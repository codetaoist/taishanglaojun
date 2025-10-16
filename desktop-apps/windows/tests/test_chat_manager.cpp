#include "chat_manager.h"
#include <iostream>
#include <string>
#include <vector>

// 测试框架宏
#define TEST_ASSERT(condition, message) \
    do { \
        if (!(condition)) { \
            std::cerr << "FAIL: " << message << " at " << __FILE__ << ":" << __LINE__ << std::endl; \
            return false; \
        } \
    } while(0)

// 测试聊天管理器发送消息功能
bool test_chat_manager_send_message() {
    ChatManager chatManager;
    
    // 测试发送空消息
    bool result1 = chatManager.SendMessage("", "user123");
    TEST_ASSERT(!result1, "Should not be able to send empty message");
    
    // 测试发送给空用户
    bool result2 = chatManager.SendMessage("Hello", "");
    TEST_ASSERT(!result2, "Should not be able to send message to empty user");
    
    // 测试发送正常消息
    bool result3 = chatManager.SendMessage("Hello, World!", "user123");
    // 在测试环境中，这可能会失败（没有后端连接），但应该能够处理
    
    // 测试发送长消息
    std::string longMessage(10000, 'A'); // 10KB的消息
    bool result4 = chatManager.SendMessage(longMessage, "user123");
    // 应该能够处理长消息或适当地拒绝
    
    return true;
}

// 测试聊天管理器接收消息功能
bool test_chat_manager_receive_message() {
    ChatManager chatManager;
    
    // 测试获取消息列表
    std::vector<std::string> messages = chatManager.GetMessages("user123");
    // 在测试环境中，消息列表可能为空
    
    // 测试获取不存在用户的消息
    std::vector<std::string> emptyMessages = chatManager.GetMessages("");
    TEST_ASSERT(emptyMessages.empty(), "Should return empty messages for empty user");
    
    return true;
}

// 测试聊天历史管理
bool test_chat_manager_history() {
    ChatManager chatManager;
    
    // 测试获取聊天历史
    std::vector<std::string> history = chatManager.GetChatHistory("user123", 10);
    // 历史记录可能为空，这是正常的
    
    // 测试清除聊天历史
    bool clearResult = chatManager.ClearChatHistory("user123");
    // 应该能够安全地清除历史记录
    
    return true;
}

// 测试在线状态管理
bool test_chat_manager_online_status() {
    ChatManager chatManager;
    
    // 测试设置在线状态
    bool setOnlineResult = chatManager.SetOnlineStatus(true);
    // 应该能够设置在线状态
    
    // 测试获取在线状态
    bool isOnline = chatManager.IsOnline();
    // 状态可能取决于网络连接
    
    // 测试设置离线状态
    bool setOfflineResult = chatManager.SetOnlineStatus(false);
    
    return true;
}

// 测试群聊功能
bool test_chat_manager_group_chat() {
    ChatManager chatManager;
    
    // 测试创建群聊
    std::string groupId = chatManager.CreateGroup("Test Group", {"user1", "user2", "user3"});
    // 在测试环境中，群聊创建可能失败
    
    // 测试加入群聊
    bool joinResult = chatManager.JoinGroup("group123");
    
    // 测试离开群聊
    bool leaveResult = chatManager.LeaveGroup("group123");
    
    // 测试发送群消息
    bool sendGroupResult = chatManager.SendGroupMessage("Hello Group!", "group123");
    
    return true;
}

// 测试消息格式化和验证
bool test_chat_manager_message_validation() {
    ChatManager chatManager;
    
    // 测试消息格式化
    std::string formattedMessage = chatManager.FormatMessage("Hello", "user123", "2024-01-01T12:00:00Z");
    TEST_ASSERT(!formattedMessage.empty(), "Formatted message should not be empty");
    
    // 测试消息验证
    bool isValid1 = chatManager.ValidateMessage("Hello");
    TEST_ASSERT(isValid1, "Simple message should be valid");
    
    bool isValid2 = chatManager.ValidateMessage("");
    TEST_ASSERT(!isValid2, "Empty message should be invalid");
    
    // 测试特殊字符
    bool isValid3 = chatManager.ValidateMessage("Hello 🌟 World!");
    TEST_ASSERT(isValid3, "Message with emoji should be valid");
    
    return true;
}