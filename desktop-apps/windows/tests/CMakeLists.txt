# Windows Desktop App Tests
cmake_minimum_required(VERSION 3.20)

# 测试可执行文件
set(TEST_SOURCES
    test_main.cpp
    test_application.cpp
    test_auth_manager.cpp
    test_chat_manager.cpp
    test_file_transfer.cpp
    test_desktop_pet.cpp
    test_http_client.cpp
)

# 创建测试可执行文件
add_executable(${PROJECT_NAME}_tests ${TEST_SOURCES})

# 链接主应用的源文件（除了main.cpp）
target_sources(${PROJECT_NAME}_tests PRIVATE
    ${CMAKE_SOURCE_DIR}/src/application.cpp
    ${CMAKE_SOURCE_DIR}/src/auth_manager.cpp
    ${CMAKE_SOURCE_DIR}/src/chat_manager.cpp
    ${CMAKE_SOURCE_DIR}/src/file_transfer.cpp
    ${CMAKE_SOURCE_DIR}/src/desktop_pet.cpp
    ${CMAKE_SOURCE_DIR}/src/http_client.cpp
    ${CMAKE_SOURCE_DIR}/src/data_sync.cpp
    ${CMAKE_SOURCE_DIR}/src/friend_manager.cpp
    ${CMAKE_SOURCE_DIR}/src/project_management.cpp
)

# 包含目录
target_include_directories(${PROJECT_NAME}_tests PRIVATE
    ${CMAKE_SOURCE_DIR}/src
    ${CMAKE_SOURCE_DIR}/include
    ${CMAKE_SOURCE_DIR}/../shared/common/include
    ${CMAKE_SOURCE_DIR}/../shared/common/protocol
)

# 链接库
target_link_libraries(${PROJECT_NAME}_tests
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
)

# 编译定义
target_compile_definitions(${PROJECT_NAME}_tests PRIVATE
    WIN32_LEAN_AND_MEAN
    NOMINMAX
    _WIN32_WINNT=0x0A00
    WINVER=0x0A00
    TESTING_MODE=1
)

# 添加测试
add_test(NAME ApplicationTests COMMAND ${PROJECT_NAME}_tests --test-application)
add_test(NAME AuthManagerTests COMMAND ${PROJECT_NAME}_tests --test-auth)
add_test(NAME ChatManagerTests COMMAND ${PROJECT_NAME}_tests --test-chat)
add_test(NAME FileTransferTests COMMAND ${PROJECT_NAME}_tests --test-file-transfer)
add_test(NAME DesktopPetTests COMMAND ${PROJECT_NAME}_tests --test-desktop-pet)
add_test(NAME HttpClientTests COMMAND ${PROJECT_NAME}_tests --test-http-client)

# 设置测试属性
set_tests_properties(ApplicationTests PROPERTIES TIMEOUT 30)
set_tests_properties(AuthManagerTests PROPERTIES TIMEOUT 60)
set_tests_properties(ChatManagerTests PROPERTIES TIMEOUT 60)
set_tests_properties(FileTransferTests PROPERTIES TIMEOUT 120)
set_tests_properties(DesktopPetTests PROPERTIES TIMEOUT 30)
set_tests_properties(HttpClientTests PROPERTIES TIMEOUT 60)