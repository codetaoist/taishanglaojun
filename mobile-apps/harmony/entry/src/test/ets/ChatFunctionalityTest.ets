/**
 * 鸿蒙AI对话功能测试
 * 测试ChatViewModel和ChatRepository的核心功能
 */

import { describe, beforeAll, beforeEach, afterEach, it, expect } from '@ohos/hypium';
import { 
  ChatMessage, 
  Conversation, 
  MessageType, 
  MessageSender, 
  MessageStatus,
  AIPersonality
} from '../../../main/ets/model/ChatMessage';
import { ChatViewModel, UIState } from '../../../main/ets/viewmodel/ChatViewModel';
import { ChatRepository } from '../../../main/ets/service/ChatRepository';

export default function ChatFunctionalityTest() {
  describe('ChatFunctionalityTest', () => {
    let chatViewModel: ChatViewModel;
    let chatRepository: ChatRepository;

    beforeAll(async () => {
      // 测试环境初始化
      console.info('开始鸿蒙AI对话功能测试');
    });

    beforeEach(async () => {
      // 每个测试前的准备
      chatViewModel = ChatViewModel.getInstance();
      chatRepository = ChatRepository.getInstance();
      
      // 重置状态
      await chatRepository.clearAllData();
    });

    afterEach(async () => {
      // 每个测试后的清理
      await chatRepository.clearAllData();
    });

    it('should create new conversation successfully', async () => {
      // 测试创建新对话
      const initialConversationCount = chatViewModel.conversations.length;
      
      await chatViewModel.createNewConversation('测试对话');
      
      expect(chatViewModel.currentConversation).not.toBeNull();
      expect(chatViewModel.currentConversation?.title).assertEqual('测试对话');
      expect(chatViewModel.conversations.length).assertEqual(initialConversationCount + 1);
      expect(chatViewModel.uiState).assertEqual(UIState.IDLE);
    });

    it('should send message successfully', async () => {
      // 测试发送消息成功
      await chatViewModel.createNewConversation('测试对话');
      const conversationId = chatViewModel.currentConversation?.conversationId;
      
      expect(conversationId).not.toBeUndefined();
      
      // 模拟发送消息
      chatViewModel.inputText = '你好，AI助手';
      await chatViewModel.sendMessage();
      
      expect(chatViewModel.inputText).assertEqual('');
      expect(chatViewModel.uiState).assertEqual(UIState.IDLE);
      expect(chatViewModel.isInputEnabled).assertTrue();
      
      // 验证消息已保存
      const messages = await chatRepository.getMessages(conversationId!);
      expect(messages.length).assertLarger(0);
      
      const userMessage = messages.find(m => m.sender === MessageSender.USER);
      expect(userMessage).not.toBeUndefined();
      expect(userMessage?.content).assertEqual('你好，AI助手');
    });

    it('should handle send message failure', async () => {
      // 测试发送消息失败的处理
      await chatViewModel.createNewConversation('测试对话');
      
      // 模拟网络错误
      const originalSendMessage = chatRepository.sendMessage;
      chatRepository.sendMessage = async () => {
        throw new Error('网络连接失败');
      };
      
      chatViewModel.inputText = '测试消息';
      await chatViewModel.sendMessage();
      
      expect(chatViewModel.hasError).assertTrue();
      expect(chatViewModel.errorInfo?.message).assertEqual('发送消息失败');
      expect(chatViewModel.isInputEnabled).assertTrue();
      
      // 恢复原始方法
      chatRepository.sendMessage = originalSendMessage;
    });

    it('should delete conversation successfully', async () => {
      // 测试删除对话
      await chatViewModel.createNewConversation('待删除对话');
      const conversationId = chatViewModel.currentConversation?.conversationId;
      const initialCount = chatViewModel.conversations.length;
      
      expect(conversationId).not.toBeUndefined();
      
      await chatViewModel.deleteConversation(conversationId!);
      
      expect(chatViewModel.conversations.length).assertEqual(initialCount - 1);
      expect(chatViewModel.currentConversation).toBeNull();
      expect(chatViewModel.messages.length).assertEqual(0);
    });

    it('should change AI personality successfully', async () => {
      // 测试更改AI人格
      await chatViewModel.createNewConversation('测试对话');
      
      const newPersonality = AIPersonality.CREATIVE;
      await chatViewModel.changeAIPersonality(newPersonality);
      
      expect(chatViewModel.selectedAIPersonality).assertEqual(newPersonality);
      expect(chatViewModel.currentConversation?.aiPersonality).assertEqual(newPersonality);
      expect(chatViewModel.isAIPersonalityDialogVisible).assertFalse();
    });

    it('should retry failed message successfully', async () => {
      // 测试重试失败消息
      await chatViewModel.createNewConversation('测试对话');
      
      // 创建一个失败的消息
      const failedMessage: ChatMessage = {
        messageId: 'test-message-id',
        conversationId: chatViewModel.currentConversation!.conversationId,
        content: '测试重试消息',
        sender: MessageSender.USER,
        type: MessageType.TEXT,
        status: MessageStatus.FAILED,
        timestamp: Date.now(),
        aiPersonality: AIPersonality.DEFAULT
      };
      
      await chatRepository.saveMessage(failedMessage);
      
      // 重试消息
      await chatViewModel.retryMessage(failedMessage.messageId);
      
      expect(chatViewModel.uiState).assertEqual(UIState.IDLE);
      expect(chatViewModel.hasError).assertFalse();
    });

    it('should search messages successfully', async () => {
      // 测试搜索消息
      await chatViewModel.createNewConversation('测试对话');
      
      // 添加一些测试消息
      const testMessages = [
        '这是第一条测试消息',
        '这是关于AI的消息',
        '这是另一条普通消息'
      ];
      
      for (const content of testMessages) {
        chatViewModel.inputText = content;
        await chatViewModel.sendMessage();
      }
      
      // 搜索包含"AI"的消息
      await chatViewModel.searchMessages('AI');
      
      expect(chatViewModel.isSearchMode).assertTrue();
      expect(chatViewModel.searchQuery).assertEqual('AI');
      expect(chatViewModel.searchResults.length).assertLarger(0);
      
      const aiMessage = chatViewModel.searchResults.find(m => m.content.includes('AI'));
      expect(aiMessage).not.toBeUndefined();
    });

    it('should clear search successfully', async () => {
      // 测试清除搜索
      await chatViewModel.searchMessages('测试查询');
      
      expect(chatViewModel.isSearchMode).assertTrue();
      
      chatViewModel.clearSearch();
      
      expect(chatViewModel.isSearchMode).assertFalse();
      expect(chatViewModel.searchQuery).assertEqual('');
      expect(chatViewModel.searchResults.length).assertEqual(0);
    });

    it('should refresh data successfully', async () => {
      // 测试刷新数据
      await chatViewModel.createNewConversation('测试对话');
      
      // 模拟数据变化
      const originalState = chatViewModel.uiState;
      
      await chatViewModel.refreshData();
      
      expect(chatViewModel.uiState).assertEqual(UIState.IDLE);
      expect(chatViewModel.hasError).assertFalse();
    });

    it('should handle UI state transitions correctly', async () => {
      // 测试UI状态转换
      expect(chatViewModel.uiState).assertEqual(UIState.IDLE);
      
      // 创建对话时应该显示加载状态
      const createPromise = chatViewModel.createNewConversation('测试对话');
      
      // 在异步操作完成前检查状态
      expect(chatViewModel.uiState).assertEqual(UIState.LOADING);
      
      await createPromise;
      
      // 操作完成后应该回到空闲状态
      expect(chatViewModel.uiState).assertEqual(UIState.IDLE);
    });

    it('should manage conversation list visibility', async () => {
      // 测试对话列表显示控制
      expect(chatViewModel.isConversationListVisible).assertFalse();
      
      chatViewModel.showConversationList();
      expect(chatViewModel.isConversationListVisible).assertTrue();
      
      chatViewModel.hideConversationList();
      expect(chatViewModel.isConversationListVisible).assertFalse();
    });

    it('should manage AI personality dialog visibility', async () => {
      // 测试AI人格对话框显示控制
      expect(chatViewModel.isAIPersonalityDialogVisible).assertFalse();
      
      chatViewModel.showAIPersonalityDialog();
      expect(chatViewModel.isAIPersonalityDialogVisible).assertTrue();
      
      chatViewModel.hideAIPersonalityDialog();
      expect(chatViewModel.isAIPersonalityDialogVisible).assertFalse();
    });

    it('should validate input before sending message', async () => {
      // 测试发送消息前的输入验证
      await chatViewModel.createNewConversation('测试对话');
      
      // 测试空消息
      chatViewModel.inputText = '';
      await chatViewModel.sendMessage();
      
      // 空消息不应该被发送
      expect(chatViewModel.messages.length).assertEqual(0);
      
      // 测试只有空格的消息
      chatViewModel.inputText = '   ';
      await chatViewModel.sendMessage();
      
      expect(chatViewModel.messages.length).assertEqual(0);
      
      // 测试有效消息
      chatViewModel.inputText = '有效消息';
      await chatViewModel.sendMessage();
      
      expect(chatViewModel.inputText).assertEqual('');
    });

    it('should handle error states correctly', async () => {
      // 测试错误状态处理
      expect(chatViewModel.hasError).assertFalse();
      expect(chatViewModel.errorInfo).toBeNull();
      
      // 模拟错误
      const originalCreateConversation = chatRepository.createConversation;
      chatRepository.createConversation = async () => {
        throw new Error('数据库错误');
      };
      
      await chatViewModel.createNewConversation('测试对话');
      
      expect(chatViewModel.hasError).assertTrue();
      expect(chatViewModel.errorInfo).not.toBeNull();
      expect(chatViewModel.errorInfo?.message).assertEqual('创建对话失败');
      
      // 清除错误
      chatViewModel.clearError();
      
      expect(chatViewModel.hasError).assertFalse();
      expect(chatViewModel.errorInfo).toBeNull();
      
      // 恢复原始方法
      chatRepository.createConversation = originalCreateConversation;
    });
  });
}