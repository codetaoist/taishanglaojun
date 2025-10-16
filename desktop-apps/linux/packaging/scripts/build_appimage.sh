#!/bin/bash
# TaishangLaojun AppImage Build Script
# This script builds an AppImage package for the TaishangLaojun desktop application

set -e

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
BUILD_DIR="$PROJECT_ROOT/build"
APPIMAGE_DIR="$BUILD_DIR/appimage"
APPDIR="$APPIMAGE_DIR/TaishangLaojun.AppDir"

# Application information
APP_NAME="TaishangLaojun"
APP_VERSION="${APP_VERSION:-1.0.0}"
APP_ARCH="${APP_ARCH:-x86_64}"

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

# Check if required tools are available
check_dependencies() {
    log_info "Checking dependencies..."
    
    local missing_deps=()
    
    # Check for linuxdeploy
    if ! command -v linuxdeploy &> /dev/null; then
        missing_deps+=("linuxdeploy")
    fi
    
    # Check for desktop-file-validate
    if ! command -v desktop-file-validate &> /dev/null; then
        log_warning "desktop-file-validate not found (optional)"
    fi
    
    # Check for appstream-util
    if ! command -v appstream-util &> /dev/null; then
        log_warning "appstream-util not found (optional)"
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Missing required dependencies: ${missing_deps[*]}"
        log_info "Please install missing dependencies:"
        for dep in "${missing_deps[@]}"; do
            case $dep in
                linuxdeploy)
                    echo "  wget https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage"
                    echo "  chmod +x linuxdeploy-x86_64.AppImage"
                    echo "  sudo mv linuxdeploy-x86_64.AppImage /usr/local/bin/linuxdeploy"
                    ;;
            esac
        done
        exit 1
    fi
    
    log_success "All required dependencies are available"
}

# Clean previous AppImage build
clean_appimage() {
    log_info "Cleaning previous AppImage build..."
    
    if [ -d "$APPIMAGE_DIR" ]; then
        rm -rf "$APPIMAGE_DIR"
    fi
    
    mkdir -p "$APPIMAGE_DIR"
    log_success "AppImage build directory cleaned"
}

# Create AppDir structure
create_appdir() {
    log_info "Creating AppDir structure..."
    
    mkdir -p "$APPDIR"
    mkdir -p "$APPDIR/usr/bin"
    mkdir -p "$APPDIR/usr/lib"
    mkdir -p "$APPDIR/usr/share/applications"
    mkdir -p "$APPDIR/usr/share/icons/hicolor/scalable/apps"
    mkdir -p "$APPDIR/usr/share/metainfo"
    mkdir -p "$APPDIR/usr/share/man/man1"
    
    log_success "AppDir structure created"
}

# Install application files
install_app_files() {
    log_info "Installing application files..."
    
    # Install binary
    if [ -f "$BUILD_DIR/src/taishang-laojun" ]; then
        cp "$BUILD_DIR/src/taishang-laojun" "$APPDIR/usr/bin/"
        chmod +x "$APPDIR/usr/bin/taishang-laojun"
        log_success "Binary installed"
    else
        log_error "Application binary not found: $BUILD_DIR/src/taishang-laojun"
        exit 1
    fi
    
    # Install desktop file
    if [ -f "$PROJECT_ROOT/resources/taishang-laojun.desktop" ]; then
        cp "$PROJECT_ROOT/resources/taishang-laojun.desktop" "$APPDIR/usr/share/applications/"
        
        # Create symlink for AppImage
        ln -sf "usr/share/applications/taishang-laojun.desktop" "$APPDIR/taishang-laojun.desktop"
        
        # Validate desktop file
        if command -v desktop-file-validate &> /dev/null; then
            if desktop-file-validate "$APPDIR/usr/share/applications/taishang-laojun.desktop"; then
                log_success "Desktop file installed and validated"
            else
                log_warning "Desktop file validation failed"
            fi
        else
            log_success "Desktop file installed"
        fi
    else
        log_error "Desktop file not found: $PROJECT_ROOT/resources/taishang-laojun.desktop"
        exit 1
    fi
    
    # Install icon
    if [ -f "$PROJECT_ROOT/resources/icons/taishang-laojun.svg" ]; then
        cp "$PROJECT_ROOT/resources/icons/taishang-laojun.svg" "$APPDIR/usr/share/icons/hicolor/scalable/apps/"
        
        # Create symlink for AppImage
        ln -sf "usr/share/icons/hicolor/scalable/apps/taishang-laojun.svg" "$APPDIR/taishang-laojun.svg"
        
        log_success "Icon installed"
    else
        log_warning "Icon file not found: $PROJECT_ROOT/resources/icons/taishang-laojun.svg"
    fi
    
    # Install AppStream metadata
    if [ -f "$PROJECT_ROOT/resources/appstream/taishang-laojun.appdata.xml" ]; then
        cp "$PROJECT_ROOT/resources/appstream/taishang-laojun.appdata.xml" "$APPDIR/usr/share/metainfo/"
        
        # Validate AppStream metadata
        if command -v appstream-util &> /dev/null; then
            if appstream-util validate "$APPDIR/usr/share/metainfo/taishang-laojun.appdata.xml"; then
                log_success "AppStream metadata installed and validated"
            else
                log_warning "AppStream metadata validation failed"
            fi
        else
            log_success "AppStream metadata installed"
        fi
    else
        log_warning "AppStream metadata not found: $PROJECT_ROOT/resources/appstream/taishang-laojun.appdata.xml"
    fi
    
    # Install manual page
    if [ -f "$PROJECT_ROOT/resources/man/taishang-laojun.1" ]; then
        cp "$PROJECT_ROOT/resources/man/taishang-laojun.1" "$APPDIR/usr/share/man/man1/"
        gzip "$APPDIR/usr/share/man/man1/taishang-laojun.1"
        log_success "Manual page installed"
    else
        log_warning "Manual page not found: $PROJECT_ROOT/resources/man/taishang-laojun.1"
    fi
}

# Create AppRun script
create_apprun() {
    log_info "Creating AppRun script..."
    
    cat > "$APPDIR/AppRun" << 'EOF'
#!/bin/bash
# AppRun script for TaishangLaojun

# Get the directory where this script is located
HERE="$(dirname "$(readlink -f "${0}")")"

# Set up environment
export PATH="${HERE}/usr/bin:${PATH}"
export LD_LIBRARY_PATH="${HERE}/usr/lib:${LD_LIBRARY_PATH}"
export XDG_DATA_DIRS="${HERE}/usr/share:${XDG_DATA_DIRS}"

# Set application-specific environment variables
export TAISHANG_APPIMAGE=1
export TAISHANG_APPDIR="${HERE}"

# Run the application
exec "${HERE}/usr/bin/taishang-laojun" "$@"
EOF
    
    chmod +x "$APPDIR/AppRun"
    log_success "AppRun script created"
}

# Bundle dependencies using linuxdeploy
bundle_dependencies() {
    log_info "Bundling dependencies with linuxdeploy..."
    
    cd "$APPIMAGE_DIR"
    
    # Run linuxdeploy
    linuxdeploy \
        --appdir "$APPDIR" \
        --executable "$APPDIR/usr/bin/taishang-laojun" \
        --desktop-file "$APPDIR/usr/share/applications/taishang-laojun.desktop" \
        --icon-file "$APPDIR/usr/share/icons/hicolor/scalable/apps/taishang-laojun.svg" \
        --output appimage
    
    log_success "Dependencies bundled successfully"
}

# Create AppImage with custom settings
create_appimage() {
    log_info "Creating AppImage..."
    
    cd "$APPIMAGE_DIR"
    
    # Set AppImage environment variables
    export ARCH="$APP_ARCH"
    export VERSION="$APP_VERSION"
    
    # Create AppImage using linuxdeploy
    if linuxdeploy \
        --appdir "$APPDIR" \
        --executable "$APPDIR/usr/bin/taishang-laojun" \
        --desktop-file "$APPDIR/usr/share/applications/taishang-laojun.desktop" \
        --icon-file "$APPDIR/usr/share/icons/hicolor/scalable/apps/taishang-laojun.svg" \
        --output appimage; then
        
        # Find the generated AppImage
        APPIMAGE_FILE=$(find . -name "*.AppImage" -type f | head -1)
        
        if [ -n "$APPIMAGE_FILE" ]; then
            # Rename to standard format
            FINAL_NAME="TaishangLaojun-${APP_VERSION}-${APP_ARCH}.AppImage"
            mv "$APPIMAGE_FILE" "$FINAL_NAME"
            
            # Make executable
            chmod +x "$FINAL_NAME"
            
            # Get file size
            FILE_SIZE=$(du -h "$FINAL_NAME" | cut -f1)
            
            log_success "AppImage created: $FINAL_NAME ($FILE_SIZE)"
            
            # Copy to packaging directory
            mkdir -p "$BUILD_DIR/packaging"
            cp "$FINAL_NAME" "$BUILD_DIR/packaging/"
            
            log_success "AppImage copied to: $BUILD_DIR/packaging/$FINAL_NAME"
        else
            log_error "AppImage file not found after creation"
            exit 1
        fi
    else
        log_error "Failed to create AppImage"
        exit 1
    fi
}

# Verify AppImage
verify_appimage() {
    log_info "Verifying AppImage..."
    
    APPIMAGE_FILE="$BUILD_DIR/packaging/TaishangLaojun-${APP_VERSION}-${APP_ARCH}.AppImage"
    
    if [ -f "$APPIMAGE_FILE" ]; then
        # Check if AppImage is executable
        if [ -x "$APPIMAGE_FILE" ]; then
            log_success "AppImage is executable"
        else
            log_error "AppImage is not executable"
            exit 1
        fi
        
        # Test AppImage help
        if "$APPIMAGE_FILE" --help &> /dev/null; then
            log_success "AppImage help command works"
        else
            log_warning "AppImage help command failed (may be normal)"
        fi
        
        # Check AppImage structure
        if "$APPIMAGE_FILE" --appimage-extract-and-run echo "test" &> /dev/null; then
            log_success "AppImage structure is valid"
        else
            log_warning "AppImage structure test failed"
        fi
        
        log_success "AppImage verification completed"
    else
        log_error "AppImage file not found: $APPIMAGE_FILE"
        exit 1
    fi
}

# Print usage information
print_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help          Show this help message"
    echo "  -c, --clean         Clean build directory before building"
    echo "  -v, --version VER   Set application version (default: 1.0.0)"
    echo "  -a, --arch ARCH     Set target architecture (default: x86_64)"
    echo ""
    echo "Environment variables:"
    echo "  APP_VERSION         Application version"
    echo "  APP_ARCH           Target architecture"
    echo ""
    echo "Example:"
    echo "  $0 --clean --version 1.2.3"
}

# Main function
main() {
    local clean_build=false
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                print_usage
                exit 0
                ;;
            -c|--clean)
                clean_build=true
                shift
                ;;
            -v|--version)
                APP_VERSION="$2"
                shift 2
                ;;
            -a|--arch)
                APP_ARCH="$2"
                shift 2
                ;;
            *)
                log_error "Unknown option: $1"
                print_usage
                exit 1
                ;;
        esac
    done
    
    log_info "Starting AppImage build for TaishangLaojun v$APP_VERSION ($APP_ARCH)"
    
    # Check if build directory exists
    if [ ! -d "$BUILD_DIR" ]; then
        log_error "Build directory not found: $BUILD_DIR"
        log_info "Please run 'cmake --build build' first"
        exit 1
    fi
    
    # Execute build steps
    check_dependencies
    
    if [ "$clean_build" = true ]; then
        clean_appimage
    fi
    
    create_appdir
    install_app_files
    create_apprun
    create_appimage
    verify_appimage
    
    log_success "AppImage build completed successfully!"
    log_info "AppImage location: $BUILD_DIR/packaging/TaishangLaojun-${APP_VERSION}-${APP_ARCH}.AppImage"
}

# Run main function with all arguments
main "$@"