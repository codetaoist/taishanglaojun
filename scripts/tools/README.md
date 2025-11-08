# 工具脚本

此目录包含各种开发和测试工具脚本。

## 密码哈希工具

### temp_hash.go
用于生成密码哈希的Go工具，主要用于开发环境创建测试用户密码。

使用方法：
```bash
cd scripts/tools
go run temp_hash.go
```

## API测试工具

### test_vector_api.sh
向量数据库API测试脚本，用于测试Milvus向量数据库的各种操作。

使用方法：
```bash
cd scripts/tools
chmod +x test_vector_api.sh
./test_vector_api.sh
```

测试内容：
- 连接测试
- 集合管理（创建/查询/统计/删除）
- 索引操作（创建/删除）
- 向量操作（插入/搜索/获取/删除）