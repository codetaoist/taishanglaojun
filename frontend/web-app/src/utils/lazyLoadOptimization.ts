import React, { ComponentType, lazy, LazyExoticComponent } from 'react';

// 预加载缓存
const preloadCache = new Map<string, Promise<{ default: ComponentType<any> }>>();

// 懒加载配置接口
interface LazyLoadConfig {
  preload?: boolean;
  priority?: 'high' | 'medium' | 'low';
  delay?: number;
  retries?: number;
  chunkName?: string;
}

// 智能懒加载函数
export function smartLazy<T extends ComponentType<any>>(
  importFn: () => Promise<{ default: T }>,
  config: LazyLoadConfig = {}
): LazyExoticComponent<T> {
  const {
    preload = false,
    priority = 'medium',
    delay = 0,
    retries = 3,
    chunkName
  } = config;

  // 创建带重试机制的导入函数
  const importWithRetry = async (): Promise<{ default: T }> => {
    let lastError: Error;
    
    for (let i = 0; i < retries; i++) {
      try {
        if (delay > 0 && i > 0) {
          await new Promise(resolve => setTimeout(resolve, delay * Math.pow(2, i - 1)));
        }
        return await importFn();
      } catch (error) {
        lastError = error as Error;
        console.warn(`Lazy load attempt ${i + 1} failed:`, error);
      }
    }
    
    throw lastError!;
  };

  // 如果启用预加载，立即开始加载
  if (preload) {
    const cacheKey = chunkName || importFn.toString();
    if (!preloadCache.has(cacheKey)) {
      preloadCache.set(cacheKey, importWithRetry());
    }
  }

  return lazy(() => {
    const cacheKey = chunkName || importFn.toString();
    if (preloadCache.has(cacheKey)) {
      return preloadCache.get(cacheKey)!;
    }
    return importWithRetry();
  });
}

// 预加载函数
export function preloadComponent(
  importFn: () => Promise<{ default: ComponentType<any> }>,
  chunkName?: string
): Promise<{ default: ComponentType<any> }> {
  const cacheKey = chunkName || importFn.toString();
  
  if (!preloadCache.has(cacheKey)) {
    preloadCache.set(cacheKey, importFn());
  }
  
  return preloadCache.get(cacheKey)!;
}

// 批量预加载
export function preloadComponents(
  components: Array<{
    importFn: () => Promise<{ default: ComponentType<any> }>;
    chunkName?: string;
    priority?: 'high' | 'medium' | 'low';
  }>
): void {
  // 按优先级排序
  const sortedComponents = components.sort((a, b) => {
    const priorityOrder = { high: 3, medium: 2, low: 1 };
    return priorityOrder[b.priority || 'medium'] - priorityOrder[a.priority || 'medium'];
  });

  // 使用 requestIdleCallback 在空闲时预加载
  const preloadNext = (index: number) => {
    if (index >= sortedComponents.length) return;

    const component = sortedComponents[index];
    preloadComponent(component.importFn, component.chunkName);

    // 使用 requestIdleCallback 或 setTimeout 作为后备
    if (typeof requestIdleCallback !== 'undefined') {
      requestIdleCallback(() => preloadNext(index + 1));
    } else {
      setTimeout(() => preloadNext(index + 1), 0);
    }
  };

  preloadNext(0);
}

// 路由预加载策略
export class RoutePreloader {
  private static instance: RoutePreloader;
  private preloadedRoutes = new Set<string>();
  private routeMap = new Map<string, () => Promise<{ default: ComponentType<any> }>>();

  static getInstance(): RoutePreloader {
    if (!RoutePreloader.instance) {
      RoutePreloader.instance = new RoutePreloader();
    }
    return RoutePreloader.instance;
  }

  // 注册路由
  registerRoute(
    path: string,
    importFn: () => Promise<{ default: ComponentType<any> }>
  ): void {
    this.routeMap.set(path, importFn);
  }

  // 预加载路由
  preloadRoute(path: string): Promise<void> {
    if (this.preloadedRoutes.has(path)) {
      return Promise.resolve();
    }

    const importFn = this.routeMap.get(path);
    if (!importFn) {
      console.warn(`Route ${path} not found in route map`);
      return Promise.resolve();
    }

    this.preloadedRoutes.add(path);
    return preloadComponent(importFn, `route-${path}`)
      .then(() => {
        console.log(`Route ${path} preloaded successfully`);
      })
      .catch((error) => {
        console.error(`Failed to preload route ${path}:`, error);
        this.preloadedRoutes.delete(path);
      });
  }

  // 预加载相关路由
  preloadRelatedRoutes(currentPath: string): void {
    const relatedRoutes = this.getRelatedRoutes(currentPath);
    relatedRoutes.forEach(route => {
      this.preloadRoute(route);
    });
  }

  // 获取相关路由（基于路径相似性）
  private getRelatedRoutes(currentPath: string): string[] {
    const routes = Array.from(this.routeMap.keys());
    const pathSegments = currentPath.split('/').filter(Boolean);
    
    return routes.filter(route => {
      if (route === currentPath) return false;
      
      const routeSegments = route.split('/').filter(Boolean);
      
      // 如果有共同的父路径，认为是相关路由
      for (let i = 0; i < Math.min(pathSegments.length, routeSegments.length); i++) {
        if (pathSegments[i] === routeSegments[i]) {
          return true;
        }
      }
      
      return false;
    });
  }
}

// 智能预加载 Hook
export function useSmartPreload() {
  const preloader = RoutePreloader.getInstance();

  const preloadOnHover = (path: string) => {
    return {
      onMouseEnter: () => preloader.preloadRoute(path),
      onFocus: () => preloader.preloadRoute(path),
    };
  };

  const preloadOnVisible = (path: string) => {
    return (element: HTMLElement | null) => {
      if (!element) return;

      const observer = new IntersectionObserver(
        (entries) => {
          entries.forEach((entry) => {
            if (entry.isIntersecting) {
              preloader.preloadRoute(path);
              observer.unobserve(element);
            }
          });
        },
        { threshold: 0.1 }
      );

      observer.observe(element);
    };
  };

  return {
    preloadOnHover,
    preloadOnVisible,
    preloadRoute: (path: string) => preloader.preloadRoute(path),
    preloadRelatedRoutes: (path: string) => preloader.preloadRelatedRoutes(path),
  };
}

// 性能监控
export class LazyLoadPerformanceMonitor {
  private static loadTimes = new Map<string, number>();
  private static failureCount = new Map<string, number>();

  static recordLoadTime(chunkName: string, loadTime: number): void {
    this.loadTimes.set(chunkName, loadTime);
  }

  static recordFailure(chunkName: string): void {
    const current = this.failureCount.get(chunkName) || 0;
    this.failureCount.set(chunkName, current + 1);
  }

  static getMetrics(): {
    averageLoadTime: number;
    totalFailures: number;
    slowestChunks: Array<{ name: string; time: number }>;
  } {
    const loadTimes = Array.from(this.loadTimes.values());
    const averageLoadTime = loadTimes.length > 0 
      ? loadTimes.reduce((sum, time) => sum + time, 0) / loadTimes.length 
      : 0;

    const totalFailures = Array.from(this.failureCount.values())
      .reduce((sum, count) => sum + count, 0);

    const slowestChunks = Array.from(this.loadTimes.entries())
      .sort(([, a], [, b]) => b - a)
      .slice(0, 5)
      .map(([name, time]) => ({ name, time }));

    return {
      averageLoadTime,
      totalFailures,
      slowestChunks,
    };
  }
}

// 导出默认配置
export const defaultLazyConfig: LazyLoadConfig = {
  preload: false,
  priority: 'medium',
  delay: 100,
  retries: 3,
};

export const highPriorityConfig: LazyLoadConfig = {
  preload: true,
  priority: 'high',
  delay: 0,
  retries: 5,
};

export const lowPriorityConfig: LazyLoadConfig = {
  preload: false,
  priority: 'low',
  delay: 200,
  retries: 2,
};