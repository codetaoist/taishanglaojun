#!/bin/bash

# RPM package post-removal script for TaishangLaojun Desktop

set -e

# Check if this is a complete removal or just an upgrade
if [ "$1" = "0" ]; then
    # Complete removal (not an upgrade)
    
    # Update desktop database
    if command -v update-desktop-database >/dev/null 2>&1; then
        update-desktop-database -q /usr/share/applications || true
    fi

    # Update icon cache
    if command -v gtk-update-icon-cache >/dev/null 2>&1; then
        gtk-update-icon-cache -q /usr/share/icons/hicolor || true
    fi

    # Update MIME database
    if command -v update-mime-database >/dev/null 2>&1; then
        update-mime-database /usr/share/mime || true
    fi

    # Compile GLib schemas
    if command -v glib-compile-schemas >/dev/null 2>&1; then
        glib-compile-schemas /usr/share/glib-2.0/schemas || true
    fi

    # Remove symlink
    rm -f /usr/local/bin/taishang-laojun 2>/dev/null || true

    # Remove systemd user service
    if [ -f /etc/systemd/user/taishang-laojun.service ]; then
        rm -f /etc/systemd/user/taishang-laojun.service
        if command -v systemctl >/dev/null 2>&1; then
            systemctl --global daemon-reload 2>/dev/null || true
        fi
    fi

    # Remove file associations
    if command -v xdg-mime >/dev/null 2>&1; then
        xdg-mime uninstall /usr/share/applications/taishang-laojun.desktop 2>/dev/null || true
    fi

    # Update shared library cache
    if command -v ldconfig >/dev/null 2>&1; then
        ldconfig || true
    fi

    # Ask user if they want to remove user data
    echo "TaishangLaojun Desktop has been removed."
    echo "User configuration and data files have been preserved."
    echo ""
    echo "To completely remove all user data, you can manually delete:"
    echo "  ~/.config/taishang-laojun"
    echo "  ~/.local/share/taishang-laojun"
    echo "  ~/.cache/taishang-laojun"
    echo ""
    echo "Or run the following command to remove all user data:"
    echo "  sudo rm -rf /home/*/.config/taishang-laojun /home/*/.local/share/taishang-laojun /home/*/.cache/taishang-laojun"

elif [ "$1" = "1" ]; then
    # This is an upgrade, don't remove user data
    echo "TaishangLaojun Desktop has been upgraded."
fi

exit 0