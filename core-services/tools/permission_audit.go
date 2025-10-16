package main

import (
    "flag"
    "fmt"
    "os"
    "strings"
    "time"

    "github.com/codetaoist/taishanglaojun/core-services/internal/config"
    "github.com/codetaoist/taishanglaojun/core-services/internal/database"
    "github.com/google/uuid"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

type Permission struct {
    ID          string    `gorm:"type:char(36);primaryKey"`
    Name        string    `gorm:"type:varchar(255)"`
    Code        string    `gorm:"type:varchar(255)"`
    Description string    `gorm:"type:text"`
    Resource    string    `gorm:"type:varchar(255)"`
    Action      string    `gorm:"type:varchar(255)"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type AuditResult struct {
    ID        string
    Code      string
    Field     string
    Value     string
    Suspected bool
    Reason    string
}

func main() {
    limit := flag.Int("limit", 50, "抽样记录数量")
    applyFix := flag.Bool("apply-fix", false, "对疑似乱码字段执行修复（LATIN1→UTF8）")
    doSeed := flag.Bool("seed", true, "填充缺失的基线权限数据")
    flag.Parse()

    // 初始化日志
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // 加载配置
    cfg, err := config.Load("")
    if err != nil {
        fmt.Printf("加载配置失败: %v\n", err)
        os.Exit(1)
    }

    // 连接数据库（沿用项目内的连接优化：client_encoding/timeout 等）
    dbMgr, err := database.New(database.Config{
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
    }, logger)
    if err != nil {
        fmt.Printf("数据库连接失败: %v\n", err)
        os.Exit(2)
    }
    db := dbMgr.GetDB()

    // 输出当前会话编码信息，便于确认环境
    printEncodingInfo(db, cfg.Database.Type)

    // 抽样检测
    var perms []Permission
    if err := db.Order("created_at ASC").Limit(*limit).Find(&perms).Error; err != nil {
        fmt.Printf("查询权限抽样失败: %v\n", err)
        os.Exit(3)
    }

    var results []AuditResult
    for _, p := range perms {
        results = append(results, auditField(p.ID, p.Code, "name", p.Name))
        results = append(results, auditField(p.ID, p.Code, "description", p.Description))
    }

    // 输出检测结果
    fmt.Printf("\n=== 权限表抽样检测（共 %d 条）===\n", len(perms))
    suspectedCount := 0
    for _, r := range results {
        if r.Suspected {
            suspectedCount++
            fmt.Printf("[疑似乱码] id=%s code=%s field=%s reason=%s value=%q\n", r.ID, r.Code, r.Field, r.Reason, r.Value)
        }
    }
    fmt.Printf("\n检测完成：疑似乱码记录 %d 条\n", suspectedCount)

    // 可选：修复疑似乱码（将 UTF-8 被当作 Latin1 解码导致的串音尝试纠正）
    if *applyFix && suspectedCount > 0 {
        if err := fixSuspectedMojibake(db, cfg.Database.Type, results); err != nil {
            fmt.Printf("执行修复失败: %v\n", err)
            os.Exit(4)
        }
        fmt.Println("修复已执行（LATIN1→UTF8），建议再次运行审计确认结果")
    }

    // 可选：填充基线权限（避免缺失导致菜单/权限筛选异常）
    if *doSeed {
        missing, err := seedBaselinePermissions(db, cfg.Database.Type)
        if err != nil {
            fmt.Printf("填充基线权限失败: %v\n", err)
            os.Exit(5)
        }
        fmt.Printf("\n基线权限填充完成：新增 %d 条\n", missing)
    }
}

func printEncodingInfo(db *gorm.DB, driver string) {
    if driver == "postgres" || driver == "postgresql" {
        type row struct{ V string }
        var srvEnc, cliEnc row
        _ = db.Raw("SELECT current_setting('server_encoding') AS v").Scan(&srvEnc).Error
        _ = db.Raw("SELECT current_setting('client_encoding') AS v").Scan(&cliEnc).Error
        type dbl struct{ Datcollate, Datctype string }
        var d dbl
        _ = db.Raw("SELECT datcollate, datctype FROM pg_database WHERE datname = current_database()").Scan(&d).Error
        fmt.Printf("当前编码(Postgres): server=%s client=%s lc_collate=%s lc_ctype=%s\n", srvEnc.V, cliEnc.V, d.Datcollate, d.Datctype)
        return
    }
    // MySQL 编码信息输出
    type enc struct{ Charset string; Collation string }
    var srv, dbinfo enc
    _ = db.Raw("SELECT @@character_set_server AS charset, @@collation_server AS collation").Scan(&srv).Error
    _ = db.Raw("SELECT @@character_set_database AS charset, @@collation_database AS collation").Scan(&dbinfo).Error
    fmt.Printf("当前编码(MySQL): server=%s/%s database=%s/%s\n", srv.Charset, srv.Collation, dbinfo.Charset, dbinfo.Collation)
}

func auditField(id, code, field, val string) AuditResult {
    suspected, reason := isSuspectedMojibake(val)
    return AuditResult{
        ID:        id,
        Code:      code,
        Field:     field,
        Value:     val,
        Suspected: suspected,
        Reason:    reason,
    }
}

// 经验规则：
// - 包含 U+FFFD（替换符）直接判定疑似损坏
// - 非 ASCII 比例较高但缺少 CJK，且含常见串音字符（Ã, Â, â, €, ™, œ, �）
// - 值为空不判定损坏
func isSuspectedMojibake(s string) (bool, string) {
    if len(s) == 0 {
        return false, "empty"
    }
    suspects := "ÃÂâ€™œ�"
    r := []rune(s)
    total := len(r)
    var hasCJK bool
    var nonASCII, suspectCount, replacement int
    for _, ch := range r {
        if ch == '\uFFFD' {
            replacement++
        }
        if ch > 0x7F {
            nonASCII++
        }
        if (ch >= 0x4E00 && ch <= 0x9FFF) || (ch >= 0x3400 && ch <= 0x4DBF) || (ch >= 0x20000 && ch <= 0x2FA1F) {
            hasCJK = true
        }
        if strings.ContainsRune(suspects, ch) {
            suspectCount++
        }
    }
    if replacement > 0 {
        return true, "contains U+FFFD replacement"
    }
    ratioNon := float32(nonASCII) / float32(total)
    ratioSus := float32(suspectCount) / float32(total)
    if !hasCJK && ratioNon > 0.3 && ratioSus > 0.1 {
        return true, "latin1-utf8 mojibake suspected"
    }
    return false, "ok"
}

// 对标记为疑似乱码的字段进行可逆修复：convert_from(convert_to(val,'LATIN1'),'UTF8')
func fixSuspectedMojibake(db *gorm.DB, driver string, results []AuditResult) error {
    tx := db.Begin()
    for _, r := range results {
        if !r.Suspected {
            continue
        }
        // 仅修复 name/description 两个字段
        if r.Field != "name" && r.Field != "description" {
            continue
        }
        var query string
        if driver == "postgres" || driver == "postgresql" {
            // PostgreSQL 修复：LATIN1 → UTF8
            query = fmt.Sprintf("UPDATE permissions SET %s = convert_from(convert_to(%s, 'LATIN1'), 'UTF8') WHERE id = ?", r.Field, r.Field)
        } else {
            // MySQL 修复：尝试将 UTF-8 被按 latin1 解码的文本纠正为 utf8mb4
            // 常见写法：CONVERT(CAST(CONVERT(val USING latin1) AS BINARY) USING utf8mb4)
            query = fmt.Sprintf("UPDATE permissions SET %s = CONVERT(CAST(CONVERT(%s USING latin1) AS BINARY) USING utf8mb4) WHERE id = ?", r.Field, r.Field)
        }
        if err := tx.Exec(query, r.ID).Error; err != nil {
            tx.Rollback()
            return fmt.Errorf("修复失败 id=%s field=%s: %w", r.ID, r.Field, err)
        }
    }
    return tx.Commit().Error
}

// 基线权限列表（与 scripts/init-permissions.go 对齐，避免缺项）
func baseline() []Permission {
    return []Permission{
        {Name: "菜单查看", Code: "menu:read", Description: "查看菜单列表和详情", Resource: "menu", Action: "read"},
        {Name: "菜单创建", Code: "menu:create", Description: "创建新菜单", Resource: "menu", Action: "create"},
        {Name: "菜单编辑", Code: "menu:update", Description: "编辑菜单信息", Resource: "menu", Action: "update"},
        {Name: "菜单删除", Code: "menu:delete", Description: "删除菜单", Resource: "menu", Action: "delete"},
        {Name: "菜单树查看", Code: "menu:tree", Description: "查看菜单树结构", Resource: "menu", Action: "tree"},

        {Name: "用户查看", Code: "user:read", Description: "查看用户列表和详情", Resource: "user", Action: "read"},
        {Name: "用户创建", Code: "user:create", Description: "创建新用户", Resource: "user", Action: "create"},
        {Name: "用户编辑", Code: "user:update", Description: "编辑用户信息", Resource: "user", Action: "update"},
        {Name: "用户删除", Code: "user:delete", Description: "删除用户", Resource: "user", Action: "delete"},
        {Name: "用户状态管理", Code: "user:status", Description: "启用/禁用用户", Resource: "user", Action: "status"},

        {Name: "角色查看", Code: "role:read", Description: "查看角色列表和详情", Resource: "role", Action: "read"},
        {Name: "角色创建", Code: "role:create", Description: "创建新角色", Resource: "role", Action: "create"},
        {Name: "角色编辑", Code: "role:update", Description: "编辑角色信息", Resource: "role", Action: "update"},
        {Name: "角色删除", Code: "role:delete", Description: "删除角色", Resource: "role", Action: "delete"},
        {Name: "角色权限分配", Code: "role:assign", Description: "为角色分配权限", Resource: "role", Action: "assign"},

        {Name: "权限查看", Code: "permission:read", Description: "查看权限列表和详情", Resource: "permission", Action: "read"},
        {Name: "权限创建", Code: "permission:create", Description: "创建新权限", Resource: "permission", Action: "create"},
        {Name: "权限编辑", Code: "permission:update", Description: "编辑权限信息", Resource: "permission", Action: "update"},
        {Name: "权限删除", Code: "permission:delete", Description: "删除权限", Resource: "permission", Action: "delete"},

        {Name: "系统配置查看", Code: "system:config:read", Description: "查看系统配置", Resource: "system", Action: "config:read"},
        {Name: "系统配置编辑", Code: "system:config:update", Description: "编辑系统配置", Resource: "system", Action: "config:update"},
        {Name: "系统日志查看", Code: "system:log:read", Description: "查看系统日志", Resource: "system", Action: "log:read"},
        {Name: "系统监控", Code: "system:monitor", Description: "系统监控和状态查看", Resource: "system", Action: "monitor"},

        {Name: "文件上传", Code: "file:upload", Description: "上传文件", Resource: "file", Action: "upload"},
        {Name: "文件下载", Code: "file:download", Description: "下载文件", Resource: "file", Action: "download"},
        {Name: "文件删除", Code: "file:delete", Description: "删除文件", Resource: "file", Action: "delete"},
        {Name: "文件查看", Code: "file:read", Description: "查看文件列表", Resource: "file", Action: "read"},

        {Name: "数据导出", Code: "data:export", Description: "导出数据", Resource: "data", Action: "export"},
        {Name: "数据导入", Code: "data:import", Description: "导入数据", Resource: "data", Action: "import"},
        {Name: "数据备份", Code: "data:backup", Description: "数据备份", Resource: "data", Action: "backup"},
        {Name: "数据恢复", Code: "data:restore", Description: "数据恢复", Resource: "data", Action: "restore"},
    }
}

func seedBaselinePermissions(db *gorm.DB, driver string) (int, error) {
    base := baseline()
    created := 0
    for _, perm := range base {
        var cnt int64
        if err := db.Model(&Permission{}).Where("code = ?", perm.Code).Count(&cnt).Error; err != nil {
            return created, err
        }
        if cnt == 0 {
            // 插入时显式提供 UUID（适配 MySQL/PG 表结构）
            newID := uuid.New().String()
            var err error
            if driver == "postgres" || driver == "postgresql" {
                err = db.Exec(
                    "INSERT INTO permissions (id, name, code, description, resource, action, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())",
                    newID, perm.Name, perm.Code, perm.Description, perm.Resource, perm.Action,
                ).Error
            } else {
                // MySQL 使用 NOW() 同样适配
                err = db.Exec(
                    "INSERT INTO permissions (id, name, code, description, resource, action, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())",
                    newID, perm.Name, perm.Code, perm.Description, perm.Resource, perm.Action,
                ).Error
            }
            if err != nil {
                return created, err
            }
            created++
        }
    }
    return created, nil
}