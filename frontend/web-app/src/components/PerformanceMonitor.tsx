import React from 'react';
import { Card, Progress, Typography, Space, Tag } from 'antd';
import { 
  useRenderCount, 
  useMemoryMonitor 
} from '../utils/performanceOptimization';

const { Text, Title } = Typography;

interface PerformanceMonitorProps {
  componentName: string;
  showInProduction?: boolean;
  memoryThreshold?: number;
  style?: React.CSSProperties;
}

const PerformanceMonitor: React.FC<PerformanceMonitorProps> = ({
  componentName,
  showInProduction = false,
  memoryThreshold = 80,
  style = {}
}) => {
  const renderCount = useRenderCount(componentName);
  const memoryInfo = useMemoryMonitor(memoryThreshold);

  // 在生产环境中默认不显示，除非明确指定
  if (process.env.NODE_ENV === 'production' && !showInProduction) {
    return null;
  }

  const memoryUsagePercent = memoryInfo 
    ? Math.round((memoryInfo.usedJSHeapSize / memoryInfo.totalJSHeapSize) * 100)
    : 0;

  const getMemoryStatus = () => {
    if (memoryUsagePercent > 90) return 'exception';
    if (memoryUsagePercent > 70) return 'active';
    return 'success';
  };

  return (
    <Card
      size="small"
      title={
        <Space>
          <span>🔧 性能监控</span>
          <Tag color="blue">{componentName}</Tag>
        </Space>
      }
      style={{
        position: 'fixed',
        top: 10,
        right: 10,
        width: 300,
        zIndex: 9999,
        opacity: 0.9,
        fontSize: '12px',
        ...style
      }}
      bodyStyle={{ padding: '12px' }}
    >
      <Space direction="vertical" size="small" style={{ width: '100%' }}>
        {/* 渲染次数 */}
        <div>
          <Text strong>渲染次数:</Text>
          <Tag color={renderCount > 10 ? 'orange' : 'green'} style={{ marginLeft: 8 }}>
            {renderCount}
          </Tag>
          {renderCount > 20 && (
            <Text type="warning" style={{ fontSize: '10px' }}>
              ⚠️ 渲染频繁
            </Text>
          )}
        </div>

        {/* 内存使用 */}
        {memoryInfo && (
          <div>
            <Text strong>内存使用:</Text>
            <div style={{ marginTop: 4 }}>
              <Progress
                percent={memoryUsagePercent}
                size="small"
                status={getMemoryStatus()}
                format={(percent) => `${percent}%`}
              />
              <div style={{ fontSize: '10px', marginTop: 2 }}>
                <Text type="secondary">
                  {memoryInfo.usedJSHeapSize}MB / {memoryInfo.totalJSHeapSize}MB
                </Text>
                {memoryUsagePercent > memoryThreshold && (
                  <Text type="danger" style={{ marginLeft: 8 }}>
                    ⚠️ 内存使用过高
                  </Text>
                )}
              </div>
            </div>
          </div>
        )}

        {/* 性能建议 */}
        {(renderCount > 20 || memoryUsagePercent > 80) && (
          <div style={{ 
            background: '#fff2e8', 
            padding: '6px', 
            borderRadius: '4px',
            border: '1px solid #ffbb96'
          }}>
            <Text style={{ fontSize: '10px', color: '#d46b08' }}>
              💡 性能建议:
            </Text>
            <ul style={{ margin: '4px 0 0 0', paddingLeft: '16px', fontSize: '10px' }}>
              {renderCount > 20 && (
                <li>考虑使用 React.memo 或 useMemo 优化渲染</li>
              )}
              {memoryUsagePercent > 80 && (
                <li>检查是否存在内存泄漏或大对象缓存</li>
              )}
            </ul>
          </div>
        )}
      </Space>
    </Card>
  );
};

export default PerformanceMonitor;