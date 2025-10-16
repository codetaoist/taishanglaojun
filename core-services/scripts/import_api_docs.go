package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/codetaoist/taishanglaojun/core-services/internal/models"
)

// APIDocImporter API文档导入器
type APIDocImporter struct {
	db     *gorm.DB
	logger *log.Logger
}

// NewAPIDocImporter 创建API文档导入器
func NewAPIDocImporter(db *gorm.DB) *APIDocImporter {
	return &APIDocImporter{
		db:     db,
		logger: log.New(os.Stdout, "[API_IMPORT] ", log.LstdFlags),
	}
}

// ImportFromMarkdownFiles 从Markdown文件导入API文档
func (importer *APIDocImporter) ImportFromMarkdownFiles(docsDir string) error {
	importer.logger.Printf("开始从目录导入API文档: %s", docsDir)

	// 扫描分类目录
	categoriesDir := filepath.Join(docsDir, "分类")
	if _, err := os.Stat(categoriesDir); os.IsNotExist(err) {
		return fmt.Errorf("分类目录不存在: %s", categoriesDir)
	}

	// 读取所有分类文件
	files, err := os.ReadDir(categoriesDir)
	if err != nil {
		return fmt.Errorf("读取分类目录失败: %v", err)
	}

	// 创建默认用户ID（用于记录创建者）
	defaultUserID := uuid.New().String()

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			filePath := filepath.Join(categoriesDir, file.Name())
			categoryName := strings.TrimSuffix(file.Name(), ".md")
			
			importer.logger.Printf("处理分类文件: %s", file.Name())
			
			if err := importer.importCategoryFile(filePath, categoryName, defaultUserID); err != nil {
				importer.logger.Printf("导入分类文件失败 %s: %v", file.Name(), err)
				continue
			}
		}
	}

	importer.logger.Println("API文档导入完成")
	return nil
}

// importCategoryFile 导入单个分类文件
func (importer *APIDocImporter) importCategoryFile(filePath, categoryName, userID string) error {
	// 创建或获取分类
	category, err := importer.createOrGetCategory(categoryName, userID)
	if err != nil {
		return fmt.Errorf("创建分类失败: %v", err)
	}

	// 读取文件内容
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 解析Markdown表格
	apis, err := importer.parseMarkdownTable(file, filePath)
	if err != nil {
		return fmt.Errorf("解析Markdown表格失败: %v", err)
	}

	// 保存API接口到数据库
	for _, api := range apis {
		api.CategoryID = category.ID
		api.CreatedBy = userID
		api.CreatedAt = time.Now()
		api.UpdatedAt = time.Now()

		if err := importer.db.Create(&api).Error; err != nil {
			importer.logger.Printf("保存API接口失败 %s %s: %v", api.Method, api.Path, err)
			continue
		}
		
		importer.logger.Printf("成功导入API: %s %s - %s", api.Method, api.Path, api.Name)
	}

	// 记录文档来源
	if err := importer.recordDocumentationSource(filePath, len(apis), userID); err != nil {
		importer.logger.Printf("记录文档来源失败: %v", err)
	}

	return nil
}

// createOrGetCategory 创建或获取分类
func (importer *APIDocImporter) createOrGetCategory(categoryName, userID string) (*models.APICategory, error) {
	var category models.APICategory
	
	// 先尝试查找现有分类
	err := importer.db.Where("name = ?", categoryName).First(&category).Error
	if err == nil {
		return &category, nil
	}

	// 如果不存在，创建新分类
	if err == gorm.ErrRecordNotFound {
		category = models.APICategory{
			ID:          uuid.New().String(),
			Name:        categoryName,
			Code:        strings.ToLower(strings.ReplaceAll(categoryName, " ", "_")),
			Description: fmt.Sprintf("%s相关的API接口", categoryName),
			Icon:        "api",
			Color:       "#1890ff",
			SortOrder:   0,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			CreatedBy:   userID,
		}

		if err := importer.db.Create(&category).Error; err != nil {
			return nil, fmt.Errorf("创建分类失败: %v", err)
		}

		importer.logger.Printf("创建新分类: %s", categoryName)
		return &category, nil
	}

	return nil, fmt.Errorf("查询分类失败: %v", err)
}

// parseMarkdownTable 解析Markdown表格
func (importer *APIDocImporter) parseMarkdownTable(file *os.File, filePath string) ([]models.APIEndpoint, error) {
	var apis []models.APIEndpoint
	scanner := bufio.NewScanner(file)
	
	// 正则表达式匹配表格行
	tableRowRegex := regexp.MustCompile(`^\|\s*([^|]+)\s*\|\s*([^|]+)\s*\|\s*([^|]+)\s*\|\s*([^|]+)\s*\|`)
	linkRegex := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	
	inTable := false
	lineNumber := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNumber++

		// 检测表格开始
		if strings.Contains(line, "| 方法 | 路径 | 名称 | 来源 |") {
			inTable = true
			continue
		}

		// 跳过表格分隔行
		if strings.Contains(line, "|------|") {
			continue
		}

		// 如果在表格中且是有效的表格行
		if inTable && strings.HasPrefix(line, "|") && strings.HasSuffix(line, "|") {
			matches := tableRowRegex.FindStringSubmatch(line)
			if len(matches) == 5 {
				method := strings.TrimSpace(matches[1])
				path := strings.TrimSpace(matches[2])
				name := strings.TrimSpace(matches[3])
				source := strings.TrimSpace(matches[4])

				// 跳过表头行
				if method == "方法" || method == "Method" {
					continue
				}

				// 解析来源链接
				sourceFile := source
				sourceURL := ""
				if linkMatches := linkRegex.FindStringSubmatch(source); len(linkMatches) == 3 {
					sourceFile = linkMatches[1]
					sourceURL = linkMatches[2]
				}

				api := models.APIEndpoint{
					ID:          uuid.New().String(),
					Method:      strings.ToUpper(method),
					Path:        path,
					Name:        name,
					Description: fmt.Sprintf("%s接口", name),
					Summary:     name,
					SourceFile:  sourceFile,
					SourcePath:  sourceURL,
					SourceLine:  lineNumber,
					DocumentURL: sourceURL,
					Status:      "active",
					Version:     "v1",
					IsPublic:    true,
					IsDeprecated: false,
					ViewCount:   0,
					TestCount:   0,
					ErrorCount:  0,
				}

				apis = append(apis, api)
			}
		}

		// 如果遇到空行或其他内容，表格结束
		if inTable && (line == "" || (!strings.HasPrefix(line, "|") && !strings.Contains(line, "|"))) {
			inTable = false
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	return apis, nil
}

// recordDocumentationSource 记录文档来源
func (importer *APIDocImporter) recordDocumentationSource(filePath string, apiCount int, userID string) error {
	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	source := models.APIDocumentationSource{
		ID:          uuid.New().String(),
		Name:        filepath.Base(filePath),
		FilePath:    filePath,
		FileType:    "markdown",
		FileSize:    fileInfo.Size(),
		FileHash:    "", // 可以添加文件哈希计算
		LastScanned: time.Now(),
		ScanStatus:  "completed",
		APICount:    apiCount,
		ErrorMsg:    "",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   userID,
	}

	return importer.db.Create(&source).Error
}

func main() {
	// 数据库连接配置
	dsn := "laojun:xKyyLNMM64zdfNwE@tcp(1.13.249.131:3306)/laojun?charset=utf8mb4&parseTime=True&loc=Local"
	
	fmt.Println("正在连接MySQL数据库...")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	
	fmt.Println("MySQL数据库连接成功!")

	// 自动迁移API文档相关表
	fmt.Println("开始自动迁移API文档相关表...")
	err = db.AutoMigrate(
		&models.APICategory{},
		&models.APIEndpoint{},
		&models.APIDocumentationSource{},
		&models.APITestRecord{},
		&models.APIChangeLog{},
	)
	if err != nil {
		log.Fatalf("自动迁移失败: %v", err)
	}
	fmt.Println("API文档相关表迁移完成!")

	// 创建导入器
	importer := NewAPIDocImporter(db)

	// 设置文档目录路径
	docsDir := "../../docs/API接口文档/整理汇总"
	
	// 检查目录是否存在
	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		log.Fatalf("文档目录不存在: %s", docsDir)
	}

	// 开始导入
	if err := importer.ImportFromMarkdownFiles(docsDir); err != nil {
		log.Fatalf("导入API文档失败: %v", err)
	}

	fmt.Println("API文档导入完成!")
}