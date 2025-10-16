package com.taishanglaojun.watch.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.wear.compose.material.*
import androidx.wear.compose.navigation.SwipeDismissableNavHost
import androidx.wear.compose.navigation.composable
import androidx.wear.compose.navigation.rememberSwipeDismissableNavController
import com.taishanglaojun.watch.data.models.TaskStatistics
import com.taishanglaojun.watch.data.models.WatchTask
import com.taishanglaojun.watch.services.ConnectionState
import com.taishanglaojun.watch.ui.theme.TaishanglaojunWatchTheme
import com.taishanglaojun.watch.ui.viewmodels.MainViewModel
import kotlinx.datetime.Clock
import kotlinx.datetime.TimeZone
import kotlinx.datetime.toLocalDateTime

/**
 * 主界面
 * 显示任务概览、连接状态和快速操作
 */
@Composable
fun MainScreen(
    onNavigateToTaskList: () -> Unit,
    onNavigateToQuickActions: () -> Unit,
    onNavigateToSettings: () -> Unit,
    viewModel: MainViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val connectionState by viewModel.connectionState.collectAsStateWithLifecycle()
    val taskStatistics by viewModel.taskStatistics.collectAsStateWithLifecycle()
    val recentTasks by viewModel.recentTasks.collectAsStateWithLifecycle()
    
    LaunchedEffect(Unit) {
        viewModel.loadData()
    }
    
    TaishanglaojunWatchTheme {
        Scaffold(
            timeText = {
                TimeText(
                    timeTextStyle = TimeTextDefaults.timeTextStyle(
                        color = MaterialTheme.colors.onBackground
                    )
                )
            },
            vignette = {
                Vignette(vignettePosition = VignettePosition.TopAndBottom)
            }
        ) {
            LazyColumn(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(horizontal = 8.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp),
                contentPadding = PaddingValues(vertical = 16.dp)
            ) {
                // 连接状态卡片
                item {
                    ConnectionStatusCard(
                        connectionState = connectionState,
                        lastSyncTime = uiState.lastSyncTime,
                        onSyncClick = { viewModel.forceSync() }
                    )
                }
                
                // 任务统计卡片
                item {
                    TaskStatisticsCard(
                        statistics = taskStatistics,
                        onViewAllClick = onNavigateToTaskList
                    )
                }
                
                // 最近任务列表
                if (recentTasks.isNotEmpty()) {
                    item {
                        Text(
                            text = "最近任务",
                            style = MaterialTheme.typography.title3,
                            color = MaterialTheme.colors.onBackground,
                            modifier = Modifier.padding(horizontal = 8.dp)
                        )
                    }
                    
                    items(recentTasks.take(3)) { task ->
                        RecentTaskCard(
                            task = task,
                            onClick = { /* 导航到任务详情 */ }
                        )
                    }
                }
                
                // 快速操作按钮
                item {
                    QuickActionButtons(
                        onTaskListClick = onNavigateToTaskList,
                        onQuickActionsClick = onNavigateToQuickActions,
                        onSettingsClick = onNavigateToSettings,
                        onSyncClick = { viewModel.forceSync() }
                    )
                }
                
                // 加载状态
                if (uiState.isLoading) {
                    item {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(16.dp),
                            contentAlignment = Alignment.Center
                        ) {
                            CircularProgressIndicator(
                                modifier = Modifier.size(24.dp),
                                strokeWidth = 2.dp
                            )
                        }
                    }
                }
                
                // 错误状态
                uiState.error?.let { error ->
                    item {
                        ErrorCard(
                            error = error,
                            onRetryClick = { viewModel.loadData() }
                        )
                    }
                }
            }
        }
    }
}

/**
 * 连接状态卡片
 */
@Composable
private fun ConnectionStatusCard(
    connectionState: ConnectionState,
    lastSyncTime: kotlinx.datetime.Instant?,
    onSyncClick: () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        onClick = onSyncClick
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            Column(
                modifier = Modifier.weight(1f)
            ) {
                Row(
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    ConnectionStatusIndicator(connectionState)
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = connectionState.displayName,
                        style = MaterialTheme.typography.body2,
                        fontWeight = FontWeight.Medium
                    )
                }
                
                lastSyncTime?.let { time ->
                    val localTime = time.toLocalDateTime(TimeZone.currentSystemDefault())
                    Text(
                        text = "上次同步: ${localTime.hour}:${localTime.minute.toString().padStart(2, '0')}",
                        style = MaterialTheme.typography.caption2,
                        color = MaterialTheme.colors.onSurfaceVariant
                    )
                }
            }
            
            Icon(
                imageVector = Icons.Default.Refresh,
                contentDescription = "同步",
                tint = MaterialTheme.colors.primary,
                modifier = Modifier.size(20.dp)
            )
        }
    }
}

/**
 * 连接状态指示器
 */
@Composable
private fun ConnectionStatusIndicator(connectionState: ConnectionState) {
    val color = when (connectionState) {
        ConnectionState.CONNECTED -> Color.Green
        ConnectionState.PAIRED -> Color.Yellow
        ConnectionState.DISCONNECTED -> Color.Red
    }
    
    Box(
        modifier = Modifier
            .size(8.dp)
            .clip(CircleShape)
            .background(color)
    )
}

/**
 * 任务统计卡片
 */
@Composable
private fun TaskStatisticsCard(
    statistics: TaskStatistics,
    onViewAllClick: () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        onClick = onViewAllClick
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(12.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "任务统计",
                    style = MaterialTheme.typography.body2,
                    fontWeight = FontWeight.Medium
                )
                
                Icon(
                    imageVector = Icons.Default.ArrowForward,
                    contentDescription = "查看全部",
                    tint = MaterialTheme.colors.primary,
                    modifier = Modifier.size(16.dp)
                )
            }
            
            Spacer(modifier = Modifier.height(8.dp))
            
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceEvenly
            ) {
                StatisticItem(
                    label = "总计",
                    value = statistics.total.toString(),
                    color = MaterialTheme.colors.onSurface
                )
                
                StatisticItem(
                    label = "进行中",
                    value = statistics.active.toString(),
                    color = Color.Blue
                )
                
                StatisticItem(
                    label = "已完成",
                    value = statistics.completed.toString(),
                    color = Color.Green
                )
                
                if (statistics.overdue > 0) {
                    StatisticItem(
                        label = "逾期",
                        value = statistics.overdue.toString(),
                        color = Color.Red
                    )
                }
            }
        }
    }
}

/**
 * 统计项目
 */
@Composable
private fun StatisticItem(
    label: String,
    value: String,
    color: Color
) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Text(
            text = value,
            style = MaterialTheme.typography.title3,
            color = color,
            fontWeight = FontWeight.Bold
        )
        Text(
            text = label,
            style = MaterialTheme.typography.caption2,
            color = MaterialTheme.colors.onSurfaceVariant
        )
    }
}

/**
 * 最近任务卡片
 */
@Composable
private fun RecentTaskCard(
    task: WatchTask,
    onClick: () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        onClick = onClick
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(12.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            // 任务状态指示器
            TaskStatusIndicator(
                status = task.status,
                modifier = Modifier.size(12.dp)
            )
            
            Spacer(modifier = Modifier.width(8.dp))
            
            Column(
                modifier = Modifier.weight(1f)
            ) {
                Text(
                    text = task.title,
                    style = MaterialTheme.typography.body2,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
                
                if (task.progress > 0) {
                    Text(
                        text = "进度: ${(task.progress * 100).toInt()}%",
                        style = MaterialTheme.typography.caption2,
                        color = MaterialTheme.colors.onSurfaceVariant
                    )
                }
            }
            
            // 优先级指示器
            if (task.priority.ordinal > 0) {
                PriorityIndicator(
                    priority = task.priority,
                    modifier = Modifier.size(16.dp)
                )
            }
        }
    }
}

/**
 * 任务状态指示器
 */
@Composable
private fun TaskStatusIndicator(
    status: com.taishanglaojun.watch.data.models.TaskStatus,
    modifier: Modifier = Modifier
) {
    val color = when (status) {
        com.taishanglaojun.watch.data.models.TaskStatus.PENDING -> Color.Gray
        com.taishanglaojun.watch.data.models.TaskStatus.IN_PROGRESS -> Color.Blue
        com.taishanglaojun.watch.data.models.TaskStatus.COMPLETED -> Color.Green
        com.taishanglaojun.watch.data.models.TaskStatus.CANCELLED -> Color.Red
    }
    
    Box(
        modifier = modifier
            .clip(CircleShape)
            .background(color)
    )
}

/**
 * 优先级指示器
 */
@Composable
private fun PriorityIndicator(
    priority: com.taishanglaojun.watch.data.models.TaskPriority,
    modifier: Modifier = Modifier
) {
    val (icon, color) = when (priority) {
        com.taishanglaojun.watch.data.models.TaskPriority.LOW -> Icons.Default.KeyboardArrowDown to Color.Green
        com.taishanglaojun.watch.data.models.TaskPriority.MEDIUM -> Icons.Default.Remove to Color.Yellow
        com.taishanglaojun.watch.data.models.TaskPriority.HIGH -> Icons.Default.KeyboardArrowUp to Color.Red
    }
    
    Icon(
        imageVector = icon,
        contentDescription = priority.name,
        tint = color,
        modifier = modifier
    )
}

/**
 * 快速操作按钮
 */
@Composable
private fun QuickActionButtons(
    onTaskListClick: () -> Unit,
    onQuickActionsClick: () -> Unit,
    onSettingsClick: () -> Unit,
    onSyncClick: () -> Unit
) {
    Column(
        modifier = Modifier.fillMaxWidth(),
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        Text(
            text = "快速操作",
            style = MaterialTheme.typography.title3,
            color = MaterialTheme.colors.onBackground,
            modifier = Modifier.padding(horizontal = 8.dp)
        )
        
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            QuickActionButton(
                icon = Icons.Default.List,
                label = "任务",
                onClick = onTaskListClick,
                modifier = Modifier.weight(1f)
            )
            
            QuickActionButton(
                icon = Icons.Default.PlayArrow,
                label = "操作",
                onClick = onQuickActionsClick,
                modifier = Modifier.weight(1f)
            )
        }
        
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            QuickActionButton(
                icon = Icons.Default.Refresh,
                label = "同步",
                onClick = onSyncClick,
                modifier = Modifier.weight(1f)
            )
            
            QuickActionButton(
                icon = Icons.Default.Settings,
                label = "设置",
                onClick = onSettingsClick,
                modifier = Modifier.weight(1f)
            )
        }
    }
}

/**
 * 快速操作按钮
 */
@Composable
private fun QuickActionButton(
    icon: ImageVector,
    label: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    Button(
        onClick = onClick,
        modifier = modifier.height(48.dp),
        colors = ButtonDefaults.buttonColors(
            backgroundColor = MaterialTheme.colors.surface
        )
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Icon(
                imageVector = icon,
                contentDescription = label,
                modifier = Modifier.size(16.dp)
            )
            Text(
                text = label,
                style = MaterialTheme.typography.caption2,
                textAlign = TextAlign.Center,
                maxLines = 1
            )
        }
    }
}

/**
 * 错误卡片
 */
@Composable
private fun ErrorCard(
    error: String,
    onRetryClick: () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        backgroundColor = MaterialTheme.colors.error.copy(alpha = 0.1f)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(12.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Icon(
                imageVector = Icons.Default.Warning,
                contentDescription = "错误",
                tint = MaterialTheme.colors.error,
                modifier = Modifier.size(24.dp)
            )
            
            Spacer(modifier = Modifier.height(8.dp))
            
            Text(
                text = error,
                style = MaterialTheme.typography.body2,
                color = MaterialTheme.colors.error,
                textAlign = TextAlign.Center
            )
            
            Spacer(modifier = Modifier.height(8.dp))
            
            Button(
                onClick = onRetryClick,
                colors = ButtonDefaults.buttonColors(
                    backgroundColor = MaterialTheme.colors.error
                )
            ) {
                Text(
                    text = "重试",
                    color = MaterialTheme.colors.onError
                )
            }
        }
    }
}