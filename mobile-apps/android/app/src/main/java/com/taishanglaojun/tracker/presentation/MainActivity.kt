package com.taishanglaojun.tracker.presentation

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Surface
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.tooling.preview.Preview
import androidx.navigation.compose.rememberNavController
import com.taishanglaojun.tracker.presentation.navigation.TaishanglaojunNavigation
import com.taishanglaojun.tracker.presentation.ui.theme.TaishanglaojunTrackerTheme
import dagger.hilt.android.AndroidEntryPoint

/**
 * 太上老君追踪器主Activity
 * 使用Jetpack Compose构建UI
 */
@AndroidEntryPoint
class MainActivity : ComponentActivity() {
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        
        setContent {
            TaishanglaojunTrackerTheme {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MaterialTheme.colorScheme.background
                ) {
                    TaishanglaojunApp()
                }
            }
        }
    }
}

@Composable
fun TaishanglaojunApp() {
    val navController = rememberNavController()
    
    Scaffold(
        modifier = Modifier.fillMaxSize()
    ) { innerPadding ->
        TaishanglaojunNavigation(
            navController = navController,
            modifier = Modifier.padding(innerPadding)
        )
    }
}

@Preview(showBackground = true)
@Composable
fun TaishanglaojunAppPreview() {
    TaishanglaojunTrackerTheme {
        TaishanglaojunApp()
    }
}