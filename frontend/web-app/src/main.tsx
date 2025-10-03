// 导入 Ant Design React 19 兼容补丁 - 必须在最前面
import '@ant-design/v5-patch-for-react-19'

import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { ErrorBoundary } from 'react-error-boundary'
import { App as AntdApp } from 'antd'
import App from './App.tsx'
import './index.css'
import { AuthProvider } from './hooks/useAuth.tsx'

function ErrorFallback({error}: {error: Error}) {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full bg-white rounded-lg shadow-md p-6">
        <h2 className="text-xl font-semibold text-red-600 mb-4">出现错误</h2>
        <p className="text-gray-600 mb-4">应用程序遇到了一个错误，请刷新页面重试。</p>
        <details className="mb-4">
          <summary className="cursor-pointer text-sm text-gray-500">错误详情</summary>
          <pre className="mt-2 text-xs text-gray-400 bg-gray-100 p-2 rounded overflow-auto">
            {error.message}
          </pre>
        </details>
        <button 
          onClick={() => window.location.reload()} 
          className="w-full bg-blue-500 text-white py-2 px-4 rounded hover:bg-blue-600 transition-colors"
        >
          刷新页面
        </button>
      </div>
    </div>
  )
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ErrorBoundary FallbackComponent={ErrorFallback}>
      <AntdApp>
        <AuthProvider>
          <App />
        </AuthProvider>
      </AntdApp>
    </ErrorBoundary>
  </StrictMode>,
)
