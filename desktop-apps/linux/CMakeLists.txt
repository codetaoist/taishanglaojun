cmake_minimum_required(VERSION 3.16)
project(taishanglaojun-desktop VERSION 1.0.0 LANGUAGES C)

# 设置C标准
set(CMAKE_C_STANDARD 11)
set(CMAKE_C_STANDARD_REQUIRED ON)

# 设置输出目录
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin)
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib)
set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib)

# 编译选项
set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -Wall -Wextra -Werror")
set(CMAKE_C_FLAGS_DEBUG "${CMAKE_C_FLAGS_DEBUG} -g -O0 -DDEBUG")
set(CMAKE_C_FLAGS_RELEASE "${CMAKE_C_FLAGS_RELEASE} -O3 -DNDEBUG")

# 包含目录
include_directories(
    ${CMAKE_CURRENT_SOURCE_DIR}/include
    ${CMAKE_CURRENT_SOURCE_DIR}/../shared/common
    ${CMAKE_CURRENT_SOURCE_DIR}/../shared/protocols
)

# 查找依赖库
find_package(PkgConfig REQUIRED)

# X11库
find_package(X11 REQUIRED)
if(NOT X11_FOUND)
    message(FATAL_ERROR "X11 not found")
endif()

# Wayland库
pkg_check_modules(WAYLAND REQUIRED wayland-client wayland-server wayland-protocols)
pkg_check_modules(WAYLAND_EGL REQUIRED wayland-egl)

# OpenGL库
find_package(OpenGL REQUIRED)
if(NOT OPENGL_FOUND)
    message(FATAL_ERROR "OpenGL not found")
endif()

# GTK4库（现代化UI框架）
pkg_check_modules(GTK4 REQUIRED gtk4)
pkg_check_modules(ADWAITA REQUIRED libadwaita-1)

# 网络库
find_package(Threads REQUIRED)

# JSON库
pkg_check_modules(JSON_C REQUIRED json-c)

# 加密库
find_package(OpenSSL REQUIRED)

# 音频库
pkg_check_modules(ALSA REQUIRED alsa)
pkg_check_modules(PULSE REQUIRED libpulse)

# 通知库
pkg_check_modules(NOTIFY REQUIRED libnotify)

# 源文件
set(COMMON_SOURCES
    ../shared/common/utils.c
    ../shared/protocols/communication.c
)

set(LINUX_SOURCES
    src/main.c
    src/application.c
    src/window_manager.c
    src/desktop_pet.c
    src/file_transfer.c
    src/data_sync.c
    src/project_manager.c
    src/ui/main_window.c
    src/ui/pet_window.c
    src/ui/settings_dialog.c
    src/ui/file_transfer_dialog.c
    src/ui/project_view.c
    src/ui/gtk_helpers.c
    src/network/client.c
    src/network/server.c
    src/storage/database.c
    src/storage/cache.c
    src/crypto/encryption.c
    src/utils/system_info.c
    src/utils/notifications.c
    src/graphics/renderer.c
    src/graphics/x11_backend.c
    src/graphics/wayland_backend.c
    src/audio/audio_manager.c
)

# 创建可执行文件
add_executable(${PROJECT_NAME}
    ${COMMON_SOURCES}
    ${LINUX_SOURCES}
)

# 链接库
target_link_libraries(${PROJECT_NAME}
    ${X11_LIBRARIES}
    ${WAYLAND_LIBRARIES}
    ${WAYLAND_EGL_LIBRARIES}
    ${OPENGL_LIBRARIES}
    ${GTK4_LIBRARIES}
    ${ADWAITA_LIBRARIES}
    ${JSON_C_LIBRARIES}
    ${OPENSSL_LIBRARIES}
    ${ALSA_LIBRARIES}
    ${PULSE_LIBRARIES}
    ${NOTIFY_LIBRARIES}
    Threads::Threads
    m
    dl
    rt
)

# 编译定义
target_compile_definitions(${PROJECT_NAME} PRIVATE
    _GNU_SOURCE
    _DEFAULT_SOURCE
    APP_VERSION_MAJOR=${PROJECT_VERSION_MAJOR}
    APP_VERSION_MINOR=${PROJECT_VERSION_MINOR}
    APP_VERSION_PATCH=${PROJECT_VERSION_PATCH}
    APP_NAME="TaishangLaojun Desktop"
)

# 包含目录
target_include_directories(${PROJECT_NAME} PRIVATE
    ${X11_INCLUDE_DIRS}
    ${WAYLAND_INCLUDE_DIRS}
    ${GTK4_INCLUDE_DIRS}
    ${ADWAITA_INCLUDE_DIRS}
    ${JSON_C_INCLUDE_DIRS}
    ${OPENSSL_INCLUDE_DIRS}
    ${ALSA_INCLUDE_DIRS}
    ${PULSE_INCLUDE_DIRS}
    ${NOTIFY_INCLUDE_DIRS}
)

# 编译选项
target_compile_options(${PROJECT_NAME} PRIVATE
    ${GTK4_CFLAGS_OTHER}
    ${ADWAITA_CFLAGS_OTHER}
    ${WAYLAND_CFLAGS_OTHER}
    ${JSON_C_CFLAGS_OTHER}
    ${ALSA_CFLAGS_OTHER}
    ${PULSE_CFLAGS_OTHER}
    ${NOTIFY_CFLAGS_OTHER}
)

# 安装配置
install(TARGETS ${PROJECT_NAME}
    RUNTIME DESTINATION bin
)

install(DIRECTORY resources/
    DESTINATION share/${PROJECT_NAME}
    FILES_MATCHING 
    PATTERN "*.png" 
    PATTERN "*.svg" 
    PATTERN "*.desktop"
    PATTERN "*.xml"
)

install(FILES resources/taishanglaojun-desktop.desktop
    DESTINATION share/applications
)

install(FILES resources/icons/taishanglaojun-desktop.svg
    DESTINATION share/icons/hicolor/scalable/apps
)

# 创建AppImage所需的目录结构
set(APPDIR ${CMAKE_BINARY_DIR}/AppDir)
add_custom_target(appimage-prepare
    COMMAND ${CMAKE_COMMAND} -E make_directory ${APPDIR}/usr/bin
    COMMAND ${CMAKE_COMMAND} -E make_directory ${APPDIR}/usr/share/applications
    COMMAND ${CMAKE_COMMAND} -E make_directory ${APPDIR}/usr/share/icons/hicolor/scalable/apps
    COMMAND ${CMAKE_COMMAND} -E copy $<TARGET_FILE:${PROJECT_NAME}> ${APPDIR}/usr/bin/
    COMMAND ${CMAKE_COMMAND} -E copy ${CMAKE_CURRENT_SOURCE_DIR}/resources/taishanglaojun-desktop.desktop ${APPDIR}/usr/share/applications/
    COMMAND ${CMAKE_COMMAND} -E copy ${CMAKE_CURRENT_SOURCE_DIR}/resources/icons/taishanglaojun-desktop.svg ${APPDIR}/usr/share/icons/hicolor/scalable/apps/
    COMMAND ${CMAKE_COMMAND} -E copy ${CMAKE_CURRENT_SOURCE_DIR}/resources/taishanglaojun-desktop.desktop ${APPDIR}/
    COMMAND ${CMAKE_COMMAND} -E copy ${CMAKE_CURRENT_SOURCE_DIR}/resources/icons/taishanglaojun-desktop.svg ${APPDIR}/
    DEPENDS ${PROJECT_NAME}
)

# 打包配置
set(CPACK_GENERATOR "DEB;RPM;TGZ")
set(CPACK_PACKAGE_NAME "taishanglaojun-desktop")
set(CPACK_PACKAGE_VENDOR "TaishangLaojun Team")
set(CPACK_PACKAGE_DESCRIPTION_SUMMARY "太上老君AI平台桌面应用")
set(CPACK_PACKAGE_DESCRIPTION "太上老君AI平台的原生Linux桌面应用，提供桌面宠物、文件传输、数据同步和项目管理功能。")
set(CPACK_PACKAGE_VERSION ${PROJECT_VERSION})
set(CPACK_PACKAGE_CONTACT "support@taishanglaojun.com")
set(CPACK_RESOURCE_FILE_LICENSE "${CMAKE_CURRENT_SOURCE_DIR}/../LICENSE")

# DEB包配置
set(CPACK_DEBIAN_PACKAGE_DEPENDS "libx11-6, libwayland-client0, libgtk-4-1, libadwaita-1-0, libjson-c5, libssl3, libasound2, libpulse0, libnotify4")
set(CPACK_DEBIAN_PACKAGE_SECTION "utils")
set(CPACK_DEBIAN_PACKAGE_PRIORITY "optional")
set(CPACK_DEBIAN_PACKAGE_HOMEPAGE "https://taishanglaojun.com")

# RPM包配置
set(CPACK_RPM_PACKAGE_GROUP "Applications/Productivity")
set(CPACK_RPM_PACKAGE_LICENSE "MIT")
set(CPACK_RPM_PACKAGE_URL "https://taishanglaojun.com")
set(CPACK_RPM_PACKAGE_REQUIRES "libX11, wayland, gtk4, libadwaita, json-c, openssl, alsa-lib, pulseaudio-libs, libnotify")

include(CPack)

# 测试配置
enable_testing()

add_subdirectory(tests)

# 自定义目标
add_custom_target(run
    COMMAND ${CMAKE_RUNTIME_OUTPUT_DIRECTORY}/${PROJECT_NAME}
    DEPENDS ${PROJECT_NAME}
    WORKING_DIRECTORY ${CMAKE_RUNTIME_OUTPUT_DIRECTORY}
)

add_custom_target(clean-all
    COMMAND ${CMAKE_COMMAND} -E remove_directory ${CMAKE_BINARY_DIR}
)

# 代码格式化
find_program(CLANG_FORMAT clang-format)
if(CLANG_FORMAT)
    add_custom_target(format
        COMMAND ${CLANG_FORMAT} -i -style=file ${LINUX_SOURCES}
        WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    )
endif()

# 静态分析
find_program(CPPCHECK cppcheck)
if(CPPCHECK)
    add_custom_target(cppcheck
        COMMAND ${CPPCHECK} --enable=all --std=c11 --project=${CMAKE_BINARY_DIR}/compile_commands.json
        WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    )
endif()

# Valgrind内存检查
find_program(VALGRIND valgrind)
if(VALGRIND)
    add_custom_target(memcheck
        COMMAND ${VALGRIND} --tool=memcheck --leak-check=full --show-leak-kinds=all --track-origins=yes ${CMAKE_RUNTIME_OUTPUT_DIRECTORY}/${PROJECT_NAME}
        DEPENDS ${PROJECT_NAME}
        WORKING_DIRECTORY ${CMAKE_RUNTIME_OUTPUT_DIRECTORY}
    )
endif()

# 生成编译数据库
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

# 调试信息
if(CMAKE_BUILD_TYPE STREQUAL "Debug")
    target_compile_definitions(${PROJECT_NAME} PRIVATE DEBUG_BUILD)
    message(STATUS "Debug build enabled")
endif()

# 显示配置信息
message(STATUS "Build type: ${CMAKE_BUILD_TYPE}")
message(STATUS "C compiler: ${CMAKE_C_COMPILER}")
message(STATUS "Install prefix: ${CMAKE_INSTALL_PREFIX}")
message(STATUS "X11 found: ${X11_FOUND}")
message(STATUS "Wayland found: ${WAYLAND_FOUND}")
message(STATUS "GTK4 found: ${GTK4_FOUND}")
message(STATUS "Adwaita found: ${ADWAITA_FOUND}")
message(STATUS "OpenGL found: ${OPENGL_FOUND}")
message(STATUS "OpenSSL found: ${OPENSSL_FOUND}")