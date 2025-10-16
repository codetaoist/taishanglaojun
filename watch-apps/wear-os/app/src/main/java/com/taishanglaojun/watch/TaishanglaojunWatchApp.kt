package com.taishanglaojun.watch

import android.app.Application
import android.content.Context
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.platform.LocalContext
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavHostController
import androidx.wear.compose.material.MaterialTheme
import androidx.wear.compose.material.Scaffold
import androidx.wear.compose.material.TimeText
import androidx.wear.compose.material.Vignette
import androidx.wear.compose.material.VignettePosition
import androidx.wear.compose.navigation.SwipeDismissableNavHost
import androidx.wear.compose.navigation.composable
import androidx.wear.compose.navigation.rememberSwipeDismissableNavController
import com.taishanglaojun.watch.data.repository.TaskRepository
import com.taishanglaojun.watch.services.ConnectivityService
import com.taishanglaojun.watch.ui.screens.MainScreen
import com.taishanglaojun.watch.ui.screens.QuickActionsScreen
import com.taishanglaojun.watch.ui.screens.SettingsScreen
import com.taishanglaojun.watch.ui.screens.TaskDetailScreen
import com.taishanglaojun.watch.ui.screens.TaskListScreen
import com.taishanglaojun.watch.ui.theme.TaishanglaojunWatchTheme
import com.taishanglaojun.watch.ui.viewmodels.MainViewModel
import dagger.hilt.android.AndroidEntryPoint
import dagger.hilt.android.HiltAndroidApp
import javax.inject.Inject

/**
 * 太上老君手表应用主入口
 */
@HiltAndroidApp
class TaishanglaojunWatchApplication : Application() {
    
    @Inject
    lateinit var connectivityService: ConnectivityService
    
    @Inject
    lateinit var taskRepository: TaskRepository
    
    override fun onCreate() {
        super.onCreate()
        
        // 初始化连接服务
        connectivityService.initialize()
        
        // 启动后台服务
        startBackgroundServices()
    }
    
    private fun startBackgroundServices() {
        // 启动数据同步服务
        connectivityService.startDataSync()
        
        // 初始化任务缓存
        taskRepository.initializeCache()
    }
}

/**
 * 主应用组合函数
 */
@Composable
fun TaishanglaojunWatchApp(
    navController: NavHostController = rememberSwipeDismissableNavController()
) {
    val context = LocalContext.current
    val mainViewModel: MainViewModel = hiltViewModel()
    
    // 监听应用状态
    val appState by mainViewModel.appState.collectAsState()
    
    // 应用启动时的初始化
    LaunchedEffect(Unit) {
        mainViewModel.initializeApp()
    }
    
    TaishanglaojunWatchTheme {
        Scaffold(
            timeText = {
                TimeText()
            },
            vignette = {
                Vignette(vignettePosition = VignettePosition.TopAndBottom)
            }
        ) {
            SwipeDismissableNavHost(
                navController = navController,
                startDestination = "main"
            ) {
                // 主屏幕
                composable("main") {
                    MainScreen(
                        navController = navController,
                        viewModel = mainViewModel
                    )
                }
                
                // 任务列表
                composable("tasks") {
                    TaskListScreen(
                        navController = navController,
                        onTaskClick = { taskId ->
                            navController.navigate("task_detail/$taskId")
                        }
                    )
                }
                
                // 任务详情
                composable("task_detail/{taskId}") { backStackEntry ->
                    val taskId = backStackEntry.arguments?.getString("taskId") ?: ""
                    TaskDetailScreen(
                        taskId = taskId,
                        navController = navController
                    )
                }
                
                // 快速操作
                composable("quick_actions") {
                    QuickActionsScreen(
                        navController = navController
                    )
                }
                
                // 设置页面
                composable("settings") {
                    SettingsScreen(
                        navController = navController
                    )
                }
            }
        }
    }
}

/**
 * 应用路由定义
 */
object WatchRoutes {
    const val MAIN = "main"
    const val TASKS = "tasks"
    const val TASK_DETAIL = "task_detail/{taskId}"
    const val QUICK_ACTIONS = "quick_actions"
    const val SETTINGS = "settings"
    
    fun taskDetail(taskId: String) = "task_detail/$taskId"
}

/**
 * 应用状态数据类
 */
data class AppState(
    val isInitialized: Boolean = false,
    val isConnected: Boolean = false,
    val lastSyncTime: Long? = null,
    val errorMessage: String? = null,
    val isLoading: Boolean = false
)

/**
 * 应用配置常量
 */
object AppConfig {
    const val APP_NAME = "太上老君手表"
    const val VERSION = "1.0.0"
    const val API_BASE_URL = "https://api.taishanglaojun.com"
    const val SYNC_INTERVAL_MS = 300_000L // 5分钟
    const val HEARTBEAT_INTERVAL_MS = 30_000L // 30秒
    const val MAX_CACHED_TASKS = 50
    const val CONNECTION_TIMEOUT_MS = 10_000L
    const val READ_TIMEOUT_MS = 30_000L
}

/**
 * 应用权限定义
 */
object AppPermissions {
    const val BODY_SENSORS = android.Manifest.permission.BODY_SENSORS
    const val ACCESS_FINE_LOCATION = android.Manifest.permission.ACCESS_FINE_LOCATION
    const val ACCESS_COARSE_LOCATION = android.Manifest.permission.ACCESS_COARSE_LOCATION
    const val WAKE_LOCK = android.Manifest.permission.WAKE_LOCK
    const val VIBRATE = android.Manifest.permission.VIBRATE
    const val INTERNET = android.Manifest.permission.INTERNET
    const val ACCESS_NETWORK_STATE = android.Manifest.permission.ACCESS_NETWORK_STATE
    
    val REQUIRED_PERMISSIONS = arrayOf(
        BODY_SENSORS,
        ACCESS_FINE_LOCATION,
        ACCESS_COARSE_LOCATION,
        WAKE_LOCK,
        VIBRATE,
        INTERNET,
        ACCESS_NETWORK_STATE
    )
}

/**
 * 应用通知渠道
 */
object NotificationChannels {
    const val TASK_UPDATES = "task_updates"
    const val SYNC_STATUS = "sync_status"
    const val SYSTEM_ALERTS = "system_alerts"
    
    data class ChannelInfo(
        val id: String,
        val name: String,
        val description: String,
        val importance: Int
    )
    
    val CHANNELS = listOf(
        ChannelInfo(
            TASK_UPDATES,
            "任务更新",
            "任务状态变更和新任务通知",
            android.app.NotificationManager.IMPORTANCE_DEFAULT
        ),
        ChannelInfo(
            SYNC_STATUS,
            "同步状态",
            "数据同步状态通知",
            android.app.NotificationManager.IMPORTANCE_LOW
        ),
        ChannelInfo(
            SYSTEM_ALERTS,
            "系统提醒",
            "系统错误和重要提醒",
            android.app.NotificationManager.IMPORTANCE_HIGH
        )
    )
}

/**
 * 应用主题配置
 */
object ThemeConfig {
    // 主色调
    val PRIMARY_COLOR = androidx.compose.ui.graphics.Color(0xFF1976D2)
    val PRIMARY_VARIANT_COLOR = androidx.compose.ui.graphics.Color(0xFF1565C0)
    val SECONDARY_COLOR = androidx.compose.ui.graphics.Color(0xFF03DAC6)
    
    // 状态颜色
    val SUCCESS_COLOR = androidx.compose.ui.graphics.Color(0xFF4CAF50)
    val WARNING_COLOR = androidx.compose.ui.graphics.Color(0xFFFF9800)
    val ERROR_COLOR = androidx.compose.ui.graphics.Color(0xFFF44336)
    val INFO_COLOR = androidx.compose.ui.graphics.Color(0xFF2196F3)
    
    // 任务优先级颜色
    val PRIORITY_LOW = androidx.compose.ui.graphics.Color(0xFF4CAF50)
    val PRIORITY_MEDIUM = androidx.compose.ui.graphics.Color(0xFFFF9800)
    val PRIORITY_HIGH = androidx.compose.ui.graphics.Color(0xFFF44336)
    val PRIORITY_URGENT = androidx.compose.ui.graphics.Color(0xFF9C27B0)
    
    // 坐标轴颜色
    val COORDINATE_S = androidx.compose.ui.graphics.Color(0xFFF44336) // 红色
    val COORDINATE_C = androidx.compose.ui.graphics.Color(0xFF4CAF50) // 绿色
    val COORDINATE_T = androidx.compose.ui.graphics.Color(0xFF2196F3) // 蓝色
}

/**
 * 应用日志标签
 */
object LogTags {
    const val APP = "TaishanglaojunWatch"
    const val CONNECTIVITY = "Connectivity"
    const val TASK_MANAGER = "TaskManager"
    const val DATA_SYNC = "DataSync"
    const val UI = "UI"
    const val LOCATION = "Location"
    const val HEALTH = "Health"
    const val NOTIFICATION = "Notification"
}