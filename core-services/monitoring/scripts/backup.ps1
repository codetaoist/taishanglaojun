# 太上老君监控系统备份 PowerShell 脚本
# 用于备份配置、数据和日志

param(
    [Parameter(Position = 0)]
    [ValidateSet("backup", "restore", "list", "cleanup", "verify", "help")]
    [string]$Command = "backup",
    
    [Alias("t")]
    [ValidateSet("config", "data", "logs", "all")]
    [string]$Type = "all",
    
    [Alias("d")]
    [string]$BackupDir = $env:BACKUP_DIR,
    
    [Alias("e")]
    [ValidateSet("development", "staging", "production")]
    [string]$Environment = "",
    
    [Alias("n")]
    [string]$Name = "",
    
    [Alias("c")]
    [ValidateSet("gzip", "zip", "none")]
    [string]$Compression = "zip",
    
    [Alias("r")]
    [int]$Retention = 30,
    
    [Alias("f")]
    [switch]$Force,
    
    [Alias("v")]
    [switch]$Verbose,
    
    [Alias("q")]
    [switch]$Quiet,
    
    [switch]$DryRun,
    [string[]]$Exclude = @(),
    [string[]]$Include = @(),
    [switch]$Encrypt,
    [string]$Password = "",
    [switch]$NoColor,
    [switch]$Help
)

# 脚本配置
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$ServiceName = "monitoring"

# 默认配置
$DefaultBackupDir = "C:\Backups\monitoring"
$DefaultRetentionDays = 30
$DefaultCompression = "zip"

# 如果指定了帮助参数，显示帮助
if ($Help) {
    $Command = "help"
}

# 设置默认值
if (-not $BackupDir) {
    $BackupDir = $DefaultBackupDir
}

if ($Retention -eq 0) {
    $Retention = $DefaultRetentionDays
}

# 颜色输出函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    
    if ($NoColor -or $Quiet) {
        if (-not $Quiet) {
            Write-Host $Message
        }
        return
    }
    
    $colorMap = @{
        "Red" = "Red"
        "Green" = "Green"
        "Yellow" = "Yellow"
        "Blue" = "Blue"
        "Cyan" = "Cyan"
        "Magenta" = "Magenta"
        "White" = "White"
    }
    
    Write-Host $Message -ForegroundColor $colorMap[$Color]
}

# 日志函数
function Write-Info {
    param([string]$Message)
    Write-ColorOutput "[INFO] $Message" "Blue"
}

function Write-Success {
    param([string]$Message)
    Write-ColorOutput "[SUCCESS] $Message" "Green"
}

function Write-Warning {
    param([string]$Message)
    Write-ColorOutput "[WARNING] $Message" "Yellow"
}

function Write-Error {
    param([string]$Message)
    Write-ColorOutput "[ERROR] $Message" "Red"
}

# 显示帮助信息
function Show-Help {
    @"
太上老君监控系统备份 PowerShell 脚本

用法: .\backup.ps1 [选项] [命令]

命令:
    backup      执行备份
    restore     恢复备份
    list        列出备份
    cleanup     清理旧备份
    verify      验证备份
    help        显示此帮助信息

选项:
    -Type, -t TYPE          备份类型 (config|data|logs|all) [默认: all]
    -BackupDir, -d DIR      备份目录 [默认: $DefaultBackupDir]
    -Environment, -e ENV    环境 (development|staging|production)
    -Name, -n NAME          备份名称 [默认: 自动生成]
    -Compression, -c TYPE   压缩类型 (gzip|zip|none) [默认: $DefaultCompression]
    -Retention, -r DAYS     保留天数 [默认: $DefaultRetentionDays]
    -Force, -f              强制执行
    -Verbose, -v            详细输出
    -Quiet, -q              静默模式
    -DryRun                 试运行模式
    -Exclude PATTERN        排除模式
    -Include PATTERN        包含模式
    -Encrypt                加密备份
    -Password PASSWORD      加密密码
    -NoColor                禁用颜色输出

环境变量:
    BACKUP_DIR              备份目录
    BACKUP_RETENTION_DAYS   保留天数
    BACKUP_COMPRESSION      压缩类型
    BACKUP_ENCRYPT_KEY      加密密钥
    DATABASE_URL            数据库连接地址
    REDIS_URL               Redis 连接地址

示例:
    .\backup.ps1 backup
    .\backup.ps1 backup -Type config -Environment production
    .\backup.ps1 restore -Name monitoring-20231201-120000
    .\backup.ps1 list -Environment production
    .\backup.ps1 cleanup -Retention 7

"@
}

# 检查依赖
function Test-Dependencies {
    $deps = @()
    
    switch ($Compression) {
        "gzip" { $deps += "7z" }
        "zip" { $deps += "7z" }
    }
    
    if ($Encrypt) {
        $deps += "gpg"
    }
    
    foreach ($dep in $deps) {
        if (-not (Get-Command $dep -ErrorAction SilentlyContinue)) {
            Write-Error "依赖未安装: $dep"
            exit 1
        }
    }
}

# 创建备份目录
function New-BackupDirectory {
    param([string]$Path)
    
    if ($DryRun) {
        Write-Info "[DRY RUN] 创建备份目录: $Path"
        return
    }
    
    if (-not (Test-Path $Path)) {
        New-Item -ItemType Directory -Path $Path -Force | Out-Null
        Write-Info "创建备份目录: $Path"
    }
}

# 生成备份名称
function New-BackupName {
    param(
        [string]$Type,
        [string]$Environment
    )
    
    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    
    if ($Environment) {
        return "${ServiceName}-${Environment}-${Type}-${timestamp}"
    } else {
        return "${ServiceName}-${Type}-${timestamp}"
    }
}

# 获取压缩扩展名
function Get-CompressionExtension {
    param([string]$CompressionType)
    
    switch ($CompressionType) {
        "gzip" { return ".tar.gz" }
        "zip" { return ".zip" }
        "none" { return ".tar" }
    }
}

# 备份配置文件
function Backup-Config {
    param(
        [string]$BackupDir,
        [string]$BackupName
    )
    
    Write-Info "备份配置文件..."
    
    $configDirs = @(
        "$ProjectRoot\config",
        "$ProjectRoot\k8s",
        "$ProjectRoot\docker",
        "$ProjectRoot\scripts"
    )
    
    $configFiles = @(
        "$ProjectRoot\Dockerfile",
        "$ProjectRoot\docker-compose.yml",
        "$ProjectRoot\package.json",
        "$ProjectRoot\go.mod",
        "$ProjectRoot\Makefile",
        "$ProjectRoot\.env.example"
    )
    
    $tempDir = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_ }
    $configBackupDir = Join-Path $tempDir "config"
    New-Item -ItemType Directory -Path $configBackupDir -Force | Out-Null
    
    # 复制配置目录
    foreach ($dir in $configDirs) {
        if (Test-Path $dir) {
            $destDir = Join-Path $configBackupDir (Split-Path $dir -Leaf)
            Copy-Item -Path $dir -Destination $destDir -Recurse -Force
            Write-Info "复制配置目录: $dir"
        }
    }
    
    # 复制配置文件
    foreach ($file in $configFiles) {
        if (Test-Path $file) {
            Copy-Item -Path $file -Destination $configBackupDir -Force
            Write-Info "复制配置文件: $file"
        }
    }
    
    # 创建压缩包
    $ext = Get-CompressionExtension $Compression
    $archivePath = Join-Path $BackupDir "${BackupName}-config${ext}"
    
    if ($DryRun) {
        Write-Info "[DRY RUN] 创建配置备份: $archivePath"
    } else {
        switch ($Compression) {
            "zip" {
                Compress-Archive -Path "$configBackupDir\*" -DestinationPath $archivePath -Force
            }
            "gzip" {
                & 7z a -ttar -so "$archivePath.tar" "$configBackupDir\*" | & 7z a -tgzip -si "$archivePath"
            }
            "none" {
                & 7z a -ttar "$archivePath" "$configBackupDir\*"
            }
        }
        Write-Success "配置备份完成: $archivePath"
    }
    
    # 清理临时目录
    Remove-Item $tempDir -Recurse -Force
}

# 备份数据
function Backup-Data {
    param(
        [string]$BackupDir,
        [string]$BackupName
    )
    
    Write-Info "备份数据..."
    
    $dataDirs = @(
        "C:\ProgramData\monitoring",
        "C:\monitoring\data",
        "$ProjectRoot\data"
    )
    
    $tempDir = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_ }
    $dataBackupDir = Join-Path $tempDir "data"
    New-Item -ItemType Directory -Path $dataBackupDir -Force | Out-Null
    
    # 备份文件数据
    foreach ($dir in $dataDirs) {
        if (Test-Path $dir) {
            $destDir = Join-Path $dataBackupDir (Split-Path $dir -Leaf)
            Copy-Item -Path $dir -Destination $destDir -Recurse -Force
            Write-Info "复制数据目录: $dir"
        }
    }
    
    # 备份数据库
    if ($env:DATABASE_URL) {
        Backup-Database $dataBackupDir
    }
    
    # 备份 Redis
    if ($env:REDIS_URL) {
        Backup-Redis $dataBackupDir
    }
    
    # 创建压缩包
    $ext = Get-CompressionExtension $Compression
    $archivePath = Join-Path $BackupDir "${BackupName}-data${ext}"
    
    if ($DryRun) {
        Write-Info "[DRY RUN] 创建数据备份: $archivePath"
    } else {
        switch ($Compression) {
            "zip" {
                Compress-Archive -Path "$dataBackupDir\*" -DestinationPath $archivePath -Force
            }
            "gzip" {
                & 7z a -ttar -so "$archivePath.tar" "$dataBackupDir\*" | & 7z a -tgzip -si "$archivePath"
            }
            "none" {
                & 7z a -ttar "$archivePath" "$dataBackupDir\*"
            }
        }
        Write-Success "数据备份完成: $archivePath"
    }
    
    # 清理临时目录
    Remove-Item $tempDir -Recurse -Force
}

# 备份数据库
function Backup-Database {
    param([string]$BackupDir)
    
    Write-Info "备份数据库..."
    
    if ($env:DATABASE_URL -match "postgres") {
        # PostgreSQL
        if (Get-Command pg_dump -ErrorAction SilentlyContinue) {
            $dbBackup = Join-Path $BackupDir "database.sql"
            if ($DryRun) {
                Write-Info "[DRY RUN] 备份 PostgreSQL 数据库"
            } else {
                & pg_dump $env:DATABASE_URL | Out-File -FilePath $dbBackup -Encoding UTF8
                Write-Success "PostgreSQL 数据库备份完成"
            }
        } else {
            Write-Warning "pg_dump 未安装，跳过数据库备份"
        }
    } elseif ($env:DATABASE_URL -match "mysql") {
        # MySQL
        if (Get-Command mysqldump -ErrorAction SilentlyContinue) {
            $dbBackup = Join-Path $BackupDir "database.sql"
            if ($DryRun) {
                Write-Info "[DRY RUN] 备份 MySQL 数据库"
            } else {
                & mysqldump --single-transaction --routines --triggers $env:DATABASE_URL | Out-File -FilePath $dbBackup -Encoding UTF8
                Write-Success "MySQL 数据库备份完成"
            }
        } else {
            Write-Warning "mysqldump 未安装，跳过数据库备份"
        }
    } else {
        Write-Warning "不支持的数据库类型，跳过数据库备份"
    }
}

# 备份 Redis
function Backup-Redis {
    param([string]$BackupDir)
    
    Write-Info "备份 Redis..."
    
    if (Get-Command redis-cli -ErrorAction SilentlyContinue) {
        $redisBackup = Join-Path $BackupDir "redis.rdb"
        if ($DryRun) {
            Write-Info "[DRY RUN] 备份 Redis 数据"
        } else {
            & redis-cli -u $env:REDIS_URL --rdb $redisBackup
            Write-Success "Redis 数据备份完成"
        }
    } else {
        Write-Warning "redis-cli 未安装，跳过 Redis 备份"
    }
}

# 备份日志
function Backup-Logs {
    param(
        [string]$BackupDir,
        [string]$BackupName
    )
    
    Write-Info "备份日志..."
    
    $logDirs = @(
        "C:\Logs\monitoring",
        "C:\monitoring\logs",
        "$ProjectRoot\logs"
    )
    
    $tempDir = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_ }
    $logsBackupDir = Join-Path $tempDir "logs"
    New-Item -ItemType Directory -Path $logsBackupDir -Force | Out-Null
    
    # 复制日志目录
    foreach ($dir in $logDirs) {
        if (Test-Path $dir) {
            $destDir = Join-Path $logsBackupDir (Split-Path $dir -Leaf)
            Copy-Item -Path $dir -Destination $destDir -Recurse -Force
            Write-Info "复制日志目录: $dir"
        }
    }
    
    # 创建压缩包
    $ext = Get-CompressionExtension $Compression
    $archivePath = Join-Path $BackupDir "${BackupName}-logs${ext}"
    
    if ($DryRun) {
        Write-Info "[DRY RUN] 创建日志备份: $archivePath"
    } else {
        switch ($Compression) {
            "zip" {
                Compress-Archive -Path "$logsBackupDir\*" -DestinationPath $archivePath -Force
            }
            "gzip" {
                & 7z a -ttar -so "$archivePath.tar" "$logsBackupDir\*" | & 7z a -tgzip -si "$archivePath"
            }
            "none" {
                & 7z a -ttar "$archivePath" "$logsBackupDir\*"
            }
        }
        Write-Success "日志备份完成: $archivePath"
    }
    
    # 清理临时目录
    Remove-Item $tempDir -Recurse -Force
}

# 执行备份
function Start-Backup {
    Write-Info "开始备份..."
    
    # 创建备份目录
    New-BackupDirectory $BackupDir
    
    # 生成备份名称
    if (-not $Name) {
        $Name = New-BackupName $Type $Environment
    }
    
    Write-Info "备份名称: $Name"
    
    # 执行备份
    switch ($Type) {
        "config" {
            Backup-Config $BackupDir $Name
        }
        "data" {
            Backup-Data $BackupDir $Name
        }
        "logs" {
            Backup-Logs $BackupDir $Name
        }
        "all" {
            Backup-Config $BackupDir $Name
            Backup-Data $BackupDir $Name
            Backup-Logs $BackupDir $Name
        }
    }
    
    # 加密备份
    if ($Encrypt) {
        Protect-Backups $BackupDir $Name
    }
    
    Write-Success "备份完成"
}

# 加密备份
function Protect-Backups {
    param(
        [string]$BackupDir,
        [string]$BackupName
    )
    
    Write-Info "加密备份文件..."
    
    $ext = Get-CompressionExtension $Compression
    $files = Get-ChildItem -Path $BackupDir -Filter "${BackupName}*${ext}"
    
    foreach ($file in $files) {
        if ($DryRun) {
            Write-Info "[DRY RUN] 加密文件: $($file.FullName)"
        } else {
            if ($Password) {
                & gpg --batch --yes --passphrase $Password --symmetric --cipher-algo AES256 $file.FullName
                Remove-Item $file.FullName -Force
                Write-Success "文件已加密: $($file.FullName).gpg"
            } else {
                & gpg --symmetric --cipher-algo AES256 $file.FullName
                Remove-Item $file.FullName -Force
                Write-Success "文件已加密: $($file.FullName).gpg"
            }
        }
    }
}

# 列出备份
function Get-BackupList {
    Write-Info "列出备份文件..."
    
    if (-not (Test-Path $BackupDir)) {
        Write-Warning "备份目录不存在: $BackupDir"
        return
    }
    
    $pattern = "${ServiceName}-"
    if ($Environment) {
        $pattern = "${ServiceName}-${Environment}-"
    }
    
    Write-Host "备份目录: $BackupDir"
    Write-Host "========================================="
    
    Get-ChildItem -Path $BackupDir -Filter "${pattern}*" | Sort-Object Name | ForEach-Object {
        $size = [math]::Round($_.Length / 1MB, 2)
        Write-Host "$($_.Name) - ${size}MB - $($_.LastWriteTime)"
    }
}

# 清理旧备份
function Remove-OldBackups {
    Write-Info "清理旧备份..."
    
    if (-not (Test-Path $BackupDir)) {
        Write-Warning "备份目录不存在: $BackupDir"
        return
    }
    
    $pattern = "${ServiceName}-"
    if ($Environment) {
        $pattern = "${ServiceName}-${Environment}-"
    }
    
    $cutoffDate = (Get-Date).AddDays(-$Retention)
    $oldFiles = Get-ChildItem -Path $BackupDir -Filter "${pattern}*" | Where-Object { $_.LastWriteTime -lt $cutoffDate }
    
    $count = 0
    foreach ($file in $oldFiles) {
        if ($DryRun) {
            Write-Info "[DRY RUN] 删除旧备份: $($file.Name)"
        } else {
            Remove-Item $file.FullName -Force
            Write-Info "删除旧备份: $($file.Name)"
        }
        $count++
    }
    
    if ($count -eq 0) {
        Write-Info "没有需要清理的旧备份"
    } else {
        Write-Success "清理了 $count 个旧备份文件"
    }
}

# 验证备份
function Test-BackupIntegrity {
    Write-Info "验证备份..."
    
    if (-not $Name) {
        Write-Error "请指定要验证的备份名称"
        exit 1
    }
    
    $ext = Get-CompressionExtension $Compression
    $files = Get-ChildItem -Path $BackupDir -Filter "${Name}*${ext}"
    
    $verified = 0
    $failed = 0
    
    foreach ($file in $files) {
        Write-Info "验证文件: $($file.Name)"
        
        $isValid = $false
        switch ($Compression) {
            "zip" {
                try {
                    Add-Type -AssemblyName System.IO.Compression.FileSystem
                    [System.IO.Compression.ZipFile]::OpenRead($file.FullName).Dispose()
                    $isValid = $true
                } catch {
                    $isValid = $false
                }
            }
            "gzip" {
                $result = & 7z t $file.FullName 2>&1
                $isValid = $LASTEXITCODE -eq 0
            }
            "none" {
                $result = & 7z t $file.FullName 2>&1
                $isValid = $LASTEXITCODE -eq 0
            }
        }
        
        if ($isValid) {
            Write-Success "文件完整: $($file.Name)"
            $verified++
        } else {
            Write-Error "文件损坏: $($file.Name)"
            $failed++
        }
    }
    
    Write-Info "验证完成: $verified 个文件完整, $failed 个文件损坏"
    
    if ($failed -gt 0) {
        exit 1
    }
}

# 恢复备份
function Restore-Backup {
    Write-Info "恢复备份..."
    
    if (-not $Name) {
        Write-Error "请指定要恢复的备份名称"
        exit 1
    }
    
    if (-not $Force) {
        $confirm = Read-Host "确定要恢复备份 '$Name' 吗? (y/N)"
        if ($confirm -ne "y" -and $confirm -ne "Y") {
            Write-Info "取消恢复操作"
            exit 0
        }
    }
    
    $ext = Get-CompressionExtension $Compression
    
    # 恢复配置
    $configFile = Join-Path $BackupDir "${Name}-config${ext}"
    if (Test-Path $configFile) {
        Write-Info "恢复配置文件..."
        if ($DryRun) {
            Write-Info "[DRY RUN] 恢复配置: $configFile"
        } else {
            switch ($Compression) {
                "zip" {
                    Expand-Archive -Path $configFile -DestinationPath $ProjectRoot -Force
                }
                "gzip" {
                    & 7z x $configFile -o"$ProjectRoot" -y
                }
                "none" {
                    & 7z x $configFile -o"$ProjectRoot" -y
                }
            }
            Write-Success "配置恢复完成"
        }
    }
    
    # 恢复数据
    $dataFile = Join-Path $BackupDir "${Name}-data${ext}"
    if (Test-Path $dataFile) {
        Write-Info "恢复数据文件..."
        if ($DryRun) {
            Write-Info "[DRY RUN] 恢复数据: $dataFile"
        } else {
            switch ($Compression) {
                "zip" {
                    Expand-Archive -Path $dataFile -DestinationPath "C:\" -Force
                }
                "gzip" {
                    & 7z x $dataFile -o"C:\" -y
                }
                "none" {
                    & 7z x $dataFile -o"C:\" -y
                }
            }
            Write-Success "数据恢复完成"
        }
    }
    
    # 恢复日志
    $logsFile = Join-Path $BackupDir "${Name}-logs${ext}"
    if (Test-Path $logsFile) {
        Write-Info "恢复日志文件..."
        if ($DryRun) {
            Write-Info "[DRY RUN] 恢复日志: $logsFile"
        } else {
            switch ($Compression) {
                "zip" {
                    Expand-Archive -Path $logsFile -DestinationPath "C:\" -Force
                }
                "gzip" {
                    & 7z x $logsFile -o"C:\" -y
                }
                "none" {
                    & 7z x $logsFile -o"C:\" -y
                }
            }
            Write-Success "日志恢复完成"
        }
    }
    
    Write-Success "备份恢复完成"
}

# 主函数
function Main {
    # 如果启用详细输出
    if ($Verbose) {
        $VerbosePreference = "Continue"
    }
    
    # 检查依赖
    Test-Dependencies
    
    # 执行命令
    switch ($Command) {
        "backup" {
            Start-Backup
        }
        "restore" {
            Restore-Backup
        }
        "list" {
            Get-BackupList
        }
        "cleanup" {
            Remove-OldBackups
        }
        "verify" {
            Test-BackupIntegrity
        }
        "help" {
            Show-Help
        }
        default {
            Write-Error "未知命令: $Command"
            Show-Help
            exit 1
        }
    }
}

# 执行主函数
Main