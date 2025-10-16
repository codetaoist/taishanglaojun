cmake_minimum_required(VERSION 3.20)
project(TaishangLaojunDesktop VERSION 1.0.0 LANGUAGES C CXX)

# 设置C++标准
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_C_STANDARD 11)
set(CMAKE_C_STANDARD_REQUIRED ON)

# 设置输出目录
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin)
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib)
set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib)

# 编译选项
if(MSVC)
    add_compile_options(/W4 /WX /utf-8)
    add_compile_definitions(_CRT_SECURE_NO_WARNINGS)
    add_compile_definitions(UNICODE _UNICODE)
else()
    add_compile_options(-Wall -Wextra -Werror)
endif()

# 包含目录
include_directories(
    ${CMAKE_CURRENT_SOURCE_DIR}/include
    ${CMAKE_CURRENT_SOURCE_DIR}/../shared/common
    ${CMAKE_CURRENT_SOURCE_DIR}/../shared/protocols
)

# 查找依赖库
find_package(OpenGL REQUIRED)

# 源文件
set(COMMON_SOURCES
    ../shared/common/utils.c
    ../shared/protocols/communication.c
)

set(WINDOWS_SOURCES
    src/main.cpp
    src/application.cpp
    src/desktop_pet.cpp
    src/file_transfer.cpp
    src/data_sync.cpp
    src/auth_manager.cpp
    src/chat_manager.cpp
    src/project_management.cpp
    src/friend_manager.cpp
    src/http_client.cpp
)

# 头文件
set(WINDOWS_HEADERS
    src/pch.h
    src/application.h
    src/desktop_pet.h
    src/file_transfer.h
    src/data_sync.h
    src/auth_manager.h
    src/chat_manager.h
    src/project_management.h
    src/friend_manager.h
    src/http_client.h
)

set(RESOURCE_FILES
    resources/app.rc
    resources/icon.ico
    resources/manifest.xml
)

# 创建可执行文件
add_executable(${PROJECT_NAME} WIN32
    ${COMMON_SOURCES}
    ${WINDOWS_SOURCES}
    ${WINDOWS_HEADERS}
    ${RESOURCE_FILES}
)

# 链接库
target_link_libraries(${PROJECT_NAME}
    ${OPENGL_LIBRARIES}
    user32
    gdi32
    shell32
    advapi32
    ole32
    oleaut32
    uuid
    comctl32
    ws2_32
    wininet
    crypt32
    bcrypt
    d3d11
    dxgi
    d2d1
    dwrite
)

# 预编译头
target_precompile_headers(${PROJECT_NAME} PRIVATE
    src/pch.h
)

# 编译定义
target_compile_definitions(${PROJECT_NAME} PRIVATE
    WIN32_LEAN_AND_MEAN
    NOMINMAX
    _WIN32_WINNT=0x0A00  # Windows 10
    WINVER=0x0A00
    APP_VERSION_MAJOR=${PROJECT_VERSION_MAJOR}
    APP_VERSION_MINOR=${PROJECT_VERSION_MINOR}
    APP_VERSION_PATCH=${PROJECT_VERSION_PATCH}
)

# 安装配置
install(TARGETS ${PROJECT_NAME}
    RUNTIME DESTINATION bin
)

install(DIRECTORY resources/
    DESTINATION share/${PROJECT_NAME}
    FILES_MATCHING PATTERN "*.png" PATTERN "*.jpg" PATTERN "*.svg"
)

# 打包配置
set(CPACK_GENERATOR "WIX;ZIP")
set(CPACK_PACKAGE_NAME "TaishangLaojun Desktop")
set(CPACK_PACKAGE_VENDOR "TaishangLaojun Team")
set(CPACK_PACKAGE_DESCRIPTION_SUMMARY "太上老君AI平台桌面应用")
set(CPACK_PACKAGE_VERSION ${PROJECT_VERSION})
set(CPACK_PACKAGE_INSTALL_DIRECTORY "TaishangLaojun")
set(CPACK_RESOURCE_FILE_LICENSE "${CMAKE_CURRENT_SOURCE_DIR}/../LICENSE")

# WiX特定配置
set(CPACK_WIX_UPGRADE_GUID "12345678-1234-1234-1234-123456789012")
set(CPACK_WIX_PRODUCT_GUID "87654321-4321-4321-4321-210987654321")
set(CPACK_WIX_PRODUCT_ICON "${CMAKE_CURRENT_SOURCE_DIR}/resources/icon.ico")
set(CPACK_WIX_UI_BANNER "${CMAKE_CURRENT_SOURCE_DIR}/resources/banner.bmp")
set(CPACK_WIX_UI_DIALOG "${CMAKE_CURRENT_SOURCE_DIR}/resources/dialog.bmp")

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
        COMMAND ${CLANG_FORMAT} -i -style=file ${WINDOWS_SOURCES}
        WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    )
endif()

# 静态分析
find_program(CPPCHECK cppcheck)
if(CPPCHECK)
    add_custom_target(cppcheck
        COMMAND ${CPPCHECK} --enable=all --std=c++17 --project=${CMAKE_BINARY_DIR}/compile_commands.json
        WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
    )
endif()

# 生成编译数据库
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)