/**
 * 聊天视图模型
 * 管理聊天界面的状态和业务逻辑
 */

import { 
  ChatMessage, 
  Conversation, 
  MessageType, 
  MessageSender, 
  MessageStatus,
  AIPersonality
} from '../model/ChatMessage';
import { ChatRepository, ChatRepositoryInterface } from '../service/ChatRepository';
import { MediaService, MediaFileInfo, MediaType } from '../service/MediaService';

// UI状态枚举
export enum UIState {
  IDLE = 'idle',
  LOADING = 'loading',
  SENDING = 'sending',
  ERROR = 'error'
}

// 错误信息接口
export interface ErrorInfo {
  code: string;
  message: string;
  details?: Record<string, any>;
}

// 聊天视图模型
@ObservedV2
export class ChatViewModel {
  // 单例实例
  private static instance: ChatViewModel;
  
  // 依赖注入
  private chatRepository: ChatRepositoryInterface;
  private mediaService: MediaService;
  
  // UI状态
  @Trace uiState: UIState = UIState.IDLE;
  @Trace errorInfo: ErrorInfo | null = null;
  
  // 当前对话
  @Trace currentConversation: Conversation | null = null;
  @Trace messages: ChatMessage[] = [];
  
  // 对话列表
  @Trace conversations: Conversation[] = [];
  @Trace isConversationListVisible: boolean = false;
  
  // 消息输入
  @Trace inputText: string = '';
  @Trace isInputEnabled: boolean = true;
  
  // AI设置
  @Trace selectedAIPersonality: AIPersonality = AIPersonality.DEFAULT;
  @Trace isAIPersonalityDialogVisible: boolean = false;
  
  // 搜索
  @Trace searchQuery: string = '';
  @Trace searchResults: ChatMessage[] = [];
  @Trace isSearchMode: boolean = false;
  
  // 统计信息
  @Trace unreadCount: number = 0;
  @Trace totalMessageCount: number = 0;

  private constructor() {
    this.chatRepository = ChatRepository.getInstance();
    this.mediaService = MediaService.getInstance();
    this.initialize();
  }

  static getInstance(): ChatViewModel {
    if (!ChatViewModel.instance) {
      ChatViewModel.instance = new ChatViewModel();
    }
    return ChatViewModel.instance;
  }

  // 初始化
  private async initialize(): Promise<void> {
    try {
      this.uiState = UIState.LOADING;
      
      await this.chatRepository.initialize();
      await this.loadConversations();
      await this.updateUnreadCount();
      
      this.uiState = UIState.IDLE;
    } catch (error) {
      console.error('Failed to initialize ChatViewModel:', error);
      this.handleError('INIT_ERROR', '初始化失败', error);
    }
  }

  // 创建新对话
  async createNewConversation(title?: string): Promise<void> {
    try {
      this.uiState = UIState.LOADING;
      this.clearError();
      
      const conversationTitle = title || `对话 ${new Date().toLocaleString()}`;
      const conversation = await this.chatRepository.createConversation(
        conversationTitle, 
        this.selectedAIPersonality
      );
      
      this.currentConversation = conversation;
      this.messages = [];
      this.conversations.unshift(conversation);
      this.isConversationListVisible = false;
      
      this.uiState = UIState.IDLE;
    } catch (error) {
      console.error('Failed to create conversation:', error);
      this.handleError('CREATE_CONVERSATION_ERROR', '创建对话失败', error);
    }
  }

  // 选择对话
  async selectConversation(conversationId: string): Promise<void> {
    try {
      this.uiState = UIState.LOADING;
      this.clearError();
      
      const conversation = await this.chatRepository.getConversation(conversationId);
      if (!conversation) {
        throw new Error('Conversation not found');
      }
      
      this.currentConversation = conversation;
      this.selectedAIPersonality = conversation.aiPersonality;
      
      // 加载消息
      await this.loadMessages(conversationId);
      
      this.isConversationListVisible = false;
      this.uiState = UIState.IDLE;
    } catch (error) {
      console.error('Failed to select conversation:', error);
      this.handleError('SELECT_CONVERSATION_ERROR', '选择对话失败', error);
    }
  }

  // 发送消息
  async sendMessage(content?: string): Promise<void> {
    const messageContent = content || this.inputText.trim();
    
    if (!messageContent || !this.currentConversation) {
      return;
    }

    try {
      this.uiState = UIState.SENDING;
      this.isInputEnabled = false;
      this.clearError();
      
      // 清空输入框
      this.inputText = '';
      
      // 发送消息
      const aiMessage = await this.chatRepository.sendMessage(
        this.currentConversation.conversationId,
        messageContent,
        MessageType.TEXT
      );
      
      // 重新加载消息列表
      await this.loadMessages(this.currentConversation.conversationId);
      
      // 更新对话列表中的对话信息
      await this.updateConversationInList(this.currentConversation.conversationId);
      
      this.uiState = UIState.IDLE;
      this.isInputEnabled = true;
    } catch (error) {
      console.error('Failed to send message:', error);
      this.handleError('SEND_MESSAGE_ERROR', '发送消息失败', error);
      this.isInputEnabled = true;
    }
  }

  // 重试消息
  async retryMessage(messageId: string): Promise<void> {
    try {
      this.uiState = UIState.SENDING;
      this.clearError();
      
      await this.chatRepository.retryMessage(messageId);
      
      // 重新加载消息
      if (this.currentConversation) {
        await this.loadMessages(this.currentConversation.conversationId);
      }
      
      this.uiState = UIState.IDLE;
    } catch (error) {
      console.error('Failed to retry message:', error);
      this.handleError('RETRY_MESSAGE_ERROR', '重试消息失败', error);
    }
  }

  // 删除消息
  async deleteMessage(messageId: string): Promise<void> {
    try {
      this.clearError();
      
      await this.chatRepository.deleteMessage(messageId);
      
      // 从当前消息列表中移除
      this.messages = this.messages.filter(m => m.messageId !== messageId);
      
      // 更新对话信息
      if (this.currentConversation) {
        await this.updateConversationInList(this.currentConversation.conversationId);
      }
    } catch (error) {
      console.error('Failed to delete message:', error);
      this.handleError('DELETE_MESSAGE_ERROR', '删除消息失败', error);
    }
  }

  // 删除对话
  async deleteConversation(conversationId: string): Promise<void> {
    try {
      this.clearError();
      
      await this.chatRepository.deleteConversation(conversationId);
      
      // 从对话列表中移除
      this.conversations = this.conversations.filter(c => c.conversationId !== conversationId);
      
      // 如果删除的是当前对话，清空当前对话
      if (this.currentConversation?.conversationId === conversationId) {
        this.currentConversation = null;
        this.messages = [];
      }
    } catch (error) {
      console.error('Failed to delete conversation:', error);
      this.handleError('DELETE_CONVERSATION_ERROR', '删除对话失败', error);
    }
  }

  // 归档对话
  async archiveConversation(conversationId: string): Promise<void> {
    try {
      this.clearError();
      
      await this.chatRepository.archiveConversation(conversationId);
      
      // 更新对话列表中的对话状态
      const conversation = this.conversations.find(c => c.conversationId === conversationId);
      if (conversation) {
        conversation.isArchived = true;
      }
    } catch (error) {
      console.error('Failed to archive conversation:', error);
      this.handleError('ARCHIVE_CONVERSATION_ERROR', '归档对话失败', error);
    }
  }

  // 搜索消息
  async searchMessages(query: string): Promise<void> {
    try {
      this.clearError();
      this.searchQuery = query;
      
      if (!query.trim()) {
        this.searchResults = [];
        this.isSearchMode = false;
        return;
      }
      
      this.isSearchMode = true;
      this.searchResults = await this.chatRepository.searchMessages(
        query, 
        this.currentConversation?.conversationId
      );
    } catch (error) {
      console.error('Failed to search messages:', error);
      this.handleError('SEARCH_ERROR', '搜索失败', error);
    }
  }

  // 清除搜索
  clearSearch(): void {
    this.searchQuery = '';
    this.searchResults = [];
    this.isSearchMode = false;
  }

  // 刷新数据
  async refreshData(): Promise<void> {
    try {
      this.uiState = UIState.LOADING;
      this.clearError();
      
      await this.loadConversations();
      
      if (this.currentConversation) {
        await this.loadMessages(this.currentConversation.conversationId);
      }
      
      await this.updateUnreadCount();
      
      this.uiState = UIState.IDLE;
    } catch (error) {
      console.error('Failed to refresh data:', error);
      this.handleError('REFRESH_ERROR', '刷新数据失败', error);
    }
  }

  // 同步数据
  async syncWithServer(): Promise<void> {
    try {
      this.clearError();
      
      await this.chatRepository.syncWithServer();
      await this.refreshData();
    } catch (error) {
      console.error('Failed to sync with server:', error);
      this.handleError('SYNC_ERROR', '同步失败', error);
    }
  }

  // 更改AI人格
  async changeAIPersonality(personality: AIPersonality): Promise<void> {
    this.selectedAIPersonality = personality;
    this.isAIPersonalityDialogVisible = false;
    
    // 如果有当前对话，更新对话的AI人格
    if (this.currentConversation) {
      this.currentConversation.aiPersonality = personality;
      try {
        await this.chatRepository.updateConversation(this.currentConversation);
      } catch (error) {
        console.error('Failed to update conversation AI personality:', error);
      }
    }
  }

  // UI控制方法
  showConversationList(): void {
    this.isConversationListVisible = true;
  }

  hideConversationList(): void {
    this.isConversationListVisible = false;
  }

  showAIPersonalityDialog(): void {
    this.isAIPersonalityDialogVisible = true;
  }

  hideAIPersonalityDialog(): void {
    this.isAIPersonalityDialogVisible = false;
  }

  updateInputText(text: string): void {
    this.inputText = text;
  }

  // 私有方法：加载对话列表
  private async loadConversations(): Promise<void> {
    this.conversations = await this.chatRepository.getConversations();
  }

  // 私有方法：加载消息
  private async loadMessages(conversationId: string): Promise<void> {
    this.messages = await this.chatRepository.getMessages(conversationId);
  }

  // 私有方法：更新对话列表中的对话信息
  private async updateConversationInList(conversationId: string): Promise<void> {
    const updatedConversation = await this.chatRepository.getConversation(conversationId);
    if (updatedConversation) {
      const index = this.conversations.findIndex(c => c.conversationId === conversationId);
      if (index !== -1) {
        this.conversations[index] = updatedConversation;
        // 将更新的对话移到列表顶部
        this.conversations.splice(index, 1);
        this.conversations.unshift(updatedConversation);
      }
    }
  }

  // 私有方法：更新未读消息数量
  private async updateUnreadCount(): Promise<void> {
    try {
      this.unreadCount = await this.chatRepository.getUnreadCount();
    } catch (error) {
      console.error('Failed to update unread count:', error);
    }
  }

  // 私有方法：错误处理
  private handleError(code: string, message: string, error?: any): void {
    this.errorInfo = {
      code,
      message,
      details: error instanceof Error ? { message: error.message } : error
    };
    this.uiState = UIState.ERROR;
  }

  // 私有方法：清除错误
  private clearError(): void {
    this.errorInfo = null;
    if (this.uiState === UIState.ERROR) {
      this.uiState = UIState.IDLE;
    }
  }

  // 计算属性：是否可以发送消息
  get canSendMessage(): boolean {
    return this.inputText.trim().length > 0 && 
           this.isInputEnabled && 
           this.currentConversation !== null &&
           this.uiState !== UIState.SENDING;
  }

  // 计算属性：是否显示加载状态
  get isLoading(): boolean {
    return this.uiState === UIState.LOADING;
  }

  // 计算属性：是否显示发送状态
  get isSending(): boolean {
    return this.uiState === UIState.SENDING;
  }

  // 计算属性：是否有错误
  get hasError(): boolean {
    return this.errorInfo !== null;
  }

  // 计算属性：当前显示的消息列表
  get displayMessages(): ChatMessage[] {
    return this.isSearchMode ? this.searchResults : this.messages;
  }

  // 计算属性：AI人格显示名称
  get aiPersonalityDisplayName(): string {
    switch (this.selectedAIPersonality) {
      case AIPersonality.DEFAULT:
        return '默认助手';
      case AIPersonality.WISE_SAGE:
        return '智慧长者';
      case AIPersonality.FRIENDLY_GUIDE:
        return '友善向导';
      case AIPersonality.SCHOLARLY:
        return '学者专家';
      case AIPersonality.POETIC:
        return '诗意文人';
      default:
        return '默认助手';
    }
  }

  // 工具方法：格式化时间
  formatTime(timestamp: number): string {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    
    if (diff < 60000) { // 1分钟内
      return '刚刚';
    } else if (diff < 3600000) { // 1小时内
      return `${Math.floor(diff / 60000)}分钟前`;
    } else if (diff < 86400000) { // 1天内
      return `${Math.floor(diff / 3600000)}小时前`;
    } else if (diff < 604800000) { // 1周内
      return `${Math.floor(diff / 86400000)}天前`;
    } else {
      return date.toLocaleDateString();
    }
  }

  // 工具方法：格式化消息状态
  getMessageStatusText(status: MessageStatus): string {
    switch (status) {
      case MessageStatus.SENDING:
        return '发送中...';
      case MessageStatus.SENT:
        return '已发送';
      case MessageStatus.FAILED:
        return '发送失败';
      default:
        return '';
    }
  }

  // 工具方法：获取消息状态颜色
  getMessageStatusColor(status: MessageStatus): string {
    switch (status) {
      case MessageStatus.SENDING:
        return '#FFA500';
      case MessageStatus.SENT:
        return '#00AA00';
      case MessageStatus.FAILED:
        return '#FF0000';
      default:
        return '#666666';
    }
  }

  // ==================== 多模态消息支持 ====================

  /**
   * 发送图片消息
   */
  async sendImageMessage(imageData: ArrayBuffer): Promise<void> {
    if (!this.currentConversation) {
      this.handleError('NO_CONVERSATION', '请先选择或创建对话', null);
      return;
    }

    try {
      this.uiState = UIState.SENDING;
      this.isInputEnabled = false;
      this.clearError();

      // 保存图片到本地
      const mediaInfo = await this.mediaService.saveImage(imageData);
      
      // 构建图片消息内容（使用本地文件路径）
      const imageContent = `[图片] ${mediaInfo.fileName}`;
      
      // 发送消息
      const aiMessage = await this.chatRepository.sendMessage(
        this.currentConversation.conversationId,
        imageContent,
        MessageType.IMAGE,
        { 
          mediaInfo: mediaInfo,
          imageUrl: mediaInfo.filePath 
        }
      );
      
      // 重新加载消息列表
      await this.loadMessages(this.currentConversation.conversationId);
      
      // 更新对话列表中的对话信息
      await this.updateConversationInList(this.currentConversation.conversationId);
      
      this.uiState = UIState.IDLE;
      this.isInputEnabled = true;
    } catch (error) {
      console.error('Failed to send image message:', error);
      this.handleError('SEND_IMAGE_ERROR', '发送图片失败', error);
      this.isInputEnabled = true;
    }
  }

  /**
   * 发送音频消息
   */
  async sendAudioMessage(audioData: ArrayBuffer, duration?: number): Promise<void> {
    if (!this.currentConversation) {
      this.handleError('NO_CONVERSATION', '请先选择或创建对话', null);
      return;
    }

    try {
      this.uiState = UIState.SENDING;
      this.isInputEnabled = false;
      this.clearError();

      // 保存音频到本地
      const mediaInfo = await this.mediaService.saveAudio(audioData);
      
      // 构建音频消息内容
      const audioContent = `[语音] ${this.formatDuration(duration || 0)}`;
      
      // 发送消息
      const aiMessage = await this.chatRepository.sendMessage(
        this.currentConversation.conversationId,
        audioContent,
        MessageType.VOICE,
        { 
          mediaInfo: mediaInfo,
          audioUrl: mediaInfo.filePath,
          duration: duration 
        }
      );
      
      // 重新加载消息列表
      await this.loadMessages(this.currentConversation.conversationId);
      
      // 更新对话列表中的对话信息
      await this.updateConversationInList(this.currentConversation.conversationId);
      
      this.uiState = UIState.IDLE;
      this.isInputEnabled = true;
    } catch (error) {
      console.error('Failed to send audio message:', error);
      this.handleError('SEND_AUDIO_ERROR', '发送语音失败', error);
      this.isInputEnabled = true;
    }
  }

  /**
   * 发送文件消息
   */
  async sendFileMessage(fileData: ArrayBuffer, fileName: string, mimeType: string): Promise<void> {
    if (!this.currentConversation) {
      this.handleError('NO_CONVERSATION', '请先选择或创建对话', null);
      return;
    }

    try {
      this.uiState = UIState.SENDING;
      this.isInputEnabled = false;
      this.clearError();

      // 根据文件类型选择保存方法
      let mediaInfo: MediaFileInfo;
      const extension = fileName.split('.').pop()?.toLowerCase() || '';
      
      if (['jpg', 'jpeg', 'png', 'gif', 'webp'].includes(extension)) {
        mediaInfo = await this.mediaService.saveImage(fileData);
      } else if (['mp3', 'wav', 'aac'].includes(extension)) {
        mediaInfo = await this.mediaService.saveAudio(fileData, extension);
      } else if (['mp4', 'mov', 'avi'].includes(extension)) {
        mediaInfo = await this.mediaService.saveVideo(fileData, extension);
      } else {
        // 对于其他文件类型，保存为通用文件
        mediaInfo = await this.mediaService.saveAudio(fileData, extension); // 复用音频保存逻辑
      }
      
      // 构建文件消息内容
      const fileContent = `[文件] ${fileName}`;
      
      // 发送消息
      const aiMessage = await this.chatRepository.sendMessage(
        this.currentConversation.conversationId,
        fileContent,
        MessageType.FILE,
        { 
          mediaInfo: mediaInfo,
          fileName: fileName,
          fileUrl: mediaInfo.filePath,
          mimeType: mimeType
        }
      );
      
      // 重新加载消息列表
      await this.loadMessages(this.currentConversation.conversationId);
      
      // 更新对话列表中的对话信息
      await this.updateConversationInList(this.currentConversation.conversationId);
      
      this.uiState = UIState.IDLE;
      this.isInputEnabled = true;
    } catch (error) {
      console.error('Failed to send file message:', error);
      this.handleError('SEND_FILE_ERROR', '发送文件失败', error);
      this.isInputEnabled = true;
    }
  }

  /**
   * 获取媒体文件信息
   */
  async getMediaInfo(filePath: string): Promise<MediaFileInfo | null> {
    try {
      return await this.mediaService.getMediaInfo(filePath);
    } catch (error) {
      console.error('Failed to get media info:', error);
      return null;
    }
  }

  /**
   * 删除媒体文件
   */
  async deleteMediaFile(filePath: string): Promise<void> {
    try {
      await this.mediaService.deleteMedia(filePath);
    } catch (error) {
      console.error('Failed to delete media file:', error);
      throw error;
    }
  }

  /**
   * 清理过期媒体文件
   */
  async cleanupExpiredMedia(): Promise<void> {
    try {
      // 清理7天前的媒体文件
      await this.mediaService.cleanupExpiredMedia(7 * 24 * 60 * 60 * 1000);
    } catch (error) {
      console.error('Failed to cleanup expired media:', error);
    }
  }

  /**
   * 获取媒体存储使用情况
   */
  async getMediaStorageUsage(): Promise<{ size: number, formattedSize: string }> {
    try {
      const size = await this.mediaService.getMediaDirectorySize();
      return {
        size: size,
        formattedSize: this.formatFileSize(size)
      };
    } catch (error) {
      console.error('Failed to get media storage usage:', error);
      return { size: 0, formattedSize: '0 B' };
    }
  }

  // ==================== 工具方法 ====================

  /**
   * 格式化时长
   */
  private formatDuration(milliseconds: number): string {
    const seconds = Math.floor(milliseconds / 1000);
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    
    if (minutes > 0) {
      return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`;
    } else {
      return `0:${remainingSeconds.toString().padStart(2, '0')}`;
    }
  }

  /**
   * 格式化文件大小
   */
  private formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 B';
    
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  /**
   * 检查是否支持的媒体类型
   */
  isSupportedMediaType(mimeType: string): boolean {
    const supportedTypes = [
      'image/jpeg', 'image/png', 'image/gif', 'image/webp',
      'audio/mpeg', 'audio/wav', 'audio/aac',
      'video/mp4', 'video/quicktime', 'video/x-msvideo'
    ];
    return supportedTypes.includes(mimeType);
  }

  /**
   * 获取媒体类型显示名称
   */
  getMediaTypeDisplayName(type: MessageType): string {
    switch (type) {
      case MessageType.IMAGE:
        return '图片';
      case MessageType.VOICE:
        return '语音';
      case MessageType.FILE:
        return '文件';
      default:
        return '文本';
    }
  }
}