# 测试文档索引

- 测试策略：`strategy.md`
- 原生多端测试方案：`plan-native.md`
- 用例集：`test-cases.md`

## 测试目录结构

项目采用分层测试结构，不同类型的测试分布在不同的目录中：

```
codetaoist/
├── tests/                      # 系统级测试
│   ├── integration/            # 服务间集成测试
│   ├── e2e/                    # 端到端测试
│   └── contracts/              # 契约测试
└── services/
    ├── api/
    │   ├── tests/              # API服务测试
    │   │   ├── integration/    # API集成测试
    │   │   └── unit/           # API单元测试
    │   └── internal/           # 与源码同目录的单元测试
    ├── auth/
    │   └── tests/              # 认证服务测试
    │       ├── integration/    # 认证集成测试
    │       └── unit/           # 认证单元测试
    └── gateway/
        └── tests/              # 网关服务测试
            ├── integration/    # 网关集成测试
            └── unit/           # 网关单元测试
```

## 测试执行指南

### 运行所有测试
```bash
# 运行系统级测试
go test ./tests/...

# 运行所有服务测试
go test ./services/...
```

### 运行特定服务测试
```bash
# API服务测试
cd services/api
go test ./tests/...

# 认证服务测试
cd services/auth
go test ./tests/...

# 网关服务测试
cd services/gateway
go test ./tests/...
```

### 运行特定类型测试
```bash
# 只运行集成测试
go test -tags=integration ./tests/...

# 只运行单元测试
go test -tags=unit ./services/...
```

## 测试覆盖率

生成测试覆盖率报告：
```bash
# 生成覆盖率报告
go test -cover ./...

# 生成HTML格式的覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```