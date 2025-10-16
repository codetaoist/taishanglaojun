# 太上老君AI平台故障排除指南

## 概述

本文档提供了太上老君AI平台常见问题的解决方案和故障排除步骤，帮助用户快速解决使用过程中遇到的问题。

## 常见问题分类

### 1. 登录和认证问题

#### 无法登录

**问题描述**: 输入正确的用户名和密码后仍无法登录

**可能原因**:
- 网络连接问题
- 账户被锁定
- 密码已过期
- 浏览器缓存问题

**解决步骤**:

1. **检查网络连接**
   ```bash
   # 测试网络连接
   ping taishanglaojun.ai
   ```

2. **清除浏览器缓存**
   - Chrome: Ctrl+Shift+Delete
   - Firefox: Ctrl+Shift+Delete
   - Safari: Cmd+Option+E

3. **尝试重置密码**
   - 点击"忘记密码"链接
   - 输入注册邮箱
   - 检查邮件并按提示重置

4. **联系技术支持**
   - 如果问题持续存在，请联系 support@taishanglaojun.ai

#### API密钥无效

**问题描述**: 使用API时收到"INVALID_API_KEY"错误

**解决步骤**:

1. **检查API密钥格式**
   ```javascript
   // 正确格式
   const apiKey = "sk-1234567890abcdef...";
   
   // 错误格式
   const apiKey = "1234567890abcdef..."; // 缺少sk-前缀
   ```

2. **验证API密钥状态**
   ```bash
   curl -H "Authorization: Bearer YOUR_API_KEY" \
        https://api.taishanglaojun.ai/v1/auth/permissions
   ```

3. **重新生成API密钥**
   - 登录用户面板
   - 进入"API设置"
   - 删除旧密钥并创建新密钥

#### 两步验证问题

**问题描述**: 两步验证码无效或无法接收

**解决步骤**:

1. **检查时间同步**
   - 确保设备时间正确
   - 同步时间服务器

2. **重新同步验证器**
   - 删除旧的验证器条目
   - 重新扫描二维码设置

3. **使用备用代码**
   - 使用设置时保存的备用代码
   - 每个备用代码只能使用一次

### 2. 聊天功能问题

#### 消息发送失败

**问题描述**: 点击发送后消息无法发送

**可能原因**:
- 网络连接不稳定
- 消息内容违规
- 达到使用限制
- 服务器临时故障

**解决步骤**:

1. **检查网络状态**
   ```javascript
   // 检查网络连接
   if (navigator.onLine) {
     console.log("网络连接正常");
   } else {
     console.log("网络连接异常");
   }
   ```

2. **检查消息内容**
   - 避免敏感或违规内容
   - 检查消息长度限制
   - 移除特殊字符

3. **检查使用配额**
   - 查看账户使用统计
   - 升级套餐或等待配额重置

4. **重试发送**
   - 等待几秒后重试
   - 刷新页面后重试

#### 响应速度慢

**问题描述**: AI回复速度异常缓慢

**解决步骤**:

1. **检查网络延迟**
   ```bash
   # 测试延迟
   ping -c 4 api.taishanglaojun.ai
   ```

2. **优化请求参数**
   ```javascript
   // 减少max_tokens以提高响应速度
   const response = await fetch('/api/chat', {
     method: 'POST',
     body: JSON.stringify({
       content: "你的问题",
       max_tokens: 500, // 减少token数量
       temperature: 0.7
     })
   });
   ```

3. **使用流式响应**
   ```javascript
   // 启用流式响应获得更快的感知速度
   const response = await fetch('/api/chat', {
     method: 'POST',
     body: JSON.stringify({
       content: "你的问题",
       stream: true
     })
   });
   ```

#### 对话历史丢失

**问题描述**: 之前的对话记录消失

**解决步骤**:

1. **检查浏览器存储**
   ```javascript
   // 检查本地存储
   console.log(localStorage.getItem('conversations'));
   console.log(sessionStorage.getItem('current_conversation'));
   ```

2. **刷新对话列表**
   - 点击刷新按钮
   - 重新登录账户

3. **检查账户同步**
   - 确认已登录正确账户
   - 检查多设备同步设置

### 3. 图像生成问题

#### 图像生成失败

**问题描述**: 提交图像生成请求后失败

**可能原因**:
- 提示词包含违规内容
- 网络上传问题
- 服务器负载过高
- 达到生成限制

**解决步骤**:

1. **检查提示词内容**
   ```javascript
   // 避免的内容类型
   const prohibitedContent = [
     "暴力", "色情", "政治敏感",
     "版权内容", "真实人物"
   ];
   
   // 优化提示词
   const optimizedPrompt = "一只可爱的卡通猫咪，坐在花园里，阳光明媚，插画风格";
   ```

2. **简化提示词**
   ```javascript
   // 过于复杂的提示词
   const complexPrompt = "一只穿着红色衣服的猫咪坐在蓝色的椅子上，背景是绿色的花园，天空是橙色的，还有紫色的云朵...";
   
   // 简化后的提示词
   const simplePrompt = "一只红衣猫咪坐在花园椅子上";
   ```

3. **调整生成参数**
   ```javascript
   const params = {
     prompt: "你的提示词",
     size: "512x512", // 使用较小尺寸提高成功率
     quality: "standard", // 使用标准质量
     n: 1 // 减少生成数量
   };
   ```

#### 图像质量不佳

**问题描述**: 生成的图像质量不符合预期

**解决步骤**:

1. **优化提示词**
   ```javascript
   // 添加质量描述词
   const qualityPrompt = "高质量，4K，超详细，专业摄影，最佳质量";
   
   // 完整提示词
   const fullPrompt = `一只可爱的小猫，${qualityPrompt}`;
   ```

2. **使用负面提示词**
   ```javascript
   const negativePrompt = "模糊，低质量，变形，噪点，水印";
   ```

3. **调整生成参数**
   ```javascript
   const params = {
     prompt: "你的提示词",
     size: "1024x1024", // 使用更高分辨率
     quality: "hd", // 使用高清质量
     style: "vivid" // 使用生动风格
   };
   ```

#### 图像加载失败

**问题描述**: 生成的图像无法正常显示

**解决步骤**:

1. **检查图像URL**
   ```javascript
   // 验证图像URL
   const img = new Image();
   img.onload = () => console.log("图像加载成功");
   img.onerror = () => console.log("图像加载失败");
   img.src = imageUrl;
   ```

2. **检查网络连接**
   ```javascript
   // 测试CDN连接
   fetch('https://cdn.taishanglaojun.ai/test')
     .then(response => console.log('CDN连接正常'))
     .catch(error => console.log('CDN连接异常'));
   ```

3. **清除缓存重试**
   - 强制刷新页面 (Ctrl+F5)
   - 清除浏览器缓存

### 4. 文档分析问题

#### 文档上传失败

**问题描述**: 无法成功上传文档文件

**可能原因**:
- 文件格式不支持
- 文件大小超限
- 网络连接问题
- 存储空间不足

**解决步骤**:

1. **检查文件格式**
   ```javascript
   const supportedFormats = [
     'application/pdf',
     'application/msword',
     'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
     'text/plain'
   ];
   
   function validateFileType(file) {
     return supportedFormats.includes(file.type);
   }
   ```

2. **检查文件大小**
   ```javascript
   const maxSize = 50 * 1024 * 1024; // 50MB
   
   function validateFileSize(file) {
     if (file.size > maxSize) {
       alert('文件大小不能超过50MB');
       return false;
     }
     return true;
   }
   ```

3. **压缩文件**
   - 使用PDF压缩工具
   - 降低图片质量
   - 移除不必要的内容

#### 文档分析结果不准确

**问题描述**: 分析结果与文档内容不符

**解决步骤**:

1. **检查文档质量**
   - 确保文字清晰可读
   - 避免扫描质量差的文档
   - 使用标准字体

2. **优化文档格式**
   ```javascript
   // 推荐的文档格式设置
   const documentSettings = {
     format: 'PDF',
     textLayer: true, // 确保有文本层
     resolution: '300dpi', // 高分辨率
     compression: 'lossless' // 无损压缩
   };
   ```

3. **分段分析**
   - 将大文档分成小段
   - 针对特定部分进行分析

#### 分析速度慢

**问题描述**: 文档分析处理时间过长

**解决步骤**:

1. **优化文档大小**
   - 减少页数
   - 压缩图片
   - 移除不必要的元素

2. **使用异步处理**
   ```javascript
   // 检查处理状态
   async function checkAnalysisStatus(documentId) {
     const response = await fetch(`/api/documents/${documentId}/status`);
     const data = await response.json();
     
     if (data.status === 'processing') {
       setTimeout(() => checkAnalysisStatus(documentId), 5000);
     } else {
       console.log('分析完成');
     }
   }
   ```

### 5. 性能问题

#### 页面加载慢

**问题描述**: 网站页面加载速度缓慢

**解决步骤**:

1. **检查网络速度**
   ```bash
   # 测试下载速度
   curl -o /dev/null -s -w "%{time_total}\n" https://taishanglaojun.ai
   ```

2. **优化浏览器**
   - 关闭不必要的标签页
   - 禁用不需要的扩展
   - 更新浏览器版本

3. **清理缓存**
   ```javascript
   // 清理应用缓存
   if ('caches' in window) {
     caches.keys().then(names => {
       names.forEach(name => caches.delete(name));
     });
   }
   ```

#### 内存使用过高

**问题描述**: 浏览器内存占用过高

**解决步骤**:

1. **监控内存使用**
   ```javascript
   // 检查内存使用情况
   if (performance.memory) {
     console.log('已使用内存:', performance.memory.usedJSHeapSize);
     console.log('总内存限制:', performance.memory.totalJSHeapSize);
   }
   ```

2. **优化使用习惯**
   - 定期关闭不用的对话
   - 清理历史记录
   - 避免同时处理大量文档

3. **重启浏览器**
   - 定期重启浏览器
   - 清理临时文件

### 6. 移动端问题

#### 移动应用崩溃

**问题描述**: 移动应用频繁崩溃或闪退

**解决步骤**:

1. **更新应用版本**
   - 检查应用商店更新
   - 安装最新版本

2. **重启设备**
   - 完全关机重启
   - 清理后台应用

3. **清理应用数据**
   ```bash
   # Android清理应用数据
   adb shell pm clear com.taishanglaojun.app
   ```

4. **重新安装应用**
   - 卸载当前版本
   - 重新下载安装

#### 语音输入不工作

**问题描述**: 语音输入功能无响应

**解决步骤**:

1. **检查权限设置**
   - 确认麦克风权限已开启
   - 检查应用权限设置

2. **测试麦克风**
   ```javascript
   // 测试麦克风访问
   navigator.mediaDevices.getUserMedia({ audio: true })
     .then(stream => console.log('麦克风正常'))
     .catch(error => console.log('麦克风异常:', error));
   ```

3. **检查网络连接**
   - 确保网络连接稳定
   - 尝试切换网络

### 7. 账户和计费问题

#### 配额用完

**问题描述**: 收到配额已用完的提示

**解决步骤**:

1. **查看使用统计**
   ```javascript
   // 获取使用统计
   fetch('/api/users/me/usage')
     .then(response => response.json())
     .then(data => console.log('使用情况:', data));
   ```

2. **升级套餐**
   - 访问账户设置
   - 选择合适的套餐
   - 完成支付升级

3. **等待配额重置**
   - 查看配额重置时间
   - 合理安排使用时间

#### 支付问题

**问题描述**: 无法完成支付或支付失败

**解决步骤**:

1. **检查支付信息**
   - 验证信用卡信息
   - 确认账户余额充足
   - 检查卡片有效期

2. **尝试其他支付方式**
   - 使用不同的信用卡
   - 尝试PayPal支付
   - 联系银行确认

3. **联系客服**
   - 提供订单号
   - 描述具体问题
   - 等待人工处理

## 诊断工具

### 1. 网络诊断

```bash
#!/bin/bash
# 网络诊断脚本

echo "=== 网络连接诊断 ==="

# 测试DNS解析
echo "测试DNS解析..."
nslookup taishanglaojun.ai

# 测试连接延迟
echo "测试连接延迟..."
ping -c 4 taishanglaojun.ai

# 测试HTTPS连接
echo "测试HTTPS连接..."
curl -I https://taishanglaojun.ai

# 测试API连接
echo "测试API连接..."
curl -I https://api.taishanglaojun.ai/v1/health
```

### 2. 浏览器诊断

```javascript
// 浏览器环境诊断
function diagnoseBrowser() {
  const info = {
    userAgent: navigator.userAgent,
    language: navigator.language,
    cookieEnabled: navigator.cookieEnabled,
    onLine: navigator.onLine,
    platform: navigator.platform,
    screen: {
      width: screen.width,
      height: screen.height,
      colorDepth: screen.colorDepth
    },
    localStorage: typeof(Storage) !== "undefined",
    sessionStorage: typeof(Storage) !== "undefined",
    webGL: !!window.WebGLRenderingContext,
    webWorker: typeof(Worker) !== "undefined"
  };
  
  console.log('浏览器诊断信息:', info);
  return info;
}

// 性能诊断
function diagnosePerformance() {
  if (performance.memory) {
    const memory = {
      used: Math.round(performance.memory.usedJSHeapSize / 1024 / 1024),
      total: Math.round(performance.memory.totalJSHeapSize / 1024 / 1024),
      limit: Math.round(performance.memory.jsHeapSizeLimit / 1024 / 1024)
    };
    console.log('内存使用情况 (MB):', memory);
  }
  
  const timing = performance.timing;
  const loadTime = timing.loadEventEnd - timing.navigationStart;
  console.log('页面加载时间:', loadTime + 'ms');
}
```

### 3. API诊断

```javascript
// API连接诊断
async function diagnoseAPI() {
  const tests = [
    { name: '健康检查', url: '/api/health' },
    { name: '认证检查', url: '/api/auth/permissions' },
    { name: '用户信息', url: '/api/users/me' }
  ];
  
  for (const test of tests) {
    try {
      const start = Date.now();
      const response = await fetch(test.url);
      const end = Date.now();
      
      console.log(`${test.name}: ${response.status} (${end - start}ms)`);
    } catch (error) {
      console.error(`${test.name}: 失败 -`, error.message);
    }
  }
}
```

## 日志收集

### 1. 浏览器控制台日志

```javascript
// 收集控制台日志
function collectLogs() {
  const logs = [];
  const originalLog = console.log;
  const originalError = console.error;
  const originalWarn = console.warn;
  
  console.log = function(...args) {
    logs.push({ type: 'log', message: args.join(' '), timestamp: new Date() });
    originalLog.apply(console, args);
  };
  
  console.error = function(...args) {
    logs.push({ type: 'error', message: args.join(' '), timestamp: new Date() });
    originalError.apply(console, args);
  };
  
  console.warn = function(...args) {
    logs.push({ type: 'warn', message: args.join(' '), timestamp: new Date() });
    originalWarn.apply(console, args);
  };
  
  return logs;
}
```

### 2. 错误监控

```javascript
// 全局错误监控
window.addEventListener('error', function(event) {
  const errorInfo = {
    message: event.message,
    filename: event.filename,
    lineno: event.lineno,
    colno: event.colno,
    error: event.error,
    timestamp: new Date()
  };
  
  console.error('JavaScript错误:', errorInfo);
  
  // 发送错误报告
  fetch('/api/errors', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(errorInfo)
  });
});

// Promise错误监控
window.addEventListener('unhandledrejection', function(event) {
  const errorInfo = {
    reason: event.reason,
    promise: event.promise,
    timestamp: new Date()
  };
  
  console.error('未处理的Promise错误:', errorInfo);
});
```

## 联系技术支持

### 支持渠道

1. **在线客服**
   - 网站右下角聊天窗口
   - 工作时间: 9:00-18:00 (工作日)

2. **邮件支持**
   - 技术支持: support@taishanglaojun.ai
   - 商务咨询: sales@taishanglaojun.ai

3. **电话支持**
   - 客服热线: 400-123-4567
   - 技术热线: 400-123-4568

### 提交问题时请提供

1. **基本信息**
   - 用户ID或邮箱
   - 问题发生时间
   - 使用的设备和浏览器

2. **问题描述**
   - 详细的问题描述
   - 重现步骤
   - 期望的结果

3. **技术信息**
   - 错误消息截图
   - 浏览器控制台日志
   - 网络请求信息

4. **诊断结果**
   - 运行诊断工具的结果
   - 网络连接测试结果

### 问题优先级

| 优先级 | 描述 | 响应时间 |
|--------|------|----------|
| 紧急 | 系统完全无法使用 | 1小时内 |
| 高 | 核心功能异常 | 4小时内 |
| 中 | 部分功能问题 | 1个工作日内 |
| 低 | 一般咨询或建议 | 3个工作日内 |

---

© 2024 太上老君AI平台. 保留所有权利。