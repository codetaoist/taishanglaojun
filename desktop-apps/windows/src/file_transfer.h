#pragma once

#include "pch.h"
#include <functional>
#include <memory>
#include <vector>
#include <string>
#include <atomic>
#include <thread>
#include <mutex>

// 设备类型枚举
enum class DeviceType {
    UNKNOWN = 0,
    DESKTOP_WINDOWS,
    DESKTOP_MACOS,
    DESKTOP_LINUX,
    MOBILE_ANDROID,
    MOBILE_IOS,
    WEB_BROWSER
};

// 传输状态枚举
enum class TransferStatus {
    PENDING = 0,
    CONNECTING,
    TRANSFERRING,
    PAUSED,
    COMPLETED,
    FAILED,
    CANCELLED
};

// 传输错误枚举
enum class TransferError {
    NONE = 0,
    NETWORK_ERROR,
    FILE_NOT_FOUND,
    PERMISSION_DENIED,
    DISK_FULL,
    CHECKSUM_MISMATCH,
    TIMEOUT,
    CANCELLED_BY_USER,
    UNKNOWN_ERROR
};

// 设备信息结构
struct DeviceInfo {
    std::string deviceId;
    std::string deviceName;
    DeviceType deviceType;
    std::string ipAddress;
    uint16_t port;
    bool isOnline;
    uint64_t lastSeen;
    std::string osVersion;
    std::string appVersion;
};

// 文件信息结构
struct FileInfo {
    std::string fileName;
    std::string filePath;
    uint64_t fileSize;
    std::string mimeType;
    std::string checksum;
    uint64_t lastModified;
    bool isDirectory;
    std::vector<FileInfo> children; // 用于目录
};

// 传输进度信息
struct TransferProgress {
    uint32_t transferId;
    std::string fileName;
    uint64_t totalBytes;
    uint64_t transferredBytes;
    float percentage;
    uint64_t speed;        // bytes per second
    uint64_t remainingTime; // seconds
    TransferStatus status;
    TransferError error;
};

// 传输会话信息
struct TransferSession {
    uint32_t sessionId;
    DeviceInfo remoteDevice;
    bool isIncoming;
    std::vector<uint32_t> activeTransfers;
    uint64_t createdTime;
    uint64_t lastActivity;
};

// 回调函数类型
using DeviceDiscoveredCallback = std::function<void(const DeviceInfo&)>;
using DeviceDisconnectedCallback = std::function<void(const std::string& deviceId)>;
using TransferProgressCallback = std::function<void(const TransferProgress&)>;
using TransferCompletedCallback = std::function<void(uint32_t transferId, bool success, const std::string& error)>;
using FileReceivedCallback = std::function<void(const FileInfo&, const std::string& savePath)>;

// 文件传输管理器类
class FileTransferManager {
public:
    FileTransferManager();
    ~FileTransferManager();

    // 初始化和清理
    bool Initialize(const std::string& deviceName, DeviceType deviceType);
    void Shutdown();

    // 服务控制
    bool StartService(uint16_t port = 0);
    void StopService();
    bool IsServiceRunning() const { return m_serviceRunning; }

    // 设备发现
    bool StartDiscovery();
    void StopDiscovery();
    bool IsDiscoveryActive() const { return m_discoveryActive; }
    std::vector<DeviceInfo> GetDiscoveredDevices() const;
    void RefreshDeviceList();

    // 连接管理
    uint32_t ConnectToDevice(const DeviceInfo& device);
    void DisconnectFromDevice(uint32_t sessionId);
    void DisconnectAll();
    std::vector<TransferSession> GetActiveSessions() const;

    // 文件传输
    uint32_t SendFile(uint32_t sessionId, const std::string& filePath);
    uint32_t SendFiles(uint32_t sessionId, const std::vector<std::string>& filePaths);
    uint32_t SendDirectory(uint32_t sessionId, const std::string& dirPath);
    bool PauseTransfer(uint32_t transferId);
    bool ResumeTransfer(uint32_t transferId);
    bool CancelTransfer(uint32_t transferId);

    // 传输状态查询
    TransferProgress GetTransferProgress(uint32_t transferId) const;
    std::vector<TransferProgress> GetAllTransfers() const;
    std::vector<TransferProgress> GetActiveTransfers() const;

    // 设置和配置
    void SetReceiveDirectory(const std::string& path);
    std::string GetReceiveDirectory() const { return m_receiveDirectory; }
    void SetMaxConcurrentTransfers(int maxTransfers);
    int GetMaxConcurrentTransfers() const { return m_maxConcurrentTransfers; }
    void SetTransferChunkSize(size_t chunkSize);
    size_t GetTransferChunkSize() const { return m_transferChunkSize; }

    // 回调设置
    void SetDeviceDiscoveredCallback(DeviceDiscoveredCallback callback);
    void SetDeviceDisconnectedCallback(DeviceDisconnectedCallback callback);
    void SetTransferProgressCallback(TransferProgressCallback callback);
    void SetTransferCompletedCallback(TransferCompletedCallback callback);
    void SetFileReceivedCallback(FileReceivedCallback callback);

    // 设备信息
    const DeviceInfo& GetLocalDevice() const { return m_localDevice; }
    void SetLocalDeviceName(const std::string& name);

    // 网络信息
    std::vector<std::string> GetLocalIPAddresses() const;
    uint16_t GetServicePort() const { return m_servicePort; }

    // 统计信息
    struct Statistics {
        uint64_t totalBytesSent;
        uint64_t totalBytesReceived;
        uint32_t totalFilesSent;
        uint32_t totalFilesReceived;
        uint32_t successfulTransfers;
        uint32_t failedTransfers;
        uint64_t totalTransferTime;
    };
    Statistics GetStatistics() const { return m_statistics; }
    void ResetStatistics();

private:
    // 内部实现方法
    bool InitializeNetwork();
    void CleanupNetwork();
    bool CreateListenSocket();
    bool CreateDiscoverySocket();
    void StartWorkerThreads();
    void StopWorkerThreads();
    
    // 网络处理
    void ServerThreadProc();
    void DiscoveryThreadProc();
    void WorkerThreadProc();
    void HandleClientConnection(SOCKET clientSocket);
    void ProcessIncomingTransfer(SOCKET socket);
    
    // 设备发现协议
    void SendDiscoveryBroadcast();
    void ProcessDiscoveryMessage(const std::string& message, const std::string& senderIP);
    void SendDiscoveryResponse(const std::string& targetIP, uint16_t targetPort);
    
    // 传输处理
    void ProcessSendFileTask(uint32_t sessionId, uint32_t transferId, const std::string& filePath);
    void ProcessReceiveFileTask(SOCKET socket, const FileInfo& fileInfo);
    bool SendFileData(SOCKET socket, const std::string& filePath, uint32_t transferId);
    bool ReceiveFileData(SOCKET socket, const std::string& savePath, const FileInfo& fileInfo, uint32_t transferId);
    
    // 工具方法
    uint32_t GenerateTransferId();
    uint32_t GenerateSessionId();
    std::string CalculateFileChecksum(const std::string& filePath);
    FileInfo GetFileInfo(const std::string& filePath);
    bool ValidateFileChecksum(const std::string& filePath, const std::string& expectedChecksum);
    
    // 回调触发
    void TriggerDeviceDiscovered(const DeviceInfo& device);
    void TriggerDeviceDisconnected(const std::string& deviceId);
    void TriggerTransferProgress(const TransferProgress& progress);
    void TriggerTransferCompleted(uint32_t transferId, bool success, const std::string& error);
    void TriggerFileReceived(const FileInfo& fileInfo, const std::string& savePath);

private:
    // 基本状态
    bool m_initialized;
    bool m_serviceRunning;
    bool m_discoveryActive;
    
    // 设备信息
    DeviceInfo m_localDevice;
    std::vector<DeviceInfo> m_discoveredDevices;
    mutable std::mutex m_devicesMutex;
    
    // 网络
    SOCKET m_listenSocket;
    SOCKET m_discoverySocket;
    uint16_t m_servicePort;
    std::vector<std::string> m_localIPs;
    
    // 线程管理
    std::thread m_serverThread;
    std::thread m_discoveryThread;
    std::vector<std::thread> m_workerThreads;
    std::atomic<bool> m_running;
    
    // 传输管理
    std::map<uint32_t, std::unique_ptr<TransferSession>> m_sessions;
    std::map<uint32_t, TransferProgress> m_transfers;
    mutable std::mutex m_transfersMutex;
    mutable std::mutex m_sessionsMutex;
    
    // 任务队列
    std::queue<std::function<void()>> m_taskQueue;
    std::mutex m_taskMutex;
    std::condition_variable m_taskCondition;
    
    // 配置
    std::string m_receiveDirectory;
    int m_maxConcurrentTransfers;
    size_t m_transferChunkSize;
    
    // 回调
    DeviceDiscoveredCallback m_deviceDiscoveredCallback;
    DeviceDisconnectedCallback m_deviceDisconnectedCallback;
    TransferProgressCallback m_transferProgressCallback;
    TransferCompletedCallback m_transferCompletedCallback;
    FileReceivedCallback m_fileReceivedCallback;
    
    // 统计
    Statistics m_statistics;
    mutable std::mutex m_statisticsMutex;
    
    // ID生成器
    std::atomic<uint32_t> m_nextTransferId;
    std::atomic<uint32_t> m_nextSessionId;
};

// 全局函数
FileTransferManager* GetFileTransferManager();
bool InitializeFileTransferSystem();
void ShutdownFileTransferSystem();