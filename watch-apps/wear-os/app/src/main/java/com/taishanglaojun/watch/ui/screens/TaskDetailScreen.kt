package com.taishanglaojun.watch.ui.screens

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.wear.compose.material.*
import com.taishanglaojun.watch.data.model.*
import com.taishanglaojun.watch.ui.theme.TaishanglaojunWatchTheme
import com.taishanglaojun.watch.ui.viewmodel.TaskDetailViewModel
import kotlinx.coroutines.launch
import java.text.SimpleDateFormat
import java.util.*

/**
 * 任务详情屏幕
 * 显示任务的详细信息，支持任务操作和进度更新
 */
@Composable
fun TaskDetailScreen(
    taskId: String,
    onNavigateBack: () -> Unit,
    viewModel: TaskDetailViewModel = hiltViewModel()
) {
    val task by viewModel.task.collectAsState()
    val isLoading by viewModel.isLoading.collectAsState()
    val error by viewModel.error.collectAsState()
    val isUpdating by viewModel.isUpdating.collectAsState()
    
    val scrollState = rememberScrollState()
    val coroutineScope = rememberCoroutineScope()
    
    LaunchedEffect(taskId) {
        viewModel.loadTask(taskId)
    }
    
    TaishanglaojunWatchTheme {
        Scaffold(
            timeText = { TimeText() },
            vignette = { Vignette(vignettePosition = VignettePosition.TopAndBottom) }
        ) {
            Box(
                modifier = Modifier.fillMaxSize()
            ) {
                when {
                    isLoading -> {
                        LoadingIndicator()
                    }
                    error != null -> {
                        ErrorCard(
                            error = error!!,
                            onRetry = { viewModel.loadTask(taskId) },
                            onBack = onNavigateBack
                        )
                    }
                    task != null -> {
                        TaskDetailContent(
                            task = task!!,
                            isUpdating = isUpdating,
                            scrollState = scrollState,
                            onAction = { action ->
                                coroutineScope.launch {
                                    when (action) {
                                        TaskDetailAction.Accept -> viewModel.acceptTask()
                                        TaskDetailAction.Start -> viewModel.startTask()
                                        TaskDetailAction.Complete -> viewModel.completeTask()
                                        TaskDetailAction.Pause -> viewModel.pauseTask()
                                        TaskDetailAction.Cancel -> viewModel.cancelTask()
                                        is TaskDetailAction.UpdateProgress -> viewModel.updateProgress(action.progress)
                                    }
                                }
                            },
                            onBack = onNavigateBack
                        )
                    }
                }
            }
        }
    }
}

/**
 * 任务详情内容
 */
@Composable
private fun TaskDetailContent(
    task: WatchTask,
    isUpdating: Boolean,
    scrollState: ScrollState,
    onAction: (TaskDetailAction) -> Unit,
    onBack: () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(scrollState)
            .padding(
                top = 32.dp,
                start = 8.dp,
                end = 8.dp,
                bottom = 32.dp
            ),
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        // 任务头部
        TaskDetailHeader(
            task = task,
            onBack = onBack
        )
        
        // 任务状态卡片
        TaskStatusCard(
            task = task,
            isUpdating = isUpdating
        )
        
        // 任务信息卡片
        TaskInfoCard(task = task)
        
        // 进度卡片（如果任务进行中）
        if (task.status == TaskStatus.IN_PROGRESS) {
            TaskProgressCard(
                task = task,
                onProgressUpdate = { progress ->
                    onAction(TaskDetailAction.UpdateProgress(progress))
                }
            )
        }
        
        // 位置信息卡片（如果有坐标）
        if (task.coordinate != null) {
            TaskLocationCard(
                coordinate = task.coordinate!!
            )
        }
        
        // 操作按钮
        TaskActionButtons(
            task = task,
            isUpdating = isUpdating,
            onAction = onAction
        )
    }
}

/**
 * 任务详情头部
 */
@Composable
private fun TaskDetailHeader(
    task: WatchTask,
    onBack: () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(12.dp)
    ) {
        Column(
            modifier = Modifier.padding(16.dp)
        ) {
            // 返回按钮和标题
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.Top
            ) {
                Button(
                    onClick = onBack,
                    modifier = Modifier.size(32.dp),
                    colors = ButtonDefaults.buttonColors(
                        backgroundColor = MaterialTheme.colors.surface
                    )
                ) {
                    Icon(
                        imageVector = Icons.Default.ArrowBack,
                        contentDescription = "返回",
                        modifier = Modifier.size(16.dp)
                    )
                }
                
                Column(
                    modifier = Modifier.weight(1f).padding(horizontal = 12.dp),
                    horizontalAlignment = Alignment.CenterHorizontally
                ) {
                    Text(
                        text = task.title,
                        style = MaterialTheme.typography.title3,
                        fontWeight = FontWeight.Bold,
                        textAlign = TextAlign.Center,
                        maxLines = 2,
                        overflow = TextOverflow.Ellipsis
                    )
                    
                    if (task.description.isNotEmpty()) {
                        Spacer(modifier = Modifier.height(4.dp))
                        Text(
                            text = task.description,
                            style = MaterialTheme.typography.caption1,
                            color = MaterialTheme.colors.onSurface.copy(alpha = 0.7f),
                            textAlign = TextAlign.Center,
                            maxLines = 3,
                            overflow = TextOverflow.Ellipsis
                        )
                    }
                }
                
                // 任务ID显示
                Text(
                    text = "#${task.id.takeLast(4)}",
                    style = MaterialTheme.typography.caption2,
                    color = MaterialTheme.colors.onSurface.copy(alpha = 0.5f)
                )
            }
        }
    }
}

/**
 * 任务状态卡片
 */
@Composable
private fun TaskStatusCard(
    task: WatchTask,
    isUpdating: Boolean
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(12.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            // 状态指示器
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                TaskStatusIndicator(
                    status = task.status,
                    modifier = Modifier.size(12.dp)
                )
                
                Text(
                    text = task.status.displayName,
                    style = MaterialTheme.typography.body2,
                    fontWeight = FontWeight.Medium
                )
                
                if (isUpdating) {
                    CircularProgressIndicator(
                        modifier = Modifier.size(12.dp),
                        strokeWidth = 1.dp,
                        color = MaterialTheme.colors.primary
                    )
                }
            }
            
            // 优先级指示器
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(4.dp)
            ) {
                PriorityIndicator(
                    priority = task.priority,
                    size = 16.dp
                )
                
                Text(
                    text = task.priority.displayName,
                    style = MaterialTheme.typography.caption1,
                    color = task.priority.color
                )
            }
        }
    }
}

/**
 * 任务信息卡片
 */
@Composable
private fun TaskInfoCard(task: WatchTask) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(12.dp)
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            Text(
                text = "任务信息",
                style = MaterialTheme.typography.title3,
                fontWeight = FontWeight.Bold
            )
            
            // 难度和奖励
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                TaskInfoItem(
                    label = "难度",
                    value = {
                        DifficultyStars(
                            difficulty = task.difficulty,
                            size = 12.dp
                        )
                    }
                )
                
                if (task.reward > 0) {
                    TaskInfoItem(
                        label = "奖励",
                        value = {
                            Text(
                                text = task.formattedReward,
                                style = MaterialTheme.typography.body2,
                                color = Color(0xFFFF9800),
                                fontWeight = FontWeight.Medium
                            )
                        }
                    )
                }
            }
            
            // 创建时间和截止时间
            if (task.createdAt != null) {
                TaskInfoItem(
                    label = "创建时间",
                    value = {
                        Text(
                            text = formatDateTime(task.createdAt!!),
                            style = MaterialTheme.typography.caption1,
                            color = MaterialTheme.colors.onSurface.copy(alpha = 0.7f)
                        )
                    }
                )
            }
            
            if (task.dueDate != null) {
                TaskInfoItem(
                    label = "截止时间",
                    value = {
                        Text(
                            text = formatDateTime(task.dueDate!!),
                            style = MaterialTheme.typography.caption1,
                            color = if (task.isOverdue) Color(0xFFF44336) else MaterialTheme.colors.onSurface.copy(alpha = 0.7f)
                        )
                    }
                )
            }
            
            // 连接要求
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                TaskInfoItem(
                    label = "需要连接",
                    value = {
                        Icon(
                            imageVector = if (task.requiresPhoneConnection) Icons.Default.Wifi else Icons.Default.WifiOff,
                            contentDescription = null,
                            modifier = Modifier.size(16.dp),
                            tint = if (task.requiresPhoneConnection) Color(0xFF2196F3) else Color(0xFF9E9E9E)
                        )
                    }
                )
                
                TaskInfoItem(
                    label = "离线完成",
                    value = {
                        Icon(
                            imageVector = if (task.canCompleteOffline) Icons.Default.Check else Icons.Default.Close,
                            contentDescription = null,
                            modifier = Modifier.size(16.dp),
                            tint = if (task.canCompleteOffline) Color(0xFF4CAF50) else Color(0xFFF44336)
                        )
                    }
                )
            }
        }
    }
}

/**
 * 任务信息项
 */
@Composable
private fun TaskInfoItem(
    label: String,
    value: @Composable () -> Unit
) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Text(
            text = label,
            style = MaterialTheme.typography.caption2,
            color = MaterialTheme.colors.onSurface.copy(alpha = 0.7f)
        )
        Spacer(modifier = Modifier.height(4.dp))
        value()
    }
}

/**
 * 任务进度卡片
 */
@Composable
private fun TaskProgressCard(
    task: WatchTask,
    onProgressUpdate: (Float) -> Unit
) {
    var showProgressDialog by remember { mutableStateOf(false) }
    
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(12.dp)
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "任务进度",
                    style = MaterialTheme.typography.title3,
                    fontWeight = FontWeight.Bold
                )
                
                Button(
                    onClick = { showProgressDialog = true },
                    modifier = Modifier.size(32.dp),
                    colors = ButtonDefaults.buttonColors(
                        backgroundColor = MaterialTheme.colors.primary
                    )
                ) {
                    Icon(
                        imageVector = Icons.Default.Edit,
                        contentDescription = "更新进度",
                        modifier = Modifier.size(16.dp),
                        tint = Color.White
                    )
                }
            }
            
            // 进度条
            Column {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    Text(
                        text = "${(task.progress * 100).toInt()}%",
                        style = MaterialTheme.typography.body1,
                        fontWeight = FontWeight.Bold,
                        color = MaterialTheme.colors.primary
                    )
                    
                    Text(
                        text = "剩余 ${task.timeRemaining}",
                        style = MaterialTheme.typography.caption1,
                        color = MaterialTheme.colors.onSurface.copy(alpha = 0.7f)
                    )
                }
                
                Spacer(modifier = Modifier.height(8.dp))
                
                LinearProgressIndicator(
                    progress = task.progress,
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(6.dp)
                        .clip(RoundedCornerShape(3.dp)),
                    color = MaterialTheme.colors.primary,
                    backgroundColor = MaterialTheme.colors.surface
                )
            }
        }
    }
    
    // 进度更新对话框
    if (showProgressDialog) {
        ProgressUpdateDialog(
            currentProgress = task.progress,
            onProgressUpdate = { progress ->
                onProgressUpdate(progress)
                showProgressDialog = false
            },
            onDismiss = { showProgressDialog = false }
        )
    }
}

/**
 * 任务位置卡片
 */
@Composable
private fun TaskLocationCard(
    coordinate: TaskCoordinate
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(12.dp)
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            Text(
                text = "位置信息",
                style = MaterialTheme.typography.title3,
                fontWeight = FontWeight.Bold
            )
            
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Column {
                    Text(
                        text = "纬度",
                        style = MaterialTheme.typography.caption2,
                        color = MaterialTheme.colors.onSurface.copy(alpha = 0.7f)
                    )
                    Text(
                        text = String.format("%.6f", coordinate.latitude),
                        style = MaterialTheme.typography.caption1,
                        fontWeight = FontWeight.Medium
                    )
                }
                
                Column {
                    Text(
                        text = "经度",
                        style = MaterialTheme.typography.caption2,
                        color = MaterialTheme.colors.onSurface.copy(alpha = 0.7f)
                    )
                    Text(
                        text = String.format("%.6f", coordinate.longitude),
                        style = MaterialTheme.typography.caption1,
                        fontWeight = FontWeight.Medium
                    )
                }
            }
            
            if (coordinate.address.isNotEmpty()) {
                Column {
                    Text(
                        text = "地址",
                        style = MaterialTheme.typography.caption2,
                        color = MaterialTheme.colors.onSurface.copy(alpha = 0.7f)
                    )
                    Text(
                        text = coordinate.address,
                        style = MaterialTheme.typography.caption1,
                        maxLines = 2,
                        overflow = TextOverflow.Ellipsis
                    )
                }
            }
        }
    }
}

/**
 * 任务操作按钮
 */
@Composable
private fun TaskActionButtons(
    task: WatchTask,
    isUpdating: Boolean,
    onAction: (TaskDetailAction) -> Unit
) {
    val availableActions = getAvailableActions(task)
    
    if (availableActions.isNotEmpty()) {
        Card(
            modifier = Modifier.fillMaxWidth(),
            shape = RoundedCornerShape(12.dp)
        ) {
            Column(
                modifier = Modifier.padding(16.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                Text(
                    text = "操作",
                    style = MaterialTheme.typography.title3,
                    fontWeight = FontWeight.Bold
                )
                
                availableActions.forEach { action ->
                    Button(
                        onClick = { onAction(action) },
                        modifier = Modifier.fillMaxWidth(),
                        enabled = !isUpdating,
                        colors = ButtonDefaults.buttonColors(
                            backgroundColor = action.color
                        )
                    ) {
                        if (isUpdating) {
                            CircularProgressIndicator(
                                modifier = Modifier.size(16.dp),
                                strokeWidth = 2.dp,
                                color = Color.White
                            )
                        } else {
                            Icon(
                                imageVector = action.icon,
                                contentDescription = action.label,
                                modifier = Modifier.size(16.dp),
                                tint = Color.White
                            )
                        }
                        
                        Spacer(modifier = Modifier.width(8.dp))
                        
                        Text(
                            text = action.label,
                            style = MaterialTheme.typography.button,
                            color = Color.White
                        )
                    }
                }
            }
        }
    }
}

/**
 * 进度更新对话框
 */
@Composable
private fun ProgressUpdateDialog(
    currentProgress: Float,
    onProgressUpdate: (Float) -> Unit,
    onDismiss: () -> Unit
) {
    var progress by remember { mutableStateOf(currentProgress) }
    
    Dialog(
        showDialog = true,
        onDismissRequest = onDismiss
    ) {
        Column(
            verticalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            Text(
                text = "更新进度",
                style = MaterialTheme.typography.title3,
                fontWeight = FontWeight.Bold
            )
            
            Text(
                text = "${(progress * 100).toInt()}%",
                style = MaterialTheme.typography.title2,
                fontWeight = FontWeight.Bold,
                color = MaterialTheme.colors.primary,
                modifier = Modifier.fillMaxWidth(),
                textAlign = TextAlign.Center
            )
            
            // 进度选择按钮
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(4.dp)
            ) {
                listOf(0.25f, 0.5f, 0.75f, 1.0f).forEach { value ->
                    Button(
                        onClick = { progress = value },
                        modifier = Modifier.weight(1f),
                        colors = ButtonDefaults.buttonColors(
                            backgroundColor = if (progress == value) {
                                MaterialTheme.colors.primary
                            } else {
                                MaterialTheme.colors.surface
                            }
                        )
                    ) {
                        Text(
                            text = "${(value * 100).toInt()}%",
                            style = MaterialTheme.typography.caption2
                        )
                    }
                }
            }
            
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                Button(
                    onClick = onDismiss,
                    modifier = Modifier.weight(1f),
                    colors = ButtonDefaults.buttonColors(
                        backgroundColor = MaterialTheme.colors.surface
                    )
                ) {
                    Text(
                        text = "取消",
                        style = MaterialTheme.typography.button
                    )
                }
                
                Button(
                    onClick = { onProgressUpdate(progress) },
                    modifier = Modifier.weight(1f),
                    colors = ButtonDefaults.buttonColors(
                        backgroundColor = MaterialTheme.colors.primary
                    )
                ) {
                    Text(
                        text = "确定",
                        style = MaterialTheme.typography.button,
                        color = Color.White
                    )
                }
            }
        }
    }
}

// MARK: - Reusable Components

/**
 * 任务状态指示器
 */
@Composable
private fun TaskStatusIndicator(
    status: TaskStatus,
    modifier: Modifier = Modifier
) {
    val color = when (status) {
        TaskStatus.PENDING -> Color(0xFF9E9E9E)
        TaskStatus.IN_PROGRESS -> Color(0xFF2196F3)
        TaskStatus.COMPLETED -> Color(0xFF4CAF50)
        TaskStatus.PAUSED -> Color(0xFFFF9800)
        TaskStatus.CANCELLED -> Color(0xFFF44336)
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
    priority: TaskPriority,
    size: androidx.compose.ui.unit.Dp
) {
    val (color, icon) = when (priority) {
        TaskPriority.LOW -> Color(0xFF4CAF50) to Icons.Default.KeyboardArrowDown
        TaskPriority.MEDIUM -> Color(0xFFFF9800) to Icons.Default.Remove
        TaskPriority.HIGH -> Color(0xFFF44336) to Icons.Default.KeyboardArrowUp
        TaskPriority.URGENT -> Color(0xFF9C27B0) to Icons.Default.PriorityHigh
    }
    
    Icon(
        imageVector = icon,
        contentDescription = priority.name,
        modifier = Modifier.size(size),
        tint = color
    )
}

/**
 * 难度星级
 */
@Composable
private fun DifficultyStars(
    difficulty: Int,
    size: androidx.compose.ui.unit.Dp
) {
    Row(
        horizontalArrangement = Arrangement.spacedBy(1.dp)
    ) {
        repeat(5) { index ->
            Icon(
                imageVector = if (index < difficulty) Icons.Default.Star else Icons.Default.StarBorder,
                contentDescription = null,
                modifier = Modifier.size(size),
                tint = if (index < difficulty) Color(0xFFFFD700) else Color(0xFF9E9E9E)
            )
        }
    }
}

/**
 * 加载指示器
 */
@Composable
private fun LoadingIndicator() {
    Box(
        modifier = Modifier.fillMaxSize(),
        contentAlignment = Alignment.Center
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            CircularProgressIndicator(
                modifier = Modifier.size(32.dp),
                color = MaterialTheme.colors.primary
            )
            
            Spacer(modifier = Modifier.height(16.dp))
            
            Text(
                text = "加载中...",
                style = MaterialTheme.typography.body2,
                color = MaterialTheme.colors.onSurface.copy(alpha = 0.7f)
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
    onRetry: () -> Unit,
    onBack: () -> Unit
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp),
        shape = RoundedCornerShape(12.dp)
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Icon(
                imageVector = Icons.Default.Error,
                contentDescription = null,
                modifier = Modifier.size(32.dp),
                tint = Color(0xFFF44336)
            )
            
            Spacer(modifier = Modifier.height(8.dp))
            
            Text(
                text = "加载失败",
                style = MaterialTheme.typography.body1,
                fontWeight = FontWeight.Medium
            )
            
            Spacer(modifier = Modifier.height(4.dp))
            
            Text(
                text = error,
                style = MaterialTheme.typography.caption1,
                textAlign = TextAlign.Center,
                color = MaterialTheme.colors.onSurface.copy(alpha = 0.7f)
            )
            
            Spacer(modifier = Modifier.height(12.dp))
            
            Row(
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                Button(
                    onClick = onBack,
                    colors = ButtonDefaults.buttonColors(
                        backgroundColor = MaterialTheme.colors.surface
                    )
                ) {
                    Text(
                        text = "返回",
                        style = MaterialTheme.typography.button
                    )
                }
                
                Button(
                    onClick = onRetry,
                    colors = ButtonDefaults.buttonColors(
                        backgroundColor = MaterialTheme.colors.primary
                    )
                ) {
                    Text(
                        text = "重试",
                        style = MaterialTheme.typography.button,
                        color = Color.White
                    )
                }
            }
        }
    }
}

// MARK: - Helper Functions

/**
 * 格式化日期时间
 */
private fun formatDateTime(date: Date): String {
    val formatter = SimpleDateFormat("MM-dd HH:mm", Locale.getDefault())
    return formatter.format(date)
}

/**
 * 获取可用的任务操作
 */
private fun getAvailableActions(task: WatchTask): List<TaskDetailAction> {
    return when (task.status) {
        TaskStatus.PENDING -> listOf(TaskDetailAction.Accept)
        TaskStatus.IN_PROGRESS -> listOf(TaskDetailAction.Complete, TaskDetailAction.Pause)
        TaskStatus.PAUSED -> listOf(TaskDetailAction.Start, TaskDetailAction.Cancel)
        else -> emptyList()
    }
}

// MARK: - Supporting Types

/**
 * 任务详情操作
 */
sealed class TaskDetailAction(
    val label: String,
    val icon: ImageVector,
    val color: Color
) {
    object Accept : TaskDetailAction("接受任务", Icons.Default.Check, Color(0xFF4CAF50))
    object Start : TaskDetailAction("开始任务", Icons.Default.PlayArrow, Color(0xFF2196F3))
    object Complete : TaskDetailAction("完成任务", Icons.Default.CheckCircle, Color(0xFF4CAF50))
    object Pause : TaskDetailAction("暂停任务", Icons.Default.Pause, Color(0xFFFF9800))
    object Cancel : TaskDetailAction("取消任务", Icons.Default.Cancel, Color(0xFFF44336))
    data class UpdateProgress(val progress: Float) : TaskDetailAction("更新进度", Icons.Default.Update, Color(0xFF2196F3))
}

/**
 * 任务状态扩展
 */
val TaskStatus.displayName: String
    get() = when (this) {
        TaskStatus.PENDING -> "待接受"
        TaskStatus.IN_PROGRESS -> "进行中"
        TaskStatus.COMPLETED -> "已完成"
        TaskStatus.PAUSED -> "已暂停"
        TaskStatus.CANCELLED -> "已取消"
    }

/**
 * 任务优先级扩展
 */
val TaskPriority.displayName: String
    get() = when (this) {
        TaskPriority.LOW -> "低"
        TaskPriority.MEDIUM -> "中"
        TaskPriority.HIGH -> "高"
        TaskPriority.URGENT -> "紧急"
    }

val TaskPriority.color: Color
    get() = when (this) {
        TaskPriority.LOW -> Color(0xFF4CAF50)
        TaskPriority.MEDIUM -> Color(0xFFFF9800)
        TaskPriority.HIGH -> Color(0xFFF44336)
        TaskPriority.URGENT -> Color(0xFF9C27B0)
    }