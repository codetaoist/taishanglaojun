package com.taishanglaojun.watch.di

import android.content.Context
import com.taishanglaojun.watch.data.repository.TaskRepository
import com.taishanglaojun.watch.services.ConnectivityService
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

/**
 * Hilt 应用模块
 * 提供应用级别的依赖注入
 */
@Module
@InstallIn(SingletonComponent::class)
object AppModule {

    @Provides
    @Singleton
    fun provideTaskRepository(
        @ApplicationContext context: Context
    ): TaskRepository {
        return TaskRepository(context)
    }

    @Provides
    @Singleton
    fun provideConnectivityService(
        @ApplicationContext context: Context,
        taskRepository: TaskRepository
    ): ConnectivityService {
        return ConnectivityService(context, taskRepository)
    }
}