import React from 'react';
import { Spin, Space } from 'antd';
import { LoadingOutlined } from '@ant-design/icons';

interface LoadingProps {
  size?: 'small' | 'default' | 'large';
  tip?: string;
  spinning?: boolean;
  children?: React.ReactNode;
  className?: string;
  style?: React.CSSProperties;
  overlay?: boolean; // 是否显示遮罩层
}

const Loading: React.FC<LoadingProps> = ({
  size = 'default',
  tip = '加载中...',
  spinning = true,
  children,
  className = '',
  style,
  overlay = false
}) => {
  const antIcon = <LoadingOutlined style={{ fontSize: size === 'large' ? 24 : size === 'small' ? 14 : 18 }} spin />;

  // 如果有子元素，使用Spin包装
  if (children) {
    return (
      <Spin 
        spinning={spinning} 
        tip={tip} 
        size={size}
        indicator={antIcon}
        className={className}
        style={style}
      >
        {children}
      </Spin>
    );
  }

  // 全屏加载遮罩
  if (overlay) {
    return (
      <div className={`fixed inset-0 bg-white bg-opacity-80 flex items-center justify-center z-50 ${className}`} style={style}>
        <Space direction="vertical" align="center" size="large">
          <div className="relative">
            <div className="w-16 h-16 border-4 border-blue-200 border-t-blue-600 rounded-full animate-spin"></div>
            <div className="absolute inset-0 flex items-center justify-center">
              <div className="w-8 h-8 bg-gradient-to-r from-blue-500 to-purple-600 rounded-full flex items-center justify-center">
                <span className="text-white font-bold text-sm">太</span>
              </div>
            </div>
          </div>
          <div className="text-gray-600 font-medium">{tip}</div>
        </Space>
      </div>
    );
  }

  // 普通加载指示器
  return (
    <div className={`flex items-center justify-center p-8 ${className}`} style={style}>
      <Space direction="vertical" align="center">
        <Spin indicator={antIcon} size={size} />
        {tip && <div className="text-gray-500 mt-2">{tip}</div>}
      </Space>
    </div>
  );
};

export default Loading;