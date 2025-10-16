-- 插入示例文化智慧数据
INSERT INTO cultural_wisdom (
    id, title, content, summary, author, author_id, source, category, school, 
    tags, difficulty, status, view_count, like_count, is_featured, is_recommended
) VALUES 
(
    'wisdom_001',
    '道可道，非常道',
    '道可道，非常道；名可名，非常名。无名天地之始，有名万物之母。故常无欲，以观其妙；常有欲，以观其徼。此两者，同出而异名，同谓之玄。玄之又玄，众妙之门。',
    '《道德经》开篇，阐述了道的不可言喻性和万物的根本原理。',
    '老子',
    'author_laozi',
    '道德经',
    '道家',
    '道教',
    '["哲学", "道德", "自然"]',
    'medium',
    'published',
    156,
    23,
    true,
    true
),
(
    'wisdom_002', 
    '学而时习之，不亦说乎',
    '子曰："学而时习之，不亦说乎？有朋自远方来，不亦乐乎？人不知而不愠，不亦君子乎？"',
    '孔子论述学习的快乐和君子的品格。',
    '孔子',
    'author_confucius',
    '论语·学而',
    '儒家',
    '孔门',
    '["教育", "修身", "智慧"]',
    'easy',
    'published',
    89,
    15,
    true,
    false
),
(
    'wisdom_003',
    '知己知彼，百战不殆',
    '孙子曰：知己知彼，百战不殆；不知彼而知己，一胜一负；不知彼不知己，每战必殆。',
    '《孙子兵法》中关于战略思维的经典论述。',
    '孙武',
    'author_sunwu',
    '孙子兵法·谋攻',
    '兵家',
    '兵家',
    '["智慧", "治国"]',
    'medium',
    'published',
    234,
    45,
    false,
    true
),
(
    'wisdom_004',
    '仁者爱人',
    '樊迟问仁。子曰："爱人。"问知。子曰："知人。"',
    '孔子对仁和智的简洁而深刻的定义。',
    '孔子',
    'author_confucius',
    '论语·颜渊',
    '儒家',
    '孔门',
    '["修身", "道德"]',
    'easy',
    'published',
    67,
    12,
    false,
    false
),
(
    'wisdom_005',
    '上善若水',
    '上善若水。水善利万物而不争，处众人之所恶，故几于道。居善地，心善渊，与善仁，言善信，政善治，事善能，动善时。夫唯不争，故无尤。',
    '老子以水为喻，阐述最高境界的善德。',
    '老子',
    'author_laozi',
    '道德经',
    '道家',
    '道教',
    '["道德", "自然", "智慧"]',
    'medium',
    'published',
    178,
    31,
    true,
    true
);

-- 插入标签关联
INSERT INTO wisdom_tag_relations (wisdom_id, tag_id) VALUES
('wisdom_001', 5), -- 哲学
('wisdom_001', 9), -- 道德
('wisdom_001', 10), -- 自然
('wisdom_002', 6), -- 教育
('wisdom_002', 1), -- 修身
('wisdom_002', 8), -- 智慧
('wisdom_003', 8), -- 智慧
('wisdom_003', 2), -- 治国
('wisdom_004', 1), -- 修身
('wisdom_004', 9), -- 道德
('wisdom_005', 9), -- 道德
('wisdom_005', 10), -- 自然
('wisdom_005', 8); -- 智慧