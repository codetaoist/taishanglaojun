# 码道 CLI 构建脚本 (PowerShell)

param(
    [string]$Version = "1.0.0"
)

Write-Host "🚀 开始构建码道 CLI 工具..." -ForegroundColor Green

# 项目信息
$AppName = "lo"
$BuildTime = (Get-Date).ToUniversalTime().ToString("yyyy-MM-dd HH:mm:ss UTC")
try {
    $GitCommit = (git rev-parse --short HEAD 2>$null)
} catch {
    $GitCommit = "unknown"
}

# 构建标志
$LDFlags = "-X 'main.Version=$Version' -X 'main.BuildTime=$BuildTime' -X 'main.GitCommit=$GitCommit'"

# 清理之前的构建
Write-Host "🧹 清理构建目录..." -ForegroundColor Yellow
if (Test-Path "dist") {
    Remove-Item -Recurse -Force "dist"
}
New-Item -ItemType Directory -Path "dist" | Out-Null

# 构建不同平台的二进制文件
Write-Host "🔨 构建多平台二进制文件..." -ForegroundColor Cyan

# 构建配置
$BuildConfigs = @(
    @{ OS = "linux"; Arch = "amd64"; Output = "lo-linux-amd64" },
    @{ OS = "linux"; Arch = "arm64"; Output = "lo-linux-arm64" },
    @{ OS = "darwin"; Arch = "amd64"; Output = "lo-darwin-amd64" },
    @{ OS = "darwin"; Arch = "arm64"; Output = "lo-darwin-arm64" },
    @{ OS = "windows"; Arch = "amd64"; Output = "lo-windows-amd64.exe" },
    @{ OS = "windows"; Arch = "arm64"; Output = "lo-windows-arm64.exe" }
)

foreach ($config in $BuildConfigs) {
    Write-Host "构建 $($config.OS) $($config.Arch)..." -ForegroundColor White
    
    $env:CGO_ENABLED = "0"
    $env:GOOS = $config.OS
    $env:GOARCH = $config.Arch
    
    $outputPath = "dist\$($config.Output)"
    
    & go build -ldflags $LDFlags -o $outputPath .\cmd\lo
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ 构建 $($config.OS) $($config.Arch) 失败" -ForegroundColor Red
        exit 1
    }
}

# 创建本地可执行文件副本
Write-Host "🔗 创建本地可执行文件..." -ForegroundColor Yellow
Copy-Item "dist\lo-windows-amd64.exe" "dist\lo.exe"

# 显示构建结果
Write-Host "📦 构建完成！生成的文件：" -ForegroundColor Green
Get-ChildItem "dist" | Format-Table Name, Length, LastWriteTime

# 计算文件校验和
Write-Host "📊 文件信息：" -ForegroundColor Cyan
Get-ChildItem "dist\lo-*" | ForEach-Object {
    $size = [math]::Round($_.Length / 1MB, 2)
    $hash = (Get-FileHash $_.FullName -Algorithm SHA256).Hash
    Write-Host "$($_.Name): ${size}MB, SHA256: $hash" -ForegroundColor White
}

Write-Host "✅ 构建完成！" -ForegroundColor Green
Write-Host "💡 使用方法：" -ForegroundColor Yellow
Write-Host "   .\dist\lo.exe --help"
Write-Host "   .\dist\lo.exe version"

# 清理环境变量
Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue