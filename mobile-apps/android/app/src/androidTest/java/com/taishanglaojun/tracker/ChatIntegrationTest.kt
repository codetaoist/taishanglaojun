package com.taishanglaojun.tracker

import androidx.compose.ui.test.*
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.test.ext.junit.runners.AndroidJUnit4
import com.taishanglaojun.tracker.data.model.*
import com.taishanglaojun.tracker.presentation.ui.screen.ChatScreen
import com.taishanglaojun.tracker.presentation.viewmodel.ChatViewModel
import dagger.hilt.android.testing.HiltAndroidRule
import dagger.hilt.android.testing.HiltAndroidTest
import kotlinx.coroutines.flow.MutableStateFlow
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import org.mockito.Mock
import org.mockito.Mockito.*
import org.mockito.MockitoAnnotations
import javax.inject.Inject

/**
 * Android AI对话功能集成测试
 * 验证UI交互和数据库集成的完整流程
 */
@HiltAndroidTest
@RunWith(AndroidJUnit4::class)
class ChatIntegrationTest {

    @get:Rule
    val hiltRule = HiltAndroidRule(this)

    @get:Rule
    val composeTestRule = createComposeRule()

    @Mock
    private lateinit var mockViewModel: ChatViewModel

    private val testMessages = MutableStateFlow(emptyList<ChatMessage>())
    private val testConversations = MutableStateFlow(emptyList<Conversation>())
    private val testCurrentConversationId = MutableStateFlow<String?>(null)
    private val testUiState = MutableStateFlow(ChatUiState())

    @Before
    fun setup() {
        MockitoAnnotations.openMocks(this)
        hiltRule.inject()

        // 模拟ViewModel的Flow
        `when`(mockViewModel.currentMessages).thenReturn(testMessages)
        `when`(mockViewModel.conversations).thenReturn(testConversations)
        `when`(mockViewModel.currentConversationId).thenReturn(testCurrentConversationId)
        `when`(mockViewModel.uiState).thenReturn(testUiState)
    }

    @Test
    fun testChatScreenInitialState() {
        composeTestRule.setContent {
            ChatScreen(
                viewModel = mockViewModel,
                onNavigateBack = {}
            )
        }

        // 验证初始UI状态
        composeTestRule.onNodeWithText("AI对话").assertIsDisplayed()
        composeTestRule.onNodeWithContentDescription("返回").assertIsDisplayed()
        composeTestRule.onNodeWithContentDescription("对话列表").assertIsDisplayed()
        composeTestRule.onNodeWithContentDescription("AI个性").assertIsDisplayed()
        composeTestRule.onNodeWithContentDescription("新对话").assertIsDisplayed()
    }

    @Test
    fun testCreateNewConversation() {
        val testConversation = Conversation(
            id = "test-conv-1",
            title = "新对话",
            createdAt = System.currentTimeMillis(),
            updatedAt = System.currentTimeMillis()
        )

        composeTestRule.setContent {
            ChatScreen(
                viewModel = mockViewModel,
                onNavigateBack = {}
            )
        }

        // 点击新对话按钮
        composeTestRule.onNodeWithContentDescription("新对话").performClick()

        // 验证ViewModel方法被调用
        verify(mockViewModel).createNewConversation()

        // 模拟对话创建成功
        testConversations.value = listOf(testConversation)
        testCurrentConversationId.value = testConversation.id

        // 验证UI更新
        composeTestRule.onNodeWithText("新对话").assertIsDisplayed()
    }

    @Test
    fun testSendMessage() {
        val conversationId = "test-conv-1"
        val messageContent = "你好，太上老君"

        // 设置测试状态
        testCurrentConversationId.value = conversationId

        composeTestRule.setContent {
            ChatScreen(
                viewModel = mockViewModel,
                onNavigateBack = {}
            )
        }

        // 输入消息
        composeTestRule.onNodeWithText("输入消息...").performTextInput(messageContent)
        
        // 点击发送按钮
        composeTestRule.onNodeWithContentDescription("发送").performClick()

        // 验证ViewModel方法被调用
        verify(mockViewModel).sendMessage(messageContent)
    }

    @Test
    fun testMessageDisplay() {
        val conversationId = "test-conv-1"
        val userMessage = ChatMessage(
            id = "msg-1",
            conversationId = conversationId,
            content = "用户消息测试",
            sender = MessageSender.USER,
            timestamp = System.currentTimeMillis(),
            status = MessageStatus.SENT
        )
        val aiMessage = ChatMessage(
            id = "msg-2",
            conversationId = conversationId,
            content = "AI回复测试",
            sender = MessageSender.AI,
            timestamp = System.currentTimeMillis(),
            status = MessageStatus.SENT
        )

        // 设置测试数据
        testCurrentConversationId.value = conversationId
        testMessages.value = listOf(userMessage, aiMessage)

        composeTestRule.setContent {
            ChatScreen(
                viewModel = mockViewModel,
                onNavigateBack = {}
            )
        }

        // 验证消息显示
        composeTestRule.onNodeWithText("用户消息测试").assertIsDisplayed()
        composeTestRule.onNodeWithText("AI回复测试").assertIsDisplayed()
    }

    @Test
    fun testLoadingState() {
        val conversationId = "test-conv-1"
        
        // 设置加载状态
        testCurrentConversationId.value = conversationId
        testUiState.value = ChatUiState(isSending = true)

        composeTestRule.setContent {
            ChatScreen(
                viewModel = mockViewModel,
                onNavigateBack = {}
            )
        }

        // 验证加载指示器显示
        composeTestRule.onNode(hasContentDescription("发送"))
            .assertExists()
        
        // 验证输入框被禁用
        composeTestRule.onNodeWithText("输入消息...")
            .assertIsNotEnabled()
    }

    @Test
    fun testErrorHandling() {
        val errorMessage = "网络连接失败，请检查网络设置"
        
        // 设置错误状态
        testUiState.value = ChatUiState(error = errorMessage)

        composeTestRule.setContent {
            ChatScreen(
                viewModel = mockViewModel,
                onNavigateBack = {}
            )
        }

        // 验证错误消息显示
        composeTestRule.onNodeWithText(errorMessage).assertIsDisplayed()
        composeTestRule.onNodeWithContentDescription("关闭").assertIsDisplayed()

        // 点击关闭错误消息
        composeTestRule.onNodeWithContentDescription("关闭").performClick()
        
        // 验证清除错误方法被调用
        verify(mockViewModel).clearError()
    }

    @Test
    fun testConversationListDialog() {
        val conversations = listOf(
            Conversation(
                id = "conv-1",
                title = "对话1",
                createdAt = System.currentTimeMillis(),
                updatedAt = System.currentTimeMillis()
            ),
            Conversation(
                id = "conv-2", 
                title = "对话2",
                createdAt = System.currentTimeMillis(),
                updatedAt = System.currentTimeMillis()
            )
        )

        testConversations.value = conversations

        composeTestRule.setContent {
            ChatScreen(
                viewModel = mockViewModel,
                onNavigateBack = {}
            )
        }

        // 打开对话列表
        composeTestRule.onNodeWithContentDescription("对话列表").performClick()

        // 验证对话列表显示
        composeTestRule.onNodeWithText("对话列表").assertIsDisplayed()
        composeTestRule.onNodeWithText("对话1").assertIsDisplayed()
        composeTestRule.onNodeWithText("对话2").assertIsDisplayed()

        // 选择对话
        composeTestRule.onNodeWithText("对话1").performClick()
        
        // 验证选择对话方法被调用
        verify(mockViewModel).selectConversation("conv-1")
    }

    @Test
    fun testAIPersonalityDialog() {
        composeTestRule.setContent {
            ChatScreen(
                viewModel = mockViewModel,
                onNavigateBack = {}
            )
        }

        // 打开AI个性选择
        composeTestRule.onNodeWithContentDescription("AI个性").performClick()

        // 验证个性选择对话框显示
        composeTestRule.onNodeWithText("选择AI个性").assertIsDisplayed()
        composeTestRule.onNodeWithText("默认").assertIsDisplayed()
        composeTestRule.onNodeWithText("智慧长者").assertIsDisplayed()
        composeTestRule.onNodeWithText("友善向导").assertIsDisplayed()

        // 选择个性
        composeTestRule.onNodeWithText("智慧长者").performClick()
        
        // 验证更新个性方法被调用
        verify(mockViewModel).updateAIPersonality(AIPersonality.WISE_SAGE)
    }

    @Test
    fun testMessageRetry() {
        val conversationId = "test-conv-1"
        val failedMessage = ChatMessage(
            id = "failed-msg",
            conversationId = conversationId,
            content = "失败的消息",
            sender = MessageSender.USER,
            timestamp = System.currentTimeMillis(),
            status = MessageStatus.FAILED
        )

        testCurrentConversationId.value = conversationId
        testMessages.value = listOf(failedMessage)

        composeTestRule.setContent {
            ChatScreen(
                viewModel = mockViewModel,
                onNavigateBack = {}
            )
        }

        // 验证失败消息显示
        composeTestRule.onNodeWithText("失败的消息").assertIsDisplayed()
        
        // 点击重试按钮
        composeTestRule.onNodeWithContentDescription("重试").performClick()
        
        // 验证重试方法被调用
        verify(mockViewModel).retryMessage("failed-msg")
    }

    @Test
    fun testWelcomeMessage() {
        val conversationId = "test-conv-1"
        
        // 设置空消息列表
        testCurrentConversationId.value = conversationId
        testMessages.value = emptyList()

        composeTestRule.setContent {
            ChatScreen(
                viewModel = mockViewModel,
                onNavigateBack = {}
            )
        }

        // 验证欢迎消息显示
        composeTestRule.onNodeWithText("欢迎使用太上老君AI助手").assertIsDisplayed()
        composeTestRule.onNodeWithText("我将以中华文化智慧为您答疑解惑").assertIsDisplayed()
    }

    @Test
    fun testNavigationBack() {
        var backPressed = false

        composeTestRule.setContent {
            ChatScreen(
                viewModel = mockViewModel,
                onNavigateBack = { backPressed = true }
            )
        }

        // 点击返回按钮
        composeTestRule.onNodeWithContentDescription("返回").performClick()

        // 验证导航回调被调用
        assert(backPressed)
    }
}