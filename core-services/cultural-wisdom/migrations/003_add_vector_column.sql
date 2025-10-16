-- 添加向量字段到文化智慧表
-- 用于存储文本的向量表示，支持语义搜索

-- 添加向量字段（使用TEXT类型存储JSON格式的float32数组）
ALTER TABLE cultural_wisdom ADD COLUMN IF NOT EXISTS vector TEXT;

-- 添加向量索引（如果使用PostgreSQL的pgvector扩展）
-- CREATE INDEX IF NOT EXISTS idx_cultural_wisdom_vector ON cultural_wisdom USING ivfflat (vector vector_cosine_ops);

-- 添加向量维度字段（可选，用于记录向量维度）
ALTER TABLE cultural_wisdom ADD COLUMN IF NOT EXISTS vector_dimension INTEGER DEFAULT 1536;

-- 添加向量生成时间戳（用于缓存管理）
ALTER TABLE cultural_wisdom ADD COLUMN IF NOT EXISTS vector_generated_at TIMESTAMP;

-- 创建向量相关的索引
CREATE INDEX IF NOT EXISTS idx_cultural_wisdom_vector_generated_at ON cultural_wisdom(vector_generated_at);
CREATE INDEX IF NOT EXISTS idx_cultural_wisdom_vector_dimension ON cultural_wisdom(vector_dimension);

-- 添加向量质量评分字段（用于搜索排序）
ALTER TABLE cultural_wisdom ADD COLUMN IF NOT EXISTS content_score FLOAT DEFAULT 0.0;
CREATE INDEX IF NOT EXISTS idx_cultural_wisdom_content_score ON cultural_wisdom(content_score);

-- 更新现有记录的内容评分（基于内容长度、标题质量等）
UPDATE cultural_wisdom 
SET content_score = CASE 
    WHEN LENGTH(content) > 500 AND LENGTH(title) > 10 THEN 0.9
    WHEN LENGTH(content) > 200 AND LENGTH(title) > 5 THEN 0.7
    WHEN LENGTH(content) > 100 THEN 0.5
    ELSE 0.3
END
WHERE content_score = 0.0;