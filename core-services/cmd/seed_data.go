package main

import (
	"log"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
	"github.com/codetaoist/taishanglaojun/core-services/internal/config"
	"github.com/codetaoist/taishanglaojun/core-services/internal/database"
	"github.com/codetaoist/taishanglaojun/core-services/internal/logger"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logConfig := logger.LogConfig{
		Level:      cfg.Logger.Level,
		Format:     cfg.Logger.Format,
		Output:     cfg.Logger.Output,
		Filename:   cfg.Logger.Filename,
		MaxSize:    cfg.Logger.MaxSize,
		MaxBackups: cfg.Logger.MaxBackups,
		MaxAge:     cfg.Logger.MaxAge,
		Compress:   cfg.Logger.Compress,
	}
	zapLogger, err := logger.New(logConfig)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	// 初始化数据库
	dbConfig := database.Config{
		Driver:          cfg.Database.Type,
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		Database:        cfg.Database.Database,
		Username:        cfg.Database.Username,
		Password:        cfg.Database.Password,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: time.Duration(cfg.Database.MaxLifetime) * time.Second,
		SSLMode:         cfg.Database.SSLMode,
		ConnectTimeout:  30 * time.Second,
	}

	db, err := database.New(dbConfig, zapLogger)
	if err != nil {
		zapLogger.Fatal("Failed to initialize database", zap.Error(err))
	}

	zapLogger.Info("开始插入文化智慧样本数据...")

	// 清理现有的乱码数据
	result := db.GetDB().Where("title LIKE ? OR content LIKE ?", "%?%", "%?%").Delete(&models.CulturalWisdom{})
	if result.Error != nil {
		zapLogger.Error("清理乱码数据失败", zap.Error(result.Error))
	} else {
		zapLogger.Info("清理乱码数据完成", zap.Int64("删除记录数", result.RowsAffected))
	}

	// 插入样本数据
	sampleData := getSampleWisdomData()
	
	for _, wisdom := range sampleData {
		// 检查是否已存在
		var existing models.CulturalWisdom
		result := db.GetDB().Where("id = ?", wisdom.ID).First(&existing)
		
		if result.Error != nil {
			// 不存在，创建新记录
			if err := db.GetDB().Create(&wisdom).Error; err != nil {
				zapLogger.Error("插入数据失败", zap.String("id", wisdom.ID), zap.Error(err))
			} else {
				zapLogger.Info("插入数据成功", zap.String("id", wisdom.ID), zap.String("title", wisdom.Title))
			}
		} else {
			// 已存在，更新记录
			if err := db.GetDB().Model(&existing).Updates(&wisdom).Error; err != nil {
				zapLogger.Error("更新数据失败", zap.String("id", wisdom.ID), zap.Error(err))
			} else {
				zapLogger.Info("更新数据成功", zap.String("id", wisdom.ID), zap.String("title", wisdom.Title))
			}
		}
	}

	zapLogger.Info("文化智慧样本数据插入完成")
}

func getSampleWisdomData() []models.CulturalWisdom {
	now := time.Now()
	
	return []models.CulturalWisdom{
		{
			ID:           "wisdom_dao_001",
			Title:        "道可道，非常道",
			Content:      "道可道，非常道；名可名，非常名。无名天地之始，有名万物之母。故常无欲，以观其妙；常有欲，以观其徼。此两者，同出而异名，同谓之玄。玄之又玄，众妙之门。",
			Summary:      "道德经开篇，阐述了道的不可言喻性和万物的本源。道是超越语言和概念的存在，是天地万物的根本。",
			Author:       "老子",
			AuthorID:     "author_laozi_001",
			Category:     "道家",
			School:       "道家",
			Tags:         models.StringSlice{"道德经", "哲学", "本源", "玄学"},
			Difficulty:   "高级",
			Status:       "published",
			ViewCount:    1250,
			LikeCount:    89,
			ShareCount:   23,
			CommentCount: 15,
			IsFeatured:   true,
			IsRecommended: true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "wisdom_confucius_001", 
			Title:        "学而时习之，不亦说乎",
			Content:      "子曰：学而时习之，不亦说乎？有朋自远方来，不亦乐乎？人不知而不愠，不亦君子乎？",
			Summary:      "论语开篇，强调学习的重要性和君子的品格。学习要持续实践，朋友来访要热情相待，被人误解也不生气。",
			Author:       "孔子",
			AuthorID:     "author_confucius_001",
			Category:     "儒家",
			School:       "儒家",
			Tags:         models.StringSlice{"论语", "学习", "君子", "品格"},
			Difficulty:   "初级",
			Status:       "published",
			ViewCount:    2100,
			LikeCount:    156,
			ShareCount:   45,
			CommentCount: 28,
			IsFeatured:   true,
			IsRecommended: true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "wisdom_buddha_001",
			Title:        "诸行无常，是生灭法",
			Content:      "诸行无常，是生灭法；生灭灭已，寂灭为乐。一切有为法，如梦幻泡影，如露亦如电，应作如是观。",
			Summary:      "佛教核心思想，阐述万物无常的道理。一切现象都在生灭变化中，只有超越生灭才能获得真正的安乐。",
			Author:       "释迦牟尼",
			AuthorID:     "author_buddha_001",
			Category:     "佛家",
			School:       "佛教",
			Tags:         models.StringSlice{"佛经", "无常", "生灭", "寂灭"},
			Difficulty:   "高级",
			Status:       "published",
			ViewCount:    980,
			LikeCount:    67,
			ShareCount:   18,
			CommentCount: 12,
			IsFeatured:   false,
			IsRecommended: true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "wisdom_sunzi_001",
			Title:        "知己知彼，百战不殆",
			Content:      "孙子曰：知己知彼，百战不殆；不知彼而知己，一胜一负；不知彼，不知己，每战必殆。",
			Summary:      "孙子兵法中的经典战略思想。了解自己和敌人的情况，才能在战争中立于不败之地。",
			Author:       "孙武",
			AuthorID:     "author_sunzi_001",
			Category:     "兵家",
			School:       "兵家",
			Tags:         models.StringSlice{"孙子兵法", "战略", "知己知彼", "军事"},
			Difficulty:   "中级",
			Status:       "published",
			ViewCount:    1680,
			LikeCount:    124,
			ShareCount:   35,
			CommentCount: 22,
			IsFeatured:   true,
			IsRecommended: false,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "wisdom_mencius_001",
			Title:        "人之初，性本善",
			Content:      "孟子曰：人之初，性本善。性相近，习相远。苟不教，性乃迁。教之道，贵以专。",
			Summary:      "孟子关于人性的经典论述。人生来本性善良，但会因为环境和教育的不同而产生差异。",
			Author:       "孟子",
			AuthorID:     "author_mencius_001",
			Category:     "儒家",
			School:       "儒家",
			Tags:         models.StringSlice{"人性", "教育", "性善论", "儒学"},
			Difficulty:   "中级",
			Status:       "published",
			ViewCount:    1420,
			LikeCount:    98,
			ShareCount:   28,
			CommentCount: 19,
			IsFeatured:   false,
			IsRecommended: true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "wisdom_zhuangzi_001",
			Title:        "庄周梦蝶",
			Content:      "昔者庄周梦为胡蝶，栩栩然胡蝶也，自喻适志与！不知周也。俄然觉，则蘧蘧然周也。不知周之梦为胡蝶与，胡蝶之梦为周与？周与胡蝶，则必有分矣。此之谓物化。",
			Summary:      "庄子著名的哲学寓言，探讨现实与梦境、主体与客体的关系，体现了道家对存在本质的深刻思考。",
			Author:       "庄子",
			AuthorID:     "author_zhuangzi_001",
			Category:     "道家",
			School:       "道家",
			Tags:         models.StringSlice{"庄子", "梦境", "物化", "哲学"},
			Difficulty:   "高级",
			Status:       "published",
			ViewCount:    1890,
			LikeCount:    142,
			ShareCount:   38,
			CommentCount: 25,
			IsFeatured:   true,
			IsRecommended: true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "wisdom_hanfeizi_001",
			Title:        "法不阿贵，绳不挠曲",
			Content:      "法不阿贵，绳不挠曲。法之所加，智者弗能辞，勇者弗敢争。刑过不避大臣，赏善不遗匹夫。",
			Summary:      "韩非子关于法治的经典论述。法律面前人人平等，不因地位高低而有所偏私，体现了法家思想的核心理念。",
			Author:       "韩非子",
			AuthorID:     "author_hanfeizi_001",
			Category:     "法家",
			School:       "法家",
			Tags:         models.StringSlice{"法治", "平等", "韩非子", "治国"},
			Difficulty:   "中级",
			Status:       "published",
			ViewCount:    1320,
			LikeCount:    87,
			ShareCount:   21,
			CommentCount: 16,
			IsFeatured:   false,
			IsRecommended: true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "wisdom_mozi_001",
			Title:        "兼相爱，交相利",
			Content:      "子墨子言曰：仁人之所以为事者，必兴天下之利，除天下之害。然当今之时，天下之害孰为大？曰：大国攻小国也，大家乱小家也，强凌弱，众暴寡，诈欺愚，贵傲贱，此天下之害也。",
			Summary:      "墨子提出的兼爱思想，主张人与人之间应该相互关爱，互相帮助，消除天下的祸害，实现社会和谐。",
			Author:       "墨子",
			AuthorID:     "author_mozi_001",
			Category:     "墨家",
			School:       "墨家",
			Tags:         models.StringSlice{"兼爱", "墨子", "社会和谐", "利他"},
			Difficulty:   "中级",
			Status:       "published",
			ViewCount:    980,
			LikeCount:    65,
			ShareCount:   15,
			CommentCount: 11,
			IsFeatured:   false,
			IsRecommended: false,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "wisdom_xunzi_001",
			Title:        "人之性恶，其善者伪也",
			Content:      "人之性恶，其善者伪也。今人之性，生而有好利焉，顺是，故争夺生而辞让亡焉；生而有疾恶焉，顺是，故残贼生而忠信亡焉。",
			Summary:      "荀子的性恶论，认为人性本恶，善良的品德是后天教化的结果。与孟子的性善论形成对比，强调教育的重要性。",
			Author:       "荀子",
			AuthorID:     "author_xunzi_001",
			Category:     "儒家",
			School:       "儒家",
			Tags:         models.StringSlice{"性恶论", "荀子", "教化", "人性"},
			Difficulty:   "高级",
			Status:       "published",
			ViewCount:    1150,
			LikeCount:    78,
			ShareCount:   19,
			CommentCount: 13,
			IsFeatured:   false,
			IsRecommended: true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "wisdom_yijing_001",
			Title:        "天行健，君子以自强不息",
			Content:      "天行健，君子以自强不息。地势坤，君子以厚德载物。",
			Summary:      "易经中的经典名句，天道刚健，君子应该像天一样自强不息；地道柔顺，君子应该像大地一样厚德载物。",
			Author:       "佚名",
			AuthorID:     "author_yijing_001",
			Category:     "阴阳家",
			School:       "易学",
			Tags:         models.StringSlice{"易经", "自强不息", "厚德载物", "天地"},
			Difficulty:   "中级",
			Status:       "published",
			ViewCount:    2350,
			LikeCount:    189,
			ShareCount:   52,
			CommentCount: 34,
			IsFeatured:   true,
			IsRecommended: true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
}