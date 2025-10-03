/**
 * 鸿蒙AI对话功能集成测试
 * 测试UI组件与数据层的完整交互
 */

import { describe, beforeAll, beforeEach, afterEach, it, expect } from '@ohos/hypium';
import { Driver, ON } from '@ohos.UiTest';
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

export default function ChatIntegrationTest() {
  describe('ChatIntegrationTest', () => {
    let driver: Driver;
    let chatViewModel: ChatViewModel;
    let chatRepository: ChatRepository;

    beforeAll(async () => {
      // 初始化UI测试驱动
      driver = Driver.create();
      console.info('开始鸿蒙AI对话功能集成测试');
    });

    beforeEach(async () => {
      // 每个测试前的准备
      chatViewModel = ChatViewModel.getInstance();
      chatRepository = ChatRepository.getInstance();
      
      // 重置数据
      await chatRepository.clearAllData();
      
      // 等待UI稳定
      await driver.delayMs(500);
    });

    afterEach(async () => {
      // 每个测试后的清理
      await chatRepository.clearAllData();
      await driver.delayMs(300);
    });

    it('should display welcome screen initially', async () => {
      // 测试初始欢迎界面显示
      
      // 验证欢迎界面元素
      const welcomeText = await driver.findComponent(ON.text('开始对话吧'));
      expect(await welcomeText.isEnabled()).assertTrue();
      
      const subtitleText = await driver.findComponent(ON.text('向AI助手提问任何问题'));
      expect(await subtitleText.isEnabled()).assertTrue();
      
      // 验证没有消息列表
      const messageList = await driver.findComponent(ON.type('List'));
      expect(await messageList.isEnabled()).assertFalse();
    });

    it('should create new conversation through UI', async () => {
      // 测试通过UI创建新对话
      
      // 点击新建对话按钮
      const newConversationButton = await driver.findComponent(ON.id('new_conversation_button'));
      await newConversationButton.click();
      
      // 等待对话创建
      await driver.delayMs(1000);
      
      // 验证对话已创建
      expect(chatViewModel.currentConversation).not.toBeNull();
      
      // 验证UI更新
      const titleText = await driver.findComponent(ON.text(chatViewModel.currentConversation!.title));
      expect(await titleText.isEnabled()).assertTrue();
      
      // 验证输入区域可见
      const inputArea = await driver.findComponent(ON.id('message_input'));
      expect(await inputArea.isEnabled()).assertTrue();
    });

    it('should send message through UI interaction', async () => {
      // 测试通过UI发送消息
      
      // 先创建对话
      await chatViewModel.createNewConversation('UI测试对话');
      await driver.delayMs(500);
      
      // 输入消息
      const messageInput = await driver.findComponent(ON.id('message_input'));
      await messageInput.inputText('这是一条UI测试消息');
      
      // 点击发送按钮
      const sendButton = await driver.findComponent(ON.id('send_button'));
      await sendButton.click();
      
      // 等待消息发送
      await driver.delayMs(2000);
      
      // 验证消息显示在界面上
      const userMessage = await driver.findComponent(ON.text('这是一条UI测试消息'));
      expect(await userMessage.isEnabled()).assertTrue();
      
      // 验证输入框已清空
      const inputText = await messageInput.getText();
      expect(inputText).assertEqual('');
    });

    it('should display loading state during message sending', async () => {
      // 测试发送消息时的加载状态
      
      await chatViewModel.createNewConversation('加载测试对话');
      await driver.delayMs(500);
      
      // 输入消息
      const messageInput = await driver.findComponent(ON.id('message_input'));
      await messageInput.inputText('测试加载状态');
      
      // 点击发送按钮
      const sendButton = await driver.findComponent(ON.id('send_button'));
      await sendButton.click();
      
      // 立即检查加载状态（在消息发送完成前）
      await driver.delayMs(100);
      expect(chatViewModel.uiState).assertEqual(UIState.SENDING);
      expect(chatViewModel.isInputEnabled).assertFalse();
      
      // 等待发送完成
      await driver.delayMs(2000);
      expect(chatViewModel.uiState).assertEqual(UIState.IDLE);
      expect(chatViewModel.isInputEnabled).assertTrue();
    });

    it('should handle error display in UI', async () => {
      // 测试UI中的错误显示
      
      await chatViewModel.createNewConversation('错误测试对话');
      await driver.delayMs(500);
      
      // 模拟网络错误
      const originalSendMessage = chatRepository.sendMessage;
      chatRepository.sendMessage = async () => {
        throw new Error('网络连接失败');
      };
      
      // 尝试发送消息
      const messageInput = await driver.findComponent(ON.id('message_input'));
      await messageInput.inputText('这条消息会失败');
      
      const sendButton = await driver.findComponent(ON.id('send_button'));
      await sendButton.click();
      
      // 等待错误处理
      await driver.delayMs(1000);
      
      // 验证错误提示显示
      expect(chatViewModel.hasError).assertTrue();
      const errorBar = await driver.findComponent(ON.id('error_bar'));
      expect(await errorBar.isEnabled()).assertTrue();
      
      const errorMessage = await driver.findComponent(ON.text('发送消息失败'));
      expect(await errorMessage.isEnabled()).assertTrue();
      
      // 点击重试按钮
      const retryButton = await driver.findComponent(ON.text('重试'));
      await retryButton.click();
      
      // 恢复原始方法
      chatRepository.sendMessage = originalSendMessage;
    });

    it('should open and interact with conversation list dialog', async () => {
      // 测试对话列表弹窗交互
      
      // 创建几个测试对话
      await chatViewModel.createNewConversation('对话1');
      await driver.delayMs(300);
      await chatViewModel.createNewConversation('对话2');
      await driver.delayMs(300);
      
      // 点击对话列表按钮
      const conversationListButton = await driver.findComponent(ON.id('conversation_list_button'));
      await conversationListButton.click();
      
      // 验证对话列表弹窗显示
      expect(chatViewModel.isConversationListVisible).assertTrue();
      
      const conversationDialog = await driver.findComponent(ON.id('conversation_list_dialog'));
      expect(await conversationDialog.isEnabled()).assertTrue();
      
      // 验证对话项显示
      const conversation1 = await driver.findComponent(ON.text('对话1'));
      expect(await conversation1.isEnabled()).assertTrue();
      
      const conversation2 = await driver.findComponent(ON.text('对话2'));
      expect(await conversation2.isEnabled()).assertTrue();
      
      // 选择一个对话
      await conversation1.click();
      
      // 验证对话切换
      await driver.delayMs(500);
      expect(chatViewModel.currentConversation?.title).assertEqual('对话1');
      expect(chatViewModel.isConversationListVisible).assertFalse();
    });

    it('should open and interact with AI personality dialog', async () => {
      // 测试AI人格选择弹窗交互
      
      await chatViewModel.createNewConversation('人格测试对话');
      await driver.delayMs(500);
      
      // 点击AI人格按钮
      const personalityButton = await driver.findComponent(ON.id('ai_personality_button'));
      await personalityButton.click();
      
      // 验证AI人格弹窗显示
      expect(chatViewModel.isAIPersonalityDialogVisible).assertTrue();
      
      const personalityDialog = await driver.findComponent(ON.id('ai_personality_dialog'));
      expect(await personalityDialog.isEnabled()).assertTrue();
      
      // 选择创意人格
      const creativePersonality = await driver.findComponent(ON.text('创意助手'));
      await creativePersonality.click();
      
      // 验证人格切换
      await driver.delayMs(300);
      expect(chatViewModel.selectedAIPersonality).assertEqual(AIPersonality.CREATIVE);
      expect(chatViewModel.isAIPersonalityDialogVisible).assertFalse();
      
      // 验证UI显示更新
      const personalityDisplay = await driver.findComponent(ON.text('创意助手'));
      expect(await personalityDisplay.isEnabled()).assertTrue();
    });

    it('should handle message retry through UI', async () => {
      // 测试通过UI重试消息
      
      await chatViewModel.createNewConversation('重试测试对话');
      await driver.delayMs(500);
      
      // 创建一个失败的消息
      const failedMessage: ChatMessage = {
        messageId: 'failed-message-id',
        conversationId: chatViewModel.currentConversation!.conversationId,
        content: '这是一条失败的消息',
        sender: MessageSender.USER,
        type: MessageType.TEXT,
        status: MessageStatus.FAILED,
        timestamp: Date.now(),
        aiPersonality: AIPersonality.DEFAULT
      };
      
      await chatRepository.saveMessage(failedMessage);
      await chatViewModel.loadMessages(chatViewModel.currentConversation!.conversationId);
      await driver.delayMs(500);
      
      // 查找并点击重试按钮
      const retryButton = await driver.findComponent(ON.id(`retry_button_${failedMessage.messageId}`));
      await retryButton.click();
      
      // 等待重试完成
      await driver.delayMs(1000);
      
      // 验证重试状态
      expect(chatViewModel.uiState).assertEqual(UIState.IDLE);
      expect(chatViewModel.hasError).assertFalse();
    });

    it('should display welcome message for new conversation', async () => {
      // 测试新对话的欢迎消息显示
      
      await chatViewModel.createNewConversation('欢迎消息测试');
      await driver.delayMs(1000);
      
      // 验证欢迎消息显示
      const welcomeMessage = await driver.findComponent(ON.text('你好！我是AI助手'));
      expect(await welcomeMessage.isEnabled()).assertTrue();
      
      // 验证消息来源是AI
      const aiAvatar = await driver.findComponent(ON.id('ai_avatar'));
      expect(await aiAvatar.isEnabled()).assertTrue();
    });

    it('should handle navigation back correctly', async () => {
      // 测试返回导航处理
      
      await chatViewModel.createNewConversation('导航测试对话');
      await driver.delayMs(500);
      
      // 发送一条消息
      const messageInput = await driver.findComponent(ON.id('message_input'));
      await messageInput.inputText('测试导航');
      
      const sendButton = await driver.findComponent(ON.id('send_button'));
      await sendButton.click();
      
      await driver.delayMs(1000);
      
      // 模拟返回按键
      await driver.pressBack();
      
      // 验证状态保存
      expect(chatViewModel.currentConversation).not.toBeNull();
      
      // 重新进入应用
      await driver.delayMs(500);
      
      // 验证状态恢复
      const conversationTitle = await driver.findComponent(ON.text(chatViewModel.currentConversation!.title));
      expect(await conversationTitle.isEnabled()).assertTrue();
      
      const testMessage = await driver.findComponent(ON.text('测试导航'));
      expect(await testMessage.isEnabled()).assertTrue();
    });

    it('should handle long message content correctly', async () => {
      // 测试长消息内容处理
      
      await chatViewModel.createNewConversation('长消息测试');
      await driver.delayMs(500);
      
      const longMessage = '这是一条非常长的消息内容，用来测试UI组件是否能够正确处理和显示长文本内容。'.repeat(10);
      
      // 输入长消息
      const messageInput = await driver.findComponent(ON.id('message_input'));
      await messageInput.inputText(longMessage);
      
      const sendButton = await driver.findComponent(ON.id('send_button'));
      await sendButton.click();
      
      await driver.delayMs(2000);
      
      // 验证长消息显示
      const messageContent = await driver.findComponent(ON.textContains('这是一条非常长的消息内容'));
      expect(await messageContent.isEnabled()).assertTrue();
      
      // 验证消息布局没有溢出
      const messageContainer = await driver.findComponent(ON.id('message_container'));
      expect(await messageContainer.isEnabled()).assertTrue();
    });

    it('should maintain scroll position during message updates', async () => {
      // 测试消息更新时滚动位置保持
      
      await chatViewModel.createNewConversation('滚动测试对话');
      await driver.delayMs(500);
      
      // 发送多条消息以创建可滚动内容
      for (let i = 1; i <= 10; i++) {
        const messageInput = await driver.findComponent(ON.id('message_input'));
        await messageInput.inputText(`测试消息 ${i}`);
        
        const sendButton = await driver.findComponent(ON.id('send_button'));
        await sendButton.click();
        
        await driver.delayMs(500);
      }
      
      // 滚动到中间位置
      const messageList = await driver.findComponent(ON.type('List'));
      await messageList.scrollToTop();
      await driver.delayMs(300);
      
      // 发送新消息
      const messageInput = await driver.findComponent(ON.id('message_input'));
      await messageInput.inputText('新消息');
      
      const sendButton = await driver.findComponent(ON.id('send_button'));
      await sendButton.click();
      
      await driver.delayMs(1000);
      
      // 验证自动滚动到底部显示新消息
      const newMessage = await driver.findComponent(ON.text('新消息'));
      expect(await newMessage.isEnabled()).assertTrue();
    });
  });
}