# 批量修复 UTF-8 编码问题
$files = @(
    "ai-integration\services\crossmodal_service.go",
    "community\handlers\chat_handler.go", 
    "community\services\message_persistence.go",
    "community\routes.go",
    "consciousness\models\evolution.go",
    "consciousness\coordinators\three_axis_coordinator.go",
    "consciousness\engines\evolution_components.go",
    "consciousness\grpc\consciousness_server.go",
    "consciousness\handlers\coordination_handler.go",
    "cultural-wisdom\models\wisdom.go",
    "cultural-wisdom\services\favorites_service.go",
    "cultural-wisdom\handlers\ai_handler.go",
    "location-tracking\handlers\interfaces.go",
    "location-tracking\models\location_point.go"
)

foreach ($file in $files) {
    if (Test-Path $file) {
        Write-Host "修复文件: $file"
        $content = Get-Content $file -Raw -Encoding UTF8
        # 替换常见的编码问题字符
        $content = $content -replace '�', '器'
        $content = $content -replace '锟', ''
        $content = $content -replace '斤拷', ''
        # 保存文件
        $content | Out-File $file -Encoding UTF8 -NoNewline
    }
}

Write-Host "编码修复完成"