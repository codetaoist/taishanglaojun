# 文件结构整理方案

## 当前问题
1. 根目录下有多个Go编译生成的二进制可执行文件散落
2. 缺少合理的目录结构来分类不同类型的文件
3. 配置文件、文档和可执行文件混杂在一起

## 建议的目录结构

```
codetaoist/
├── .env.example              # 环境变量示例文件
├── .gitignore                # Git忽略文件
├── .gitlab-ci.yml            # GitLab CI配置
├── README.md                 # 项目说明文档
├── go.mod                    # Go模块定义
├── go.sum                    # Go模块校验和
├── bin/                      # 编译后的二进制文件
│   ├── create_table
│   ├── debug_config
│   ├── gen_jwt
│   ├── init_db
│   ├── main
│   ├── start_all
│   ├── start_all_simple
│   ├── temp_hash
│   ├── test_admin_password
│   └── test_ollama
├── cmd/                      # 命令行工具源码
│   ├── api/
│   ├── service_tools/
│   └── ...
├── config/                   # 配置文件
├── db/                       # 数据库相关文件
├── docs/                     # 文档
├── openapi/                  # API规范
├── scripts/                  # 脚本文件
├── services/                 # 微服务源码
│   ├── api/
│   ├── auth/
│   └── gateway/
├── tests/                    # 测试文件
└── volumes/                  # 数据卷
```

## 整理步骤

1. 创建 `bin/` 目录
2. 将所有二进制可执行文件移动到 `bin/` 目录
3. 更新 `.gitignore` 文件，忽略 `bin/` 目录
4. 更新任何引用这些二进制文件的脚本或文档

## 注意事项

- 二进制文件不应提交到版本控制系统，应在构建时生成
- 可以考虑在 `scripts/` 目录中添加构建脚本，自动编译并放置到 `bin/` 目录
- 如果某些工具是临时性的，可以考虑删除或移动到 `tools/` 目录