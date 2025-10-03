package com.taishanglaojun.watch.ui.theme

import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.wear.compose.material.Colors
import androidx.wear.compose.material.MaterialTheme

// Primary colors for the Taishanglaojun brand
private val Primary = Color(0xFF6750A4)
private val OnPrimary = Color(0xFFFFFFFF)
private val PrimaryContainer = Color(0xFFEADDFF)
private val OnPrimaryContainer = Color(0xFF21005D)

// Secondary colors
private val Secondary = Color(0xFF625B71)
private val OnSecondary = Color(0xFFFFFFFF)
private val SecondaryContainer = Color(0xFFE8DEF8)
private val OnSecondaryContainer = Color(0xFF1D192B)

// Surface colors for watch
private val Surface = Color(0xFF1C1B1F)
private val OnSurface = Color(0xFFE6E1E5)
private val SurfaceVariant = Color(0xFF49454F)
private val OnSurfaceVariant = Color(0xFFCAC4D0)

// Background colors
private val Background = Color(0xFF000000)
private val OnBackground = Color(0xFFE6E1E5)

// Error colors
private val Error = Color(0xFFBA1A1A)
private val OnError = Color(0xFFFFFFFF)
private val ErrorContainer = Color(0xFFFFDAD6)
private val OnErrorContainer = Color(0xFF410002)

// Outline colors
private val Outline = Color(0xFF938F99)
private val OutlineVariant = Color(0xFF49454F)

// Task status colors
val TaskPendingColor = Color(0xFFFFB74D)
val TaskInProgressColor = Color(0xFF42A5F5)
val TaskCompletedColor = Color(0xFF66BB6A)
val TaskCancelledColor = Color(0xFFEF5350)

// Priority colors
val PriorityLowColor = Color(0xFF81C784)
val PriorityMediumColor = Color(0xFFFFB74D)
val PriorityHighColor = Color(0xFFE57373)

// Connection status colors
val ConnectedColor = Color(0xFF4CAF50)
val DisconnectedColor = Color(0xFFF44336)
val SyncingColor = Color(0xFF2196F3)

private val WearColorPalette = Colors(
    primary = Primary,
    primaryVariant = PrimaryContainer,
    secondary = Secondary,
    secondaryVariant = SecondaryContainer,
    background = Background,
    surface = Surface,
    error = Error,
    onPrimary = OnPrimary,
    onSecondary = OnSecondary,
    onBackground = OnBackground,
    onSurface = OnSurface,
    onSurfaceVariant = OnSurfaceVariant,
    onError = OnError
)

@Composable
fun TaishanglaojunWatchTheme(
    content: @Composable () -> Unit
) {
    MaterialTheme(
        colors = WearColorPalette,
        typography = Typography,
        content = content
    )
}

// Extension functions for semantic colors
@Composable
fun getTaskStatusColor(status: com.taishanglaojun.watch.model.TaskStatus): Color {
    return when (status) {
        com.taishanglaojun.watch.model.TaskStatus.PENDING -> TaskPendingColor
        com.taishanglaojun.watch.model.TaskStatus.ACCEPTED -> TaskInProgressColor
        com.taishanglaojun.watch.model.TaskStatus.IN_PROGRESS -> TaskInProgressColor
        com.taishanglaojun.watch.model.TaskStatus.COMPLETED -> TaskCompletedColor
        com.taishanglaojun.watch.model.TaskStatus.CANCELLED -> TaskCancelledColor
    }
}

@Composable
fun getPriorityColor(priority: com.taishanglaojun.watch.model.TaskPriority): Color {
    return when (priority) {
        com.taishanglaojun.watch.model.TaskPriority.LOW -> PriorityLowColor
        com.taishanglaojun.watch.model.TaskPriority.MEDIUM -> PriorityMediumColor
        com.taishanglaojun.watch.model.TaskPriority.HIGH -> PriorityHighColor
    }
}

@Composable
fun getConnectionStatusColor(isConnected: Boolean, isSyncing: Boolean): Color {
    return when {
        isSyncing -> SyncingColor
        isConnected -> ConnectedColor
        else -> DisconnectedColor
    }
}