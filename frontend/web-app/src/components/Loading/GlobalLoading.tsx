import React from 'react';
import { Spin, Typography } from 'antd';
import { LoadingOutlined } from '@ant-design/icons';

const { Text } = Typography;

interface GlobalLoadingProps {
  size?: 'small' | 'default' | 'large';
  tip?: string;
  spinning?: boolean;
  children?: React.ReactNode;
  style?: React.CSSProperties;
}

const GlobalLoading: React.FC<GlobalLoadingProps> = ({
  size = 'large',
  tip = '加载中...',
  spinning = true,
  children,
  style,
}) => {
  const antIcon = <LoadingOutlined style={{ fontSize: size === 'large' ? 24 : 16 }} spin />;

  if (children) {
    return (
      <Spin 
        indicator={antIcon} 
        spinning={spinning} 
        tip={tip}
        size={size}
        style={style}
      >
        {children}
      </Spin>
    );
  }

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '200px',
        padding: '20px',
        ...style,
      }}
    >
      <Spin 
        indicator={antIcon} 
        size={size}
        style={{ marginBottom: '12px' }}
      />
      <Text type="secondary" style={{ fontSize: '14px' }}>
        {tip}
      </Text>
    </div>
  );
};

// 页面级加载组件
export const PageLoading: React.FC<{ tip?: string }> = ({ tip = '页面加载中...' }) => {
  return (
    <div
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        backgroundColor: 'rgba(255, 255, 255, 0.9)',
        zIndex: 9999,
      }}
    >
      <GlobalLoading tip={tip} />
    </div>
  );
};

// 内容区域加载组件
export const ContentLoading: React.FC<{ tip?: string; height?: string | number }> = ({ 
  tip = '内容加载中...', 
  height = '300px' 
}) => {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        height,
        width: '100%',
      }}
    >
      <GlobalLoading tip={tip} />
    </div>
  );
};

// 按钮加载组件
export const ButtonLoading: React.FC<{ loading?: boolean; children: React.ReactNode }> = ({ 
  loading = false, 
  children 
}) => {
  return (
    <Spin 
      spinning={loading} 
      indicator={<LoadingOutlined style={{ fontSize: 14 }} spin />}
      size="small"
    >
      {children}
    </Spin>
  );
};

export default GlobalLoading;