import Metal
import MetalKit
import simd

class MetalRenderer: NSObject, MTKViewDelegate {
    
    // MARK: - Properties
    private let device: MTLDevice
    private let commandQueue: MTLCommandQueue
    private var renderPipelineState: MTLRenderPipelineState?
    private var vertexBuffer: MTLBuffer?
    private var uniformBuffer: MTLBuffer?
    
    // 渲染参数
    private var viewportSize: vector_uint2 = vector_uint2(0, 0)
    private var time: Float = 0.0
    
    // 顶点数据结构
    struct Vertex {
        var position: vector_float3
        var color: vector_float4
        var texCoord: vector_float2
    }
    
    // Uniform数据结构
    struct Uniforms {
        var projectionMatrix: matrix_float4x4
        var modelViewMatrix: matrix_float4x4
        var time: Float
        var resolution: vector_float2
    }
    
    // MARK: - Initialization
    
    init?(metalView: MTKView) {
        guard let device = metalView.device else {
            NSLog("Metal设备不可用")
            return nil
        }
        
        self.device = device
        
        guard let commandQueue = device.makeCommandQueue() else {
            NSLog("无法创建Metal命令队列")
            return nil
        }
        
        self.commandQueue = commandQueue
        
        super.init()
        
        metalView.delegate = self
        metalView.device = device
        
        setupMetal()
        setupVertexData()
    }
    
    // MARK: - Metal Setup
    
    private func setupMetal() {
        // 创建着色器库
        guard let library = device.makeDefaultLibrary() else {
            NSLog("无法创建着色器库")
            return
        }
        
        // 获取顶点和片段着色器函数
        guard let vertexFunction = library.makeFunction(name: "vertex_main"),
              let fragmentFunction = library.makeFunction(name: "fragment_main") else {
            NSLog("无法获取着色器函数")
            return
        }
        
        // 创建渲染管线描述符
        let pipelineDescriptor = MTLRenderPipelineDescriptor()
        pipelineDescriptor.vertexFunction = vertexFunction
        pipelineDescriptor.fragmentFunction = fragmentFunction
        pipelineDescriptor.colorAttachments[0].pixelFormat = .bgra8Unorm
        
        // 设置顶点描述符
        let vertexDescriptor = MTLVertexDescriptor()
        
        // 位置属性
        vertexDescriptor.attributes[0].format = .float3
        vertexDescriptor.attributes[0].offset = 0
        vertexDescriptor.attributes[0].bufferIndex = 0
        
        // 颜色属性
        vertexDescriptor.attributes[1].format = .float4
        vertexDescriptor.attributes[1].offset = MemoryLayout<vector_float3>.stride
        vertexDescriptor.attributes[1].bufferIndex = 0
        
        // 纹理坐标属性
        vertexDescriptor.attributes[2].format = .float2
        vertexDescriptor.attributes[2].offset = MemoryLayout<vector_float3>.stride + MemoryLayout<vector_float4>.stride
        vertexDescriptor.attributes[2].bufferIndex = 0
        
        // 缓冲区布局
        vertexDescriptor.layouts[0].stride = MemoryLayout<Vertex>.stride
        vertexDescriptor.layouts[0].stepRate = 1
        vertexDescriptor.layouts[0].stepFunction = .perVertex
        
        pipelineDescriptor.vertexDescriptor = vertexDescriptor
        
        // 创建渲染管线状态
        do {
            renderPipelineState = try device.makeRenderPipelineState(descriptor: pipelineDescriptor)
        } catch {
            NSLog("创建渲染管线状态失败: \\(error)")
        }
        
        // 创建uniform缓冲区
        uniformBuffer = device.makeBuffer(length: MemoryLayout<Uniforms>.stride, options: [])
    }
    
    private func setupVertexData() {
        // 创建一个简单的四边形
        let vertices: [Vertex] = [
            // 第一个三角形
            Vertex(position: vector_float3(-0.5, -0.5, 0.0), color: vector_float4(1.0, 0.0, 0.0, 1.0), texCoord: vector_float2(0.0, 1.0)),
            Vertex(position: vector_float3(0.5, -0.5, 0.0), color: vector_float4(0.0, 1.0, 0.0, 1.0), texCoord: vector_float2(1.0, 1.0)),
            Vertex(position: vector_float3(-0.5, 0.5, 0.0), color: vector_float4(0.0, 0.0, 1.0, 1.0), texCoord: vector_float2(0.0, 0.0)),
            
            // 第二个三角形
            Vertex(position: vector_float3(0.5, -0.5, 0.0), color: vector_float4(0.0, 1.0, 0.0, 1.0), texCoord: vector_float2(1.0, 1.0)),
            Vertex(position: vector_float3(0.5, 0.5, 0.0), color: vector_float4(1.0, 1.0, 0.0, 1.0), texCoord: vector_float2(1.0, 0.0)),
            Vertex(position: vector_float3(-0.5, 0.5, 0.0), color: vector_float4(0.0, 0.0, 1.0, 1.0), texCoord: vector_float2(0.0, 0.0))
        ]
        
        vertexBuffer = device.makeBuffer(bytes: vertices, length: vertices.count * MemoryLayout<Vertex>.stride, options: [])
    }
    
    // MARK: - Matrix Utilities
    
    private func makeOrthographicMatrix(left: Float, right: Float, bottom: Float, top: Float, near: Float, far: Float) -> matrix_float4x4 {
        let ral = right + left
        let rsl = right - left
        let tab = top + bottom
        let tsb = top - bottom
        let fan = far + near
        let fsn = far - near
        
        return matrix_float4x4(columns: (
            vector_float4(2.0 / rsl, 0.0, 0.0, 0.0),
            vector_float4(0.0, 2.0 / tsb, 0.0, 0.0),
            vector_float4(0.0, 0.0, -2.0 / fsn, 0.0),
            vector_float4(-ral / rsl, -tab / tsb, -fan / fsn, 1.0)
        ))
    }
    
    private func makeTranslationMatrix(x: Float, y: Float, z: Float) -> matrix_float4x4 {
        return matrix_float4x4(columns: (
            vector_float4(1.0, 0.0, 0.0, 0.0),
            vector_float4(0.0, 1.0, 0.0, 0.0),
            vector_float4(0.0, 0.0, 1.0, 0.0),
            vector_float4(x, y, z, 1.0)
        ))
    }
    
    private func makeRotationMatrix(angle: Float, axis: vector_float3) -> matrix_float4x4 {
        let c = cos(angle)
        let s = sin(angle)
        let x = axis.x
        let y = axis.y
        let z = axis.z
        
        return matrix_float4x4(columns: (
            vector_float4(c + x*x*(1-c), y*x*(1-c) + z*s, z*x*(1-c) - y*s, 0.0),
            vector_float4(x*y*(1-c) - z*s, c + y*y*(1-c), z*y*(1-c) + x*s, 0.0),
            vector_float4(x*z*(1-c) + y*s, y*z*(1-c) - x*s, c + z*z*(1-c), 0.0),
            vector_float4(0.0, 0.0, 0.0, 1.0)
        ))
    }
    
    // MARK: - Public Methods
    
    func updateFrame() {
        time += 1.0 / 60.0 // 假设60FPS
    }
    
    func updateViewport() {
        // 视口更新逻辑
    }
    
    // MARK: - MTKViewDelegate
    
    func mtkView(_ view: MTKView, drawableSizeWillChange size: CGSize) {
        viewportSize.x = UInt32(size.width)
        viewportSize.y = UInt32(size.height)
    }
    
    func draw(in view: MTKView) {
        guard let renderPipelineState = renderPipelineState,
              let vertexBuffer = vertexBuffer,
              let uniformBuffer = uniformBuffer,
              let drawable = view.currentDrawable,
              let renderPassDescriptor = view.currentRenderPassDescriptor else {
            return
        }
        
        // 更新uniform数据
        updateUniforms()
        
        // 创建命令缓冲区
        guard let commandBuffer = commandQueue.makeCommandBuffer() else { return }
        
        // 创建渲染编码器
        guard let renderEncoder = commandBuffer.makeRenderCommandEncoder(descriptor: renderPassDescriptor) else { return }
        
        // 设置渲染管线状态
        renderEncoder.setRenderPipelineState(renderPipelineState)
        
        // 设置顶点缓冲区
        renderEncoder.setVertexBuffer(vertexBuffer, offset: 0, index: 0)
        renderEncoder.setVertexBuffer(uniformBuffer, offset: 0, index: 1)
        
        // 设置片段缓冲区
        renderEncoder.setFragmentBuffer(uniformBuffer, offset: 0, index: 0)
        
        // 绘制
        renderEncoder.drawPrimitives(type: .triangle, vertexStart: 0, vertexCount: 6)
        
        // 结束编码
        renderEncoder.endEncoding()
        
        // 提交命令缓冲区
        commandBuffer.present(drawable)
        commandBuffer.commit()
        
        // 更新时间
        updateFrame()
    }
    
    private func updateUniforms() {
        guard let uniformBuffer = uniformBuffer else { return }
        
        let aspect = Float(viewportSize.x) / Float(viewportSize.y)
        let projectionMatrix = makeOrthographicMatrix(left: -aspect, right: aspect, bottom: -1.0, top: 1.0, near: -1.0, far: 1.0)
        
        let rotationAngle = time * 0.5
        let rotationMatrix = makeRotationMatrix(angle: rotationAngle, axis: vector_float3(0.0, 0.0, 1.0))
        
        var uniforms = Uniforms(
            projectionMatrix: projectionMatrix,
            modelViewMatrix: rotationMatrix,
            time: time,
            resolution: vector_float2(Float(viewportSize.x), Float(viewportSize.y))
        )
        
        let uniformBufferPointer = uniformBuffer.contents().bindMemory(to: Uniforms.self, capacity: 1)
        uniformBufferPointer.pointee = uniforms
    }
}