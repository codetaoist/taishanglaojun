import React, { useState, useEffect } from 'react';
import { DataManager, DataSyncStatus, DatabaseHealthStatus, DatabaseStats } from '../services/dataManager';

interface DataDashboardProps {
  className?: string;
}

interface SyncStatus {
  healthy: boolean;
  errors: string[];
  stats: DatabaseStats | null;
  timestamp: string;
}

const DataDashboard: React.FC<DataDashboardProps> = ({ className = '' }) => {
  const [healthStatus, setHealthStatus] = useState<DatabaseHealthStatus | null>(null);
  const [stats, setStats] = useState<DatabaseStats | null>(null);
  const [syncStatus, setSyncStatus] = useState<SyncStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // 加载数据
  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);

      const [healthData, statsData] = await Promise.all([
        DataManager.checkDatabaseHealth(),
        DataManager.getDatabaseStatistics(),
      ]);

      setHealthStatus(healthData);
      setStats(statsData);
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载数据失败');
    } finally {
      setLoading(false);
    }
  };

  // 组件挂载时加载数据
  useEffect(() => {
    loadData();

    // 添加同步状态监听器
    const handleSyncStatus = (status: SyncStatus) => {
      setSyncStatus(status);
    };

    DataSyncStatus.addListener(handleSyncStatus);

    // 定期检查同步状态
    const interval = setInterval(() => {
      DataSyncStatus.checkSyncStatus();
    }, 30000); // 每30秒检查一次

    // 立即检查一次
    DataSyncStatus.checkSyncStatus();

    return () => {
      DataSyncStatus.removeListener(handleSyncStatus);
      clearInterval(interval);
    };
  }, []);

  // 渲染健康状态指示器
  const renderHealthIndicator = (healthy: boolean, label: string, error?: string) => (
    <div className="flex items-center space-x-2">
      <div
        className={`w-3 h-3 rounded-full ${
          healthy ? 'bg-green-500' : 'bg-red-500'
        }`}
      />
      <span className="text-sm font-medium">{label}</span>
      {error && (
        <span className="text-xs text-red-600 truncate" title={error}>
          ({error})
        </span>
      )}
    </div>
  );

  // 渲染统计卡片
  const renderStatCard = (title: string, value: number | string, subtitle?: string) => (
    <div className="bg-white rounded-lg p-4 shadow-sm border">
      <h3 className="text-sm font-medium text-gray-500 mb-1">{title}</h3>
      <p className="text-2xl font-bold text-gray-900">{value}</p>
      {subtitle && <p className="text-xs text-gray-500 mt-1">{subtitle}</p>}
    </div>
  );

  // 格式化文件大小
  const formatFileSize = (bytes: number): string => {
    const units = ['B', 'KB', 'MB', 'GB'];
    let size = bytes;
    let unitIndex = 0;

    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024;
      unitIndex++;
    }

    return `${size.toFixed(2)} ${units[unitIndex]}`;
  };

  if (loading) {
    return (
      <div className={`p-6 ${className}`}>
        <div className="animate-pulse">
          <div className="h-6 bg-gray-200 rounded mb-4"></div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
            {[1, 2, 3].map((i) => (
              <div key={i} className="h-24 bg-gray-200 rounded"></div>
            ))}
          </div>
          <div className="h-32 bg-gray-200 rounded"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`p-6 ${className}`}>
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <h3 className="text-lg font-medium text-red-800 mb-2">加载失败</h3>
          <p className="text-red-600">{error}</p>
          <button
            onClick={loadData}
            className="mt-3 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 transition-colors"
          >
            重试
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className={`p-6 space-y-6 ${className}`}>
      {/* 标题和刷新按钮 */}
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold text-gray-900">数据管理仪表板</h2>
        <button
          onClick={loadData}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
        >
          刷新数据
        </button>
      </div>

      {/* 同步状态 */}
      {syncStatus && (
        <div className={`p-4 rounded-lg border ${
          syncStatus.healthy ? 'bg-green-50 border-green-200' : 'bg-red-50 border-red-200'
        }`}>
          <div className="flex items-center space-x-2 mb-2">
            <div
              className={`w-3 h-3 rounded-full ${
                syncStatus.healthy ? 'bg-green-500' : 'bg-red-500'
              }`}
            />
            <span className="font-medium">
              数据同步状态: {syncStatus.healthy ? '正常' : '异常'}
            </span>
            <span className="text-xs text-gray-500">
              最后更新: {new Date(syncStatus.timestamp).toLocaleTimeString()}
            </span>
          </div>
          {syncStatus.errors.length > 0 && (
            <div className="mt-2">
              <p className="text-sm text-red-600 mb-1">错误信息:</p>
              <ul className="text-xs text-red-500 space-y-1">
                {syncStatus.errors.map((error, index) => (
                  <li key={index}>• {error}</li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}

      {/* 数据库健康状态 */}
      {healthStatus && (
        <div className="bg-white rounded-lg p-6 shadow-sm border">
          <h3 className="text-lg font-medium text-gray-900 mb-4">数据库健康状态</h3>
          <div className="space-y-3">
            {renderHealthIndicator(
              healthStatus.main_db_healthy,
              '主数据库',
              healthStatus.main_db_error
            )}
            {renderHealthIndicator(
              healthStatus.chat_db_healthy,
              '聊天数据库',
              healthStatus.chat_db_error
            )}
            {renderHealthIndicator(
              healthStatus.storage_db_healthy,
              '存储数据库',
              healthStatus.storage_db_error
            )}
          </div>
        </div>
      )}

      {/* 数据库统计 */}
      {stats && (
        <div className="space-y-4">
          <h3 className="text-lg font-medium text-gray-900">数据库统计</h3>
          
          {/* 主数据库统计 */}
          <div>
            <h4 className="text-md font-medium text-gray-700 mb-3">主数据库</h4>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {renderStatCard('用户数量', stats.main_db.users, '注册用户总数')}
              {renderStatCard('项目数量', stats.main_db.projects, '创建项目总数')}
            </div>
          </div>

          {/* 聊天数据库统计 */}
          <div>
            <h4 className="text-md font-medium text-gray-700 mb-3">聊天数据库</h4>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {renderStatCard('会话数量', stats.chat_db.sessions, 'AI聊天会话总数')}
              {renderStatCard('消息数量', stats.chat_db.messages, '聊天消息总数')}
            </div>
          </div>

          {/* 存储数据库统计 */}
          <div>
            <h4 className="text-md font-medium text-gray-700 mb-3">存储数据库</h4>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {renderStatCard('文件数量', stats.storage_db.files, '存储文件总数')}
              {renderStatCard(
                '存储大小',
                formatFileSize(stats.storage_db.total_size),
                '文件总大小'
              )}
            </div>
          </div>
        </div>
      )}

      {/* 数据架构说明 */}
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <h3 className="text-lg font-medium text-blue-800 mb-2">数据架构说明</h3>
        <div className="text-sm text-blue-700 space-y-2">
          <p>
            <strong>分离式存储架构:</strong> 应用采用多数据库分离存储策略，提高性能和可维护性。
          </p>
          <ul className="list-disc list-inside space-y-1 ml-4">
            <li><strong>主数据库 (taishang.db):</strong> 存储用户信息、项目数据、好友关系等核心业务数据</li>
            <li><strong>聊天数据库 (chat.db):</strong> 专门存储AI聊天会话和消息，优化聊天性能</li>
            <li><strong>存储数据库 (storage.db):</strong> 管理文件信息、备份记录等存储相关数据</li>
          </ul>
          <p>
            <strong>优势:</strong> 模块化设计、性能优化、数据隔离、便于维护和扩展。
          </p>
        </div>
      </div>
    </div>
  );
};

export default DataDashboard;