import { apiClient } from '../services/api';

export function logClientEvent(level: string, message: string, module?: string, extra?: any): Promise<void> {
  return apiClient
    .createSystemLog({ level, message, module, extra })
    .then(() => void 0)
    .catch(() => void 0);
}

export function initClientLogging(): void {
  // 捕获未处理的运行时错误
  window.addEventListener('error', (event: ErrorEvent) => {
    const err: any = event.error || event.message;
    const message = typeof err === 'string' ? err : (err?.message || 'Uncaught error');
    const extra: Record<string, any> = {
      stack: err?.stack,
      filename: (event as any).filename,
      lineno: (event as any).lineno,
      colno: (event as any).colno,
      url: window.location.href,
      userAgent: navigator.userAgent,
    };
    logClientEvent('error', message, 'frontend', extra);
  });

  // 捕获未处理的 Promise 拒绝
  window.addEventListener('unhandledrejection', (event: PromiseRejectionEvent) => {
    const reason: any = event.reason;
    let message = 'Unhandled promise rejection';
    const extra: Record<string, any> = {};

    if (reason instanceof Error) {
      message = reason.message;
      extra.stack = reason.stack;
    } else if (typeof reason === 'string') {
      message = reason;
    } else {
      extra.reason = reason;
    }

    extra.url = window.location.href;
    extra.userAgent = navigator.userAgent;

    logClientEvent('error', message, 'frontend', extra);
  });
}