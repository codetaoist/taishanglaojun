#include "pch.h"
#include "desktop_pet.h"
#include "application.h"
#include "network.h"
#include "utils.h"
#include "http_client.h"
#include <json/json.h>
#include <sstream>
#include <sapi.h>
#include <sphelper.h>

// 全局桌面宠物管理器实例
static DesktopPetManager* g_pet_manager = nullptr;

// 桌面宠物窗口类名
static const wchar_t* PET_WINDOW_CLASS = L"TaishanglaojunDesktopPet";

// 桌面宠物管理器实现
DesktopPetManager* desktop_pet_manager_create(void) {
    DesktopPetManager* manager = (DesktopPetManager*)calloc(1, sizeof(DesktopPetManager));
    if (!manager) {
        return nullptr;
    }
    
    manager->max_pets = 4; // 最多支持4个桌面宠物
    manager->pets = (DesktopPet*)calloc(manager->max_pets, sizeof(DesktopPet));
    if (!manager->pets) {
        free(manager);
        return nullptr;
    }
    
    manager->pets_enabled = true;
    wcscpy_s((wchar_t*)manager->skins_directory, MAX_PATH, L"assets\\skins");
    wcscpy_s((wchar_t*)manager->voices_directory, MAX_PATH, L"assets\\voices");
    strcpy_s(manager->ai_service_url, sizeof(manager->ai_service_url), "http://localhost:8080/api/v1/ai/chat");
    
    return manager;
}

void desktop_pet_manager_destroy(DesktopPetManager* manager) {
    if (!manager) return;
    
    desktop_pet_manager_shutdown(manager);
    
    if (manager->pets) {
        free(manager->pets);
    }
    free(manager);
}

bool desktop_pet_manager_initialize(DesktopPetManager* manager, HWND main_window) {
    if (!manager) return false;
    
    manager->main_window = main_window;
    
    // 注册桌面宠物窗口类
    WNDCLASSEXW wc = {0};
    wc.cbSize = sizeof(WNDCLASSEXW);
    wc.style = CS_HREDRAW | CS_VREDRAW;
    wc.lpfnWndProc = desktop_pet_window_proc;
    wc.hInstance = GetModuleHandle(nullptr);
    wc.hCursor = LoadCursor(nullptr, IDC_ARROW);
    wc.hbrBackground = nullptr; // 透明背景
    wc.lpszClassName = PET_WINDOW_CLASS;
    
    if (!RegisterClassExW(&wc)) {
        return false;
    }
    
    // 创建默认桌面宠物
    PetConfig default_config;
    desktop_pet_get_default_config(&default_config);
    
    DesktopPet* pet = desktop_pet_create(&default_config);
    if (pet && desktop_pet_initialize(pet, main_window)) {
        manager->pets[0] = *pet;
        manager->pet_count = 1;
        free(pet);
        
        // 显示桌面宠物
        desktop_pet_show(&manager->pets[0]);
    }
    
    g_pet_manager = manager;
    return true;
}

void desktop_pet_manager_shutdown(DesktopPetManager* manager) {
    if (!manager) return;
    
    // 关闭所有桌面宠物
    for (int i = 0; i < manager->pet_count; i++) {
        desktop_pet_shutdown(&manager->pets[i]);
    }
    manager->pet_count = 0;
    
    // 注销窗口类
    UnregisterClassW(PET_WINDOW_CLASS, GetModuleHandle(nullptr));
    
    g_pet_manager = nullptr;
}

void desktop_pet_manager_update(DesktopPetManager* manager) {
    if (!manager || !manager->pets_enabled) return;
    
    for (int i = 0; i < manager->pet_count; i++) {
        DesktopPet* pet = &manager->pets[i];
        desktop_pet_update_animation(pet);
        desktop_pet_update_behavior(pet);
    }
}

// 桌面宠物实现
DesktopPet* desktop_pet_create(const PetConfig* config) {
    DesktopPet* pet = (DesktopPet*)calloc(1, sizeof(DesktopPet));
    if (!pet) return nullptr;
    
    // 复制配置
    if (config) {
        pet->config = *config;
    } else {
        desktop_pet_get_default_config(&pet->config);
    }
    
    // 初始化状态
    pet->current_state = PET_STATE_IDLE;
    pet->current_mood = PET_MOOD_CALM;
    pet->current_action = PET_ACTION_NONE;
    
    // 初始化位置（屏幕右下角）
    RECT screen_bounds;
    desktop_pet_get_screen_bounds(&screen_bounds);
    pet->position.x = screen_bounds.right - pet->config.width - 50;
    pet->position.y = screen_bounds.bottom - pet->config.height - 100;
    pet->position.target_x = pet->position.x;
    pet->position.target_y = pet->position.y;
    pet->position.is_moving = false;
    
    // 初始化时间
    pet->last_interaction_time = desktop_pet_get_current_time_ms();
    pet->last_action_time = pet->last_interaction_time;
    pet->next_random_action_time = pet->last_interaction_time + 
        PET_RANDOM_ACTION_MIN_MS + (rand() % (PET_RANDOM_ACTION_MAX_MS - PET_RANDOM_ACTION_MIN_MS));
    
    // 初始化同步对象
    InitializeCriticalSection(&pet->state_lock);
    
    return pet;
}

void desktop_pet_destroy(DesktopPet* pet) {
    if (!pet) return;
    
    desktop_pet_shutdown(pet);
    DeleteCriticalSection(&pet->state_lock);
    free(pet);
}

bool desktop_pet_initialize(DesktopPet* pet, HWND parent_window) {
    if (!pet) return false;
    
    // 创建桌面宠物窗口
    pet->hwnd = CreateWindowExW(
        WS_EX_LAYERED | WS_EX_TRANSPARENT | WS_EX_TOPMOST | WS_EX_TOOLWINDOW,
        PET_WINDOW_CLASS,
        L"Desktop Pet",
        WS_POPUP,
        pet->position.x, pet->position.y,
        pet->config.width, pet->config.height,
        nullptr, nullptr,
        GetModuleHandle(nullptr),
        pet
    );
    
    if (!pet->hwnd) {
        return false;
    }
    
    // 设置窗口透明度
    SetLayeredWindowAttributes(pet->hwnd, RGB(255, 0, 255), pet->config.transparency, LWA_COLORKEY | LWA_ALPHA);
    
    // 初始化图形系统
    if (!desktop_pet_initialize_graphics(pet)) {
        DestroyWindow(pet->hwnd);
        pet->hwnd = nullptr;
        return false;
    }
    
    // 加载默认皮肤
    char default_skin_path[MAX_PATH];
    sprintf_s(default_skin_path, sizeof(default_skin_path), "%s\\default\\idle.png", pet->config.skin_path);
    desktop_pet_load_skin(pet, default_skin_path);
    
    // 创建动画线程
    pet->should_exit = false;
    pet->animation_thread = CreateThread(nullptr, 0, desktop_pet_animation_thread, pet, 0, nullptr);
    pet->ai_thread = CreateThread(nullptr, 0, desktop_pet_ai_thread, pet, 0, nullptr);
    
    return true;
}

void desktop_pet_shutdown(DesktopPet* pet) {
    if (!pet) return;
    
    // 停止线程
    pet->should_exit = true;
    if (pet->animation_thread) {
        WaitForSingleObject(pet->animation_thread, 3000);
        CloseHandle(pet->animation_thread);
        pet->animation_thread = nullptr;
    }
    if (pet->ai_thread) {
        WaitForSingleObject(pet->ai_thread, 3000);
        CloseHandle(pet->ai_thread);
        pet->ai_thread = nullptr;
    }
    
    // 清理图形资源
    desktop_pet_cleanup_graphics(pet);
    
    // 销毁窗口
    if (pet->hwnd) {
        DestroyWindow(pet->hwnd);
        pet->hwnd = nullptr;
    }
}

bool desktop_pet_show(DesktopPet* pet) {
    if (!pet || !pet->hwnd) return false;
    
    ShowWindow(pet->hwnd, SW_SHOW);
    UpdateWindow(pet->hwnd);
    return true;
}

bool desktop_pet_hide(DesktopPet* pet) {
    if (!pet || !pet->hwnd) return false;
    
    ShowWindow(pet->hwnd, SW_HIDE);
    return true;
}

bool desktop_pet_set_position(DesktopPet* pet, int x, int y) {
    if (!pet) return false;
    
    EnterCriticalSection(&pet->state_lock);
    pet->position.x = x;
    pet->position.y = y;
    pet->position.target_x = x;
    pet->position.target_y = y;
    pet->position.is_moving = false;
    LeaveCriticalSection(&pet->state_lock);
    
    if (pet->hwnd) {
        SetWindowPos(pet->hwnd, nullptr, x, y, 0, 0, SWP_NOSIZE | SWP_NOZORDER);
    }
    
    return true;
}

bool desktop_pet_move_to(DesktopPet* pet, int x, int y, int duration_ms) {
    if (!pet) return false;
    
    EnterCriticalSection(&pet->state_lock);
    pet->position.target_x = x;
    pet->position.target_y = y;
    pet->position.is_moving = true;
    LeaveCriticalSection(&pet->state_lock);
    
    // 设置行走状态
    desktop_pet_set_state(pet, PET_STATE_WALKING);
    
    return true;
}

bool desktop_pet_set_state(DesktopPet* pet, PetState state) {
    if (!pet) return false;
    
    EnterCriticalSection(&pet->state_lock);
    PetState old_state = pet->current_state;
    pet->current_state = state;
    LeaveCriticalSection(&pet->state_lock);
    
    // 触发状态变化回调
    if (pet->on_state_change_callback && old_state != state) {
        pet->on_state_change_callback(old_state, state, pet->callback_user_data);
    }
    
    // 根据状态播放相应动画
    switch (state) {
        case PET_STATE_IDLE:
            desktop_pet_play_animation(pet, 0, true);
            break;
        case PET_STATE_WALKING:
            desktop_pet_play_animation(pet, 1, true);
            break;
        case PET_STATE_TALKING:
            desktop_pet_play_animation(pet, 2, false);
            break;
        case PET_STATE_THINKING:
            desktop_pet_play_animation(pet, 3, true);
            break;
        case PET_STATE_SLEEPING:
            desktop_pet_play_animation(pet, 4, true);
            break;
        case PET_STATE_PLAYING:
            desktop_pet_play_animation(pet, 5, false);
            break;
        case PET_STATE_WORKING:
            desktop_pet_play_animation(pet, 6, true);
            break;
        case PET_STATE_NOTIFICATION:
            desktop_pet_play_animation(pet, 7, false);
            break;
    }
    
    return true;
}

bool desktop_pet_set_mood(DesktopPet* pet, PetMood mood) {
    if (!pet) return false;
    
    EnterCriticalSection(&pet->state_lock);
    pet->current_mood = mood;
    LeaveCriticalSection(&pet->state_lock);
    
    return true;
}

bool desktop_pet_perform_action(DesktopPet* pet, PetAction action) {
    if (!pet) return false;
    
    EnterCriticalSection(&pet->state_lock);
    pet->current_action = action;
    pet->last_action_time = desktop_pet_get_current_time_ms();
    LeaveCriticalSection(&pet->state_lock);
    
    // 执行具体动作
    switch (action) {
        case PET_ACTION_MOVE_LEFT:
            desktop_pet_move_to(pet, pet->position.x - 100, pet->position.y, 2000);
            break;
        case PET_ACTION_MOVE_RIGHT:
            desktop_pet_move_to(pet, pet->position.x + 100, pet->position.y, 2000);
            break;
        case PET_ACTION_MOVE_UP:
            desktop_pet_move_to(pet, pet->position.x, pet->position.y - 50, 1500);
            break;
        case PET_ACTION_MOVE_DOWN:
            desktop_pet_move_to(pet, pet->position.x, pet->position.y + 50, 1500);
            break;
        case PET_ACTION_JUMP:
            desktop_pet_set_state(pet, PET_STATE_PLAYING);
            break;
        case PET_ACTION_DANCE:
            desktop_pet_set_state(pet, PET_STATE_PLAYING);
            break;
        case PET_ACTION_WAVE:
        case PET_ACTION_NOD:
        case PET_ACTION_SHAKE_HEAD:
            desktop_pet_set_state(pet, PET_STATE_TALKING);
            break;
    }
    
    return true;
}

// 动画系统实现
bool desktop_pet_load_animation(DesktopPet* pet, int animation_id, const char* animation_path) {
    if (!pet || animation_id < 0 || animation_id >= 16) return false;
    
    // TODO: 实现动画加载逻辑
    // 这里应该加载动画文件（PNG序列或GIF）并创建动画帧
    
    return true;
}

bool desktop_pet_play_animation(DesktopPet* pet, int animation_id, bool loop) {
    if (!pet || animation_id < 0 || animation_id >= 16) return false;
    
    EnterCriticalSection(&pet->state_lock);
    pet->current_animation = animation_id;
    pet->animations[animation_id].current_frame = 0;
    pet->animations[animation_id].loop = loop;
    pet->animations[animation_id].playing = true;
    pet->animations[animation_id].last_frame_time = desktop_pet_get_current_time_ms();
    LeaveCriticalSection(&pet->state_lock);
    
    return true;
}

bool desktop_pet_stop_animation(DesktopPet* pet) {
    if (!pet) return false;
    
    EnterCriticalSection(&pet->state_lock);
    if (pet->current_animation >= 0 && pet->current_animation < 16) {
        pet->animations[pet->current_animation].playing = false;
    }
    LeaveCriticalSection(&pet->state_lock);
    
    return true;
}

void desktop_pet_update_animation(DesktopPet* pet) {
    if (!pet) return;
    
    EnterCriticalSection(&pet->state_lock);
    
    if (pet->current_animation >= 0 && pet->current_animation < 16) {
        PetAnimation* anim = &pet->animations[pet->current_animation];
        
        if (anim->playing && anim->frame_count > 0) {
            uint64_t current_time = desktop_pet_get_current_time_ms();
            
            if (current_time - anim->last_frame_time >= anim->frames[anim->current_frame].duration_ms) {
                anim->current_frame++;
                
                if (anim->current_frame >= anim->frame_count) {
                    if (anim->loop) {
                        anim->current_frame = 0;
                    } else {
                        anim->playing = false;
                        anim->current_frame = anim->frame_count - 1;
                    }
                }
                
                anim->last_frame_time = current_time;
            }
        }
    }
    
    LeaveCriticalSection(&pet->state_lock);
    
    // 更新位置动画
    if (pet->position.is_moving) {
        int dx = pet->position.target_x - pet->position.x;
        int dy = pet->position.target_y - pet->position.y;
        
        if (abs(dx) <= 2 && abs(dy) <= 2) {
            // 到达目标位置
            pet->position.x = pet->position.target_x;
            pet->position.y = pet->position.target_y;
            pet->position.is_moving = false;
            desktop_pet_set_state(pet, PET_STATE_IDLE);
        } else {
            // 继续移动
            pet->position.x += (dx > 0) ? 2 : (dx < 0) ? -2 : 0;
            pet->position.y += (dy > 0) ? 2 : (dy < 0) ? -2 : 0;
            
            if (pet->hwnd) {
                SetWindowPos(pet->hwnd, nullptr, pet->position.x, pet->position.y, 0, 0, SWP_NOSIZE | SWP_NOZORDER);
            }
        }
    }
}

// AI交互实现
bool desktop_pet_process_user_input(DesktopPet* pet, const char* input) {
    if (!pet || !input) return false;
    
    strcpy_s(pet->last_user_input, sizeof(pet->last_user_input), input);
    pet->last_interaction_time = desktop_pet_get_current_time_ms();
    pet->user_interaction_mode = true;
    
    // 设置思考状态
    desktop_pet_set_state(pet, PET_STATE_THINKING);
    
    return true;
}

bool desktop_pet_get_ai_response(DesktopPet* pet, const char* input, PetAIResponse* response) {
    if (!pet || !input || !response) return false;
    
    // 获取全局宠物管理器
    if (!g_pet_manager || strlen(g_pet_manager->ai_service_url) == 0) {
        // 如果没有配置AI服务URL，使用默认响应
        strcpy_s(response->response_text, sizeof(response->response_text), "我明白了！让我来帮助你。");
        response->suggested_action = PET_ACTION_NOD;
        response->suggested_mood = PET_MOOD_HAPPY;
        response->confidence = 85;
        return true;
    }
    
    try {
        // 使用全局HTTP客户端
        if (!TaishangLaojun::g_httpClient) {
            return false;
        }
        
        // 构建AI服务请求
        Json::Value requestJson;
        requestJson["message"] = input;
        requestJson["context"] = "desktop_pet";
        requestJson["user_id"] = ""; // TODO: 从认证管理器获取用户ID
        
        Json::StreamWriterBuilder builder;
        std::string requestBody = Json::writeString(builder, requestJson);
        
        // 设置请求头
        std::map<std::string, std::string> headers;
        headers["Content-Type"] = "application/json";
        
        // 添加认证头（如果有）
        if (strlen(g_pet_manager->ai_api_key) > 0) {
            headers["Authorization"] = "Bearer " + std::string(g_pet_manager->ai_api_key);
        }
        
        // 发送POST请求到AI服务
        TaishangLaojun::HttpResponse httpResponse = TaishangLaojun::g_httpClient->post(
            g_pet_manager->ai_service_url, requestBody, headers);
        
        if (httpResponse.success && httpResponse.status_code == 200) {
            // 解析AI服务响应
            Json::Value responseJson;
            Json::CharReaderBuilder readerBuilder;
            std::string errors;
            std::istringstream responseStream(httpResponse.body);
            
            if (Json::parseFromStream(readerBuilder, responseStream, &responseJson, &errors)) {
                // 提取响应文本
                if (responseJson.isMember("response") && responseJson["response"].isString()) {
                    std::string responseText = responseJson["response"].asString();
                    strncpy_s(response->response_text, sizeof(response->response_text), 
                             responseText.c_str(), _TRUNCATE);
                }
                
                // 提取建议动作
                if (responseJson.isMember("suggested_action") && responseJson["suggested_action"].isString()) {
                    std::string action = responseJson["suggested_action"].asString();
                    if (action == "nod") response->suggested_action = PET_ACTION_NOD;
                    else if (action == "wave") response->suggested_action = PET_ACTION_WAVE;
                    else if (action == "jump") response->suggested_action = PET_ACTION_JUMP;
                    else if (action == "dance") response->suggested_action = PET_ACTION_DANCE;
                    else if (action == "shake_head") response->suggested_action = PET_ACTION_SHAKE_HEAD;
                    else response->suggested_action = PET_ACTION_NONE;
                } else {
                    response->suggested_action = PET_ACTION_NOD; // 默认动作
                }
                
                // 提取建议情绪
                if (responseJson.isMember("suggested_mood") && responseJson["suggested_mood"].isString()) {
                    std::string mood = responseJson["suggested_mood"].asString();
                    if (mood == "happy") response->suggested_mood = PET_MOOD_HAPPY;
                    else if (mood == "excited") response->suggested_mood = PET_MOOD_EXCITED;
                    else if (mood == "calm") response->suggested_mood = PET_MOOD_CALM;
                    else if (mood == "curious") response->suggested_mood = PET_MOOD_CURIOUS;
                    else if (mood == "focused") response->suggested_mood = PET_MOOD_FOCUSED;
                    else response->suggested_mood = PET_MOOD_HAPPY;
                } else {
                    response->suggested_mood = PET_MOOD_HAPPY; // 默认情绪
                }
                
                // 提取置信度
                if (responseJson.isMember("confidence") && responseJson["confidence"].isNumeric()) {
                    response->confidence = responseJson["confidence"].asInt();
                } else {
                    response->confidence = 85; // 默认置信度
                }
                
                return true;
            }
        }
        
        // 如果AI服务调用失败，使用默认响应
        strcpy_s(response->response_text, sizeof(response->response_text), "抱歉，我现在无法理解你的话。");
        response->suggested_action = PET_ACTION_SHAKE_HEAD;
        response->suggested_mood = PET_MOOD_CALM;
        response->confidence = 50;
        
    } catch (const std::exception& e) {
        // 异常处理，使用默认响应
        strcpy_s(response->response_text, sizeof(response->response_text), "出现了一些问题，请稍后再试。");
        response->suggested_action = PET_ACTION_SHAKE_HEAD;
        response->suggested_mood = PET_MOOD_CALM;
        response->confidence = 30;
    }
    
    return true;
}

void desktop_pet_apply_ai_response(DesktopPet* pet, const PetAIResponse* response) {
    if (!pet || !response) return;
    
    pet->last_ai_response = *response;
    
    // 应用AI建议的动作和情绪
    desktop_pet_set_mood(pet, response->suggested_mood);
    desktop_pet_perform_action(pet, response->suggested_action);
    
    // 播放语音响应
    desktop_pet_speak(pet, response->response_text);
    
    // 设置说话状态
    desktop_pet_set_state(pet, PET_STATE_TALKING);
}

// 语音系统实现
bool desktop_pet_speak(DesktopPet* pet, const char* text) {
    if (!pet || !text) return false;
    
    strcpy_s(pet->current_voice.text, sizeof(pet->current_voice.text), text);
    pet->current_voice.is_playing = true;
    
    // 使用Windows SAPI进行TTS语音合成
    try {
        // 初始化COM
        static bool com_initialized = false;
        if (!com_initialized) {
            HRESULT hr = CoInitialize(nullptr);
            if (FAILED(hr)) {
                return false;
            }
            com_initialized = true;
        }
        
        // 创建SAPI语音对象
        ISpVoice* pVoice = nullptr;
        HRESULT hr = CoCreateInstance(CLSID_SpVoice, nullptr, CLSCTX_ALL, IID_ISpVoice, (void**)&pVoice);
        
        if (SUCCEEDED(hr) && pVoice) {
            // 转换为宽字符
            wchar_t wide_text[512];
            MultiByteToWideChar(CP_UTF8, 0, text, -1, wide_text, 512);
            
            // 异步播放语音
            hr = pVoice->Speak(wide_text, SPF_ASYNC | SPF_IS_NOT_XML, nullptr);
            
            // 释放语音对象
            pVoice->Release();
            
            if (SUCCEEDED(hr)) {
                return true;
            }
        }
    } catch (...) {
        // 如果SAPI失败，使用系统默认的MessageBeep作为备选
        MessageBeep(MB_OK);
    }
    
    return false;
}

bool desktop_pet_play_sound(DesktopPet* pet, const char* sound_file) {
    if (!pet || !sound_file) return false;
    
    // 使用Windows多媒体API播放音频文件
    wchar_t wide_path[MAX_PATH];
    MultiByteToWideChar(CP_UTF8, 0, sound_file, -1, wide_path, MAX_PATH);
    
    return PlaySoundW(wide_path, nullptr, SND_FILENAME | SND_ASYNC) != FALSE;
}

void desktop_pet_stop_speaking(DesktopPet* pet) {
    if (!pet) return;
    
    pet->current_voice.is_playing = false;
    PlaySoundW(nullptr, nullptr, SND_PURGE);
}

// 行为系统实现
void desktop_pet_update_behavior(DesktopPet* pet) {
    if (!pet) return;
    
    uint64_t current_time = desktop_pet_get_current_time_ms();
    
    // 检查是否需要触发随机动作
    if (!pet->user_interaction_mode && current_time >= pet->next_random_action_time) {
        desktop_pet_trigger_random_action(pet);
        pet->next_random_action_time = current_time + 
            PET_RANDOM_ACTION_MIN_MS + (rand() % (PET_RANDOM_ACTION_MAX_MS - PET_RANDOM_ACTION_MIN_MS));
    }
    
    // 检查是否空闲太久
    if (desktop_pet_is_idle_too_long(pet)) {
        if (pet->current_state != PET_STATE_SLEEPING) {
            desktop_pet_set_state(pet, PET_STATE_SLEEPING);
        }
    }
    
    // 重置用户交互模式
    if (pet->user_interaction_mode && current_time - pet->last_interaction_time > 10000) {
        pet->user_interaction_mode = false;
    }
}

void desktop_pet_trigger_random_action(DesktopPet* pet) {
    if (!pet) return;
    
    PetAction actions[] = {
        PET_ACTION_MOVE_LEFT, PET_ACTION_MOVE_RIGHT,
        PET_ACTION_JUMP, PET_ACTION_WAVE, PET_ACTION_DANCE
    };
    
    int action_count = sizeof(actions) / sizeof(actions[0]);
    PetAction random_action = actions[rand() % action_count];
    
    desktop_pet_perform_action(pet, random_action);
}

bool desktop_pet_is_idle_too_long(DesktopPet* pet) {
    if (!pet) return false;
    
    uint64_t current_time = desktop_pet_get_current_time_ms();
    return (current_time - pet->last_interaction_time) > PET_IDLE_TIMEOUT_MS;
}

// 渲染系统实现
bool desktop_pet_initialize_graphics(DesktopPet* pet) {
    if (!pet || !pet->hwnd) return false;
    
    // 创建Direct2D工厂
    HRESULT hr = D2D1CreateFactory(D2D1_FACTORY_TYPE_SINGLE_THREADED, &pet->d2d_factory);
    if (FAILED(hr)) return false;
    
    // 获取窗口客户区大小
    RECT rc;
    GetClientRect(pet->hwnd, &rc);
    
    // 创建渲染目标
    D2D1_SIZE_U size = D2D1::SizeU(rc.right - rc.left, rc.bottom - rc.top);
    hr = pet->d2d_factory->CreateHwndRenderTarget(
        D2D1::RenderTargetProperties(),
        D2D1::HwndRenderTargetProperties(pet->hwnd, size),
        &pet->render_target
    );
    
    return SUCCEEDED(hr);
}

void desktop_pet_cleanup_graphics(DesktopPet* pet) {
    if (!pet) return;
    
    if (pet->current_frame_bitmap) {
        pet->current_frame_bitmap->Release();
        pet->current_frame_bitmap = nullptr;
    }
    
    if (pet->render_target) {
        pet->render_target->Release();
        pet->render_target = nullptr;
    }
    
    if (pet->d2d_factory) {
        pet->d2d_factory->Release();
        pet->d2d_factory = nullptr;
    }
}

void desktop_pet_render(DesktopPet* pet) {
    if (!pet || !pet->render_target) return;
    
    pet->render_target->BeginDraw();
    pet->render_target->Clear(D2D1::ColorF(D2D1::ColorF::White, 0.0f)); // 透明背景
    
    // 渲染当前动画帧
    if (pet->current_frame_bitmap) {
        D2D1_SIZE_F size = pet->render_target->GetSize();
        D2D1_RECT_F destRect = D2D1::RectF(0, 0, size.width, size.height);
        pet->render_target->DrawBitmap(pet->current_frame_bitmap, destRect);
    }
    
    HRESULT hr = pet->render_target->EndDraw();
    if (hr == D2DERR_RECREATE_TARGET) {
        desktop_pet_cleanup_graphics(pet);
        desktop_pet_initialize_graphics(pet);
    }
}

bool desktop_pet_load_skin(DesktopPet* pet, const char* skin_path) {
    if (!pet || !skin_path) return false;
    
    // TODO: 实现皮肤加载
    // 这里应该加载PNG/GIF文件并创建Direct2D位图
    
    return true;
}

// 事件处理实现
void desktop_pet_on_mouse_click(DesktopPet* pet, int x, int y, bool is_double_click) {
    if (!pet) return;
    
    pet->last_interaction_time = desktop_pet_get_current_time_ms();
    pet->user_interaction_mode = true;
    
    if (is_double_click) {
        if (pet->on_double_click_callback) {
            pet->on_double_click_callback(x, y, pet->callback_user_data);
        }
        // 双击触发特殊动作
        desktop_pet_perform_action(pet, PET_ACTION_DANCE);
    } else {
        if (pet->on_click_callback) {
            pet->on_click_callback(x, y, pet->callback_user_data);
        }
        // 单击触发问候
        desktop_pet_perform_action(pet, PET_ACTION_WAVE);
    }
}

void desktop_pet_on_mouse_right_click(DesktopPet* pet, int x, int y) {
    if (!pet) return;
    
    if (pet->on_right_click_callback) {
        pet->on_right_click_callback(x, y, pet->callback_user_data);
    }
    
    // 右键显示上下文菜单
    // TODO: 实现上下文菜单
}

// 配置管理实现
void desktop_pet_get_default_config(PetConfig* config) {
    if (!config) return;
    
    config->width = PET_DEFAULT_WIDTH;
    config->height = PET_DEFAULT_HEIGHT;
    config->animation_speed = PET_DEFAULT_ANIMATION_SPEED;
    config->always_on_top = true;
    config->click_through = false;
    config->auto_hide = false;
    config->transparency = PET_DEFAULT_TRANSPARENCY;
    strcpy_s(config->skin_path, sizeof(config->skin_path), "assets\\skins\\default");
    strcpy_s(config->voice_pack, sizeof(config->voice_pack), "default");
}

// 工具函数实现
bool desktop_pet_is_point_inside(const DesktopPet* pet, int x, int y) {
    if (!pet) return false;
    
    return (x >= pet->position.x && x < pet->position.x + pet->config.width &&
            y >= pet->position.y && y < pet->position.y + pet->config.height);
}

void desktop_pet_get_screen_bounds(RECT* bounds) {
    if (!bounds) return;
    
    SystemParametersInfo(SPI_GETWORKAREA, 0, bounds, 0);
}

bool desktop_pet_clamp_to_screen(DesktopPet* pet) {
    if (!pet) return false;
    
    RECT screen_bounds;
    desktop_pet_get_screen_bounds(&screen_bounds);
    
    bool clamped = false;
    
    if (pet->position.x < screen_bounds.left) {
        pet->position.x = screen_bounds.left;
        clamped = true;
    }
    if (pet->position.y < screen_bounds.top) {
        pet->position.y = screen_bounds.top;
        clamped = true;
    }
    if (pet->position.x + pet->config.width > screen_bounds.right) {
        pet->position.x = screen_bounds.right - pet->config.width;
        clamped = true;
    }
    if (pet->position.y + pet->config.height > screen_bounds.bottom) {
        pet->position.y = screen_bounds.bottom - pet->config.height;
        clamped = true;
    }
    
    if (clamped && pet->hwnd) {
        SetWindowPos(pet->hwnd, nullptr, pet->position.x, pet->position.y, 0, 0, SWP_NOSIZE | SWP_NOZORDER);
    }
    
    return clamped;
}

uint64_t desktop_pet_get_current_time_ms(void) {
    return GetTickCount64();
}

// 窗口过程实现
LRESULT CALLBACK desktop_pet_window_proc(HWND hwnd, UINT msg, WPARAM wParam, LPARAM lParam) {
    DesktopPet* pet = nullptr;
    
    if (msg == WM_NCCREATE) {
        CREATESTRUCT* cs = (CREATESTRUCT*)lParam;
        pet = (DesktopPet*)cs->lpCreateParams;
        SetWindowLongPtr(hwnd, GWLP_USERDATA, (LONG_PTR)pet);
    } else {
        pet = (DesktopPet*)GetWindowLongPtr(hwnd, GWLP_USERDATA);
    }
    
    switch (msg) {
        case WM_PAINT: {
            PAINTSTRUCT ps;
            BeginPaint(hwnd, &ps);
            if (pet) {
                desktop_pet_render(pet);
            }
            EndPaint(hwnd, &ps);
            return 0;
        }
        
        case WM_LBUTTONDOWN: {
            if (pet) {
                int x = LOWORD(lParam);
                int y = HIWORD(lParam);
                desktop_pet_on_mouse_click(pet, x, y, false);
            }
            return 0;
        }
        
        case WM_LBUTTONDBLCLK: {
            if (pet) {
                int x = LOWORD(lParam);
                int y = HIWORD(lParam);
                desktop_pet_on_mouse_click(pet, x, y, true);
            }
            return 0;
        }
        
        case WM_RBUTTONDOWN: {
            if (pet) {
                int x = LOWORD(lParam);
                int y = HIWORD(lParam);
                desktop_pet_on_mouse_right_click(pet, x, y);
            }
            return 0;
        }
        
        case WM_TIMER: {
            if (pet) {
                desktop_pet_update_animation(pet);
                InvalidateRect(hwnd, nullptr, FALSE);
            }
            return 0;
        }
        
        case WM_DESTROY: {
            PostQuitMessage(0);
            return 0;
        }
    }
    
    return DefWindowProc(hwnd, msg, wParam, lParam);
}

// 线程函数实现
DWORD WINAPI desktop_pet_animation_thread(LPVOID param) {
    DesktopPet* pet = (DesktopPet*)param;
    if (!pet) return 1;
    
    while (!pet->should_exit) {
        desktop_pet_update_animation(pet);
        
        if (pet->hwnd) {
            InvalidateRect(pet->hwnd, nullptr, FALSE);
        }
        
        Sleep(1000 / pet->config.animation_speed); // 控制动画帧率
    }
    
    return 0;
}

DWORD WINAPI desktop_pet_ai_thread(LPVOID param) {
    DesktopPet* pet = (DesktopPet*)param;
    if (!pet) return 1;
    
    while (!pet->should_exit) {
        // 检查是否有待处理的用户输入
        if (pet->user_interaction_mode && strlen(pet->last_user_input) > 0) {
            PetAIResponse response;
            if (desktop_pet_get_ai_response(pet, pet->last_user_input, &response)) {
                desktop_pet_apply_ai_response(pet, &response);
            }
            
            // 清空输入
            pet->last_user_input[0] = '\0';
        }
        
        Sleep(100); // 100ms检查间隔
    }
    
    return 0;
}