#!/bin/bash

# TaishangLaojun Desktop Linux Build Script

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目信息
PROJECT_NAME="taishanglaojun-desktop"
BUILD_DIR="build"
INSTALL_PREFIX="/usr/local"

# 函数定义
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    print_info "检查构建依赖..."
    
    local deps=("cmake" "gcc" "pkg-config")
    local missing_deps=()
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing_deps+=("$dep")
        fi
    done
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_error "缺少以下依赖: ${missing_deps[*]}"
        print_info "请安装缺少的依赖后重试"
        exit 1
    fi
    
    # 检查GTK4和libadwaita
    if ! pkg-config --exists gtk4; then
        print_error "GTK4 未找到，请安装 libgtk-4-dev"
        exit 1
    fi
    
    if ! pkg-config --exists libadwaita-1; then
        print_warning "libadwaita-1 未找到，将尝试继续构建"
    fi
    
    print_success "所有依赖检查完成"
}

# 清理构建目录
clean_build() {
    print_info "清理构建目录..."
    if [ -d "$BUILD_DIR" ]; then
        rm -rf "$BUILD_DIR"
    fi
    mkdir -p "$BUILD_DIR"
    print_success "构建目录已清理"
}

# 配置构建
configure_build() {
    print_info "配置构建..."
    cd "$BUILD_DIR"
    
    cmake .. \
        -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_INSTALL_PREFIX="$INSTALL_PREFIX" \
        -DCMAKE_EXPORT_COMPILE_COMMANDS=ON
    
    cd ..
    print_success "构建配置完成"
}

# 编译项目
build_project() {
    print_info "开始编译..."
    cd "$BUILD_DIR"
    
    make -j$(nproc)
    
    cd ..
    print_success "编译完成"
}

# 运行测试
run_tests() {
    print_info "运行测试..."
    cd "$BUILD_DIR"
    
    if [ -f "test_runner" ]; then
        ./test_runner
        print_success "测试通过"
    else
        print_warning "未找到测试程序，跳过测试"
    fi
    
    cd ..
}

# 安装应用
install_app() {
    print_info "安装应用程序..."
    cd "$BUILD_DIR"
    
    sudo make install
    
    cd ..
    print_success "安装完成"
}

# 创建AppImage
create_appimage() {
    print_info "创建AppImage..."
    cd "$BUILD_DIR"
    
    if command -v linuxdeploy &> /dev/null; then
        make appimage
        print_success "AppImage创建完成"
    else
        print_warning "linuxdeploy 未找到，跳过AppImage创建"
    fi
    
    cd ..
}

# 创建包
create_packages() {
    print_info "创建安装包..."
    cd "$BUILD_DIR"
    
    # 创建DEB包
    if command -v dpkg-deb &> /dev/null; then
        make package
        print_success "DEB包创建完成"
    fi
    
    # 创建RPM包
    if command -v rpmbuild &> /dev/null; then
        cpack -G RPM
        print_success "RPM包创建完成"
    fi
    
    cd ..
}

# 显示帮助信息
show_help() {
    echo "TaishangLaojun Desktop Linux 构建脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help      显示此帮助信息"
    echo "  -c, --clean     清理构建目录"
    echo "  -b, --build     编译项目"
    echo "  -t, --test      运行测试"
    echo "  -i, --install   安装应用程序"
    echo "  -a, --appimage  创建AppImage"
    echo "  -p, --package   创建安装包"
    echo "  --all           执行完整构建流程"
    echo ""
    echo "示例:"
    echo "  $0 --all                # 完整构建流程"
    echo "  $0 -c -b                # 清理并编译"
    echo "  $0 -b -t                # 编译并测试"
}

# 主函数
main() {
    local do_clean=false
    local do_build=false
    local do_test=false
    local do_install=false
    local do_appimage=false
    local do_package=false
    local do_all=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -c|--clean)
                do_clean=true
                shift
                ;;
            -b|--build)
                do_build=true
                shift
                ;;
            -t|--test)
                do_test=true
                shift
                ;;
            -i|--install)
                do_install=true
                shift
                ;;
            -a|--appimage)
                do_appimage=true
                shift
                ;;
            -p|--package)
                do_package=true
                shift
                ;;
            --all)
                do_all=true
                shift
                ;;
            *)
                print_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 如果没有指定任何选项，显示帮助
    if [ "$do_clean" = false ] && [ "$do_build" = false ] && [ "$do_test" = false ] && \
       [ "$do_install" = false ] && [ "$do_appimage" = false ] && [ "$do_package" = false ] && \
       [ "$do_all" = false ]; then
        show_help
        exit 0
    fi
    
    # 检查依赖
    check_dependencies
    
    # 执行操作
    if [ "$do_all" = true ]; then
        clean_build
        configure_build
        build_project
        run_tests
        create_appimage
        create_packages
    else
        if [ "$do_clean" = true ]; then
            clean_build
            configure_build
        fi
        
        if [ "$do_build" = true ]; then
            if [ ! -d "$BUILD_DIR" ] || [ ! -f "$BUILD_DIR/Makefile" ]; then
                clean_build
                configure_build
            fi
            build_project
        fi
        
        if [ "$do_test" = true ]; then
            run_tests
        fi
        
        if [ "$do_install" = true ]; then
            install_app
        fi
        
        if [ "$do_appimage" = true ]; then
            create_appimage
        fi
        
        if [ "$do_package" = true ]; then
            create_packages
        fi
    fi
    
    print_success "构建脚本执行完成！"
}

# 运行主函数
main "$@"