import '@ant-design/v5-patch-for-react-19';
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { Provider } from 'react-redux';
import { PersistGate } from 'redux-persist/integration/react';
import { App as AntdApp } from 'antd';
import { ErrorBoundary } from 'react-error-boundary';
import { store, persistor } from './store';
import { AuthProvider } from './contexts/AuthContext';
import { MenuProvider } from './contexts/MenuContext';
import App from './App';
import './i18n';
import './index.css';
import { initClientLogging, logClientEvent } from './utils/clientLogging';

// 错误回退组件
function ErrorFallback({ error, resetErrorBoundary }: { error: Error; resetErrorBoundary: () => void }) {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-6 text-center">
        <div className="text-red-500 text-6xl mb-4">⚠️</div>
        <h2 className="text-xl font-semibold text-gray-900 mb-2">应用出现错误</h2>
        <p className="text-gray-600 mb-4">很抱歉，应用遇到了一个意外错误。</p>
        <details className="text-left mb-4">
          <summary className="cursor-pointer text-sm text-gray-500 hover:text-gray-700">
            查看错误详情
          </summary>
          <pre className="mt-2 text-xs text-red-600 bg-red-50 p-2 rounded overflow-auto max-h-32">
            {error.message}
          </pre>
        </details>
        <button
          onClick={resetErrorBoundary}
          className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded transition-colors"
        >
          重新加载应用
        </button>
      </div>
    </div>
  );
}

// 加载组件
function LoadingComponent() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="text-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
        <p className="text-gray-600">正在加载应用...</p>
      </div>
    </div>
  );
}

// 初始化客户端日志捕获（在挂载前执行一次）
initClientLogging();

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ErrorBoundary
      FallbackComponent={ErrorFallback}
      onError={(error, errorInfo) => {
        console.error('应用错误:', error, errorInfo);
        // 接入错误上报到系统日志
        logClientEvent('error', error.message, 'react', {
          componentStack: errorInfo.componentStack,
          stack: (error as any)?.stack,
          url: window.location.href,
          userAgent: navigator.userAgent,
        });
      }}
      onReset={() => {
        // 重置应用状态
        window.location.reload();
      }}
    >
      <Provider store={store}>
        <PersistGate loading={<LoadingComponent />} persistor={persistor}>
          <AntdApp>
            <AuthProvider>
              <MenuProvider>
                <App />
              </MenuProvider>
            </AuthProvider>
          </AntdApp>
        </PersistGate>
      </Provider>
    </ErrorBoundary>
  </StrictMode>
);
