# Wear OS 应用自动化部署脚本
# 作者: 太上老君团队
# 版本: 1.0.0
# 描述: 自动化构建、测试和部署 Wear OS 应用

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("debug", "release")]
    [string]$BuildType = "debug",
    
    [Parameter(Mandatory=$false)]
    [switch]$SkipTests,
    
    [Parameter(Mandatory=$false)]
    [switch]$SkipLint,
    
    [Parameter(Mandatory=$false)]
    [switch]$InstallToDevice,
    
    [Parameter(Mandatory=$false)]
    [switch]$GenerateApk,
    
    [Parameter(Mandatory=$false)]
    [switch]$CleanBuild,
    
    [Parameter(Mandatory=$false)]
    [string]$OutputDir = ".\build\outputs"
)

# 设置错误处理
$ErrorActionPreference = "Stop"

# 颜色输出函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    
    $originalColor = $Host.UI.RawUI.ForegroundColor
    $Host.UI.RawUI.ForegroundColor = $Color
    Write-Host $Message
    $Host.UI.RawUI.ForegroundColor = $originalColor
}

# 检查必要工具
function Test-Prerequisites {
    Write-ColorOutput "检查部署前提条件..." "Yellow"
    
    # 检查 Java
    try {
        $javaVersion = java -version 2>&1 | Select-String "version"
        Write-ColorOutput "✓ Java: $javaVersion" "Green"
    } catch {
        Write-ColorOutput "✗ Java 未安装或未配置" "Red"
        exit 1
    }
    
    # 检查 Android SDK
    if (-not $env:ANDROID_HOME) {
        Write-ColorOutput "✗ ANDROID_HOME 环境变量未设置" "Red"
        exit 1
    } else {
        Write-ColorOutput "✓ Android SDK: $env:ANDROID_HOME" "Green"
    }
    
    # 检查 ADB
    try {
        $adbVersion = adb version 2>&1 | Select-String "Android Debug Bridge"
        Write-ColorOutput "✓ ADB: $adbVersion" "Green"
    } catch {
        Write-ColorOutput "✗ ADB 未安装或未配置" "Red"
        exit 1
    }
    
    # 检查项目目录
    $projectDir = "d:\work\taishanglaojun\watch-apps\wear-os"
    if (-not (Test-Path $projectDir)) {
        Write-ColorOutput "✗ 项目目录不存在: $projectDir" "Red"
        exit 1
    } else {
        Write-ColorOutput "✓ 项目目录: $projectDir" "Green"
    }
    
    Write-ColorOutput "前提条件检查完成" "Green"
}

# 清理构建
function Invoke-CleanBuild {
    Write-ColorOutput "清理构建目录..." "Yellow"
    
    Set-Location "d:\work\taishanglaojun\watch-apps\wear-os"
    
    try {
        .\gradlew.bat clean
        Write-ColorOutput "✓ 构建目录清理完成" "Green"
    } catch {
        Write-ColorOutput "✗ 构建目录清理失败: $_" "Red"
        exit 1
    }
}

# 运行 Lint 检查
function Invoke-LintCheck {
    if ($SkipLint) {
        Write-ColorOutput "跳过 Lint 检查" "Yellow"
        return
    }
    
    Write-ColorOutput "运行 Lint 检查..." "Yellow"
    
    try {
        .\gradlew.bat lint
        Write-ColorOutput "✓ Lint 检查完成" "Green"
        
        # 检查 Lint 报告
        $lintReport = ".\app\build\reports\lint\lint.html"
        if (Test-Path $lintReport) {
            Write-ColorOutput "Lint 报告生成: $lintReport" "Cyan"
        }
    } catch {
        Write-ColorOutput "⚠ Lint 检查发现问题，但继续构建" "Yellow"
    }
}

# 运行单元测试
function Invoke-UnitTests {
    if ($SkipTests) {
        Write-ColorOutput "跳过单元测试" "Yellow"
        return
    }
    
    Write-ColorOutput "运行单元测试..." "Yellow"
    
    try {
        .\gradlew.bat test
        Write-ColorOutput "✓ 单元测试完成" "Green"
        
        # 检查测试报告
        $testReport = ".\app\build\reports\tests\testDebugUnitTest\index.html"
        if (Test-Path $testReport) {
            Write-ColorOutput "测试报告生成: $testReport" "Cyan"
        }
    } catch {
        Write-ColorOutput "✗ 单元测试失败: $_" "Red"
        exit 1
    }
}

# 构建应用
function Invoke-BuildApp {
    Write-ColorOutput "构建 $BuildType 版本..." "Yellow"
    
    try {
        if ($BuildType -eq "release") {
            .\gradlew.bat assembleRelease
            Write-ColorOutput "✓ Release 版本构建完成" "Green"
        } else {
            .\gradlew.bat assembleDebug
            Write-ColorOutput "✓ Debug 版本构建完成" "Green"
        }
    } catch {
        Write-ColorOutput "✗ 应用构建失败: $_" "Red"
        exit 1
    }
}

# 生成 APK
function Copy-ApkFiles {
    if (-not $GenerateApk) {
        return
    }
    
    Write-ColorOutput "复制 APK 文件..." "Yellow"
    
    # 创建输出目录
    if (-not (Test-Path $OutputDir)) {
        New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
    }
    
    # 复制 APK 文件
    $apkSource = ".\app\build\outputs\apk\$BuildType\*.apk"
    $apkFiles = Get-ChildItem $apkSource -ErrorAction SilentlyContinue
    
    if ($apkFiles) {
        foreach ($apk in $apkFiles) {
            $destination = Join-Path $OutputDir $apk.Name
            Copy-Item $apk.FullName $destination -Force
            Write-ColorOutput "✓ APK 复制到: $destination" "Green"
        }
    } else {
        Write-ColorOutput "⚠ 未找到 APK 文件" "Yellow"
    }
}

# 检查连接的设备
function Test-ConnectedDevices {
    Write-ColorOutput "检查连接的设备..." "Yellow"
    
    try {
        $devices = adb devices | Select-String "device$"
        if ($devices.Count -eq 0) {
            Write-ColorOutput "⚠ 未检测到连接的设备" "Yellow"
            return $false
        } else {
            Write-ColorOutput "✓ 检测到 $($devices.Count) 个设备" "Green"
            foreach ($device in $devices) {
                Write-ColorOutput "  - $device" "Cyan"
            }
            return $true
        }
    } catch {
        Write-ColorOutput "✗ 设备检查失败: $_" "Red"
        return $false
    }
}

# 安装到设备
function Install-ToDevice {
    if (-not $InstallToDevice) {
        return
    }
    
    if (-not (Test-ConnectedDevices)) {
        Write-ColorOutput "跳过设备安装（无连接设备）" "Yellow"
        return
    }
    
    Write-ColorOutput "安装应用到设备..." "Yellow"
    
    try {
        if ($BuildType -eq "release") {
            .\gradlew.bat installRelease
        } else {
            .\gradlew.bat installDebug
        }
        Write-ColorOutput "✓ 应用安装完成" "Green"
    } catch {
        Write-ColorOutput "✗ 应用安装失败: $_" "Red"
        exit 1
    }
}

# 生成部署报告
function New-DeploymentReport {
    Write-ColorOutput "生成部署报告..." "Yellow"
    
    $reportPath = Join-Path $OutputDir "deployment-report.txt"
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    
    $report = @"
Wear OS 应用部署报告
====================

部署时间: $timestamp
构建类型: $BuildType
跳过测试: $SkipTests
跳过 Lint: $SkipLint
安装到设备: $InstallToDevice
生成 APK: $GenerateApk
清理构建: $CleanBuild

项目信息:
- 项目路径: d:\work\taishanglaojun\watch-apps\wear-os
- 输出目录: $OutputDir

构建产物:
"@

    # 添加 APK 信息
    $apkFiles = Get-ChildItem "$OutputDir\*.apk" -ErrorAction SilentlyContinue
    if ($apkFiles) {
        $report += "`n- APK 文件:`n"
        foreach ($apk in $apkFiles) {
            $size = [math]::Round($apk.Length / 1MB, 2)
            $report += "  * $($apk.Name) ($size MB)`n"
        }
    }
    
    # 添加报告文件信息
    $reportFiles = @(
        ".\app\build\reports\lint\lint.html",
        ".\app\build\reports\tests\testDebugUnitTest\index.html"
    )
    
    $report += "`n- 报告文件:`n"
    foreach ($file in $reportFiles) {
        if (Test-Path $file) {
            $report += "  * $file`n"
        }
    }
    
    $report | Out-File -FilePath $reportPath -Encoding UTF8
    Write-ColorOutput "✓ 部署报告生成: $reportPath" "Green"
}

# 主函数
function Main {
    Write-ColorOutput "=== Wear OS 应用自动化部署 ===" "Cyan"
    Write-ColorOutput "构建类型: $BuildType" "Cyan"
    
    $startTime = Get-Date
    
    try {
        # 1. 检查前提条件
        Test-Prerequisites
        
        # 2. 切换到项目目录
        Set-Location "d:\work\taishanglaojun\watch-apps\wear-os"
        
        # 3. 清理构建（如果需要）
        if ($CleanBuild) {
            Invoke-CleanBuild
        }
        
        # 4. 运行 Lint 检查
        Invoke-LintCheck
        
        # 5. 运行单元测试
        Invoke-UnitTests
        
        # 6. 构建应用
        Invoke-BuildApp
        
        # 7. 复制 APK 文件
        Copy-ApkFiles
        
        # 8. 安装到设备
        Install-ToDevice
        
        # 9. 生成部署报告
        New-DeploymentReport
        
        $endTime = Get-Date
        $duration = $endTime - $startTime
        
        Write-ColorOutput "=== 部署完成 ===" "Green"
        Write-ColorOutput "总耗时: $($duration.TotalMinutes.ToString('F2')) 分钟" "Green"
        
    } catch {
        Write-ColorOutput "=== 部署失败 ===" "Red"
        Write-ColorOutput "错误: $_" "Red"
        exit 1
    }
}

# 执行主函数
Main