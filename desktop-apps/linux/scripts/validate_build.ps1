# TaishangLaojun Linux Build Validation Script (PowerShell)
# This script validates the build environment and configuration for Linux desktop application

param(
    [switch]$Verbose = $false
)

# Color functions for output
function Write-Success {
    param([string]$Message)
    Write-Host "✓ $Message" -ForegroundColor Green
}

function Write-Error {
    param([string]$Message)
    Write-Host "✗ $Message" -ForegroundColor Red
}

function Write-Warning {
    param([string]$Message)
    Write-Host "⚠ $Message" -ForegroundColor Yellow
}

function Write-Info {
    param([string]$Message)
    Write-Host "ℹ $Message" -ForegroundColor Cyan
}

function Write-Header {
    param([string]$Message)
    Write-Host "`n=== $Message ===" -ForegroundColor Magenta
}

# Initialize counters
$script:ErrorCount = 0
$script:WarningCount = 0

function Test-FileExists {
    param(
        [string]$Path,
        [string]$Description
    )
    
    if (Test-Path $Path) {
        Write-Success "$Description exists: $Path"
        return $true
    } else {
        Write-Error "$Description missing: $Path"
        $script:ErrorCount++
        return $false
    }
}

function Test-DirectoryExists {
    param(
        [string]$Path,
        [string]$Description
    )
    
    if (Test-Path $Path -PathType Container) {
        Write-Success "$Description directory exists: $Path"
        return $true
    } else {
        Write-Error "$Description directory missing: $Path"
        $script:ErrorCount++
        return $false
    }
}

# Main validation function
function Start-Validation {
    Write-Header "TaishangLaojun Linux Build Validation"
    
    # Get project root
    $ProjectRoot = Split-Path -Parent $PSScriptRoot
    Write-Info "Project root: $ProjectRoot"
    
    # Check project structure
    Write-Header "Checking Project Structure"
    
    $RequiredDirs = @(
        @{Path = "$ProjectRoot\src"; Description = "Source code"},
        @{Path = "$ProjectRoot\include"; Description = "Header files"},
        @{Path = "$ProjectRoot\tests"; Description = "Test files"},
        @{Path = "$ProjectRoot\docs"; Description = "Documentation"},
        @{Path = "$ProjectRoot\resources"; Description = "Application resources"},
        @{Path = "$ProjectRoot\packaging"; Description = "Packaging files"},
        @{Path = "$ProjectRoot\scripts"; Description = "Build scripts"}
    )
    
    foreach ($dir in $RequiredDirs) {
        Test-DirectoryExists -Path $dir.Path -Description $dir.Description
    }
    
    # Check main build files
    Write-Header "Checking Build Configuration"
    
    $BuildFiles = @(
        @{Path = "$ProjectRoot\CMakeLists.txt"; Description = "Main CMake configuration"},
        @{Path = "$ProjectRoot\build.sh"; Description = "Build script"},
        @{Path = "$ProjectRoot\src\CMakeLists.txt"; Description = "Source CMake configuration"},
        @{Path = "$ProjectRoot\tests\CMakeLists.txt"; Description = "Tests CMake configuration"}
    )
    
    foreach ($file in $BuildFiles) {
        Test-FileExists -Path $file.Path -Description $file.Description
    }
    
    # Check source files
    Write-Header "Checking Source Files"
    
    $SourceFiles = @(
        @{Path = "$ProjectRoot\src\main.c"; Description = "Main source file"},
        @{Path = "$ProjectRoot\src\app.c"; Description = "Application source"},
        @{Path = "$ProjectRoot\src\ui.c"; Description = "UI source"},
        @{Path = "$ProjectRoot\src\config.c"; Description = "Configuration source"},
        @{Path = "$ProjectRoot\src\utils.c"; Description = "Utilities source"}
    )
    
    foreach ($file in $SourceFiles) {
        Test-FileExists -Path $file.Path -Description $file.Description
    }
    
    # Check header files
    Write-Header "Checking Header Files"
    
    $HeaderFiles = @(
        @{Path = "$ProjectRoot\include\app.h"; Description = "Application header"},
        @{Path = "$ProjectRoot\include\ui.h"; Description = "UI header"},
        @{Path = "$ProjectRoot\include\config.h"; Description = "Configuration header"},
        @{Path = "$ProjectRoot\include\utils.h"; Description = "Utilities header"}
    )
    
    foreach ($file in $HeaderFiles) {
        Test-FileExists -Path $file.Path -Description $file.Description
    }
    
    # Check resource files
    Write-Header "Checking Resource Files"
    
    $ResourceFiles = @(
        @{Path = "$ProjectRoot\resources\taishang-laojun.desktop"; Description = "Desktop entry file"},
        @{Path = "$ProjectRoot\resources\icons\taishang-laojun.svg"; Description = "SVG icon"},
        @{Path = "$ProjectRoot\resources\appstream\taishang-laojun.appdata.xml"; Description = "AppStream metadata"},
        @{Path = "$ProjectRoot\resources\man\taishang-laojun.1"; Description = "Manual page"}
    )
    
    foreach ($file in $ResourceFiles) {
        Test-FileExists -Path $file.Path -Description $file.Description
    }
    
    # Check packaging files
    Write-Header "Checking Packaging Configuration"
    
    # DEB packaging
    $DebFiles = @(
        @{Path = "$ProjectRoot\packaging\debian\control"; Description = "DEB control file"},
        @{Path = "$ProjectRoot\packaging\debian\changelog"; Description = "DEB changelog"},
        @{Path = "$ProjectRoot\packaging\debian\rules"; Description = "DEB rules"},
        @{Path = "$ProjectRoot\packaging\debian\postinst"; Description = "DEB post-install script"},
        @{Path = "$ProjectRoot\packaging\debian\prerm"; Description = "DEB pre-remove script"},
        @{Path = "$ProjectRoot\packaging\debian\postrm"; Description = "DEB post-remove script"}
    )
    
    foreach ($file in $DebFiles) {
        Test-FileExists -Path $file.Path -Description $file.Description
    }
    
    # RPM packaging
    $RpmFiles = @(
        @{Path = "$ProjectRoot\packaging\rpm\taishang-laojun.spec"; Description = "RPM spec file"},
        @{Path = "$ProjectRoot\packaging\rpm\postinst.sh"; Description = "RPM post-install script"},
        @{Path = "$ProjectRoot\packaging\rpm\prerm.sh"; Description = "RPM pre-remove script"},
        @{Path = "$ProjectRoot\packaging\rpm\postrm.sh"; Description = "RPM post-remove script"}
    )
    
    foreach ($file in $RpmFiles) {
        Test-FileExists -Path $file.Path -Description $file.Description
    }
    
    # Flatpak packaging
    Test-FileExists -Path "$ProjectRoot\packaging\flatpak\taishang-laojun.yml" -Description "Flatpak manifest"
    
    # Snap packaging
    Test-FileExists -Path "$ProjectRoot\packaging\snap\snapcraft.yaml" -Description "Snap configuration"
    
    # AppImage packaging
    Test-FileExists -Path "$ProjectRoot\packaging\scripts\build_appimage.sh" -Description "AppImage build script"
    
    # Check test files
    Write-Header "Checking Test Files"
    
    $TestFiles = @(
        @{Path = "$ProjectRoot\tests\test_main.c"; Description = "Main test file"},
        @{Path = "$ProjectRoot\tests\test_app.c"; Description = "Application tests"},
        @{Path = "$ProjectRoot\tests\test_ui.c"; Description = "UI tests"},
        @{Path = "$ProjectRoot\tests\test_config.c"; Description = "Configuration tests"},
        @{Path = "$ProjectRoot\tests\test_utils.c"; Description = "Utilities tests"}
    )
    
    foreach ($file in $TestFiles) {
        Test-FileExists -Path $file.Path -Description $file.Description
    }
    
    # Check documentation
    Write-Header "Checking Documentation"
    
    $DocFiles = @(
        @{Path = "$ProjectRoot\README.md"; Description = "Main README"},
        @{Path = "$ProjectRoot\docs\BUILD.md"; Description = "Build documentation"}
    )
    
    foreach ($file in $DocFiles) {
        Test-FileExists -Path $file.Path -Description $file.Description
    }
    
    # Check CI/CD configuration
    Write-Header "Checking CI/CD Configuration"
    
    Test-FileExists -Path "$ProjectRoot\.github\workflows\linux-build.yml" -Description "GitHub Actions workflow"
    
    # Validate file permissions (simulate for Windows)
    Write-Header "Checking Script Permissions"
    
    $ScriptFiles = @(
        "$ProjectRoot\build.sh",
        "$ProjectRoot\packaging\scripts\build_appimage.sh",
        "$ProjectRoot\packaging\scripts\verify_packages.sh",
        "$ProjectRoot\packaging\debian\rules",
        "$ProjectRoot\packaging\rpm\postinst.sh",
        "$ProjectRoot\packaging\rpm\prerm.sh",
        "$ProjectRoot\packaging\rpm\postrm.sh"
    )
    
    foreach ($script in $ScriptFiles) {
        if (Test-Path $script) {
            Write-Success "Script file exists: $script"
        } else {
            Write-Warning "Script file missing: $script"
            $script:WarningCount++
        }
    }
    
    # Summary
    Write-Header "Validation Summary"
    
    if ($script:ErrorCount -eq 0) {
        Write-Success "All critical files and directories are present!"
        Write-Info "The project structure is ready for Linux build and packaging."
        
        Write-Host "`n📋 Next Steps:" -ForegroundColor Yellow
        Write-Host "1. Install build dependencies on Linux system"
        Write-Host "2. Run: cmake -B build -G Ninja -DCMAKE_BUILD_TYPE=Release"
        Write-Host "3. Run: cmake --build build"
        Write-Host "4. Run: ctest --test-dir build"
        Write-Host "5. Create packages: ninja -C build package-all"
        
        return $true
    } else {
        Write-Error "Found $script:ErrorCount critical issues that need to be resolved."
        if ($script:WarningCount -gt 0) {
            Write-Warning "Found $script:WarningCount warnings."
        }
        return $false
    }
}

# Run validation
try {
    $ValidationResult = Start-Validation
    
    if ($ValidationResult) {
        Write-Host "`n🎉 Validation completed successfully!" -ForegroundColor Green
        exit 0
    } else {
        Write-Host "`n❌ Validation failed. Please fix the issues above." -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Error "Validation script failed: $($_.Exception.Message)"
    exit 1
}