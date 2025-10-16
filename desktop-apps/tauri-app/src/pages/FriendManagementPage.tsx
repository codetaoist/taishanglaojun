import React, { useState, useEffect } from 'react';
import {
  UserPlus,
  User,
  Shield,
  Trash2,
  MoreVertical,
  Search,
  RefreshCw,
  Check,
  X,
  MessageCircle,
  Phone,
  Video,
} from 'lucide-react';
import { invoke } from '@tauri-apps/api/core';

interface Friend {
  id: string;
  username: string;
  display_name: string;
  email: string;
  avatar_url?: string;
  status: 'Pending' | 'Accepted' | 'Blocked' | 'Declined';
  online_status: 'Online' | 'Offline' | 'Away' | 'Busy' | 'Invisible';
  last_seen: string;
  mutual_friends_count: number;
  added_at: string;
  updated_at: string;
}

interface FriendRequest {
  id: string;
  from_user_id: string;
  to_user_id: string;
  from_username: string;
  from_display_name: string;
  from_avatar_url?: string;
  message?: string;
  status: 'Pending' | 'Accepted' | 'Declined';
  created_at: string;
  updated_at: string;
}

interface FriendResponse {
  success: boolean;
  message: string;
  friends?: Friend[];
  friend?: Friend;
  requests?: FriendRequest[];
  request?: FriendRequest;
}

interface AddFriendRequest {
  username: string;
  message?: string;
}

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`friend-tabpanel-${index}`}
      aria-labelledby={`friend-tab-${index}`}
      {...other}
    >
      {value === index && <div className="p-6">{children}</div>}
    </div>
  );
}

const FriendManagementPage: React.FC = () => {
  const [tabValue, setTabValue] = useState(0);
  const [friends, setFriends] = useState<Friend[]>([]);
  const [friendRequests, setFriendRequests] = useState<FriendRequest[]>([]);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error' | 'info'; text: string } | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [addFriendDialog, setAddFriendDialog] = useState(false);
  const [addFriendForm, setAddFriendForm] = useState<AddFriendRequest>({
    username: '',
    message: '',
  });
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedFriend, setSelectedFriend] = useState<Friend | null>(null);

  useEffect(() => {
    loadFriends();
    loadFriendRequests();
  }, []);

  const loadFriends = async () => {
    try {
      const authToken = localStorage.getItem('auth_token') || '';
      const response = await invoke<FriendResponse>('friend_get_list', {
        authToken,
      });

      if (response.success && response.friends) {
        setFriends(response.friends);
      } else {
        // 尝试从缓存加载
        const cachedFriends = await invoke<Friend[]>('friend_get_cached_list');
        setFriends(cachedFriends);
      }
    } catch (error) {
      console.error('加载好友列表失败:', error);
      // 尝试从缓存加载
      try {
        const cachedFriends = await invoke<Friend[]>('friend_get_cached_list');
        setFriends(cachedFriends);
      } catch (cacheError) {
        console.error('从缓存加载好友列表失败:', cacheError);
      }
    }
  };

  const loadFriendRequests = async () => {
    try {
      const authToken = localStorage.getItem('auth_token') || '';
      const response = await invoke<FriendResponse>('friend_get_requests', {
        authToken,
      });

      if (response.success && response.requests) {
        setFriendRequests(response.requests);
      } else {
        // 尝试从缓存加载
        const cachedRequests = await invoke<FriendRequest[]>('friend_get_cached_requests');
        setFriendRequests(cachedRequests);
      }
    } catch (error) {
      console.error('加载好友请求失败:', error);
      // 尝试从缓存加载
      try {
        const cachedRequests = await invoke<FriendRequest[]>('friend_get_cached_requests');
        setFriendRequests(cachedRequests);
      } catch (cacheError) {
        console.error('从缓存加载好友请求失败:', cacheError);
      }
    }
  };

  const handleAddFriend = async () => {
    if (!addFriendForm.username.trim()) {
      setMessage({ type: 'error', text: '请输入用户名' });
      return;
    }

    setLoading(true);
    try {
      const authToken = localStorage.getItem('auth_token') || '';
      const response = await invoke<FriendResponse>('friend_send_request', {
        authToken,
        request: addFriendForm,
      });

      if (response.success) {
        setMessage({ type: 'success', text: '好友请求已发送' });
        setAddFriendDialog(false);
        setAddFriendForm({ username: '', message: '' });
        loadFriendRequests();
      } else {
        setMessage({ type: 'error', text: response.message || '发送好友请求失败' });
      }
    } catch (error) {
      console.error('发送好友请求失败:', error);
      setMessage({ type: 'error', text: '发送好友请求失败' });
    } finally {
      setLoading(false);
    }
  };

  const handleAcceptRequest = async (requestId: string) => {
    setLoading(true);
    try {
      const authToken = localStorage.getItem('auth_token') || '';
      const response = await invoke<FriendResponse>('friend_accept_request', {
        authToken,
        requestId,
      });

      if (response.success) {
        setMessage({ type: 'success', text: '已接受好友请求' });
        loadFriends();
        loadFriendRequests();
      } else {
        setMessage({ type: 'error', text: response.message || '接受好友请求失败' });
      }
    } catch (error) {
      console.error('接受好友请求失败:', error);
      setMessage({ type: 'error', text: '接受好友请求失败' });
    } finally {
      setLoading(false);
    }
  };

  const handleDeclineRequest = async (requestId: string) => {
    setLoading(true);
    try {
      const authToken = localStorage.getItem('auth_token') || '';
      const response = await invoke<FriendResponse>('friend_decline_request', {
        authToken,
        requestId,
      });

      if (response.success) {
        setMessage({ type: 'success', text: '已拒绝好友请求' });
        loadFriendRequests();
      } else {
        setMessage({ type: 'error', text: response.message || '拒绝好友请求失败' });
      }
    } catch (error) {
      console.error('拒绝好友请求失败:', error);
      setMessage({ type: 'error', text: '拒绝好友请求失败' });
    } finally {
      setLoading(false);
    }
  };

  const handleRemoveFriend = async (friendId: string) => {
    setLoading(true);
    try {
      const authToken = localStorage.getItem('auth_token') || '';
      const response = await invoke<FriendResponse>('friend_remove', {
        authToken,
        friendId,
      });

      if (response.success) {
        setMessage({ type: 'success', text: '已删除好友' });
        loadFriends();
      } else {
        setMessage({ type: 'error', text: response.message || '删除好友失败' });
      }
    } catch (error) {
      console.error('删除好友失败:', error);
      setMessage({ type: 'error', text: '删除好友失败' });
    } finally {
      setLoading(false);
    }
  };

  const handleBlockFriend = async (friendId: string) => {
    setLoading(true);
    try {
      const authToken = localStorage.getItem('auth_token') || '';
      const response = await invoke<FriendResponse>('friend_block', {
        authToken,
        friendId,
      });

      if (response.success) {
        setMessage({ type: 'success', text: '已屏蔽好友' });
        loadFriends();
      } else {
        setMessage({ type: 'error', text: response.message || '屏蔽好友失败' });
      }
    } catch (error) {
      console.error('屏蔽好友失败:', error);
      setMessage({ type: 'error', text: '屏蔽好友失败' });
    } finally {
      setLoading(false);
    }
  };

  const filteredFriends = friends.filter(friend =>
    friend.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
    friend.display_name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'Online': return 'bg-green-500';
      case 'Away': return 'bg-yellow-500';
      case 'Busy': return 'bg-red-500';
      case 'Offline': return 'bg-gray-500';
      default: return 'bg-gray-500';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'Online': return '在线';
      case 'Away': return '离开';
      case 'Busy': return '忙碌';
      case 'Offline': return '离线';
      default: return '未知';
    }
  };

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 mb-2">好友管理</h1>
        <p className="text-gray-600">管理您的好友列表和好友请求</p>
      </div>

      {message && (
        <div className={`mb-4 p-4 rounded-lg ${
          message.type === 'success' ? 'bg-green-100 text-green-700' :
          message.type === 'error' ? 'bg-red-100 text-red-700' :
          'bg-blue-100 text-blue-700'
        }`}>
          {message.text}
        </div>
      )}

      <div className="bg-white rounded-lg shadow">
        <div className="border-b border-gray-200">
          <nav className="-mb-px flex">
            <button
              onClick={() => setTabValue(0)}
              className={`py-2 px-4 border-b-2 font-medium text-sm ${
                tabValue === 0
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              好友列表 ({friends.length})
            </button>
            <button
              onClick={() => setTabValue(1)}
              className={`py-2 px-4 border-b-2 font-medium text-sm ${
                tabValue === 1
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              好友请求 ({friendRequests.length})
            </button>
          </nav>
        </div>

        <TabPanel value={tabValue} index={0}>
          <div className="space-y-4">
            <div className="flex justify-between items-center">
              <div className="relative flex-1 max-w-md">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
                <input
                  type="text"
                  placeholder="搜索好友..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent w-full"
                />
              </div>
              <div className="flex space-x-2">
                <button
                  onClick={loadFriends}
                  disabled={loading}
                  className="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 disabled:opacity-50 flex items-center space-x-2"
                >
                  <RefreshCw className="w-4 h-4" />
                  <span>刷新</span>
                </button>
                <button
                  onClick={() => setAddFriendDialog(true)}
                  className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 flex items-center space-x-2"
                >
                  <UserPlus className="w-4 h-4" />
                  <span>添加好友</span>
                </button>
              </div>
            </div>

            <div className="space-y-2">
              {filteredFriends.map((friend) => (
                <div key={friend.id} className="flex items-center justify-between p-4 border border-gray-200 rounded-lg hover:bg-gray-50">
                  <div className="flex items-center space-x-3">
                    <div className="relative">
                      <div className="w-10 h-10 bg-gray-300 rounded-full flex items-center justify-center">
                        {friend.avatar_url ? (
                          <img src={friend.avatar_url} alt={friend.display_name} className="w-10 h-10 rounded-full" />
                        ) : (
                          <User className="w-6 h-6 text-gray-600" />
                        )}
                      </div>
                      <div className={`absolute -bottom-1 -right-1 w-3 h-3 rounded-full border-2 border-white ${getStatusColor(friend.online_status)}`}></div>
                    </div>
                    <div>
                      <div className="font-medium text-gray-900">{friend.display_name}</div>
                      <div className="text-sm text-gray-500">@{friend.username}</div>
                      <div className="text-xs text-gray-400">{getStatusText(friend.online_status)}</div>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    <button className="p-2 text-gray-400 hover:text-blue-600 rounded-lg hover:bg-blue-50">
                      <MessageCircle className="w-4 h-4" />
                    </button>
                    <button className="p-2 text-gray-400 hover:text-green-600 rounded-lg hover:bg-green-50">
                      <Phone className="w-4 h-4" />
                    </button>
                    <button className="p-2 text-gray-400 hover:text-purple-600 rounded-lg hover:bg-purple-50">
                      <Video className="w-4 h-4" />
                    </button>
                    <div className="relative">
                      <button
                        onClick={(e) => {
                          setAnchorEl(e.currentTarget);
                          setSelectedFriend(friend);
                        }}
                        className="p-2 text-gray-400 hover:text-gray-600 rounded-lg hover:bg-gray-100"
                      >
                        <MoreVertical className="w-4 h-4" />
                      </button>
                      {anchorEl && selectedFriend?.id === friend.id && (
                        <div className="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-lg border border-gray-200 z-10">
                          <button
                            onClick={() => {
                              handleBlockFriend(friend.id);
                              setAnchorEl(null);
                              setSelectedFriend(null);
                            }}
                            className="w-full px-4 py-2 text-left text-gray-700 hover:bg-gray-100 flex items-center space-x-2"
                          >
                            <Shield className="w-4 h-4" />
                            <span>屏蔽</span>
                          </button>
                          <button
                            onClick={() => {
                              handleRemoveFriend(friend.id);
                              setAnchorEl(null);
                              setSelectedFriend(null);
                            }}
                            className="w-full px-4 py-2 text-left text-red-600 hover:bg-red-50 flex items-center space-x-2"
                          >
                            <Trash2 className="w-4 h-4" />
                            <span>删除好友</span>
                          </button>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              ))}
              {filteredFriends.length === 0 && (
                <div className="text-center py-8 text-gray-500">
                  {searchTerm ? '没有找到匹配的好友' : '暂无好友'}
                </div>
              )}
            </div>
          </div>
        </TabPanel>

        <TabPanel value={tabValue} index={1}>
          <div className="space-y-4">
            <div className="flex justify-between items-center">
              <h3 className="text-lg font-medium text-gray-900">待处理的好友请求</h3>
              <button
                onClick={loadFriendRequests}
                disabled={loading}
                className="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 disabled:opacity-50 flex items-center space-x-2"
              >
                <RefreshCw className="w-4 h-4" />
                <span>刷新</span>
              </button>
            </div>

            <div className="space-y-2">
              {friendRequests.map((request) => (
                <div key={request.id} className="flex items-center justify-between p-4 border border-gray-200 rounded-lg">
                  <div className="flex items-center space-x-3">
                    <div className="w-10 h-10 bg-gray-300 rounded-full flex items-center justify-center">
                      {request.from_avatar_url ? (
                        <img src={request.from_avatar_url} alt={request.from_display_name} className="w-10 h-10 rounded-full" />
                      ) : (
                        <User className="w-6 h-6 text-gray-600" />
                      )}
                    </div>
                    <div>
                      <div className="font-medium text-gray-900">{request.from_display_name}</div>
                      <div className="text-sm text-gray-500">@{request.from_username}</div>
                      {request.message && (
                        <div className="text-sm text-gray-600 mt-1">"{request.message}"</div>
                      )}
                      <div className="text-xs text-gray-400">
                        {new Date(request.created_at).toLocaleDateString()}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    <button
                      onClick={() => handleAcceptRequest(request.id)}
                      disabled={loading}
                      className="px-3 py-1 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50 flex items-center space-x-1"
                    >
                      <Check className="w-4 h-4" />
                      <span>接受</span>
                    </button>
                    <button
                      onClick={() => handleDeclineRequest(request.id)}
                      disabled={loading}
                      className="px-3 py-1 bg-red-600 text-white rounded-lg hover:bg-red-700 disabled:opacity-50 flex items-center space-x-1"
                    >
                      <X className="w-4 h-4" />
                      <span>拒绝</span>
                    </button>
                  </div>
                </div>
              ))}
              {friendRequests.length === 0 && (
                <div className="text-center py-8 text-gray-500">
                  暂无好友请求
                </div>
              )}
            </div>
          </div>
        </TabPanel>
      </div>

      {/* 添加好友对话框 */}
      {addFriendDialog && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h3 className="text-lg font-medium text-gray-900 mb-4">添加好友</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">用户名</label>
                <input
                  type="text"
                  value={addFriendForm.username}
                  onChange={(e) => setAddFriendForm({ ...addFriendForm, username: e.target.value })}
                  placeholder="输入用户名"
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">消息（可选）</label>
                <textarea
                  value={addFriendForm.message}
                  onChange={(e) => setAddFriendForm({ ...addFriendForm, message: e.target.value })}
                  placeholder="添加一条消息..."
                  rows={3}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
            </div>
            <div className="flex justify-end space-x-2 mt-6">
              <button
                onClick={() => setAddFriendDialog(false)}
                className="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200"
              >
                取消
              </button>
              <button
                onClick={handleAddFriend}
                disabled={loading}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
              >
                发送请求
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 点击外部关闭菜单 */}
      {anchorEl && (
        <div
          className="fixed inset-0 z-0"
          onClick={() => {
            setAnchorEl(null);
            setSelectedFriend(null);
          }}
        />
      )}
    </div>
  );
};

export default FriendManagementPage;