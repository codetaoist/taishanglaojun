package com.taishanglaojun.watch.ui.screens

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
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
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.wear.compose.material.*
import androidx.wear.compose.navigation.SwipeDismissableNavHost
import androidx.wear.compose.navigation.composable
import androidx.wear.compose.navigation.rememberSwipeDismissableNavController
import com.taishanglaojun.watch.data.model.*
import com.taishanglaojun.watch.ui.theme.TaishanglaojunWatchTheme
import com.taishanglaojun.watch.ui.viewmodel.TaskListViewModel
import kotlinx.coroutines.launch

/**
 * 任务列表屏幕
 * 显示和管理任务列表，支持筛选、排序和快速操作
 */
@Composable
fun TaskListScreen(
    onNavigateToDetail: (String) -> Unit,
    onNavigateBack: () -> Unit,
    viewModel: TaskListViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val tasks by viewModel.tasks.collectAsState()
    val isLoading by viewModel.isLoading.collectAsState()
    val error by viewModel.error.collectAsState()
    
    val listState = rememberScalingLazyListState()
    val coroutineScope = rememberCoroutineScope()
    
    LaunchedEffect(Unit) {
        viewModel.loadTasks()
    }
    
    TaishanglaojunWatchTheme {
        Scaffold(
            timeText = { TimeText() },
            vignette = { Vignette(vignettePosition = VignettePosition.TopAndBottom) },
            positionIndicator = { PositionIndicator(scalingLazyListState = listState) }
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
                            onRetry = { viewModel.loadTasks() }
                        )
                    }
                    tasks.isEmpty() -> {
                        EmptyTasksView(
                            onRefresh = { viewModel.refreshTasks() }
                        )
                    }
                    else -> {
                        TaskListContent(
                            tasks = tasks,
                            uiState = uiState,
                            listState = listState,
                            onTaskClick = onNavigateToDetail,
                            onTaskAction = { task, action ->
                                when (action) {
                                    TaskAction.Accept -> viewModel.acceptTask(task.id)
                                    TaskAction.Start -> viewModel.startTask(task.id)
                                    TaskAction.Complete -> viewModel.completeTask(task.id)
                                    TaskAction.Pause -> viewModel.pauseTask(task.id)
                                }
                            },
                            onFilterChange = { filter ->
                                viewModel.updateFilter(filter)
                            },
                            onSortChange = { sort ->
                                viewModel.updateSort(sort)
                            },
                            onRefresh = {
                                coroutineScope.launch {
                                    viewModel.refreshTasks()
                                }
                            }
                        )
                    }
                }
            }
        }
    }
}

/**
 * 任务列表内容
 */
@Composable
private fun TaskListContent(
    tasks: List<WatchTask>,
    uiState: TaskListUiState,
    listState: ScalingLazyListState,
    onTaskClick: (String) -> Unit,
    onTaskAction: (WatchTask, TaskAction) -> Unit,
    onFilterChange: (TaskFilter) -> Unit,
    onSortChange: (TaskSort) -> Unit,
    onRefresh: () -> Unit
) {
    ScalingLazyColumn(
        state = listState,
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(
            top = 32.dp,
            start = 8.dp,
            end = 8.dp,
            bottom = 32.dp
        ),
        verticalArrangement = Arrangement.spacedBy(4.dp)
    ) {
        // 标题和统计
        item {
            TaskListHeader(
                totalTasks = tasks.size,
                activeTasks = tasks.count { it.status == TaskStatus.IN_PROGRESS },
                onRefresh = onRefresh
            )
        }
        
        // 筛选和排序控件
        item {
            TaskFilterControls(
                currentFilter = uiState.filter,
                currentSort = uiState.sort,
                onFilterChange = onFilterChange,
                onSortChange = onSortChange
            )
        }
        
        // 任务列表
        items(tasks) { task ->
            TaskListItem(
                task = task,
                onClick = { onTaskClick(task.id) },
                onAction = { action -> onTaskAction(task, action) }
            )
        }
    }
}

/**
 * 任务列表头部
 */
@Composable
private fun TaskListHeader(
    totalTasks: Int,
    activeTasks: Int,
    onRefresh: () -> Unit
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 4.dp),
        shape = RoundedCornerShape(12.dp)
    ) {
        Column(
            modifier = Modifier.padding(12.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Text(
                text = "任务列表",
                style = MaterialTheme.typography.title3,
                fontWeight = FontWeight.Bold
            )
            
            Spacer(modifier = Modifier.height(4.dp))
            
            Row(
                horizontalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                TaskStatItem(
                    label = "总计",
                    value = totalTasks.toString(),
                    color = MaterialTheme.colors.primary
                )
                
                TaskStatItem(
                    label = "进行中",
                    value = activeTasks.toString(),
                    color = Color(0xFF4CAF50)
                )
            }
            
            Spacer(modifier = Modifier.height(8.dp))
            
            Button(
                onClick = onRefresh,
                modifier = Modifier.size(32.dp),
                colors = ButtonDefaults.buttonColors(
                    backgroundColor = MaterialTheme.colors.surface
                )
            ) {
                Icon(
                    imageVector = Icons.Default.Refresh,
                    contentDescription = "刷新",
                    modifier = Modifier.size(16.dp)
                )
            }
        }
    }
}

/**
 * 任务统计项
 */
@Composable
private fun TaskStatItem(
    label: String,
    value: String,
    color: Color
) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Text(
            text = value,
            style = MaterialTheme.typography.title2,
            fontWeight = FontWeight.Bold,
            color = color
        )
        Text(
            text = label,
            style = MaterialTheme.typography.caption2,
            color = MaterialTheme.colors.onSurface.copy(alpha = 0.7f)
        )
    }
}

/**
 * 任务筛选控件
 */
@Composable
private fun TaskFilterControls(
    currentFilter: TaskFilter,
    currentSort: TaskSort,
    onFilterChange: (TaskFilter) -> Unit,
    onSortChange: (TaskSort) -> Unit
) {
    var showFilterDialog by remember { mutableStateOf(false) }
    var showSortDialog by remember { mutableStateOf(false) }
    
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 4.dp),
        horizontalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        // 筛选按钮
        Chip(
            onClick = { showFilterDialog = true },
            modifier = Modifier.weight(1f),
            colors = ChipDefaults.chipColors(
                backgroundColor = if (currentFilter != TaskFilter.ALL) {
                    MaterialTheme.colors.primary
                } else {
                    MaterialTheme.colors.surface
                }
            )
        ) {
            Icon(
                imageVector = Icons.Default.FilterList,
                contentDescription = "筛选",
                modifier = Modifier.size(12.dp)
            )
            Spacer(modifier = Modifier.width(4.dp))
            Text(
                text = currentFilter.displayName,
                style = MaterialTheme.typography.caption2,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis
            )
        }
        
        // 排序按钮
        Chip(
            onClick = { showSortDialog = true },
            modifier = Modifier.weight(1f),
            colors = ChipDefaults.chipColors(
                backgroundColor = MaterialTheme.colors.surface
            )
        ) {
            Icon(
                imageVector = Icons.Default.Sort,
                contentDescription = "排序",
                modifier = Modifier.size(12.dp)
            )
            Spacer(modifier = Modifier.width(4.dp))
            Text(
                text = currentSort.displayName,
                style = MaterialTheme.typography.caption2,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis
            )
        }
    }
    
    // 筛选对话框
    if (showFilterDialog) {
        FilterDialog(
            currentFilter = currentFilter,
            onFilterSelected = { filter ->
                onFilterChange(filter)
                showFilterDialog = false
            },
            onDismiss = { showFilterDialog = false }
        )
    }
    
    // 排序对话框
    if (showSortDialog) {
        SortDialog(
            currentSort = currentSort,
            onSortSelected = { sort ->
                onSortChange(sort)
                showSortDialog = false
            },
            onDismiss = { showSortDialog = false }
        )
    }
}

/**
 * 任务列表项
 */
@Composable
private fun TaskListItem(
    task: WatchTask,
    onClick: () -> Unit,
    onAction: (TaskAction) -> Unit
) {
    var showActionDialog by remember { mutableStateOf(false) }
    
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { onClick() }
            .padding(horizontal = 4.dp),
        shape = RoundedCornerShape(8.dp)
    ) {
        Column(
            modifier = Modifier.padding(12.dp)
        ) {
            // 任务头部
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.Top
            ) {
                Column(
                    modifier = Modifier.weight(1f)
                ) {
                    Text(
                        text = task.title,
                        style = MaterialTheme.typography.body2,
                        fontWeight = FontWeight.Medium,
                        maxLines = 2,
                        overflow = TextOverflow.Ellipsis
                    )
                    
                    if (task.description.isNotEmpty()) {
                        Spacer(modifier = Modifier.height(2.dp))
                        Text(
                            text = task.description,
                            style = MaterialTheme.typography.caption2,
                            color = MaterialTheme.colors.onSurface.copy(alpha = 0.7f),
                            maxLines = 1,
                            overflow = TextOverflow.Ellipsis
                        )
                    }
                }
                
                Spacer(modifier = Modifier.width(8.dp))
                
                // 任务状态指示器
                TaskStatusIndicator(
                    status = task.status,
                    modifier = Modifier.size(8.dp)
                )
            }
            
            Spacer(modifier = Modifier.height(8.dp))
            
            // 任务信息行
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                // 优先级和难度
                Row(
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    PriorityIndicator(
                        priority = task.priority,
                        size = 12.dp
                    )
                    
                    DifficultyStars(
                        difficulty = task.difficulty,
                        size = 8.dp
                    )
                }
                
                // 奖励和操作按钮
                Row(
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    if (task.reward > 0) {
                        Text(
                            text = task.formattedReward,
                            style = MaterialTheme.typography.caption2,
                            color = Color(0xFFFF9800),
                            fontWeight = FontWeight.Medium
                        )
                    }
                    
                    // 快速操作按钮
                    val availableAction = getAvailableAction(task)
                    if (availableAction != null) {
                        Button(
                            onClick = { showActionDialog = true },
                            modifier = Modifier.size(24.dp),
                            colors = ButtonDefaults.buttonColors(
                                backgroundColor = availableAction.color
                            )
                        ) {
                            Icon(
                                imageVector = availableAction.icon,
                                contentDescription = availableAction.label,
                                modifier = Modifier.size(12.dp),
                                tint = Color.White
                            )
                        }
                    }
                }
            }
            
            // 进度条（如果任务进行中）
            if (task.status == TaskStatus.IN_PROGRESS && task.progress > 0) {
                Spacer(modifier = Modifier.height(6.dp))
                TaskProgressBar(
                    progress = task.progress,
                    modifier = Modifier.fillMaxWidth()
                )
            }
        }
    }
    
    // 操作对话框
    if (showActionDialog) {
        TaskActionDialog(
            task = task,
            onActionSelected = { action ->
                onAction(action)
                showActionDialog = false
            },
            onDismiss = { showActionDialog = false }
        )
    }
}

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
 * 任务进度条
 */
@Composable
private fun TaskProgressBar(
    progress: Float,
    modifier: Modifier = Modifier
) {
    Column(modifier = modifier) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            Text(
                text = "进度",
                style = MaterialTheme.typography.caption2,
                color = MaterialTheme.colors.onSurface.copy(alpha = 0.7f)
            )
            Text(
                text = "${(progress * 100).toInt()}%",
                style = MaterialTheme.typography.caption2,
                color = MaterialTheme.colors.primary,
                fontWeight = FontWeight.Medium
            )
        }
        
        Spacer(modifier = Modifier.height(2.dp))
        
        LinearProgressIndicator(
            progress = progress,
            modifier = Modifier
                .fillMaxWidth()
                .height(3.dp)
                .clip(RoundedCornerShape(1.5.dp)),
            color = MaterialTheme.colors.primary,
            backgroundColor = MaterialTheme.colors.surface
        )
    }
}

/**
 * 空任务视图
 */
@Composable
private fun EmptyTasksView(
    onRefresh: () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Icon(
            imageVector = Icons.Default.Assignment,
            contentDescription = null,
            modifier = Modifier.size(48.dp),
            tint = MaterialTheme.colors.onSurface.copy(alpha = 0.5f)
        )
        
        Spacer(modifier = Modifier.height(16.dp))
        
        Text(
            text = "暂无任务",
            style = MaterialTheme.typography.body1,
            textAlign = TextAlign.Center,
            color = MaterialTheme.colors.onSurface.copy(alpha = 0.7f)
        )
        
        Spacer(modifier = Modifier.height(8.dp))
        
        Text(
            text = "点击刷新获取最新任务",
            style = MaterialTheme.typography.caption1,
            textAlign = TextAlign.Center,
            color = MaterialTheme.colors.onSurface.copy(alpha = 0.5f)
        )
        
        Spacer(modifier = Modifier.height(16.dp))
        
        Button(
            onClick = onRefresh,
            colors = ButtonDefaults.buttonColors(
                backgroundColor = MaterialTheme.colors.primary
            )
        ) {
            Icon(
                imageVector = Icons.Default.Refresh,
                contentDescription = "刷新",
                modifier = Modifier.size(16.dp)
            )
            Spacer(modifier = Modifier.width(4.dp))
            Text(
                text = "刷新",
                style = MaterialTheme.typography.button
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
    onRetry: () -> Unit
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
            
            Button(
                onClick = onRetry,
                colors = ButtonDefaults.buttonColors(
                    backgroundColor = MaterialTheme.colors.primary
                )
            ) {
                Text(
                    text = "重试",
                    style = MaterialTheme.typography.button
                )
            }
        }
    }
}

// MARK: - Dialog Components

/**
 * 筛选对话框
 */
@Composable
private fun FilterDialog(
    currentFilter: TaskFilter,
    onFilterSelected: (TaskFilter) -> Unit,
    onDismiss: () -> Unit
) {
    Dialog(
        showDialog = true,
        onDismissRequest = onDismiss
    ) {
        Column {
            Text(
                text = "筛选任务",
                style = MaterialTheme.typography.title3,
                modifier = Modifier.padding(bottom = 8.dp)
            )
            
            TaskFilter.values().forEach { filter ->
                Chip(
                    onClick = { onFilterSelected(filter) },
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(vertical = 2.dp),
                    colors = ChipDefaults.chipColors(
                        backgroundColor = if (filter == currentFilter) {
                            MaterialTheme.colors.primary
                        } else {
                            MaterialTheme.colors.surface
                        }
                    )
                ) {
                    Text(
                        text = filter.displayName,
                        style = MaterialTheme.typography.body2
                    )
                }
            }
        }
    }
}

/**
 * 排序对话框
 */
@Composable
private fun SortDialog(
    currentSort: TaskSort,
    onSortSelected: (TaskSort) -> Unit,
    onDismiss: () -> Unit
) {
    Dialog(
        showDialog = true,
        onDismissRequest = onDismiss
    ) {
        Column {
            Text(
                text = "排序方式",
                style = MaterialTheme.typography.title3,
                modifier = Modifier.padding(bottom = 8.dp)
            )
            
            TaskSort.values().forEach { sort ->
                Chip(
                    onClick = { onSortSelected(sort) },
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(vertical = 2.dp),
                    colors = ChipDefaults.chipColors(
                        backgroundColor = if (sort == currentSort) {
                            MaterialTheme.colors.primary
                        } else {
                            MaterialTheme.colors.surface
                        }
                    )
                ) {
                    Text(
                        text = sort.displayName,
                        style = MaterialTheme.typography.body2
                    )
                }
            }
        }
    }
}

/**
 * 任务操作对话框
 */
@Composable
private fun TaskActionDialog(
    task: WatchTask,
    onActionSelected: (TaskAction) -> Unit,
    onDismiss: () -> Unit
) {
    val availableActions = getAvailableActions(task)
    
    Dialog(
        showDialog = true,
        onDismissRequest = onDismiss
    ) {
        Column {
            Text(
                text = "任务操作",
                style = MaterialTheme.typography.title3,
                modifier = Modifier.padding(bottom = 8.dp)
            )
            
            availableActions.forEach { action ->
                Chip(
                    onClick = { onActionSelected(action) },
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(vertical = 2.dp),
                    colors = ChipDefaults.chipColors(
                        backgroundColor = action.color
                    )
                ) {
                    Icon(
                        imageVector = action.icon,
                        contentDescription = action.label,
                        modifier = Modifier.size(16.dp),
                        tint = Color.White
                    )
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = action.label,
                        style = MaterialTheme.typography.body2,
                        color = Color.White
                    )
                }
            }
        }
    }
}

// MARK: - Helper Functions

/**
 * 获取可用的任务操作
 */
private fun getAvailableAction(task: WatchTask): TaskAction? {
    return when (task.status) {
        TaskStatus.PENDING -> TaskAction.Accept
        TaskStatus.IN_PROGRESS -> TaskAction.Complete
        else -> null
    }
}

/**
 * 获取所有可用的任务操作
 */
private fun getAvailableActions(task: WatchTask): List<TaskAction> {
    return when (task.status) {
        TaskStatus.PENDING -> listOf(TaskAction.Accept)
        TaskStatus.IN_PROGRESS -> listOf(TaskAction.Complete, TaskAction.Pause)
        TaskStatus.PAUSED -> listOf(TaskAction.Start)
        else -> emptyList()
    }
}

// MARK: - Supporting Types

/**
 * 任务操作枚举
 */
enum class TaskAction(
    val label: String,
    val icon: ImageVector,
    val color: Color
) {
    Accept("接受", Icons.Default.Check, Color(0xFF4CAF50)),
    Start("开始", Icons.Default.PlayArrow, Color(0xFF2196F3)),
    Complete("完成", Icons.Default.CheckCircle, Color(0xFF4CAF50)),
    Pause("暂停", Icons.Default.Pause, Color(0xFFFF9800))
}

/**
 * 任务筛选枚举
 */
enum class TaskFilter(val displayName: String) {
    ALL("全部"),
    PENDING("待接受"),
    IN_PROGRESS("进行中"),
    COMPLETED("已完成"),
    HIGH_PRIORITY("高优先级"),
    QUICK_ACTION("快捷操作")
}

/**
 * 任务排序枚举
 */
enum class TaskSort(val displayName: String) {
    CREATED_DESC("创建时间"),
    PRIORITY_DESC("优先级"),
    DIFFICULTY_ASC("难度"),
    REWARD_DESC("奖励"),
    STATUS("状态")
}