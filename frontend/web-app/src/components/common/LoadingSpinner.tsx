import React from 'react';
import { Spin } from 'antd';
import type { SpinProps } from 'antd';

interface LoadingSpinnerProps extends SpinProps {
  text?: string;
  fullScreen?: boolean;
}

const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({ 
  text = '加载中...', 
  fullScreen = false, 
  ...props 
}) => {
  const spinElement = (
    <Spin 
      tip={text} 
      size="large" 
      {...props}
    />
  );

  if (fullScreen) {
    return (
      <div className="fixed inset-0 flex items-center justify-center bg-white bg-opacity-80 z-50">
        {spinElement}
      </div>
    );
  }

  return (
    <div className="flex items-center justify-center p-8">
      {spinElement}
    </div>
  );
};

export default LoadingSpinner;