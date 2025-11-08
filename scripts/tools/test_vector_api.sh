#!/bin/bash

# 向量数据库API测试脚本

API_BASE="http://localhost:8082/api/v1/vector"

echo "=== 向量数据库API测试 ==="

# 1. 测试连接向量数据库
echo -e "\n1. 测试连接向量数据库..."
curl -X POST "${API_BASE}/vector-db/connect" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "milvus",
    "host": "localhost",
    "port": 19530,
    "user": "",
    "password": "",
    "database": ""
  }' | jq .

# 2. 获取向量数据库信息
echo -e "\n2. 获取向量数据库信息..."
curl -X GET "${API_BASE}/vector-db/info" | jq .

# 3. 列出所有集合
echo -e "\n3. 列出所有集合..."
curl -X GET "${API_BASE}/collections" | jq .

# 4. 创建集合
echo -e "\n4. 创建集合..."
curl -X POST "${API_BASE}/collections" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-collection",
    "dimension": 1536,
    "metric_type": "cosine",
    "index_type": "HNSW"
  }' | jq .

# 5. 获取集合信息
echo -e "\n5. 获取集合信息..."
curl -X GET "${API_BASE}/collections/test-collection" | jq .

# 6. 获取集合统计信息
echo -e "\n6. 获取集合统计信息..."
curl -X GET "${API_BASE}/collections/test-collection/stats" | jq .

# 7. 创建索引
echo -e "\n7. 创建索引..."
curl -X POST "${API_BASE}/collections/test-collection/indexes" \
  -H "Content-Type: application/json" \
  -d '{
    "field_name": "vector",
    "index_type": "HNSW",
    "metric_type": "cosine",
    "params": {
      "M": 16,
      "efConstruction": 64
    }
  }' | jq .

# 8. 插入向量
echo -e "\n8. 插入向量..."
curl -X POST "${API_BASE}/collections/test-collection/vectors" \
  -H "Content-Type: application/json" \
  -d '{
    "vectors": [
      {
        "id": "test-vector-1",
        "values": [0.1, 0.2, 0.3, 0.4, 0.5],
        "metadata": {
          "title": "测试文档1",
          "content": "这是第一个测试文档的内容"
        }
      },
      {
        "id": "test-vector-2",
        "values": [0.5, 0.4, 0.3, 0.2, 0.1],
        "metadata": {
          "title": "测试文档2",
          "content": "这是第二个测试文档的内容"
        }
      }
    ]
  }' | jq .

# 9. 搜索向量
echo -e "\n9. 搜索向量..."
curl -X POST "${API_BASE}/collections/test-collection/search" \
  -H "Content-Type: application/json" \
  -d '{
    "vector": [0.1, 0.2, 0.3, 0.4, 0.5],
    "top_k": 5,
    "include_values": true,
    "include_metadata": true,
    "filter": {
      "title": "测试文档1"
    }
  }' | jq .

# 10. 获取单个向量
echo -e "\n10. 获取单个向量..."
curl -X GET "${API_BASE}/collections/test-collection/vectors/test-vector-1?include_vector=true" | jq .

# 11. 删除向量
echo -e "\n11. 删除向量..."
curl -X DELETE "${API_BASE}/collections/test-collection/vectors/test-vector-1" | jq .

# 12. 删除索引
echo -e "\n12. 删除索引..."
curl -X DELETE "${API_BASE}/collections/test-collection/indexes/vector" | jq .

# 13. 删除集合
echo -e "\n13. 删除集合..."
curl -X DELETE "${API_BASE}/collections/test-collection" | jq .

echo -e "\n=== 测试完成 ==="