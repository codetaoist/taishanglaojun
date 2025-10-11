#!/bin/bash

# Build validation script for TaishangLaojun Desktop Linux
# This script validates the build environment and configuration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
check_directory() {
    log_info "Checking project directory structure..."
    
    if [ ! -f "CMakeLists.txt" ]; then
        log_error "CMakeLists.txt not found. Are you in the project root?"
        exit 1
    fi
    
    if [ ! -d "src" ]; then
        log_error "src directory not found"
        exit 1
    fi
    
    if [ ! -d "tests" ]; then
        log_error "tests directory not found"
        exit 1
    fi
    
    if [ ! -d "packaging" ]; then
        log_error "packaging directory not found"
        exit 1
    fi
    
    if [ ! -d "resources" ]; then
        log_error "resources directory not found"
        exit 1
    fi
    
    log_success "Project directory structure is valid"
}

# Check build dependencies
check_dependencies() {
    log_info "Checking build dependencies..."
    
    local missing_deps=()
    
    # Check for required tools
    for tool in cmake ninja-build pkg-config gcc; do
        if ! command -v $tool >/dev/null 2>&1; then
            missing_deps+=($tool)
        fi
    done
    
    # Check for required libraries using pkg-config
    for lib in gtk4 libadwaita-1 glib-2.0 json-c openssl sqlite3; do
        if ! pkg-config --exists $lib 2>/dev/null; then
            missing_deps+=("lib${lib}-dev")
        fi
    done
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        log_info "Install them with:"
        log_info "  Ubuntu/Debian: sudo apt install ${missing_deps[*]}"
        log_info "  Fedora: sudo dnf install ${missing_deps[*]}"
        exit 1
    fi
    
    log_success "All build dependencies are available"
}

# Validate CMake configuration
validate_cmake() {
    log_info "Validating CMake configuration..."
    
    # Create a temporary build directory
    local temp_build_dir=$(mktemp -d)
    
    # Try to configure the project
    if cmake -S . -B "$temp_build_dir" -G Ninja -DCMAKE_BUILD_TYPE=Release >/dev/null 2>&1; then
        log_success "CMake configuration is valid"
    else
        log_error "CMake configuration failed"
        rm -rf "$temp_build_dir"
        exit 1
    fi
    
    # Clean up
    rm -rf "$temp_build_dir"
}

# Check source files
check_source_files() {
    log_info "Checking source files..."
    
    local required_files=(
        "src/main.c"
        "src/application.c"
        "include/application.h"
    )
    
    for file in "${required_files[@]}"; do
        if [ ! -f "$file" ]; then
            log_error "Required file missing: $file"
            exit 1
        fi
    done
    
    log_success "All required source files are present"
}

# Check resource files
check_resources() {
    log_info "Checking resource files..."
    
    local required_resources=(
        "resources/taishang-laojun.desktop"
        "resources/icons/taishang-laojun.svg"
        "resources/appstream/taishang-laojun.appdata.xml"
        "resources/man/taishang-laojun.1"
    )
    
    for resource in "${required_resources[@]}"; do
        if [ ! -f "$resource" ]; then
            log_error "Required resource missing: $resource"
            exit 1
        fi
    done
    
    log_success "All required resource files are present"
}

# Check packaging files
check_packaging() {
    log_info "Checking packaging files..."
    
    local required_packaging=(
        "packaging/CMakeLists.txt"
        "packaging/debian/postinst"
        "packaging/debian/prerm"
        "packaging/debian/postrm"
        "packaging/rpm/postinst.sh"
        "packaging/rpm/prerm.sh"
        "packaging/rpm/postrm.sh"
        "packaging/flatpak/taishang-laojun.yml"
        "packaging/snap/snapcraft.yaml"
        "packaging/scripts/copy_dependencies.sh"
        "packaging/scripts/verify_packages.sh"
    )
    
    for package_file in "${required_packaging[@]}"; do
        if [ ! -f "$package_file" ]; then
            log_error "Required packaging file missing: $package_file"
            exit 1
        fi
    done
    
    log_success "All required packaging files are present"
}

# Check test files
check_tests() {
    log_info "Checking test files..."
    
    local required_tests=(
        "tests/CMakeLists.txt"
        "tests/test_main.c"
        "tests/test_network.h"
        "tests/test_storage.h"
        "tests/test_system.h"
        "tests/test_graphics.h"
        "tests/test_audio.h"
        "tests/test_ui.h"
    )
    
    for test_file in "${required_tests[@]}"; do
        if [ ! -f "$test_file" ]; then
            log_error "Required test file missing: $test_file"
            exit 1
        fi
    done
    
    log_success "All required test files are present"
}

# Validate desktop file
validate_desktop_file() {
    log_info "Validating desktop file..."
    
    if command -v desktop-file-validate >/dev/null 2>&1; then
        if desktop-file-validate resources/taishang-laojun.desktop 2>/dev/null; then
            log_success "Desktop file is valid"
        else
            log_warning "Desktop file validation failed (non-critical)"
        fi
    else
        log_warning "desktop-file-validate not available, skipping validation"
    fi
}

# Validate AppStream metadata
validate_appstream() {
    log_info "Validating AppStream metadata..."
    
    if command -v appstream-util >/dev/null 2>&1; then
        if appstream-util validate resources/appstream/taishang-laojun.appdata.xml 2>/dev/null; then
            log_success "AppStream metadata is valid"
        else
            log_warning "AppStream metadata validation failed (non-critical)"
        fi
    else
        log_warning "appstream-util not available, skipping validation"
    fi
}

# Check file permissions
check_permissions() {
    log_info "Checking file permissions..."
    
    # Check that scripts are executable
    local scripts=(
        "packaging/scripts/copy_dependencies.sh"
        "packaging/scripts/verify_packages.sh"
        "packaging/debian/postinst"
        "packaging/debian/prerm"
        "packaging/debian/postrm"
        "packaging/rpm/postinst.sh"
        "packaging/rpm/prerm.sh"
        "packaging/rpm/postrm.sh"
    )
    
    for script in "${scripts[@]}"; do
        if [ ! -x "$script" ]; then
            log_warning "Script $script is not executable, fixing..."
            chmod +x "$script" 2>/dev/null || log_warning "Could not make $script executable"
        fi
    done
    
    log_success "File permissions checked"
}

# Main validation function
main() {
    log_info "Starting TaishangLaojun Desktop build validation..."
    echo
    
    check_directory
    check_dependencies
    validate_cmake
    check_source_files
    check_resources
    check_packaging
    check_tests
    validate_desktop_file
    validate_appstream
    check_permissions
    
    echo
    log_success "All validation checks passed!"
    log_info "The project is ready for building and packaging."
    echo
    log_info "To build the project:"
    log_info "  mkdir build && cd build"
    log_info "  cmake .. -G Ninja -DCMAKE_BUILD_TYPE=Release"
    log_info "  ninja"
    echo
    log_info "To create packages:"
    log_info "  ninja appimage      # Create AppImage"
    log_info "  ninja package-deb   # Create DEB package"
    log_info "  ninja package-rpm   # Create RPM package"
    log_info "  ninja flatpak       # Create Flatpak"
    log_info "  ninja snap          # Create Snap"
    log_info "  ninja all-packages  # Create all packages"
}

# Run main function
main "$@"