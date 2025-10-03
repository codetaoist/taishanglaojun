package com.taishanglaojun.tracker

import android.app.Application
import dagger.hilt.android.HiltAndroidApp

/**
 * 太上老君追踪器应用程序类
 * 配置Hilt依赖注入和全局应用设置
 */
@HiltAndroidApp
class TaishanglaojunTrackerApplication : Application() {
    
    override fun onCreate() {
        super.onCreate()
        
        // 初始化应用程序
        initializeApp()
    }
    
    private fun initializeApp() {
        // 这里可以添加应用初始化逻辑
        // 例如：日志配置、崩溃报告、性能监控等
    }
}