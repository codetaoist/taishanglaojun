#!/bin/bash

# 码道 CLI 构建脚本

set -e

echo "🚀 开始构建码道 CLI 工具..."

# 项目信息
APP_NAME="ct"
VERSION=${VERSION:-"1.0.0"}
BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S UTC')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS="-X 'main.Version=${VERSION}' -X 'main.BuildTime=${BUILD_TIME}' -X 'main.GitCommit=${GIT_COMMIT}'"

# 清理之前的构建
echo "🧹 清理构建目录..."
rm -rf dist/
mkdir -p dist/

# 构建不同平台的二进制文件
echo "🔨 构建多平台二进制文件..."

# Linux AMD64
echo "构建 Linux AMD64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "${LDFLAGS}" \
    -o dist/lo-linux-amd64 \
    ./cmd/lo

# Linux ARM64
echo "构建 Linux ARM64..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
    -ldflags "${LDFLAGS}" \
    -o dist/lo-linux-arm64 \
    ./cmd/lo

# macOS AMD64
echo "构建 macOS AMD64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
    -ldflags "${LDFLAGS}" \
    -o dist/lo-darwin-amd64 \
    ./cmd/lo

# macOS ARM64 (Apple Silicon)
echo "构建 macOS ARM64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
    -ldflags "${LDFLAGS}" \
    -o dist/lo-darwin-arm64 \
    ./cmd/lo

# Windows AMD64
echo "构建 Windows AMD64..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
    -ldflags "${LDFLAGS}" \
    -o dist/lo-windows-amd64.exe \
    ./cmd/lo

# Windows ARM64
echo "构建 Windows ARM64..."
CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build \
    -ldflags "${LDFLAGS}" \
    -o dist/lo-windows-arm64.exe \
    ./cmd/lo

# 创建符号链接（本地平台）
echo "🔗 创建本地符号链接..."
case "$(uname -s)" in
    Linux*)
        case "$(uname -m)" in
            x86_64) ln -sf lo-linux-amd64 dist/lo ;;
            aarch64) ln -sf lo-linux-arm64 dist/lo ;;
        esac
        ;;
    Darwin*)
        case "$(uname -m)" in
            x86_64) ln -sf lo-darwin-amd64 dist/lo ;;
            arm64) ln -sf lo-darwin-arm64 dist/lo ;;
        esac
        ;;
    MINGW*|MSYS*|CYGWIN*)
        case "$(uname -m)" in
            x86_64) cp dist/lo-windows-amd64.exe dist/lo.exe ;;
        esac
        ;;
esac

# 显示构建结果
echo "📦 构建完成！生成的文件："
ls -la dist/

# 计算文件大小和校验和
echo "📊 文件信息："
for file in dist/lo-*; do
    if [ -f "$file" ]; then
        size=$(du -h "$file" | cut -f1)
        checksum=$(sha256sum "$file" | cut -d' ' -f1)
        echo "$(basename "$file"): ${size}, SHA256: ${checksum}"
    fi
done

echo "✅ 构建完成！"
echo "💡 使用方法："
echo "   ./dist/lo --help"
echo "   ./dist/lo version"