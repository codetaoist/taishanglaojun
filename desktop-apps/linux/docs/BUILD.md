# Building TaishangLaojun Desktop for Linux

This document provides comprehensive instructions for building TaishangLaojun Desktop on Linux systems.

## Prerequisites

### System Requirements

- **Operating System**: Linux (Ubuntu 20.04+, Fedora 35+, Arch Linux, or equivalent)
- **Architecture**: x86_64 or ARM64
- **Memory**: Minimum 2GB RAM (4GB recommended)
- **Storage**: At least 2GB free space for build dependencies and output

### Build Dependencies

#### Ubuntu/Debian
```bash
sudo apt update
sudo apt install -y \
    build-essential \
    cmake \
    ninja-build \
    pkg-config \
    git \
    libgtk-4-dev \
    libadwaita-1-dev \
    libglib2.0-dev \
    libjson-c-dev \
    libssl-dev \
    libasound2-dev \
    libpulse-dev \
    libsndfile1-dev \
    libnotify-dev \
    libappindicator3-dev \
    libx11-dev \
    libwayland-dev \
    libgl1-mesa-dev \
    libegl1-mesa-dev \
    libdbus-1-dev \
    libsqlite3-dev \
    desktop-file-utils \
    shared-mime-info
```

#### Fedora/RHEL/CentOS
```bash
sudo dnf install -y \
    gcc gcc-c++ \
    cmake \
    ninja-build \
    pkgconfig \
    git \
    gtk4-devel \
    libadwaita-devel \
    glib2-devel \
    json-c-devel \
    openssl-devel \
    alsa-lib-devel \
    pulseaudio-libs-devel \
    libsndfile-devel \
    libnotify-devel \
    libappindicator-gtk3-devel \
    libX11-devel \
    wayland-devel \
    mesa-libGL-devel \
    mesa-libEGL-devel \
    dbus-devel \
    sqlite-devel \
    desktop-file-utils \
    shared-mime-info
```

#### Arch Linux
```bash
sudo pacman -S \
    base-devel \
    cmake \
    ninja \
    pkgconf \
    git \
    gtk4 \
    libadwaita \
    glib2 \
    json-c \
    openssl \
    alsa-lib \
    libpulse \
    libsndfile \
    libnotify \
    libappindicator-gtk3 \
    libx11 \
    wayland \
    mesa \
    dbus \
    sqlite \
    desktop-file-utils \
    shared-mime-info
```

### Optional Dependencies

For additional features and packaging:

```bash
# For testing
sudo apt install -y glib-testing-framework valgrind

# For code coverage
sudo apt install -y gcov lcov

# For static analysis
sudo apt install -y cppcheck clang-tools

# For packaging
sudo apt install -y dpkg-dev rpm-build flatpak-builder snapcraft

# For AppImage creation
wget https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage
chmod +x linuxdeploy-x86_64.AppImage
sudo mv linuxdeploy-x86_64.AppImage /usr/local/bin/linuxdeploy
```

## Building from Source

### 1. Clone the Repository

```bash
git clone https://github.com/taishanglaojun/desktop-apps.git
cd desktop-apps/linux
```

### 2. Configure the Build

#### Standard Release Build
```bash
mkdir build
cd build
cmake .. -G Ninja \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_INSTALL_PREFIX=/usr \
    -DBUILD_TESTS=ON \
    -DBUILD_PACKAGES=ON
```

#### Debug Build
```bash
mkdir build-debug
cd build-debug
cmake .. -G Ninja \
    -DCMAKE_BUILD_TYPE=Debug \
    -DCMAKE_INSTALL_PREFIX=/usr \
    -DBUILD_TESTS=ON \
    -DENABLE_COVERAGE=ON
```

#### Custom Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `CMAKE_BUILD_TYPE` | `Release` | Build type (Release, Debug, RelWithDebInfo) |
| `CMAKE_INSTALL_PREFIX` | `/usr/local` | Installation prefix |
| `BUILD_TESTS` | `OFF` | Build test suite |
| `BUILD_PACKAGES` | `OFF` | Enable package generation |
| `ENABLE_COVERAGE` | `OFF` | Enable code coverage (Debug builds only) |
| `ENABLE_ASAN` | `OFF` | Enable AddressSanitizer |
| `ENABLE_TSAN` | `OFF` | Enable ThreadSanitizer |
| `ENABLE_UBSAN` | `OFF` | Enable UndefinedBehaviorSanitizer |
| `USE_SYSTEM_LIBS` | `ON` | Use system libraries instead of bundled |

### 3. Build the Application

```bash
ninja
```

Or with make:
```bash
make -j$(nproc)
```

### 4. Run Tests (Optional)

```bash
ninja test
# or
ctest --output-on-failure
```

### 5. Install

```bash
sudo ninja install
# or
sudo make install
```

## Development Build

For development with automatic rebuilding:

```bash
mkdir build-dev
cd build-dev
cmake .. -G Ninja \
    -DCMAKE_BUILD_TYPE=Debug \
    -DCMAKE_EXPORT_COMPILE_COMMANDS=ON \
    -DBUILD_TESTS=ON \
    -DENABLE_COVERAGE=ON

# Build and run
ninja && ./src/taishang-laojun
```

## Packaging

### AppImage

```bash
cd build
ninja appimage
```

The AppImage will be created in `packaging/appimage/`.

### DEB Package

```bash
cd build
ninja package-deb
```

### RPM Package

```bash
cd build
ninja package-rpm
```

### Flatpak

```bash
cd build
ninja flatpak
```

### Snap

```bash
cd build
ninja snap
```

### All Packages

```bash
cd build
ninja all-packages
```

## Code Quality and Analysis

### Static Analysis

```bash
ninja cppcheck
```

### Code Coverage

```bash
# Debug build with coverage enabled
ninja coverage
```

Coverage report will be generated in `coverage/html/index.html`.

### Memory Checking

```bash
ninja memcheck
```

### Code Formatting

```bash
ninja format
```

## Troubleshooting

### Common Build Issues

#### Missing Dependencies
```bash
# Check for missing packages
pkg-config --list-all | grep -E "(gtk|glib|json|ssl)"

# Install missing development packages
sudo apt install -y libgtk-4-dev  # example
```

#### CMake Configuration Issues
```bash
# Clear CMake cache
rm -rf CMakeCache.txt CMakeFiles/

# Reconfigure with verbose output
cmake .. -DCMAKE_VERBOSE_MAKEFILE=ON
```

#### Linker Errors
```bash
# Check library paths
ldconfig -p | grep -E "(gtk|glib|json)"

# Update library cache
sudo ldconfig
```

### Runtime Issues

#### Missing Shared Libraries
```bash
# Check dependencies
ldd ./src/taishang-laojun

# Install missing runtime packages
sudo apt install -y libgtk-4-1  # example
```

#### Permission Issues
```bash
# Fix file permissions
chmod +x ./src/taishang-laojun

# Check SELinux context (RHEL/Fedora)
ls -Z ./src/taishang-laojun
```

## Performance Optimization

### Compiler Optimizations

```bash
cmake .. -G Ninja \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_C_FLAGS="-O3 -march=native -flto" \
    -DCMAKE_EXE_LINKER_FLAGS="-flto"
```

### Profile-Guided Optimization (PGO)

```bash
# Build with profiling
cmake .. -DCMAKE_C_FLAGS="-fprofile-generate"
ninja
./src/taishang-laojun  # Run typical workload
ninja clean

# Build with profile data
cmake .. -DCMAKE_C_FLAGS="-fprofile-use"
ninja
```

## Cross-Compilation

### For ARM64

```bash
# Install cross-compilation toolchain
sudo apt install -y gcc-aarch64-linux-gnu

# Configure for ARM64
cmake .. -G Ninja \
    -DCMAKE_SYSTEM_NAME=Linux \
    -DCMAKE_SYSTEM_PROCESSOR=aarch64 \
    -DCMAKE_C_COMPILER=aarch64-linux-gnu-gcc \
    -DCMAKE_FIND_ROOT_PATH=/usr/aarch64-linux-gnu
```

## Continuous Integration

### GitHub Actions Example

```yaml
name: Build and Test
on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Install dependencies
      run: |
        sudo apt update
        sudo apt install -y build-essential cmake ninja-build libgtk-4-dev
    - name: Configure
      run: |
        mkdir build
        cd build
        cmake .. -G Ninja -DBUILD_TESTS=ON
    - name: Build
      run: |
        cd build
        ninja
    - name: Test
      run: |
        cd build
        ctest --output-on-failure
```

## Additional Resources

- [CMake Documentation](https://cmake.org/documentation/)
- [GTK4 Documentation](https://docs.gtk.org/gtk4/)
- [GLib Documentation](https://docs.gtk.org/glib/)
- [Project Wiki](https://github.com/taishanglaojun/desktop-apps/wiki)

For more help, visit our [documentation site](https://docs.taishanglaojun.com) or open an issue on [GitHub](https://github.com/taishanglaojun/desktop-apps/issues).