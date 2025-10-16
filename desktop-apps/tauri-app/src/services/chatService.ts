import { invoke } from '@tauri-apps/api/core';

export interface ChatMessage {
  id: string;
  chatId: string;
  senderId: string;
  senderName: string;
  content: string;
  timestamp: string;
  type: 'text' | 'image' | 'file';
  isRead: boolean;
}

export interface Chat {
  id: string;
  name: string;
  type: 'private' | 'group';
  lastMessage: string;
  lastMessageTime: string;
  unreadCount: number;
  isOnline: boolean;
  avatar?: string;
  participants?: string[];
}

export interface ChatResponse {
  success: boolean;
  message: string;
  chats?: Chat[];
  messages?: ChatMessage[];
  chat?: Chat;
}

export interface SendMessageRequest {
  chat_id: string;
  content: string;
  message_type: 'Text' | 'Image' | 'File';
  reply_to_message_id?: string;
}

export interface CreateChatRequest {
  chat_type: 'Private' | 'Group';
  name: string;
  participants: string[];
}

class ChatService {
  private authToken: string = 'dummy_token'; // 临时token，需要从认证系统获取
  private isConnected: boolean = false;
  private messageListeners: ((message: ChatMessage) => void)[] = [];
  private connectionListeners: ((connected: boolean) => void)[] = [];

  // 设置认证token
  setAuthToken(token: string) {
    this.authToken = token;
  }

  // 获取聊天列表
  async getChatList(): Promise<ChatResponse> {
    try {
      const response = await invoke<ChatResponse>('chat_get_list', {
        authToken: this.authToken,
      });
      return response;
    } catch (error) {
      console.error('获取聊天列表失败:', error);
      throw error;
    }
  }

  // 获取聊天消息
  async getChatMessages(chatId: string, limit?: number, offset?: number): Promise<ChatResponse> {
    try {
      const response = await invoke<ChatResponse>('chat_get_messages', {
        authToken: this.authToken,
        chatId,
        limit,
        offset,
      });
      return response;
    } catch (error) {
      console.error('获取聊天消息失败:', error);
      throw error;
    }
  }

  // 发送消息
  async sendMessage(request: SendMessageRequest): Promise<ChatResponse> {
    try {
      const response = await invoke<ChatResponse>('chat_send_message', {
        authToken: this.authToken,
        request,
      });
      return response;
    } catch (error) {
      console.error('发送消息失败:', error);
      throw error;
    }
  }

  // 创建聊天
  async createChat(request: CreateChatRequest): Promise<ChatResponse> {
    try {
      const response = await invoke<ChatResponse>('chat_create', {
        authToken: this.authToken,
        request,
      });
      return response;
    } catch (error) {
      console.error('创建聊天失败:', error);
      throw error;
    }
  }

  // 删除聊天
  async deleteChat(chatId: string): Promise<ChatResponse> {
    try {
      const response = await invoke<ChatResponse>('chat_delete', {
        authToken: this.authToken,
        chatId,
      });
      return response;
    } catch (error) {
      console.error('删除聊天失败:', error);
      throw error;
    }
  }

  // 标记消息为已读
  async markMessagesRead(chatId: string, messageIds: string[]): Promise<ChatResponse> {
    try {
      const response = await invoke<ChatResponse>('chat_mark_read', {
        authToken: this.authToken,
        chatId,
        messageIds,
      });
      return response;
    } catch (error) {
      console.error('标记消息已读失败:', error);
      throw error;
    }
  }

  // 连接WebSocket
  async connectWebSocket(): Promise<void> {
    try {
      const response = await invoke<ChatResponse>('chat_connect_websocket', {
        authToken: this.authToken,
      });
      
      if (response.success) {
        this.isConnected = true;
        this.notifyConnectionListeners(true);
        console.log('WebSocket连接成功');
        
        // 模拟接收到新消息时的处理
        // 在实际实现中，这应该通过WebSocket事件来触发
        setTimeout(() => {
          const mockMessage: ChatMessage = {
            id: Date.now().toString(),
            chatId: '1',
            senderId: 'system',
            senderName: '系统',
            content: 'WebSocket连接已建立',
            timestamp: new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }),
            type: 'text',
            isRead: false,
          };
          this.notifyMessageListeners(mockMessage);
        }, 1000);
      } else {
        throw new Error(response.message);
      }
    } catch (error) {
      console.error('WebSocket连接失败:', error);
      this.isConnected = false;
      this.notifyConnectionListeners(false);
      throw error;
    }
  }

  // 断开WebSocket连接
  async disconnectWebSocket(): Promise<void> {
    try {
      const response = await invoke<ChatResponse>('chat_disconnect_websocket');
      
      if (response.success) {
        this.isConnected = false;
        this.notifyConnectionListeners(false);
        console.log('WebSocket连接已断开');
      }
    } catch (error) {
      console.error('断开WebSocket连接失败:', error);
    }
  }

  // 检查连接状态
  async checkConnectionStatus(): Promise<boolean> {
    try {
      const connected = await invoke<boolean>('chat_is_connected');
      this.isConnected = connected;
      return connected;
    } catch (error) {
      console.error('检查连接状态失败:', error);
      return false;
    }
  }

  // 获取缓存的聊天列表
  async getCachedChatList(): Promise<Chat[]> {
    try {
      const chats = await invoke<Chat[]>('chat_get_cached_list');
      return chats;
    } catch (error) {
      console.error('获取缓存聊天列表失败:', error);
      return [];
    }
  }

  // 获取缓存的聊天消息
  async getCachedMessages(chatId: string): Promise<ChatMessage[]> {
    try {
      const messages = await invoke<ChatMessage[]>('chat_get_cached_messages', {
        chatId,
      });
      return messages;
    } catch (error) {
      console.error('获取缓存消息失败:', error);
      return [];
    }
  }

  // 设置当前用户
  async setCurrentUser(userId: string): Promise<void> {
    try {
      await invoke<ChatResponse>('chat_set_current_user', {
        userId,
      });
    } catch (error) {
      console.error('设置当前用户失败:', error);
      throw error;
    }
  }

  // 添加消息监听器
  addMessageListener(listener: (message: ChatMessage) => void) {
    this.messageListeners.push(listener);
  }

  // 移除消息监听器
  removeMessageListener(listener: (message: ChatMessage) => void) {
    const index = this.messageListeners.indexOf(listener);
    if (index > -1) {
      this.messageListeners.splice(index, 1);
    }
  }

  // 添加连接状态监听器
  addConnectionListener(listener: (connected: boolean) => void) {
    this.connectionListeners.push(listener);
  }

  // 移除连接状态监听器
  removeConnectionListener(listener: (connected: boolean) => void) {
    const index = this.connectionListeners.indexOf(listener);
    if (index > -1) {
      this.connectionListeners.splice(index, 1);
    }
  }

  // 通知消息监听器
  private notifyMessageListeners(message: ChatMessage) {
    this.messageListeners.forEach(listener => listener(message));
  }

  // 通知连接状态监听器
  private notifyConnectionListeners(connected: boolean) {
    this.connectionListeners.forEach(listener => listener(connected));
  }

  // 获取连接状态
  getConnectionStatus(): boolean {
    return this.isConnected;
  }
}

// 导出单例实例
export const chatService = new ChatService();
export default chatService;