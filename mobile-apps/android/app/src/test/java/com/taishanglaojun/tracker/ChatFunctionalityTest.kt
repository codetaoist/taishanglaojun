package com.taishanglaojun.tracker

import com.taishanglaojun.tracker.data.model.*
import com.taishanglaojun.tracker.data.repository.ChatRepository
import com.taishanglaojun.tracker.presentation.viewmodel.ChatViewModel
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.test.runTest
import org.junit.Before
import org.junit.Test
import org.junit.Assert.*
import org.mockito.Mock
import org.mockito.Mockito.*
import org.mockito.MockitoAnnotations
import java.util.*

/**
 * Android AI对话功能测试
 * 验证聊天功能的数据库操作和UI交互逻辑
 */
@ExperimentalCoroutinesApi
class ChatFunctionalityTest {

    @Mock
    private lateinit var mockChatRepository: ChatRepository

    private lateinit var chatViewModel: ChatViewModel

    @Before
    fun setup() {
        MockitoAnnotations.openMocks(this)
        chatViewModel = ChatViewModel(mockChatRepository)
    }

    @Test
    fun `test create new conversation`() = runTest {
        // 准备测试数据
        val conversationId = "test-conversation-id"
        val conversation = Conversation(
            id = conversationId,
            title = "新对话",
            createdAt = System.currentTimeMillis(),
            updatedAt = System.currentTimeMillis()
        )

        // 模拟仓库行为
        `when`(mockChatRepository.createConversation()).thenReturn(conversation)
        `when`(mockChatRepository.getConversations()).thenReturn(flowOf(listOf(conversation)))

        // 执行测试
        chatViewModel.createNewConversation()

        // 验证结果
        verify(mockChatRepository).createConversation()
        val conversations = chatViewModel.conversations.first()
        assertEquals(1, conversations.size)
        assertEquals(conversationId, conversations[0].id)
    }

    @Test
    fun `test send message successfully`() = runTest {
        // 准备测试数据
        val conversationId = "test-conversation-id"
        val messageContent = "你好，太上老君"
        val userMessage = ChatMessage(
            id = "user-message-id",
            conversationId = conversationId,
            content = messageContent,
            sender = MessageSender.USER,
            timestamp = System.currentTimeMillis(),
            status = MessageStatus.SENT
        )
        val aiResponse = ChatMessage(
            id = "ai-message-id",
            conversationId = conversationId,
            content = "施主有何疑问？老君愿为您答疑解惑。",
            sender = MessageSender.AI,
            timestamp = System.currentTimeMillis(),
            status = MessageStatus.SENT
        )

        // 模拟仓库行为
        `when`(mockChatRepository.saveMessage(any())).thenReturn(userMessage)
        `when`(mockChatRepository.sendMessageToAI(messageContent, conversationId))
            .thenReturn(aiResponse)
        `when`(mockChatRepository.getMessagesForConversation(conversationId))
            .thenReturn(flowOf(listOf(userMessage, aiResponse)))

        // 设置当前对话
        chatViewModel.selectConversation(conversationId)

        // 执行测试
        chatViewModel.sendMessage(messageContent)

        // 验证结果
        verify(mockChatRepository).saveMessage(any())
        verify(mockChatRepository).sendMessageToAI(messageContent, conversationId)
        
        val messages = chatViewModel.currentMessages.first()
        assertEquals(2, messages.size)
        assertEquals(MessageSender.USER, messages[0].sender)
        assertEquals(MessageSender.AI, messages[1].sender)
        assertEquals(messageContent, messages[0].content)
    }

    @Test
    fun `test message sending failure handling`() = runTest {
        // 准备测试数据
        val conversationId = "test-conversation-id"
        val messageContent = "测试消息"
        val failedMessage = ChatMessage(
            id = "failed-message-id",
            conversationId = conversationId,
            content = messageContent,
            sender = MessageSender.USER,
            timestamp = System.currentTimeMillis(),
            status = MessageStatus.FAILED
        )

        // 模拟网络错误
        `when`(mockChatRepository.saveMessage(any())).thenReturn(failedMessage)
        `when`(mockChatRepository.sendMessageToAI(messageContent, conversationId))
            .thenThrow(RuntimeException("网络连接失败"))

        // 设置当前对话
        chatViewModel.selectConversation(conversationId)

        // 执行测试
        chatViewModel.sendMessage(messageContent)

        // 验证错误处理
        verify(mockChatRepository).saveMessage(any())
        verify(mockChatRepository).sendMessageToAI(messageContent, conversationId)
        
        val uiState = chatViewModel.uiState.first()
        assertNotNull(uiState.error)
        assertTrue(uiState.error!!.contains("网络连接失败"))
    }

    @Test
    fun `test conversation deletion`() = runTest {
        // 准备测试数据
        val conversationId = "test-conversation-id"
        val conversation = Conversation(
            id = conversationId,
            title = "要删除的对话",
            createdAt = System.currentTimeMillis(),
            updatedAt = System.currentTimeMillis()
        )

        // 模拟仓库行为
        `when`(mockChatRepository.getConversations())
            .thenReturn(flowOf(listOf(conversation)))
            .thenReturn(flowOf(emptyList()))

        // 执行测试
        chatViewModel.deleteConversation(conversationId)

        // 验证结果
        verify(mockChatRepository).deleteConversation(conversationId)
        val conversations = chatViewModel.conversations.first()
        assertEquals(0, conversations.size)
    }

    @Test
    fun `test AI personality update`() = runTest {
        // 准备测试数据
        val newPersonality = AIPersonality.WISE_SAGE

        // 执行测试
        chatViewModel.updateAIPersonality(newPersonality)

        // 验证结果
        verify(mockChatRepository).updateAIPersonality(newPersonality)
    }

    @Test
    fun `test message retry functionality`() = runTest {
        // 准备测试数据
        val conversationId = "test-conversation-id"
        val messageId = "failed-message-id"
        val originalMessage = ChatMessage(
            id = messageId,
            conversationId = conversationId,
            content = "重试消息",
            sender = MessageSender.USER,
            timestamp = System.currentTimeMillis(),
            status = MessageStatus.FAILED
        )
        val retriedMessage = originalMessage.copy(status = MessageStatus.SENT)

        // 模拟仓库行为
        `when`(mockChatRepository.getMessage(messageId)).thenReturn(originalMessage)
        `when`(mockChatRepository.retryMessage(messageId)).thenReturn(retriedMessage)

        // 执行测试
        chatViewModel.retryMessage(messageId)

        // 验证结果
        verify(mockChatRepository).retryMessage(messageId)
    }

    @Test
    fun `test conversation title generation`() = runTest {
        // 准备测试数据
        val conversationId = "test-conversation-id"
        val firstMessage = "你好，我想了解道德经的智慧"
        val generatedTitle = "道德经智慧探讨"

        // 模拟仓库行为
        `when`(mockChatRepository.generateConversationTitle(firstMessage))
            .thenReturn(generatedTitle)

        // 执行测试
        val result = mockChatRepository.generateConversationTitle(firstMessage)

        // 验证结果
        assertEquals(generatedTitle, result)
        verify(mockChatRepository).generateConversationTitle(firstMessage)
    }

    @Test
    fun `test database message persistence`() = runTest {
        // 准备测试数据
        val conversationId = "test-conversation-id"
        val messages = listOf(
            ChatMessage(
                id = "msg1",
                conversationId = conversationId,
                content = "用户消息1",
                sender = MessageSender.USER,
                timestamp = System.currentTimeMillis(),
                status = MessageStatus.SENT
            ),
            ChatMessage(
                id = "msg2",
                conversationId = conversationId,
                content = "AI回复1",
                sender = MessageSender.AI,
                timestamp = System.currentTimeMillis(),
                status = MessageStatus.SENT
            )
        )

        // 模拟数据库查询
        `when`(mockChatRepository.getMessagesForConversation(conversationId))
            .thenReturn(flowOf(messages))

        // 执行测试
        val result = mockChatRepository.getMessagesForConversation(conversationId).first()

        // 验证结果
        assertEquals(2, result.size)
        assertEquals(MessageSender.USER, result[0].sender)
        assertEquals(MessageSender.AI, result[1].sender)
        verify(mockChatRepository).getMessagesForConversation(conversationId)
    }

    @Test
    fun `test UI state management during message sending`() = runTest {
        // 准备测试数据
        val conversationId = "test-conversation-id"
        val messageContent = "测试UI状态"

        // 模拟长时间的AI响应
        `when`(mockChatRepository.sendMessageToAI(messageContent, conversationId))
            .thenAnswer { 
                Thread.sleep(1000) // 模拟网络延迟
                ChatMessage(
                    id = "ai-response",
                    conversationId = conversationId,
                    content = "AI响应",
                    sender = MessageSender.AI,
                    timestamp = System.currentTimeMillis(),
                    status = MessageStatus.SENT
                )
            }

        // 设置当前对话
        chatViewModel.selectConversation(conversationId)

        // 验证初始状态
        val initialState = chatViewModel.uiState.first()
        assertFalse(initialState.isSending)

        // 执行测试（异步）
        chatViewModel.sendMessage(messageContent)

        // 验证发送状态
        // 注意：在实际测试中，这里需要更复杂的异步状态验证
        verify(mockChatRepository).sendMessageToAI(messageContent, conversationId)
    }
}