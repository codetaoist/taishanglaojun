# 太上老君手表应用部署脚本
# Deployment script for Taishanglaojun Watch Applications

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("debug", "release")]
    [string]$BuildType = "debug",
    
    [Parameter(Mandatory=$false)]
    [ValidateSet("wear-os", "apple-watch", "all")]
    [string]$Platform = "all",
    
    [Parameter(Mandatory=$false)]
    [string]$DeviceId = "",
    
    [Parameter(Mandatory=$false)]
    [switch]$Install = $false,
    
    [Parameter(Mandatory=$false)]
    [switch]$Launch = $false
)

# 颜色定义
$Red = "Red"
$Green = "Green"
$Yellow = "Yellow"
$Blue = "Blue"

# 日志函数
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor $Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor $Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor $Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor $Red
}

# 检查命令是否存在
function Test-Command {
    param([string]$Command)
    
    $null = Get-Command $Command -ErrorAction SilentlyContinue
    return $?
}

# 项目路径配置
$ProjectRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
$WatchAppsDir = Join-Path $ProjectRoot "watch-apps"
$AppleWatchDir = Join-Path $WatchAppsDir "apple-watch"
$WearOSDir = Join-Path $WatchAppsDir "wear-os"
$BuildOutputDir = Join-Path $ProjectRoot "build\watch-apps"

Write-Info "开始部署太上老君手表应用..."
Write-Info "构建类型: $BuildType"
Write-Info "目标平台: $Platform"
Write-Info "项目根目录: $ProjectRoot"

# 确保构建输出目录存在
if (!(Test-Path $BuildOutputDir)) {
    New-Item -ItemType Directory -Path $BuildOutputDir -Force | Out-Null
}

# 部署 Wear OS 应用
function Deploy-WearOS {
    Write-Info "开始部署 Wear OS 应用..."
    
    if (!(Test-Path $WearOSDir)) {
        Write-Error "Wear OS 项目目录不存在: $WearOSDir"
        return $false
    }
    
    # 检查 ADB
    if (!(Test-Command "adb")) {
        Write-Error "ADB 命令未找到，请确保 Android SDK 已正确安装并添加到 PATH"
        return $false
    }
    
    # 检查设备连接
    $devices = adb devices
    if ($devices -match "device$") {
        Write-Info "检测到已连接的 Android 设备"
    } else {
        Write-Warning "未检测到已连接的 Android 设备"
        if (!$DeviceId) {
            Write-Info "请连接 Wear OS 设备或启动模拟器"
            return $false
        }
    }
    
    # 确定 APK 文件路径
    $apkFileName = if ($BuildType -eq "release") { "taishanglaojun-watch-release.apk" } else { "taishanglaojun-watch-debug.apk" }
    $apkPath = Join-Path $BuildOutputDir $apkFileName
    
    if (!(Test-Path $apkPath)) {
        Write-Error "APK 文件不存在: $apkPath"
        Write-Info "请先运行构建脚本生成 APK 文件"
        return $false
    }
    
    try {
        # 安装 APK
        if ($Install) {
            Write-Info "安装 Wear OS 应用到设备..."
            if ($DeviceId) {
                $result = adb -s $DeviceId install -r $apkPath
            } else {
                $result = adb install -r $apkPath
            }
            
            if ($LASTEXITCODE -eq 0) {
                Write-Success "Wear OS 应用安装成功"
            } else {
                Write-Error "Wear OS 应用安装失败"
                return $false
            }
        }
        
        # 启动应用
        if ($Launch) {
            Write-Info "启动 Wear OS 应用..."
            $packageName = "com.taishanglaojun.watch"
            $activityName = "com.taishanglaojun.watch.MainActivity"
            
            if ($DeviceId) {
                adb -s $DeviceId shell am start -n "$packageName/$activityName"
            } else {
                adb shell am start -n "$packageName/$activityName"
            }
            
            if ($LASTEXITCODE -eq 0) {
                Write-Success "Wear OS 应用启动成功"
            } else {
                Write-Warning "Wear OS 应用启动可能失败"
            }
        }
        
        return $true
    }
    catch {
        Write-Error "部署 Wear OS 应用时发生错误: $($_.Exception.Message)"
        return $false
    }
}

# 部署 Apple Watch 应用
function Deploy-AppleWatch {
    Write-Info "开始部署 Apple Watch 应用..."
    
    if (!(Test-Path $AppleWatchDir)) {
        Write-Error "Apple Watch 项目目录不存在: $AppleWatchDir"
        return $false
    }
    
    # 检查是否在 macOS 上运行
    if ($env:OS -ne "Windows_NT") {
        Write-Warning "Apple Watch 应用部署需要在 macOS 上进行"
        Write-Info "请在 macOS 设备上使用 Xcode 进行部署"
        return $false
    }
    
    Write-Warning "在 Windows 环境下无法直接部署 Apple Watch 应用"
    Write-Info "请按照以下步骤在 macOS 上部署:"
    Write-Info "1. 在 macOS 上打开 Xcode"
    Write-Info "2. 打开项目文件: $AppleWatchDir\TaishanglaojunWatch.xcodeproj"
    Write-Info "3. 连接 Apple Watch 设备"
    Write-Info "4. 选择目标设备并点击运行"
    
    return $true
}

# 生成部署报告
function New-DeploymentReport {
    $reportPath = Join-Path $BuildOutputDir "deployment-report.txt"
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    
    $report = @"
太上老君手表应用部署报告
========================

部署时间: $timestamp
构建类型: $BuildType
目标平台: $Platform
设备ID: $DeviceId
安装应用: $Install
启动应用: $Launch

部署结果:
"@
    
    $report | Out-File -FilePath $reportPath -Encoding UTF8
    Write-Success "部署报告已保存到: $reportPath"
}

# 主部署逻辑
$deploymentSuccess = $true

switch ($Platform) {
    "wear-os" {
        $deploymentSuccess = Deploy-WearOS
    }
    "apple-watch" {
        $deploymentSuccess = Deploy-AppleWatch
    }
    "all" {
        $wearOSResult = Deploy-WearOS
        $appleWatchResult = Deploy-AppleWatch
        $deploymentSuccess = $wearOSResult -and $appleWatchResult
    }
}

# 生成部署报告
New-DeploymentReport

if ($deploymentSuccess) {
    Write-Success "部署完成！"
} else {
    Write-Error "部署过程中遇到错误"
    exit 1
}

# 显示有用的命令
Write-Info ""
Write-Info "有用的命令:"
Write-Info "查看 Wear OS 设备: adb devices"
Write-Info "查看应用日志: adb logcat | findstr taishanglaojun"
Write-Info "卸载应用: adb uninstall com.taishanglaojun.watch"