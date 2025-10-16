# Packaging CMakeLists.txt
cmake_minimum_required(VERSION 3.16)

# AppImage packaging
find_program(LINUXDEPLOY_TOOL linuxdeploy)
find_program(APPIMAGETOOL_TOOL appimagetool)

if(LINUXDEPLOY_TOOL AND APPIMAGETOOL_TOOL)
    set(APPDIR ${CMAKE_CURRENT_BINARY_DIR}/AppDir)
    
    # Create AppImage target
    add_custom_target(appimage
        COMMAND ${CMAKE_COMMAND} -E remove_directory ${APPDIR}
        COMMAND ${CMAKE_COMMAND} -E make_directory ${APPDIR}
        COMMAND ${LINUXDEPLOY_TOOL} --appdir ${APPDIR} --executable ${CMAKE_BINARY_DIR}/bin/taishang-laojun --desktop-file ${CMAKE_SOURCE_DIR}/data/taishang-laojun.desktop --icon-file ${CMAKE_SOURCE_DIR}/data/taishang-laojun.png
        COMMAND ${APPIMAGETOOL_TOOL} ${APPDIR} ${CMAKE_CURRENT_BINARY_DIR}/TaishangLaojun-${PROJECT_VERSION}-x86_64.AppImage
        DEPENDS taishang-laojun
        WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
        COMMENT "Creating AppImage package"
    )
    
    message(STATUS "AppImage packaging enabled")
else()
    message(WARNING "linuxdeploy or appimagetool not found, AppImage packaging disabled")
endif()

# Alternative AppImage creation without linuxdeploy
add_custom_target(appimage-manual
    COMMAND ${CMAKE_COMMAND} -E remove_directory ${CMAKE_CURRENT_BINARY_DIR}/AppDir
    COMMAND ${CMAKE_COMMAND} -E make_directory ${CMAKE_CURRENT_BINARY_DIR}/AppDir/usr/bin
    COMMAND ${CMAKE_COMMAND} -E make_directory ${CMAKE_CURRENT_BINARY_DIR}/AppDir/usr/lib
    COMMAND ${CMAKE_COMMAND} -E make_directory ${CMAKE_CURRENT_BINARY_DIR}/AppDir/usr/share/applications
    COMMAND ${CMAKE_COMMAND} -E make_directory ${CMAKE_CURRENT_BINARY_DIR}/AppDir/usr/share/icons/hicolor/256x256/apps
    
    # Copy executable
    COMMAND ${CMAKE_COMMAND} -E copy ${CMAKE_BINARY_DIR}/bin/taishang-laojun ${CMAKE_CURRENT_BINARY_DIR}/AppDir/usr/bin/
    
    # Copy desktop file and icon
    COMMAND ${CMAKE_COMMAND} -E copy ${CMAKE_SOURCE_DIR}/data/taishang-laojun.desktop ${CMAKE_CURRENT_BINARY_DIR}/AppDir/usr/share/applications/
    COMMAND ${CMAKE_COMMAND} -E copy ${CMAKE_SOURCE_DIR}/data/taishang-laojun.png ${CMAKE_CURRENT_BINARY_DIR}/AppDir/usr/share/icons/hicolor/256x256/apps/
    
    # Create AppRun script
    COMMAND ${CMAKE_COMMAND} -E echo "#!/bin/bash" > ${CMAKE_CURRENT_BINARY_DIR}/AppDir/AppRun
    COMMAND ${CMAKE_COMMAND} -E echo "cd \"\$(dirname \"\$0\")\"" >> ${CMAKE_CURRENT_BINARY_DIR}/AppDir/AppRun
    COMMAND ${CMAKE_COMMAND} -E echo "exec ./usr/bin/taishang-laojun \"\$@\"" >> ${CMAKE_CURRENT_BINARY_DIR}/AppDir/AppRun
    COMMAND chmod +x ${CMAKE_CURRENT_BINARY_DIR}/AppDir/AppRun
    
    # Copy desktop file to root
    COMMAND ${CMAKE_COMMAND} -E copy ${CMAKE_SOURCE_DIR}/data/taishang-laojun.desktop ${CMAKE_CURRENT_BINARY_DIR}/AppDir/
    COMMAND ${CMAKE_COMMAND} -E copy ${CMAKE_SOURCE_DIR}/data/taishang-laojun.png ${CMAKE_CURRENT_BINARY_DIR}/AppDir/
    
    # Copy dependencies
    COMMAND ${CMAKE_CURRENT_SOURCE_DIR}/scripts/copy_dependencies.sh ${CMAKE_BINARY_DIR}/bin/taishang-laojun ${CMAKE_CURRENT_BINARY_DIR}/AppDir/usr/lib
    
    DEPENDS taishang-laojun
    WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
    COMMENT "Creating AppImage directory structure manually"
)

# DEB package configuration
set(CPACK_GENERATOR "DEB")
set(CPACK_PACKAGE_NAME "taishang-laojun")
set(CPACK_PACKAGE_VENDOR "TaishangLaojun Team")
set(CPACK_PACKAGE_DESCRIPTION_SUMMARY "TaishangLaojun AI Platform Desktop Application")
set(CPACK_PACKAGE_DESCRIPTION "Native Linux desktop application for TaishangLaojun AI Platform, providing chat interface, project management, file transfer, and system integration.")
set(CPACK_PACKAGE_VERSION ${PROJECT_VERSION})
set(CPACK_PACKAGE_CONTACT "support@taishanglaojun.com")
set(CPACK_PACKAGE_HOMEPAGE_URL "https://taishanglaojun.com")

# DEB specific settings
set(CPACK_DEBIAN_PACKAGE_MAINTAINER "TaishangLaojun Team <support@taishanglaojun.com>")
set(CPACK_DEBIAN_PACKAGE_SECTION "utils")
set(CPACK_DEBIAN_PACKAGE_PRIORITY "optional")
set(CPACK_DEBIAN_PACKAGE_DEPENDS "libgtk-4-1 (>= 4.6), libglib2.0-0 (>= 2.66), libjson-glib-1.0-0, libsoup-3.0-0, libwebkit2gtk-4.1-0, libsqlite3-0, libgl1-mesa-glx, libegl1-mesa, libsndfile1, libdbus-1-3")
set(CPACK_DEBIAN_PACKAGE_RECOMMENDS "libnotify4, libayatana-appindicator3-1")
set(CPACK_DEBIAN_PACKAGE_SUGGESTS "pulseaudio, alsa-utils")

# DEB package control files
set(CPACK_DEBIAN_PACKAGE_CONTROL_EXTRA 
    "${CMAKE_CURRENT_SOURCE_DIR}/debian/postinst"
    "${CMAKE_CURRENT_SOURCE_DIR}/debian/prerm"
    "${CMAKE_CURRENT_SOURCE_DIR}/debian/postrm"
)

# RPM package configuration
set(CPACK_RPM_PACKAGE_SUMMARY "TaishangLaojun AI Platform Desktop Application")
set(CPACK_RPM_PACKAGE_DESCRIPTION "Native Linux desktop application for TaishangLaojun AI Platform, providing chat interface, project management, file transfer, and system integration.")
set(CPACK_RPM_PACKAGE_GROUP "Applications/Internet")
set(CPACK_RPM_PACKAGE_LICENSE "MIT")
set(CPACK_RPM_PACKAGE_URL "https://taishanglaojun.com")
set(CPACK_RPM_PACKAGE_VENDOR "TaishangLaojun Team")
set(CPACK_RPM_PACKAGE_REQUIRES "gtk4 >= 4.6, glib2 >= 2.66, json-glib, libsoup3, webkit2gtk4.1, sqlite, mesa-libGL, mesa-libEGL, libsndfile, dbus")
set(CPACK_RPM_PACKAGE_SUGGESTS "libnotify, pulseaudio, alsa-utils")

# RPM package scripts
set(CPACK_RPM_POST_INSTALL_SCRIPT_FILE "${CMAKE_CURRENT_SOURCE_DIR}/rpm/postinst.sh")
set(CPACK_RPM_PRE_UNINSTALL_SCRIPT_FILE "${CMAKE_CURRENT_SOURCE_DIR}/rpm/prerm.sh")
set(CPACK_RPM_POST_UNINSTALL_SCRIPT_FILE "${CMAKE_CURRENT_SOURCE_DIR}/rpm/postrm.sh")

# Flatpak configuration
configure_file(
    ${CMAKE_CURRENT_SOURCE_DIR}/flatpak/com.taishanglaojun.Desktop.yml.in
    ${CMAKE_CURRENT_BINARY_DIR}/com.taishanglaojun.Desktop.yml
    @ONLY
)

add_custom_target(flatpak
    COMMAND flatpak-builder --force-clean ${CMAKE_CURRENT_BINARY_DIR}/flatpak-build ${CMAKE_CURRENT_BINARY_DIR}/com.taishanglaojun.Desktop.yml
    COMMAND flatpak build-export ${CMAKE_CURRENT_BINARY_DIR}/flatpak-repo ${CMAKE_CURRENT_BINARY_DIR}/flatpak-build
    COMMAND flatpak build-bundle ${CMAKE_CURRENT_BINARY_DIR}/flatpak-repo ${CMAKE_CURRENT_BINARY_DIR}/TaishangLaojun-${PROJECT_VERSION}.flatpak com.taishanglaojun.Desktop
    DEPENDS taishang-laojun
    WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
    COMMENT "Creating Flatpak package"
)

# Snap configuration
configure_file(
    ${CMAKE_CURRENT_SOURCE_DIR}/snap/snapcraft.yaml.in
    ${CMAKE_CURRENT_BINARY_DIR}/snap/snapcraft.yaml
    @ONLY
)

add_custom_target(snap
    COMMAND ${CMAKE_COMMAND} -E copy_directory ${CMAKE_CURRENT_SOURCE_DIR}/snap ${CMAKE_CURRENT_BINARY_DIR}/snap
    COMMAND snapcraft --destructive-mode
    DEPENDS taishang-laojun
    WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
    COMMENT "Creating Snap package"
)

# Archive packages
add_custom_target(tar-package
    COMMAND ${CMAKE_COMMAND} -E tar czf TaishangLaojun-${PROJECT_VERSION}-linux-x86_64.tar.gz 
        --format=gnutar
        -C ${CMAKE_BINARY_DIR}
        bin/taishang-laojun
        share/
    DEPENDS taishang-laojun
    WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
    COMMENT "Creating tar.gz package"
)

add_custom_target(zip-package
    COMMAND ${CMAKE_COMMAND} -E tar cf TaishangLaojun-${PROJECT_VERSION}-linux-x86_64.zip 
        --format=zip
        -C ${CMAKE_BINARY_DIR}
        bin/taishang-laojun
        share/
    DEPENDS taishang-laojun
    WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
    COMMENT "Creating zip package"
)

# Package verification
add_custom_target(verify-packages
    COMMAND ${CMAKE_CURRENT_SOURCE_DIR}/scripts/verify_packages.sh ${CMAKE_CURRENT_BINARY_DIR}
    WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
    COMMENT "Verifying package integrity"
)

# Clean packaging artifacts
add_custom_target(clean-packages
    COMMAND ${CMAKE_COMMAND} -E remove_directory ${CMAKE_CURRENT_BINARY_DIR}/AppDir
    COMMAND ${CMAKE_COMMAND} -E remove_directory ${CMAKE_CURRENT_BINARY_DIR}/flatpak-build
    COMMAND ${CMAKE_COMMAND} -E remove_directory ${CMAKE_CURRENT_BINARY_DIR}/flatpak-repo
    COMMAND ${CMAKE_COMMAND} -E remove_directory ${CMAKE_CURRENT_BINARY_DIR}/snap
    COMMAND ${CMAKE_COMMAND} -E remove ${CMAKE_CURRENT_BINARY_DIR}/*.AppImage
    COMMAND ${CMAKE_COMMAND} -E remove ${CMAKE_CURRENT_BINARY_DIR}/*.deb
    COMMAND ${CMAKE_COMMAND} -E remove ${CMAKE_CURRENT_BINARY_DIR}/*.rpm
    COMMAND ${CMAKE_COMMAND} -E remove ${CMAKE_CURRENT_BINARY_DIR}/*.flatpak
    COMMAND ${CMAKE_COMMAND} -E remove ${CMAKE_CURRENT_BINARY_DIR}/*.snap
    COMMAND ${CMAKE_COMMAND} -E remove ${CMAKE_CURRENT_BINARY_DIR}/*.tar.gz
    COMMAND ${CMAKE_COMMAND} -E remove ${CMAKE_CURRENT_BINARY_DIR}/*.zip
    WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
    COMMENT "Cleaning packaging artifacts"
)

# All packages target
add_custom_target(all-packages
    DEPENDS appimage-manual tar-package zip-package
    COMMENT "Building all available packages"
)

include(CPack)