import React, { useEffect, useRef, useState } from 'react';
import { Card, Progress, Typography, Space, Tag, Button, Drawer, Statistic, Row, Col } from 'antd';
import { 
  DashboardOutlined, 
  ClockCircleOutlined, 
  ThunderboltOutlined,
  MemoryOutlined,
  WifiOutlined,
  EyeOutlined
} from '@ant-design/icons';

const { Text, Title } = Typography;

interface PerformanceMetrics {
  // Core Web Vitals
  fcp: number; // First Contentful Paint
  lcp: number; // Largest Contentful Paint
  fid: number; // First Input Delay
  cls: number; // Cumulative Layout Shift
  
  // 其他性能指标
  ttfb: number; // Time to First Byte
  domContentLoaded: number;
  loadComplete: number;
  
  // 内存使用
  memoryUsage: {
    used: number;
    total: number;
    percentage: number;
  };
  
  // 网络信息
  connection: {
    effectiveType: string;
    downlink: number;
    rtt: number;
  };
}

interface Props {
  enabled?: boolean;
  showFloatingButton?: boolean;
  autoReport?: boolean;
  reportInterval?: number;
}

const PerformanceMonitor: React.FC<Props> = ({
  enabled = process.env.NODE_ENV === 'development',
  showFloatingButton = true,
  autoReport = false,
  reportInterval = 30000
}) => {
  const [metrics, setMetrics] = useState<PerformanceMetrics | null>(null);
  const [isVisible, setIsVisible] = useState(false);
  const [isCollecting, setIsCollecting] = useState(false);
  const intervalRef = useRef<NodeJS.Timeout>();

  // 收集性能指标
  const collectMetrics = async (): Promise<PerformanceMetrics> => {
    const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
    const paint = performance.getEntriesByType('paint');
    
    // Core Web Vitals
    const fcp = paint.find(entry => entry.name === 'first-contentful-paint')?.startTime || 0;
    
    // 获取LCP
    let lcp = 0;
    if ('PerformanceObserver' in window) {
      try {
        const lcpObserver = new PerformanceObserver((list) => {
          const entries = list.getEntries();
          const lastEntry = entries[entries.length - 1];
          lcp = lastEntry.startTime;
        });
        lcpObserver.observe({ entryTypes: ['largest-contentful-paint'] });
      } catch (e) {
        console.warn('LCP observation failed:', e);
      }
    }

    // 内存使用情况
    const memoryInfo = (performance as any).memory;
    const memoryUsage = memoryInfo ? {
      used: Math.round(memoryInfo.usedJSHeapSize / 1024 / 1024),
      total: Math.round(memoryInfo.totalJSHeapSize / 1024 / 1024),
      percentage: Math.round((memoryInfo.usedJSHeapSize / memoryInfo.totalJSHeapSize) * 100)
    } : { used: 0, total: 0, percentage: 0 };

    // 网络信息
    const connection = (navigator as any).connection || (navigator as any).mozConnection || (navigator as any).webkitConnection;
    const connectionInfo = connection ? {
      effectiveType: connection.effectiveType || 'unknown',
      downlink: connection.downlink || 0,
      rtt: connection.rtt || 0
    } : { effectiveType: 'unknown', downlink: 0, rtt: 0 };

    return {
      fcp,
      lcp,
      fid: 0, // 需要通过PerformanceObserver获取
      cls: 0, // 需要通过PerformanceObserver获取
      ttfb: navigation.responseStart - navigation.requestStart,
      domContentLoaded: navigation.domContentLoadedEventEnd - navigation.navigationStart,
      loadComplete: navigation.loadEventEnd - navigation.navigationStart,
      memoryUsage,
      connection: connectionInfo
    };
  };

  // 获取性能评分
  const getPerformanceScore = (value: number, thresholds: { good: number; poor: number }) => {
    if (value <= thresholds.good) return { score: 'good', color: 'green' };
    if (value <= thresholds.poor) return { score: 'needs-improvement', color: 'orange' };
    return { score: 'poor', color: 'red' };
  };

  // 格式化时间
  const formatTime = (ms: number) => {
    if (ms < 1000) return `${Math.round(ms)}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  };

  // 更新性能指标
  const updateMetrics = async () => {
    setIsCollecting(true);
    try {
      const newMetrics = await collectMetrics();
      setMetrics(newMetrics);
    } catch (error) {
      console.error('Failed to collect performance metrics:', error);
    } finally {
      setIsCollecting(false);
    }
  };

  // 发送性能报告
  const reportMetrics = (metrics: PerformanceMetrics) => {
    // 这里可以发送到分析服务
    console.log('Performance Metrics Report:', metrics);
    
    // 可以发送到后端API
    // apiManager.reportPerformance(metrics);
  };

  useEffect(() => {
    if (!enabled) return;

    // 初始收集
    updateMetrics();

    // 自动报告
    if (autoReport) {
      intervalRef.current = setInterval(() => {
        updateMetrics().then(() => {
          if (metrics) {
            reportMetrics(metrics);
          }
        });
      }, reportInterval);
    }

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [enabled, autoReport, reportInterval]);

  if (!enabled) return null;

  return (
    <>
      {/* 浮动按钮 */}
      {showFloatingButton && (
        <Button
          type="primary"
          shape="circle"
          icon={<DashboardOutlined />}
          size="large"
          className="fixed bottom-4 right-4 z-50 shadow-lg"
          onClick={() => setIsVisible(true)}
          loading={isCollecting}
        />
      )}

      {/* 性能监控面板 */}
      <Drawer
        title={
          <Space>
            <DashboardOutlined />
            <span>性能监控</span>
            <Button size="small" onClick={updateMetrics} loading={isCollecting}>
              刷新
            </Button>
          </Space>
        }
        placement="right"
        width={400}
        open={isVisible}
        onClose={() => setIsVisible(false)}
      >
        {metrics && (
          <Space direction="vertical" className="w-full" size="large">
            {/* Core Web Vitals */}
            <Card size="small" title="Core Web Vitals">
              <Space direction="vertical" className="w-full">
                <div>
                  <Text strong>First Contentful Paint (FCP)</Text>
                  <div className="flex justify-between items-center mt-1">
                    <Text>{formatTime(metrics.fcp)}</Text>
                    <Tag color={getPerformanceScore(metrics.fcp, { good: 1800, poor: 3000 }).color}>
                      {getPerformanceScore(metrics.fcp, { good: 1800, poor: 3000 }).score}
                    </Tag>
                  </div>
                </div>

                <div>
                  <Text strong>Largest Contentful Paint (LCP)</Text>
                  <div className="flex justify-between items-center mt-1">
                    <Text>{formatTime(metrics.lcp)}</Text>
                    <Tag color={getPerformanceScore(metrics.lcp, { good: 2500, poor: 4000 }).color}>
                      {getPerformanceScore(metrics.lcp, { good: 2500, poor: 4000 }).score}
                    </Tag>
                  </div>
                </div>

                <div>
                  <Text strong>Time to First Byte (TTFB)</Text>
                  <div className="flex justify-between items-center mt-1">
                    <Text>{formatTime(metrics.ttfb)}</Text>
                    <Tag color={getPerformanceScore(metrics.ttfb, { good: 800, poor: 1800 }).color}>
                      {getPerformanceScore(metrics.ttfb, { good: 800, poor: 1800 }).score}
                    </Tag>
                  </div>
                </div>
              </Space>
            </Card>

            {/* 加载时间 */}
            <Card size="small" title={<Space><ClockCircleOutlined />加载时间</Space>}>
              <Row gutter={16}>
                <Col span={12}>
                  <Statistic
                    title="DOM加载"
                    value={metrics.domContentLoaded}
                    suffix="ms"
                    precision={0}
                  />
                </Col>
                <Col span={12}>
                  <Statistic
                    title="完全加载"
                    value={metrics.loadComplete}
                    suffix="ms"
                    precision={0}
                  />
                </Col>
              </Row>
            </Card>

            {/* 内存使用 */}
            <Card size="small" title={<Space><MemoryOutlined />内存使用</Space>}>
              <Space direction="vertical" className="w-full">
                <div className="flex justify-between">
                  <Text>已使用: {metrics.memoryUsage.used}MB</Text>
                  <Text>总计: {metrics.memoryUsage.total}MB</Text>
                </div>
                <Progress
                  percent={metrics.memoryUsage.percentage}
                  status={metrics.memoryUsage.percentage > 80 ? 'exception' : 'normal'}
                  strokeColor={
                    metrics.memoryUsage.percentage > 80 ? '#ff4d4f' :
                    metrics.memoryUsage.percentage > 60 ? '#faad14' : '#52c41a'
                  }
                />
              </Space>
            </Card>

            {/* 网络信息 */}
            <Card size="small" title={<Space><WifiOutlined />网络信息</Space>}>
              <Space direction="vertical" className="w-full">
                <div className="flex justify-between">
                  <Text>连接类型:</Text>
                  <Tag>{metrics.connection.effectiveType}</Tag>
                </div>
                <div className="flex justify-between">
                  <Text>下行速度:</Text>
                  <Text>{metrics.connection.downlink} Mbps</Text>
                </div>
                <div className="flex justify-between">
                  <Text>往返时间:</Text>
                  <Text>{metrics.connection.rtt} ms</Text>
                </div>
              </Space>
            </Card>

            {/* 性能建议 */}
            <Card size="small" title={<Space><ThunderboltOutlined />性能建议</Space>}>
              <Space direction="vertical" className="w-full">
                {metrics.fcp > 3000 && (
                  <Text type="warning">• FCP较慢，考虑优化关键渲染路径</Text>
                )}
                {metrics.lcp > 4000 && (
                  <Text type="warning">• LCP较慢，优化最大内容元素加载</Text>
                )}
                {metrics.memoryUsage.percentage > 80 && (
                  <Text type="danger">• 内存使用率过高，检查内存泄漏</Text>
                )}
                {metrics.ttfb > 1800 && (
                  <Text type="warning">• 服务器响应较慢，优化后端性能</Text>
                )}
                {metrics.connection.effectiveType === 'slow-2g' && (
                  <Text type="warning">• 网络连接较慢，启用离线模式</Text>
                )}
              </Space>
            </Card>
          </Space>
        )}
      </Drawer>
    </>
  );
};

export default PerformanceMonitor;