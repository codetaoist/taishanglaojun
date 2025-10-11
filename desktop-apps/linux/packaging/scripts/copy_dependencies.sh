#!/bin/bash

# Script to copy shared library dependencies for AppImage packaging
# Usage: copy_dependencies.sh <executable> <target_lib_dir>

set -e

EXECUTABLE="$1"
TARGET_LIB_DIR="$2"

if [ -z "$EXECUTABLE" ] || [ -z "$TARGET_LIB_DIR" ]; then
    echo "Usage: $0 <executable> <target_lib_dir>"
    exit 1
fi

if [ ! -f "$EXECUTABLE" ]; then
    echo "Error: Executable '$EXECUTABLE' not found"
    exit 1
fi

# Create target directory
mkdir -p "$TARGET_LIB_DIR"

echo "Copying dependencies for $EXECUTABLE to $TARGET_LIB_DIR"

# System library directories to exclude
SYSTEM_DIRS=(
    "/lib"
    "/lib64"
    "/usr/lib"
    "/usr/lib64"
    "/lib/x86_64-linux-gnu"
    "/usr/lib/x86_64-linux-gnu"
)

# Libraries to exclude (system libraries that should be available on all systems)
EXCLUDE_LIBS=(
    "linux-vdso.so"
    "ld-linux-x86-64.so"
    "libc.so"
    "libm.so"
    "libdl.so"
    "libpthread.so"
    "librt.so"
    "libresolv.so"
    "libnss_"
    "libnsl.so"
    "libutil.so"
    "libcrypt.so"
    "libgcc_s.so"
    "libstdc++.so"
)

# Function to check if a library should be excluded
should_exclude() {
    local lib="$1"
    local lib_basename=$(basename "$lib")
    
    # Check if it's in a system directory
    for sys_dir in "${SYSTEM_DIRS[@]}"; do
        if [[ "$lib" == "$sys_dir"* ]]; then
            return 0
        fi
    done
    
    # Check if it's in the exclude list
    for exclude_lib in "${EXCLUDE_LIBS[@]}"; do
        if [[ "$lib_basename" == *"$exclude_lib"* ]]; then
            return 0
        fi
    done
    
    return 1
}

# Function to copy library and its dependencies recursively
copy_lib_recursive() {
    local lib="$1"
    local copied_libs="$2"
    
    # Skip if already processed
    if [[ "$copied_libs" == *"|$lib|"* ]]; then
        return
    fi
    
    # Skip if should be excluded
    if should_exclude "$lib"; then
        return
    fi
    
    # Skip if library doesn't exist
    if [ ! -f "$lib" ]; then
        return
    fi
    
    echo "Copying: $lib"
    cp "$lib" "$TARGET_LIB_DIR/"
    
    # Mark as copied
    copied_libs="$copied_libs|$lib|"
    
    # Get dependencies of this library
    local deps=$(ldd "$lib" 2>/dev/null | grep "=>" | awk '{print $3}' | grep -v "^$")
    
    # Recursively copy dependencies
    for dep in $deps; do
        copy_lib_recursive "$dep" "$copied_libs"
    done
}

# Get direct dependencies of the executable
echo "Analyzing dependencies of $EXECUTABLE..."
DEPS=$(ldd "$EXECUTABLE" | grep "=>" | awk '{print $3}' | grep -v "^$")

# Copy each dependency and its dependencies
copied_libs="|"
for dep in $DEPS; do
    copy_lib_recursive "$dep" "$copied_libs"
done

# Copy GTK and GLib modules if they exist
echo "Copying GTK and GLib modules..."

# GTK modules
GTK_MODULES_DIR="/usr/lib/x86_64-linux-gnu/gtk-4.0"
if [ -d "$GTK_MODULES_DIR" ]; then
    mkdir -p "$TARGET_LIB_DIR/../share/gtk-4.0"
    cp -r "$GTK_MODULES_DIR" "$TARGET_LIB_DIR/../share/"
fi

# GLib schemas
GLIB_SCHEMAS_DIR="/usr/share/glib-2.0/schemas"
if [ -d "$GLIB_SCHEMAS_DIR" ]; then
    mkdir -p "$TARGET_LIB_DIR/../share/glib-2.0"
    cp -r "$GLIB_SCHEMAS_DIR" "$TARGET_LIB_DIR/../share/glib-2.0/"
fi

# GIO modules
GIO_MODULES_DIR="/usr/lib/x86_64-linux-gnu/gio/modules"
if [ -d "$GIO_MODULES_DIR" ]; then
    mkdir -p "$TARGET_LIB_DIR/gio"
    cp -r "$GIO_MODULES_DIR" "$TARGET_LIB_DIR/gio/"
fi

# Pixbuf loaders
PIXBUF_LOADERS_DIR="/usr/lib/x86_64-linux-gnu/gdk-pixbuf-2.0"
if [ -d "$PIXBUF_LOADERS_DIR" ]; then
    mkdir -p "$TARGET_LIB_DIR/../share/"
    cp -r "$PIXBUF_LOADERS_DIR" "$TARGET_LIB_DIR/../share/"
fi

# Create library path script
cat > "$TARGET_LIB_DIR/../AppRun" << 'EOF'
#!/bin/bash

# Get the directory where this script is located
APPDIR="$(dirname "$(readlink -f "$0")")"

# Set library path
export LD_LIBRARY_PATH="$APPDIR/usr/lib:$LD_LIBRARY_PATH"

# Set GTK and GLib paths
export GTK_PATH="$APPDIR/usr/share/gtk-4.0"
export GIO_MODULE_DIR="$APPDIR/usr/lib/gio/modules"
export GDK_PIXBUF_MODULE_DIR="$APPDIR/usr/share/gdk-pixbuf-2.0/2.10.0/loaders"
export GDK_PIXBUF_MODULEDIR="$APPDIR/usr/share/gdk-pixbuf-2.0/2.10.0/loaders"

# Set XDG paths
export XDG_DATA_DIRS="$APPDIR/usr/share:$XDG_DATA_DIRS"

# Execute the application
exec "$APPDIR/usr/bin/taishang-laojun" "$@"
EOF

chmod +x "$TARGET_LIB_DIR/../AppRun"

echo "Dependencies copied successfully!"
echo "Libraries copied to: $TARGET_LIB_DIR"
echo "Total libraries: $(ls -1 "$TARGET_LIB_DIR" | wc -l)"

# Verify dependencies
echo "Verifying dependencies..."
MISSING_DEPS=$(ldd "$EXECUTABLE" | grep "not found" || true)
if [ -n "$MISSING_DEPS" ]; then
    echo "Warning: Missing dependencies found:"
    echo "$MISSING_DEPS"
else
    echo "All dependencies satisfied!"
fi