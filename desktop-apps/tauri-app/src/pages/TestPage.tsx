import React, { useState } from 'react';
import { invoke } from '@tauri-apps/api/core';

interface TestResult {
  endpoint: string;
  status: 'pending' | 'success' | 'error';
  response?: any;
  error?: string;
  duration?: number;
}

const TestPage: React.FC = () => {
  const [testResults, setTestResults] = useState<TestResult[]>([]);
  const [isRunning, setIsRunning] = useState(false);

  const endpoints = [
    { name: '健康检查', command: 'health_check' },
    { name: '用户模块', command: 'get_user_modules' },
    { name: '用户偏好', command: 'get_user_preferences' }
  ];

  const testEndpoint = async (endpoint: { name: string; command: string }): Promise<TestResult> => {
    const startTime = Date.now();
    
    try {
      const response = await invoke(endpoint.command);
      const duration = Date.now() - startTime;
      
      return {
        endpoint: endpoint.name,
        status: 'success',
        response: JSON.stringify(response, null, 2),
        duration
      };
    } catch (error) {
      const duration = Date.now() - startTime;
      return {
        endpoint: endpoint.name,
        status: 'error',
        error: error instanceof Error ? error.message : String(error),
        duration
      };
    }
  };

  const runAllTests = async () => {
    setIsRunning(true);
    setTestResults([]);
    
    const results: TestResult[] = [];
    
    for (const endpoint of endpoints) {
      // 添加待测试状态
      const pendingResult: TestResult = {
        endpoint: endpoint.name,
        status: 'pending'
      };
      results.push(pendingResult);
      setTestResults([...results]);
      
      // 执行测试
      const result = await testEndpoint(endpoint);
      results[results.length - 1] = result;
      setTestResults([...results]);
    }
    
    setIsRunning(false);
  };

  const getStatusIcon = (status: TestResult['status']) => {
    switch (status) {
      case 'pending':
        return '⏳';
      case 'success':
        return '✅';
      case 'error':
        return '❌';
      default:
        return '❓';
    }
  };

  const getStatusColor = (status: TestResult['status']) => {
    switch (status) {
      case 'pending':
        return 'text-yellow-600';
      case 'success':
        return 'text-green-600';
      case 'error':
        return 'text-red-600';
      default:
        return 'text-gray-600';
    }
  };

  return (
    <div className="p-4 sm:p-6 max-w-4xl mx-auto">
      <div className="bg-white rounded-lg shadow-lg p-4 sm:p-6">
        <h1 className="text-xl sm:text-2xl font-bold text-gray-800 mb-6">
          前后端连接测试
        </h1>
        
        <div className="mb-6">
          <button
            onClick={runAllTests}
            disabled={isRunning}
            className={`w-full sm:w-auto px-6 py-3 rounded-lg font-medium ${
              isRunning
                ? 'bg-gray-400 cursor-not-allowed'
                : 'bg-blue-600 hover:bg-blue-700'
            } text-white transition-colors`}
          >
            {isRunning ? '测试中...' : '开始测试'}
          </button>
        </div>

        <div className="space-y-4">
          {testResults.map((result, index) => (
            <div
              key={index}
              className="border border-gray-200 rounded-lg p-4 bg-gray-50"
            >
              <div className="flex flex-col sm:flex-row sm:items-center justify-between mb-2 gap-2">
                <div className="flex items-center space-x-3">
                  <span className="text-2xl">{getStatusIcon(result.status)}</span>
                  <h3 className="font-semibold text-gray-800">
                    {result.endpoint}
                  </h3>
                  <span className={`font-medium ${getStatusColor(result.status)}`}>
                    {result.status === 'pending' ? '测试中' : 
                     result.status === 'success' ? '成功' : '失败'}
                  </span>
                </div>
                {result.duration && (
                  <span className="text-sm text-gray-500">
                    {result.duration}ms
                  </span>
                )}
              </div>
              
              {result.response && (
                <div className="mt-3">
                  <h4 className="font-medium text-gray-700 mb-2">响应内容:</h4>
                  <pre className="bg-green-50 border border-green-200 rounded p-3 text-sm text-gray-800 overflow-x-auto">
                    {result.response}
                  </pre>
                </div>
              )}
              
              {result.error && (
                <div className="mt-3">
                  <h4 className="font-medium text-gray-700 mb-2">错误信息:</h4>
                  <pre className="bg-red-50 border border-red-200 rounded p-3 text-sm text-red-800 overflow-x-auto">
                    {result.error}
                  </pre>
                </div>
              )}
            </div>
          ))}
        </div>

        {testResults.length === 0 && (
          <div className="text-center py-8 text-gray-500">
            点击"开始测试"按钮来测试前后端连接
          </div>
        )}
      </div>
    </div>
  );
};

export default TestPage;