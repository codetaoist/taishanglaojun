# 测试说明

本目录包含项目的各种测试脚本和测试用例。

## 测试类型

### 1. 集成测试

#### 认证服务集成测试
- 文件：`test_auth_integration.sh`
- 描述：测试认证服务和API服务的集成功能
- 运行方式：
  ```bash
  ./tests/test_auth_integration.sh
  ```
- 前提条件：
  - 认证服务运行在 http://localhost:8081
  - API服务运行在 http://localhost:8082
  - 数据库已初始化并包含默认管理员用户

### 2. 端到端测试

- 文件：`e2e/app.spec.ts`
- 描述：使用Playwright进行端到端测试
- 运行方式：
  ```bash
  npx playwright test
  ```

### 3. 合约测试

- 文件：`contracts/test_contracts.py`
- 描述：验证API合约的一致性
- 运行方式：
  ```bash
  python tests/contracts/test_contracts.py
  ```

### 4. API集成测试

- 文件：`integration/api_test.go`
- 描述：Go语言编写的API集成测试
- 运行方式：
  ```bash
  cd tests/integration
  go test -v
  ```

## 使用Docker Compose运行完整测试

1. 启动所有服务：
   ```bash
   docker-compose up -d
   ```

2. 等待服务启动完成（约30秒）

3. 运行认证服务集成测试：
   ```bash
   ./tests/test_auth_integration.sh
   ```

4. 停止服务：
   ```bash
   docker-compose down
   ```

## 测试报告

测试结果将输出到控制台，包含每个测试步骤的响应和状态。失败的测试会显示错误信息和响应代码。

## 添加新测试

1. 将测试脚本放在适当的子目录中
2. 更新本README文件，说明新测试的用途和运行方式
3. 确保测试脚本具有执行权限