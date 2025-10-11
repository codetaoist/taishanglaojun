# 修复相对导入路径脚本

$rootPath = "d:\work\taishanglaojun\core-services"

Write-Host "开始修复相对导入路径..." -ForegroundColor Green

# 定义需要修复的相对导入路径映射
$importMappings = @{
    '"../audit"' = '"github.com/codetaoist/taishanglaojun/core-services/audit"'
    '"../permission"' = '"github.com/codetaoist/taishanglaojun/core-services/permission"'
    '"../vision"' = '"github.com/codetaoist/taishanglaojun/core-services/ai-integration/vision"'
    '"../voice"' = '"github.com/codetaoist/taishanglaojun/core-services/ai-integration/voice"'
}

# 获取所有 .go 文件
$goFiles = Get-ChildItem -Path $rootPath -Recurse -Filter "*.go" | Where-Object { $_.FullName -notlike "*\.git*" }

$modifiedFiles = 0

foreach ($file in $goFiles) {
    $content = Get-Content -Path $file.FullName -Raw -Encoding UTF8
    $originalContent = $content
    
    # 应用所有映射
    foreach ($mapping in $importMappings.GetEnumerator()) {
        $oldImport = $mapping.Key
        $newImport = $mapping.Value
        
        if ($content -match [regex]::Escape($oldImport)) {
            $content = $content -replace [regex]::Escape($oldImport), $newImport
        }
    }
    
    # 如果内容有变化，保存文件
    if ($content -ne $originalContent) {
        $content | Set-Content -Path $file.FullName -Encoding UTF8 -NoNewline
        $modifiedFiles++
        Write-Host "已修复: $($file.FullName)" -ForegroundColor Yellow
    }
}

Write-Host "`n修复完成!" -ForegroundColor Green
Write-Host "已修改文件数: $modifiedFiles" -ForegroundColor Cyan