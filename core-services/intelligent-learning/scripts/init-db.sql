-- 智能学习系统数据库初始化脚本
-- 创建数据库和基础表结构

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- 学习者表
CREATE TABLE IF NOT EXISTS learners (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    full_name VARCHAR(200) NOT NULL,
    avatar_url VARCHAR(500),
    bio TEXT,
    learning_preferences JSONB DEFAULT '{}',
    skill_level VARCHAR(50) DEFAULT 'beginner',
    timezone VARCHAR(50) DEFAULT 'UTC',
    language VARCHAR(10) DEFAULT 'en',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 学习目标表
CREATE TABLE IF NOT EXISTS learning_goals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    learner_id UUID NOT NULL REFERENCES learners(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    target_date DATE,
    priority VARCHAR(20) DEFAULT 'medium',
    status VARCHAR(20) DEFAULT 'active',
    progress INTEGER DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 技能表
CREATE TABLE IF NOT EXISTS skills (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    learner_id UUID NOT NULL REFERENCES learners(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50),
    level INTEGER DEFAULT 1 CHECK (level >= 1 AND level <= 10),
    experience_points INTEGER DEFAULT 0,
    last_practiced_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(learner_id, name)
);

-- 学习活动表
CREATE TABLE IF NOT EXISTS learning_activities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    learner_id UUID NOT NULL REFERENCES learners(id) ON DELETE CASCADE,
    content_id UUID,
    activity_type VARCHAR(50) NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    duration_minutes INTEGER,
    completion_rate INTEGER DEFAULT 0 CHECK (completion_rate >= 0 AND completion_rate <= 100),
    score INTEGER,
    metadata JSONB DEFAULT '{}',
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 学习内容表
CREATE TABLE IF NOT EXISTS learning_contents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(200) NOT NULL,
    description TEXT,
    content_type VARCHAR(50) NOT NULL,
    content_url VARCHAR(500),
    content_data JSONB,
    difficulty_level VARCHAR(20) DEFAULT 'beginner',
    estimated_duration INTEGER, -- 分钟
    tags TEXT[],
    knowledge_nodes UUID[],
    prerequisites UUID[],
    learning_objectives TEXT[],
    author_id UUID,
    status VARCHAR(20) DEFAULT 'draft',
    version INTEGER DEFAULT 1,
    view_count INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    rating DECIMAL(3,2) DEFAULT 0.0,
    metadata JSONB DEFAULT '{}',
    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 内容交互表
CREATE TABLE IF NOT EXISTS content_interactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    learner_id UUID NOT NULL REFERENCES learners(id) ON DELETE CASCADE,
    content_id UUID NOT NULL REFERENCES learning_contents(id) ON DELETE CASCADE,
    interaction_type VARCHAR(50) NOT NULL,
    interaction_data JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 知识节点表
CREATE TABLE IF NOT EXISTS knowledge_nodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    node_type VARCHAR(50) NOT NULL,
    properties JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 知识关系表
CREATE TABLE IF NOT EXISTS knowledge_relations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_node_id UUID NOT NULL REFERENCES knowledge_nodes(id) ON DELETE CASCADE,
    target_node_id UUID NOT NULL REFERENCES knowledge_nodes(id) ON DELETE CASCADE,
    relation_type VARCHAR(50) NOT NULL,
    properties JSONB DEFAULT '{}',
    weight DECIMAL(5,4) DEFAULT 1.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(source_node_id, target_node_id, relation_type)
);

-- 学习路径表
CREATE TABLE IF NOT EXISTS learning_paths (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    learner_id UUID NOT NULL REFERENCES learners(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    path_data JSONB NOT NULL,
    difficulty_level VARCHAR(20) DEFAULT 'beginner',
    estimated_duration INTEGER, -- 分钟
    status VARCHAR(20) DEFAULT 'active',
    progress INTEGER DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 推荐记录表
CREATE TABLE IF NOT EXISTS recommendations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    learner_id UUID NOT NULL REFERENCES learners(id) ON DELETE CASCADE,
    content_id UUID REFERENCES learning_contents(id) ON DELETE CASCADE,
    recommendation_type VARCHAR(50) NOT NULL,
    score DECIMAL(5,4) NOT NULL,
    reason TEXT,
    metadata JSONB DEFAULT '{}',
    is_clicked BOOLEAN DEFAULT false,
    is_completed BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_learners_username ON learners(username);
CREATE INDEX IF NOT EXISTS idx_learners_email ON learners(email);
CREATE INDEX IF NOT EXISTS idx_learners_active ON learners(is_active);

CREATE INDEX IF NOT EXISTS idx_learning_goals_learner_id ON learning_goals(learner_id);
CREATE INDEX IF NOT EXISTS idx_learning_goals_status ON learning_goals(status);
CREATE INDEX IF NOT EXISTS idx_learning_goals_target_date ON learning_goals(target_date);

CREATE INDEX IF NOT EXISTS idx_skills_learner_id ON skills(learner_id);
CREATE INDEX IF NOT EXISTS idx_skills_category ON skills(category);
CREATE INDEX IF NOT EXISTS idx_skills_level ON skills(level);

CREATE INDEX IF NOT EXISTS idx_learning_activities_learner_id ON learning_activities(learner_id);
CREATE INDEX IF NOT EXISTS idx_learning_activities_content_id ON learning_activities(content_id);
CREATE INDEX IF NOT EXISTS idx_learning_activities_type ON learning_activities(activity_type);
CREATE INDEX IF NOT EXISTS idx_learning_activities_created_at ON learning_activities(created_at);

CREATE INDEX IF NOT EXISTS idx_learning_contents_status ON learning_contents(status);
CREATE INDEX IF NOT EXISTS idx_learning_contents_type ON learning_contents(content_type);
CREATE INDEX IF NOT EXISTS idx_learning_contents_difficulty ON learning_contents(difficulty_level);
CREATE INDEX IF NOT EXISTS idx_learning_contents_tags ON learning_contents USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_learning_contents_knowledge_nodes ON learning_contents USING GIN(knowledge_nodes);
CREATE INDEX IF NOT EXISTS idx_learning_contents_published_at ON learning_contents(published_at);

CREATE INDEX IF NOT EXISTS idx_content_interactions_learner_id ON content_interactions(learner_id);
CREATE INDEX IF NOT EXISTS idx_content_interactions_content_id ON content_interactions(content_id);
CREATE INDEX IF NOT EXISTS idx_content_interactions_type ON content_interactions(interaction_type);
CREATE INDEX IF NOT EXISTS idx_content_interactions_created_at ON content_interactions(created_at);

CREATE INDEX IF NOT EXISTS idx_knowledge_nodes_type ON knowledge_nodes(node_type);
CREATE INDEX IF NOT EXISTS idx_knowledge_nodes_name ON knowledge_nodes(name);

CREATE INDEX IF NOT EXISTS idx_knowledge_relations_source ON knowledge_relations(source_node_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_relations_target ON knowledge_relations(target_node_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_relations_type ON knowledge_relations(relation_type);

CREATE INDEX IF NOT EXISTS idx_learning_paths_learner_id ON learning_paths(learner_id);
CREATE INDEX IF NOT EXISTS idx_learning_paths_status ON learning_paths(status);
CREATE INDEX IF NOT EXISTS idx_learning_paths_difficulty ON learning_paths(difficulty_level);

CREATE INDEX IF NOT EXISTS idx_recommendations_learner_id ON recommendations(learner_id);
CREATE INDEX IF NOT EXISTS idx_recommendations_content_id ON recommendations(content_id);
CREATE INDEX IF NOT EXISTS idx_recommendations_type ON recommendations(recommendation_type);
CREATE INDEX IF NOT EXISTS idx_recommendations_score ON recommendations(score);
CREATE INDEX IF NOT EXISTS idx_recommendations_created_at ON recommendations(created_at);

-- 全文搜索索引
CREATE INDEX IF NOT EXISTS idx_learning_contents_title_search ON learning_contents USING GIN(to_tsvector('english', title));
CREATE INDEX IF NOT EXISTS idx_learning_contents_description_search ON learning_contents USING GIN(to_tsvector('english', description));
CREATE INDEX IF NOT EXISTS idx_knowledge_nodes_name_search ON knowledge_nodes USING GIN(to_tsvector('english', name));
CREATE INDEX IF NOT EXISTS idx_knowledge_nodes_description_search ON knowledge_nodes USING GIN(to_tsvector('english', description));

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为需要的表创建更新时间触发器
CREATE TRIGGER update_learners_updated_at BEFORE UPDATE ON learners FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_learning_goals_updated_at BEFORE UPDATE ON learning_goals FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_skills_updated_at BEFORE UPDATE ON skills FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_learning_contents_updated_at BEFORE UPDATE ON learning_contents FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_knowledge_nodes_updated_at BEFORE UPDATE ON knowledge_nodes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_knowledge_relations_updated_at BEFORE UPDATE ON knowledge_relations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_learning_paths_updated_at BEFORE UPDATE ON learning_paths FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 插入示例数据
INSERT INTO learners (username, email, full_name, bio, skill_level) VALUES
('john_doe', 'john@example.com', 'John Doe', 'Passionate learner interested in technology and science', 'intermediate'),
('jane_smith', 'jane@example.com', 'Jane Smith', 'Lifelong learner with a focus on personal development', 'beginner'),
('alex_chen', 'alex@example.com', 'Alex Chen', 'Advanced learner specializing in AI and machine learning', 'advanced')
ON CONFLICT (username) DO NOTHING;

-- 插入示例知识节点
INSERT INTO knowledge_nodes (name, description, node_type) VALUES
('Programming Fundamentals', 'Basic concepts of programming including variables, loops, and functions', 'concept'),
('Data Structures', 'Understanding of arrays, lists, trees, and other data organization methods', 'concept'),
('Algorithms', 'Problem-solving techniques and algorithmic thinking', 'concept'),
('Object-Oriented Programming', 'Principles of OOP including encapsulation, inheritance, and polymorphism', 'concept'),
('Database Design', 'Fundamentals of relational database design and normalization', 'concept')
ON CONFLICT DO NOTHING;

-- 插入示例学习内容
INSERT INTO learning_contents (title, description, content_type, difficulty_level, estimated_duration, tags, status) VALUES
('Introduction to Go Programming', 'Learn the basics of Go programming language', 'video', 'beginner', 120, ARRAY['programming', 'go', 'basics'], 'published'),
('Advanced Data Structures in Go', 'Deep dive into complex data structures implementation', 'article', 'advanced', 180, ARRAY['programming', 'go', 'data-structures'], 'published'),
('Building REST APIs with Gin', 'Hands-on tutorial for creating REST APIs using Gin framework', 'tutorial', 'intermediate', 240, ARRAY['programming', 'go', 'web', 'api'], 'published')
ON CONFLICT DO NOTHING;