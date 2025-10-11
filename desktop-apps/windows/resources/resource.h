#pragma once

// 图标资源ID
#define IDI_MAIN_ICON                   100
#define IDI_TRAY_ICON                   101

// 菜单资源ID
#define IDR_MAIN_MENU                   200
#define IDR_MAIN_ACCELERATOR            201

// 对话框资源ID
#define IDD_ABOUT                       300
#define IDD_SETTINGS                    301
#define IDD_PET_CONFIG                  302
#define IDD_TRANSFER_PROGRESS           303

// 控件ID
#define IDC_STATIC                      -1
#define IDC_CHECK_AUTOSTART             1001
#define IDC_CHECK_MINIMIZE_TO_TRAY      1002
#define IDC_CHECK_CLOSE_TO_TRAY         1003
#define IDC_CHECK_ENABLE_PET            1004
#define IDC_COMBO_PET_SKIN              1005
#define IDC_EDIT_SERVER_URL             1006
#define IDC_EDIT_SERVER_PORT            1007
#define IDC_PROGRESS_TRANSFER           1008
#define IDC_LIST_DEVICES                1009
#define IDC_BUTTON_REFRESH              1010
#define IDC_BUTTON_CONNECT              1011
#define IDC_BUTTON_DISCONNECT           1012

// 菜单命令ID
#define ID_FILE_NEW                     2001
#define ID_FILE_OPEN                    2002
#define ID_FILE_SAVE                    2003
#define ID_FILE_EXIT                    2004
#define ID_TOOLS_PET                    2101
#define ID_TOOLS_TRANSFER               2102
#define ID_TOOLS_SYNC                   2103
#define ID_TOOLS_SETTINGS               2104
#define ID_HELP_ABOUT                   2201

// 托盘菜单ID
#define ID_TRAY_SHOW                    3001
#define ID_TRAY_HIDE                    3002
#define ID_TRAY_PET                     3003
#define ID_TRAY_TRANSFER                3004
#define ID_TRAY_SYNC                    3005
#define ID_TRAY_SETTINGS                3006
#define ID_TRAY_ABOUT                   3007
#define ID_TRAY_EXIT                    3008

// 字符串资源ID
#define IDS_APP_TITLE                   4001
#define IDS_TRAY_TOOLTIP                4002
#define IDS_CONFIRM_EXIT                4003
#define IDS_PET_GREETING                4004
#define IDS_TRANSFER_COMPLETE           4005
#define IDS_SYNC_SUCCESS                4006
#define IDS_CONNECTION_ERROR            4007
#define IDS_FILE_NOT_FOUND              4008
#define IDS_PERMISSION_DENIED           4009
#define IDS_NETWORK_ERROR               4010

// 自定义消息
#define WM_TRAY_ICON                    (WM_USER + 1)
#define WM_PET_ACTION                   (WM_USER + 2)
#define WM_TRANSFER_UPDATE              (WM_USER + 3)
#define WM_SYNC_COMPLETE                (WM_USER + 4)

// 定时器ID
#define TIMER_PET_ANIMATION             5001
#define TIMER_PET_BEHAVIOR              5002
#define TIMER_SYNC_CHECK                5003
#define TIMER_HEARTBEAT                 5004

// 常量定义
#define MAX_PET_SKINS                   20
#define MAX_TRANSFER_SESSIONS           10
#define MAX_DEVICE_NAME_LENGTH          64
#define MAX_FILE_PATH_LENGTH            512
#define DEFAULT_SERVER_PORT             8080
#define TRAY_ICON_ID                    1