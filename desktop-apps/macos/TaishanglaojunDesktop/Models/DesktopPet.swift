import Cocoa
import Metal
import MetalKit
import AVFoundation
import CoreAnimation
import Foundation

// MARK: - 桌面宠物状态枚举
enum PetState: Int, CaseIterable {
    case idle = 0
    case walking = 1
    case talking = 2
    case thinking = 3
    case sleeping = 4
    case playing = 5
    case working = 6
    case notification = 7
    
    var animationName: String {
        switch self {
        case .idle: return "idle"
        case .walking: return "walking"
        case .talking: return "talking"
        case .thinking: return "thinking"
        case .sleeping: return "sleeping"
        case .playing: return "playing"
        case .working: return "working"
        case .notification: return "notification"
        }
    }
}

// MARK: - 桌面宠物动作枚举
enum PetAction: Int, CaseIterable {
    case none = 0
    case moveLeft = 1
    case moveRight = 2
    case moveUp = 3
    case moveDown = 4
    case jump = 5
    case dance = 6
    case wave = 7
    case nod = 8
    case shakeHead = 9
}

// MARK: - 桌面宠物情绪枚举
enum PetMood: Int, CaseIterable {
    case happy = 0
    case excited = 1
    case calm = 2
    case tired = 3
    case bored = 4
    case curious = 5
    case focused = 6
}

// MARK: - 桌面宠物配置结构
struct PetConfig {
    var width: CGFloat = 200
    var height: CGFloat = 200
    var animationSpeed: Int = 60
    var alwaysOnTop: Bool = true
    var clickThrough: Bool = false
    var autoHide: Bool = false
    var transparency: CGFloat = 1.0
    var skinPath: String = "default"
    var voicePack: String = "default"
}

// MARK: - 桌面宠物位置结构
struct PetPosition {
    var x: CGFloat = 0
    var y: CGFloat = 0
    var targetX: CGFloat = 0
    var targetY: CGFloat = 0
    var isMoving: Bool = false
}

// MARK: - 桌面宠物动画帧
class PetAnimationFrame {
    let image: NSImage
    let duration: TimeInterval
    let offsetX: CGFloat
    let offsetY: CGFloat
    
    init(image: NSImage, duration: TimeInterval, offsetX: CGFloat = 0, offsetY: CGFloat = 0) {
        self.image = image
        self.duration = duration
        self.offsetX = offsetX
        self.offsetY = offsetY
    }
}

// MARK: - 桌面宠物动画
class PetAnimation {
    var frames: [PetAnimationFrame] = []
    var currentFrame: Int = 0
    var lastFrameTime: TimeInterval = 0
    var loop: Bool = true
    var isPlaying: Bool = false
    
    func addFrame(_ frame: PetAnimationFrame) {
        frames.append(frame)
    }
    
    func getCurrentFrame() -> PetAnimationFrame? {
        guard currentFrame < frames.count else { return nil }
        return frames[currentFrame]
    }
    
    func updateFrame() -> Bool {
        guard isPlaying && !frames.isEmpty else { return false }
        
        let currentTime = CACurrentMediaTime()
        let currentFrameObj = frames[currentFrame]
        
        if currentTime - lastFrameTime >= currentFrameObj.duration {
            currentFrame += 1
            
            if currentFrame >= frames.count {
                if loop {
                    currentFrame = 0
                } else {
                    isPlaying = false
                    currentFrame = frames.count - 1
                    return false
                }
            }
            
            lastFrameTime = currentTime
            return true
        }
        
        return false
    }
    
    func play() {
        isPlaying = true
        lastFrameTime = CACurrentMediaTime()
    }
    
    func stop() {
        isPlaying = false
        currentFrame = 0
    }
}

// MARK: - 桌面宠物AI响应
struct PetAIResponse {
    let responseText: String
    let suggestedAction: PetAction
    let suggestedMood: PetMood
    let confidence: Int
}

// MARK: - 桌面宠物语音
class PetVoice {
    var text: String = ""
    var audioFile: String = ""
    var duration: TimeInterval = 0
    var isPlaying: Bool = false
    
    private var speechSynthesizer: AVSpeechSynthesizer?
    private var audioPlayer: AVAudioPlayer?
    
    init() {
        speechSynthesizer = AVSpeechSynthesizer()
    }
    
    func speak(_ text: String) {
        self.text = text
        isPlaying = true
        
        let utterance = AVSpeechUtterance(string: text)
        utterance.voice = AVSpeechSynthesisVoice(language: "zh-CN")
        utterance.rate = 0.5
        utterance.pitchMultiplier = 1.2
        
        speechSynthesizer?.speak(utterance)
    }
    
    func playSound(_ audioFile: String) {
        guard let url = Bundle.main.url(forResource: audioFile, withExtension: nil) else { return }
        
        do {
            audioPlayer = try AVAudioPlayer(contentsOf: url)
            audioPlayer?.play()
            isPlaying = true
        } catch {
            print("Failed to play sound: \(error)")
        }
    }
    
    func stop() {
        speechSynthesizer?.stopSpeaking(at: .immediate)
        audioPlayer?.stop()
        isPlaying = false
    }
}

// MARK: - 桌面宠物主类
class DesktopPet: NSObject {
    
    // MARK: - Properties
    private var window: NSWindow?
    private var petView: PetView?
    
    var config: PetConfig
    var position: PetPosition
    var currentState: PetState = .idle
    var currentMood: PetMood = .calm
    var currentAction: PetAction = .none
    
    // 动画系统
    private var animations: [PetState: PetAnimation] = [:]
    private var currentAnimation: PetAnimation?
    
    // AI交互
    private var lastUserInput: String = ""
    private var lastAIResponse: PetAIResponse?
    private var lastInteractionTime: TimeInterval = 0
    
    // 语音系统
    private var voice: PetVoice
    
    // 行为系统
    private var lastActionTime: TimeInterval = 0
    private var nextRandomActionTime: TimeInterval = 0
    private var userInteractionMode: Bool = false
    
    // 定时器
    private var animationTimer: Timer?
    private var behaviorTimer: Timer?
    
    // 回调
    var onClickCallback: ((CGPoint) -> Void)?
    var onDoubleClickCallback: ((CGPoint) -> Void)?
    var onRightClickCallback: ((CGPoint) -> Void)?
    var onStateChangeCallback: ((PetState, PetState) -> Void)?
    
    // 常量
    private static let idleTimeoutMs: TimeInterval = 30000
    private static let randomActionMinMs: TimeInterval = 10000
    private static let randomActionMaxMs: TimeInterval = 60000
    
    // MARK: - Initialization
    init(config: PetConfig = PetConfig()) {
        self.config = config
        self.position = PetPosition()
        self.voice = PetVoice()
        
        super.init()
        
        setupInitialPosition()
        setupAnimations()
        updateRandomActionTime()
    }
    
    deinit {
        shutdown()
    }
    
    // MARK: - Setup Methods
    private func setupInitialPosition() {
        guard let screen = NSScreen.main else { return }
        let screenFrame = screen.visibleFrame
        
        position.x = screenFrame.maxX - config.width - 50
        position.y = screenFrame.minY + 100
        position.targetX = position.x
        position.targetY = position.y
    }
    
    private func setupAnimations() {
        // 加载默认动画
        for state in PetState.allCases {
            let animation = PetAnimation()
            loadAnimationForState(state, animation: animation)
            animations[state] = animation
        }
    }
    
    private func loadAnimationForState(_ state: PetState, animation: PetAnimation) {
        // 加载动画帧
        let animationName = state.animationName
        let skinPath = config.skinPath
        
        // 尝试加载动画序列
        for i in 1...8 { // 假设每个动画最多8帧
            if let image = NSImage(named: "\(skinPath)_\(animationName)_\(i)") {
                let frame = PetAnimationFrame(image: image, duration: 1.0 / Double(config.animationSpeed))
                animation.addFrame(frame)
            }
        }
        
        // 如果没有找到序列，尝试加载单帧
        if animation.frames.isEmpty {
            if let image = NSImage(named: "\(skinPath)_\(animationName)") {
                let frame = PetAnimationFrame(image: image, duration: 1.0)
                animation.addFrame(frame)
            }
        }
        
        // 如果还是没有，使用默认图像
        if animation.frames.isEmpty {
            if let defaultImage = NSImage(named: "default_pet") {
                let frame = PetAnimationFrame(image: defaultImage, duration: 1.0)
                animation.addFrame(frame)
            }
        }
    }
    
    // MARK: - Public Methods
    func initialize(parentWindow: NSWindow? = nil) -> Bool {
        guard createWindow() else { return false }
        
        setupTimers()
        show()
        
        return true
    }
    
    func shutdown() {
        animationTimer?.invalidate()
        behaviorTimer?.invalidate()
        voice.stop()
        window?.close()
        window = nil
        petView = nil
    }
    
    func show() {
        window?.makeKeyAndOrderFront(nil)
        window?.level = config.alwaysOnTop ? .floating : .normal
    }
    
    func hide() {
        window?.orderOut(nil)
    }
    
    func setPosition(_ x: CGFloat, _ y: CGFloat) {
        position.x = x
        position.y = y
        position.targetX = x
        position.targetY = y
        position.isMoving = false
        
        window?.setFrameOrigin(NSPoint(x: x, y: y))
    }
    
    func moveTo(_ x: CGFloat, _ y: CGFloat, duration: TimeInterval = 2.0) {
        position.targetX = x
        position.targetY = y
        position.isMoving = true
        
        setState(.walking)
    }
    
    func setState(_ state: PetState) {
        let oldState = currentState
        currentState = state
        
        // 播放相应动画
        if let animation = animations[state] {
            currentAnimation = animation
            animation.play()
        }
        
        // 触发状态变化回调
        if oldState != state {
            onStateChangeCallback?(oldState, state)
        }
    }
    
    func setMood(_ mood: PetMood) {
        currentMood = mood
    }
    
    func performAction(_ action: PetAction) {
        currentAction = action
        lastActionTime = CACurrentMediaTime()
        
        switch action {
        case .moveLeft:
            moveTo(position.x - 100, position.y)
        case .moveRight:
            moveTo(position.x + 100, position.y)
        case .moveUp:
            moveTo(position.x, position.y + 50)
        case .moveDown:
            moveTo(position.x, position.y - 50)
        case .jump, .dance:
            setState(.playing)
        case .wave, .nod, .shakeHead:
            setState(.talking)
        case .none:
            break
        }
    }
    
    // MARK: - AI Interaction
    func processUserInput(_ input: String) {
        lastUserInput = input
        lastInteractionTime = CACurrentMediaTime()
        userInteractionMode = true
        
        setState(.thinking)
        
        // 异步获取AI响应
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            if let response = self?.getAIResponse(input) {
                DispatchQueue.main.async {
                    self?.applyAIResponse(response)
                }
            }
        }
    }
    
    private func getAIResponse(_ input: String) -> PetAIResponse? {
        // 获取AI服务配置
        let manager = DesktopPetManager.shared
        guard !manager.aiServiceURL.isEmpty else {
            return getDefaultAIResponse()
        }
        
        // 构建JSON请求体
        let requestBody: [String: Any] = [
            "message": input,
            "context": [
                "pet_state": currentState.rawValue,
                "pet_mood": currentMood.rawValue,
                "timestamp": Date().timeIntervalSince1970
            ]
        ]
        
        do {
            let jsonData = try JSONSerialization.data(withJSONObject: requestBody, options: [])
            
            // 设置请求头
            var headers = ["Content-Type": "application/json"]
            if !manager.aiAPIKey.isEmpty {
                headers["Authorization"] = "Bearer \(manager.aiAPIKey)"
            }
            
            // 发送HTTP POST请求
            let httpClient = HttpClient.shared
            let response = httpClient.post(manager.aiServiceURL, body: jsonData, headers: headers)
            
            if response.success, let responseData = response.body {
                // 解析JSON响应
                if let jsonObject = try JSONSerialization.jsonObject(with: responseData, options: []) as? [String: Any] {
                    let responseText = jsonObject["response"] as? String ?? "我收到了你的消息！"
                    
                    // 解析建议的动作
                    var suggestedAction: PetAction = .none
                    if let actionString = jsonObject["suggested_action"] as? String,
                       let actionValue = Int(actionString) {
                        suggestedAction = PetAction(rawValue: actionValue) ?? .none
                    }
                    
                    // 解析建议的情绪
                    var suggestedMood: PetMood = .happy
                    if let moodString = jsonObject["suggested_mood"] as? String,
                       let moodValue = Int(moodString) {
                        suggestedMood = PetMood(rawValue: moodValue) ?? .happy
                    }
                    
                    // 解析置信度
                    let confidence = jsonObject["confidence"] as? Int ?? 80
                    
                    return PetAIResponse(
                        responseText: responseText,
                        suggestedAction: suggestedAction,
                        suggestedMood: suggestedMood,
                        confidence: confidence
                    )
                }
            } else {
                print("AI服务请求失败: \(response.errorMessage ?? "未知错误")")
            }
        } catch {
            print("AI服务请求处理错误: \(error.localizedDescription)")
        }
        
        // 如果AI服务不可用，返回默认响应
        return getDefaultAIResponse()
    }
    
    private func getDefaultAIResponse() -> PetAIResponse {
        // 默认AI响应
        let responses = [
            "我很高兴和你聊天！",
            "今天天气真不错呢！",
            "你想让我做什么呢？",
            "我正在学习新的技能！"
        ]
        
        let actions: [PetAction] = [.none, .wave, .nod, .dance]
        let moods: [PetMood] = [.happy, .excited, .curious, .calm]
        
        return PetAIResponse(
            responseText: responses.randomElement() ?? "你好！",
            suggestedAction: actions.randomElement() ?? .none,
            suggestedMood: moods.randomElement() ?? .happy,
            confidence: Int.random(in: 70...95)
        )
    }
    
    private func applyAIResponse(_ response: PetAIResponse) {
        lastAIResponse = response
        
        setMood(response.suggestedMood)
        performAction(response.suggestedAction)
        speak(response.responseText)
        setState(.talking)
    }
    
    // MARK: - Voice System
    func speak(_ text: String) {
        voice.speak(text)
    }
    
    func playSound(_ soundFile: String) {
        voice.playSound(soundFile)
    }
    
    func stopSpeaking() {
        voice.stop()
    }
    
    // MARK: - Behavior System
    private func updateBehavior() {
        let currentTime = CACurrentMediaTime()
        
        // 检查是否需要触发随机动作
        if !userInteractionMode && currentTime >= nextRandomActionTime {
            triggerRandomAction()
            updateRandomActionTime()
        }
        
        // 检查是否空闲太久
        if isIdleTooLong() && currentState != .sleeping {
            setState(.sleeping)
        }
        
        // 重置用户交互模式
        if userInteractionMode && currentTime - lastInteractionTime > 10.0 {
            userInteractionMode = false
        }
    }
    
    private func triggerRandomAction() {
        let actions: [PetAction] = [.moveLeft, .moveRight, .jump, .wave, .dance]
        let randomAction = actions.randomElement() ?? .wave
        performAction(randomAction)
    }
    
    private func isIdleTooLong() -> Bool {
        let currentTime = CACurrentMediaTime()
        return (currentTime - lastInteractionTime) > (Self.idleTimeoutMs / 1000.0)
    }
    
    private func updateRandomActionTime() {
        let minTime = Self.randomActionMinMs / 1000.0
        let maxTime = Self.randomActionMaxMs / 1000.0
        let randomDelay = Double.random(in: minTime...maxTime)
        nextRandomActionTime = CACurrentMediaTime() + randomDelay
    }
    
    // MARK: - Animation System
    private func updateAnimation() {
        guard let animation = currentAnimation else { return }
        
        if animation.updateFrame() {
            petView?.needsDisplay = true
        }
        
        // 更新位置动画
        updatePositionAnimation()
    }
    
    private func updatePositionAnimation() {
        guard position.isMoving else { return }
        
        let dx = position.targetX - position.x
        let dy = position.targetY - position.y
        
        if abs(dx) <= 2 && abs(dy) <= 2 {
            // 到达目标位置
            position.x = position.targetX
            position.y = position.targetY
            position.isMoving = false
            setState(.idle)
        } else {
            // 继续移动
            let moveSpeed: CGFloat = 2
            position.x += dx > 0 ? moveSpeed : (dx < 0 ? -moveSpeed : 0)
            position.y += dy > 0 ? moveSpeed : (dy < 0 ? -moveSpeed : 0)
            
            window?.setFrameOrigin(NSPoint(x: position.x, y: position.y))
        }
    }
    
    // MARK: - Window Management
    private func createWindow() -> Bool {
        let windowRect = NSRect(x: position.x, y: position.y, width: config.width, height: config.height)
        
        window = NSWindow(
            contentRect: windowRect,
            styleMask: [.borderless],
            backing: .buffered,
            defer: false
        )
        
        guard let window = window else { return false }
        
        window.isOpaque = false
        window.backgroundColor = NSColor.clear
        window.hasShadow = false
        window.ignoresMouseEvents = config.clickThrough
        window.level = config.alwaysOnTop ? .floating : .normal
        window.alphaValue = config.transparency
        
        // 创建宠物视图
        petView = PetView(frame: window.contentView?.bounds ?? .zero)
        petView?.pet = self
        window.contentView = petView
        
        return true
    }
    
    private func setupTimers() {
        // 动画定时器
        animationTimer = Timer.scheduledTimer(withTimeInterval: 1.0 / Double(config.animationSpeed), repeats: true) { [weak self] _ in
            self?.updateAnimation()
        }
        
        // 行为定时器
        behaviorTimer = Timer.scheduledTimer(withTimeInterval: 0.1, repeats: true) { [weak self] _ in
            self?.updateBehavior()
        }
    }
    
    // MARK: - Event Handling
    func onMouseClick(at point: CGPoint, isDoubleClick: Bool = false) {
        lastInteractionTime = CACurrentMediaTime()
        userInteractionMode = true
        
        if isDoubleClick {
            onDoubleClickCallback?(point)
            performAction(.dance)
        } else {
            onClickCallback?(point)
            performAction(.wave)
        }
    }
    
    func onMouseRightClick(at point: CGPoint) {
        onRightClickCallback?(point)
        // TODO: 显示上下文菜单
    }
    
    // MARK: - Utility Methods
    func isPointInside(_ point: CGPoint) -> Bool {
        let petRect = NSRect(x: position.x, y: position.y, width: config.width, height: config.height)
        return petRect.contains(point)
    }
    
    func clampToScreen() {
        guard let screen = NSScreen.main else { return }
        let screenFrame = screen.visibleFrame
        
        var clamped = false
        
        if position.x < screenFrame.minX {
            position.x = screenFrame.minX
            clamped = true
        }
        if position.y < screenFrame.minY {
            position.y = screenFrame.minY
            clamped = true
        }
        if position.x + config.width > screenFrame.maxX {
            position.x = screenFrame.maxX - config.width
            clamped = true
        }
        if position.y + config.height > screenFrame.maxY {
            position.y = screenFrame.maxY - config.height
            clamped = true
        }
        
        if clamped {
            window?.setFrameOrigin(NSPoint(x: position.x, y: position.y))
        }
    }
}

// MARK: - 桌面宠物视图
class PetView: NSView {
    weak var pet: DesktopPet?
    
    override func draw(_ dirtyRect: NSRect) {
        super.draw(dirtyRect)
        
        guard let pet = pet,
              let animation = pet.currentAnimation,
              let frame = animation.getCurrentFrame() else { return }
        
        // 绘制当前动画帧
        let imageRect = NSRect(
            x: frame.offsetX,
            y: frame.offsetY,
            width: bounds.width - frame.offsetX,
            height: bounds.height - frame.offsetY
        )
        
        frame.image.draw(in: imageRect)
    }
    
    override func mouseDown(with event: NSEvent) {
        let point = convert(event.locationInWindow, from: nil)
        pet?.onMouseClick(at: point, isDoubleClick: event.clickCount == 2)
    }
    
    override func rightMouseDown(with event: NSEvent) {
        let point = convert(event.locationInWindow, from: nil)
        pet?.onMouseRightClick(at: point)
    }
    
    override var acceptsFirstResponder: Bool {
        return true
    }
}

// MARK: - 桌面宠物管理器
class DesktopPetManager: NSObject {
    
    // MARK: - Properties
    private var pets: [DesktopPet] = []
    private let maxPets: Int = 4
    
    var petsEnabled: Bool = true
    var skinsDirectory: String = "assets/skins"
    var voicesDirectory: String = "assets/voices"
    var aiServiceURL: String = "http://localhost:8080/api/v1/ai/chat"
    var aiAPIKey: String = ""
    
    weak var mainWindow: NSWindow?
    
    // MARK: - Singleton
    static let shared = DesktopPetManager()
    
    private override init() {
        super.init()
    }
    
    // MARK: - Public Methods
    func initialize(mainWindow: NSWindow? = nil) -> Bool {
        self.mainWindow = mainWindow
        
        // 创建默认桌面宠物
        let defaultConfig = PetConfig()
        let pet = DesktopPet(config: defaultConfig)
        
        if pet.initialize(parentWindow: mainWindow) {
            pets.append(pet)
            return true
        }
        
        return false
    }
    
    func shutdown() {
        for pet in pets {
            pet.shutdown()
        }
        pets.removeAll()
    }
    
    func update() {
        guard petsEnabled else { return }
        
        // 桌面宠物的更新由各自的定时器处理
        // 这里可以处理全局逻辑
    }
    
    func addPet(config: PetConfig = PetConfig()) -> DesktopPet? {
        guard pets.count < maxPets else { return nil }
        
        let pet = DesktopPet(config: config)
        if pet.initialize(parentWindow: mainWindow) {
            pets.append(pet)
            return pet
        }
        
        return nil
    }
    
    func removePet(_ pet: DesktopPet) {
        if let index = pets.firstIndex(of: pet) {
            pets[index].shutdown()
            pets.remove(at: index)
        }
    }
    
    func getPets() -> [DesktopPet] {
        return pets
    }
    
    func showAllPets() {
        for pet in pets {
            pet.show()
        }
    }
    
    func hideAllPets() {
        for pet in pets {
            pet.hide()
        }
    }
}