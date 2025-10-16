#pragma once

#include "pch.h"

#ifdef __cplusplus
extern "C" {
#endif

// 桌面宠物状态
typedef enum {
    PET_STATE_IDLE = 0,
    PET_STATE_WALKING = 1,
    PET_STATE_TALKING = 2,
    PET_STATE_THINKING = 3,
    PET_STATE_SLEEPING = 4,
    PET_STATE_PLAYING = 5,
    PET_STATE_WORKING = 6,
    PET_STATE_NOTIFICATION = 7
} PetState;

// 桌面宠物动作
typedef enum {
    PET_ACTION_NONE = 0,
    PET_ACTION_MOVE_LEFT = 1,
    PET_ACTION_MOVE_RIGHT = 2,
    PET_ACTION_MOVE_UP = 3,
    PET_ACTION_MOVE_DOWN = 4,
    PET_ACTION_JUMP = 5,
    PET_ACTION_DANCE = 6,
    PET_ACTION_WAVE = 7,
    PET_ACTION_NOD = 8,
    PET_ACTION_SHAKE_HEAD = 9
} PetAction;

// 桌面宠物情绪
typedef enum {
    PET_MOOD_HAPPY = 0,
    PET_MOOD_EXCITED = 1,
    PET_MOOD_CALM = 2,
    PET_MOOD_TIRED = 3,
    PET_MOOD_BORED = 4,
    PET_MOOD_CURIOUS = 5,
    PET_MOOD_FOCUSED = 6
} PetMood;

// 桌面宠物配置
typedef struct {
    int width;
    int height;
    int animation_speed;
    bool always_on_top;
    bool click_through;
    bool auto_hide;
    int transparency;
    char skin_path[MAX_PATH];
    char voice_pack[64];
} PetConfig;

// 桌面宠物位置
typedef struct {
    int x;
    int y;
    int target_x;
    int target_y;
    bool is_moving;
} PetPosition;

// 桌面宠物动画帧
typedef struct {
    HBITMAP bitmap;
    int duration_ms;
    int offset_x;
    int offset_y;
} PetAnimationFrame;

// 桌面宠物动画
typedef struct {
    PetAnimationFrame* frames;
    int frame_count;
    int current_frame;
    uint64_t last_frame_time;
    bool loop;
    bool playing;
} PetAnimation;

// 桌面宠物语音
typedef struct {
    char text[512];
    char audio_file[MAX_PATH];
    int duration_ms;
    bool is_playing;
} PetVoice;

// 桌面宠物AI响应
typedef struct {
    char response_text[1024];
    PetAction suggested_action;
    PetMood suggested_mood;
    int confidence;
} PetAIResponse;

// 桌面宠物主结构
typedef struct {
    HWND hwnd;
    HDC hdc;
    HBITMAP background_bitmap;
    
    PetConfig config;
    PetPosition position;
    PetState current_state;
    PetMood current_mood;
    PetAction current_action;
    
    // 动画系统
    PetAnimation animations[16]; // 支持多种动画
    int current_animation;
    
    // AI交互
    char last_user_input[512];
    PetAIResponse last_ai_response;
    uint64_t last_interaction_time;
    
    // 语音系统
    PetVoice current_voice;
    
    // 行为系统
    uint64_t last_action_time;
    uint64_t next_random_action_time;
    bool user_interaction_mode;
    
    // 渲染系统
    ID2D1Factory* d2d_factory;
    ID2D1HwndRenderTarget* render_target;
    ID2D1Bitmap* current_frame_bitmap;
    
    // 事件回调
    void (*on_click_callback)(int x, int y, void* user_data);
    void (*on_double_click_callback)(int x, int y, void* user_data);
    void (*on_right_click_callback)(int x, int y, void* user_data);
    void (*on_state_change_callback)(PetState old_state, PetState new_state, void* user_data);
    void* callback_user_data;
    
    // 线程同步
    CRITICAL_SECTION state_lock;
    HANDLE animation_thread;
    HANDLE ai_thread;
    volatile bool should_exit;
    
} DesktopPet;

// 桌面宠物管理器
typedef struct {
    DesktopPet* pets;
    int pet_count;
    int max_pets;
    
    // 全局配置
    bool pets_enabled;
    char skins_directory[MAX_PATH];
    char voices_directory[MAX_PATH];
    
    // AI服务配置
    char ai_service_url[256];
    char ai_api_key[128];
    
    // 窗口管理
    HWND main_window;
    
} DesktopPetManager;

// 函数声明

// 桌面宠物管理器
DesktopPetManager* desktop_pet_manager_create(void);
void desktop_pet_manager_destroy(DesktopPetManager* manager);
bool desktop_pet_manager_initialize(DesktopPetManager* manager, HWND main_window);
void desktop_pet_manager_shutdown(DesktopPetManager* manager);
void desktop_pet_manager_update(DesktopPetManager* manager);

// 桌面宠物创建和销毁
DesktopPet* desktop_pet_create(const PetConfig* config);
void desktop_pet_destroy(DesktopPet* pet);
bool desktop_pet_initialize(DesktopPet* pet, HWND parent_window);
void desktop_pet_shutdown(DesktopPet* pet);

// 桌面宠物控制
bool desktop_pet_show(DesktopPet* pet);
bool desktop_pet_hide(DesktopPet* pet);
bool desktop_pet_set_position(DesktopPet* pet, int x, int y);
bool desktop_pet_move_to(DesktopPet* pet, int x, int y, int duration_ms);
bool desktop_pet_set_state(DesktopPet* pet, PetState state);
bool desktop_pet_set_mood(DesktopPet* pet, PetMood mood);
bool desktop_pet_perform_action(DesktopPet* pet, PetAction action);

// 动画系统
bool desktop_pet_load_animation(DesktopPet* pet, int animation_id, const char* animation_path);
bool desktop_pet_play_animation(DesktopPet* pet, int animation_id, bool loop);
bool desktop_pet_stop_animation(DesktopPet* pet);
void desktop_pet_update_animation(DesktopPet* pet);

// AI交互
bool desktop_pet_process_user_input(DesktopPet* pet, const char* input);
bool desktop_pet_get_ai_response(DesktopPet* pet, const char* input, PetAIResponse* response);
void desktop_pet_apply_ai_response(DesktopPet* pet, const PetAIResponse* response);

// 语音系统
bool desktop_pet_speak(DesktopPet* pet, const char* text);
bool desktop_pet_play_sound(DesktopPet* pet, const char* sound_file);
void desktop_pet_stop_speaking(DesktopPet* pet);

// 行为系统
void desktop_pet_update_behavior(DesktopPet* pet);
void desktop_pet_trigger_random_action(DesktopPet* pet);
bool desktop_pet_is_idle_too_long(DesktopPet* pet);

// 渲染系统
bool desktop_pet_initialize_graphics(DesktopPet* pet);
void desktop_pet_cleanup_graphics(DesktopPet* pet);
void desktop_pet_render(DesktopPet* pet);
bool desktop_pet_load_skin(DesktopPet* pet, const char* skin_path);

// 事件处理
void desktop_pet_on_mouse_click(DesktopPet* pet, int x, int y, bool is_double_click);
void desktop_pet_on_mouse_right_click(DesktopPet* pet, int x, int y);
void desktop_pet_on_mouse_move(DesktopPet* pet, int x, int y);
void desktop_pet_on_key_press(DesktopPet* pet, int key_code);

// 配置管理
bool desktop_pet_load_config(PetConfig* config, const char* config_file);
bool desktop_pet_save_config(const PetConfig* config, const char* config_file);
void desktop_pet_get_default_config(PetConfig* config);

// 皮肤和资源管理
bool desktop_pet_load_skin_pack(DesktopPet* pet, const char* skin_pack_path);
bool desktop_pet_enumerate_skins(const char* skins_directory, char*** skin_names, int* count);
void desktop_pet_free_skin_list(char** skin_names, int count);

// 工具函数
bool desktop_pet_is_point_inside(const DesktopPet* pet, int x, int y);
void desktop_pet_get_screen_bounds(RECT* bounds);
bool desktop_pet_clamp_to_screen(DesktopPet* pet);
uint64_t desktop_pet_get_current_time_ms(void);

// 窗口过程
LRESULT CALLBACK desktop_pet_window_proc(HWND hwnd, UINT msg, WPARAM wParam, LPARAM lParam);

// 线程函数
DWORD WINAPI desktop_pet_animation_thread(LPVOID param);
DWORD WINAPI desktop_pet_ai_thread(LPVOID param);

// 常量定义
#define PET_DEFAULT_WIDTH           200
#define PET_DEFAULT_HEIGHT          200
#define PET_DEFAULT_ANIMATION_SPEED 60
#define PET_DEFAULT_TRANSPARENCY    255
#define PET_MAX_ANIMATION_FRAMES    32
#define PET_IDLE_TIMEOUT_MS         30000
#define PET_RANDOM_ACTION_MIN_MS    10000
#define PET_RANDOM_ACTION_MAX_MS    60000
#define PET_AI_RESPONSE_TIMEOUT_MS  5000
#define PET_VOICE_MAX_DURATION_MS   10000

// 错误代码
#define PET_ERROR_SUCCESS           0
#define PET_ERROR_INVALID_PARAM     1
#define PET_ERROR_MEMORY_ALLOC      2
#define PET_ERROR_WINDOW_CREATE     3
#define PET_ERROR_GRAPHICS_INIT     4
#define PET_ERROR_ANIMATION_LOAD    5
#define PET_ERROR_AI_SERVICE        6
#define PET_ERROR_VOICE_SYSTEM      7

#ifdef __cplusplus
}
#endif