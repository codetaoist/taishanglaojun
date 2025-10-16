import React, { useState, useEffect, useCallback } from 'react';
import { 
  Card, 
  Row, 
  Col, 
  Typography, 
  Space, 
  Tag, 
  Statistic,
  Button,
  Badge,
  Alert,
  Progress,
  Tooltip
} from 'antd';
import {
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  CloseCircleOutlined,
  ThunderboltOutlined,
  DatabaseOutlined,
  CloudServerOutlined,
  SafetyOutlined,
  MonitorOutlined,
  ReloadOutlined,
  SettingOutlined,
  SyncOutlined,
  InfoCircleOutlined,
  DesktopOutlined,
  SaveOutlined,
  WifiOutlined,
  SecurityScanOutlined,
  ApiOutlined,
  HeartOutlined,
  WarningOutlined,
  FireOutlined
} from '@ant-design/icons';
import dashboardService from '../../services/dashboardService';
import advancedAIService from '../../services/advancedAiService';

const { Text, Title } = Typography;

interface SystemMetric {
  name: string;
  value: number;
  status: 'healthy' | 'warning' | 'critical';
  unit: string;
  description: string;
  threshold: {
    warning: number;
    critical: number;
  };
}

interface ServiceStatus {
  name: string;
  status: 'online' | 'offline' | 'maintenance' | 'degraded';
  uptime?: string;
  responseTime?: number;
  description: string;
  lastCheck: string;
}

interface SystemStatusProps {
  className?: string;
}

const SystemStatus: React.FC<SystemStatusProps> = ({ className }) => {
  const [metrics, setMetrics] = useState<SystemMetric[]>([]);
  const [services, setServices] = useState<ServiceStatus[]>([]);
  const [loading, setLoading] = useState(false);
  const [lastUpdate, setLastUpdate] = useState<string>('');
  const [error, setError] = useState<string | null>(null);

  const computeStatusByThreshold = useCallback((value: number, warning: number, critical: number): 'healthy' | 'warning' | 'critical' => {
    if (value >= critical) return 'critical';
    if (value >= warning) return 'warning';
    return 'healthy';
  }, []);

  const fetchMetrics = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      const [metricsResp, aiPerfResp] = await Promise.all([
        dashboardService.getSystemMetrics(),
        advancedAIService.getPerformanceMetrics().catch(() => ({ success: false, data: undefined }))
      ]);

      const metricsData = metricsResp.success ? metricsResp.data : undefined;
      const aiLatency = (aiPerfResp as any)?.data?.latency;

      const newMetrics: SystemMetric[] = [];

      if (metricsData?.cpu?.usage !== undefined) {
        const val = Number(metricsData.cpu.usage) || 0;
        newMetrics.push({
          name: 'CPU使用率',
          value: Math.round(val * 10) / 10,
          status: computeStatusByThreshold(val, 70, 90),
          unit: '%',
          description: '系统CPU使用情况',
          threshold: { warning: 70, critical: 90 }
        });
      }

      if (metricsData?.memory?.usage !== undefined) {
        const val = Number(metricsData.memory.usage) || 0;
        newMetrics.push({
          name: '内存使用率',
          value: Math.round(val * 10) / 10,
          status: computeStatusByThreshold(val, 80, 95),
          unit: '%',
          description: '系统内存使用情况',
          threshold: { warning: 80, critical: 95 }
        });
      }

      if (metricsData?.disk?.usage !== undefined) {
        const val = Number(metricsData.disk.usage) || 0;
        newMetrics.push({
          name: '磁盘使用率',
          value: Math.round(val * 10) / 10,
          status: computeStatusByThreshold(val, 85, 95),
          unit: '%',
          description: '磁盘空间使用情况',
          threshold: { warning: 85, critical: 95 }
        });
      }

      if (metricsData?.network?.latency !== undefined) {
        const val = Number(metricsData.network.latency) || 0;
        newMetrics.push({
          name: '网络延迟',
          value: Math.round(val),
          status: computeStatusByThreshold(val, 100, 200),
          unit: 'ms',
          description: '网络响应延迟',
          threshold: { warning: 100, critical: 200 }
        });
      }

      if (aiLatency !== undefined) {
        const val = Number(aiLatency) || 0;
        newMetrics.push({
          name: 'AI响应速度',
          value: Math.round(val),
          status: computeStatusByThreshold(val, 1000, 2000),
          unit: 'ms',
          description: 'AI服务响应时间',
          threshold: { warning: 1000, critical: 2000 }
        });
      }

      setMetrics(newMetrics);
      setLastUpdate(new Date().toLocaleTimeString());
    } catch (e) {
      setError('系统指标数据加载失败');
    } finally {
      setLoading(false);
    }
  }, [computeStatusByThreshold]);

  const fetchServiceStatuses = useCallback(async () => {
    try {
      const [aiStatusResp, healthResp] = await Promise.all([
        advancedAIService.getSystemStatus().catch(() => ({ success: false } as any)),
        dashboardService.getHealthStatus().catch(() => ({ success: false } as any))
      ]);

      const servicesList: ServiceStatus[] = [];

      // AI智能服务
      let aiStatus: 'online' | 'degraded' | 'offline' = 'online';
      let aiRespTime: number | undefined = undefined;
      let lastCheck = '刚刚';

      const overall = (aiStatusResp as any)?.data?.overall_health ?? (aiStatusResp as any)?.OverallHealth;
      if (overall !== undefined) {
        const h = Number(overall);
        if (h < 0.5) aiStatus = 'offline';
        else if (h < 0.8) aiStatus = 'degraded';
        else aiStatus = 'online';
      }

      // 补充 AI 响应时间
      const perf = await advancedAIService.getPerformanceMetrics().catch(() => ({ success: false } as any));
      aiRespTime = (perf as any)?.data?.latency;

      servicesList.push({
        name: 'AI智能服务',
        status: aiStatus,
        uptime: undefined,
        responseTime: aiRespTime,
        description: '智能对话和分析服务',
        lastCheck,
      });

      // 用户认证服务（以总体健康 /health 作为近似）
      let authStatus: 'online' | 'offline' | 'degraded' = 'online';
      const healthOk = (healthResp as any)?.success && ((healthResp as any)?.data?.status === 'ok' || (healthResp as any)?.data?.status === 'healthy');
      if (!healthOk) authStatus = 'degraded';
      servicesList.push({
        name: '用户认证服务',
        status: authStatus,
        description: '用户登录和权限管理',
        lastCheck: '刚刚',
      });

      // 其他服务保留占位，使用健康状态近似。
      servicesList.push({ name: '智慧库服务', status: authStatus, description: '知识库和内容管理', lastCheck: '刚刚' });
      servicesList.push({ name: '项目管理服务', status: authStatus, description: '项目和任务管理', lastCheck: '刚刚' });
      servicesList.push({ name: '安全监控服务', status: authStatus, description: '系统安全和威胁检测', lastCheck: '刚刚' });
      servicesList.push({ name: '数据备份服务', status: authStatus, description: '数据备份和恢复', lastCheck: '刚刚' });

      setServices(servicesList);
    } catch (e) {
      // 若失败，显示最小信息但不中断渲染
      setServices([
        { name: 'AI智能服务', status: 'degraded', description: '智能对话和分析服务', lastCheck: '刚刚' },
      ]);
    }
  }, []);

  const refreshAll = useCallback(async () => {
    await Promise.all([fetchMetrics(), fetchServiceStatuses()]);
  }, [fetchMetrics, fetchServiceStatuses]);

  useEffect(() => {
    refreshAll();
    const interval = setInterval(refreshAll, 30000);
    return () => clearInterval(interval);
  }, [refreshAll]);

  const getStatusColor = (status: string) => {
    const colorMap = {
      healthy: '#52c41a',
      online: '#52c41a',
      warning: '#fa8c16',
      degraded: '#fa8c16',
      critical: '#f5222d',
      offline: '#f5222d',
      maintenance: '#1890ff'
    };
    return colorMap[status as keyof typeof colorMap] || '#d9d9d9';
  };

  const getStatusIcon = (status: string) => {
    const iconMap = {
      healthy: <CheckCircleOutlined style={{ color: '#52c41a' }} />,
      online: <CheckCircleOutlined style={{ color: '#52c41a' }} />,
      warning: <ExclamationCircleOutlined style={{ color: '#fa8c16' }} />,
      degraded: <ExclamationCircleOutlined style={{ color: '#fa8c16' }} />,
      critical: <CloseCircleOutlined style={{ color: '#f5222d' }} />,
      offline: <CloseCircleOutlined style={{ color: '#f5222d' }} />,
      maintenance: <SyncOutlined spin style={{ color: '#1890ff' }} />
    };
    return iconMap[status as keyof typeof iconMap] || <InfoCircleOutlined />;
  };

  // 已废弃的总体状态计算，保留下面 getOverallStatus 用于 UI 展示

  const getProgressColor = (status: string) => {
    const colorMap = {
      healthy: '#52c41a',
      warning: '#fa8c16',
      critical: '#f5222d'
    };
    return colorMap[status as keyof typeof colorMap] || '#d9d9d9';
  };

  const handleRefresh = () => {
    setLoading(true);
    refreshAll().finally(() => setLoading(false));
  };

  const getOverallStatus = () => {
    const criticalCount = metrics.filter(m => m.status === 'critical').length;
    const warningCount = metrics.filter(m => m.status === 'warning').length;
    const offlineServices = services.filter(s => s.status === 'offline').length;
    const degradedServices = services.filter(s => s.status === 'degraded').length;

    if (criticalCount > 0 || offlineServices > 0) {
      return { status: 'critical', text: '系统异常', color: '#f5222d' };
    }
    if (warningCount > 0 || degradedServices > 0) {
      return { status: 'warning', text: '需要关注', color: '#fa8c16' };
    }
    return { status: 'healthy', text: '运行正常', color: '#52c41a' };
  };

  const overallStatus = getOverallStatus();

  return (
    <div className={className}>
      <Card
        title={
          <div className="flex items-center justify-between">
            <Space>
              <MonitorOutlined className="text-blue-500" />
              <Title level={4} className="m-0">系统状态</Title>
              <Badge 
                status={overallStatus.status as any} 
                text={overallStatus.text}
                style={{ color: overallStatus.color }}
              />
            </Space>
            <Space>
              <Text type="secondary" className="text-sm">
                最后更新: {lastUpdate}
              </Text>
              <Button 
                type="text" 
                icon={<ReloadOutlined />} 
                loading={loading}
                onClick={handleRefresh}
                size="small"
              >
                刷新
              </Button>
              <Button 
                type="text" 
                icon={<SettingOutlined />} 
                size="small"
              >
                设置
              </Button>
            </Space>
          </div>
        }
        className="shadow-lg"
      >
        {error && (
          <Alert type="error" message={error} showIcon className="mb-4" />
        )}
        {/* 系统概览警告 */}
        {overallStatus.status !== 'healthy' && (
          <Alert
            message={`系统状态: ${overallStatus.text}`}
            description={
              overallStatus.status === 'critical' 
                ? '检测到严重问题，请立即处理'
                : '检测到一些需要关注的问题'
            }
            type={overallStatus.status === 'critical' ? 'error' : 'warning'}
            showIcon
            className="mb-6"
            action={
              <Button size="small" type="primary" ghost>
                查看详情
              </Button>
            }
          />
        )}

        {/* 系统指标 */}
        <div className="mb-8">
          <Title level={5} className="mb-4 flex items-center">
            <ThunderboltOutlined className="mr-2 text-yellow-500" />
            系统指标
          </Title>
          <Row gutter={[16, 16]}>
            {metrics.map((metric, index) => (
              <Col xs={24} sm={12} lg={8} key={index}>
                <Card size="small" className="h-full">
                  <div className="flex items-center justify-between mb-3">
                    <Space>
                      {metric.name === 'CPU使用率' && <DesktopOutlined />}
                {metric.name === '内存使用率' && <CloudServerOutlined />}
                {metric.name === '磁盘使用率' && <SaveOutlined />}
                {metric.name === '网络延迟' && <WifiOutlined />}
                      {metric.name === 'AI响应速度' && <ThunderboltOutlined />}
                      {metric.name === '数据库连接' && <DatabaseOutlined />}
                      <Text strong className="text-sm">{metric.name}</Text>
                    </Space>
                    {getStatusIcon(metric.status)}
                  </div>
                  <div className="mb-2">
                    <Progress
                      percent={metric.unit === '%' ? metric.value : Math.min((metric.value / metric.threshold.warning) * 100, 100)}
                      strokeColor={getProgressColor(metric.status)}
                      showInfo={false}
                      size="small"
                    />
                  </div>
                  <div className="flex items-center justify-between">
                    <Text className="text-lg font-semibold">
                      {metric.value}{metric.unit}
                    </Text>
                    <Tooltip title={metric.description}>
                      <Tag color={getStatusColor(metric.status)} className="text-xs">
                        {metric.status.toUpperCase()}
                      </Tag>
                    </Tooltip>
                  </div>
                </Card>
              </Col>
            ))}
          </Row>
        </div>

        {/* 服务状态 */}
        <div>
          <Title level={5} className="mb-4 flex items-center">
            <CloudServerOutlined className="mr-2 text-blue-500" />
            服务状态
          </Title>
          <Row gutter={[16, 16]}>
            {services.map((service, index) => (
              <Col xs={24} sm={12} lg={8} key={index}>
                <Card size="small" className="h-full">
                  <div className="flex items-center justify-between mb-3">
                    <Space>
                      {service.name.includes('AI') && <ThunderboltOutlined />}
                      {service.name.includes('认证') && <SecurityScanOutlined />}
                      {service.name.includes('智慧') && <DatabaseOutlined />}
                      {service.name.includes('项目') && <ApiOutlined />}
                      {service.name.includes('安全') && <SecurityScanOutlined />}
                      {service.name.includes('备份') && <SaveOutlined />}
                      <Text strong className="text-sm">{service.name}</Text>
                    </Space>
                    {getStatusIcon(service.status)}
                  </div>
                  <div className="space-y-2">
                    <div className="flex justify-between">
                      <Text type="secondary" className="text-xs">运行时间</Text>
                      <Text className="text-xs font-medium">{service.uptime ?? '-'}</Text>
                    </div>
                    <div className="flex justify-between">
                      <Text type="secondary" className="text-xs">响应时间</Text>
                      <Text className="text-xs font-medium">
                        {service.responseTime > 0 ? `${service.responseTime}ms` : '-'}
                      </Text>
                    </div>
                    <div className="flex justify-between">
                      <Text type="secondary" className="text-xs">最后检查</Text>
                      <Text className="text-xs">{service.lastCheck}</Text>
                    </div>
                  </div>
                  <div className="mt-3">
                    <Tag 
                      color={getStatusColor(service.status)} 
                      className="text-xs w-full text-center"
                    >
                      {service.status.toUpperCase()}
                    </Tag>
                  </div>
                </Card>
              </Col>
            ))}
          </Row>
        </div>

        {/* 快速统计 */}
        <div className="mt-8 p-4 bg-gray-50 rounded-lg">
          <Row gutter={16}>
            <Col span={6}>
              <Statistic
                title="在线服务"
                value={services.filter(s => s.status === 'online').length}
                suffix={`/ ${services.length}`}
                valueStyle={{ color: '#52c41a' }}
                prefix={<CheckCircleOutlined />}
              />
            </Col>
            <Col span={6}>
              <Statistic
                title="健康指标"
                value={metrics.filter(m => m.status === 'healthy').length}
                suffix={`/ ${metrics.length}`}
                valueStyle={{ color: '#52c41a' }}
                prefix={<HeartOutlined />}
              />
            </Col>
            <Col span={6}>
              <Statistic
                title="警告项目"
                value={metrics.filter(m => m.status === 'warning').length + services.filter(s => s.status === 'degraded').length}
                valueStyle={{ color: '#fa8c16' }}
                prefix={<WarningOutlined />}
              />
            </Col>
            <Col span={6}>
              <Statistic
                title="系统负载"
                value={metrics.find(m => m.name === 'CPU使用率')?.value || 0}
                suffix="%"
                valueStyle={{ color: overallStatus.color }}
                prefix={<FireOutlined />}
              />
            </Col>
          </Row>
        </div>
      </Card>
    </div>
  );
};

export default SystemStatus;