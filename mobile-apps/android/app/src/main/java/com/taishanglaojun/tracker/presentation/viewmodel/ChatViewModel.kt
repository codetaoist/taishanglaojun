package com.taishanglaojun.tracker.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.taishanglaojun.tracker.data.model.*
import com.taishanglaojun.tracker.data.repository.ChatRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class ChatViewModel @Inject constructor(
    private val chatRepository: ChatRepository
) : ViewModel() {
    
    private val _uiState = MutableStateFlow(ChatUiState())
    val uiState: StateFlow<ChatUiState> = _uiState.asStateFlow()
    
    private val _currentConversationId = MutableStateFlow<String?>(null)
    val currentConversationId: StateFlow<String?> = _currentConversationId.asStateFlow()
    
    // 当前对话的消息列表
    val currentMessages: StateFlow<List<ChatMessage>> = _currentConversationId
        .filterNotNull()
        .flatMapLatest { conversationId ->
            chatRepository.getMessagesByConversation(conversationId)
        }
        .stateIn(
            scope = viewModelScope,
            started = SharingStarted.WhileSubscribed(5000),
            initialValue = emptyList()
        )
    
    // 所有对话列表
    val conversations: StateFlow<List<Conversation>> = chatRepository.getAllConversations()
        .stateIn(
            scope = viewModelScope,
            started = SharingStarted.WhileSubscribed(5000),
            initialValue = emptyList()
        )
    
    init {
        // 初始化时同步数据
        viewModelScope.launch {
            syncData()
        }
    }
    
    fun createNewConversation(title: String = "新对话", aiPersonality: AIPersonality = AIPersonality.DEFAULT) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true)
            
            val result = chatRepository.createConversation(title, aiPersonality)
            if (result.isSuccess) {
                val conversation = result.getOrThrow()
                _currentConversationId.value = conversation.id
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    error = null
                )
            } else {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    error = result.exceptionOrNull()?.message ?: "创建对话失败"
                )
            }
        }
    }
    
    fun selectConversation(conversationId: String) {
        _currentConversationId.value = conversationId
        _uiState.value = _uiState.value.copy(error = null)
    }
    
    fun sendMessage(content: String, messageType: MessageType = MessageType.TEXT) {
        val conversationId = _currentConversationId.value
        if (conversationId == null) {
            _uiState.value = _uiState.value.copy(error = "请先选择或创建对话")
            return
        }
        
        if (content.isBlank()) {
            _uiState.value = _uiState.value.copy(error = "消息内容不能为空")
            return
        }
        
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isSending = true, error = null)
            
            val result = chatRepository.sendMessage(conversationId, content, messageType)
            if (result.isSuccess) {
                _uiState.value = _uiState.value.copy(
                    isSending = false,
                    lastSentMessage = content
                )
            } else {
                _uiState.value = _uiState.value.copy(
                    isSending = false,
                    error = result.exceptionOrNull()?.message ?: "发送消息失败"
                )
            }
        }
    }
    
    fun sendImageMessage(imageUri: String, caption: String = "") {
        val conversationId = _currentConversationId.value
        if (conversationId == null) {
            _uiState.value = _uiState.value.copy(error = "请先选择或创建对话")
            return
        }
        
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isSending = true, error = null)
            
            val result = chatRepository.sendImageMessage(conversationId, imageUri, caption)
            if (result.isSuccess) {
                _uiState.value = _uiState.value.copy(
                    isSending = false,
                    lastSentMessage = "图片消息"
                )
            } else {
                _uiState.value = _uiState.value.copy(
                    isSending = false,
                    error = result.exceptionOrNull()?.message ?: "发送图片失败"
                )
            }
        }
    }
    
    fun sendAudioMessage(audioUri: String, duration: Long) {
        val conversationId = _currentConversationId.value
        if (conversationId == null) {
            _uiState.value = _uiState.value.copy(error = "请先选择或创建对话")
            return
        }
        
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isSending = true, error = null)
            
            val result = chatRepository.sendAudioMessage(conversationId, audioUri, duration)
            if (result.isSuccess) {
                _uiState.value = _uiState.value.copy(
                    isSending = false,
                    lastSentMessage = "语音消息"
                )
            } else {
                _uiState.value = _uiState.value.copy(
                    isSending = false,
                    error = result.exceptionOrNull()?.message ?: "发送语音失败"
                )
            }
        }
    }
    
    fun retryMessage(messageId: String) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true)
            
            val result = chatRepository.retryMessage(messageId)
            if (result.isSuccess) {
                _uiState.value = _uiState.value.copy(isLoading = false, error = null)
            } else {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    error = result.exceptionOrNull()?.message ?: "重试失败"
                )
            }
        }
    }
    
    fun deleteConversation(conversationId: String) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true)
            
            val result = chatRepository.deleteConversation(conversationId)
            if (result.isSuccess) {
                if (_currentConversationId.value == conversationId) {
                    _currentConversationId.value = null
                }
                _uiState.value = _uiState.value.copy(isLoading = false, error = null)
            } else {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    error = result.exceptionOrNull()?.message ?: "删除对话失败"
                )
            }
        }
    }
    
    fun archiveConversation(conversationId: String) {
        viewModelScope.launch {
            chatRepository.archiveConversation(conversationId)
            if (_currentConversationId.value == conversationId) {
                _currentConversationId.value = null
            }
        }
    }
    
    fun searchMessages(query: String) {
        val conversationId = _currentConversationId.value ?: return
        
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isSearching = true)
            
            try {
                val results = chatRepository.searchMessages(conversationId, query)
                _uiState.value = _uiState.value.copy(
                    isSearching = false,
                    searchResults = results,
                    searchQuery = query
                )
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isSearching = false,
                    error = "搜索失败: ${e.message}"
                )
            }
        }
    }
    
    fun clearSearch() {
        _uiState.value = _uiState.value.copy(
            searchResults = emptyList(),
            searchQuery = ""
        )
    }
    
    fun clearError() {
        _uiState.value = _uiState.value.copy(error = null)
    }
    
    private suspend fun syncData() {
        try {
            chatRepository.syncConversations()
        } catch (e: Exception) {
            // 同步失败不影响正常使用
        }
    }
    
    fun refreshData() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isRefreshing = true)
            
            try {
                syncData()
                _currentConversationId.value?.let { conversationId ->
                    chatRepository.syncMessages(conversationId)
                }
                _uiState.value = _uiState.value.copy(isRefreshing = false, error = null)
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isRefreshing = false,
                    error = "刷新失败: ${e.message}"
                )
            }
        }
    }
    
    fun updateAIPersonality(personality: AIPersonality) {
        val conversationId = _currentConversationId.value ?: return
        
        viewModelScope.launch {
            try {
                val conversation = chatRepository.getConversationById(conversationId)
                conversation?.let {
                    val updatedConversation = it.copy(aiPersonality = personality)
                    chatRepository.updateConversation(updatedConversation)
                }
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(error = "更新AI个性失败: ${e.message}")
            }
        }
    }
}

data class ChatUiState(
    val isLoading: Boolean = false,
    val isSending: Boolean = false,
    val isRefreshing: Boolean = false,
    val isSearching: Boolean = false,
    val error: String? = null,
    val lastSentMessage: String? = null,
    val searchQuery: String = "",
    val searchResults: List<ChatMessage> = emptyList()
)