package codegen

import (
	"fmt"
	"strings"
	"text/template"
	"bytes"
)

// CodeGenerator 代码生成器
type CodeGenerator struct {
	templates map[string]*template.Template
}

// NewCodeGenerator 创建新的代码生成器
func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{
		templates: make(map[string]*template.Template),
	}
}

// RegisterTemplate 注册代码模板
func (cg *CodeGenerator) RegisterTemplate(name, templateStr string) error {
	tmpl, err := template.New(name).Parse(templateStr)
	if err != nil {
		return fmt.Errorf("解析模板失败: %w", err)
	}
	
	cg.templates[name] = tmpl
	return nil
}

// GenerateCode 生成代码
func (cg *CodeGenerator) GenerateCode(templateName string, data interface{}) (string, error) {
	tmpl, exists := cg.templates[templateName]
	if !exists {
		return "", fmt.Errorf("模板 %s 不存在", templateName)
	}
	
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("执行模板失败: %w", err)
	}
	
	return buf.String(), nil
}

// GoStructTemplate Go结构体模板
const GoStructTemplate = `package {{.Package}}

// {{.Name}} {{.Description}}
type {{.Name}} struct {
{{range .Fields}}	{{.Name}} {{.Type}} \`json:"{{.JSONTag}}"\` // {{.Comment}}
{{end}}}

// New{{.Name}} 创建新的{{.Name}}实例
func New{{.Name}}() *{{.Name}} {
	return &{{.Name}}{}
}
`

// GoHandlerTemplate Go HTTP处理器模板
const GoHandlerTemplate = `package {{.Package}}

import (
	"encoding/json"
	"net/http"
)

// {{.Name}}Handler {{.Description}}
func {{.Name}}Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		{{.Name}}Get(w, r)
	case http.MethodPost:
		{{.Name}}Post(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// {{.Name}}Get 处理GET请求
func {{.Name}}Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "{{.Name}} GET endpoint"})
}

// {{.Name}}Post 处理POST请求
func {{.Name}}Post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "{{.Name}} POST endpoint"})
}
`

// InitDefaultTemplates 初始化默认模板
func (cg *CodeGenerator) InitDefaultTemplates() error {
	templates := map[string]string{
		"go-struct":  GoStructTemplate,
		"go-handler": GoHandlerTemplate,
	}
	
	for name, tmpl := range templates {
		if err := cg.RegisterTemplate(name, tmpl); err != nil {
			return fmt.Errorf("注册模板 %s 失败: %w", name, err)
		}
	}
	
	return nil
}

// FormatCode 格式化代码
func FormatCode(code string) string {
	// 简单的代码格式化
	lines := strings.Split(code, "\n")
	var formatted []string
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			formatted = append(formatted, trimmed)
		}
	}
	
	return strings.Join(formatted, "\n")
}