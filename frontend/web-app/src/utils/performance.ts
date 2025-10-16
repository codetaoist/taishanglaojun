// 防抖函数
export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number,
  immediate?: boolean
): (...args: Parameters<T>) => void {
  let timeout: NodeJS.Timeout | null = null
  
  return function executedFunction(...args: Parameters<T>) {
    const later = () => {
      timeout = null
      if (!immediate) func(...args)
    }
    
    const callNow = immediate && !timeout
    
    if (timeout) clearTimeout(timeout)
    timeout = setTimeout(later, wait)
    
    if (callNow) func(...args)
  }
}

// 节流函数
export function throttle<T extends (...args: any[]) => any>(
  func: T,
  limit: number
): (...args: Parameters<T>) => void {
  let inThrottle: boolean = false
  
  return function executedFunction(...args: Parameters<T>) {
    if (!inThrottle) {
      func(...args)
      inThrottle = true
      setTimeout(() => inThrottle = false, limit)
    }
  }
}

// 请求动画帧节流
export function rafThrottle<T extends (...args: any[]) => any>(
  func: T
): (...args: Parameters<T>) => void {
  let rafId: number | null = null
  
  return function executedFunction(...args: Parameters<T>) {
    if (rafId) return
    
    rafId = requestAnimationFrame(() => {
      func(...args)
      rafId = null
    })
  }
}

// 空闲时执行
export function runWhenIdle<T extends (...args: any[]) => any>(
  func: T,
  timeout: number = 5000
): (...args: Parameters<T>) => void {
  return function executedFunction(...args: Parameters<T>) {
    if ('requestIdleCallback' in window) {
      requestIdleCallback(() => func(...args), { timeout })
    } else {
      setTimeout(() => func(...args), 0)
    }
  }
}

// 批量执行
export class BatchProcessor<T> {
  private batch: T[] = []
  private processor: (items: T[]) => void
  private batchSize: number
  private delay: number
  private timeoutId: NodeJS.Timeout | null = null

  constructor(
    processor: (items: T[]) => void,
    batchSize: number = 10,
    delay: number = 100
  ) {
    this.processor = processor
    this.batchSize = batchSize
    this.delay = delay
  }

  add(item: T): void {
    this.batch.push(item)
    
    if (this.batch.length >= this.batchSize) {
      this.flush()
    } else {
      this.scheduleFlush()
    }
  }

  private scheduleFlush(): void {
    if (this.timeoutId) return
    
    this.timeoutId = setTimeout(() => {
      this.flush()
    }, this.delay)
  }

  flush(): void {
    if (this.timeoutId) {
      clearTimeout(this.timeoutId)
      this.timeoutId = null
    }
    
    if (this.batch.length > 0) {
      const items = [...this.batch]
      this.batch = []
      this.processor(items)
    }
  }

  clear(): void {
    if (this.timeoutId) {
      clearTimeout(this.timeoutId)
      this.timeoutId = null
    }
    this.batch = []
  }
}

// 内存管理
export class MemoryManager {
  private static instance: MemoryManager
  private observers: Set<() => void> = new Set()
  private memoryThreshold: number = 50 * 1024 * 1024 // 50MB

  static getInstance(): MemoryManager {
    if (!MemoryManager.instance) {
      MemoryManager.instance = new MemoryManager()
    }
    return MemoryManager.instance
  }

  // 监听内存使用
  startMonitoring(): void {
    if ('memory' in performance) {
      const checkMemory = () => {
        const memory = (performance as any).memory
        if (memory.usedJSHeapSize > this.memoryThreshold) {
          this.notifyObservers()
        }
      }
      
      setInterval(checkMemory, 10000) // 每10秒检查一次
    }
  }

  // 添加内存压力观察者
  addObserver(callback: () => void): void {
    this.observers.add(callback)
  }

  // 移除观察者
  removeObserver(callback: () => void): void {
    this.observers.delete(callback)
  }

  // 通知观察者
  private notifyObservers(): void {
    this.observers.forEach(callback => callback())
  }

  // 获取内存信息
  getMemoryInfo(): any {
    if ('memory' in performance) {
      return (performance as any).memory
    }
    return null
  }

  // 强制垃圾回收（仅在开发环境）
  forceGC(): void {
    if (process.env.NODE_ENV === 'development' && 'gc' in window) {
      (window as any).gc()
    }
  }
}

// 图片懒加载
export class LazyImageLoader {
  private observer: IntersectionObserver | null = null
  private images: Set<HTMLImageElement> = new Set()

  constructor(options?: IntersectionObserverInit) {
    if ('IntersectionObserver' in window) {
      this.observer = new IntersectionObserver(
        this.handleIntersection.bind(this),
        {
          rootMargin: '50px',
          threshold: 0.1,
          ...options,
        }
      )
    }
  }

  observe(img: HTMLImageElement): void {
    if (this.observer) {
      this.images.add(img)
      this.observer.observe(img)
    } else {
      // 降级处理
      this.loadImage(img)
    }
  }

  unobserve(img: HTMLImageElement): void {
    if (this.observer) {
      this.images.delete(img)
      this.observer.unobserve(img)
    }
  }

  private handleIntersection(entries: IntersectionObserverEntry[]): void {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        const img = entry.target as HTMLImageElement
        this.loadImage(img)
        this.unobserve(img)
      }
    })
  }

  private loadImage(img: HTMLImageElement): void {
    const src = img.dataset.src
    if (src) {
      img.src = src
      img.removeAttribute('data-src')
    }
  }

  disconnect(): void {
    if (this.observer) {
      this.observer.disconnect()
      this.images.clear()
    }
  }
}

// 虚拟滚动
export interface VirtualScrollOptions {
  itemHeight: number
  containerHeight: number
  overscan?: number
}

export class VirtualScroll {
  private options: Required<VirtualScrollOptions>
  private scrollTop: number = 0
  private totalItems: number = 0

  constructor(options: VirtualScrollOptions) {
    this.options = {
      overscan: 5,
      ...options,
    }
  }

  updateScrollTop(scrollTop: number): void {
    this.scrollTop = scrollTop
  }

  updateTotalItems(totalItems: number): void {
    this.totalItems = totalItems
  }

  getVisibleRange(): { start: number; end: number } {
    const { itemHeight, containerHeight, overscan } = this.options
    
    const start = Math.max(0, Math.floor(this.scrollTop / itemHeight) - overscan)
    const visibleCount = Math.ceil(containerHeight / itemHeight)
    const end = Math.min(this.totalItems - 1, start + visibleCount + overscan * 2)
    
    return { start, end }
  }

  getTotalHeight(): number {
    return this.totalItems * this.options.itemHeight
  }

  getOffsetY(): number {
    const { start } = this.getVisibleRange()
    return start * this.options.itemHeight
  }
}

// 资源预加载
export class ResourcePreloader {
  private cache: Map<string, Promise<any>> = new Map()

  // 预加载图片
  preloadImage(src: string): Promise<HTMLImageElement> {
    if (this.cache.has(src)) {
      return this.cache.get(src)!
    }

    const promise = new Promise<HTMLImageElement>((resolve, reject) => {
      const img = new Image()
      img.onload = () => resolve(img)
      img.onerror = reject
      img.src = src
    })

    this.cache.set(src, promise)
    return promise
  }

  // 预加载多个图片
  preloadImages(srcs: string[]): Promise<HTMLImageElement[]> {
    return Promise.all(srcs.map(src => this.preloadImage(src)))
  }

  // 预加载脚本
  preloadScript(src: string): Promise<HTMLScriptElement> {
    if (this.cache.has(src)) {
      return this.cache.get(src)!
    }

    const promise = new Promise<HTMLScriptElement>((resolve, reject) => {
      const script = document.createElement('script')
      script.onload = () => resolve(script)
      script.onerror = reject
      script.src = src
      document.head.appendChild(script)
    })

    this.cache.set(src, promise)
    return promise
  }

  // 预加载样式
  preloadStyle(href: string): Promise<HTMLLinkElement> {
    if (this.cache.has(href)) {
      return this.cache.get(href)!
    }

    const promise = new Promise<HTMLLinkElement>((resolve, reject) => {
      const link = document.createElement('link')
      link.rel = 'stylesheet'
      link.onload = () => resolve(link)
      link.onerror = reject
      link.href = href
      document.head.appendChild(link)
    })

    this.cache.set(href, promise)
    return promise
  }

  // 预加载数据
  preloadData(url: string, options?: RequestInit): Promise<Response> {
    const key = `${url}:${JSON.stringify(options)}`
    
    if (this.cache.has(key)) {
      return this.cache.get(key)!
    }

    const promise = fetch(url, options)
    this.cache.set(key, promise)
    return promise
  }

  // 清除缓存
  clearCache(): void {
    this.cache.clear()
  }

  // 获取缓存大小
  getCacheSize(): number {
    return this.cache.size
  }
}

// 性能监控
export class PerformanceMonitor {
  private static instance: PerformanceMonitor
  private metrics: Map<string, number[]> = new Map()
  private observers: Set<(metric: string, value: number) => void> = new Set()

  static getInstance(): PerformanceMonitor {
    if (!PerformanceMonitor.instance) {
      PerformanceMonitor.instance = new PerformanceMonitor()
    }
    return PerformanceMonitor.instance
  }

  // 记录性能指标
  recordMetric(name: string, value: number): void {
    if (!this.metrics.has(name)) {
      this.metrics.set(name, [])
    }
    
    const values = this.metrics.get(name)!
    values.push(value)
    
    // 只保留最近100个值
    if (values.length > 100) {
      values.shift()
    }
    
    this.notifyObservers(name, value)
  }

  // 测量函数执行时间
  measureFunction<T extends (...args: any[]) => any>(
    name: string,
    func: T
  ): (...args: Parameters<T>) => ReturnType<T> {
    return (...args: Parameters<T>): ReturnType<T> => {
      const start = performance.now()
      const result = func(...args)
      const end = performance.now()
      
      this.recordMetric(name, end - start)
      
      return result
    }
  }

  // 测量异步函数执行时间
  measureAsyncFunction<T extends (...args: any[]) => Promise<any>>(
    name: string,
    func: T
  ): (...args: Parameters<T>) => ReturnType<T> {
    return async (...args: Parameters<T>): Promise<any> => {
      const start = performance.now()
      const result = await func(...args)
      const end = performance.now()
      
      this.recordMetric(name, end - start)
      
      return result
    }
  }

  // 获取指标统计
  getMetricStats(name: string): {
    count: number
    min: number
    max: number
    avg: number
    median: number
  } | null {
    const values = this.metrics.get(name)
    if (!values || values.length === 0) {
      return null
    }
    
    const sorted = [...values].sort((a, b) => a - b)
    const count = values.length
    const min = sorted[0]
    const max = sorted[count - 1]
    const avg = values.reduce((sum, val) => sum + val, 0) / count
    const median = count % 2 === 0
      ? (sorted[count / 2 - 1] + sorted[count / 2]) / 2
      : sorted[Math.floor(count / 2)]
    
    return { count, min, max, avg, median }
  }

  // 添加观察者
  addObserver(callback: (metric: string, value: number) => void): void {
    this.observers.add(callback)
  }

  // 移除观察者
  removeObserver(callback: (metric: string, value: number) => void): void {
    this.observers.delete(callback)
  }

  // 通知观察者
  private notifyObservers(metric: string, value: number): void {
    this.observers.forEach(callback => callback(metric, value))
  }

  // 获取所有指标
  getAllMetrics(): Map<string, number[]> {
    return new Map(this.metrics)
  }

  // 清除指标
  clearMetrics(): void {
    this.metrics.clear()
  }
}

// Web Workers 管理
export class WorkerManager {
  private workers: Map<string, Worker> = new Map()
  private workerPool: Map<string, Worker[]> = new Map()

  // 创建 Worker
  createWorker(name: string, scriptUrl: string): Worker {
    const worker = new Worker(scriptUrl)
    this.workers.set(name, worker)
    return worker
  }

  // 获取 Worker
  getWorker(name: string): Worker | undefined {
    return this.workers.get(name)
  }

  // 创建 Worker 池
  createWorkerPool(name: string, scriptUrl: string, poolSize: number = 4): Worker[] {
    const workers: Worker[] = []
    
    for (let i = 0; i < poolSize; i++) {
      workers.push(new Worker(scriptUrl))
    }
    
    this.workerPool.set(name, workers)
    return workers
  }

  // 获取可用的 Worker
  getAvailableWorker(poolName: string): Worker | undefined {
    const pool = this.workerPool.get(poolName)
    return pool ? pool[0] : undefined // 简单的轮询策略
  }

  // 执行任务
  executeTask<T>(
    workerName: string,
    data: any,
    transferable?: Transferable[]
  ): Promise<T> {
    return new Promise((resolve, reject) => {
      const worker = this.getWorker(workerName)
      if (!worker) {
        reject(new Error(`Worker ${workerName} not found`))
        return
      }

      const handleMessage = (event: MessageEvent) => {
        worker.removeEventListener('message', handleMessage)
        worker.removeEventListener('error', handleError)
        resolve(event.data)
      }

      const handleError = (error: ErrorEvent) => {
        worker.removeEventListener('message', handleMessage)
        worker.removeEventListener('error', handleError)
        reject(error)
      }

      worker.addEventListener('message', handleMessage)
      worker.addEventListener('error', handleError)
      worker.postMessage(data, transferable || [])
    })
  }

  // 终止 Worker
  terminateWorker(name: string): void {
    const worker = this.workers.get(name)
    if (worker) {
      worker.terminate()
      this.workers.delete(name)
    }
  }

  // 终止 Worker 池
  terminateWorkerPool(name: string): void {
    const pool = this.workerPool.get(name)
    if (pool) {
      pool.forEach(worker => worker.terminate())
      this.workerPool.delete(name)
    }
  }

  // 终止所有 Workers
  terminateAll(): void {
    this.workers.forEach(worker => worker.terminate())
    this.workerPool.forEach(pool => pool.forEach(worker => worker.terminate()))
    this.workers.clear()
    this.workerPool.clear()
  }
}

// 导出实例
export const memoryManager = MemoryManager.getInstance()
export const performanceMonitor = PerformanceMonitor.getInstance()
export const resourcePreloader = new ResourcePreloader()
export const workerManager = new WorkerManager()

export default {
  debounce,
  throttle,
  rafThrottle,
  runWhenIdle,
  BatchProcessor,
  MemoryManager,
  LazyImageLoader,
  VirtualScroll,
  ResourcePreloader,
  PerformanceMonitor,
  WorkerManager,
  memoryManager,
  performanceMonitor,
  resourcePreloader,
  workerManager,
}