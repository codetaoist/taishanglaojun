-- 创建文化智慧表
CREATE TABLE IF NOT EXISTS cultural_wisdom (
    id VARCHAR(255) PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    summary TEXT,
    author VARCHAR(255) NOT NULL,
    author_id VARCHAR(255),
    source VARCHAR(255),
    category VARCHAR(100),
    school VARCHAR(100),
    tags TEXT, -- JSON数组存储标签
    difficulty VARCHAR(50) DEFAULT 'medium',
    status VARCHAR(50) DEFAULT 'published',
    view_count INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    share_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    is_featured BOOLEAN DEFAULT FALSE,
    is_recommended BOOLEAN DEFAULT FALSE,
    metadata TEXT, -- JSON存储额外元数据
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_cultural_wisdom_category ON cultural_wisdom(category);
CREATE INDEX IF NOT EXISTS idx_cultural_wisdom_school ON cultural_wisdom(school);
CREATE INDEX IF NOT EXISTS idx_cultural_wisdom_author ON cultural_wisdom(author);
CREATE INDEX IF NOT EXISTS idx_cultural_wisdom_status ON cultural_wisdom(status);
CREATE INDEX IF NOT EXISTS idx_cultural_wisdom_created_at ON cultural_wisdom(created_at);
CREATE INDEX IF NOT EXISTS idx_cultural_wisdom_view_count ON cultural_wisdom(view_count);
CREATE INDEX IF NOT EXISTS idx_cultural_wisdom_like_count ON cultural_wisdom(like_count);
CREATE INDEX IF NOT EXISTS idx_cultural_wisdom_is_featured ON cultural_wisdom(is_featured);
CREATE INDEX IF NOT EXISTS idx_cultural_wisdom_is_recommended ON cultural_wisdom(is_recommended);

-- 全文搜索索引（PostgreSQL）
-- CREATE INDEX IF NOT EXISTS idx_cultural_wisdom_fulltext ON cultural_wisdom USING gin(to_tsvector('chinese', title || ' ' || content));

-- 创建分类表
CREATE TABLE IF NOT EXISTS wisdom_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    parent_id INTEGER REFERENCES wisdom_categories(id),
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建学派表
CREATE TABLE IF NOT EXISTS wisdom_schools (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    founder VARCHAR(255),
    period VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建标签表
CREATE TABLE IF NOT EXISTS wisdom_tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    usage_count INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建智慧内容与标签的关联表
CREATE TABLE IF NOT EXISTS wisdom_tag_relations (
    wisdom_id VARCHAR(255) REFERENCES cultural_wisdom(id) ON DELETE CASCADE,
    tag_id INTEGER REFERENCES wisdom_tags(id) ON DELETE CASCADE,
    PRIMARY KEY (wisdom_id, tag_id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建用户收藏表
CREATE TABLE IF NOT EXISTS wisdom_favorites (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    wisdom_id VARCHAR(255) REFERENCES cultural_wisdom(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, wisdom_id)
);

-- 创建用户点赞表
CREATE TABLE IF NOT EXISTS wisdom_likes (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    wisdom_id VARCHAR(255) REFERENCES cultural_wisdom(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, wisdom_id)
);

-- 创建评论表
CREATE TABLE IF NOT EXISTS wisdom_comments (
    id SERIAL PRIMARY KEY,
    wisdom_id VARCHAR(255) REFERENCES cultural_wisdom(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    parent_id INTEGER REFERENCES wisdom_comments(id),
    content TEXT NOT NULL,
    like_count INTEGER DEFAULT 0,
    is_approved BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- 创建用户笔记表
CREATE TABLE IF NOT EXISTS wisdom_notes (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    wisdom_id VARCHAR(255) REFERENCES cultural_wisdom(id) ON DELETE CASCADE,
    title VARCHAR(255),
    content TEXT NOT NULL,
    is_private BOOLEAN DEFAULT TRUE,
    tags TEXT[], -- 笔记标签数组
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, wisdom_id)
);

-- 插入默认分类数据
INSERT INTO wisdom_categories (name, description, sort_order) VALUES
('道家', '道家思想和哲学', 1),
('儒家', '儒家思想和教育理念', 2),
('佛家', '佛教思想和修行理念', 3),
('法家', '法家思想和治国理念', 4),
('墨家', '墨家思想和兼爱理念', 5),
('兵家', '兵法和军事思想', 6),
('纵横家', '外交和谋略思想', 7),
('阴阳家', '阴阳五行思想', 8),
('名家', '逻辑和辩论思想', 9),
('杂家', '综合各家思想', 10)
ON CONFLICT (name) DO NOTHING;

-- 插入默认学派数据
INSERT INTO wisdom_schools (name, description, founder, period) VALUES
('道教', '以道德经为核心的哲学体系', '老子', '春秋时期'),
('孔门', '以仁义礼智为核心的教育体系', '孔子', '春秋时期'),
('禅宗', '以直指人心为特色的佛教宗派', '达摩', '南北朝'),
('理学', '宋明理学思想体系', '程颢程颐', '宋代'),
('心学', '以心即理为核心的哲学体系', '王阳明', '明代')
ON CONFLICT (name) DO NOTHING;

-- 插入默认标签数据
INSERT INTO wisdom_tags (name, description) VALUES
('修身', '个人修养和品德提升'),
('治国', '国家治理和政治思想'),
('齐家', '家庭和睦和家族管理'),
('平天下', '天下大同和社会理想'),
('哲学', '哲学思辨和理论思考'),
('教育', '教育理念和方法'),
('修行', '精神修炼和心性提升'),
('智慧', '人生智慧和处世之道'),
('道德', '道德品质和伦理规范'),
('自然', '自然规律和天人合一')
ON CONFLICT (name) DO NOTHING;