package com.taishanglaojun.tracker.presentation.ui.screen

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.google.android.gms.maps.model.CameraPosition
import com.google.android.gms.maps.model.LatLng
import com.google.maps.android.compose.*
import com.taishanglaojun.tracker.data.model.LocationPoint
import com.taishanglaojun.tracker.presentation.ui.component.PermissionRequestDialog
import com.taishanglaojun.tracker.presentation.ui.component.TrackingStatsCard
import com.taishanglaojun.tracker.presentation.viewmodel.LocationViewModel
import com.taishanglaojun.tracker.utils.PermissionUtils

/**
 * 位置追踪主界面
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TrackingScreen(
    viewModel: LocationViewModel = hiltViewModel()
) {
    val context = LocalContext.current
    
    // 收集状态
    val isTracking by viewModel.isTracking.collectAsStateWithLifecycle()
    val isPaused by viewModel.isPaused.collectAsStateWithLifecycle()
    val currentLocation by viewModel.currentLocation.collectAsStateWithLifecycle()
    val totalDistance by viewModel.totalDistance.collectAsStateWithLifecycle()
    val pointCount by viewModel.pointCount.collectAsStateWithLifecycle()
    val hasLocationPermission by viewModel.hasLocationPermission.collectAsStateWithLifecycle()
    val hasBackgroundPermission by viewModel.hasBackgroundLocationPermission.collectAsStateWithLifecycle()
    val errorMessage by viewModel.errorMessage.collectAsStateWithLifecycle()
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    // 权限请求状态
    var showPermissionDialog by remember { mutableStateOf(false) }
    var showBackgroundPermissionDialog by remember { mutableStateOf(false) }

    // 地图状态
    val cameraPositionState = rememberCameraPositionState {
        position = CameraPosition.fromLatLngZoom(
            currentLocation?.let { LatLng(it.latitude, it.longitude) } 
                ?: LatLng(39.9042, 116.4074), // 默认北京坐标
            15f
        )
    }

    // 更新地图位置
    LaunchedEffect(currentLocation) {
        currentLocation?.let { location ->
            cameraPositionState.animate(
                CameraUpdateOptions(
                    position = CameraPosition.fromLatLngZoom(
                        LatLng(location.latitude, location.longitude),
                        cameraPositionState.position.zoom
                    )
                )
            )
        }
    }

    // 检查权限
    LaunchedEffect(Unit) {
        viewModel.checkPermissions()
    }

    Column(
        modifier = Modifier.fillMaxSize()
    ) {
        // 顶部应用栏
        TopAppBar(
            title = { 
                Text(
                    text = "位置追踪",
                    fontWeight = FontWeight.Bold
                ) 
            },
            actions = {
                // 同步按钮
                IconButton(
                    onClick = { viewModel.syncTrajectories() }
                ) {
                    Icon(
                        imageVector = Icons.Default.Sync,
                        contentDescription = "同步"
                    )
                }
                
                // 设置按钮
                IconButton(
                    onClick = { /* 打开设置 */ }
                ) {
                    Icon(
                        imageVector = Icons.Default.Settings,
                        contentDescription = "设置"
                    )
                }
            }
        )

        // 地图区域
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .weight(1f)
        ) {
            GoogleMap(
                modifier = Modifier.fillMaxSize(),
                cameraPositionState = cameraPositionState,
                properties = MapProperties(
                    isMyLocationEnabled = hasLocationPermission,
                    mapType = MapType.NORMAL
                ),
                uiSettings = MapUiSettings(
                    myLocationButtonEnabled = true,
                    zoomControlsEnabled = false
                )
            ) {
                // 显示当前位置标记
                currentLocation?.let { location ->
                    Marker(
                        state = MarkerState(
                            position = LatLng(location.latitude, location.longitude)
                        ),
                        title = "当前位置",
                        snippet = "精度: ${location.getFormattedAccuracy()}"
                    )
                }
            }

            // 加载指示器
            if (uiState.isLoading) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    CircularProgressIndicator()
                }
            }
        }

        // 统计信息卡片
        TrackingStatsCard(
            isTracking = isTracking,
            isPaused = isPaused,
            totalDistance = totalDistance,
            pointCount = pointCount,
            currentLocation = currentLocation,
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
        )

        // 控制按钮区域
        TrackingControlButtons(
            isTracking = isTracking,
            isPaused = isPaused,
            hasLocationPermission = hasLocationPermission,
            hasBackgroundPermission = hasBackgroundPermission,
            onStartClick = {
                if (!hasLocationPermission) {
                    showPermissionDialog = true
                } else if (!hasBackgroundPermission) {
                    showBackgroundPermissionDialog = true
                } else {
                    viewModel.startTracking()
                }
            },
            onStopClick = { viewModel.stopTracking() },
            onPauseClick = { viewModel.pauseTracking() },
            onResumeClick = { viewModel.resumeTracking() },
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
        )
    }

    // 权限请求对话框
    if (showPermissionDialog) {
        PermissionRequestDialog(
            title = "位置权限",
            message = "需要位置权限才能开始追踪您的位置",
            onConfirm = {
                showPermissionDialog = false
                PermissionUtils.requestLocationPermission(context as androidx.activity.ComponentActivity)
            },
            onDismiss = { showPermissionDialog = false }
        )
    }

    if (showBackgroundPermissionDialog) {
        PermissionRequestDialog(
            title = "后台位置权限",
            message = "需要后台位置权限才能在应用后台时继续追踪位置",
            onConfirm = {
                showBackgroundPermissionDialog = false
                PermissionUtils.requestBackgroundLocationPermission(context as androidx.activity.ComponentActivity)
            },
            onDismiss = { showBackgroundPermissionDialog = false }
        )
    }

    // 错误消息显示
    errorMessage?.let { message ->
        LaunchedEffect(message) {
            // 显示Snackbar或Toast
            viewModel.clearErrorMessage()
        }
    }

    // 同步消息显示
    uiState.syncMessage?.let { message ->
        LaunchedEffect(message) {
            // 显示同步成功消息
            viewModel.clearSyncMessage()
        }
    }
}

/**
 * 追踪控制按钮组件
 */
@Composable
private fun TrackingControlButtons(
    isTracking: Boolean,
    isPaused: Boolean,
    hasLocationPermission: Boolean,
    hasBackgroundPermission: Boolean,
    onStartClick: () -> Unit,
    onStopClick: () -> Unit,
    onPauseClick: () -> Unit,
    onResumeClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    Row(
        modifier = modifier,
        horizontalArrangement = Arrangement.SpaceEvenly,
        verticalAlignment = Alignment.CenterVertically
    ) {
        when {
            !isTracking -> {
                // 开始按钮
                FloatingActionButton(
                    onClick = onStartClick,
                    containerColor = MaterialTheme.colorScheme.primary,
                    modifier = Modifier.size(72.dp)
                ) {
                    Icon(
                        imageVector = Icons.Default.PlayArrow,
                        contentDescription = "开始追踪",
                        modifier = Modifier.size(32.dp)
                    )
                }
            }
            isPaused -> {
                // 恢复按钮
                FloatingActionButton(
                    onClick = onResumeClick,
                    containerColor = MaterialTheme.colorScheme.primary,
                    modifier = Modifier.size(64.dp)
                ) {
                    Icon(
                        imageVector = Icons.Default.PlayArrow,
                        contentDescription = "恢复追踪",
                        modifier = Modifier.size(28.dp)
                    )
                }
                
                Spacer(modifier = Modifier.width(16.dp))
                
                // 停止按钮
                FloatingActionButton(
                    onClick = onStopClick,
                    containerColor = MaterialTheme.colorScheme.error,
                    modifier = Modifier.size(64.dp)
                ) {
                    Icon(
                        imageVector = Icons.Default.Stop,
                        contentDescription = "停止追踪",
                        modifier = Modifier.size(28.dp)
                    )
                }
            }
            else -> {
                // 暂停按钮
                FloatingActionButton(
                    onClick = onPauseClick,
                    containerColor = MaterialTheme.colorScheme.secondary,
                    modifier = Modifier.size(64.dp)
                ) {
                    Icon(
                        imageVector = Icons.Default.Pause,
                        contentDescription = "暂停追踪",
                        modifier = Modifier.size(28.dp)
                    )
                }
                
                Spacer(modifier = Modifier.width(16.dp))
                
                // 停止按钮
                FloatingActionButton(
                    onClick = onStopClick,
                    containerColor = MaterialTheme.colorScheme.error,
                    modifier = Modifier.size(64.dp)
                ) {
                    Icon(
                        imageVector = Icons.Default.Stop,
                        contentDescription = "停止追踪",
                        modifier = Modifier.size(28.dp)
                    )
                }
            }
        }
    }

    // 权限提示
    if (!hasLocationPermission || !hasBackgroundPermission) {
        Spacer(modifier = Modifier.height(8.dp))
        
        Card(
            modifier = Modifier.fillMaxWidth(),
            colors = CardDefaults.cardColors(
                containerColor = MaterialTheme.colorScheme.errorContainer
            )
        ) {
            Column(
                modifier = Modifier.padding(12.dp)
            ) {
                Row(
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Icon(
                        imageVector = Icons.Default.Warning,
                        contentDescription = null,
                        tint = MaterialTheme.colorScheme.onErrorContainer,
                        modifier = Modifier.size(20.dp)
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = "权限提示",
                        style = MaterialTheme.typography.titleSmall,
                        color = MaterialTheme.colorScheme.onErrorContainer,
                        fontWeight = FontWeight.Bold
                    )
                }
                
                Spacer(modifier = Modifier.height(4.dp))
                
                Text(
                    text = when {
                        !hasLocationPermission -> "需要位置权限才能开始追踪"
                        !hasBackgroundPermission -> "需要后台位置权限才能在后台继续追踪"
                        else -> ""
                    },
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onErrorContainer
                )
            }
        }
    }
}