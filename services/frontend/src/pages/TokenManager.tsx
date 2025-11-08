import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import './TokenManager.css';

interface TokenInfo {
  id: string;
  createdAt: string;
  expiresAt: string;
  lastUsedAt?: string;
  deviceInfo?: string;
  ipAddress?: string;
}

const TokenManagerPage: React.FC = () => {
  const { user, token, revokeToken } = useAuth();
  const [tokens, setTokens] = useState<TokenInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [revokeLoading, setRevokeLoading] = useState<string | null>(null);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  useEffect(() => {
    // 模拟获取令牌列表
    const fetchTokens = async () => {
      try {
        // 这里应该调用API获取用户的令牌列表
        // const response = await tokenApi.getUserTokens();
        // setTokens(response.data);
        
        // 模拟数据
        setTokens([
          {
            id: '1',
            createdAt: new Date(Date.now() - 8 * 24 * 60 * 60 * 1000).toISOString(),
            expiresAt: new Date(Date.now() + 22 * 24 * 60 * 60 * 1000).toISOString(),
            lastUsedAt: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
            deviceInfo: 'Chrome on macOS',
            ipAddress: '192.168.1.100'
          },
          {
            id: '2',
            createdAt: new Date(Date.now() - 15 * 24 * 60 * 60 * 1000).toISOString(),
            expiresAt: new Date(Date.now() + 15 * 24 * 60 * 60 * 1000).toISOString(),
            lastUsedAt: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString(),
            deviceInfo: 'Safari on iPhone',
            ipAddress: '192.168.1.101'
          }
        ]);
      } catch (error) {
        console.error('获取令牌列表失败:', error);
        showMessage('获取令牌列表失败', 'error');
      } finally {
        setLoading(false);
      }
    };

    fetchTokens();
  }, []);

  const showMessage = (text: string, type: 'success' | 'error') => {
    setMessage({ text, type });
    setTimeout(() => setMessage(null), 3000);
  };

  const handleRevokeToken = async (tokenId: string) => {
    setRevokeLoading(tokenId);
    try {
      // 这里应该调用API撤销特定的令牌
      // await tokenApi.revokeToken(tokenId);
      
      // 对于当前令牌，使用AuthContext的revokeToken方法
      if (tokens.length === 1 || tokenId === '1') {
        const success = await revokeToken('用户主动撤销');
        if (success) {
          showMessage('令牌已撤销，您将被重定向到登录页', 'success');
          setTimeout(() => {
            window.location.href = '/login';
          }, 2000);
        } else {
          showMessage('撤销令牌失败', 'error');
        }
      } else {
        // 模拟撤销其他令牌
        setTokens(tokens.filter(t => t.id !== tokenId));
        showMessage('令牌已撤销', 'success');
      }
    } catch (error) {
      console.error('撤销令牌失败:', error);
      showMessage('撤销令牌失败', 'error');
    } finally {
      setRevokeLoading(null);
    }
  };

  const handleRevokeAllTokens = async () => {
    if (!window.confirm('确定要撤销所有令牌吗？这将在所有设备上登出您的账户。')) {
      return;
    }

    setRevokeLoading('all');
    try {
      // 这里应该调用API撤销所有令牌
      // await tokenApi.revokeAllTokens();
      
      const success = await revokeToken('用户撤销所有令牌');
      if (success) {
        showMessage('所有令牌已撤销，您将被重定向到登录页', 'success');
        setTimeout(() => {
          window.location.href = '/login';
        }, 2000);
      } else {
        showMessage('撤销所有令牌失败', 'error');
      }
    } catch (error) {
      console.error('撤销所有令牌失败:', error);
      showMessage('撤销所有令牌失败', 'error');
    } finally {
      setRevokeLoading(null);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const isCurrentToken = (tokenId: string) => {
    // 这里应该比较当前令牌的ID，我们假设第一个令牌是当前令牌
    return tokenId === '1';
  };

  if (loading) {
    return <div className="token-manager-loading">加载中...</div>;
  }

  return (
    <div className="token-manager-container">
      <div className="token-manager-header">
        <h1>令牌管理</h1>
        <p>管理您的访问令牌，查看活动会话并撤销不需要的令牌</p>
      </div>

      {message && (
        <div className={`token-manager-message ${message.type}`}>
          {message.text}
        </div>
      )}

      <div className="token-manager-actions">
        <button 
          className="token-manager-button revoke-all"
          onClick={handleRevokeAllTokens}
          disabled={revokeLoading === 'all'}
        >
          {revokeLoading === 'all' ? '处理中...' : '撤销所有令牌'}
        </button>
      </div>

      <div className="token-manager-list">
        <h2>活动令牌</h2>
        {tokens.length === 0 ? (
          <p className="token-manager-empty">没有活动令牌</p>
        ) : (
          <div className="token-manager-items">
            {tokens.map((tokenItem) => (
              <div key={tokenItem.id} className="token-manager-item">
                <div className="token-manager-info">
                  <div className="token-manager-header-item">
                    <h3>
                      {isCurrentToken(tokenItem.id) ? '当前会话' : `令牌 #${tokenItem.id}`}
                      {isCurrentToken(tokenItem.id) && (
                        <span className="token-manager-current">当前</span>
                      )}
                    </h3>
                    <div className="token-manager-dates">
                      <div>创建时间: {formatDate(tokenItem.createdAt)}</div>
                      <div>过期时间: {formatDate(tokenItem.expiresAt)}</div>
                      {tokenItem.lastUsedAt && (
                        <div>最后使用: {formatDate(tokenItem.lastUsedAt)}</div>
                      )}
                    </div>
                  </div>
                  <div className="token-manager-details">
                    {tokenItem.deviceInfo && (
                      <div className="token-manager-detail">
                        <span className="token-manager-label">设备:</span>
                        <span>{tokenItem.deviceInfo}</span>
                      </div>
                    )}
                    {tokenItem.ipAddress && (
                      <div className="token-manager-detail">
                        <span className="token-manager-label">IP地址:</span>
                        <span>{tokenItem.ipAddress}</span>
                      </div>
                    )}
                  </div>
                </div>
                <div className="token-manager-actions-item">
                  <button
                    className="token-manager-button revoke"
                    onClick={() => handleRevokeToken(tokenItem.id)}
                    disabled={revokeLoading === tokenItem.id || revokeLoading === 'all'}
                  >
                    {revokeLoading === tokenItem.id ? '处理中...' : '撤销'}
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="token-manager-security-info">
        <h2>安全提示</h2>
        <ul>
          <li>定期检查您的活动令牌，确保没有未经授权的访问</li>
          <li>如果您在公共设备上登录，请记得及时撤销令牌</li>
          <li>如果您怀疑账户被盗，请立即撤销所有令牌</li>
          <li>令牌会自动过期，但建议定期更换密码以确保安全</li>
        </ul>
      </div>
    </div>
  );
};

export default TokenManagerPage;