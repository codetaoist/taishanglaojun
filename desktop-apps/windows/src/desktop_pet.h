#pragma once

#include "pch.h"
#include <d3d11.h>
#include <directxmath.h>

// 桌面宠物状态枚举
enum class PetState {
    IDLE = 0,
    WALKING,
    RUNNING,
    SITTING,
    SLEEPING,
    PLAYING,
    TALKING,
    THINKING,
    HAPPY,
    SAD,
    ANGRY,
    SURPRISED
};

// 桌面宠物心情枚举
enum class PetMood {
    NEUTRAL = 0,
    HAPPY,
    SAD,
    EXCITED,
    TIRED,
    ANGRY,
    CURIOUS,
    PLAYFUL
};

// 桌面宠物动作枚举
enum class PetAction {
    NONE = 0,
    WAVE,
    JUMP,
    DANCE,
    SPIN,
    NOD,
    SHAKE_HEAD,
    CLAP,
    STRETCH,
    YAWN,
    WINK
};

// 动画帧结构
struct AnimationFrame {
    std::wstring imagePath;
    int duration;           // 帧持续时间(毫秒)
    DirectX::XMFLOAT2 offset;  // 相对偏移
    float scale;            // 缩放比例
    float rotation;         // 旋转角度
};

// 动画序列结构
struct Animation {
    std::wstring name;
    std::vector<AnimationFrame> frames;
    bool loop;
    int totalDuration;
};

// AI响应结构
struct AIResponse {
    std::string text;
    PetState suggestedState;
    PetMood suggestedMood;
    PetAction suggestedAction;
    std::string emotion;
    float confidence;
};

// 桌面宠物配置
struct PetConfig {
    std::wstring name;
    std::wstring skinPath;
    DirectX::XMFLOAT2 position;
    DirectX::XMFLOAT2 size;
    bool draggable;
    bool clickThrough;
    bool alwaysOnTop;
    float opacity;
    int behaviorUpdateInterval;  // 行为更新间隔(毫秒)
    bool enableAI;
    bool enableVoice;
    std::string aiModel;
    std::string voiceId;
};

// 桌面宠物类
class DesktopPet {
public:
    DesktopPet();
    ~DesktopPet();

    // 初始化和清理
    bool Initialize(const PetConfig& config);
    void Shutdown();

    // 窗口管理
    bool CreatePetWindow();
    void DestroyPetWindow();
    HWND GetWindow() const { return m_hWnd; }

    // 显示控制
    void Show();
    void Hide();
    bool IsVisible() const { return m_visible; }

    // 位置和移动
    void SetPosition(const DirectX::XMFLOAT2& pos);
    DirectX::XMFLOAT2 GetPosition() const { return m_position; }
    void MoveTo(const DirectX::XMFLOAT2& targetPos, int duration);
    void StartDragging(const DirectX::XMFLOAT2& mousePos);
    void UpdateDragging(const DirectX::XMFLOAT2& mousePos);
    void StopDragging();

    // 状态管理
    void SetState(PetState state);
    PetState GetState() const { return m_currentState; }
    void SetMood(PetMood mood);
    PetMood GetMood() const { return m_currentMood; }

    // 动作执行
    void PerformAction(PetAction action);
    void PlayAnimation(const std::wstring& animationName, bool loop = false);
    void StopAnimation();

    // AI交互
    void ProcessUserInput(const std::string& input);
    void SpeakText(const std::string& text);
    void StopSpeaking();

    // 更新和渲染
    void Update(float deltaTime);
    void Render();

    // 事件处理
    void OnMouseClick(const DirectX::XMFLOAT2& mousePos, bool doubleClick = false);
    void OnMouseRightClick(const DirectX::XMFLOAT2& mousePos);
    void OnMouseMove(const DirectX::XMFLOAT2& mousePos);

    // 配置管理
    void SetConfig(const PetConfig& config) { m_config = config; }
    const PetConfig& GetConfig() const { return m_config; }

private:
    // 内部方法
    bool InitializeGraphics();
    void CleanupGraphics();
    bool LoadSkin(const std::wstring& skinPath);
    bool LoadAnimations();
    void UpdateAnimation(float deltaTime);
    void UpdateBehavior(float deltaTime);
    void UpdateMovement(float deltaTime);
    void ClampToScreen();
    
    // AI相关
    void ProcessAIResponse(const AIResponse& response);
    void SendAIRequest(const std::string& input);
    
    // 渲染相关
    void RenderPet();
    void RenderUI();
    
    // 窗口过程
    static LRESULT CALLBACK PetWndProc(HWND hwnd, UINT msg, WPARAM wParam, LPARAM lParam);

private:
    // 窗口相关
    HWND m_hWnd;
    bool m_visible;
    
    // 配置
    PetConfig m_config;
    
    // 状态
    PetState m_currentState;
    PetMood m_currentMood;
    PetAction m_currentAction;
    
    // 位置和移动
    DirectX::XMFLOAT2 m_position;
    DirectX::XMFLOAT2 m_targetPosition;
    DirectX::XMFLOAT2 m_velocity;
    bool m_moving;
    float m_moveProgress;
    int m_moveDuration;
    
    // 拖拽
    bool m_dragging;
    DirectX::XMFLOAT2 m_dragOffset;
    
    // 动画
    std::map<std::wstring, Animation> m_animations;
    std::wstring m_currentAnimation;
    int m_currentFrame;
    float m_animationTime;
    bool m_animationLoop;
    
    // 行为
    float m_behaviorTimer;
    float m_idleTimer;
    
    // 图形资源
    ID3D11Device* m_pDevice;
    ID3D11DeviceContext* m_pContext;
    IDXGISwapChain* m_pSwapChain;
    ID3D11RenderTargetView* m_pRenderTargetView;
    ID3D11Texture2D* m_pCurrentTexture;
    ID3D11ShaderResourceView* m_pCurrentSRV;
    
    // 时间
    std::chrono::high_resolution_clock::time_point m_lastUpdateTime;
    
    // AI和语音
    bool m_aiProcessing;
    bool m_speaking;
    std::thread m_aiThread;
    std::thread m_voiceThread;
    std::mutex m_stateMutex;
};

// 桌面宠物管理器
class DesktopPetManager {
public:
    DesktopPetManager();
    ~DesktopPetManager();

    // 初始化和清理
    bool Initialize();
    void Shutdown();

    // 宠物管理
    DesktopPet* CreatePet(const PetConfig& config);
    void DestroyPet(DesktopPet* pet);
    void DestroyAllPets();
    
    // 获取宠物
    DesktopPet* GetPet(int index);
    int GetPetCount() const { return static_cast<int>(m_pets.size()); }
    
    // 更新
    void Update();
    
    // 配置
    void SetEnabled(bool enabled) { m_enabled = enabled; }
    bool IsEnabled() const { return m_enabled; }
    
    // 全局设置
    void SetGlobalOpacity(float opacity);
    void SetGlobalScale(float scale);
    void ShowAllPets();
    void HideAllPets();

private:
    std::vector<std::unique_ptr<DesktopPet>> m_pets;
    bool m_enabled;
    bool m_initialized;
    float m_globalOpacity;
    float m_globalScale;
    
    // 更新线程
    std::thread m_updateThread;
    std::atomic<bool> m_running;
    std::mutex m_petsMutex;
};

// 全局函数
DesktopPetManager* GetPetManager();
bool InitializePetSystem();
void ShutdownPetSystem();