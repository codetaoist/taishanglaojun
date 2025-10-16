/**
 * 性能优化工具函数
 */

import { useCallback, useRef, useEffect, useMemo, useState } from 'react';

/**
 * 防抖Hook
 * @param callback 回调函数
 * @param delay 延迟时间（毫秒）
 * @returns 防抖后的函数
 */
export const useDebounce = <T extends (...args: any[]) => any>(
  callback: T,
  delay: number
): T => {
  const timeoutRef = useRef<NodeJS.Timeout>();

  return useCallback(
    ((...args: Parameters<T>) => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
      timeoutRef.current = setTimeout(() => {
        callback(...args);
      }, delay);
    }) as T,
    [callback, delay]
  );
};

/**
 * 节流Hook
 * @param callback 回调函数
 * @param delay 延迟时间（毫秒）
 * @returns 节流后的函数
 */
export const useThrottle = <T extends (...args: any[]) => any>(
  callback: T,
  delay: number
): T => {
  const lastCallRef = useRef<number>(0);

  return useCallback(
    ((...args: Parameters<T>) => {
      const now = Date.now();
      if (now - lastCallRef.current >= delay) {
        lastCallRef.current = now;
        callback(...args);
      }
    }) as T,
    [callback, delay]
  );
};

/**
 * 深度比较Hook，用于优化useEffect依赖
 * @param value 要比较的值
 * @returns 深度比较后的值
 */
export const useDeepMemo = <T>(value: T): T => {
  const ref = useRef<T>(value);
  const signalRef = useRef<number>(0);

  if (!deepEqual(value, ref.current)) {
    ref.current = value;
    signalRef.current += 1;
  }

  return useMemo(() => ref.current, [signalRef.current]);
};

/**
 * 深度比较函数
 */
const deepEqual = (a: any, b: any): boolean => {
  if (a === b) return true;
  
  if (a == null || b == null) return a === b;
  
  if (typeof a !== typeof b) return false;
  
  if (typeof a !== 'object') return a === b;
  
  if (Array.isArray(a) !== Array.isArray(b)) return false;
  
  const keysA = Object.keys(a);
  const keysB = Object.keys(b);
  
  if (keysA.length !== keysB.length) return false;
  
  for (const key of keysA) {
    if (!keysB.includes(key)) return false;
    if (!deepEqual(a[key], b[key])) return false;
  }
  
  return true;
};

/**
 * 虚拟滚动Hook
 * @param items 所有项目
 * @param itemHeight 每项高度
 * @param containerHeight 容器高度
 * @param overscan 预渲染项目数
 * @returns 虚拟滚动相关数据
 */
export const useVirtualScroll = <T>(
  items: T[],
  itemHeight: number,
  containerHeight: number,
  overscan: number = 5
) => {
  const scrollTop = useRef<number>(0);
  
  const visibleRange = useMemo(() => {
    const startIndex = Math.max(0, Math.floor(scrollTop.current / itemHeight) - overscan);
    const endIndex = Math.min(
      items.length - 1,
      Math.ceil((scrollTop.current + containerHeight) / itemHeight) + overscan
    );
    
    return { startIndex, endIndex };
  }, [items.length, itemHeight, containerHeight, overscan, scrollTop.current]);

  const visibleItems = useMemo(() => {
    return items.slice(visibleRange.startIndex, visibleRange.endIndex + 1);
  }, [items, visibleRange.startIndex, visibleRange.endIndex]);

  const totalHeight = items.length * itemHeight;
  const offsetY = visibleRange.startIndex * itemHeight;

  const handleScroll = useCallback((event: React.UIEvent<HTMLDivElement>) => {
    scrollTop.current = event.currentTarget.scrollTop;
  }, []);

  return {
    visibleItems,
    totalHeight,
    offsetY,
    handleScroll,
    startIndex: visibleRange.startIndex,
    endIndex: visibleRange.endIndex,
  };
};

/**
 * 内存监控Hook
 * @param threshold 内存阈值（MB）
 * @returns 内存使用情况
 */
export const useMemoryMonitor = (threshold: number = 50) => {
  const memoryInfo = useMemo(() => {
    if ('memory' in performance) {
      const memory = (performance as any).memory;
      return {
        used: Math.round(memory.usedJSHeapSize / 1024 / 1024),
        total: Math.round(memory.totalJSHeapSize / 1024 / 1024),
        limit: Math.round(memory.jsHeapSizeLimit / 1024 / 1024),
        percentage: Math.round((memory.usedJSHeapSize / memory.totalJSHeapSize) * 100),
      };
    }
    return null;
  }, []);

  const isOverThreshold = useMemo(() => {
    return memoryInfo ? memoryInfo.used > threshold : false;
  }, [memoryInfo, threshold]);

  return { memoryInfo, isOverThreshold };
};

/**
 * 组件渲染次数监控Hook
 * @param componentName 组件名称
 */
export const useRenderCount = (componentName: string) => {
  const renderCount = useRef(0);
  
  useEffect(() => {
    renderCount.current += 1;
    if (process.env.NODE_ENV === 'development') {
      console.log(`${componentName} rendered ${renderCount.current} times`);
    }
  });

  return renderCount.current;
};

/**
 * 批量状态更新Hook
 * @param initialState 初始状态
 * @returns [state, batchUpdate]
 */
export const useBatchState = <T extends Record<string, any>>(initialState: T) => {
  const [state, setState] = useState(initialState);
  const pendingUpdates = useRef<Partial<T>>({});
  const timeoutRef = useRef<NodeJS.Timeout>();

  const batchUpdate = useCallback((updates: Partial<T>) => {
    pendingUpdates.current = { ...pendingUpdates.current, ...updates };
    
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }
    
    timeoutRef.current = setTimeout(() => {
      setState(prevState => ({ ...prevState, ...pendingUpdates.current }));
      pendingUpdates.current = {};
    }, 0);
  }, []);

  return [state, batchUpdate] as const;
};

/**
 * 图片懒加载Hook
 * @param src 图片源
 * @param placeholder 占位图
 * @returns 图片加载状态和源
 */
export const useLazyImage = (src: string, placeholder?: string) => {
  const [imageSrc, setImageSrc] = useState(placeholder || '');
  const [isLoaded, setIsLoaded] = useState(false);
  const [isError, setIsError] = useState(false);
  const imgRef = useRef<HTMLImageElement>();

  useEffect(() => {
    const img = new Image();
    img.onload = () => {
      setImageSrc(src);
      setIsLoaded(true);
      setIsError(false);
    };
    img.onerror = () => {
      setIsError(true);
      setIsLoaded(false);
    };
    img.src = src;
    imgRef.current = img;

    return () => {
      if (imgRef.current) {
        imgRef.current.onload = null;
        imgRef.current.onerror = null;
      }
    };
  }, [src]);

  return { imageSrc, isLoaded, isError };
};

/**
 * 组件可见性检测Hook
 * @param threshold 可见性阈值
 * @returns [ref, isVisible]
 */
export const useIntersectionObserver = (threshold: number = 0.1) => {
  const [isVisible, setIsVisible] = useState(false);
  const ref = useRef<HTMLElement>(null);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        setIsVisible(entry.isIntersecting);
      },
      { threshold }
    );

    if (ref.current) {
      observer.observe(ref.current);
    }

    return () => {
      if (ref.current) {
        observer.unobserve(ref.current);
      }
    };
  }, [threshold]);

  return [ref, isVisible] as const;
};

export default {
  useDebounce,
  useThrottle,
  useDeepMemo,
  useVirtualScroll,
  useMemoryMonitor,
  useRenderCount,
  useBatchState,
  useLazyImage,
  useIntersectionObserver,
};