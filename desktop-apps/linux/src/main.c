#include <gtk/gtk.h>
#include <adwaita.h>
#include <glib/gi18n.h>
#include <locale.h>
#include "application.h"

int main(int argc, char *argv[]) {
    TaishangApplication *app;
    int status;
    
    // 设置本地化
    setlocale(LC_ALL, "");
    bindtextdomain(GETTEXT_PACKAGE, LOCALEDIR);
    bind_textdomain_codeset(GETTEXT_PACKAGE, "UTF-8");
    textdomain(GETTEXT_PACKAGE);
    
    // 初始化Adwaita
    adw_init();
    
    // 创建应用程序实例
    app = taishang_application_new();
    
    // 运行应用程序
    status = g_application_run(G_APPLICATION(app), argc, argv);
    
    // 清理
    g_object_unref(app);
    
    return status;
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