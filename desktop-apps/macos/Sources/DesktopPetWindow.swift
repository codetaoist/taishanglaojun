import Cocoa
import MetalKit
import Foundation

// MARK: - Desktop Pet Window
class DesktopPetWindow: NSWindow {
    
    // MARK: - Properties
    private var petView: DesktopPetView!
    private var isDragging = false
    private var dragOffset = NSPoint.zero
    
    // MARK: - Initialization
    override init(contentRect: NSRect, styleMask style: NSWindow.StyleMask, backing backingStoreType: NSWindow.BackingStoreType, defer flag: Bool) {
        super.init(contentRect: contentRect, styleMask: [.borderless], backing: backingStoreType, defer: flag)
        
        setupWindow()
        setupPetView()
    }
    
    // MARK: - Window Setup
    private func setupWindow() {
        // 窗口属性
        isOpaque = false
        backgroundColor = NSColor.clear
        hasShadow = false
        level = .floating
        collectionBehavior = [.canJoinAllSpaces, .stationary]
        isMovableByWindowBackground = false
        
        // 忽略鼠标事件，除非在宠物区域
        ignoresMouseEvents = false
        
        // 设置窗口大小
        setContentSize(NSSize(width: 200, height: 200))
        
        // 居中显示
        center()
    }
    
    private func setupPetView() {
        petView = DesktopPetView(frame: contentView!.bounds)
        petView.autoresizingMask = [.width, .height]
        petView.petWindow = self
        contentView?.addSubview(petView)
    }
    
    // MARK: - Mouse Events
    override func mouseDown(with event: NSEvent) {
        let locationInWindow = event.locationInWindow
        
        if petView.isPointInPet(locationInWindow) {
            isDragging = true
            dragOffset = NSPoint(x: locationInWindow.x - frame.origin.x,
                               y: locationInWindow.y - frame.origin.y)
            petView.startInteraction()
        }
    }
    
    override func mouseDragged(with event: NSEvent) {
        if isDragging {
            let screenPoint = NSEvent.mouseLocation
            let newOrigin = NSPoint(x: screenPoint.x - dragOffset.x,
                                  y: screenPoint.y - dragOffset.y)
            setFrameOrigin(newOrigin)
        }
    }
    
    override func mouseUp(with event: NSEvent) {
        if isDragging {
            isDragging = false
            petView.endInteraction()
        }
    }
    
    override func rightMouseDown(with event: NSEvent) {
        let locationInWindow = event.locationInWindow
        
        if petView.isPointInPet(locationInWindow) {
            showContextMenu(at: locationInWindow)
        }
    }
    
    // MARK: - Context Menu
    private func showContextMenu(at point: NSPoint) {
        let menu = NSMenu()
        
        // 宠物状态
        let statusItem = NSMenuItem(title: "状态: 快乐", action: nil, keyEquivalent: "")
        statusItem.isEnabled = false
        menu.addItem(statusItem)
        
        menu.addItem(NSMenuItem.separator())
        
        // 喂食
        let feedItem = NSMenuItem(title: "喂食", action: #selector(feedPet), keyEquivalent: "")
        feedItem.target = self
        menu.addItem(feedItem)
        
        // 玩耍
        let playItem = NSMenuItem(title: "玩耍", action: #selector(playWithPet), keyEquivalent: "")
        playItem.target = self
        menu.addItem(playItem)
        
        // 休息
        let restItem = NSMenuItem(title: "休息", action: #selector(petRest), keyEquivalent: "")
        restItem.target = self
        menu.addItem(restItem)
        
        menu.addItem(NSMenuItem.separator())
        
        // 设置
        let settingsItem = NSMenuItem(title: "设置...", action: #selector(showPetSettings), keyEquivalent: "")
        settingsItem.target = self
        menu.addItem(settingsItem)
        
        // 隐藏
        let hideItem = NSMenuItem(title: "隐藏宠物", action: #selector(hidePet), keyEquivalent: "")
        hideItem.target = self
        menu.addItem(hideItem)
        
        // 显示菜单
        let screenPoint = convertToScreen(NSRect(origin: point, size: .zero)).origin
        menu.popUp(positioning: nil, at: NSPoint(x: 0, y: 0), in: nil)
    }
    
    // MARK: - Pet Actions
    @objc private func feedPet() {
        petView.performAction(.feed)
        NotificationManager.shared.sendPetNotification(
            title: "宠物喂食",
            body: "你的宠物很开心！",
            action: "feed"
        )
    }
    
    @objc private func playWithPet() {
        petView.performAction(.play)
        NotificationManager.shared.sendPetNotification(
            title: "宠物玩耍",
            body: "你的宠物正在开心地玩耍！",
            action: "play"
        )
    }
    
    @objc private func petRest() {
        petView.performAction(.rest)
        NotificationManager.shared.sendPetNotification(
            title: "宠物休息",
            body: "你的宠物正在休息...",
            action: "rest"
        )
    }
    
    @objc private func showPetSettings() {
        NotificationCenter.default.post(name: .showPetSettings, object: nil)
    }
    
    @objc private func hidePet() {
        NotificationCenter.default.post(name: .hideDesktopPet, object: nil)
    }
    
    // MARK: - Public Methods
    func showPet() {
        orderFront(nil)
        petView.startAnimation()
    }
    
    func hidePet() {
        orderOut(nil)
        petView.stopAnimation()
    }
    
    func updatePetState(_ state: PetState) {
        petView.updateState(state)
    }
}

// MARK: - Desktop Pet View
class DesktopPetView: MTKView {
    
    // MARK: - Properties
    weak var petWindow: DesktopPetWindow?
    private var petRenderer: PetRenderer!
    private var animationTimer: Timer?
    private var currentState: PetState = .idle
    private var currentAction: PetAction?
    
    // MARK: - Initialization
    override init(frame frameRect: NSRect, device: MTLDevice?) {
        super.init(frame: frameRect, device: device ?? MTLCreateSystemDefaultDevice())
        setupView()
    }
    
    required init(coder: NSCoder) {
        super.init(coder: coder)
        setupView()
    }
    
    private func setupView() {
        // Metal 设置
        device = MTLCreateSystemDefaultDevice()
        clearColor = MTLClearColor(red: 0, green: 0, blue: 0, alpha: 0)
        
        // 渲染器
        petRenderer = PetRenderer(device: device!)
        delegate = petRenderer
        
        // 视图属性
        wantsLayer = true
        layer?.backgroundColor = NSColor.clear.cgColor
        
        // 开始动画
        startAnimation()
    }
    
    // MARK: - Animation
    func startAnimation() {
        animationTimer = Timer.scheduledTimer(withTimeInterval: 1.0/60.0, repeats: true) { _ in
            self.needsDisplay = true
        }
    }
    
    func stopAnimation() {
        animationTimer?.invalidate()
        animationTimer = nil
    }
    
    // MARK: - Pet Interaction
    func isPointInPet(_ point: NSPoint) -> Bool {
        // 检查点是否在宠物区域内
        let petBounds = NSRect(x: bounds.width * 0.25, y: bounds.height * 0.25,
                              width: bounds.width * 0.5, height: bounds.height * 0.5)
        return petBounds.contains(point)
    }
    
    func startInteraction() {
        currentAction = .interact
        petRenderer.setAction(.interact)
    }
    
    func endInteraction() {
        currentAction = nil
        petRenderer.setAction(nil)
    }
    
    func performAction(_ action: PetAction) {
        currentAction = action
        petRenderer.setAction(action)
        
        // 动作完成后恢复空闲状态
        DispatchQueue.main.asyncAfter(deadline: .now() + 2.0) {
            self.currentAction = nil
            self.petRenderer.setAction(nil)
        }
    }
    
    func updateState(_ state: PetState) {
        currentState = state
        petRenderer.setState(state)
    }
}

// MARK: - Pet Renderer
class PetRenderer: NSObject, MTKViewDelegate {
    
    // MARK: - Properties
    private let device: MTLDevice
    private let commandQueue: MTLCommandQueue
    private var pipelineState: MTLRenderPipelineState!
    private var vertexBuffer: MTLBuffer!
    private var uniformBuffer: MTLBuffer!
    
    private var time: Float = 0
    private var currentState: PetState = .idle
    private var currentAction: PetAction?
    
    // MARK: - Initialization
    init(device: MTLDevice) {
        self.device = device
        self.commandQueue = device.makeCommandQueue()!
        
        super.init()
        
        setupMetal()
        setupVertexData()
    }
    
    private func setupMetal() {
        let library = device.makeDefaultLibrary()!
        let vertexFunction = library.makeFunction(name: "vertex_pet")!
        let fragmentFunction = library.makeFunction(name: "fragment_pet")!
        
        let pipelineDescriptor = MTLRenderPipelineDescriptor()
        pipelineDescriptor.vertexFunction = vertexFunction
        pipelineDescriptor.fragmentFunction = fragmentFunction
        pipelineDescriptor.colorAttachments[0].pixelFormat = .bgra8Unorm
        
        // 启用混合
        pipelineDescriptor.colorAttachments[0].isBlendingEnabled = true
        pipelineDescriptor.colorAttachments[0].rgbBlendOperation = .add
        pipelineDescriptor.colorAttachments[0].alphaBlendOperation = .add
        pipelineDescriptor.colorAttachments[0].sourceRGBBlendFactor = .sourceAlpha
        pipelineDescriptor.colorAttachments[0].sourceAlphaBlendFactor = .sourceAlpha
        pipelineDescriptor.colorAttachments[0].destinationRGBBlendFactor = .oneMinusSourceAlpha
        pipelineDescriptor.colorAttachments[0].destinationAlphaBlendFactor = .oneMinusSourceAlpha
        
        do {
            pipelineState = try device.makeRenderPipelineState(descriptor: pipelineDescriptor)
        } catch {
            fatalError("Failed to create pipeline state: \(error)")
        }
        
        // 创建uniform缓冲区
        uniformBuffer = device.makeBuffer(length: MemoryLayout<PetUniforms>.size, options: [])
    }
    
    private func setupVertexData() {
        let vertices: [Float] = [
            -1.0, -1.0, 0.0, 1.0,  // 左下
             1.0, -1.0, 1.0, 1.0,  // 右下
            -1.0,  1.0, 0.0, 0.0,  // 左上
             1.0,  1.0, 1.0, 0.0   // 右上
        ]
        
        vertexBuffer = device.makeBuffer(bytes: vertices, length: vertices.count * MemoryLayout<Float>.size, options: [])
    }
    
    // MARK: - Public Methods
    func setState(_ state: PetState) {
        currentState = state
    }
    
    func setAction(_ action: PetAction?) {
        currentAction = action
    }
    
    // MARK: - MTKViewDelegate
    func mtkView(_ view: MTKView, drawableSizeWillChange size: CGSize) {
        // 处理视图大小变化
    }
    
    func draw(in view: MTKView) {
        guard let drawable = view.currentDrawable,
              let renderPassDescriptor = view.currentRenderPassDescriptor else { return }
        
        time += 1.0/60.0
        
        // 更新uniform数据
        var uniforms = PetUniforms()
        uniforms.time = time
        uniforms.resolution = simd_float2(Float(view.bounds.width), Float(view.bounds.height))
        uniforms.state = currentState.rawValue
        uniforms.action = currentAction?.rawValue ?? -1
        
        let uniformPointer = uniformBuffer.contents().bindMemory(to: PetUniforms.self, capacity: 1)
        uniformPointer.pointee = uniforms
        
        // 创建命令缓冲区
        let commandBuffer = commandQueue.makeCommandBuffer()!
        let renderEncoder = commandBuffer.makeRenderCommandEncoder(descriptor: renderPassDescriptor)!
        
        renderEncoder.setRenderPipelineState(pipelineState)
        renderEncoder.setVertexBuffer(vertexBuffer, offset: 0, index: 0)
        renderEncoder.setFragmentBuffer(uniformBuffer, offset: 0, index: 0)
        
        renderEncoder.drawPrimitives(type: .triangleStrip, vertexStart: 0, vertexCount: 4)
        renderEncoder.endEncoding()
        
        commandBuffer.present(drawable)
        commandBuffer.commit()
    }
}

// MARK: - Pet Data Models
enum PetState: Int32, CaseIterable {
    case idle = 0
    case happy = 1
    case sad = 2
    case sleeping = 3
    case eating = 4
    case playing = 5
}

enum PetAction: Int32, CaseIterable {
    case feed = 0
    case play = 1
    case rest = 2
    case interact = 3
}

struct PetUniforms {
    var time: Float = 0
    var resolution: simd_float2 = simd_float2(0, 0)
    var state: Int32 = 0
    var action: Int32 = -1
}

// MARK: - Desktop Pet Manager
class DesktopPetManager: NSObject {
    
    // MARK: - Properties
    static let shared = DesktopPetManager()
    
    private var petWindow: DesktopPetWindow?
    private var isVisible = false
    private var petState: PetState = .idle
    
    // MARK: - Initialization
    override init() {
        super.init()
        setupNotifications()
    }
    
    private func setupNotifications() {
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(showPet),
            name: .showDesktopPet,
            object: nil
        )
        
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(hidePet),
            name: .hideDesktopPet,
            object: nil
        )
        
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(togglePet),
            name: .toggleDesktopPet,
            object: nil
        )
    }
    
    // MARK: - Public Methods
    @objc func showPet() {
        if petWindow == nil {
            createPetWindow()
        }
        
        petWindow?.showPet()
        isVisible = true
    }
    
    @objc func hidePet() {
        petWindow?.hidePet()
        isVisible = false
    }
    
    @objc func togglePet() {
        if isVisible {
            hidePet()
        } else {
            showPet()
        }
    }
    
    func updatePetState(_ state: PetState) {
        petState = state
        petWindow?.updatePetState(state)
    }
    
    // MARK: - Private Methods
    private func createPetWindow() {
        let screenFrame = NSScreen.main?.frame ?? NSRect.zero
        let windowFrame = NSRect(x: screenFrame.width - 250,
                               y: screenFrame.height - 250,
                               width: 200,
                               height: 200)
        
        petWindow = DesktopPetWindow(
            contentRect: windowFrame,
            styleMask: .borderless,
            backing: .buffered,
            defer: false
        )
    }
    
    deinit {
        NotificationCenter.default.removeObserver(self)
    }
}