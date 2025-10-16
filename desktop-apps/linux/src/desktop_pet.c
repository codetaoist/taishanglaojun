#include "desktop_pet.h"
#include "application.h"
#include "network.h"
#include <json-c/json.h>

// 全局桌面宠物管理器实例
static DesktopPetManager* g_pet_manager = NULL;

// 桌面宠物管理器实现
DesktopPetManager* desktop_pet_manager_create(void) {
    DesktopPetManager* manager = calloc(1, sizeof(DesktopPetManager));
    if (!manager) {
        return NULL;
    }
    
    manager->max_pets = 4; // 最多支持4个桌面宠物
    manager->pets = calloc(manager->max_pets, sizeof(DesktopPet));
    if (!manager->pets) {
        free(manager);
        return NULL;
    }
    
    manager->pets_enabled = true;
    strncpy(manager->skins_directory, "assets/skins", sizeof(manager->skins_directory) - 1);
    strncpy(manager->voices_directory, "assets/voices", sizeof(manager->voices_directory) - 1);
    strncpy(manager->ai_service_url, "http://localhost:8080/api/v1/ai/chat", sizeof(manager->ai_service_url) - 1);
    
    // 检测可用的显示后端
    manager->preferred_backend = desktop_pet_detect_display_backend();
    
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

bool desktop_pet_manager_initialize(DesktopPetManager* manager, GtkApplication* app) {
    if (!manager) return false;
    
    manager->app = app;
    
    // 创建默认桌面宠物
    PetConfig default_config;
    desktop_pet_get_default_config(&default_config);
    default_config.display_backend = manager->preferred_backend;
    
    DesktopPet* pet = desktop_pet_create(&default_config);
    if (pet && desktop_pet_initialize(pet, manager->main_window)) {
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
    
    g_pet_manager = NULL;
}

void desktop_pet_manager_update(DesktopPetManager* manager) {
    if (!manager || !manager->pets_enabled) return;
    
    // 桌面宠物的更新由GTK定时器处理
    // 这里可以处理全局逻辑
}

// 桌面宠物实现
DesktopPet* desktop_pet_create(const PetConfig* config) {
    DesktopPet* pet = calloc(1, sizeof(DesktopPet));
    if (!pet) return NULL;
    
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
    GdkRectangle screen_bounds;
    desktop_pet_get_screen_bounds(&screen_bounds);
    pet->position.x = screen_bounds.x + screen_bounds.width - pet->config.width - 50;
    pet->position.y = screen_bounds.y + screen_bounds.height - pet->config.height - 100;
    pet->position.target_x = pet->position.x;
    pet->position.target_y = pet->position.y;
    pet->position.is_moving = false;
    
    // 初始化时间
    pet->last_interaction_time = desktop_pet_get_current_time_ms();
    pet->last_action_time = pet->last_interaction_time;
    pet->next_random_action_time = pet->last_interaction_time + 
        PET_RANDOM_ACTION_MIN_MS + (rand() % (PET_RANDOM_ACTION_MAX_MS - PET_RANDOM_ACTION_MIN_MS));
    
    // 初始化同步对象
    pthread_mutex_init(&pet->state_mutex, NULL);
    
    // 设置活动后端
    pet->active_backend = pet->config.display_backend;
    
    return pet;
}

void desktop_pet_destroy(DesktopPet* pet) {
    if (!pet) return;
    
    desktop_pet_shutdown(pet);
    pthread_mutex_destroy(&pet->state_mutex);
    free(pet);
}

bool desktop_pet_initialize(DesktopPet* pet, GtkWidget* parent_window) {
    if (!pet) return false;
    
    // 创建GTK窗口
    pet->window = gtk_window_new(GTK_WINDOW_TOPLEVEL);
    if (!pet->window) {
        PET_LOG_ERROR("Failed to create GTK window");
        return false;
    }
    
    // 设置窗口属性
    gtk_window_set_title(GTK_WINDOW(pet->window), "Desktop Pet");
    gtk_window_set_default_size(GTK_WINDOW(pet->window), pet->config.width, pet->config.height);
    gtk_window_set_decorated(GTK_WINDOW(pet->window), FALSE);
    gtk_window_set_resizable(GTK_WINDOW(pet->window), FALSE);
    gtk_window_set_keep_above(GTK_WINDOW(pet->window), pet->config.always_on_top);
    gtk_window_set_skip_taskbar_hint(GTK_WINDOW(pet->window), TRUE);
    gtk_window_set_skip_pager_hint(GTK_WINDOW(pet->window), TRUE);
    gtk_window_set_accept_focus(GTK_WINDOW(pet->window), !pet->config.click_through);
    
    // 设置窗口位置
    gtk_window_move(GTK_WINDOW(pet->window), pet->position.x, pet->position.y);
    
    // 设置透明背景
    GdkScreen* screen = gtk_window_get_screen(GTK_WINDOW(pet->window));
    GdkVisual* visual = gdk_screen_get_rgba_visual(screen);
    if (visual) {
        gtk_widget_set_visual(pet->window, visual);
    }
    
    // 创建绘图区域
    pet->drawing_area = gtk_drawing_area_new();
    gtk_widget_set_size_request(pet->drawing_area, pet->config.width, pet->config.height);
    gtk_container_add(GTK_CONTAINER(pet->window), pet->drawing_area);
    
    // 连接信号
    g_signal_connect(pet->drawing_area, "draw", G_CALLBACK(desktop_pet_on_draw), pet);
    g_signal_connect(pet->window, "button-press-event", G_CALLBACK(desktop_pet_on_button_press), pet);
    g_signal_connect(pet->window, "motion-notify-event", G_CALLBACK(desktop_pet_on_motion_notify), pet);
    g_signal_connect(pet->window, "key-press-event", G_CALLBACK(desktop_pet_on_key_press_event), pet);
    
    // 设置事件掩码
    gtk_widget_set_events(pet->window, 
        GDK_BUTTON_PRESS_MASK | 
        GDK_BUTTON_RELEASE_MASK | 
        GDK_POINTER_MOTION_MASK | 
        GDK_KEY_PRESS_MASK);
    
    // 应用CSS样式
    GtkCssProvider* css_provider = gtk_css_provider_new();
    gtk_css_provider_load_from_data(css_provider, PET_WINDOW_CSS, -1, NULL);
    GtkStyleContext* style_context = gtk_widget_get_style_context(pet->window);
    gtk_style_context_add_provider(style_context, GTK_STYLE_PROVIDER(css_provider), GTK_STYLE_PROVIDER_PRIORITY_APPLICATION);
    gtk_style_context_add_class(style_context, "pet-window");
    g_object_unref(css_provider);
    
    // 初始化图形系统
    if (!desktop_pet_initialize_graphics(pet)) {
        gtk_widget_destroy(pet->window);
        pet->window = NULL;
        return false;
    }
    
    // 初始化音频系统
    desktop_pet_initialize_audio(pet);
    
    // 加载默认皮肤
    char default_skin_path[PATH_MAX];
    snprintf(default_skin_path, sizeof(default_skin_path), "%s/default/idle.png", pet->config.skin_path);
    desktop_pet_load_skin(pet, default_skin_path);
    
    // 创建定时器
    pet->animation_timer_id = g_timeout_add(1000 / pet->config.animation_speed, 
                                           desktop_pet_animation_timer_callback, pet);
    pet->behavior_timer_id = g_timeout_add(100, desktop_pet_behavior_timer_callback, pet);
    
    // 创建线程
    pet->should_exit = false;
    pthread_create(&pet->animation_thread, NULL, desktop_pet_animation_thread, pet);
    pthread_create(&pet->ai_thread, NULL, desktop_pet_ai_thread, pet);
    
    return true;
}

void desktop_pet_shutdown(DesktopPet* pet) {
    if (!pet) return;
    
    // 停止线程
    pet->should_exit = true;
    if (pet->animation_thread) {
        pthread_join(pet->animation_thread, NULL);
    }
    if (pet->ai_thread) {
        pthread_join(pet->ai_thread, NULL);
    }
    
    // 移除定时器
    if (pet->animation_timer_id > 0) {
        g_source_remove(pet->animation_timer_id);
        pet->animation_timer_id = 0;
    }
    if (pet->behavior_timer_id > 0) {
        g_source_remove(pet->behavior_timer_id);
        pet->behavior_timer_id = 0;
    }
    
    // 清理音频系统
    desktop_pet_cleanup_audio(pet);
    
    // 清理图形资源
    desktop_pet_cleanup_graphics(pet);
    
    // 销毁窗口
    if (pet->window) {
        gtk_widget_destroy(pet->window);
        pet->window = NULL;
    }
}

bool desktop_pet_show(DesktopPet* pet) {
    if (!pet || !pet->window) return false;
    
    gtk_widget_show_all(pet->window);
    return true;
}

bool desktop_pet_hide(DesktopPet* pet) {
    if (!pet || !pet->window) return false;
    
    gtk_widget_hide(pet->window);
    return true;
}

bool desktop_pet_set_position(DesktopPet* pet, int x, int y) {
    if (!pet) return false;
    
    PET_LOCK(pet);
    pet->position.x = x;
    pet->position.y = y;
    pet->position.target_x = x;
    pet->position.target_y = y;
    pet->position.is_moving = false;
    PET_UNLOCK(pet);
    
    if (pet->window) {
        gtk_window_move(GTK_WINDOW(pet->window), x, y);
    }
    
    return true;
}

bool desktop_pet_move_to(DesktopPet* pet, int x, int y, int duration_ms) {
    if (!pet) return false;
    
    PET_LOCK(pet);
    pet->position.target_x = x;
    pet->position.target_y = y;
    pet->position.is_moving = true;
    PET_UNLOCK(pet);
    
    // 设置行走状态
    desktop_pet_set_state(pet, PET_STATE_WALKING);
    
    return true;
}

bool desktop_pet_set_state(DesktopPet* pet, PetState state) {
    if (!pet) return false;
    
    PET_LOCK(pet);
    PetState old_state = pet->current_state;
    pet->current_state = state;
    PET_UNLOCK(pet);
    
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
    
    PET_LOCK(pet);
    pet->current_mood = mood;
    PET_UNLOCK(pet);
    
    return true;
}

bool desktop_pet_perform_action(DesktopPet* pet, PetAction action) {
    if (!pet) return false;
    
    PET_LOCK(pet);
    pet->current_action = action;
    pet->last_action_time = desktop_pet_get_current_time_ms();
    PET_UNLOCK(pet);
    
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
        case PET_ACTION_NONE:
            break;
    }
    
    return true;
}

// 动画系统实现
bool desktop_pet_load_animation(DesktopPet* pet, int animation_id, const char* animation_path) {
    if (!pet || animation_id < 0 || animation_id >= 16) return false;
    
    // TODO: 实现动画加载逻辑
    // 这里应该加载PNG序列或GIF文件并创建动画帧
    
    return true;
}

bool desktop_pet_play_animation(DesktopPet* pet, int animation_id, bool loop) {
    if (!pet || animation_id < 0 || animation_id >= 16) return false;
    
    PET_LOCK(pet);
    pet->current_animation = animation_id;
    pet->animations[animation_id].current_frame = 0;
    pet->animations[animation_id].loop = loop;
    pet->animations[animation_id].playing = true;
    pet->animations[animation_id].last_frame_time = desktop_pet_get_current_time_ms();
    PET_UNLOCK(pet);
    
    return true;
}

bool desktop_pet_stop_animation(DesktopPet* pet) {
    if (!pet) return false;
    
    PET_LOCK(pet);
    if (pet->current_animation >= 0 && pet->current_animation < 16) {
        pet->animations[pet->current_animation].playing = false;
    }
    PET_UNLOCK(pet);
    
    return true;
}

void desktop_pet_update_animation(DesktopPet* pet) {
    if (!pet) return;
    
    PET_LOCK(pet);
    
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
    
    PET_UNLOCK(pet);
    
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
            
            if (pet->window) {
                gtk_window_move(GTK_WINDOW(pet->window), pet->position.x, pet->position.y);
            }
        }
    }
}

// AI交互实现
bool desktop_pet_process_user_input(DesktopPet* pet, const char* input) {
    if (!pet || !input) return false;
    
    strncpy(pet->last_user_input, input, sizeof(pet->last_user_input) - 1);
    pet->last_user_input[sizeof(pet->last_user_input) - 1] = '\0';
    pet->last_interaction_time = desktop_pet_get_current_time_ms();
    pet->user_interaction_mode = true;
    
    // 设置思考状态
    desktop_pet_set_state(pet, PET_STATE_THINKING);
    
    return true;
}

bool desktop_pet_get_ai_response(DesktopPet* pet, const char* input, PetAIResponse* response) {
    if (!pet || !input || !response) return false;
    
    // 默认响应值
    strncpy(response->response_text, "I'm sorry, I'm having trouble connecting to my AI service right now.", sizeof(response->response_text) - 1);
    response->response_text[sizeof(response->response_text) - 1] = '\0';
    response->suggested_action = PET_ACTION_NONE;
    response->suggested_mood = PET_MOOD_CALM;
    response->confidence = 50;
    
    // 检查AI服务URL配置
    if (!g_pet_manager || !g_pet_manager->ai_service_url || strlen(g_pet_manager->ai_service_url) == 0) {
        strncpy(response->response_text, "AI service is not configured.", sizeof(response->response_text) - 1);
        response->response_text[sizeof(response->response_text) - 1] = '\0';
        return false;
    }
    
    // 构建JSON请求体
    json_object* request_json = json_object_new_object();
    json_object* message_obj = json_object_new_string(input);
    json_object* model_obj = json_object_new_string("gpt-3.5-turbo");
    json_object* max_tokens_obj = json_object_new_int(150);
    json_object* temperature_obj = json_object_new_double(0.7);
    
    json_object_object_add(request_json, "message", message_obj);
    json_object_object_add(request_json, "model", model_obj);
    json_object_object_add(request_json, "max_tokens", max_tokens_obj);
    json_object_object_add(request_json, "temperature", temperature_obj);
    
    const char* json_string = json_object_to_json_string(request_json);
    
    // 准备HTTP请求头
    const char* headers[3];
    headers[0] = "Content-Type: application/json";
    
    char auth_header[512];
    int header_count = 1;
    if (g_pet_manager->ai_api_key && strlen(g_pet_manager->ai_api_key) > 0) {
        snprintf(auth_header, sizeof(auth_header), "Authorization: Bearer %s", g_pet_manager->ai_api_key);
        headers[1] = auth_header;
        headers[2] = NULL;
        header_count = 2;
    } else {
        headers[1] = NULL;
    }
    
    // 发送HTTP POST请求
    http_response_t* http_response = http_client_post(g_http_client, 
                                                     g_pet_manager->ai_service_url, 
                                                     json_string, 
                                                     headers, 
                                                     header_count);
    
    bool success = false;
    if (http_response && http_response->success && http_response->status_code == 200) {
        // 解析AI服务响应
        json_object* response_json = json_tokener_parse(http_response->body);
        if (response_json) {
            json_object* response_text_obj;
            json_object* action_obj;
            json_object* mood_obj;
            json_object* confidence_obj;
            
            // 提取响应文本
            if (json_object_object_get_ex(response_json, "response", &response_text_obj)) {
                const char* ai_response_text = json_object_get_string(response_text_obj);
                if (ai_response_text && strlen(ai_response_text) > 0) {
                    strncpy(response->response_text, ai_response_text, sizeof(response->response_text) - 1);
                    response->response_text[sizeof(response->response_text) - 1] = '\0';
                }
            }
            
            // 提取建议动作
            if (json_object_object_get_ex(response_json, "suggested_action", &action_obj)) {
                const char* action_str = json_object_get_string(action_obj);
                if (action_str) {
                    if (strcmp(action_str, "wave") == 0) response->suggested_action = PET_ACTION_WAVE;
                    else if (strcmp(action_str, "nod") == 0) response->suggested_action = PET_ACTION_NOD;
                    else if (strcmp(action_str, "jump") == 0) response->suggested_action = PET_ACTION_JUMP;
                    else if (strcmp(action_str, "dance") == 0) response->suggested_action = PET_ACTION_DANCE;
                    else response->suggested_action = PET_ACTION_NONE;
                }
            }
            
            // 提取建议情绪
            if (json_object_object_get_ex(response_json, "suggested_mood", &mood_obj)) {
                const char* mood_str = json_object_get_string(mood_obj);
                if (mood_str) {
                    if (strcmp(mood_str, "happy") == 0) response->suggested_mood = PET_MOOD_HAPPY;
                    else if (strcmp(mood_str, "sad") == 0) response->suggested_mood = PET_MOOD_SAD;
                    else if (strcmp(mood_str, "excited") == 0) response->suggested_mood = PET_MOOD_EXCITED;
                    else if (strcmp(mood_str, "angry") == 0) response->suggested_mood = PET_MOOD_ANGRY;
                    else response->suggested_mood = PET_MOOD_CALM;
                }
            }
            
            // 提取置信度
            if (json_object_object_get_ex(response_json, "confidence", &confidence_obj)) {
                response->confidence = json_object_get_int(confidence_obj);
            } else {
                response->confidence = 80; // 默认置信度
            }
            
            json_object_put(response_json);
            success = true;
        }
    } else {
        // HTTP请求失败，使用简单的本地响应逻辑作为后备
        if (strstr(input, "hello") || strstr(input, "hi")) {
            strncpy(response->response_text, "Hello! How can I help you today?", sizeof(response->response_text) - 1);
            response->response_text[sizeof(response->response_text) - 1] = '\0';
            response->suggested_action = PET_ACTION_WAVE;
            response->suggested_mood = PET_MOOD_HAPPY;
            response->confidence = 90;
        } else if (strstr(input, "sad") || strstr(input, "upset")) {
            strncpy(response->response_text, "I'm sorry to hear that. Is there anything I can do to help?", sizeof(response->response_text) - 1);
            response->response_text[sizeof(response->response_text) - 1] = '\0';
            response->suggested_action = PET_ACTION_NOD;
            response->suggested_mood = PET_MOOD_SAD;
            response->confidence = 80;
        } else {
            strncpy(response->response_text, "That's interesting! Tell me more about it.", sizeof(response->response_text) - 1);
            response->response_text[sizeof(response->response_text) - 1] = '\0';
            response->suggested_action = PET_ACTION_NOD;
            response->suggested_mood = PET_MOOD_HAPPY;
            response->confidence = 70;
        }
        success = true;
    }
    
    // 清理资源
    if (http_response) {
        http_response_free(http_response);
    }
    json_object_put(request_json);
    
    return success;
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
    
    strncpy(pet->current_voice.text, text, sizeof(pet->current_voice.text) - 1);
    pet->current_voice.text[sizeof(pet->current_voice.text) - 1] = '\0';
    pet->current_voice.is_playing = true;
    
    // 使用espeak进行TTS语音合成
    char command[1024];
    char escaped_text[512];
    
    // 转义文本中的特殊字符
    const char* src = text;
    char* dst = escaped_text;
    size_t remaining = sizeof(escaped_text) - 1;
    
    while (*src && remaining > 1) {
        if (*src == '"' || *src == '\\' || *src == '$' || *src == '`') {
            if (remaining > 2) {
                *dst++ = '\\';
                *dst++ = *src++;
                remaining -= 2;
            } else {
                break;
            }
        } else {
            *dst++ = *src++;
            remaining--;
        }
    }
    *dst = '\0';
    
    // 构建espeak命令
    snprintf(command, sizeof(command), "espeak -s 150 -v zh \"%s\" 2>/dev/null &", escaped_text);
    
    // 执行TTS命令
    int result = system(command);
    if (result == -1) {
        // 如果espeak不可用，尝试使用festival
        snprintf(command, sizeof(command), "echo \"%s\" | festival --tts 2>/dev/null &", escaped_text);
        result = system(command);
        
        if (result == -1) {
            // 如果都不可用，使用系统通知音
            system("pactl play-sample bell 2>/dev/null || aplay /usr/share/sounds/alsa/Front_Left.wav 2>/dev/null &");
            PET_LOG_WARNING("TTS engines not available, using system sound");
        }
    }
    
    PET_LOG_INFO("Pet speaking: %s", text);
    
    return true;
}

bool desktop_pet_play_sound(DesktopPet* pet, const char* sound_file) {
    if (!pet || !sound_file) return false;
    
    return desktop_pet_audio_play_file(pet, sound_file);
}

void desktop_pet_stop_speaking(DesktopPet* pet) {
    if (!pet) return;
    
    pet->current_voice.is_playing = false;
    
    // 停止TTS播放进程
    system("pkill -f espeak 2>/dev/null");
    system("pkill -f festival 2>/dev/null");
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
    if (!pet) return false;
    
    // GTK图形系统已经在窗口创建时初始化
    // 这里可以进行额外的图形初始化
    
    return true;
}

void desktop_pet_cleanup_graphics(DesktopPet* pet) {
    if (!pet) return;
    
    // 清理动画帧
    for (int i = 0; i < 16; i++) {
        PetAnimation* anim = &pet->animations[i];
        for (int j = 0; j < anim->frame_count; j++) {
            if (anim->frames[j].pixbuf) {
                g_object_unref(anim->frames[j].pixbuf);
            }
            if (anim->frames[j].surface) {
                cairo_surface_destroy(anim->frames[j].surface);
            }
        }
        if (anim->frames) {
            free(anim->frames);
            anim->frames = NULL;
        }
        anim->frame_count = 0;
    }
    
    // 清理显示后端特定资源
    switch (pet->active_backend) {
        case DISPLAY_BACKEND_X11:
            desktop_pet_cleanup_x11(pet);
            break;
        case DISPLAY_BACKEND_WAYLAND:
            desktop_pet_cleanup_wayland(pet);
            break;
        default:
            break;
    }
}

void desktop_pet_render(DesktopPet* pet, cairo_t* cr) {
    if (!pet || !cr) return;
    
    // 清除背景（透明）
    cairo_set_operator(cr, CAIRO_OPERATOR_CLEAR);
    cairo_paint(cr);
    cairo_set_operator(cr, CAIRO_OPERATOR_OVER);
    
    // 渲染当前动画帧
    PET_LOCK(pet);
    if (pet->current_animation >= 0 && pet->current_animation < 16) {
        PetAnimation* anim = &pet->animations[pet->current_animation];
        if (anim->current_frame < anim->frame_count && anim->frames) {
            PetAnimationFrame* frame = &anim->frames[anim->current_frame];
            
            if (frame->surface) {
                cairo_set_source_surface(cr, frame->surface, frame->offset_x, frame->offset_y);
                cairo_paint_with_alpha(cr, pet->config.transparency);
            } else if (frame->pixbuf) {
                gdk_cairo_set_source_pixbuf(cr, frame->pixbuf, frame->offset_x, frame->offset_y);
                cairo_paint_with_alpha(cr, pet->config.transparency);
            }
        }
    }
    PET_UNLOCK(pet);
}

bool desktop_pet_load_skin(DesktopPet* pet, const char* skin_path) {
    if (!pet || !skin_path) return false;
    
    // TODO: 实现皮肤加载
    // 这里应该加载PNG/GIF文件并创建GdkPixbuf
    
    PET_LOG_INFO("Loading skin: %s", skin_path);
    
    return true;
}

// 显示后端管理实现
DisplayBackend desktop_pet_detect_display_backend(void) {
    // 检查环境变量
    const char* wayland_display = getenv("WAYLAND_DISPLAY");
    const char* x11_display = getenv("DISPLAY");
    
    if (wayland_display && strlen(wayland_display) > 0) {
        return DISPLAY_BACKEND_WAYLAND;
    } else if (x11_display && strlen(x11_display) > 0) {
        return DISPLAY_BACKEND_X11;
    }
    
    return DISPLAY_BACKEND_AUTO;
}

bool desktop_pet_initialize_x11(DesktopPet* pet) {
    if (!pet) return false;
    
#ifdef HAVE_X11
    pet->x11_data.display = XOpenDisplay(NULL);
    if (!pet->x11_data.display) {
        PET_LOG_ERROR("Failed to open X11 display");
        return false;
    }
    
    pet->x11_data.screen = DefaultScreen(pet->x11_data.display);
    pet->x11_data.root_window = RootWindow(pet->x11_data.display, pet->x11_data.screen);
    
    // 检查复合扩展
    int composite_event, composite_error;
    pet->x11_data.composite_available = XCompositeQueryExtension(pet->x11_data.display, 
                                                                &composite_event, &composite_error);
    
    // 检查Xfixes扩展
    int xfixes_event, xfixes_error;
    pet->x11_data.xfixes_available = XFixesQueryExtension(pet->x11_data.display, 
                                                         &xfixes_event, &xfixes_error);
    
    PET_LOG_INFO("X11 initialized - Composite: %s, Xfixes: %s", 
                 pet->x11_data.composite_available ? "yes" : "no",
                 pet->x11_data.xfixes_available ? "yes" : "no");
    
    return true;
#else
    PET_LOG_ERROR("X11 support not compiled");
    return false;
#endif
}

bool desktop_pet_initialize_wayland(DesktopPet* pet) {
    if (!pet) return false;
    
#ifdef HAVE_WAYLAND
    pet->wayland_data.display = wl_display_connect(NULL);
    if (!pet->wayland_data.display) {
        PET_LOG_ERROR("Failed to connect to Wayland display");
        return false;
    }
    
    // TODO: 实现Wayland初始化
    PET_LOG_INFO("Wayland initialized");
    
    return true;
#else
    PET_LOG_ERROR("Wayland support not compiled");
    return false;
#endif
}

void desktop_pet_cleanup_x11(DesktopPet* pet) {
    if (!pet) return;
    
#ifdef HAVE_X11
    if (pet->x11_data.display) {
        XCloseDisplay(pet->x11_data.display);
        pet->x11_data.display = NULL;
    }
#endif
}

void desktop_pet_cleanup_wayland(DesktopPet* pet) {
    if (!pet) return;
    
#ifdef HAVE_WAYLAND
    if (pet->wayland_data.display) {
        wl_display_disconnect(pet->wayland_data.display);
        pet->wayland_data.display = NULL;
    }
#endif
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

// GTK 回调函数实现
gboolean desktop_pet_on_draw(GtkWidget* widget, cairo_t* cr, gpointer user_data) {
    DesktopPet* pet = (DesktopPet*)user_data;
    if (pet) {
        desktop_pet_render(pet, cr);
    }
    return FALSE;
}

gboolean desktop_pet_on_button_press(GtkWidget* widget, GdkEventButton* event, gpointer user_data) {
    DesktopPet* pet = (DesktopPet*)user_data;
    if (pet) {
        bool is_double_click = (event->type == GDK_2BUTTON_PRESS);
        
        if (event->button == 1) { // 左键
            desktop_pet_on_mouse_click(pet, (int)event->x, (int)event->y, is_double_click);
        } else if (event->button == 3) { // 右键
            desktop_pet_on_mouse_right_click(pet, (int)event->x, (int)event->y);
        }
    }
    return TRUE;
}

gboolean desktop_pet_on_motion_notify(GtkWidget* widget, GdkEventMotion* event, gpointer user_data) {
    DesktopPet* pet = (DesktopPet*)user_data;
    if (pet) {
        desktop_pet_on_mouse_move(pet, (int)event->x, (int)event->y);
    }
    return FALSE;
}

gboolean desktop_pet_on_key_press_event(GtkWidget* widget, GdkEventKey* event, gpointer user_data) {
    DesktopPet* pet = (DesktopPet*)user_data;
    if (pet) {
        desktop_pet_on_key_press(pet, event->keyval);
    }
    return FALSE;
}

gboolean desktop_pet_animation_timer_callback(gpointer user_data) {
    DesktopPet* pet = (DesktopPet*)user_data;
    if (pet && pet->drawing_area) {
        gtk_widget_queue_draw(pet->drawing_area);
    }
    return TRUE; // 继续定时器
}

gboolean desktop_pet_behavior_timer_callback(gpointer user_data) {
    DesktopPet* pet = (DesktopPet*)user_data;
    if (pet) {
        desktop_pet_update_behavior(pet);
    }
    return TRUE; // 继续定时器
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
    strncpy(config->skin_path, "assets/skins/default", sizeof(config->skin_path) - 1);
    config->skin_path[sizeof(config->skin_path) - 1] = '\0';
    strncpy(config->voice_pack, "default", sizeof(config->voice_pack) - 1);
    config->voice_pack[sizeof(config->voice_pack) - 1] = '\0';
    config->display_backend = DISPLAY_BACKEND_AUTO;
}

// 音频系统实现
bool desktop_pet_initialize_audio(DesktopPet* pet) {
    if (!pet) return false;
    
    // TODO: 初始化ALSA或PulseAudio
    PET_LOG_INFO("Audio system initialized");
    
    return true;
}

void desktop_pet_cleanup_audio(DesktopPet* pet) {
    if (!pet) return;
    
    // TODO: 清理音频资源
}

bool desktop_pet_audio_play_file(DesktopPet* pet, const char* audio_file) {
    if (!pet || !audio_file) return false;
    
    // TODO: 实现音频文件播放
    PET_LOG_INFO("Playing audio file: %s", audio_file);
    
    return true;
}

// 工具函数实现
bool desktop_pet_is_point_inside(const DesktopPet* pet, int x, int y) {
    if (!pet) return false;
    
    return (x >= pet->position.x && x < pet->position.x + pet->config.width &&
            y >= pet->position.y && y < pet->position.y + pet->config.height);
}

void desktop_pet_get_screen_bounds(GdkRectangle* bounds) {
    if (!bounds) return;
    
    GdkScreen* screen = gdk_screen_get_default();
    if (screen) {
        bounds->x = 0;
        bounds->y = 0;
        bounds->width = gdk_screen_get_width(screen);
        bounds->height = gdk_screen_get_height(screen);
    } else {
        bounds->x = 0;
        bounds->y = 0;
        bounds->width = 1920;
        bounds->height = 1080;
    }
}

bool desktop_pet_clamp_to_screen(DesktopPet* pet) {
    if (!pet) return false;
    
    GdkRectangle screen_bounds;
    desktop_pet_get_screen_bounds(&screen_bounds);
    
    bool clamped = false;
    
    if (pet->position.x < screen_bounds.x) {
        pet->position.x = screen_bounds.x;
        clamped = true;
    }
    if (pet->position.y < screen_bounds.y) {
        pet->position.y = screen_bounds.y;
        clamped = true;
    }
    if (pet->position.x + pet->config.width > screen_bounds.x + screen_bounds.width) {
        pet->position.x = screen_bounds.x + screen_bounds.width - pet->config.width;
        clamped = true;
    }
    if (pet->position.y + pet->config.height > screen_bounds.y + screen_bounds.height) {
        pet->position.y = screen_bounds.y + screen_bounds.height - pet->config.height;
        clamped = true;
    }
    
    if (clamped && pet->window) {
        gtk_window_move(GTK_WINDOW(pet->window), pet->position.x, pet->position.y);
    }
    
    return clamped;
}

uint64_t desktop_pet_get_current_time_ms(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (uint64_t)ts.tv_sec * 1000 + ts.tv_nsec / 1000000;
}

// 线程函数实现
void* desktop_pet_animation_thread(void* param) {
    DesktopPet* pet = (DesktopPet*)param;
    if (!pet) return NULL;
    
    while (!pet->should_exit) {
        desktop_pet_update_animation(pet);
        
        // 触发重绘
        if (pet->drawing_area) {
            g_idle_add((GSourceFunc)gtk_widget_queue_draw, pet->drawing_area);
        }
        
        usleep(1000000 / pet->config.animation_speed); // 控制动画帧率
    }
    
    return NULL;
}

void* desktop_pet_ai_thread(void* param) {
    DesktopPet* pet = (DesktopPet*)param;
    if (!pet) return NULL;
    
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
        
        usleep(100000); // 100ms检查间隔
    }
    
    return NULL;
}

void desktop_pet_on_mouse_move(DesktopPet* pet, int x, int y) {
    // TODO: 实现鼠标移动处理
}

void desktop_pet_on_key_press(DesktopPet* pet, int key_code) {
    // TODO: 实现按键处理
}