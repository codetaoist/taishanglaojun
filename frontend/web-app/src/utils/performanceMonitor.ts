// 性能监控工具
import React from 'react';

interface PerformanceMetrics {
  name: string;
  startTime: number;
  endTime?: number;
  duration?: number;
}

class PerformanceMonitor {
  private metrics: Map<string, PerformanceMetrics> = new Map();
  private isEnabled: boolean;

  constructor() {
    this.isEnabled = process.env.NODE_ENV === 'development' || 
                     localStorage.getItem('performance-monitor') === 'true';
  }

  // 开始性能测量
  start(name: string): void {
    if (!this.isEnabled) return;

    this.metrics.set(name, {
      name,
      startTime: performance.now(),
    });
  }

  // 结束性能测量
  end(name: string): number | null {
    if (!this.isEnabled) return null;

    const metric = this.metrics.get(name);
    if (!metric) {
      console.warn(`Performance metric "${name}" not found`);
      return null;
    }

    const endTime = performance.now();
    const duration = endTime - metric.startTime;

    metric.endTime = endTime;
    metric.duration = duration;

    console.log(`⏱️ Performance: ${name} took ${duration.toFixed(2)}ms`);
    return duration;
  }

  // 测量函数执行时间
  measure<T>(name: string, fn: () => T): T {
    if (!this.isEnabled) return fn();

    this.start(name);
    const result = fn();
    this.end(name);
    return result;
  }

  // 测量异步函数执行时间
  async measureAsync<T>(name: string, fn: () => Promise<T>): Promise<T> {
    if (!this.isEnabled) return fn();

    this.start(name);
    const result = await fn();
    this.end(name);
    return result;
  }

  // 获取所有性能指标
  getMetrics(): PerformanceMetrics[] {
    return Array.from(this.metrics.values());
  }

  // 清除所有性能指标
  clear(): void {
    this.metrics.clear();
  }

  // 获取页面加载性能指标
  getPageLoadMetrics(): Record<string, number> {
    if (!this.isEnabled || typeof window === 'undefined') return {};

    const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
    
    return {
      // DNS 查询时间
      dnsLookup: navigation.domainLookupEnd - navigation.domainLookupStart,
      // TCP 连接时间
      tcpConnect: navigation.connectEnd - navigation.connectStart,
      // 请求响应时间
      request: navigation.responseEnd - navigation.requestStart,
      // DOM 解析时间
      domParse: navigation.domContentLoadedEventEnd - navigation.domContentLoadedEventStart,
      // 页面完全加载时间
      pageLoad: navigation.loadEventEnd - navigation.loadEventStart,
      // 首次内容绘制时间
      firstContentfulPaint: this.getFirstContentfulPaint(),
      // 最大内容绘制时间
      largestContentfulPaint: this.getLargestContentfulPaint(),
    };
  }

  // 获取首次内容绘制时间
  private getFirstContentfulPaint(): number {
    const fcpEntry = performance.getEntriesByName('first-contentful-paint')[0];
    return fcpEntry ? fcpEntry.startTime : 0;
  }

  // 获取最大内容绘制时间
  private getLargestContentfulPaint(): number {
    const lcpEntries = performance.getEntriesByType('largest-contentful-paint');
    const lastEntry = lcpEntries[lcpEntries.length - 1];
    return lastEntry ? lastEntry.startTime : 0;
  }

  // 监控资源加载性能
  getResourceMetrics(): Array<{
    name: string;
    type: string;
    size: number;
    duration: number;
  }> {
    if (!this.isEnabled) return [];

    const resources = performance.getEntriesByType('resource') as PerformanceResourceTiming[];
    
    return resources.map(resource => ({
      name: resource.name,
      type: this.getResourceType(resource.name),
      size: resource.transferSize || 0,
      duration: resource.responseEnd - resource.requestStart,
    }));
  }

  // 获取资源类型
  private getResourceType(url: string): string {
    if (url.includes('.js')) return 'JavaScript';
    if (url.includes('.css')) return 'CSS';
    if (url.includes('.png') || url.includes('.jpg') || url.includes('.svg')) return 'Image';
    if (url.includes('.woff') || url.includes('.ttf')) return 'Font';
    return 'Other';
  }

  // 输出性能报告
  generateReport(): void {
    if (!this.isEnabled) return;

    console.group('📊 Performance Report');
    
    // 页面加载指标
    const pageMetrics = this.getPageLoadMetrics();
    console.group('📄 Page Load Metrics');
    Object.entries(pageMetrics).forEach(([key, value]) => {
      console.log(`${key}: ${value.toFixed(2)}ms`);
    });
    console.groupEnd();

    // 自定义指标
    const customMetrics = this.getMetrics();
    if (customMetrics.length > 0) {
      console.group('⚡ Custom Metrics');
      customMetrics.forEach(metric => {
        if (metric.duration) {
          console.log(`${metric.name}: ${metric.duration.toFixed(2)}ms`);
        }
      });
      console.groupEnd();
    }

    // 资源加载指标
    const resourceMetrics = this.getResourceMetrics();
    if (resourceMetrics.length > 0) {
      console.group('📦 Resource Metrics');
      const groupedResources = resourceMetrics.reduce((acc, resource) => {
        if (!acc[resource.type]) acc[resource.type] = [];
        acc[resource.type].push(resource);
        return acc;
      }, {} as Record<string, typeof resourceMetrics>);

      Object.entries(groupedResources).forEach(([type, resources]) => {
        const totalSize = resources.reduce((sum, r) => sum + r.size, 0);
        const avgDuration = resources.reduce((sum, r) => sum + r.duration, 0) / resources.length;
        console.log(`${type}: ${resources.length} files, ${(totalSize / 1024).toFixed(2)}KB, avg ${avgDuration.toFixed(2)}ms`);
      });
      console.groupEnd();
    }

    console.groupEnd();
  }
}

// 创建全局实例
export const performanceMonitor = new PerformanceMonitor();

// React Hook for performance monitoring
export const usePerformanceMonitor = () => {
  return {
    start: performanceMonitor.start.bind(performanceMonitor),
    end: performanceMonitor.end.bind(performanceMonitor),
    measure: performanceMonitor.measure.bind(performanceMonitor),
    measureAsync: performanceMonitor.measureAsync.bind(performanceMonitor),
    generateReport: performanceMonitor.generateReport.bind(performanceMonitor),
  };
};

// 页面性能监控 Hook
export const usePagePerformance = (pageName: string) => {
  React.useEffect(() => {
    performanceMonitor.start(`page-${pageName}`);
    
    return () => {
      performanceMonitor.end(`page-${pageName}`);
    };
  }, [pageName]);
};

// 组件渲染性能监控 Hook
export const useRenderPerformance = (componentName: string) => {
  const renderCount = React.useRef(0);
  
  React.useEffect(() => {
    renderCount.current += 1;
    const metricName = `${componentName}-render-${renderCount.current}`;
    performanceMonitor.start(metricName);
    
    return () => {
      performanceMonitor.end(metricName);
    };
  });
};

export default performanceMonitor;