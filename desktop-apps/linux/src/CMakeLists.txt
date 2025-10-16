# TaishangLaojun Source Files CMake Configuration

# Set minimum CMake version
cmake_minimum_required(VERSION 3.16)

# Source files
set(TAISHANG_SOURCES
    main.c
    app.c
    ui.c
    config.c
    utils.c
)

# Header files (for IDE support)
set(TAISHANG_HEADERS
    ${CMAKE_SOURCE_DIR}/include/app.h
    ${CMAKE_SOURCE_DIR}/include/ui.h
    ${CMAKE_SOURCE_DIR}/include/config.h
    ${CMAKE_SOURCE_DIR}/include/utils.h
    ${CMAKE_SOURCE_DIR}/include/common.h
)

# Create the main executable
add_executable(taishang-laojun ${TAISHANG_SOURCES} ${TAISHANG_HEADERS})

# Set target properties
set_target_properties(taishang-laojun PROPERTIES
    C_STANDARD 11
    C_STANDARD_REQUIRED ON
    C_EXTENSIONS OFF
    OUTPUT_NAME "taishang-laojun"
)

# Include directories
target_include_directories(taishang-laojun PRIVATE
    ${CMAKE_SOURCE_DIR}/include
    ${CMAKE_BINARY_DIR}/include  # For generated headers
)

# Link libraries
target_link_libraries(taishang-laojun PRIVATE
    ${GTK3_LIBRARIES}
    ${GLIB2_LIBRARIES}
    ${JSON_C_LIBRARIES}
    ${OPENSSL_LIBRARIES}
    ${SQLITE3_LIBRARIES}
    ${CURL_LIBRARIES}
    PkgConfig::LIBNOTIFY
    Threads::Threads
)

# Compiler definitions
target_compile_definitions(taishang-laojun PRIVATE
    _GNU_SOURCE
    G_LOG_DOMAIN="TaishangLaojun"
    TAISHANG_VERSION="${PROJECT_VERSION}"
    TAISHANG_DATADIR="${CMAKE_INSTALL_FULL_DATADIR}"
    TAISHANG_LOCALEDIR="${CMAKE_INSTALL_FULL_LOCALEDIR}"
    $<$<CONFIG:Debug>:DEBUG>
    $<$<CONFIG:Debug>:G_ENABLE_DEBUG>
)

# Compiler options
target_compile_options(taishang-laojun PRIVATE
    -Wall
    -Wextra
    -Wformat-security
    -Wno-unused-parameter
    $<$<CONFIG:Debug>:-g3>
    $<$<CONFIG:Debug>:-O0>
    $<$<CONFIG:Release>:-O2>
    $<$<CONFIG:Release>:-DNDEBUG>
)

# Install the executable
install(TARGETS taishang-laojun
    RUNTIME DESTINATION ${CMAKE_INSTALL_BINDIR}
    COMPONENT Runtime
)

# Install debug symbols if available
if(CMAKE_BUILD_TYPE STREQUAL "Debug" OR CMAKE_BUILD_TYPE STREQUAL "RelWithDebInfo")
    install(FILES $<TARGET_FILE:taishang-laojun>.debug
        DESTINATION ${CMAKE_INSTALL_LIBDIR}/debug${CMAKE_INSTALL_BINDIR}
        COMPONENT Debug
        OPTIONAL
    )
endif()

# Add custom target for source code analysis
if(ENABLE_STATIC_ANALYSIS)
    find_program(CPPCHECK_EXECUTABLE cppcheck)
    if(CPPCHECK_EXECUTABLE)
        add_custom_target(cppcheck-src
            COMMAND ${CPPCHECK_EXECUTABLE}
                --enable=all
                --std=c11
                --verbose
                --quiet
                --error-exitcode=1
                --inline-suppr
                --suppress=missingIncludeSystem
                --suppress=unusedFunction
                -I ${CMAKE_SOURCE_DIR}/include
                ${CMAKE_CURRENT_SOURCE_DIR}
            WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
            COMMENT "Running cppcheck on source files"
        )
        add_dependencies(taishang-laojun cppcheck-src)
    endif()
endif()

# Add custom target for code formatting
find_program(CLANG_FORMAT_EXECUTABLE clang-format)
if(CLANG_FORMAT_EXECUTABLE)
    add_custom_target(format-src
        COMMAND ${CLANG_FORMAT_EXECUTABLE}
            -i
            -style=file
            ${TAISHANG_SOURCES}
        WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
        COMMENT "Formatting source files with clang-format"
    )
endif()

# Add custom target for generating compile commands
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

# Platform-specific configurations
if(UNIX AND NOT APPLE)
    # Linux-specific settings
    target_compile_definitions(taishang-laojun PRIVATE
        _DEFAULT_SOURCE
        _POSIX_C_SOURCE=200809L
    )
    
    # Link additional libraries for Linux
    target_link_libraries(taishang-laojun PRIVATE
        m  # Math library
        dl # Dynamic loading library
    )
endif()

# Coverage support
if(ENABLE_COVERAGE AND CMAKE_BUILD_TYPE STREQUAL "Debug")
    target_compile_options(taishang-laojun PRIVATE
        --coverage
        -fprofile-arcs
        -ftest-coverage
    )
    target_link_options(taishang-laojun PRIVATE
        --coverage
    )
endif()

# Sanitizer support
if(ENABLE_SANITIZERS AND CMAKE_BUILD_TYPE STREQUAL "Debug")
    target_compile_options(taishang-laojun PRIVATE
        -fsanitize=address
        -fsanitize=undefined
        -fno-omit-frame-pointer
    )
    target_link_options(taishang-laojun PRIVATE
        -fsanitize=address
        -fsanitize=undefined
    )
endif()

# Resource embedding (if enabled)
if(EMBED_RESOURCES)
    # Generate resource files
    add_custom_command(
        OUTPUT ${CMAKE_BINARY_DIR}/resources.c
        COMMAND ${CMAKE_SOURCE_DIR}/scripts/embed_resources.sh
            ${CMAKE_SOURCE_DIR}/resources
            ${CMAKE_BINARY_DIR}/resources.c
        DEPENDS ${CMAKE_SOURCE_DIR}/resources
        COMMENT "Embedding resources"
    )
    
    target_sources(taishang-laojun PRIVATE
        ${CMAKE_BINARY_DIR}/resources.c
    )
endif()

# Print configuration summary
message(STATUS "TaishangLaojun executable configuration:")
message(STATUS "  Target name: taishang-laojun")
message(STATUS "  Source files: ${TAISHANG_SOURCES}")
message(STATUS "  Install destination: ${CMAKE_INSTALL_BINDIR}")
message(STATUS "  C standard: C11")
message(STATUS "  Build type: ${CMAKE_BUILD_TYPE}")
if(ENABLE_COVERAGE)
    message(STATUS "  Coverage: Enabled")
endif()
if(ENABLE_SANITIZERS)
    message(STATUS "  Sanitizers: Enabled")
endif()
if(ENABLE_STATIC_ANALYSIS)
    message(STATUS "  Static analysis: Enabled")
endif()