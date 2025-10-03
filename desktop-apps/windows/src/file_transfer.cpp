#include "pch.h"
#include "file_transfer.h"
#include "network.h"
#include "utils.h"
#include <winsock2.h>
#include <ws2tcpip.h>
#include <iphlpapi.h>
#include <thread>
#include <mutex>
#include <condition_variable>
#include <atomic>
#include <queue>
#include <unordered_map>

#pragma comment(lib, "ws2_32.lib")
#pragma comment(lib, "iphlpapi.lib")

// Windows特定的文件传输实现
class WindowsFileTransferManager {
private:
    FileTransferManager* manager_;
    std::mutex mutex_;
    std::condition_variable cv_;
    std::atomic<bool> running_;
    std::thread discovery_thread_;
    std::thread server_thread_;
    std::vector<std::thread> worker_threads_;
    std::queue<std::function<void()>> task_queue_;
    std::unordered_map<uint32_t, std::unique_ptr<FileTransferSession>> sessions_;
    
    SOCKET listen_socket_;
    SOCKET discovery_socket_;
    
public:
    WindowsFileTransferManager(FileTransferManager* manager) 
        : manager_(manager), running_(false), listen_socket_(INVALID_SOCKET), discovery_socket_(INVALID_SOCKET) {
        
        // 初始化Winsock
        WSADATA wsaData;
        if (WSAStartup(MAKEWORD(2, 2), &wsaData) != 0) {
            throw std::runtime_error("Failed to initialize Winsock");
        }
        
        // 生成设备ID
        file_transfer_generate_device_id(manager_->local_device_id, sizeof(manager_->local_device_id));
        manager_->local_device_type = DEVICE_TYPE_DESKTOP_WINDOWS;
        
        // 创建工作线程
        int thread_count = std::thread::hardware_concurrency();
        if (thread_count == 0) thread_count = 4;
        
        for (int i = 0; i < thread_count; ++i) {
            worker_threads_.emplace_back([this]() { WorkerThreadProc(); });
        }
    }
    
    ~WindowsFileTransferManager() {
        Stop();
        
        // 停止工作线程
        running_ = false;
        cv_.notify_all();
        
        for (auto& thread : worker_threads_) {
            if (thread.joinable()) {
                thread.join();
            }
        }
        
        WSACleanup();
    }
    
    bool Start(uint16_t port) {
        std::lock_guard<std::mutex> lock(mutex_);
        
        if (running_) {
            return true;
        }
        
        manager_->listen_port = port;
        
        // 创建监听套接字
        if (!CreateListenSocket(port)) {
            return false;
        }
        
        // 创建发现套接字
        if (!CreateDiscoverySocket()) {
            closesocket(listen_socket_);
            listen_socket_ = INVALID_SOCKET;
            return false;
        }
        
        running_ = true;
        manager_->is_running = true;
        
        // 启动服务器线程
        server_thread_ = std::thread([this]() { ServerThreadProc(); });
        
        // 启动发现线程
        discovery_thread_ = std::thread([this]() { DiscoveryThreadProc(); });
        
        return true;
    }
    
    void Stop() {
        std::lock_guard<std::mutex> lock(mutex_);
        
        if (!running_) {
            return;
        }
        
        running_ = false;
        manager_->is_running = false;
        manager_->should_exit = true;
        
        // 关闭套接字
        if (listen_socket_ != INVALID_SOCKET) {
            closesocket(listen_socket_);
            listen_socket_ = INVALID_SOCKET;
        }
        
        if (discovery_socket_ != INVALID_SOCKET) {
            closesocket(discovery_socket_);
            discovery_socket_ = INVALID_SOCKET;
        }
        
        // 等待线程结束
        if (server_thread_.joinable()) {
            server_thread_.join();
        }
        
        if (discovery_thread_.joinable()) {
            discovery_thread_.join();
        }
        
        // 关闭所有会话
        sessions_.clear();
    }
    
    bool StartDiscovery() {
        manager_->discovery_enabled = true;
        return true;
    }
    
    void StopDiscovery() {
        manager_->discovery_enabled = false;
    }
    
    uint32_t ConnectToDevice(const DeviceInfo* device) {
        if (!device || !running_) {
            return 0;
        }
        
        // 创建连接套接字
        SOCKET client_socket = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);
        if (client_socket == INVALID_SOCKET) {
            return 0;
        }
        
        // 连接到设备
        sockaddr_in addr = {};
        addr.sin_family = AF_INET;
        addr.sin_addr.s_addr = htonl(device->ip_address);
        addr.sin_port = htons(device->port);
        
        if (connect(client_socket, (sockaddr*)&addr, sizeof(addr)) == SOCKET_ERROR) {
            closesocket(client_socket);
            return 0;
        }
        
        // 发送连接请求
        ConnectRequest request = {};
        strncpy_s(request.device_id, manager_->local_device_id, sizeof(request.device_id) - 1);
        strncpy_s(request.device_name, manager_->local_device_name, sizeof(request.device_name) - 1);
        request.device_type = manager_->local_device_type;
        request.protocol_version = FILE_TRANSFER_PROTOCOL_VERSION;
        request.request_encryption = manager_->encryption_enabled;
        
        FileTransferHeader header = {};
        header.magic = FILE_TRANSFER_MAGIC;
        header.version = FILE_TRANSFER_PROTOCOL_VERSION;
        header.message_type = FT_MSG_CONNECT_REQUEST;
        header.message_id = GenerateMessageId();
        header.data_length = sizeof(request);
        header.timestamp = file_transfer_get_current_time_ms();
        header.checksum = file_transfer_calculate_checksum(&request, sizeof(request));
        
        if (!SendMessage(client_socket, &header, &request)) {
            closesocket(client_socket);
            return 0;
        }
        
        // 接收连接响应
        FileTransferHeader response_header;
        void* response_data = nullptr;
        
        if (!ReceiveMessage(client_socket, &response_header, &response_data)) {
            closesocket(client_socket);
            return 0;
        }
        
        if (response_header.message_type != FT_MSG_CONNECT_RESPONSE) {
            free(response_data);
            closesocket(client_socket);
            return 0;
        }
        
        ConnectResponse* response = (ConnectResponse*)response_data;
        
        if (!response->connection_accepted) {
            free(response_data);
            closesocket(client_socket);
            return 0;
        }
        
        // 创建会话
        uint32_t session_id = response->session_id;
        auto session = std::make_unique<FileTransferSession>();
        session->session_id = session_id;
        strncpy_s(session->session_token, response->session_token, sizeof(session->session_token) - 1);
        session->remote_device = *device;
        session->status = FT_STATUS_CONNECTED;
        session->start_time = file_transfer_get_current_time_ms();
        session->last_activity_time = session->start_time;
        session->chunk_size = std::min(response->max_chunk_size, manager_->max_chunk_size);
        
        sessions_[session_id] = std::move(session);
        
        free(response_data);
        
        // 通知连接成功
        if (manager_->device_connected_callback) {
            manager_->device_connected_callback(device, session_id, manager_->callback_user_data);
        }
        
        return session_id;
    }
    
    uint32_t SendFile(uint32_t session_id, const char* file_path) {
        auto it = sessions_.find(session_id);
        if (it == sessions_.end()) {
            return 0;
        }
        
        FileTransferSession* session = it->second.get();
        if (session->status != FT_STATUS_CONNECTED) {
            return 0;
        }
        
        // 获取文件信息
        WIN32_FILE_ATTRIBUTE_DATA file_attr;
        if (!GetFileAttributesExA(file_path, GetFileExInfoStandard, &file_attr)) {
            return 0;
        }
        
        // 填充文件信息
        FileInfo file_info = {};
        const char* file_name = strrchr(file_path, '\\');
        if (file_name) {
            file_name++;
        } else {
            file_name = file_path;
        }
        
        strncpy_s(file_info.file_name, file_name, sizeof(file_info.file_name) - 1);
        strncpy_s(file_info.file_path, file_path, sizeof(file_info.file_path) - 1);
        
        LARGE_INTEGER file_size;
        file_size.LowPart = file_attr.nFileSizeLow;
        file_size.HighPart = file_attr.nFileSizeHigh;
        file_info.file_size = file_size.QuadPart;
        
        FILETIME ft = file_attr.ftLastWriteTime;
        ULARGE_INTEGER uli;
        uli.LowPart = ft.dwLowDateTime;
        uli.HighPart = ft.dwHighDateTime;
        file_info.modified_time = uli.QuadPart;
        
        file_info.file_hash = file_transfer_calculate_file_hash(file_path);
        file_info.is_directory = (file_attr.dwFileAttributes & FILE_ATTRIBUTE_DIRECTORY) != 0;
        
        // 发送文件请求
        FileRequest request = {};
        request.file_info = file_info;
        request.chunk_size = session->chunk_size;
        request.resume_transfer = false;
        request.resume_offset = 0;
        
        uint32_t transfer_id = GenerateTransferId();
        
        // 添加到任务队列
        AddTask([this, session_id, transfer_id, file_path, file_info]() {
            SendFileTask(session_id, transfer_id, file_path, file_info);
        });
        
        return transfer_id;
    }
    
private:
    bool CreateListenSocket(uint16_t port) {
        listen_socket_ = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);
        if (listen_socket_ == INVALID_SOCKET) {
            return false;
        }
        
        // 设置套接字选项
        int opt = 1;
        setsockopt(listen_socket_, SOL_SOCKET, SO_REUSEADDR, (char*)&opt, sizeof(opt));
        
        // 绑定地址
        sockaddr_in addr = {};
        addr.sin_family = AF_INET;
        addr.sin_addr.s_addr = INADDR_ANY;
        addr.sin_port = htons(port);
        
        if (bind(listen_socket_, (sockaddr*)&addr, sizeof(addr)) == SOCKET_ERROR) {
            closesocket(listen_socket_);
            listen_socket_ = INVALID_SOCKET;
            return false;
        }
        
        // 开始监听
        if (listen(listen_socket_, SOMAXCONN) == SOCKET_ERROR) {
            closesocket(listen_socket_);
            listen_socket_ = INVALID_SOCKET;
            return false;
        }
        
        return true;
    }
    
    bool CreateDiscoverySocket() {
        discovery_socket_ = socket(AF_INET, SOCK_DGRAM, IPPROTO_UDP);
        if (discovery_socket_ == INVALID_SOCKET) {
            return false;
        }
        
        // 设置广播选项
        int opt = 1;
        setsockopt(discovery_socket_, SOL_SOCKET, SO_BROADCAST, (char*)&opt, sizeof(opt));
        
        // 绑定地址
        sockaddr_in addr = {};
        addr.sin_family = AF_INET;
        addr.sin_addr.s_addr = INADDR_ANY;
        addr.sin_port = htons(FILE_TRANSFER_DISCOVERY_PORT);
        
        if (bind(discovery_socket_, (sockaddr*)&addr, sizeof(addr)) == SOCKET_ERROR) {
            closesocket(discovery_socket_);
            discovery_socket_ = INVALID_SOCKET;
            return false;
        }
        
        return true;
    }
    
    void ServerThreadProc() {
        while (running_) {
            fd_set read_fds;
            FD_ZERO(&read_fds);
            FD_SET(listen_socket_, &read_fds);
            
            timeval timeout = { 1, 0 }; // 1秒超时
            
            int result = select(0, &read_fds, nullptr, nullptr, &timeout);
            if (result == SOCKET_ERROR || !running_) {
                break;
            }
            
            if (result > 0 && FD_ISSET(listen_socket_, &read_fds)) {
                // 接受新连接
                sockaddr_in client_addr;
                int addr_len = sizeof(client_addr);
                SOCKET client_socket = accept(listen_socket_, (sockaddr*)&client_addr, &addr_len);
                
                if (client_socket != INVALID_SOCKET) {
                    // 处理客户端连接
                    AddTask([this, client_socket, client_addr]() {
                        HandleClientConnection(client_socket, client_addr);
                    });
                }
            }
        }
    }
    
    void DiscoveryThreadProc() {
        while (running_) {
            if (manager_->discovery_enabled) {
                // 发送发现广播
                SendDiscoveryBroadcast();
                
                // 接收发现响应
                ReceiveDiscoveryMessages();
            }
            
            std::this_thread::sleep_for(std::chrono::milliseconds(FILE_TRANSFER_DISCOVERY_INTERVAL_MS));
        }
    }
    
    void WorkerThreadProc() {
        while (running_) {
            std::function<void()> task;
            
            {
                std::unique_lock<std::mutex> lock(mutex_);
                cv_.wait(lock, [this]() { return !task_queue_.empty() || !running_; });
                
                if (!running_) {
                    break;
                }
                
                if (!task_queue_.empty()) {
                    task = std::move(task_queue_.front());
                    task_queue_.pop();
                }
            }
            
            if (task) {
                try {
                    task();
                } catch (const std::exception& e) {
                    // 记录错误
                    OutputDebugStringA(("File transfer task error: " + std::string(e.what()) + "\n").c_str());
                }
            }
        }
    }
    
    void AddTask(std::function<void()> task) {
        {
            std::lock_guard<std::mutex> lock(mutex_);
            task_queue_.push(std::move(task));
        }
        cv_.notify_one();
    }
    
    void HandleClientConnection(SOCKET client_socket, const sockaddr_in& client_addr) {
        // TODO: 实现客户端连接处理
        closesocket(client_socket);
    }
    
    void SendDiscoveryBroadcast() {
        DiscoveryRequest request = {};
        strncpy_s(request.device_id, manager_->local_device_id, sizeof(request.device_id) - 1);
        strncpy_s(request.device_name, manager_->local_device_name, sizeof(request.device_name) - 1);
        request.device_type = manager_->local_device_type;
        request.listen_port = manager_->listen_port;
        request.supports_encryption = manager_->encryption_enabled;
        request.max_chunk_size = manager_->max_chunk_size;
        
        FileTransferHeader header = {};
        header.magic = FILE_TRANSFER_MAGIC;
        header.version = FILE_TRANSFER_PROTOCOL_VERSION;
        header.message_type = FT_MSG_DISCOVERY_REQUEST;
        header.message_id = GenerateMessageId();
        header.data_length = sizeof(request);
        header.timestamp = file_transfer_get_current_time_ms();
        header.checksum = file_transfer_calculate_checksum(&request, sizeof(request));
        
        // 广播到所有网络接口
        sockaddr_in broadcast_addr = {};
        broadcast_addr.sin_family = AF_INET;
        broadcast_addr.sin_addr.s_addr = INADDR_BROADCAST;
        broadcast_addr.sin_port = htons(FILE_TRANSFER_DISCOVERY_PORT);
        
        char buffer[sizeof(FileTransferHeader) + sizeof(DiscoveryRequest)];
        memcpy(buffer, &header, sizeof(header));
        memcpy(buffer + sizeof(header), &request, sizeof(request));
        
        sendto(discovery_socket_, buffer, sizeof(buffer), 0, (sockaddr*)&broadcast_addr, sizeof(broadcast_addr));
    }
    
    void ReceiveDiscoveryMessages() {
        fd_set read_fds;
        FD_ZERO(&read_fds);
        FD_SET(discovery_socket_, &read_fds);
        
        timeval timeout = { 0, 100000 }; // 100ms超时
        
        int result = select(0, &read_fds, nullptr, nullptr, &timeout);
        if (result > 0 && FD_ISSET(discovery_socket_, &read_fds)) {
            char buffer[1024];
            sockaddr_in sender_addr;
            int addr_len = sizeof(sender_addr);
            
            int received = recvfrom(discovery_socket_, buffer, sizeof(buffer), 0, (sockaddr*)&sender_addr, &addr_len);
            if (received >= sizeof(FileTransferHeader)) {
                ProcessDiscoveryMessage(buffer, received, sender_addr);
            }
        }
    }
    
    void ProcessDiscoveryMessage(const char* buffer, int size, const sockaddr_in& sender_addr) {
        if (size < sizeof(FileTransferHeader)) {
            return;
        }
        
        const FileTransferHeader* header = (const FileTransferHeader*)buffer;
        
        if (header->magic != FILE_TRANSFER_MAGIC || header->version != FILE_TRANSFER_PROTOCOL_VERSION) {
            return;
        }
        
        if (header->message_type == FT_MSG_DISCOVERY_REQUEST) {
            // 处理发现请求
            if (size >= sizeof(FileTransferHeader) + sizeof(DiscoveryRequest)) {
                const DiscoveryRequest* request = (const DiscoveryRequest*)(buffer + sizeof(FileTransferHeader));
                
                // 检查是否是自己的请求
                if (strcmp(request->device_id, manager_->local_device_id) == 0) {
                    return;
                }
                
                // 发送发现响应
                SendDiscoveryResponse(sender_addr, request);
            }
        } else if (header->message_type == FT_MSG_DISCOVERY_RESPONSE) {
            // 处理发现响应
            if (size >= sizeof(FileTransferHeader) + sizeof(DiscoveryResponse)) {
                const DiscoveryResponse* response = (const DiscoveryResponse*)(buffer + sizeof(FileTransferHeader));
                
                // 检查是否是自己的响应
                if (strcmp(response->device_id, manager_->local_device_id) == 0) {
                    return;
                }
                
                // 添加到已发现设备列表
                AddDiscoveredDevice(response, sender_addr);
            }
        }
    }
    
    void SendDiscoveryResponse(const sockaddr_in& sender_addr, const DiscoveryRequest* request) {
        DiscoveryResponse response = {};
        strncpy_s(response.device_id, manager_->local_device_id, sizeof(response.device_id) - 1);
        strncpy_s(response.device_name, manager_->local_device_name, sizeof(response.device_name) - 1);
        response.device_type = manager_->local_device_type;
        response.listen_port = manager_->listen_port;
        response.supports_encryption = manager_->encryption_enabled;
        response.max_chunk_size = manager_->max_chunk_size;
        response.accepts_connections = true;
        
        FileTransferHeader header = {};
        header.magic = FILE_TRANSFER_MAGIC;
        header.version = FILE_TRANSFER_PROTOCOL_VERSION;
        header.message_type = FT_MSG_DISCOVERY_RESPONSE;
        header.message_id = GenerateMessageId();
        header.data_length = sizeof(response);
        header.timestamp = file_transfer_get_current_time_ms();
        header.checksum = file_transfer_calculate_checksum(&response, sizeof(response));
        
        char buffer[sizeof(FileTransferHeader) + sizeof(DiscoveryResponse)];
        memcpy(buffer, &header, sizeof(header));
        memcpy(buffer + sizeof(header), &response, sizeof(response));
        
        sendto(discovery_socket_, buffer, sizeof(buffer), 0, (sockaddr*)&sender_addr, sizeof(sender_addr));
    }
    
    void AddDiscoveredDevice(const DiscoveryResponse* response, const sockaddr_in& sender_addr) {
        // 检查是否已经存在
        for (int i = 0; i < manager_->discovered_device_count; ++i) {
            if (strcmp(manager_->discovered_devices[i].device_id, response->device_id) == 0) {
                // 更新现有设备信息
                manager_->discovered_devices[i].ip_address = ntohl(sender_addr.sin_addr.s_addr);
                manager_->discovered_devices[i].port = response->listen_port;
                manager_->discovered_devices[i].last_seen = file_transfer_get_current_time_ms();
                return;
            }
        }
        
        // 添加新设备
        if (manager_->discovered_device_count < 32) {
            DeviceInfo* device = &manager_->discovered_devices[manager_->discovered_device_count++];
            strncpy_s(device->device_id, response->device_id, sizeof(device->device_id) - 1);
            strncpy_s(device->device_name, response->device_name, sizeof(device->device_name) - 1);
            device->device_type = response->device_type;
            device->ip_address = ntohl(sender_addr.sin_addr.s_addr);
            device->port = response->listen_port;
            device->last_seen = file_transfer_get_current_time_ms();
            device->is_trusted = false;
            device->supports_encryption = response->supports_encryption;
            device->max_chunk_size = response->max_chunk_size;
            
            // 通知设备发现
            if (manager_->device_discovered_callback) {
                manager_->device_discovered_callback(device, manager_->callback_user_data);
            }
        }
    }
    
    void SendFileTask(uint32_t session_id, uint32_t transfer_id, const std::string& file_path, const FileInfo& file_info) {
        // TODO: 实现文件发送任务
    }
    
    bool SendMessage(SOCKET socket, const FileTransferHeader* header, const void* data) {
        // 发送消息头
        int sent = send(socket, (const char*)header, sizeof(FileTransferHeader), 0);
        if (sent != sizeof(FileTransferHeader)) {
            return false;
        }
        
        // 发送数据
        if (header->data_length > 0 && data) {
            sent = send(socket, (const char*)data, header->data_length, 0);
            if (sent != (int)header->data_length) {
                return false;
            }
        }
        
        return true;
    }
    
    bool ReceiveMessage(SOCKET socket, FileTransferHeader* header, void** data) {
        // 接收消息头
        int received = recv(socket, (char*)header, sizeof(FileTransferHeader), MSG_WAITALL);
        if (received != sizeof(FileTransferHeader)) {
            return false;
        }
        
        // 验证魔数和版本
        if (header->magic != FILE_TRANSFER_MAGIC || header->version != FILE_TRANSFER_PROTOCOL_VERSION) {
            return false;
        }
        
        // 接收数据
        *data = nullptr;
        if (header->data_length > 0) {
            *data = malloc(header->data_length);
            if (!*data) {
                return false;
            }
            
            received = recv(socket, (char*)*data, header->data_length, MSG_WAITALL);
            if (received != (int)header->data_length) {
                free(*data);
                *data = nullptr;
                return false;
            }
            
            // 验证校验和
            uint32_t calculated_checksum = file_transfer_calculate_checksum(*data, header->data_length);
            if (calculated_checksum != header->checksum) {
                free(*data);
                *data = nullptr;
                return false;
            }
        }
        
        return true;
    }
    
    uint32_t GenerateMessageId() {
        static std::atomic<uint32_t> counter(1);
        return counter.fetch_add(1);
    }
    
    uint32_t GenerateTransferId() {
        static std::atomic<uint32_t> counter(1);
        return counter.fetch_add(1);
    }
};

// 全局管理器实例
static std::unique_ptr<WindowsFileTransferManager> g_windows_manager;

// C接口实现
extern "C" {

FileTransferManager* file_transfer_manager_create(const char* device_name, DeviceType device_type) {
    try {
        FileTransferManager* manager = (FileTransferManager*)calloc(1, sizeof(FileTransferManager));
        if (!manager) {
            return nullptr;
        }
        
        if (device_name) {
            strncpy_s(manager->local_device_name, device_name, sizeof(manager->local_device_name) - 1);
        } else {
            strcpy_s(manager->local_device_name, "Windows Desktop");
        }
        
        manager->local_device_type = device_type;
        manager->max_chunk_size = DEFAULT_CHUNK_SIZE;
        manager->encryption_enabled = true;
        
        g_windows_manager = std::make_unique<WindowsFileTransferManager>(manager);
        
        return manager;
    } catch (const std::exception&) {
        return nullptr;
    }
}

void file_transfer_manager_destroy(FileTransferManager* manager) {
    if (manager) {
        g_windows_manager.reset();
        free(manager);
    }
}

bool file_transfer_manager_start(FileTransferManager* manager, uint16_t port) {
    if (!manager || !g_windows_manager) {
        return false;
    }
    
    return g_windows_manager->Start(port);
}

void file_transfer_manager_stop(FileTransferManager* manager) {
    if (manager && g_windows_manager) {
        g_windows_manager->Stop();
    }
}

bool file_transfer_start_discovery(FileTransferManager* manager) {
    if (!manager || !g_windows_manager) {
        return false;
    }
    
    return g_windows_manager->StartDiscovery();
}

void file_transfer_stop_discovery(FileTransferManager* manager) {
    if (manager && g_windows_manager) {
        g_windows_manager->StopDiscovery();
    }
}

uint32_t file_transfer_connect_to_device(FileTransferManager* manager, const DeviceInfo* device) {
    if (!manager || !device || !g_windows_manager) {
        return 0;
    }
    
    return g_windows_manager->ConnectToDevice(device);
}

uint32_t file_transfer_send_file(FileTransferManager* manager, uint32_t session_id, const char* file_path) {
    if (!manager || !file_path || !g_windows_manager) {
        return 0;
    }
    
    return g_windows_manager->SendFile(session_id, file_path);
}

// 工具函数实现
void file_transfer_generate_device_id(char* device_id, size_t size) {
    if (!device_id || size < MAX_DEVICE_ID_LENGTH) {
        return;
    }
    
    // 使用MAC地址和计算机名生成设备ID
    char computer_name[MAX_COMPUTERNAME_LENGTH + 1];
    DWORD name_size = sizeof(computer_name);
    GetComputerNameA(computer_name, &name_size);
    
    // 获取第一个网络适配器的MAC地址
    IP_ADAPTER_INFO adapter_info[16];
    DWORD buffer_size = sizeof(adapter_info);
    
    if (GetAdaptersInfo(adapter_info, &buffer_size) == ERROR_SUCCESS) {
        sprintf_s(device_id, size, "WIN_%s_%02X%02X%02X%02X%02X%02X",
                  computer_name,
                  adapter_info[0].Address[0], adapter_info[0].Address[1],
                  adapter_info[0].Address[2], adapter_info[0].Address[3],
                  adapter_info[0].Address[4], adapter_info[0].Address[5]);
    } else {
        sprintf_s(device_id, size, "WIN_%s_%08X", computer_name, GetTickCount());
    }
}

uint64_t file_transfer_get_current_time_ms(void) {
    return GetTickCount64();
}

uint32_t file_transfer_calculate_checksum(const void* data, size_t size) {
    if (!data || size == 0) {
        return 0;
    }
    
    uint32_t checksum = 0;
    const uint8_t* bytes = (const uint8_t*)data;
    
    for (size_t i = 0; i < size; ++i) {
        checksum = (checksum << 1) ^ bytes[i];
    }
    
    return checksum;
}

uint32_t file_transfer_calculate_file_hash(const char* file_path) {
    if (!file_path) {
        return 0;
    }
    
    HANDLE file = CreateFileA(file_path, GENERIC_READ, FILE_SHARE_READ, nullptr, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, nullptr);
    if (file == INVALID_HANDLE_VALUE) {
        return 0;
    }
    
    uint32_t hash = 0;
    char buffer[4096];
    DWORD bytes_read;
    
    while (ReadFile(file, buffer, sizeof(buffer), &bytes_read, nullptr) && bytes_read > 0) {
        for (DWORD i = 0; i < bytes_read; ++i) {
            hash = (hash << 1) ^ buffer[i];
        }
    }
    
    CloseHandle(file);
    return hash;
}

bool file_transfer_file_exists(const char* path) {
    if (!path) {
        return false;
    }
    
    DWORD attributes = GetFileAttributesA(path);
    return (attributes != INVALID_FILE_ATTRIBUTES);
}

uint64_t file_transfer_get_file_size(const char* path) {
    if (!path) {
        return 0;
    }
    
    WIN32_FILE_ATTRIBUTE_DATA file_attr;
    if (!GetFileAttributesExA(path, GetFileExInfoStandard, &file_attr)) {
        return 0;
    }
    
    LARGE_INTEGER file_size;
    file_size.LowPart = file_attr.nFileSizeLow;
    file_size.HighPart = file_attr.nFileSizeHigh;
    
    return file_size.QuadPart;
}

const char* file_transfer_status_to_string(FileTransferStatus status) {
    switch (status) {
        case FT_STATUS_IDLE: return "Idle";
        case FT_STATUS_DISCOVERING: return "Discovering";
        case FT_STATUS_CONNECTING: return "Connecting";
        case FT_STATUS_AUTHENTICATING: return "Authenticating";
        case FT_STATUS_CONNECTED: return "Connected";
        case FT_STATUS_TRANSFERRING: return "Transferring";
        case FT_STATUS_PAUSED: return "Paused";
        case FT_STATUS_COMPLETED: return "Completed";
        case FT_STATUS_CANCELLED: return "Cancelled";
        case FT_STATUS_ERROR: return "Error";
        case FT_STATUS_DISCONNECTED: return "Disconnected";
        default: return "Unknown";
    }
}

const char* file_transfer_error_to_string(FileTransferError error) {
    switch (error) {
        case FT_ERROR_NONE: return "No error";
        case FT_ERROR_NETWORK_FAILURE: return "Network failure";
        case FT_ERROR_CONNECTION_TIMEOUT: return "Connection timeout";
        case FT_ERROR_AUTH_FAILED: return "Authentication failed";
        case FT_ERROR_FILE_NOT_FOUND: return "File not found";
        case FT_ERROR_FILE_ACCESS_DENIED: return "File access denied";
        case FT_ERROR_INSUFFICIENT_SPACE: return "Insufficient space";
        case FT_ERROR_TRANSFER_CANCELLED: return "Transfer cancelled";
        case FT_ERROR_PROTOCOL_ERROR: return "Protocol error";
        case FT_ERROR_CHECKSUM_MISMATCH: return "Checksum mismatch";
        case FT_ERROR_DEVICE_NOT_FOUND: return "Device not found";
        case FT_ERROR_INVALID_REQUEST: return "Invalid request";
        case FT_ERROR_UNSUPPORTED_VERSION: return "Unsupported version";
        default: return "Unknown error";
    }
}

const char* device_type_to_string(DeviceType type) {
    switch (type) {
        case DEVICE_TYPE_DESKTOP_WINDOWS: return "Windows Desktop";
        case DEVICE_TYPE_DESKTOP_MACOS: return "macOS Desktop";
        case DEVICE_TYPE_DESKTOP_LINUX: return "Linux Desktop";
        case DEVICE_TYPE_MOBILE_ANDROID: return "Android Mobile";
        case DEVICE_TYPE_MOBILE_IOS: return "iOS Mobile";
        case DEVICE_TYPE_WEB_BROWSER: return "Web Browser";
        default: return "Unknown Device";
    }
}

} // extern "C"