package main

import (
    "fmt"
    "log"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    // 使用与本地开发一致的DSN（如需修改，请根据实际环境调整）
    dsn := "host=localhost user=postgres password=password dbname=taishanglaojun port=5432 sslmode=disable"

    fmt.Println("连接到 PostgreSQL 用于编码检测...")
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("数据库连接失败: %v", err)
    }

    // 查询服务器与客户端编码
    var serverEncoding string
    if err := db.Raw("SELECT current_setting('server_encoding')").Scan(&serverEncoding).Error; err != nil {
        log.Fatalf("查询 server_encoding 失败: %v", err)
    }

    var clientEncoding string
    if err := db.Raw("SELECT current_setting('client_encoding')").Scan(&clientEncoding).Error; err != nil {
        log.Fatalf("查询 client_encoding 失败: %v", err)
    }

    type CollateInfo struct {
        DatCollate string `gorm:"column:datcollate"`
        DatCType   string `gorm:"column:datctype"`
    }
    var collate CollateInfo
    if err := db.Raw("SELECT datcollate, datctype FROM pg_database WHERE datname = current_database()").Scan(&collate).Error; err != nil {
        log.Fatalf("查询数据库排序/类型失败: %v", err)
    }

    fmt.Printf("\n编码信息:\n")
    fmt.Printf("server_encoding: %s\n", serverEncoding)
    fmt.Printf("client_encoding: %s\n", clientEncoding)
    fmt.Printf("datcollate: %s\n", collate.DatCollate)
    fmt.Printf("datctype: %s\n", collate.DatCType)

    // 查询 permissions 表的示例数据
    fmt.Println("\n示例权限数据 (前5条):")
    type PermissionRow struct {
        Name        string `gorm:"column:name"`
        Description string `gorm:"column:description"`
        Code        string `gorm:"column:code"`
    }
    var rows []PermissionRow
    if err := db.Raw("SELECT name, description, code FROM permissions ORDER BY id LIMIT 5").Scan(&rows).Error; err != nil {
        log.Printf("查询 permissions 示例数据失败: %v", err)
    } else {
        for i, r := range rows {
            fmt.Printf("%d) name=%s | code=%s | description=%s\n", i+1, r.Name, r.Code, r.Description)
        }
    }

    fmt.Println("\n编码检测完成。")
}