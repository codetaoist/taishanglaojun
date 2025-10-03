#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <signal.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <errno.h>
#include <locale.h>
#include <getopt.h>

#include "application.h"
#include "utils.h"

// 全局变量
static Application* g_app = NULL;
static volatile sig_atomic_t g_running = 1;

// 函数声明
static void signal_handler(int sig);
static bool setup_signal_handlers(void);
static void print_usage(const char* program_name);
static void print_version(void);
static bool check_single_instance(void);
static bool setup_directories(void);
static void cleanup_and_exit(int exit_code);

// 主函数
int main(int argc, char* argv[])
{
    int opt;
    bool daemon_mode = false;
    bool verbose = false;
    const char* config_file = NULL;
    
    // 命令行选项
    static struct option long_options[] = {
        {"daemon",    no_argument,       0, 'd'},
        {"verbose",   no_argument,       0, 'v'},
        {"config",    required_argument, 0, 'c'},
        {"help",      no_argument,       0, 'h'},
        {"version",   no_argument,       0, 'V'},
        {0, 0, 0, 0}
    };
    
    // 解析命令行参数
    while ((opt = getopt_long(argc, argv, "dvc:hV", long_options, NULL)) != -1) {
        switch (opt) {
        case 'd':
            daemon_mode = true;
            break;
        case 'v':
            verbose = true;
            break;
        case 'c':
            config_file = optarg;
            break;
        case 'h':
            print_usage(argv[0]);
            return EXIT_SUCCESS;
        case 'V':
            print_version();
            return EXIT_SUCCESS;
        default:
            print_usage(argv[0]);
            return EXIT_FAILURE;
        }
    }
    
    // 设置本地化
    setlocale(LC_ALL, "");
    
    // 初始化日志系统
    if (verbose) {
        log_init(NULL, LOG_LEVEL_DEBUG);
    } else {
        log_init("taishanglaojun-desktop.log", LOG_LEVEL_INFO);
    }
    
    LOG_INFO("太上老君AI平台桌面版启动 v%d.%d.%d", 
             APP_VERSION_MAJOR, APP_VERSION_MINOR, APP_VERSION_PATCH);
    
    // 检查单实例运行
    if (!check_single_instance()) {
        LOG_ERROR("应用程序已在运行");
        fprintf(stderr, "错误: 太上老君AI平台桌面版已在运行\n");
        return EXIT_FAILURE;
    }
    
    // 设置信号处理器
    if (!setup_signal_handlers()) {
        LOG_ERROR("设置信号处理器失败");
        return EXIT_FAILURE;
    }
    
    // 创建必要的目录
    if (!setup_directories()) {
        LOG_ERROR("创建应用目录失败");
        return EXIT_FAILURE;
    }
    
    // 如果是守护进程模式，进行守护进程化
    if (daemon_mode) {
        pid_t pid = fork();
        if (pid < 0) {
            LOG_ERROR("fork失败: %s", strerror(errno));
            return EXIT_FAILURE;
        }
        if (pid > 0) {
            // 父进程退出
            return EXIT_SUCCESS;
        }
        
        // 子进程继续
        if (setsid() < 0) {
            LOG_ERROR("setsid失败: %s", strerror(errno));
            return EXIT_FAILURE;
        }
        
        // 改变工作目录
        if (chdir("/") < 0) {
            LOG_ERROR("chdir失败: %s", strerror(errno));
            return EXIT_FAILURE;
        }
        
        // 关闭标准文件描述符
        close(STDIN_FILENO);
        close(STDOUT_FILENO);
        close(STDERR_FILENO);
    }
    
    // 创建应用程序实例
    g_app = application_create();
    if (!g_app) {
        LOG_ERROR("创建应用程序实例失败");
        cleanup_and_exit(EXIT_FAILURE);
    }
    
    // 初始化应用程序
    if (!application_initialize(g_app, argc, argv, config_file)) {
        LOG_ERROR("初始化应用程序失败");
        cleanup_and_exit(EXIT_FAILURE);
    }
    
    LOG_INFO("应用程序初始化完成");
    
    // 主循环
    while (g_running && application_is_running(g_app)) {
        if (!application_process_events(g_app)) {
            LOG_WARN("处理事件时出现错误");
        }
        
        // 处理应用程序逻辑
        application_update(g_app);
        
        // 短暂休眠以避免CPU占用过高
        usleep(1000); // 1ms
    }
    
    LOG_INFO("应用程序正在退出");
    cleanup_and_exit(EXIT_SUCCESS);
    
    return EXIT_SUCCESS; // 不会到达这里
}

// 信号处理器
static void signal_handler(int sig)
{
    switch (sig) {
    case SIGINT:
    case SIGTERM:
        LOG_INFO("收到退出信号 %d", sig);
        g_running = 0;
        if (g_app) {
            application_request_quit(g_app);
        }
        break;
    case SIGHUP:
        LOG_INFO("收到重新加载信号");
        if (g_app) {
            application_reload_config(g_app);
        }
        break;
    case SIGUSR1:
        LOG_INFO("收到用户信号1");
        // 可以用于调试或状态输出
        if (g_app) {
            application_print_status(g_app);
        }
        break;
    case SIGUSR2:
        LOG_INFO("收到用户信号2");
        // 可以用于切换调试模式
        if (g_app) {
            application_toggle_debug(g_app);
        }
        break;
    default:
        LOG_WARN("收到未处理的信号 %d", sig);
        break;
    }
}

// 设置信号处理器
static bool setup_signal_handlers(void)
{
    struct sigaction sa;
    
    // 设置信号处理函数
    sa.sa_handler = signal_handler;
    sigemptyset(&sa.sa_mask);
    sa.sa_flags = SA_RESTART;
    
    // 注册信号处理器
    if (sigaction(SIGINT, &sa, NULL) == -1) {
        LOG_ERROR("注册SIGINT处理器失败: %s", strerror(errno));
        return false;
    }
    
    if (sigaction(SIGTERM, &sa, NULL) == -1) {
        LOG_ERROR("注册SIGTERM处理器失败: %s", strerror(errno));
        return false;
    }
    
    if (sigaction(SIGHUP, &sa, NULL) == -1) {
        LOG_ERROR("注册SIGHUP处理器失败: %s", strerror(errno));
        return false;
    }
    
    if (sigaction(SIGUSR1, &sa, NULL) == -1) {
        LOG_ERROR("注册SIGUSR1处理器失败: %s", strerror(errno));
        return false;
    }
    
    if (sigaction(SIGUSR2, &sa, NULL) == -1) {
        LOG_ERROR("注册SIGUSR2处理器失败: %s", strerror(errno));
        return false;
    }
    
    // 忽略SIGPIPE信号
    signal(SIGPIPE, SIG_IGN);
    
    return true;
}

// 打印使用说明
static void print_usage(const char* program_name)
{
    printf("太上老君AI平台桌面版 v%d.%d.%d\n\n", 
           APP_VERSION_MAJOR, APP_VERSION_MINOR, APP_VERSION_PATCH);
    printf("用法: %s [选项]\n\n", program_name);
    printf("选项:\n");
    printf("  -d, --daemon          以守护进程模式运行\n");
    printf("  -v, --verbose         启用详细日志输出\n");
    printf("  -c, --config FILE     指定配置文件路径\n");
    printf("  -h, --help            显示此帮助信息\n");
    printf("  -V, --version         显示版本信息\n");
    printf("\n");
    printf("信号:\n");
    printf("  SIGINT/SIGTERM        优雅退出\n");
    printf("  SIGHUP                重新加载配置\n");
    printf("  SIGUSR1               打印状态信息\n");
    printf("  SIGUSR2               切换调试模式\n");
    printf("\n");
    printf("更多信息请访问: https://taishanglaojun.com\n");
}

// 打印版本信息
static void print_version(void)
{
    printf("太上老君AI平台桌面版 %d.%d.%d\n", 
           APP_VERSION_MAJOR, APP_VERSION_MINOR, APP_VERSION_PATCH);
    printf("构建日期: %s %s\n", __DATE__, __TIME__);
    printf("编译器: %s\n", __VERSION__);
    printf("\n");
    printf("版权所有 (C) 2024 太上老君团队\n");
    printf("本软件基于MIT许可证发布\n");
}

// 检查单实例运行
static bool check_single_instance(void)
{
    const char* lock_file = "/tmp/taishanglaojun-desktop.lock";
    FILE* fp;
    pid_t pid;
    
    // 尝试打开锁文件
    fp = fopen(lock_file, "r");
    if (fp) {
        // 文件存在，检查进程是否还在运行
        if (fscanf(fp, "%d", &pid) == 1) {
            fclose(fp);
            
            // 检查进程是否存在
            if (kill(pid, 0) == 0) {
                // 进程存在
                return false;
            }
        } else {
            fclose(fp);
        }
    }
    
    // 创建新的锁文件
    fp = fopen(lock_file, "w");
    if (!fp) {
        LOG_ERROR("无法创建锁文件 %s: %s", lock_file, strerror(errno));
        return false;
    }
    
    fprintf(fp, "%d\n", getpid());
    fclose(fp);
    
    return true;
}

// 创建必要的目录
static bool setup_directories(void)
{
    char* app_dir = get_app_data_directory();
    if (!app_dir) {
        LOG_ERROR("获取应用数据目录失败");
        return false;
    }
    
    // 创建主目录
    if (!create_directories(app_dir)) {
        LOG_ERROR("创建应用数据目录失败: %s", app_dir);
        free(app_dir);
        return false;
    }
    
    // 创建子目录
    char* cache_dir = path_join(app_dir, "cache");
    char* logs_dir = path_join(app_dir, "logs");
    char* config_dir = path_join(app_dir, "config");
    char* temp_dir = path_join(app_dir, "temp");
    
    bool success = true;
    
    if (!create_directories(cache_dir)) {
        LOG_ERROR("创建缓存目录失败: %s", cache_dir);
        success = false;
    }
    
    if (!create_directories(logs_dir)) {
        LOG_ERROR("创建日志目录失败: %s", logs_dir);
        success = false;
    }
    
    if (!create_directories(config_dir)) {
        LOG_ERROR("创建配置目录失败: %s", config_dir);
        success = false;
    }
    
    if (!create_directories(temp_dir)) {
        LOG_ERROR("创建临时目录失败: %s", temp_dir);
        success = false;
    }
    
    // 清理
    free(app_dir);
    free(cache_dir);
    free(logs_dir);
    free(config_dir);
    free(temp_dir);
    
    return success;
}

// 清理并退出
static void cleanup_and_exit(int exit_code)
{
    // 清理应用程序
    if (g_app) {
        application_shutdown(g_app);
        application_destroy(g_app);
        g_app = NULL;
    }
    
    // 删除锁文件
    unlink("/tmp/taishanglaojun-desktop.lock");
    
    // 清理日志系统
    log_cleanup();
    
    LOG_INFO("应用程序已退出，退出代码: %d", exit_code);
    
    exit(exit_code);
}