import React from 'react';
import { Breadcrumb, Card, Space, Button, Divider } from 'antd';
import { HomeOutlined, LeftOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

interface BreadcrumbItem {
  title: string;
  path?: string;
  icon?: React.ReactNode;
}

interface PageContainerProps {
  title?: string;
  subtitle?: string;
  breadcrumbs?: BreadcrumbItem[];
  extra?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
  style?: React.CSSProperties;
  showBack?: boolean;
  onBack?: () => void;
  loading?: boolean;
  ghost?: boolean; // 是否使用透明背景
  size?: 'small' | 'default' | 'large';
}

const PageContainer: React.FC<PageContainerProps> = ({
  title,
  subtitle,
  breadcrumbs,
  extra,
  children,
  className = '',
  style,
  showBack = false,
  onBack,
  loading = false,
  ghost = false,
  size = 'default'
}) => {
  const navigate = useNavigate();

  const handleBack = () => {
    if (onBack) {
      onBack();
    } else {
      navigate(-1);
    }
  };

  // 生成面包屑项
  const breadcrumbItems = breadcrumbs?.map((item, index) => ({
    title: (
      <span className="flex items-center space-x-1">
        {item.icon}
        <span>{item.title}</span>
      </span>
    ),
    href: item.path,
    onClick: item.path ? (e: React.MouseEvent) => {
      e.preventDefault();
      navigate(item.path!);
    } : undefined
  }));

  const paddingMap = {
    small: 'p-4',
    default: 'p-6',
    large: 'p-8'
  };

  return (
    <div className={`min-h-full ${className}`} style={style}>
      {/* 页面头部 */}
      <div className={`bg-white ${ghost ? 'bg-transparent' : 'border-b border-gray-200'} ${paddingMap[size]}`}>
        {/* 面包屑导航 */}
        {breadcrumbs && breadcrumbs.length > 0 && (
          <div className="mb-4">
            <Breadcrumb
              items={[
                {
                  title: (
                    <span className="flex items-center space-x-1">
                      <HomeOutlined />
                      <span>首页</span>
                    </span>
                  ),
                  href: '/',
                  onClick: (e: React.MouseEvent) => {
                    e.preventDefault();
                    navigate('/');
                  }
                },
                ...breadcrumbItems || []
              ]}
            />
          </div>
        )}

        {/* 页面标题区域 */}
        {(title || showBack || extra) && (
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              {/* 返回按钮 */}
              {showBack && (
                <Button
                  type="text"
                  icon={<LeftOutlined />}
                  onClick={handleBack}
                  className="flex items-center justify-center w-8 h-8 hover:bg-gray-100 rounded-lg"
                />
              )}

              {/* 标题和副标题 */}
              {title && (
                <div>
                  <h1 className="text-2xl font-bold text-gray-900 mb-1">
                    {title}
                  </h1>
                  {subtitle && (
                    <p className="text-gray-500 text-sm">
                      {subtitle}
                    </p>
                  )}
                </div>
              )}
            </div>

            {/* 额外操作区域 */}
            {extra && (
              <div className="flex items-center space-x-2">
                {extra}
              </div>
            )}
          </div>
        )}
      </div>

      {/* 页面内容 */}
      <div className={`flex-1 ${paddingMap[size]}`}>
        {ghost ? (
          children
        ) : (
          <Card
            loading={loading}
            className="shadow-sm"
            styles={{ body: { padding: size === 'small' ? 16 : size === 'large' ? 32 : 24 } }}
          >
            {children}
          </Card>
        )}
      </div>
    </div>
  );
};

export default PageContainer;