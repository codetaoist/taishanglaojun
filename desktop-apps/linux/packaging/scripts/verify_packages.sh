#!/bin/bash

# Script to verify package integrity and functionality
# Usage: verify_packages.sh <package_dir>

set -e

PACKAGE_DIR="$1"

if [ -z "$PACKAGE_DIR" ]; then
    echo "Usage: $0 <package_dir>"
    exit 1
fi

if [ ! -d "$PACKAGE_DIR" ]; then
    echo "Error: Package directory '$PACKAGE_DIR' not found"
    exit 1
fi

echo "Verifying packages in $PACKAGE_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status="$1"
    local message="$2"
    
    case "$status" in
        "PASS")
            echo -e "${GREEN}[PASS]${NC} $message"
            ;;
        "FAIL")
            echo -e "${RED}[FAIL]${NC} $message"
            ;;
        "WARN")
            echo -e "${YELLOW}[WARN]${NC} $message"
            ;;
        "INFO")
            echo -e "[INFO] $message"
            ;;
    esac
}

# Verification results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run test
run_test() {
    local test_name="$1"
    local test_command="$2"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if eval "$test_command" >/dev/null 2>&1; then
        print_status "PASS" "$test_name"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        print_status "FAIL" "$test_name"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# Verify AppImage
if [ -f "$PACKAGE_DIR"/*.AppImage ]; then
    print_status "INFO" "Verifying AppImage packages..."
    
    for appimage in "$PACKAGE_DIR"/*.AppImage; do
        print_status "INFO" "Checking $(basename "$appimage")"
        
        # Check if file is executable
        run_test "AppImage is executable" "[ -x '$appimage' ]"
        
        # Check file size (should be reasonable, not empty)
        run_test "AppImage has reasonable size" "[ $(stat -c%s '$appimage') -gt 10485760 ]" # > 10MB
        
        # Check if it's a valid AppImage
        run_test "AppImage has valid format" "file '$appimage' | grep -q 'ELF.*executable'"
        
        # Try to extract and verify contents
        if command -v unsquashfs >/dev/null 2>&1; then
            TEMP_DIR=$(mktemp -d)
            if "$appimage" --appimage-extract >/dev/null 2>&1; then
                run_test "AppImage can be extracted" "[ -d 'squashfs-root' ]"
                run_test "AppImage contains executable" "[ -f 'squashfs-root/usr/bin/taishang-laojun' ]"
                run_test "AppImage contains desktop file" "[ -f 'squashfs-root/taishang-laojun.desktop' ]"
                rm -rf squashfs-root
            else
                print_status "WARN" "Could not extract AppImage for verification"
            fi
        else
            print_status "WARN" "unsquashfs not available, skipping AppImage extraction test"
        fi
    done
else
    print_status "WARN" "No AppImage packages found"
fi

# Verify DEB packages
if [ -f "$PACKAGE_DIR"/*.deb ]; then
    print_status "INFO" "Verifying DEB packages..."
    
    for deb in "$PACKAGE_DIR"/*.deb; do
        print_status "INFO" "Checking $(basename "$deb")"
        
        # Check if dpkg can read the package
        if command -v dpkg >/dev/null 2>&1; then
            run_test "DEB package is valid" "dpkg --info '$deb' >/dev/null"
            run_test "DEB package has correct architecture" "dpkg --info '$deb' | grep -q 'Architecture: amd64'"
            run_test "DEB package has maintainer info" "dpkg --info '$deb' | grep -q 'Maintainer:'"
            run_test "DEB package has description" "dpkg --info '$deb' | grep -q 'Description:'"
            
            # Check package contents
            run_test "DEB contains executable" "dpkg --contents '$deb' | grep -q 'usr/bin/taishang-laojun'"
            run_test "DEB contains desktop file" "dpkg --contents '$deb' | grep -q 'usr/share/applications/'"
        else
            print_status "WARN" "dpkg not available, skipping DEB verification"
        fi
        
        # Check file size
        run_test "DEB has reasonable size" "[ $(stat -c%s '$deb') -gt 1048576 ]" # > 1MB
    done
else
    print_status "WARN" "No DEB packages found"
fi

# Verify RPM packages
if [ -f "$PACKAGE_DIR"/*.rpm ]; then
    print_status "INFO" "Verifying RPM packages..."
    
    for rpm in "$PACKAGE_DIR"/*.rpm; do
        print_status "INFO" "Checking $(basename "$rpm")"
        
        # Check if rpm can read the package
        if command -v rpm >/dev/null 2>&1; then
            run_test "RPM package is valid" "rpm -qp '$rpm' >/dev/null"
            run_test "RPM package has correct architecture" "rpm -qp --queryformat '%{ARCH}' '$rpm' | grep -q 'x86_64'"
            run_test "RPM package has summary" "rpm -qp --queryformat '%{SUMMARY}' '$rpm' | grep -q '.'"
            run_test "RPM package has description" "rpm -qp --queryformat '%{DESCRIPTION}' '$rpm' | grep -q '.'"
            
            # Check package contents
            run_test "RPM contains executable" "rpm -qlp '$rpm' | grep -q 'usr/bin/taishang-laojun'"
            run_test "RPM contains desktop file" "rpm -qlp '$rpm' | grep -q 'usr/share/applications/'"
        else
            print_status "WARN" "rpm not available, skipping RPM verification"
        fi
        
        # Check file size
        run_test "RPM has reasonable size" "[ $(stat -c%s '$rpm') -gt 1048576 ]" # > 1MB
    done
else
    print_status "WARN" "No RPM packages found"
fi

# Verify Flatpak packages
if [ -f "$PACKAGE_DIR"/*.flatpak ]; then
    print_status "INFO" "Verifying Flatpak packages..."
    
    for flatpak in "$PACKAGE_DIR"/*.flatpak; do
        print_status "INFO" "Checking $(basename "$flatpak")"
        
        # Check if flatpak can read the package
        if command -v flatpak >/dev/null 2>&1; then
            run_test "Flatpak bundle is valid" "flatpak info --show-metadata '$flatpak' >/dev/null"
            run_test "Flatpak has correct app ID" "flatpak info '$flatpak' | grep -q 'com.taishanglaojun.Desktop'"
        else
            print_status "WARN" "flatpak not available, skipping Flatpak verification"
        fi
        
        # Check file size
        run_test "Flatpak has reasonable size" "[ $(stat -c%s '$flatpak') -gt 10485760 ]" # > 10MB
    done
else
    print_status "WARN" "No Flatpak packages found"
fi

# Verify Snap packages
if [ -f "$PACKAGE_DIR"/*.snap ]; then
    print_status "INFO" "Verifying Snap packages..."
    
    for snap in "$PACKAGE_DIR"/*.snap; do
        print_status "INFO" "Checking $(basename "$snap")"
        
        # Check if snap can read the package
        if command -v snap >/dev/null 2>&1; then
            run_test "Snap package is valid" "snap info '$snap' >/dev/null"
        else
            print_status "WARN" "snap not available, skipping Snap verification"
        fi
        
        # Check file size
        run_test "Snap has reasonable size" "[ $(stat -c%s '$snap') -gt 10485760 ]" # > 10MB
    done
else
    print_status "WARN" "No Snap packages found"
fi

# Verify archive packages
for archive in "$PACKAGE_DIR"/*.tar.gz "$PACKAGE_DIR"/*.zip; do
    if [ -f "$archive" ]; then
        print_status "INFO" "Verifying $(basename "$archive")"
        
        case "$archive" in
            *.tar.gz)
                run_test "Archive can be listed" "tar -tzf '$archive' >/dev/null"
                run_test "Archive contains executable" "tar -tzf '$archive' | grep -q 'bin/taishang-laojun'"
                ;;
            *.zip)
                if command -v unzip >/dev/null 2>&1; then
                    run_test "Archive can be listed" "unzip -l '$archive' >/dev/null"
                    run_test "Archive contains executable" "unzip -l '$archive' | grep -q 'bin/taishang-laojun'"
                else
                    print_status "WARN" "unzip not available, skipping ZIP verification"
                fi
                ;;
        esac
        
        # Check file size
        run_test "Archive has reasonable size" "[ $(stat -c%s '$archive') -gt 1048576 ]" # > 1MB
    fi
done

# Security checks
print_status "INFO" "Running security checks..."

# Check for common security issues in packages
for package in "$PACKAGE_DIR"/*; do
    if [ -f "$package" ]; then
        # Check for setuid/setgid bits (should not be present)
        case "$package" in
            *.deb)
                if command -v dpkg >/dev/null 2>&1; then
                    if dpkg --contents "$package" | grep -q '^-rws\|^-r-s\|^-rwx.*s'; then
                        print_status "WARN" "$(basename "$package") contains setuid/setgid files"
                    else
                        run_test "No setuid/setgid files in $(basename "$package")" "true"
                    fi
                fi
                ;;
            *.rpm)
                if command -v rpm >/dev/null 2>&1; then
                    if rpm -qlp "$package" --dump | awk '{print $1, $5}' | grep -q 's'; then
                        print_status "WARN" "$(basename "$package") contains setuid/setgid files"
                    else
                        run_test "No setuid/setgid files in $(basename "$package")" "true"
                    fi
                fi
                ;;
        esac
    fi
done

# Summary
echo ""
print_status "INFO" "Verification Summary:"
print_status "INFO" "Total tests: $TOTAL_TESTS"
print_status "INFO" "Passed: $PASSED_TESTS"
print_status "INFO" "Failed: $FAILED_TESTS"

if [ $FAILED_TESTS -eq 0 ]; then
    print_status "PASS" "All package verifications passed!"
    exit 0
else
    print_status "FAIL" "$FAILED_TESTS test(s) failed"
    exit 1
fi