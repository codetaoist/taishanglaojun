package com.taishanglaojun.tracker.presentation.navigation

import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.navigation.NavHostController
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import com.taishanglaojun.tracker.presentation.ui.screen.ChatScreen
import com.taishanglaojun.tracker.presentation.ui.screen.HomeScreen
import com.taishanglaojun.tracker.presentation.ui.screen.TrackingScreen

/**
 * 太上老君应用导航组件
 * 定义应用的导航结构和路由
 */
@Composable
fun TaishanglaojunNavigation(
    navController: NavHostController,
    modifier: Modifier = Modifier
) {
    NavHost(
        navController = navController,
        startDestination = "home",
        modifier = modifier
    ) {
        composable("home") {
            HomeScreen(
                onNavigateToTracking = {
                    navController.navigate("tracking")
                },
                onNavigateToChat = {
                    navController.navigate("chat")
                }
            )
        }
        
        composable("tracking") {
            TrackingScreen(
                onNavigateBack = {
                    navController.popBackStack()
                }
            )
        }
        
        composable("chat") {
            ChatScreen(
                onNavigateBack = {
                    navController.popBackStack()
                }
            )
        }
    }
}