-- 插入样本文化智慧数据
-- 确保数据库连接使用UTF-8编码

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

-- 清理现有的乱码数据
DELETE FROM cultural_wisdom WHERE title LIKE '%?%' OR content LIKE '%?%';

-- 插入正确的文化智慧样本数据
INSERT INTO cultural_wisdom (
    id, title, content, summary, author, author_id, category, school, 
    tags, difficulty, status, view_count, like_count, share_count, 
    comment_count, is_featured, is_recommended, created_at, updated_at
) VALUES 
(
    'wisdom_dao_001',
    '道可道，非常道',
    '道可道，非常道；名可名，非常名。无名天地之始，有名万物之母。故常无欲，以观其妙；常有欲，以观其徼。此两者，同出而异名，同谓之玄。玄之又玄，众妙之门。',
    '道德经开篇，阐述了道的不可言喻性和万物的本源。道是超越语言和概念的存在，是天地万物的根本。',
    '老子',
    'author_laozi_001',
    '道家',
    '道教',
    '["哲学", "道", "本源", "玄妙"]',
    '5',
    'published',
    156,
    23,
    8,
    5,
    true,
    true,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
),
(
    'wisdom_confucius_001',
    '仁者爱人',
    '仁者爱人，有礼者敬人。爱人者，人恒爱之；敬人者，人恒敬之。夫仁者，己欲立而立人，己欲达而达人。能近取譬，可谓仁之方也已。',
    '孔子论仁，强调仁者以爱人为本，礼者以敬人为要。仁的实践在于推己及人，将心比心。',
    '孔子',
    'author_confucius_001',
    '儒家',
    '孔门',
    '["仁爱", "礼仪", "修身", "教育"]',
    '3',
    'published',
    234,
    45,
    12,
    8,
    true,
    true,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
),
(
    'wisdom_buddha_001',
    '诸行无常',
    '诸行无常，是生灭法。生灭灭已，寂灭为乐。一切有为法，如梦幻泡影，如露亦如电，应作如是观。',
    '佛陀教导世间万法皆无常，生灭变化是自然规律。通过观照无常，可以达到内心的寂静安乐。',
    '释迦牟尼',
    'author_buddha_001',
    '佛家',
    '禅宗',
    '["无常", "生灭", "寂灭", "观照"]',
    '4',
    'published',
    189,
    31,
    6,
    3,
    false,
    true,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
),
(
    'wisdom_mencius_001',
    '民为贵，社稷次之，君为轻',
    '民为贵，社稷次之，君为轻。是故得乎丘民而为天子，得乎天子为诸侯，得乎诸侯为大夫。诸侯危社稷，则变置。牺牲既成，粢盛既洁，祭祀以时，然而旱干水溢，则变置社稷。',
    '孟子提出民本思想，认为人民是国家的根本，君主应当以民为重，这是古代民主思想的重要体现。',
    '孟子',
    'author_mencius_001',
    '儒家',
    '孔门',
    '["民本", "治国", "政治", "仁政"]',
    '4',
    'published',
    167,
    28,
    9,
    4,
    true,
    false,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
),
(
    'wisdom_zhuangzi_001',
    '庄周梦蝶',
    '昔者庄周梦为胡蝶，栩栩然胡蝶也，自喻适志与，不知周也。俄然觉，则蘧蘧然周也。不知周之梦为胡蝶与，胡蝶之梦为周与？周与胡蝶，则必有分矣。此之谓物化。',
    '庄子通过梦蝶的故事，探讨了现实与梦境、主体与客体的界限问题，体现了道家对存在本质的深刻思考。',
    '庄子',
    'author_zhuangzi_001',
    '道家',
    '道教',
    '["梦境", "现实", "物化", "哲学"]',
    '5',
    'published',
    298,
    52,
    15,
    11,
    true,
    true,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);