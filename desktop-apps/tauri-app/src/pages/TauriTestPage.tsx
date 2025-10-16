import { useState } from 'react';
import { invoke } from '@tauri-apps/api/core';
import { apiService } from '../services/api';

export default function TauriTestPage() {
  const [results, setResults] = useState<Record<string, any>>({});
  const [loading, setLoading] = useState<Record<string, boolean>>({});

  const testCommand = async (commandName: string, params: any = {}) => {
    setLoading(prev => ({ ...prev, [commandName]: true }));
    try {
      const result = await invoke(commandName, params);
      setResults(prev => ({ ...prev, [commandName]: { success: true, data: result } }));
    } catch (error) {
      setResults(prev => ({ ...prev, [commandName]: { success: false, error: String(error) } }));
    } finally {
      setLoading(prev => ({ ...prev, [commandName]: false }));
    }
  };

  const testApiService = async (serviceName: string, serviceFunction: () => Promise<any>) => {
    setLoading(prev => ({ ...prev, [serviceName]: true }));
    try {
      const result = await serviceFunction();
      setResults(prev => ({ ...prev, [serviceName]: { success: true, data: result } }));
    } catch (error) {
      setResults(prev => ({ ...prev, [serviceName]: { success: false, error: String(error) } }));
    } finally {
      setLoading(prev => ({ ...prev, [serviceName]: false }));
    }
  };

  const commands = [
    { name: 'health_check', params: {} },
    { name: 'get_user_modules', params: {} },
    { name: 'validate_session', params: {} },
    { name: 'get_user_info', params: {} },
  ];

  const apiTests = [
    { name: 'getUserModules', function: () => apiService.getUserModules() },
    { name: 'getUserPreferences', function: () => apiService.getUserPreferences() },
    { name: 'updateUserPreferences', function: () => apiService.updateUserPreferences({ theme: 'dark' }) },
  ];

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">Tauri 命令和API测试</h1>
      
      {/* Tauri 命令测试 */}
      <div className="mb-8">
        <h2 className="text-xl font-semibold mb-4">Tauri 命令测试</h2>
        <div className="grid gap-4">
          {commands.map(({ name, params }) => (
            <div key={name} className="border rounded-lg p-4">
              <div className="flex items-center justify-between mb-2">
                <h3 className="font-semibold">{name}</h3>
                <button
                  onClick={() => testCommand(name, params)}
                  disabled={loading[name]}
                  className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
                >
                  {loading[name] ? '测试中...' : '测试'}
                </button>
              </div>
              
              {results[name] && (
                <div className={`mt-2 p-3 rounded ${
                  results[name].success ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                }`}>
                  <div className="font-medium">
                    {results[name].success ? '✅ 成功' : '❌ 失败'}
                  </div>
                  <pre className="mt-1 text-sm overflow-auto">
                    {JSON.stringify(results[name].success ? results[name].data : results[name].error, null, 2)}
                  </pre>
                </div>
              )}
            </div>
          ))}
        </div>
        
        <div className="mt-4">
          <button
            onClick={() => {
              commands.forEach(({ name, params }) => testCommand(name, params));
            }}
            className="px-6 py-3 bg-green-500 text-white rounded-lg hover:bg-green-600"
          >
            测试所有Tauri命令
          </button>
        </div>
      </div>

      {/* API 服务测试 */}
      <div>
        <h2 className="text-xl font-semibold mb-4">API 服务测试</h2>
        <div className="grid gap-4">
          {apiTests.map(({ name, function: testFunction }) => (
            <div key={name} className="border rounded-lg p-4">
              <div className="flex items-center justify-between mb-2">
                <h3 className="font-semibold">{name}</h3>
                <button
                  onClick={() => testApiService(name, testFunction)}
                  disabled={loading[name]}
                  className="px-4 py-2 bg-purple-500 text-white rounded hover:bg-purple-600 disabled:opacity-50"
                >
                  {loading[name] ? '测试中...' : '测试'}
                </button>
              </div>
              
              {results[name] && (
                <div className={`mt-2 p-3 rounded ${
                  results[name].success ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                }`}>
                  <div className="font-medium">
                    {results[name].success ? '✅ 成功' : '❌ 失败'}
                  </div>
                  <pre className="mt-1 text-sm overflow-auto">
                    {JSON.stringify(results[name].success ? results[name].data : results[name].error, null, 2)}
                  </pre>
                </div>
              )}
            </div>
          ))}
        </div>
        
        <div className="mt-4">
          <button
            onClick={() => {
              apiTests.forEach(({ name, function: testFunction }) => testApiService(name, testFunction));
            }}
            className="px-6 py-3 bg-purple-500 text-white rounded-lg hover:bg-purple-600"
          >
            测试所有API服务
          </button>
        </div>
      </div>
    </div>
  );
}