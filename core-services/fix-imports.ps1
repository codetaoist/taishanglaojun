# 修复导入路径脚本
# 将 github.com/taishanglaojun/core-services 替换为 github.com/codetaoist/taishanglaojun/core-services

$rootPath = "d:\work\taishanglaojun\core-services"
$oldImport = "github.com/taishanglaojun/core-services"
$newImport = "github.com/codetaoist/taishanglaojun/core-services"

Write-Host "开始修复导入路径..." -ForegroundColor Green

# 获取所有 .go 文件
$goFiles = Get-ChildItem -Path $rootPath -Recurse -Filter "*.go" | Where-Object { $_.FullName -notlike "*\.git*" }

$totalFiles = $goFiles.Count
$processedFiles = 0
$modifiedFiles = 0

foreach ($file in $goFiles) {
    $processedFiles++
    Write-Progress -Activity "修复导入路径" -Status "处理文件: $($file.Name)" -PercentComplete (($processedFiles / $totalFiles) * 100)
    
    try {
        $content = Get-Content -Path $file.FullName -Raw -Encoding UTF8
        
        if ($content -match [regex]::Escape($oldImport)) {
            $newContent = $content -replace [regex]::Escape($oldImport), $newImport
            Set-Content -Path $file.FullName -Value $newContent -Encoding UTF8 -NoNewline
            $modifiedFiles++
            Write-Host "已修复: $($file.FullName)" -ForegroundColor Yellow
        }
    }
    catch {
        Write-Warning "处理文件失败: $($file.FullName) - $($_.Exception.Message)"
    }
}

Write-Host "`n修复完成!" -ForegroundColor Green
Write-Host "总文件数: $totalFiles" -ForegroundColor Cyan
Write-Host "已修改文件数: $modifiedFiles" -ForegroundColor Cyan

# 同时修复 go.mod 文件
$goModFiles = Get-ChildItem -Path $rootPath -Recurse -Filter "go.mod" | Where-Object { $_.FullName -notlike "*\.git*" }

Write-Host "`n开始修复 go.mod 文件..." -ForegroundColor Green

foreach ($file in $goModFiles) {
    try {
        $content = Get-Content -Path $file.FullName -Raw -Encoding UTF8
        
        if ($content -match [regex]::Escape($oldImport)) {
            $newContent = $content -replace [regex]::Escape($oldImport), $newImport
            Set-Content -Path $file.FullName -Value $newContent -Encoding UTF8 -NoNewline
            Write-Host "已修复 go.mod: $($file.FullName)" -ForegroundColor Yellow
        }
    }
    catch {
        Write-Warning "处理 go.mod 文件失败: $($file.FullName) - $($_.Exception.Message)"
    }
}

Write-Host "`n所有导入路径修复完成!" -ForegroundColor Green