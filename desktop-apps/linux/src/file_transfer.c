#include "file_transfer.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <pthread.h>
#include <sys/socket.h>
#include <sys/stat.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <ifaddrs.h>
#include <net/if.h>
#include <errno.h>
#include <fcntl.h>
#include <time.h>
#include <signal.h>

// MARK: - Internal Structures

typedef struct {
    FileTransferManager* manager;
    int socket_fd;
    struct sockaddr_in client_addr;
    socklen_t addr_len;
} ConnectionContext;

typedef struct {
    FileTransferManager* manager;
    uint32_t session_id;
    uint32_t transfer_id;
    char file_path[MAX_PATH_LENGTH];
} TransferContext;

// MARK: - Global Variables

static volatile sig_atomic_t g_shutdown_requested = 0;

// MARK: - Signal Handlers

static void signal_handler(int sig) {
    if (sig == SIGINT || sig == SIGTERM) {
        g_shutdown_requested = 1;
    }
}

// MARK: - Private Function Declarations

static void* server_thread_func(void* arg);
static void* discovery_thread_func(void* arg);
static void* client_handler_thread(void* arg);
static void* file_send_thread(void* arg);
static void* file_receive_thread(void* arg);

static bool create_listen_socket(FileTransferManager* manager);
static bool create_discovery_socket(FileTransferManager* manager);
static void handle_client_connection(FileTransferManager* manager, int client_fd, struct sockaddr_in* client_addr);
static void process_message(FileTransferManager* manager, int client_fd, FileTransferHeader* header, void* data);

static void send_discovery_broadcast(FileTransferManager* manager);
static void handle_discovery_request(FileTransferManager* manager, DiscoveryRequest* request, struct sockaddr_in* from_addr);
static void handle_discovery_response(FileTransferManager* manager, DiscoveryResponse* response, struct sockaddr_in* from_addr);

static bool send_message(int socket_fd, FileTransferHeader* header, void* data);
static bool receive_message(int socket_fd, FileTransferHeader* header, void** data);

static char* generate_device_id(void);
static char* get_default_device_name(void);
static uint32_t generate_session_id(void);
static uint32_t generate_transfer_id(void);
static uint32_t generate_message_id(void);
static uint64_t get_current_time_ms(void);
static uint32_t calculate_checksum(const void* data, size_t length);
static uint32_t calculate_file_hash(const char* file_path);

static bool file_exists(const char* file_path);
static uint64_t get_file_size(const char* file_path);
static bool create_directories(const char* path);

// MARK: - Public API Implementation

FileTransferManager* file_transfer_manager_create(const char* device_name) {
    FileTransferManager* manager = calloc(1, sizeof(FileTransferManager));
    if (!manager) {
        return NULL;
    }
    
    // Initialize device info
    manager->local_device.device_id = generate_device_id();
    if (!manager->local_device.device_id) {
        free(manager);
        return NULL;
    }
    
    if (device_name) {
        strncpy(manager->local_device.device_name, device_name, MAX_DEVICE_NAME_LENGTH - 1);
    } else {
        char* default_name = get_default_device_name();
        if (default_name) {
            strncpy(manager->local_device.device_name, default_name, MAX_DEVICE_NAME_LENGTH - 1);
            free(default_name);
        } else {
            strcpy(manager->local_device.device_name, "Linux Desktop");
        }
    }
    
    manager->local_device.device_type = DEVICE_TYPE_DESKTOP_LINUX;
    manager->local_device.listen_port = DEFAULT_LISTEN_PORT;
    manager->local_device.supports_encryption = true;
    manager->local_device.max_chunk_size = DEFAULT_CHUNK_SIZE;
    
    // Initialize configuration
    manager->config.listen_port = DEFAULT_LISTEN_PORT;
    manager->config.max_chunk_size = DEFAULT_CHUNK_SIZE;
    manager->config.enable_encryption = true;
    manager->config.enable_compression = false;
    manager->config.connection_timeout = 30000; // 30 seconds
    manager->config.transfer_timeout = 300000;  // 5 minutes
    manager->config.max_concurrent_transfers = 5;
    manager->config.max_discovered_devices = 50;
    
    // Initialize mutexes
    if (pthread_mutex_init(&manager->mutex, NULL) != 0) {
        free(manager->local_device.device_id);
        free(manager);
        return NULL;
    }
    
    if (pthread_mutex_init(&manager->devices_mutex, NULL) != 0) {
        pthread_mutex_destroy(&manager->mutex);
        free(manager->local_device.device_id);
        free(manager);
        return NULL;
    }
    
    if (pthread_mutex_init(&manager->sessions_mutex, NULL) != 0) {
        pthread_mutex_destroy(&manager->devices_mutex);
        pthread_mutex_destroy(&manager->mutex);
        free(manager->local_device.device_id);
        free(manager);
        return NULL;
    }
    
    // Initialize socket file descriptors
    manager->listen_socket = -1;
    manager->discovery_socket = -1;
    
    // Setup signal handlers
    signal(SIGINT, signal_handler);
    signal(SIGTERM, signal_handler);
    signal(SIGPIPE, SIG_IGN); // Ignore broken pipe signals
    
    printf("File transfer manager created for device: %s (%s)\n", 
           manager->local_device.device_name, manager->local_device.device_id);
    
    return manager;
}

void file_transfer_manager_destroy(FileTransferManager* manager) {
    if (!manager) return;
    
    // Stop the manager first
    file_transfer_manager_stop(manager);
    
    // Clean up discovered devices
    pthread_mutex_lock(&manager->devices_mutex);
    for (int i = 0; i < manager->discovered_devices_count; i++) {
        if (manager->discovered_devices[i].device_id) {
            free(manager->discovered_devices[i].device_id);
        }
    }
    pthread_mutex_unlock(&manager->devices_mutex);
    
    // Clean up active sessions
    pthread_mutex_lock(&manager->sessions_mutex);
    for (int i = 0; i < manager->active_sessions_count; i++) {
        if (manager->active_sessions[i].session_token) {
            free(manager->active_sessions[i].session_token);
        }
    }
    pthread_mutex_unlock(&manager->sessions_mutex);
    
    // Destroy mutexes
    pthread_mutex_destroy(&manager->sessions_mutex);
    pthread_mutex_destroy(&manager->devices_mutex);
    pthread_mutex_destroy(&manager->mutex);
    
    // Free device ID
    if (manager->local_device.device_id) {
        free(manager->local_device.device_id);
    }
    
    free(manager);
    printf("File transfer manager destroyed\n");
}

bool file_transfer_manager_start(FileTransferManager* manager, uint16_t port) {
    if (!manager || manager->is_running) {
        return false;
    }
    
    pthread_mutex_lock(&manager->mutex);
    
    // Update port if specified
    if (port > 0) {
        manager->config.listen_port = port;
        manager->local_device.listen_port = port;
    }
    
    // Create sockets
    if (!create_listen_socket(manager)) {
        pthread_mutex_unlock(&manager->mutex);
        return false;
    }
    
    if (!create_discovery_socket(manager)) {
        close(manager->listen_socket);
        manager->listen_socket = -1;
        pthread_mutex_unlock(&manager->mutex);
        return false;
    }
    
    // Start server thread
    if (pthread_create(&manager->server_thread, NULL, server_thread_func, manager) != 0) {
        close(manager->discovery_socket);
        close(manager->listen_socket);
        manager->discovery_socket = -1;
        manager->listen_socket = -1;
        pthread_mutex_unlock(&manager->mutex);
        return false;
    }
    
    // Start discovery thread
    if (pthread_create(&manager->discovery_thread, NULL, discovery_thread_func, manager) != 0) {
        manager->shutdown_requested = true;
        pthread_join(manager->server_thread, NULL);
        close(manager->discovery_socket);
        close(manager->listen_socket);
        manager->discovery_socket = -1;
        manager->listen_socket = -1;
        pthread_mutex_unlock(&manager->mutex);
        return false;
    }
    
    manager->is_running = true;
    manager->shutdown_requested = false;
    
    pthread_mutex_unlock(&manager->mutex);
    
    printf("File transfer manager started on port %d\n", manager->config.listen_port);
    return true;
}

void file_transfer_manager_stop(FileTransferManager* manager) {
    if (!manager || !manager->is_running) {
        return;
    }
    
    pthread_mutex_lock(&manager->mutex);
    
    manager->shutdown_requested = true;
    manager->discovery_enabled = false;
    
    // Close sockets to wake up threads
    if (manager->listen_socket >= 0) {
        close(manager->listen_socket);
        manager->listen_socket = -1;
    }
    
    if (manager->discovery_socket >= 0) {
        close(manager->discovery_socket);
        manager->discovery_socket = -1;
    }
    
    pthread_mutex_unlock(&manager->mutex);
    
    // Wait for threads to finish
    if (manager->server_thread) {
        pthread_join(manager->server_thread, NULL);
        manager->server_thread = 0;
    }
    
    if (manager->discovery_thread) {
        pthread_join(manager->discovery_thread, NULL);
        manager->discovery_thread = 0;
    }
    
    pthread_mutex_lock(&manager->mutex);
    manager->is_running = false;
    pthread_mutex_unlock(&manager->mutex);
    
    printf("File transfer manager stopped\n");
}

bool file_transfer_manager_start_discovery(FileTransferManager* manager) {
    if (!manager || !manager->is_running) {
        return false;
    }
    
    pthread_mutex_lock(&manager->mutex);
    manager->discovery_enabled = true;
    pthread_mutex_unlock(&manager->mutex);
    
    // Send initial discovery broadcast
    send_discovery_broadcast(manager);
    
    printf("Device discovery started\n");
    return true;
}

void file_transfer_manager_stop_discovery(FileTransferManager* manager) {
    if (!manager) return;
    
    pthread_mutex_lock(&manager->mutex);
    manager->discovery_enabled = false;
    pthread_mutex_unlock(&manager->mutex);
    
    printf("Device discovery stopped\n");
}

uint32_t file_transfer_manager_connect_to_device(FileTransferManager* manager, const DeviceInfo* device) {
    if (!manager || !device || !manager->is_running) {
        return 0;
    }
    
    // Create socket for connection
    int sock = socket(AF_INET, SOCK_STREAM, 0);
    if (sock < 0) {
        printf("Failed to create socket for connection\n");
        return 0;
    }
    
    // Set socket timeout
    struct timeval timeout;
    timeout.tv_sec = manager->config.connection_timeout / 1000;
    timeout.tv_usec = (manager->config.connection_timeout % 1000) * 1000;
    setsockopt(sock, SOL_SOCKET, SO_RCVTIMEO, &timeout, sizeof(timeout));
    setsockopt(sock, SOL_SOCKET, SO_SNDTIMEO, &timeout, sizeof(timeout));
    
    // Connect to device
    struct sockaddr_in server_addr;
    memset(&server_addr, 0, sizeof(server_addr));
    server_addr.sin_family = AF_INET;
    server_addr.sin_port = htons(device->port);
    server_addr.sin_addr.s_addr = htonl(device->ip_address);
    
    if (connect(sock, (struct sockaddr*)&server_addr, sizeof(server_addr)) < 0) {
        printf("Failed to connect to device %s: %s\n", device->device_name, strerror(errno));
        close(sock);
        return 0;
    }
    
    // Generate session ID
    uint32_t session_id = generate_session_id();
    
    // Send connection request
    ConnectRequest request;
    memset(&request, 0, sizeof(request));
    strncpy(request.device_id, manager->local_device.device_id, MAX_DEVICE_ID_LENGTH - 1);
    strncpy(request.device_name, manager->local_device.device_name, MAX_DEVICE_NAME_LENGTH - 1);
    request.device_type = manager->local_device.device_type;
    request.protocol_version = PROTOCOL_VERSION;
    request.request_encryption = manager->config.enable_encryption;
    
    FileTransferHeader header;
    memset(&header, 0, sizeof(header));
    header.magic = PROTOCOL_MAGIC;
    header.version = PROTOCOL_VERSION;
    header.message_type = MSG_TYPE_CONNECT_REQUEST;
    header.message_id = generate_message_id();
    header.session_id = session_id;
    header.data_length = sizeof(request);
    header.timestamp = get_current_time_ms();
    
    if (!send_message(sock, &header, &request)) {
        printf("Failed to send connection request\n");
        close(sock);
        return 0;
    }
    
    // Receive connection response
    FileTransferHeader response_header;
    void* response_data = NULL;
    
    if (!receive_message(sock, &response_header, &response_data)) {
        printf("Failed to receive connection response\n");
        close(sock);
        return 0;
    }
    
    if (response_header.message_type != MSG_TYPE_CONNECT_RESPONSE) {
        printf("Unexpected response message type: %d\n", response_header.message_type);
        if (response_data) free(response_data);
        close(sock);
        return 0;
    }
    
    ConnectResponse* response = (ConnectResponse*)response_data;
    
    if (!response->connection_accepted) {
        printf("Connection rejected by device: %d\n", response->error_code);
        free(response_data);
        close(sock);
        return 0;
    }
    
    // Create session
    pthread_mutex_lock(&manager->sessions_mutex);
    
    if (manager->active_sessions_count >= MAX_ACTIVE_SESSIONS) {
        printf("Maximum number of active sessions reached\n");
        pthread_mutex_unlock(&manager->sessions_mutex);
        free(response_data);
        close(sock);
        return 0;
    }
    
    FileTransferSession* session = &manager->active_sessions[manager->active_sessions_count++];
    memset(session, 0, sizeof(FileTransferSession));
    
    session->session_id = session_id;
    session->session_token = strdup(response->session_token);
    session->remote_device = *device;
    session->direction = TRANSFER_DIRECTION_SEND;
    session->status = TRANSFER_STATUS_CONNECTED;
    session->chunk_size = (response->max_chunk_size < manager->config.max_chunk_size) ? 
                         response->max_chunk_size : manager->config.max_chunk_size;
    session->start_time = get_current_time_ms();
    session->last_activity_time = session->start_time;
    session->socket_fd = sock;
    
    pthread_mutex_unlock(&manager->sessions_mutex);
    
    free(response_data);
    
    printf("Connected to device %s (Session ID: %u)\n", device->device_name, session_id);
    
    // Notify callback
    if (manager->device_connected_callback) {
        manager->device_connected_callback(device, session_id);
    }
    
    return session_id;
}

void file_transfer_manager_disconnect_from_device(FileTransferManager* manager, uint32_t session_id) {
    if (!manager) return;
    
    pthread_mutex_lock(&manager->sessions_mutex);
    
    for (int i = 0; i < manager->active_sessions_count; i++) {
        if (manager->active_sessions[i].session_id == session_id) {
            FileTransferSession* session = &manager->active_sessions[i];
            
            // Close socket
            if (session->socket_fd >= 0) {
                close(session->socket_fd);
            }
            
            // Free session token
            if (session->session_token) {
                free(session->session_token);
            }
            
            // Notify callback
            if (manager->device_disconnected_callback) {
                manager->device_disconnected_callback(&session->remote_device, session_id);
            }
            
            // Remove session from array
            memmove(&manager->active_sessions[i], &manager->active_sessions[i + 1],
                   (manager->active_sessions_count - i - 1) * sizeof(FileTransferSession));
            manager->active_sessions_count--;
            
            printf("Disconnected from device (Session ID: %u)\n", session_id);
            break;
        }
    }
    
    pthread_mutex_unlock(&manager->sessions_mutex);
}

uint32_t file_transfer_manager_send_file(FileTransferManager* manager, uint32_t session_id, const char* file_path) {
    if (!manager || !file_path || !manager->is_running) {
        return 0;
    }
    
    // Check if file exists
    if (!file_exists(file_path)) {
        printf("File not found: %s\n", file_path);
        return 0;
    }
    
    // Find session
    pthread_mutex_lock(&manager->sessions_mutex);
    
    FileTransferSession* session = NULL;
    for (int i = 0; i < manager->active_sessions_count; i++) {
        if (manager->active_sessions[i].session_id == session_id) {
            session = &manager->active_sessions[i];
            break;
        }
    }
    
    if (!session || session->status != TRANSFER_STATUS_CONNECTED) {
        pthread_mutex_unlock(&manager->sessions_mutex);
        printf("Invalid session or session not connected\n");
        return 0;
    }
    
    pthread_mutex_unlock(&manager->sessions_mutex);
    
    // Generate transfer ID
    uint32_t transfer_id = generate_transfer_id();
    
    // Create transfer context
    TransferContext* context = malloc(sizeof(TransferContext));
    if (!context) {
        return 0;
    }
    
    context->manager = manager;
    context->session_id = session_id;
    context->transfer_id = transfer_id;
    strncpy(context->file_path, file_path, MAX_PATH_LENGTH - 1);
    context->file_path[MAX_PATH_LENGTH - 1] = '\0';
    
    // Start file send thread
    pthread_t send_thread;
    if (pthread_create(&send_thread, NULL, file_send_thread, context) != 0) {
        free(context);
        return 0;
    }
    
    pthread_detach(send_thread);
    
    printf("Started file transfer: %s (Transfer ID: %u)\n", file_path, transfer_id);
    return transfer_id;
}

// MARK: - Private Function Implementations

static void* server_thread_func(void* arg) {
    FileTransferManager* manager = (FileTransferManager*)arg;
    
    printf("Server thread started\n");
    
    while (!manager->shutdown_requested && !g_shutdown_requested) {
        struct sockaddr_in client_addr;
        socklen_t addr_len = sizeof(client_addr);
        
        int client_fd = accept(manager->listen_socket, (struct sockaddr*)&client_addr, &addr_len);
        
        if (client_fd < 0) {
            if (errno == EINTR || manager->shutdown_requested) {
                break;
            }
            printf("Accept failed: %s\n", strerror(errno));
            continue;
        }
        
        printf("New client connection from %s:%d\n", 
               inet_ntoa(client_addr.sin_addr), ntohs(client_addr.sin_port));
        
        // Create connection context
        ConnectionContext* context = malloc(sizeof(ConnectionContext));
        if (context) {
            context->manager = manager;
            context->socket_fd = client_fd;
            context->client_addr = client_addr;
            context->addr_len = addr_len;
            
            // Start client handler thread
            pthread_t client_thread;
            if (pthread_create(&client_thread, NULL, client_handler_thread, context) == 0) {
                pthread_detach(client_thread);
            } else {
                printf("Failed to create client handler thread\n");
                close(client_fd);
                free(context);
            }
        } else {
            printf("Failed to allocate connection context\n");
            close(client_fd);
        }
    }
    
    printf("Server thread stopped\n");
    return NULL;
}

static void* discovery_thread_func(void* arg) {
    FileTransferManager* manager = (FileTransferManager*)arg;
    
    printf("Discovery thread started\n");
    
    time_t last_broadcast = 0;
    const time_t broadcast_interval = 5; // 5 seconds
    
    while (!manager->shutdown_requested && !g_shutdown_requested) {
        // Send discovery broadcast periodically
        time_t current_time = time(NULL);
        if (manager->discovery_enabled && (current_time - last_broadcast) >= broadcast_interval) {
            send_discovery_broadcast(manager);
            last_broadcast = current_time;
        }
        
        // Listen for discovery messages
        struct sockaddr_in from_addr;
        socklen_t addr_len = sizeof(from_addr);
        char buffer[1024];
        
        struct timeval timeout;
        timeout.tv_sec = 1;
        timeout.tv_usec = 0;
        
        fd_set read_fds;
        FD_ZERO(&read_fds);
        FD_SET(manager->discovery_socket, &read_fds);
        
        int result = select(manager->discovery_socket + 1, &read_fds, NULL, NULL, &timeout);
        
        if (result > 0 && FD_ISSET(manager->discovery_socket, &read_fds)) {
            ssize_t bytes_received = recvfrom(manager->discovery_socket, buffer, sizeof(buffer), 0,
                                            (struct sockaddr*)&from_addr, &addr_len);
            
            if (bytes_received >= sizeof(FileTransferHeader)) {
                FileTransferHeader* header = (FileTransferHeader*)buffer;
                
                if (header->magic == PROTOCOL_MAGIC && header->version == PROTOCOL_VERSION) {
                    void* data = (bytes_received > sizeof(FileTransferHeader)) ? 
                                (buffer + sizeof(FileTransferHeader)) : NULL;
                    
                    if (header->message_type == MSG_TYPE_DISCOVERY_REQUEST) {
                        DiscoveryRequest* request = (DiscoveryRequest*)data;
                        if (request) {
                            handle_discovery_request(manager, request, &from_addr);
                        }
                    } else if (header->message_type == MSG_TYPE_DISCOVERY_RESPONSE) {
                        DiscoveryResponse* response = (DiscoveryResponse*)data;
                        if (response) {
                            handle_discovery_response(manager, response, &from_addr);
                        }
                    }
                }
            }
        }
    }
    
    printf("Discovery thread stopped\n");
    return NULL;
}

static void* client_handler_thread(void* arg) {
    ConnectionContext* context = (ConnectionContext*)arg;
    FileTransferManager* manager = context->manager;
    int client_fd = context->socket_fd;
    
    printf("Client handler thread started for fd %d\n", client_fd);
    
    while (!manager->shutdown_requested && !g_shutdown_requested) {
        FileTransferHeader header;
        void* data = NULL;
        
        if (!receive_message(client_fd, &header, &data)) {
            break;
        }
        
        process_message(manager, client_fd, &header, data);
        
        if (data) {
            free(data);
        }
    }
    
    close(client_fd);
    free(context);
    
    printf("Client handler thread stopped for fd %d\n", client_fd);
    return NULL;
}

static void* file_send_thread(void* arg) {
    TransferContext* context = (TransferContext*)arg;
    FileTransferManager* manager = context->manager;
    
    printf("File send thread started for transfer %u\n", context->transfer_id);
    
    // TODO: Implement file sending logic
    // 1. Open file
    // 2. Get file info
    // 3. Send file info message
    // 4. Send file chunks
    // 5. Handle acknowledgments
    // 6. Update progress
    
    free(context);
    printf("File send thread completed for transfer %u\n", context->transfer_id);
    return NULL;
}

static void* file_receive_thread(void* arg) {
    TransferContext* context = (TransferContext*)arg;
    FileTransferManager* manager = context->manager;
    
    printf("File receive thread started for transfer %u\n", context->transfer_id);
    
    // TODO: Implement file receiving logic
    // 1. Receive file info
    // 2. Create destination file
    // 3. Receive file chunks
    // 4. Send acknowledgments
    // 5. Update progress
    // 6. Verify file integrity
    
    free(context);
    printf("File receive thread completed for transfer %u\n", context->transfer_id);
    return NULL;
}

static bool create_listen_socket(FileTransferManager* manager) {
    manager->listen_socket = socket(AF_INET, SOCK_STREAM, 0);
    if (manager->listen_socket < 0) {
        printf("Failed to create listen socket: %s\n", strerror(errno));
        return false;
    }
    
    // Set socket options
    int opt = 1;
    if (setsockopt(manager->listen_socket, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt)) < 0) {
        printf("Failed to set SO_REUSEADDR: %s\n", strerror(errno));
        close(manager->listen_socket);
        manager->listen_socket = -1;
        return false;
    }
    
    // Bind socket
    struct sockaddr_in server_addr;
    memset(&server_addr, 0, sizeof(server_addr));
    server_addr.sin_family = AF_INET;
    server_addr.sin_addr.s_addr = INADDR_ANY;
    server_addr.sin_port = htons(manager->config.listen_port);
    
    if (bind(manager->listen_socket, (struct sockaddr*)&server_addr, sizeof(server_addr)) < 0) {
        printf("Failed to bind listen socket: %s\n", strerror(errno));
        close(manager->listen_socket);
        manager->listen_socket = -1;
        return false;
    }
    
    // Start listening
    if (listen(manager->listen_socket, 10) < 0) {
        printf("Failed to listen on socket: %s\n", strerror(errno));
        close(manager->listen_socket);
        manager->listen_socket = -1;
        return false;
    }
    
    printf("Listen socket created on port %d\n", manager->config.listen_port);
    return true;
}

static bool create_discovery_socket(FileTransferManager* manager) {
    manager->discovery_socket = socket(AF_INET, SOCK_DGRAM, 0);
    if (manager->discovery_socket < 0) {
        printf("Failed to create discovery socket: %s\n", strerror(errno));
        return false;
    }
    
    // Set socket options
    int opt = 1;
    if (setsockopt(manager->discovery_socket, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt)) < 0) {
        printf("Failed to set SO_REUSEADDR on discovery socket: %s\n", strerror(errno));
        close(manager->discovery_socket);
        manager->discovery_socket = -1;
        return false;
    }
    
    if (setsockopt(manager->discovery_socket, SOL_SOCKET, SO_BROADCAST, &opt, sizeof(opt)) < 0) {
        printf("Failed to set SO_BROADCAST on discovery socket: %s\n", strerror(errno));
        close(manager->discovery_socket);
        manager->discovery_socket = -1;
        return false;
    }
    
    // Bind socket
    struct sockaddr_in discovery_addr;
    memset(&discovery_addr, 0, sizeof(discovery_addr));
    discovery_addr.sin_family = AF_INET;
    discovery_addr.sin_addr.s_addr = INADDR_ANY;
    discovery_addr.sin_port = htons(DISCOVERY_PORT);
    
    if (bind(manager->discovery_socket, (struct sockaddr*)&discovery_addr, sizeof(discovery_addr)) < 0) {
        printf("Failed to bind discovery socket: %s\n", strerror(errno));
        close(manager->discovery_socket);
        manager->discovery_socket = -1;
        return false;
    }
    
    printf("Discovery socket created on port %d\n", DISCOVERY_PORT);
    return true;
}

static void send_discovery_broadcast(FileTransferManager* manager) {
    DiscoveryRequest request;
    memset(&request, 0, sizeof(request));
    
    strncpy(request.device_id, manager->local_device.device_id, MAX_DEVICE_ID_LENGTH - 1);
    strncpy(request.device_name, manager->local_device.device_name, MAX_DEVICE_NAME_LENGTH - 1);
    request.device_type = manager->local_device.device_type;
    request.listen_port = manager->local_device.listen_port;
    request.supports_encryption = manager->local_device.supports_encryption;
    request.max_chunk_size = manager->local_device.max_chunk_size;
    
    FileTransferHeader header;
    memset(&header, 0, sizeof(header));
    header.magic = PROTOCOL_MAGIC;
    header.version = PROTOCOL_VERSION;
    header.message_type = MSG_TYPE_DISCOVERY_REQUEST;
    header.message_id = generate_message_id();
    header.session_id = 0;
    header.data_length = sizeof(request);
    header.checksum = calculate_checksum(&request, sizeof(request));
    header.timestamp = get_current_time_ms();
    
    // Prepare broadcast message
    char buffer[sizeof(FileTransferHeader) + sizeof(DiscoveryRequest)];
    memcpy(buffer, &header, sizeof(header));
    memcpy(buffer + sizeof(header), &request, sizeof(request));
    
    // Send broadcast
    struct sockaddr_in broadcast_addr;
    memset(&broadcast_addr, 0, sizeof(broadcast_addr));
    broadcast_addr.sin_family = AF_INET;
    broadcast_addr.sin_addr.s_addr = INADDR_BROADCAST;
    broadcast_addr.sin_port = htons(DISCOVERY_PORT);
    
    ssize_t bytes_sent = sendto(manager->discovery_socket, buffer, sizeof(buffer), 0,
                               (struct sockaddr*)&broadcast_addr, sizeof(broadcast_addr));
    
    if (bytes_sent < 0) {
        printf("Failed to send discovery broadcast: %s\n", strerror(errno));
    }
}

static void handle_discovery_request(FileTransferManager* manager, DiscoveryRequest* request, struct sockaddr_in* from_addr) {
    // Don't respond to our own requests
    if (strcmp(request->device_id, manager->local_device.device_id) == 0) {
        return;
    }
    
    printf("Received discovery request from %s (%s)\n", request->device_name, request->device_id);
    
    // Send discovery response
    DiscoveryResponse response;
    memset(&response, 0, sizeof(response));
    
    strncpy(response.device_id, manager->local_device.device_id, MAX_DEVICE_ID_LENGTH - 1);
    strncpy(response.device_name, manager->local_device.device_name, MAX_DEVICE_NAME_LENGTH - 1);
    response.device_type = manager->local_device.device_type;
    response.listen_port = manager->local_device.listen_port;
    response.supports_encryption = manager->local_device.supports_encryption;
    response.max_chunk_size = manager->local_device.max_chunk_size;
    response.accepts_connections = true;
    
    FileTransferHeader header;
    memset(&header, 0, sizeof(header));
    header.magic = PROTOCOL_MAGIC;
    header.version = PROTOCOL_VERSION;
    header.message_type = MSG_TYPE_DISCOVERY_RESPONSE;
    header.message_id = generate_message_id();
    header.session_id = 0;
    header.data_length = sizeof(response);
    header.checksum = calculate_checksum(&response, sizeof(response));
    header.timestamp = get_current_time_ms();
    
    // Prepare response message
    char buffer[sizeof(FileTransferHeader) + sizeof(DiscoveryResponse)];
    memcpy(buffer, &header, sizeof(header));
    memcpy(buffer + sizeof(header), &response, sizeof(response));
    
    // Send response
    ssize_t bytes_sent = sendto(manager->discovery_socket, buffer, sizeof(buffer), 0,
                               (struct sockaddr*)from_addr, sizeof(*from_addr));
    
    if (bytes_sent < 0) {
        printf("Failed to send discovery response: %s\n", strerror(errno));
    }
}

static void handle_discovery_response(FileTransferManager* manager, DiscoveryResponse* response, struct sockaddr_in* from_addr) {
    // Don't add our own device
    if (strcmp(response->device_id, manager->local_device.device_id) == 0) {
        return;
    }
    
    printf("Received discovery response from %s (%s)\n", response->device_name, response->device_id);
    
    pthread_mutex_lock(&manager->devices_mutex);
    
    // Check if device already exists
    bool device_exists = false;
    for (int i = 0; i < manager->discovered_devices_count; i++) {
        if (strcmp(manager->discovered_devices[i].device_id, response->device_id) == 0) {
            // Update existing device
            manager->discovered_devices[i].ip_address = ntohl(from_addr->sin_addr.s_addr);
            manager->discovered_devices[i].port = response->listen_port;
            manager->discovered_devices[i].last_seen = get_current_time_ms();
            manager->discovered_devices[i].supports_encryption = response->supports_encryption;
            manager->discovered_devices[i].max_chunk_size = response->max_chunk_size;
            device_exists = true;
            break;
        }
    }
    
    // Add new device if not exists and we have space
    if (!device_exists && manager->discovered_devices_count < MAX_DISCOVERED_DEVICES) {
        DeviceInfo* device = &manager->discovered_devices[manager->discovered_devices_count++];
        memset(device, 0, sizeof(DeviceInfo));
        
        device->device_id = strdup(response->device_id);
        strncpy(device->device_name, response->device_name, MAX_DEVICE_NAME_LENGTH - 1);
        device->device_type = response->device_type;
        device->ip_address = ntohl(from_addr->sin_addr.s_addr);
        device->port = response->listen_port;
        device->last_seen = get_current_time_ms();
        device->is_trusted = false;
        device->supports_encryption = response->supports_encryption;
        device->max_chunk_size = response->max_chunk_size;
        
        // Notify callback
        if (manager->device_discovered_callback) {
            manager->device_discovered_callback(device);
        }
    }
    
    pthread_mutex_unlock(&manager->devices_mutex);
}

static void process_message(FileTransferManager* manager, int client_fd, FileTransferHeader* header, void* data) {
    switch (header->message_type) {
        case MSG_TYPE_CONNECT_REQUEST:
            // TODO: Handle connection request
            break;
        case MSG_TYPE_FILE_REQUEST:
            // TODO: Handle file request
            break;
        case MSG_TYPE_FILE_CHUNK:
            // TODO: Handle file chunk
            break;
        case MSG_TYPE_FILE_ACK:
            // TODO: Handle file acknowledgment
            break;
        case MSG_TYPE_HEARTBEAT:
            // TODO: Handle heartbeat
            break;
        default:
            printf("Unknown message type: %d\n", header->message_type);
            break;
    }
}

static bool send_message(int socket_fd, FileTransferHeader* header, void* data) {
    // Calculate checksum
    if (data && header->data_length > 0) {
        header->checksum = calculate_checksum(data, header->data_length);
    }
    
    // Send header
    ssize_t bytes_sent = send(socket_fd, header, sizeof(FileTransferHeader), 0);
    if (bytes_sent != sizeof(FileTransferHeader)) {
        printf("Failed to send message header: %s\n", strerror(errno));
        return false;
    }
    
    // Send data if present
    if (data && header->data_length > 0) {
        bytes_sent = send(socket_fd, data, header->data_length, 0);
        if (bytes_sent != header->data_length) {
            printf("Failed to send message data: %s\n", strerror(errno));
            return false;
        }
    }
    
    return true;
}

static bool receive_message(int socket_fd, FileTransferHeader* header, void** data) {
    // Receive header
    ssize_t bytes_received = recv(socket_fd, header, sizeof(FileTransferHeader), MSG_WAITALL);
    if (bytes_received != sizeof(FileTransferHeader)) {
        if (bytes_received == 0) {
            printf("Connection closed by peer\n");
        } else {
            printf("Failed to receive message header: %s\n", strerror(errno));
        }
        return false;
    }
    
    // Validate header
    if (header->magic != PROTOCOL_MAGIC || header->version != PROTOCOL_VERSION) {
        printf("Invalid message header\n");
        return false;
    }
    
    // Receive data if present
    *data = NULL;
    if (header->data_length > 0) {
        *data = malloc(header->data_length);
        if (!*data) {
            printf("Failed to allocate memory for message data\n");
            return false;
        }
        
        bytes_received = recv(socket_fd, *data, header->data_length, MSG_WAITALL);
        if (bytes_received != header->data_length) {
            printf("Failed to receive message data: %s\n", strerror(errno));
            free(*data);
            *data = NULL;
            return false;
        }
        
        // Verify checksum
        uint32_t calculated_checksum = calculate_checksum(*data, header->data_length);
        if (calculated_checksum != header->checksum) {
            printf("Message checksum mismatch\n");
            free(*data);
            *data = NULL;
            return false;
        }
    }
    
    return true;
}

// MARK: - Utility Function Implementations

static char* generate_device_id(void) {
    // Try to get MAC address
    struct ifaddrs *ifaddrs_ptr = NULL;
    if (getifaddrs(&ifaddrs_ptr) == 0) {
        for (struct ifaddrs *ifa = ifaddrs_ptr; ifa != NULL; ifa = ifa->ifa_next) {
            if (ifa->ifa_addr && ifa->ifa_addr->sa_family == AF_PACKET) {
                // Found a network interface
                char* device_id = malloc(64);
                if (device_id) {
                    snprintf(device_id, 64, "LINUX_%s", ifa->ifa_name);
                    freeifaddrs(ifaddrs_ptr);
                    return device_id;
                }
            }
        }
        freeifaddrs(ifaddrs_ptr);
    }
    
    // Fallback to hostname
    char hostname[256];
    if (gethostname(hostname, sizeof(hostname)) == 0) {
        char* device_id = malloc(64);
        if (device_id) {
            snprintf(device_id, 64, "LINUX_%s", hostname);
            return device_id;
        }
    }
    
    // Final fallback to random ID
    char* device_id = malloc(64);
    if (device_id) {
        snprintf(device_id, 64, "LINUX_%ld", time(NULL));
    }
    
    return device_id;
}

static char* get_default_device_name(void) {
    char hostname[256];
    if (gethostname(hostname, sizeof(hostname)) == 0) {
        char* device_name = malloc(strlen(hostname) + 20);
        if (device_name) {
            snprintf(device_name, strlen(hostname) + 20, "%s (Linux)", hostname);
            return device_name;
        }
    }
    
    return NULL;
}

static uint32_t generate_session_id(void) {
    return (uint32_t)time(NULL) ^ (uint32_t)getpid();
}

static uint32_t generate_transfer_id(void) {
    static uint32_t counter = 0;
    return (uint32_t)time(NULL) ^ (++counter);
}

static uint32_t generate_message_id(void) {
    static uint32_t counter = 0;
    return ++counter;
}

static uint64_t get_current_time_ms(void) {
    struct timespec ts;
    clock_gettime(CLOCK_REALTIME, &ts);
    return (uint64_t)ts.tv_sec * 1000 + ts.tv_nsec / 1000000;
}

static uint32_t calculate_checksum(const void* data, size_t length) {
    const uint8_t* bytes = (const uint8_t*)data;
    uint32_t checksum = 0;
    
    for (size_t i = 0; i < length; i++) {
        checksum = (checksum << 1) ^ bytes[i];
    }
    
    return checksum;
}

static uint32_t calculate_file_hash(const char* file_path) {
    FILE* file = fopen(file_path, "rb");
    if (!file) {
        return 0;
    }
    
    uint32_t hash = 0;
    uint8_t buffer[4096];
    size_t bytes_read;
    
    while ((bytes_read = fread(buffer, 1, sizeof(buffer), file)) > 0) {
        hash = calculate_checksum(buffer, bytes_read) ^ hash;
    }
    
    fclose(file);
    return hash;
}

static bool file_exists(const char* file_path) {
    struct stat st;
    return stat(file_path, &st) == 0 && S_ISREG(st.st_mode);
}

static uint64_t get_file_size(const char* file_path) {
    struct stat st;
    if (stat(file_path, &st) == 0) {
        return st.st_size;
    }
    return 0;
}

static bool create_directories(const char* path) {
    char* path_copy = strdup(path);
    if (!path_copy) {
        return false;
    }
    
    char* dir = dirname(path_copy);
    
    struct stat st;
    if (stat(dir, &st) != 0) {
        if (mkdir(dir, 0755) != 0 && errno != EEXIST) {
            free(path_copy);
            return false;
        }
    }
    
    free(path_copy);
    return true;
}

// MARK: - Callback Setters

void file_transfer_manager_set_progress_callback(FileTransferManager* manager, FileTransferProgressCallback callback) {
    if (manager) {
        manager->progress_callback = callback;
    }
}

void file_transfer_manager_set_complete_callback(FileTransferManager* manager, FileTransferCompleteCallback callback) {
    if (manager) {
        manager->complete_callback = callback;
    }
}

void file_transfer_manager_set_error_callback(FileTransferManager* manager, FileTransferErrorCallback callback) {
    if (manager) {
        manager->error_callback = callback;
    }
}

void file_transfer_manager_set_device_discovered_callback(FileTransferManager* manager, DeviceDiscoveredCallback callback) {
    if (manager) {
        manager->device_discovered_callback = callback;
    }
}

void file_transfer_manager_set_device_connected_callback(FileTransferManager* manager, DeviceConnectedCallback callback) {
    if (manager) {
        manager->device_connected_callback = callback;
    }
}

void file_transfer_manager_set_device_disconnected_callback(FileTransferManager* manager, DeviceDisconnectedCallback callback) {
    if (manager) {
        manager->device_disconnected_callback = callback;
    }
}

void file_transfer_manager_set_file_receive_request_callback(FileTransferManager* manager, FileReceiveRequestCallback callback) {
    if (manager) {
        manager->file_receive_request_callback = callback;
    }
}