#include <metal_stdlib>
using namespace metal;

// 顶点输入结构
struct VertexIn {
    float3 position [[attribute(0)]];
    float4 color [[attribute(1)]];
    float2 texCoord [[attribute(2)]];
};

// 顶点输出结构
struct VertexOut {
    float4 position [[position]];
    float4 color;
    float2 texCoord;
    float time;
};

// Uniform结构
struct Uniforms {
    float4x4 projectionMatrix;
    float4x4 modelViewMatrix;
    float time;
    float2 resolution;
};

// 顶点着色器
vertex VertexOut vertex_main(VertexIn in [[stage_in]],
                            constant Uniforms& uniforms [[buffer(1)]]) {
    VertexOut out;
    
    // 应用变换矩阵
    float4 position = float4(in.position, 1.0);
    position = uniforms.modelViewMatrix * position;
    position = uniforms.projectionMatrix * position;
    
    out.position = position;
    out.color = in.color;
    out.texCoord = in.texCoord;
    out.time = uniforms.time;
    
    return out;
}

// 片段着色器 - 基础渲染
fragment float4 fragment_main(VertexOut in [[stage_in]],
                             constant Uniforms& uniforms [[buffer(0)]]) {
    // 基础颜色插值
    float4 color = in.color;
    
    // 添加时间动画效果
    float pulse = sin(in.time * 2.0) * 0.1 + 0.9;
    color.rgb *= pulse;
    
    // 添加渐变效果
    float2 uv = in.texCoord;
    float gradient = smoothstep(0.0, 1.0, uv.y);
    color.rgb = mix(color.rgb, color.rgb * 0.5, gradient);
    
    return color;
}

// 片段着色器 - 太极图案
fragment float4 fragment_taiji(VertexOut in [[stage_in]],
                              constant Uniforms& uniforms [[buffer(0)]]) {
    float2 uv = in.texCoord * 2.0 - 1.0; // 转换到[-1,1]范围
    float2 center = float2(0.0, 0.0);
    
    float dist = length(uv - center);
    
    // 外圆边界
    if (dist > 1.0) {
        discard_fragment();
    }
    
    // 太极图案计算
    float angle = atan2(uv.y, uv.x);
    float normalizedAngle = (angle + M_PI_F) / (2.0 * M_PI_F); // 归一化到[0,1]
    
    // 旋转动画
    float rotation = uniforms.time * 0.5;
    normalizedAngle += rotation / (2.0 * M_PI_F);
    normalizedAngle = fmod(normalizedAngle, 1.0);
    
    // 阴阳分界
    bool isYang = normalizedAngle < 0.5;
    
    // 小圆计算
    float2 yangCenter = float2(0.0, 0.5);
    float2 yinCenter = float2(0.0, -0.5);
    
    float yangDist = length(uv - yangCenter);
    float yinDist = length(uv - yinCenter);
    
    float4 yangColor = float4(1.0, 1.0, 1.0, 1.0); // 白色
    float4 yinColor = float4(0.1, 0.1, 0.1, 1.0);  // 黑色
    
    float4 finalColor;
    
    if (isYang) {
        // 阳区域（白色）
        if (yinDist < 0.25) {
            // 阴中有阳（小黑点）
            finalColor = yinColor;
        } else {
            finalColor = yangColor;
        }
    } else {
        // 阴区域（黑色）
        if (yangDist < 0.25) {
            // 阳中有阴（小白点）
            finalColor = yangColor;
        } else {
            finalColor = yinColor;
        }
    }
    
    // 添加边缘光晕效果
    float edgeFactor = 1.0 - smoothstep(0.9, 1.0, dist);
    finalColor.rgb *= edgeFactor;
    
    // 添加发光效果
    float glow = exp(-dist * 2.0) * 0.3;
    finalColor.rgb += float3(0.2, 0.4, 0.8) * glow;
    
    return finalColor;
}

// 片段着色器 - 粒子效果
fragment float4 fragment_particles(VertexOut in [[stage_in]],
                                  constant Uniforms& uniforms [[buffer(0)]]) {
    float2 uv = in.texCoord;
    float time = uniforms.time;
    
    float4 color = float4(0.0, 0.0, 0.0, 0.0);
    
    // 生成多个粒子
    for (int i = 0; i < 20; i++) {
        float fi = float(i);
        
        // 粒子位置动画
        float2 particlePos;
        particlePos.x = sin(time * 0.5 + fi * 0.3) * 0.3 + 0.5;
        particlePos.y = fmod(time * 0.2 + fi * 0.1, 1.0);
        
        // 粒子距离
        float dist = length(uv - particlePos);
        
        // 粒子大小和亮度
        float size = 0.02 + sin(time + fi) * 0.01;
        float brightness = 1.0 - smoothstep(0.0, size, dist);
        
        // 粒子颜色
        float3 particleColor = float3(
            0.5 + sin(time + fi) * 0.5,
            0.5 + sin(time + fi + 2.0) * 0.5,
            0.5 + sin(time + fi + 4.0) * 0.5
        );
        
        color.rgb += particleColor * brightness * 0.1;
        color.a += brightness * 0.1;
    }
    
    return color;
}

// 片段着色器 - 背景渐变
fragment float4 fragment_background(VertexOut in [[stage_in]],
                                   constant Uniforms& uniforms [[buffer(0)]]) {
    float2 uv = in.texCoord;
    float time = uniforms.time;
    
    // 创建动态渐变背景
    float3 color1 = float3(0.1, 0.2, 0.4); // 深蓝色
    float3 color2 = float3(0.8, 0.9, 1.0); // 浅蓝色
    float3 color3 = float3(0.9, 0.8, 0.6); // 金色
    
    // 时间动画
    float wave1 = sin(time * 0.3 + uv.x * 3.0) * 0.5 + 0.5;
    float wave2 = sin(time * 0.5 + uv.y * 2.0) * 0.5 + 0.5;
    
    // 混合颜色
    float3 finalColor = mix(color1, color2, uv.y);
    finalColor = mix(finalColor, color3, wave1 * wave2 * 0.3);
    
    // 添加噪声
    float noise = fract(sin(dot(uv * 100.0, float2(12.9898, 78.233))) * 43758.5453);
    finalColor += noise * 0.02;
    
    return float4(finalColor, 1.0);
}

// 计算函数 - 噪声
float noise(float2 p) {
    return fract(sin(dot(p, float2(12.9898, 78.233))) * 43758.5453);
}

// MARK: - Pet Shaders

struct PetUniforms {
    float time;
    float2 resolution;
    int state;
    int action;
};

vertex VertexOut vertex_pet(uint vertexID [[vertex_id]],
                           constant float4 *vertices [[buffer(0)]]) {
    VertexOut out;
    out.position = float4(vertices[vertexID].xy, 0.0, 1.0);
    out.texCoord = vertices[vertexID].zw;
    return out;
}

// Pet body shape function
float petBody(float2 p, float time) {
    // Main body (ellipse)
    float2 bodyPos = p - float2(0.0, -0.1);
    float body = length(bodyPos / float2(0.3, 0.4)) - 1.0;
    
    // Head (circle)
    float2 headPos = p - float2(0.0, 0.3);
    float head = length(headPos) - 0.25;
    
    // Combine body and head
    float pet = min(body, head);
    
    // Add some animation
    float bounce = sin(time * 3.0) * 0.02;
    pet += bounce;
    
    return pet;
}

// Pet eyes function
float2 petEyes(float2 p, float time, int state) {
    float2 leftEye = p - float2(-0.08, 0.35);
    float2 rightEye = p - float2(0.08, 0.35);
    
    float eyeSize = 0.04;
    
    // Blinking animation
    float blink = step(0.9, fract(time * 0.5));
    if (state == 3) { // sleeping
        blink = 1.0;
    }
    
    float leftEyeDist = length(leftEye) - eyeSize;
    float rightEyeDist = length(rightEye) - eyeSize;
    
    if (blink > 0.5) {
        leftEyeDist = abs(leftEye.y) - 0.01;
        rightEyeDist = abs(rightEye.y) - 0.01;
    }
    
    return float2(leftEyeDist, rightEyeDist);
}

// Pet mouth function
float petMouth(float2 p, int state, int action) {
    float2 mouthPos = p - float2(0.0, 0.25);
    
    // Different mouth shapes based on state
    float mouth;
    if (state == 1 || action == 1) { // happy or playing
        // Smile
        mouth = abs(length(mouthPos) - 0.08) - 0.02;
        mouth = max(mouth, mouthPos.y);
    } else if (state == 2) { // sad
        // Frown
        mouth = abs(length(mouthPos + float2(0.0, 0.05)) - 0.08) - 0.02;
        mouth = max(mouth, -mouthPos.y - 0.05);
    } else if (state == 4 || action == 0) { // eating or feeding
        // Open mouth
        float2 mouthEllipse = mouthPos / float2(0.06, 0.08);
        mouth = length(mouthEllipse) - 1.0;
    } else {
        // Neutral
        mouth = abs(mouthPos.y) - 0.01;
        mouth = max(mouth, abs(mouthPos.x) - 0.04);
    }
    
    return mouth;
}

// Pet ears function
float petEars(float2 p, float time) {
    // Left ear
    float2 leftEarPos = p - float2(-0.15, 0.45);
    float leftEar = length(leftEarPos / float2(0.8, 1.2)) - 0.15;
    
    // Right ear
    float2 rightEarPos = p - float2(0.15, 0.45);
    float rightEar = length(rightEarPos / float2(0.8, 1.2)) - 0.15;
    
    // Ear wiggle animation
    float wiggle = sin(time * 4.0) * 0.01;
    leftEar += wiggle;
    rightEar -= wiggle;
    
    return min(leftEar, rightEar);
}

// Pet color function
float3 petColor(float2 p, float time, int state, int action) {
    float3 baseColor = float3(1.0, 0.8, 0.6); // Light orange
    
    // Add some patterns
    float pattern = sin(p.x * 10.0) * sin(p.y * 10.0) * 0.1;
    baseColor += pattern * float3(0.1, 0.05, 0.0);
    
    // State-based color modifications
    if (state == 1 || action == 1) { // happy or playing
        baseColor += float3(0.2, 0.1, 0.0); // More vibrant
    } else if (state == 2) { // sad
        baseColor *= 0.7; // Darker
    } else if (state == 3) { // sleeping
        baseColor *= 0.8; // Slightly darker
    }
    
    // Action-based effects
    if (action == 3) { // interact
        float glow = sin(time * 8.0) * 0.2 + 0.8;
        baseColor *= glow;
    }
    
    return baseColor;
}

// Sparkle effects
float sparkles(float2 p, float time, int action) {
    if (action != 1 && action != 3) return 0.0; // Only for play and interact
    
    float sparkle = 0.0;
    
    for (int i = 0; i < 5; i++) {
        float2 sparklePos = float2(
            sin(time * 2.0 + float(i)) * 0.4,
            cos(time * 1.5 + float(i)) * 0.4
        );
        
        float dist = length(p - sparklePos);
        float size = 0.02 + sin(time * 6.0 + float(i)) * 0.01;
        sparkle += smoothstep(size, 0.0, dist);
    }
    
    return sparkle;
}

fragment float4 fragment_pet(VertexOut in [[stage_in]],
                            constant PetUniforms &uniforms [[buffer(0)]]) {
    float2 uv = in.texCoord;
    float2 p = (uv - 0.5) * 2.0;
    p.x *= uniforms.resolution.x / uniforms.resolution.y;
    
    float time = uniforms.time;
    int state = uniforms.state;
    int action = uniforms.action;
    
    // Pet body
    float petDist = petBody(p, time);
    
    // Pet ears
    float earsDist = petEars(p, time);
    
    // Combine body and ears
    float petShape = min(petDist, earsDist);
    
    // Pet eyes
    float2 eyesDist = petEyes(p, time, state);
    
    // Pet mouth
    float mouthDist = petMouth(p, state, action);
    
    // Base pet color
    float3 color = float3(0.0);
    
    if (petShape < 0.0) {
        color = petColor(p, time, state, action);
        
        // Add eyes
        if (eyesDist.x < 0.0 || eyesDist.y < 0.0) {
            color = float3(0.0); // Black eyes
        }
        
        // Add mouth
        if (mouthDist < 0.0) {
            if (state == 4 || action == 0) { // eating
                color = float3(0.2, 0.1, 0.1); // Dark mouth
            } else {
                color = float3(0.8, 0.4, 0.4); // Pink mouth
            }
        }
    }
    
    // Add sparkle effects
    float sparkleEffect = sparkles(p, time, action);
    color += sparkleEffect * float3(1.0, 1.0, 0.8);
    
    // Soft edges
    float alpha = 1.0 - smoothstep(-0.02, 0.02, petShape);
    
    // Add glow effect for interaction
    if (action == 3) { // interact
        float glow = exp(-length(p) * 2.0) * 0.3;
        color += glow * float3(1.0, 0.8, 0.6);
    }
    
    return float4(color, alpha);
}

// 计算函数 - 分形噪声
float fractalNoise(float2 p, int octaves) {
    float value = 0.0;
    float amplitude = 0.5;
    float frequency = 1.0;
    
    for (int i = 0; i < octaves; i++) {
        value += amplitude * noise(p * frequency);
        amplitude *= 0.5;
        frequency *= 2.0;
    }
    
    return value;
}