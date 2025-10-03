#pragma once

// Windows头文件
#define WIN32_LEAN_AND_MEAN
#define NOMINMAX
#include <windows.h>
#include <windowsx.h>
#include <commctrl.h>
#include <shellapi.h>
#include <shlobj.h>
#include <shlwapi.h>
#include <wininet.h>
#include <winsock2.h>
#include <ws2tcpip.h>

// DirectX头文件
#include <d3d11.h>
#include <d2d1.h>
#include <dwrite.h>
#include <dxgi.h>

// 加密相关
#include <bcrypt.h>
#include <wincrypt.h>

// C标准库
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <stdbool.h>
#include <time.h>
#include <math.h>

// C++标准库
#include <iostream>
#include <string>
#include <vector>
#include <map>
#include <unordered_map>
#include <set>
#include <queue>
#include <thread>
#include <mutex>
#include <condition_variable>
#include <atomic>
#include <memory>
#include <algorithm>
#include <functional>
#include <chrono>
#include <fstream>
#include <sstream>
#include <regex>
#include <future>
#include <exception>

// JSON库（使用nlohmann/json或类似库）
// #include <nlohmann/json.hpp>

// 共享头文件
#include "utils.h"
#include "communication.h"

// 应用程序特定头文件
#include "application.h"
#include "window_manager.h"
#include "desktop_pet.h"

// 常用宏定义
#define SAFE_DELETE(p) do { if(p) { delete (p); (p) = nullptr; } } while(0)
#define SAFE_DELETE_ARRAY(p) do { if(p) { delete[] (p); (p) = nullptr; } } while(0)
#define SAFE_RELEASE(p) do { if(p) { (p)->Release(); (p) = nullptr; } } while(0)

// 错误处理宏
#define CHECK_HR(hr) do { if(FAILED(hr)) { LOG_ERROR("HRESULT failed: 0x%08X", hr); return false; } } while(0)
#define CHECK_WIN32(result) do { if(!(result)) { LOG_ERROR("Win32 API failed: %d", GetLastError()); return false; } } while(0)

// 字符串转换宏
#define WSTR_TO_STR(wstr) std::string(wstr.begin(), wstr.end())
#define STR_TO_WSTR(str) std::wstring(str.begin(), str.end())

// 应用程序常量
#define APP_NAME L"太上老君AI平台"
#define APP_CLASS_NAME L"TaishangLaojunDesktopApp"
#define APP_WINDOW_TITLE L"太上老君AI平台 - 桌面版"
#define APP_VERSION L"1.0.0"

// 窗口消息定义
#define WM_TRAY_ICON (WM_USER + 1)
#define WM_PET_UPDATE (WM_USER + 2)
#define WM_FILE_TRANSFER (WM_USER + 3)
#define WM_DATA_SYNC (WM_USER + 4)
#define WM_NOTIFICATION (WM_USER + 5)

// 配置常量
#define CONFIG_FILE_NAME L"config.ini"
#define LOG_FILE_NAME L"app.log"
#define DATABASE_FILE_NAME L"data.db"
#define CACHE_DIR_NAME L"cache"
#define TEMP_DIR_NAME L"temp"

// 网络配置
#define DEFAULT_SERVER_HOST "api.taishanglaojun.com"
#define DEFAULT_SERVER_PORT 443
#define DEFAULT_WEBSOCKET_PORT 8080
#define CONNECTION_TIMEOUT_MS 30000
#define HEARTBEAT_INTERVAL_MS 30000

// UI配置
#define MAIN_WINDOW_WIDTH 1200
#define MAIN_WINDOW_HEIGHT 800
#define PET_WINDOW_WIDTH 200
#define PET_WINDOW_HEIGHT 200
#define ANIMATION_FRAME_RATE 60

// 文件传输配置
#define MAX_FILE_SIZE (1024 * 1024 * 1024)  // 1GB
#define FILE_CHUNK_SIZE (64 * 1024)         // 64KB
#define MAX_CONCURRENT_TRANSFERS 5

// 数据同步配置
#define SYNC_INTERVAL_MS 5000
#define MAX_SYNC_RETRIES 3
#define SYNC_BATCH_SIZE 100