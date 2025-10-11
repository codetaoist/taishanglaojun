# Test CMakeLists.txt
cmake_minimum_required(VERSION 3.16)

# Find GLib testing framework
pkg_check_modules(GLIB_TEST REQUIRED glib-2.0 gobject-2.0 gio-2.0)

# Test include directories
include_directories(
    ${CMAKE_SOURCE_DIR}/src
    ${CMAKE_SOURCE_DIR}/include
    ${CMAKE_CURRENT_SOURCE_DIR}
    ${GLIB_TEST_INCLUDE_DIRS}
    ${GTK4_INCLUDE_DIRS}
    ${JSON_GLIB_INCLUDE_DIRS}
    ${SOUP_INCLUDE_DIRS}
    ${WEBKIT_INCLUDE_DIRS}
    ${OPENGL_INCLUDE_DIRS}
    ${EGL_INCLUDE_DIRS}
    ${SNDFILE_INCLUDE_DIRS}
    ${DBUS_INCLUDE_DIRS}
)

# Test source files
set(TEST_SOURCES
    test_main.c
    test_network.c
    test_storage.c
    test_system.c
    test_graphics.c
    test_audio.c
    test_ui.c
)

# Test headers
set(TEST_HEADERS
    test_network.h
    test_storage.h
    test_system.h
    test_graphics.h
    test_audio.h
    test_ui.h
)

# Application source files (for testing)
set(APP_TEST_SOURCES
    ${CMAKE_SOURCE_DIR}/src/application.c
    ${CMAKE_SOURCE_DIR}/src/ui/main_window.c
    ${CMAKE_SOURCE_DIR}/src/ui/chat_interface.c
    ${CMAKE_SOURCE_DIR}/src/ui/project_manager.c
    ${CMAKE_SOURCE_DIR}/src/ui/settings_dialog.c
    ${CMAKE_SOURCE_DIR}/src/ui/friend_list.c
    ${CMAKE_SOURCE_DIR}/src/ui/notification_manager.c
    ${CMAKE_SOURCE_DIR}/src/network/api_client.c
    ${CMAKE_SOURCE_DIR}/src/network/websocket_client.c
    ${CMAKE_SOURCE_DIR}/src/storage/database.c
    ${CMAKE_SOURCE_DIR}/src/storage/cache_manager.c
    ${CMAKE_SOURCE_DIR}/src/storage/file_manager.c
    ${CMAKE_SOURCE_DIR}/src/system/dbus_client.c
    ${CMAKE_SOURCE_DIR}/src/system/desktop_integration.c
    ${CMAKE_SOURCE_DIR}/src/system/system_info.c
    ${CMAKE_SOURCE_DIR}/src/graphics/renderer.c
    ${CMAKE_SOURCE_DIR}/src/audio/audio_manager.c
)

# Create test executable
add_executable(taishang-tests
    ${TEST_SOURCES}
    ${TEST_HEADERS}
    ${APP_TEST_SOURCES}
)

# Link test libraries
target_link_libraries(taishang-tests
    ${GLIB_TEST_LIBRARIES}
    ${GTK4_LIBRARIES}
    ${JSON_GLIB_LIBRARIES}
    ${SOUP_LIBRARIES}
    ${WEBKIT_LIBRARIES}
    ${OPENGL_LIBRARIES}
    ${EGL_LIBRARIES}
    ${SNDFILE_LIBRARIES}
    ${DBUS_LIBRARIES}
    SQLite::SQLite3
    m
)

# Optional test libraries
if(PULSEAUDIO_FOUND)
    target_link_libraries(taishang-tests ${PULSEAUDIO_LIBRARIES})
endif()

if(ALSA_FOUND)
    target_link_libraries(taishang-tests ${ALSA_LIBRARIES})
endif()

if(LIBNOTIFY_FOUND)
    target_link_libraries(taishang-tests ${LIBNOTIFY_LIBRARIES})
endif()

if(APPINDICATOR_FOUND)
    target_link_libraries(taishang-tests ${APPINDICATOR_LIBRARIES})
endif()

# Test compile definitions
target_compile_definitions(taishang-tests PRIVATE
    TESTING_BUILD
    TEST_DATA_DIR="${CMAKE_CURRENT_SOURCE_DIR}/data"
    TEST_OUTPUT_DIR="${CMAKE_CURRENT_BINARY_DIR}/output"
)

# Test compile options
target_compile_options(taishang-tests PRIVATE
    ${GLIB_TEST_CFLAGS_OTHER}
    ${GTK4_CFLAGS_OTHER}
    ${JSON_GLIB_CFLAGS_OTHER}
    ${SOUP_CFLAGS_OTHER}
    ${WEBKIT_CFLAGS_OTHER}
    ${EGL_CFLAGS_OTHER}
    ${SNDFILE_CFLAGS_OTHER}
    ${DBUS_CFLAGS_OTHER}
)

# Individual test executables for better isolation
add_executable(test-network
    test_main.c
    test_network.c
    ${CMAKE_SOURCE_DIR}/src/network/api_client.c
    ${CMAKE_SOURCE_DIR}/src/network/websocket_client.c
)

add_executable(test-storage
    test_main.c
    test_storage.c
    ${CMAKE_SOURCE_DIR}/src/storage/database.c
    ${CMAKE_SOURCE_DIR}/src/storage/cache_manager.c
    ${CMAKE_SOURCE_DIR}/src/storage/file_manager.c
)

add_executable(test-system
    test_main.c
    test_system.c
    ${CMAKE_SOURCE_DIR}/src/system/dbus_client.c
    ${CMAKE_SOURCE_DIR}/src/system/desktop_integration.c
    ${CMAKE_SOURCE_DIR}/src/system/system_info.c
)

add_executable(test-graphics
    test_main.c
    test_graphics.c
    ${CMAKE_SOURCE_DIR}/src/graphics/renderer.c
)

add_executable(test-audio
    test_main.c
    test_audio.c
    ${CMAKE_SOURCE_DIR}/src/audio/audio_manager.c
)

add_executable(test-ui
    test_main.c
    test_ui.c
    ${CMAKE_SOURCE_DIR}/src/ui/main_window.c
    ${CMAKE_SOURCE_DIR}/src/ui/chat_interface.c
    ${CMAKE_SOURCE_DIR}/src/ui/project_manager.c
    ${CMAKE_SOURCE_DIR}/src/ui/settings_dialog.c
    ${CMAKE_SOURCE_DIR}/src/ui/friend_list.c
    ${CMAKE_SOURCE_DIR}/src/ui/notification_manager.c
)

# Link libraries for individual tests
foreach(test_target test-network test-storage test-system test-graphics test-audio test-ui)
    target_link_libraries(${test_target}
        ${GLIB_TEST_LIBRARIES}
        ${GTK4_LIBRARIES}
        ${JSON_GLIB_LIBRARIES}
        ${SOUP_LIBRARIES}
        ${WEBKIT_LIBRARIES}
        ${OPENGL_LIBRARIES}
        ${EGL_LIBRARIES}
        ${SNDFILE_LIBRARIES}
        ${DBUS_LIBRARIES}
        SQLite::SQLite3
        m
    )
    
    target_include_directories(${test_target} PRIVATE
        ${CMAKE_SOURCE_DIR}/src
        ${CMAKE_SOURCE_DIR}/include
        ${CMAKE_CURRENT_SOURCE_DIR}
        ${GLIB_TEST_INCLUDE_DIRS}
        ${GTK4_INCLUDE_DIRS}
        ${JSON_GLIB_INCLUDE_DIRS}
        ${SOUP_INCLUDE_DIRS}
        ${WEBKIT_INCLUDE_DIRS}
        ${OPENGL_INCLUDE_DIRS}
        ${EGL_INCLUDE_DIRS}
        ${SNDFILE_INCLUDE_DIRS}
        ${DBUS_INCLUDE_DIRS}
    )
    
    target_compile_definitions(${test_target} PRIVATE
        TESTING_BUILD
        TEST_DATA_DIR="${CMAKE_CURRENT_SOURCE_DIR}/data"
        TEST_OUTPUT_DIR="${CMAKE_CURRENT_BINARY_DIR}/output"
    )
endforeach()

# Add tests to CTest
add_test(NAME NetworkTests COMMAND test-network)
add_test(NAME StorageTests COMMAND test-storage)
add_test(NAME SystemTests COMMAND test-system)
add_test(NAME GraphicsTests COMMAND test-graphics)
add_test(NAME AudioTests COMMAND test-audio)
add_test(NAME UITests COMMAND test-ui)
add_test(NAME AllTests COMMAND taishang-tests)

# Test environment setup
set_tests_properties(NetworkTests PROPERTIES
    ENVIRONMENT "G_TEST_SRCDIR=${CMAKE_CURRENT_SOURCE_DIR};G_TEST_BUILDDIR=${CMAKE_CURRENT_BINARY_DIR}"
)

set_tests_properties(StorageTests PROPERTIES
    ENVIRONMENT "G_TEST_SRCDIR=${CMAKE_CURRENT_SOURCE_DIR};G_TEST_BUILDDIR=${CMAKE_CURRENT_BINARY_DIR}"
)

set_tests_properties(SystemTests PROPERTIES
    ENVIRONMENT "G_TEST_SRCDIR=${CMAKE_CURRENT_SOURCE_DIR};G_TEST_BUILDDIR=${CMAKE_CURRENT_BINARY_DIR}"
)

set_tests_properties(GraphicsTests PROPERTIES
    ENVIRONMENT "G_TEST_SRCDIR=${CMAKE_CURRENT_SOURCE_DIR};G_TEST_BUILDDIR=${CMAKE_CURRENT_BINARY_DIR};DISPLAY=:99"
)

set_tests_properties(AudioTests PROPERTIES
    ENVIRONMENT "G_TEST_SRCDIR=${CMAKE_CURRENT_SOURCE_DIR};G_TEST_BUILDDIR=${CMAKE_CURRENT_BINARY_DIR}"
)

set_tests_properties(UITests PROPERTIES
    ENVIRONMENT "G_TEST_SRCDIR=${CMAKE_CURRENT_SOURCE_DIR};G_TEST_BUILDDIR=${CMAKE_CURRENT_BINARY_DIR};DISPLAY=:99"
)

set_tests_properties(AllTests PROPERTIES
    ENVIRONMENT "G_TEST_SRCDIR=${CMAKE_CURRENT_SOURCE_DIR};G_TEST_BUILDDIR=${CMAKE_CURRENT_BINARY_DIR};DISPLAY=:99"
)

# Test data directory
file(MAKE_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}/data)
file(MAKE_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}/output)

# Copy test data files
configure_file(${CMAKE_CURRENT_SOURCE_DIR}/data/test_config.json
               ${CMAKE_CURRENT_BINARY_DIR}/data/test_config.json COPYONLY)

# Coverage target (if gcov is available)
find_program(GCOV_PATH gcov)
find_program(LCOV_PATH lcov)
find_program(GENHTML_PATH genhtml)

if(GCOV_PATH AND LCOV_PATH AND GENHTML_PATH AND CMAKE_BUILD_TYPE STREQUAL "Debug")
    target_compile_options(taishang-tests PRIVATE --coverage)
    target_link_libraries(taishang-tests --coverage)
    
    add_custom_target(coverage
        COMMAND ${LCOV_PATH} --directory . --zerocounters
        COMMAND ${CMAKE_CTEST_COMMAND} --output-on-failure
        COMMAND ${LCOV_PATH} --directory . --capture --output-file coverage.info
        COMMAND ${LCOV_PATH} --remove coverage.info '/usr/*' --output-file coverage.info
        COMMAND ${LCOV_PATH} --remove coverage.info '*/tests/*' --output-file coverage.info
        COMMAND ${GENHTML_PATH} -o coverage coverage.info
        WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
        COMMENT "Generating code coverage report"
    )
endif()

# Valgrind testing
if(VALGRIND)
    add_custom_target(test-memcheck
        COMMAND ${VALGRIND} --tool=memcheck --leak-check=full --show-leak-kinds=all --track-origins=yes --error-exitcode=1 ${CMAKE_CURRENT_BINARY_DIR}/taishang-tests
        DEPENDS taishang-tests
        WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
        COMMENT "Running tests with Valgrind memory checking"
    )
endif()

# Performance testing
add_custom_target(test-performance
    COMMAND ${CMAKE_CURRENT_BINARY_DIR}/taishang-tests --performance
    DEPENDS taishang-tests
    WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
    COMMENT "Running performance tests"
)

# Stress testing
add_custom_target(test-stress
    COMMAND ${CMAKE_CURRENT_BINARY_DIR}/taishang-tests --stress
    DEPENDS taishang-tests
    WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
    COMMENT "Running stress tests"
)