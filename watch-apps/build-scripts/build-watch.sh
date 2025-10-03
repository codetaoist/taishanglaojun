#!/bin/bash

# 太上老君手表应用构建脚本
# Build script for Taishanglaojun Watch Applications

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
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

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        log_error "$1 命令未找到，请确保已安装相关工具"
        exit 1
    fi
}

# 构建配置
BUILD_TYPE=${1:-debug}
PLATFORM=${2:-all}
PROJECT_ROOT=$(dirname $(dirname $(realpath $0)))
WATCH_APPS_DIR="$PROJECT_ROOT/watch-apps"
APPLE_WATCH_DIR="$WATCH_APPS_DIR/apple-watch"
WEAR_OS_DIR="$WATCH_APPS_DIR/wear-os"
BUILD_OUTPUT_DIR="$PROJECT_ROOT/build/watch-apps"

log_info "开始构建太上老君手表应用..."
log_info "构建类型: $BUILD_TYPE"
log_info "目标平台: $PLATFORM"
log_info "项目根目录: $PROJECT_ROOT"

# 创建构建输出目录
mkdir -p "$BUILD_OUTPUT_DIR"

# 构建 Wear OS 应用
build_wear_os() {
    log_info "开始构建 Wear OS 应用..."
    
    if [ ! -d "$WEAR_OS_DIR" ]; then
        log_error "Wear OS 项目目录不存在: $WEAR_OS_DIR"
        return 1
    fi
    
    cd "$WEAR_OS_DIR"
    
    # 检查 Gradle 包装器
    if [ ! -f "./gradlew" ]; then
        log_error "Gradle 包装器未找到，请确保项目已正确初始化"
        return 1
    fi
    
    # 清理项目
    log_info "清理 Wear OS 项目..."
    ./gradlew clean
    
    # 构建项目
    if [ "$BUILD_TYPE" = "release" ]; then
        log_info "构建 Wear OS Release 版本..."
        ./gradlew assembleRelease
        
        # 复制 APK 文件
        if [ -f "app/build/outputs/apk/release/app-release.apk" ]; then
            cp "app/build/outputs/apk/release/app-release.apk" "$BUILD_OUTPUT_DIR/taishanglaojun-watch-release.apk"
            log_success "Wear OS Release APK 已生成: $BUILD_OUTPUT_DIR/taishanglaojun-watch-release.apk"
        fi
    else
        log_info "构建 Wear OS Debug 版本..."
        ./gradlew assembleDebug
        
        # 复制 APK 文件
        if [ -f "app/build/outputs/apk/debug/app-debug.apk" ]; then
            cp "app/build/outputs/apk/debug/app-debug.apk" "$BUILD_OUTPUT_DIR/taishanglaojun-watch-debug.apk"
            log_success "Wear OS Debug APK 已生成: $BUILD_OUTPUT_DIR/taishanglaojun-watch-debug.apk"
        fi
    fi
    
    # 运行测试
    log_info "运行 Wear OS 单元测试..."
    ./gradlew testDebugUnitTest
    
    log_success "Wear OS 应用构建完成"
}

# 构建 Apple Watch 应用
build_apple_watch() {
    log_info "开始构建 Apple Watch 应用..."
    
    if [ ! -d "$APPLE_WATCH_DIR" ]; then
        log_error "Apple Watch 项目目录不存在: $APPLE_WATCH_DIR"
        return 1
    fi
    
    # 检查 Xcode 命令行工具
    check_command xcodebuild
    
    cd "$APPLE_WATCH_DIR"
    
    # 检查项目文件
    if [ ! -f "TaishanglaojunWatch.xcodeproj/project.pbxproj" ]; then
        log_error "Xcode 项目文件未找到"
        return 1
    fi
    
    # 清理项目
    log_info "清理 Apple Watch 项目..."
    xcodebuild clean -project TaishanglaojunWatch.xcodeproj -scheme TaishanglaojunWatch
    
    # 构建项目
    if [ "$BUILD_TYPE" = "release" ]; then
        log_info "构建 Apple Watch Release 版本..."
        xcodebuild -project TaishanglaojunWatch.xcodeproj \
                   -scheme TaishanglaojunWatch \
                   -configuration Release \
                   -destination 'generic/platform=watchOS' \
                   build
    else
        log_info "构建 Apple Watch Debug 版本..."
        xcodebuild -project TaishanglaojunWatch.xcodeproj \
                   -scheme TaishanglaojunWatch \
                   -configuration Debug \
                   -destination 'generic/platform=watchOS' \
                   build
    fi
    
    # 运行测试
    log_info "运行 Apple Watch 测试..."
    xcodebuild test -project TaishanglaojunWatch.xcodeproj \
                   -scheme TaishanglaojunWatch \
                   -destination 'platform=watchOS Simulator,name=Apple Watch Series 9 (45mm)'
    
    log_success "Apple Watch 应用构建完成"
}

# 执行构建
case $PLATFORM in
    "wear-os"|"wearos")
        build_wear_os
        ;;
    "apple-watch"|"applewatch"|"ios")
        build_apple_watch
        ;;
    "all"|*)
        if [[ "$OSTYPE" == "darwin"* ]]; then
            build_apple_watch
        else
            log_warning "在非 macOS 系统上跳过 Apple Watch 构建"
        fi
        build_wear_os
        ;;
esac

# 生成构建报告
log_info "生成构建报告..."
BUILD_REPORT="$BUILD_OUTPUT_DIR/build-report.txt"
cat > "$BUILD_REPORT" << EOF
太上老君手表应用构建报告
========================

构建时间: $(date)
构建类型: $BUILD_TYPE
目标平台: $PLATFORM
项目版本: 1.0.0

构建输出:
EOF

if [ -f "$BUILD_OUTPUT_DIR/taishanglaojun-watch-debug.apk" ]; then
    echo "- Wear OS Debug APK: $(ls -lh $BUILD_OUTPUT_DIR/taishanglaojun-watch-debug.apk | awk '{print $5}')" >> "$BUILD_REPORT"
fi

if [ -f "$BUILD_OUTPUT_DIR/taishanglaojun-watch-release.apk" ]; then
    echo "- Wear OS Release APK: $(ls -lh $BUILD_OUTPUT_DIR/taishanglaojun-watch-release.apk | awk '{print $5}')" >> "$BUILD_REPORT"
fi

echo "" >> "$BUILD_REPORT"
echo "构建完成时间: $(date)" >> "$BUILD_REPORT"

log_success "构建完成！构建报告已保存到: $BUILD_REPORT"
log_info "构建输出目录: $BUILD_OUTPUT_DIR"