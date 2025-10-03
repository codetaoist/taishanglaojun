/**
 * AI聊天页面
 * 提供完整的AI对话界面
 */

import { 
  ChatMessage, 
  Conversation, 
  MessageSender, 
  MessageStatus, 
  MessageType,
  AIPersonality 
} from '../model/ChatMessage';
import { ChatViewModel, UIState } from '../viewmodel/ChatViewModel';

@Entry
@Component
struct ChatPage {
  @State private chatViewModel: ChatViewModel = ChatViewModel.getInstance();
  @State private scrollController: Scroller = new Scroller();
  @State private inputController: TextInputController = new TextInputController();

  aboutToAppear() {
    // 页面即将出现时的初始化
  }

  aboutToDisappear() {
    // 页面即将消失时的清理
  }

  build() {
    Column() {
      // 顶部导航栏
      this.buildTopBar()
      
      // 错误提示
      if (this.chatViewModel.hasError) {
        this.buildErrorBar()
      }
      
      // 主要内容区域
      if (this.chatViewModel.currentConversation) {
        this.buildChatContent()
      } else {
        this.buildWelcomeScreen()
      }
      
      // 底部输入区域
      if (this.chatViewModel.currentConversation) {
        this.buildInputArea()
      }
    }
    .width('100%')
    .height('100%')
    .backgroundColor('#F5F5F5')
    .overlay(
      // 对话列表弹窗
      this.buildConversationListDialog(),
      { align: Alignment.Center }
    )
    .overlay(
      // AI人格选择弹窗
      this.buildAIPersonalityDialog(),
      { align: Alignment.Center }
    )
  }

  // 构建顶部导航栏
  @Builder
  buildTopBar() {
    Row() {
      // 对话列表按钮
      Button() {
        Image($r('app.media.ic_list'))
          .width(24)
          .height(24)
          .fillColor(Color.White)
      }
      .width(40)
      .height(40)
      .backgroundColor('#007AFF')
      .borderRadius(20)
      .onClick(() => {
        this.chatViewModel.showConversationList();
      })

      // 标题区域
      Column() {
        Text(this.chatViewModel.currentConversation?.title || 'AI助手')
          .fontSize(18)
          .fontWeight(FontWeight.Medium)
          .fontColor('#333333')
          .maxLines(1)
          .textOverflow({ overflow: TextOverflow.Ellipsis })
        
        if (this.chatViewModel.currentConversation) {
          Text(this.chatViewModel.aiPersonalityDisplayName)
            .fontSize(12)
            .fontColor('#666666')
            .margin({ top: 2 })
        }
      }
      .layoutWeight(1)
      .alignItems(HorizontalAlign.Center)
      .margin({ left: 12, right: 12 })

      // 右侧按钮组
      Row({ space: 8 }) {
        // AI人格选择按钮
        Button() {
          Image($r('app.media.ic_personality'))
            .width(20)
            .height(20)
            .fillColor('#666666')
        }
        .width(36)
        .height(36)
        .backgroundColor(Color.Transparent)
        .borderRadius(18)
        .onClick(() => {
          this.chatViewModel.showAIPersonalityDialog();
        })

        // 新建对话按钮
        Button() {
          Image($r('app.media.ic_add'))
            .width(20)
            .height(20)
            .fillColor('#666666')
        }
        .width(36)
        .height(36)
        .backgroundColor(Color.Transparent)
        .borderRadius(18)
        .onClick(() => {
          this.chatViewModel.createNewConversation();
        })
      }
    }
    .width('100%')
    .height(56)
    .padding({ left: 16, right: 16 })
    .backgroundColor(Color.White)
    .shadow({
      radius: 4,
      color: '#1A000000',
      offsetX: 0,
      offsetY: 2
    })
  }

  // 构建错误提示栏
  @Builder
  buildErrorBar() {
    Row() {
      Image($r('app.media.ic_error'))
        .width(16)
        .height(16)
        .fillColor('#FF3B30')
        .margin({ right: 8 })
      
      Text(this.chatViewModel.errorInfo?.message || '发生错误')
        .fontSize(14)
        .fontColor('#FF3B30')
        .layoutWeight(1)
      
      Button('重试')
        .fontSize(12)
        .fontColor('#007AFF')
        .backgroundColor(Color.Transparent)
        .padding({ left: 8, right: 8, top: 4, bottom: 4 })
        .onClick(() => {
          this.chatViewModel.refreshData();
        })
    }
    .width('100%')
    .padding({ left: 16, right: 16, top: 8, bottom: 8 })
    .backgroundColor('#FFF2F2')
    .border({ width: { bottom: 1 }, color: '#FFE6E6' })
  }

  // 构建聊天内容区域
  @Builder
  buildChatContent() {
    Column() {
      if (this.chatViewModel.isLoading) {
        // 加载状态
        this.buildLoadingView()
      } else if (this.chatViewModel.displayMessages.length === 0) {
        // 空状态
        this.buildEmptyView()
      } else {
        // 消息列表
        this.buildMessageList()
      }
    }
    .layoutWeight(1)
    .width('100%')
  }

  // 构建加载视图
  @Builder
  buildLoadingView() {
    Column() {
      LoadingProgress()
        .width(40)
        .height(40)
        .color('#007AFF')
      
      Text('加载中...')
        .fontSize(14)
        .fontColor('#666666')
        .margin({ top: 12 })
    }
    .width('100%')
    .height('100%')
    .justifyContent(FlexAlign.Center)
    .alignItems(HorizontalAlign.Center)
  }

  // 构建空状态视图
  @Builder
  buildEmptyView() {
    Column() {
      Image($r('app.media.ic_chat_empty'))
        .width(80)
        .height(80)
        .fillColor('#CCCCCC')
      
      Text('开始对话吧')
        .fontSize(16)
        .fontColor('#666666')
        .margin({ top: 16 })
      
      Text('向AI助手提问任何问题')
        .fontSize(14)
        .fontColor('#999999')
        .margin({ top: 8 })
    }
    .width('100%')
    .height('100%')
    .justifyContent(FlexAlign.Center)
    .alignItems(HorizontalAlign.Center)
  }

  // 构建消息列表
  @Builder
  buildMessageList() {
    List({ scroller: this.scrollController }) {
      ForEach(this.chatViewModel.displayMessages, (message: ChatMessage, index: number) => {
        ListItem() {
          this.buildMessageItem(message)
        }
        .margin({ bottom: index === this.chatViewModel.displayMessages.length - 1 ? 16 : 8 })
      })
    }
    .width('100%')
    .layoutWeight(1)
    .padding({ left: 16, right: 16, top: 16 })
    .scrollBar(BarState.Off)
    .onScrollToIndex((first: number, last: number) => {
      // 滚动到底部时自动标记消息为已读
      if (last === this.chatViewModel.displayMessages.length - 1) {
        // TODO: 标记消息为已读
      }
    })
  }

  // 构建消息项
  @Builder
  buildMessageItem(message: ChatMessage) {
    Row() {
      if (message.sender === MessageSender.USER) {
        // 用户消息 - 右对齐
        Blank()
        
        Column() {
          Text(message.content)
            .fontSize(16)
            .fontColor(Color.White)
            .padding({ left: 12, right: 12, top: 8, bottom: 8 })
            .backgroundColor('#007AFF')
            .borderRadius(12)
            .maxLines(999)
          
          Text(this.chatViewModel.formatTime(message.timestamp))
            .fontSize(12)
            .fontColor('#999999')
            .margin({ top: 4 })
            .alignSelf(ItemAlign.End)
        }
        .alignItems(HorizontalAlign.End)
        .constraintSize({ maxWidth: '70%' })
      } else {
        // AI消息 - 左对齐
        Row() {
          // AI头像
          Image($r('app.media.ic_ai_avatar'))
            .width(32)
            .height(32)
            .borderRadius(16)
            .margin({ right: 8, top: 4 })
          
          Column() {
            // 消息内容
            Text(message.content || '思考中...')
              .fontSize(16)
              .fontColor('#333333')
              .padding({ left: 12, right: 12, top: 8, bottom: 8 })
              .backgroundColor(Color.White)
              .borderRadius(12)
              .maxLines(999)
              .shadow({
                radius: 2,
                color: '#1A000000',
                offsetX: 0,
                offsetY: 1
              })
            
            // 状态和时间
            Row() {
              Text(this.chatViewModel.formatTime(message.timestamp))
                .fontSize(12)
                .fontColor('#999999')
              
              if (message.status !== MessageStatus.SENT) {
                Text(this.chatViewModel.getMessageStatusText(message.status))
                  .fontSize(12)
                  .fontColor(this.chatViewModel.getMessageStatusColor(message.status))
                  .margin({ left: 8 })
              }
              
              // 重试按钮
              if (message.status === MessageStatus.FAILED) {
                Button('重试')
                  .fontSize(12)
                  .fontColor('#007AFF')
                  .backgroundColor(Color.Transparent)
                  .padding({ left: 8, right: 8, top: 2, bottom: 2 })
                  .margin({ left: 8 })
                  .onClick(() => {
                    this.chatViewModel.retryMessage(message.messageId);
                  })
              }
            }
            .margin({ top: 4 })
            .alignItems(VerticalAlign.Center)
          }
          .alignItems(HorizontalAlign.Start)
          .constraintSize({ maxWidth: '70%' })
        }
        .alignItems(VerticalAlign.Top)
        
        Blank()
      }
    }
    .width('100%')
    .alignItems(VerticalAlign.Top)
    .gesture(
      // 长按菜单
      LongPressGesture({ repeat: false })
        .onAction(() => {
          this.showMessageMenu(message);
        })
    )
  }

  // 构建欢迎屏幕
  @Builder
  buildWelcomeScreen() {
    Column() {
      Image($r('app.media.ic_ai_welcome'))
        .width(120)
        .height(120)
        .fillColor('#007AFF')
      
      Text('欢迎使用AI助手')
        .fontSize(24)
        .fontWeight(FontWeight.Medium)
        .fontColor('#333333')
        .margin({ top: 24 })
      
      Text('选择一个AI人格，开始您的智慧对话之旅')
        .fontSize(16)
        .fontColor('#666666')
        .margin({ top: 12 })
        .textAlign(TextAlign.Center)
      
      Button('开始对话')
        .fontSize(16)
        .fontColor(Color.White)
        .backgroundColor('#007AFF')
        .borderRadius(24)
        .padding({ left: 32, right: 32, top: 12, bottom: 12 })
        .margin({ top: 32 })
        .onClick(() => {
          this.chatViewModel.createNewConversation();
        })
    }
    .width('100%')
    .layoutWeight(1)
    .justifyContent(FlexAlign.Center)
    .alignItems(HorizontalAlign.Center)
    .padding({ left: 32, right: 32 })
  }

  // 构建输入区域
  @Builder
  buildInputArea() {
    Row() {
      // 文本输入框
      TextInput({ 
        placeholder: '输入消息...',
        controller: this.inputController
      })
        .layoutWeight(1)
        .height(40)
        .fontSize(16)
        .backgroundColor(Color.White)
        .borderRadius(20)
        .padding({ left: 16, right: 16 })
        .enabled(this.chatViewModel.isInputEnabled)
        .onChange((value: string) => {
          this.chatViewModel.updateInputText(value);
        })
        .onSubmit(() => {
          if (this.chatViewModel.canSendMessage) {
            this.sendMessage();
          }
        })
      
      // 发送按钮
      Button() {
        if (this.chatViewModel.isSending) {
          LoadingProgress()
            .width(20)
            .height(20)
            .color(Color.White)
        } else {
          Image($r('app.media.ic_send'))
            .width(20)
            .height(20)
            .fillColor(Color.White)
        }
      }
      .width(40)
      .height(40)
      .backgroundColor(this.chatViewModel.canSendMessage ? '#007AFF' : '#CCCCCC')
      .borderRadius(20)
      .margin({ left: 8 })
      .enabled(this.chatViewModel.canSendMessage)
      .onClick(() => {
        this.sendMessage();
      })
    }
    .width('100%')
    .padding({ left: 16, right: 16, top: 12, bottom: 12 })
    .backgroundColor('#F8F8F8')
    .border({ width: { top: 1 }, color: '#E5E5E5' })
  }

  // 构建对话列表弹窗
  @Builder
  buildConversationListDialog() {
    if (this.chatViewModel.isConversationListVisible) {
      Column() {
        // 弹窗标题
        Row() {
          Text('对话列表')
            .fontSize(18)
            .fontWeight(FontWeight.Medium)
            .fontColor('#333333')
            .layoutWeight(1)
          
          Button() {
            Image($r('app.media.ic_close'))
              .width(20)
              .height(20)
              .fillColor('#666666')
          }
          .width(32)
          .height(32)
          .backgroundColor(Color.Transparent)
          .borderRadius(16)
          .onClick(() => {
            this.chatViewModel.hideConversationList();
          })
        }
        .width('100%')
        .padding({ left: 20, right: 20, top: 16, bottom: 16 })
        .border({ width: { bottom: 1 }, color: '#E5E5E5' })
        
        // 对话列表
        List() {
          ForEach(this.chatViewModel.conversations, (conversation: Conversation) => {
            ListItem() {
              this.buildConversationItem(conversation)
            }
          })
        }
        .width('100%')
        .layoutWeight(1)
        .scrollBar(BarState.Off)
      }
      .width('90%')
      .height('70%')
      .backgroundColor(Color.White)
      .borderRadius(12)
      .shadow({
        radius: 12,
        color: '#33000000',
        offsetX: 0,
        offsetY: 4
      })
    }
  }

  // 构建对话项
  @Builder
  buildConversationItem(conversation: Conversation) {
    Row() {
      Column() {
        Text(conversation.title)
          .fontSize(16)
          .fontColor('#333333')
          .maxLines(1)
          .textOverflow({ overflow: TextOverflow.Ellipsis })
        
        Text(`${conversation.messageCount}条消息 · ${this.chatViewModel.formatTime(conversation.updatedAt)}`)
          .fontSize(12)
          .fontColor('#666666')
          .margin({ top: 4 })
      }
      .layoutWeight(1)
      .alignItems(HorizontalAlign.Start)
      
      if (conversation.conversationId === this.chatViewModel.currentConversation?.conversationId) {
        Image($r('app.media.ic_check'))
          .width(16)
          .height(16)
          .fillColor('#007AFF')
      }
    }
    .width('100%')
    .padding({ left: 20, right: 20, top: 12, bottom: 12 })
    .onClick(() => {
      this.chatViewModel.selectConversation(conversation.conversationId);
    })
    .gesture(
      LongPressGesture({ repeat: false })
        .onAction(() => {
          this.showConversationMenu(conversation);
        })
    )
  }

  // 构建AI人格选择弹窗
  @Builder
  buildAIPersonalityDialog() {
    if (this.chatViewModel.isAIPersonalityDialogVisible) {
      Column() {
        // 弹窗标题
        Row() {
          Text('选择AI人格')
            .fontSize(18)
            .fontWeight(FontWeight.Medium)
            .fontColor('#333333')
            .layoutWeight(1)
          
          Button() {
            Image($r('app.media.ic_close'))
              .width(20)
              .height(20)
              .fillColor('#666666')
          }
          .width(32)
          .height(32)
          .backgroundColor(Color.Transparent)
          .borderRadius(16)
          .onClick(() => {
            this.chatViewModel.hideAIPersonalityDialog();
          })
        }
        .width('100%')
        .padding({ left: 20, right: 20, top: 16, bottom: 16 })
        .border({ width: { bottom: 1 }, color: '#E5E5E5' })
        
        // AI人格列表
        Column() {
          this.buildPersonalityOption(AIPersonality.DEFAULT, '默认助手', '通用AI助手，适合各种问题')
          this.buildPersonalityOption(AIPersonality.WISE_SAGE, '智慧长者', '富有智慧的长者，善于人生指导')
          this.buildPersonalityOption(AIPersonality.FRIENDLY_GUIDE, '友善向导', '友好热情的向导，乐于助人')
          this.buildPersonalityOption(AIPersonality.SCHOLARLY, '学者专家', '严谨的学者，专业知识丰富')
          this.buildPersonalityOption(AIPersonality.POETIC, '诗意文人', '富有诗意的文人，语言优美')
        }
        .width('100%')
        .padding({ top: 8, bottom: 8 })
      }
      .width('85%')
      .backgroundColor(Color.White)
      .borderRadius(12)
      .shadow({
        radius: 12,
        color: '#33000000',
        offsetX: 0,
        offsetY: 4
      })
    }
  }

  // 构建人格选项
  @Builder
  buildPersonalityOption(personality: AIPersonality, name: string, description: string) {
    Row() {
      Column() {
        Text(name)
          .fontSize(16)
          .fontColor('#333333')
          .fontWeight(FontWeight.Medium)
        
        Text(description)
          .fontSize(14)
          .fontColor('#666666')
          .margin({ top: 4 })
      }
      .layoutWeight(1)
      .alignItems(HorizontalAlign.Start)
      
      if (personality === this.chatViewModel.selectedAIPersonality) {
        Image($r('app.media.ic_check'))
          .width(20)
          .height(20)
          .fillColor('#007AFF')
      }
    }
    .width('100%')
    .padding({ left: 20, right: 20, top: 16, bottom: 16 })
    .onClick(() => {
      this.chatViewModel.changeAIPersonality(personality);
    })
  }

  // 发送消息
  private sendMessage() {
    this.chatViewModel.sendMessage();
    this.inputController.caretPosition(0);
    
    // 滚动到底部
    Timer.setTimeout(() => {
      this.scrollController.scrollToIndex(this.chatViewModel.displayMessages.length - 1);
    }, 100);
  }

  // 显示消息菜单
  private showMessageMenu(message: ChatMessage) {
    // TODO: 实现消息长按菜单
    console.info('Show message menu for:', message.messageId);
  }

  // 显示对话菜单
  private showConversationMenu(conversation: Conversation) {
    // TODO: 实现对话长按菜单
    console.info('Show conversation menu for:', conversation.conversationId);
  }
}