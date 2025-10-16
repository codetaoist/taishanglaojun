import React, { Suspense, lazy, ComponentType, LazyExoticComponent } from 'react';
import { Spin, Result, Button } from 'antd';
import { ReloadOutlined } from '@ant-design/icons';
import ErrorBoundary from './ErrorBoundary';

// 懒加载配置接口
interface LazyLoadConfig {
  fallback?: React.ReactNode;
  errorFallback?: React.ReactNode;
  retryable?: boolean;
  preload?: boolean;
  timeout?: number;
}

// 默认加载组件
const DefaultFallback: React.FC = () => (
  <div className="flex items-center justify-center min-h-[200px]">
    <div className="flex flex-col items-center">
      <Spin size="large" />
      <div className="text-gray-500 mt-2">加载中...</div>
    </div>
  </div>
);

// 默认错误组件
const DefaultErrorFallback: React.FC<{ retry?: () => void }> = ({ retry }) => (
  <div className="flex items-center justify-center min-h-[200px]">
    <Result
      status="error"
      title="组件加载失败"
      subTitle="网络连接异常或组件不存在"
      extra={
        retry && (
          <Button type="primary" icon={<ReloadOutlined />} onClick={retry}>
            重试
          </Button>
        )
      }
    />
  </div>
);

// 创建懒加载组件的高阶函数
export function createLazyComponent<T extends ComponentType<any>>(
  importFunc: () => Promise<{ default: T }>,
  config: LazyLoadConfig = {}
): LazyExoticComponent<T> {
  const {
    fallback = <DefaultFallback />,
    errorFallback,
    retryable = true,
    preload = false,
    timeout = 10000
  } = config;

  // 创建带超时的导入函数
  const importWithTimeout = (): Promise<{ default: T }> => {
    return Promise.race([
      importFunc(),
      new Promise<never>((_, reject) =>
        setTimeout(() => reject(new Error('Component load timeout')), timeout)
      )
    ]);
  };

  const LazyComponent = lazy(importWithTimeout);

  // 预加载功能
  if (preload) {
    // 在空闲时间预加载
    if ('requestIdleCallback' in window) {
      requestIdleCallback(() => {
        importFunc().catch(() => {
          // 预加载失败时静默处理
        });
      });
    } else {
      // 降级到setTimeout
      setTimeout(() => {
        importFunc().catch(() => {
          // 预加载失败时静默处理
        });
      }, 100);
    }
  }

  return LazyComponent;
}

// 懒加载包装器组件
interface LazyWrapperProps extends LazyLoadConfig {
  children: React.ReactNode;
  className?: string;
}

export const LazyWrapper: React.FC<LazyWrapperProps> = ({
  children,
  fallback = <DefaultFallback />,
  errorFallback,
  retryable = true,
  className
}) => {
  const [retryKey, setRetryKey] = React.useState(0);

  const handleRetry = () => {
    setRetryKey(prev => prev + 1);
  };

  const errorFallbackComponent = errorFallback || (
    <DefaultErrorFallback retry={retryable ? handleRetry : undefined} />
  );

  return (
    <div className={className} key={retryKey}>
      <ErrorBoundary fallback={errorFallbackComponent}>
        <Suspense fallback={fallback}>
          {children}
        </Suspense>
      </ErrorBoundary>
    </div>
  );
};

// 路由级别的懒加载组件
interface LazyRouteProps extends LazyLoadConfig {
  component: LazyExoticComponent<ComponentType<any>>;
  props?: any;
}

export const LazyRoute: React.FC<LazyRouteProps> = ({
  component: Component,
  props = {},
  ...config
}) => {
  return (
    <LazyWrapper {...config}>
      <Component {...props} />
    </LazyWrapper>
  );
};

// 条件懒加载Hook
export const useConditionalLazy = <T extends ComponentType<any>>(
  condition: boolean,
  importFunc: () => Promise<{ default: T }>,
  config: LazyLoadConfig = {}
) => {
  const [LazyComponent, setLazyComponent] = React.useState<LazyExoticComponent<T> | null>(null);
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState<Error | null>(null);

  React.useEffect(() => {
    if (condition && !LazyComponent && !loading) {
      setLoading(true);
      setError(null);

      const timeoutId = setTimeout(() => {
        setError(new Error('Component load timeout'));
        setLoading(false);
      }, config.timeout || 10000);

      importFunc()
        .then(module => {
          clearTimeout(timeoutId);
          const Component = lazy(() => Promise.resolve(module));
          setLazyComponent(Component);
          setLoading(false);
        })
        .catch(err => {
          clearTimeout(timeoutId);
          setError(err);
          setLoading(false);
        });
    }
  }, [condition, LazyComponent, loading, importFunc, config.timeout]);

  return { LazyComponent, loading, error };
};

// 预加载工具函数
export const preloadComponent = <T extends ComponentType<any>>(
  importFunc: () => Promise<{ default: T }>
): Promise<{ default: T }> => {
  return importFunc();
};

// 批量预加载
export const preloadComponents = (
  importFuncs: Array<() => Promise<{ default: ComponentType<any> }>>
): Promise<Array<{ default: ComponentType<any> }>> => {
  return Promise.all(importFuncs.map(func => func()));
};

// 智能预加载Hook（基于用户行为）
export const useSmartPreload = (
  importFunc: () => Promise<{ default: ComponentType<any> }>,
  triggers: {
    onHover?: boolean;
    onVisible?: boolean;
    onIdle?: boolean;
    delay?: number;
  } = {}
) => {
  const [preloaded, setPreloaded] = React.useState(false);
  const elementRef = React.useRef<HTMLElement>(null);

  const preload = React.useCallback(() => {
    if (!preloaded) {
      importFunc()
        .then(() => setPreloaded(true))
        .catch(() => {
          // 预加载失败时静默处理
        });
    }
  }, [importFunc, preloaded]);

  React.useEffect(() => {
    const element = elementRef.current;
    if (!element) return;

    // 鼠标悬停预加载
    if (triggers.onHover) {
      const handleMouseEnter = () => preload();
      element.addEventListener('mouseenter', handleMouseEnter);
      return () => element.removeEventListener('mouseenter', handleMouseEnter);
    }

    // 可见性预加载
    if (triggers.onVisible && 'IntersectionObserver' in window) {
      const observer = new IntersectionObserver(
        (entries) => {
          if (entries[0].isIntersecting) {
            preload();
            observer.disconnect();
          }
        },
        { threshold: 0.1 }
      );
      observer.observe(element);
      return () => observer.disconnect();
    }

    // 空闲时预加载
    if (triggers.onIdle) {
      const timeoutId = setTimeout(() => {
        if ('requestIdleCallback' in window) {
          requestIdleCallback(preload);
        } else {
          preload();
        }
      }, triggers.delay || 1000);
      return () => clearTimeout(timeoutId);
    }
  }, [triggers, preload]);

  return { elementRef, preloaded, preload };
};

// 导出默认配置
export const defaultLazyConfig: LazyLoadConfig = {
  fallback: <DefaultFallback />,
  retryable: true,
  timeout: 10000
};

export default LazyWrapper;