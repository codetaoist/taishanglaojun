#include <iostream>
#include <string>
#include <vector>
#include <windows.h>

// 测试框架宏
#define TEST_ASSERT(condition, message) \
    do { \
        if (!(condition)) { \
            std::cerr << "FAIL: " << message << " at " << __FILE__ << ":" << __LINE__ << std::endl; \
            return false; \
        } \
    } while(0)

#define TEST_ASSERT_EQ(expected, actual, message) \
    do { \
        if ((expected) != (actual)) { \
            std::cerr << "FAIL: " << message << " - Expected: " << (expected) << ", Actual: " << (actual) \
                      << " at " << __FILE__ << ":" << __LINE__ << std::endl; \
            return false; \
        } \
    } while(0)

#define RUN_TEST(test_func) \
    do { \
        std::cout << "Running " << #test_func << "..." << std::endl; \
        if (!test_func()) { \
            std::cerr << "Test " << #test_func << " failed!" << std::endl; \
            failed_tests++; \
        } else { \
            std::cout << "Test " << #test_func << " passed!" << std::endl; \
            passed_tests++; \
        } \
    } while(0)

// 前向声明测试函数
bool test_application_init();
bool test_application_shutdown();
bool test_auth_manager_login();
bool test_auth_manager_logout();
bool test_chat_manager_send_message();
bool test_chat_manager_receive_message();
bool test_file_transfer_upload();
bool test_file_transfer_download();
bool test_desktop_pet_show();
bool test_desktop_pet_hide();
bool test_http_client_get();
bool test_http_client_post();

int main(int argc, char* argv[]) {
    std::cout << "=== Windows Desktop App Test Suite ===" << std::endl;
    
    int passed_tests = 0;
    int failed_tests = 0;
    
    // 检查命令行参数
    std::string test_filter = "";
    if (argc > 1) {
        test_filter = argv[1];
    }
    
    // 运行测试
    if (test_filter.empty() || test_filter == "--test-application") {
        RUN_TEST(test_application_init);
        RUN_TEST(test_application_shutdown);
    }
    
    if (test_filter.empty() || test_filter == "--test-auth") {
        RUN_TEST(test_auth_manager_login);
        RUN_TEST(test_auth_manager_logout);
    }
    
    if (test_filter.empty() || test_filter == "--test-chat") {
        RUN_TEST(test_chat_manager_send_message);
        RUN_TEST(test_chat_manager_receive_message);
    }
    
    if (test_filter.empty() || test_filter == "--test-file-transfer") {
        RUN_TEST(test_file_transfer_upload);
        RUN_TEST(test_file_transfer_download);
    }
    
    if (test_filter.empty() || test_filter == "--test-desktop-pet") {
        RUN_TEST(test_desktop_pet_show);
        RUN_TEST(test_desktop_pet_hide);
    }
    
    if (test_filter.empty() || test_filter == "--test-http-client") {
        RUN_TEST(test_http_client_get);
        RUN_TEST(test_http_client_post);
    }
    
    // 输出测试结果
    std::cout << std::endl << "=== Test Results ===" << std::endl;
    std::cout << "Passed: " << passed_tests << std::endl;
    std::cout << "Failed: " << failed_tests << std::endl;
    std::cout << "Total:  " << (passed_tests + failed_tests) << std::endl;
    
    if (failed_tests > 0) {
        std::cout << "Some tests failed!" << std::endl;
        return 1;
    } else {
        std::cout << "All tests passed!" << std::endl;
        return 0;
    }
}