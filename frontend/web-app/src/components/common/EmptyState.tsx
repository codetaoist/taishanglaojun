import React from 'react';
import { Empty, Button, Space } from 'antd';
import { 
  FileTextOutlined, 
  InboxOutlined, 
  SearchOutlined,
  ExclamationCircleOutlined,
  PlusOutlined
} from '@ant-design/icons';

interface EmptyStateProps {
  type?: 'default' | 'search' | 'error' | 'nodata' | 'custom';
  title?: string;
  description?: string;
  image?: React.ReactNode;
  action?: React.ReactNode;
  className?: string;
  style?: React.CSSProperties;
}

const EmptyState: React.FC<EmptyStateProps> = ({
  type = 'default',
  title,
  description,
  image,
  action,
  className = '',
  style
}) => {
  // 预定义的空状态配置
  const presetConfigs = {
    default: {
      image: <InboxOutlined style={{ fontSize: 64, color: '#d9d9d9' }} />,
      title: '暂无数据',
      description: '当前没有可显示的内容'
    },
    search: {
      image: <SearchOutlined style={{ fontSize: 64, color: '#d9d9d9' }} />,
      title: '未找到相关内容',
      description: '请尝试调整搜索条件或关键词'
    },
    error: {
      image: <ExclamationCircleOutlined style={{ fontSize: 64, color: '#ff4d4f' }} />,
      title: '加载失败',
      description: '数据加载时出现错误，请稍后重试'
    },
    nodata: {
      image: <FileTextOutlined style={{ fontSize: 64, color: '#d9d9d9' }} />,
      title: '还没有内容',
      description: '开始创建您的第一个项目吧'
    },
    custom: {
      image: null,
      title: '',
      description: ''
    }
  };

  const config = presetConfigs[type];
  const finalImage = image || config.image;
  const finalTitle = title || config.title;
  const finalDescription = description || config.description;

  return (
    <div className={`flex items-center justify-center min-h-64 ${className}`} style={style}>
      <div className="text-center max-w-md mx-auto p-8">
        {/* 图标/图片 */}
        <div className="mb-6">
          {finalImage}
        </div>

        {/* 标题 */}
        {finalTitle && (
          <h3 className="text-lg font-semibold text-gray-800 mb-3">
            {finalTitle}
          </h3>
        )}

        {/* 描述 */}
        {finalDescription && (
          <p className="text-gray-500 mb-6 leading-relaxed">
            {finalDescription}
          </p>
        )}

        {/* 操作按钮 */}
        {action && (
          <div className="mt-6">
            {action}
          </div>
        )}
      </div>
    </div>
  );
};

// 预设的操作按钮组件
export const EmptyActions = {
  Refresh: ({ onClick, loading = false }: { onClick?: () => void; loading?: boolean }) => (
    <Button type="primary" onClick={onClick} loading={loading}>
      刷新页面
    </Button>
  ),
  
  Create: ({ onClick, text = '创建新项目' }: { onClick?: () => void; text?: string }) => (
    <Button type="primary" icon={<PlusOutlined />} onClick={onClick}>
      {text}
    </Button>
  ),
  
  GoBack: ({ onClick }: { onClick?: () => void }) => (
    <Button onClick={onClick}>
      返回上一页
    </Button>
  ),
  
  Multiple: ({ actions }: { actions: React.ReactNode[] }) => (
    <Space>
      {actions.map((action, index) => (
        <React.Fragment key={index}>{action}</React.Fragment>
      ))}
    </Space>
  )
};

export default EmptyState;