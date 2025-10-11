# 安全服务 API 文档

## 概述

安全服务提供全面的网络安全防护功能，包括威胁检测、漏洞扫描、安全监控、报告生成等核心安全能力。

## 基础信息

- **服务名称**: Security Service
- **版本**: v1.0.0
- **基础路径**: `/api/security`
- **认证方式**: Bearer Token

## API 端点

### 1. 威胁检测 API

#### 1.1 获取威胁统计
```http
GET /api/security/threats/statistics
```

**查询参数:**
- `timeRange` (string, optional): 时间范围，默认 "24h"

**响应示例:**
```json
{
  "success": true,
  "data": {
    "total_alerts": 156,
    "active_alerts": 23,
    "alerts_by_severity": {
      "critical": 5,
      "high": 18,
      "medium": 45,
      "low": 88
    },
    "recent_alerts": [...],
    "blocked_ips": 12,
    "active_rules": 45
  }
}
```

#### 1.2 获取最近告警
```http
GET /api/security/threats/alerts/recent
```

**查询参数:**
- `limit` (int, optional): 返回数量限制，默认 10

#### 1.3 更新威胁告警
```http
PUT /api/security/threats/alerts/{id}
```

**请求体:**
```json
{
  "status": "resolved",
  "notes": "已处理该威胁"
}
```

### 2. 漏洞扫描 API

#### 2.1 创建扫描任务
```http
POST /api/security/vulnerabilities/scan
```

**请求体:**
```json
{
  "name": "Web应用扫描",
  "type": "web",
  "target": "https://example.com",
  "config": {
    "depth": 3,
    "timeout": 300,
    "concurrent": 10
  }
}
```

#### 2.2 获取扫描任务列表
```http
GET /api/security/vulnerabilities/scans
```

**查询参数:**
- `page` (int, optional): 页码，默认 1
- `limit` (int, optional): 每页数量，默认 20
- `status` (string, optional): 状态过滤

#### 2.3 获取扫描结果
```http
GET /api/security/vulnerabilities/scans/{id}/results
```

#### 2.4 获取漏洞建议
```http
GET /api/security/vulnerabilities/{id}/recommendations
```

#### 2.5 安排定期扫描
```http
POST /api/security/vulnerabilities/schedule
```

**请求体:**
```json
{
  "name": "每日安全扫描",
  "scan_type": "web",
  "target": "https://example.com",
  "interval": "daily",
  "config": {...}
}
```

#### 2.6 导出漏洞报告
```http
GET /api/security/vulnerabilities/export
```

**查询参数:**
- `format` (string): 导出格式 (json, csv, pdf)
- `scan_id` (string, optional): 特定扫描ID

### 3. 渗透测试 API

#### 3.1 创建渗透测试项目
```http
POST /api/security/pentest/projects
```

#### 3.2 获取项目列表
```http
GET /api/security/pentest/projects
```

#### 3.3 获取测试结果
```http
GET /api/security/pentest/projects/{id}/results
```

### 4. 安全教育 API

#### 4.1 获取课程列表
```http
GET /api/security/education/courses
```

#### 4.2 获取实验环境
```http
GET /api/security/education/labs
```

#### 4.3 获取认证信息
```http
GET /api/security/education/certifications
```

### 5. 安全审计 API

#### 5.1 获取审计日志
```http
GET /api/security/audit/logs
```

**查询参数:**
- `start_time` (string): 开始时间
- `end_time` (string): 结束时间
- `event_type` (string, optional): 事件类型
- `user_id` (string, optional): 用户ID

#### 5.2 获取审计报告
```http
GET /api/security/audit/reports
```

#### 5.3 生成审计报告
```http
POST /api/security/audit/reports
```

#### 5.4 获取合规状态
```http
GET /api/security/audit/compliance
```

### 6. 安全监控 API

#### 6.1 获取安全指标
```http
GET /api/security/monitoring/metrics
```

**查询参数:**
- `timeRange` (string, optional): 时间范围，默认 "24h"

**响应示例:**
```json
{
  "success": true,
  "data": {
    "total_events": 1250,
    "threat_alerts": 45,
    "blocked_requests": 123,
    "active_sessions": 89,
    "top_source_ips": [...],
    "threat_trends": [...],
    "attack_vectors": {...},
    "geographic_data": {...}
  }
}
```

#### 6.2 获取告警列表
```http
GET /api/security/monitoring/alerts
```

**查询参数:**
- `page` (int, optional): 页码
- `limit` (int, optional): 每页数量
- `severity` (string, optional): 严重程度过滤
- `status` (string, optional): 状态过滤

#### 6.3 确认告警
```http
POST /api/security/monitoring/alerts/{id}/acknowledge
```

#### 6.4 获取监控仪表板
```http
GET /api/security/monitoring/dashboard
```

#### 6.5 获取健康状态
```http
GET /api/security/monitoring/health
```

### 7. 安全报告 API

#### 7.1 获取报告列表
```http
GET /api/security/reports
```

**查询参数:**
- `page` (int, optional): 页码
- `limit` (int, optional): 每页数量
- `type` (string, optional): 报告类型
- `status` (string, optional): 状态

#### 7.2 生成报告
```http
POST /api/security/reports/generate
```

**请求体:**
```json
{
  "type": "summary",
  "time_range": "7d",
  "format": "pdf",
  "sections": [
    "executive_summary",
    "threat_analysis",
    "vulnerability_assessment"
  ],
  "title": "周度安全报告",
  "description": "本周安全状况总结"
}
```

#### 7.3 获取特定报告
```http
GET /api/security/reports/{id}
```

#### 7.4 导出报告
```http
GET /api/security/reports/{id}/export
```

**查询参数:**
- `format` (string): 导出格式 (pdf, html, json, csv)

#### 7.5 获取报告模板
```http
GET /api/security/reports/templates
```

#### 7.6 安排定时报告
```http
POST /api/security/reports/schedule
```

**请求体:**
```json
{
  "name": "月度安全报告",
  "type": "summary",
  "schedule": "0 0 1 * *",
  "format": "pdf",
  "recipients": ["admin@example.com"],
  "enabled": true
}
```

## 数据模型

### ThreatAlert (威胁告警)
```json
{
  "id": "string",
  "title": "string",
  "description": "string",
  "severity": "critical|high|medium|low",
  "status": "active|acknowledged|resolved",
  "source_ip": "string",
  "target": "string",
  "attack_type": "string",
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

### Vulnerability (漏洞)
```json
{
  "id": "string",
  "title": "string",
  "description": "string",
  "severity": "critical|high|medium|low",
  "cvss_score": "number",
  "cve_id": "string",
  "affected_component": "string",
  "fix_available": "boolean",
  "discovered_at": "datetime"
}
```

### ScanJob (扫描任务)
```json
{
  "id": "string",
  "name": "string",
  "type": "web|network|host",
  "target": "string",
  "status": "pending|running|completed|failed",
  "progress": "number",
  "config": "object",
  "created_at": "datetime",
  "started_at": "datetime",
  "completed_at": "datetime"
}
```

### SecurityReport (安全报告)
```json
{
  "id": "string",
  "title": "string",
  "type": "summary|technical|compliance|incident",
  "status": "generating|completed|failed",
  "format": "pdf|html|json|csv",
  "file_size": "string",
  "created_at": "datetime",
  "generated_at": "datetime"
}
```

## 错误码

| 错误码 | 描述 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 未授权访问 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 409 | 资源冲突 |
| 429 | 请求频率限制 |
| 500 | 服务器内部错误 |
| 503 | 服务不可用 |

## 状态码说明

### 扫描任务状态
- `pending`: 等待执行
- `running`: 正在执行
- `completed`: 已完成
- `failed`: 执行失败
- `cancelled`: 已取消

### 告警状态
- `active`: 活跃状态
- `acknowledged`: 已确认
- `resolved`: 已解决
- `false_positive`: 误报

### 严重程度级别
- `critical`: 严重 (9.0-10.0)
- `high`: 高危 (7.0-8.9)
- `medium`: 中危 (4.0-6.9)
- `low`: 低危 (0.1-3.9)

## 使用示例

### 创建漏洞扫描任务
```bash
curl -X POST "https://api.example.com/api/security/vulnerabilities/scan" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "网站安全扫描",
    "type": "web",
    "target": "https://example.com",
    "config": {
      "depth": 3,
      "timeout": 300
    }
  }'
```

### 获取安全指标
```bash
curl -X GET "https://api.example.com/api/security/monitoring/metrics?timeRange=7d" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 生成安全报告
```bash
curl -X POST "https://api.example.com/api/security/reports/generate" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "summary",
    "time_range": "30d",
    "format": "pdf",
    "title": "月度安全报告"
  }'
```

## 注意事项

1. **认证**: 所有API调用都需要有效的Bearer Token
2. **频率限制**: API调用受到频率限制，建议合理控制请求频率
3. **数据保留**: 审计日志和报告数据有保留期限制
4. **权限控制**: 不同用户角色对API的访问权限不同
5. **异步操作**: 扫描和报告生成为异步操作，需要轮询状态

## 更新日志

### v1.0.0 (2024-12)
- 初始版本发布
- 支持威胁检测、漏洞扫描、安全监控
- 提供报告生成和导出功能
- 集成安全审计和合规检查