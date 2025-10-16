import React, { useState, useRef, useEffect } from 'react';
import { 
  Modal, 
  Input, 
  Button, 
  List, 
  Avatar, 
  Space, 
  Typography, 
  Divider,
  message,
  Card,
  Tag,
  Tooltip
} from 'antd';
import {
  CustomerServiceOutlined,
  SendOutlined,
  CloseOutlined,
  SmileOutlined,
  PaperClipOutlined,
  PhoneOutlined,
  MailOutlined
} from '@ant-design/icons';

const { Text, Title } = Typography;
const { TextArea } = Input;

interface Message {
  id: string;
  content: string;
  sender: 'user' | 'service';
  timestamp: Date;
  type: 'text' | 'image' | 'file';
}

interface CustomerServiceProps {
  visible: boolean;
  onClose: () => void;
}

const CustomerService: React.FC<CustomerServiceProps> = ({ visible, onClose }) => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputValue, setInputValue] = useState('');
  const [isConnected, setIsConnected] = useState(false);
  const [isTyping, setIsTyping] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // 初始化消息
  useEffect(() => {
    if (visible && messages.length === 0) {
      const welcomeMessage: Message = {
        id: '1',
        content: '您好！欢迎使用太上老君AI平台客服服务。我是您的专属客服小助手，有什么可以帮助您的吗？',
        sender: 'service',
        timestamp: new Date(),
        type: 'text'
      };
      setMessages([welcomeMessage]);
      setIsConnected(true);
    }
  }, [visible, messages.length]);

  // 自动滚动到底部
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  // 发送消息
  const sendMessage = () => {
    if (!inputValue.trim()) return;

    const userMessage: Message = {
      id: Date.now().toString(),
      content: inputValue,
      sender: 'user',
      timestamp: new Date(),
      type: 'text'
    };

    setMessages(prev => [...prev, userMessage]);
    setInputValue('');
    setIsTyping(true);

    // 模拟客服回复
    setTimeout(() => {
      const serviceMessage: Message = {
        id: (Date.now() + 1).toString(),
        content: getAutoReply(inputValue),
        sender: 'service',
        timestamp: new Date(),
        type: 'text'
      };
      setMessages(prev => [...prev, serviceMessage]);
      setIsTyping(false);
    }, 1000 + Math.random() * 2000);
  };

  // 自动回复逻辑
  const getAutoReply = (userInput: string): string => {
    const input = userInput.toLowerCase();
    
    if (input.includes('登录') || input.includes('密码')) {
      return '关于登录问题，请确认您的用户名和密码是否正确。如果忘记密码，可以点击登录页面的"忘记密码"进行重置。如需进一步帮助，请提供您的注册邮箱。';
    }
    
    if (input.includes('ai') || input.includes('智能')) {
      return 'AI功能使用很简单！您可以在左侧菜单找到"AI智能服务"，包含智能对话、多模态AI等功能。有具体问题可以详细描述，我来为您解答。';
    }
    
    if (input.includes('学习') || input.includes('课程')) {
      return '智能学习模块提供个性化学习体验。您可以创建学习计划、跟踪进度、参与课程等。建议从"智能学习"菜单开始探索。';
    }
    
    if (input.includes('收费') || input.includes('价格') || input.includes('费用')) {
      return '我们提供多种服务套餐，包括免费版和高级版。具体价格信息请查看"设置"页面的"订阅管理"，或联系销售团队获取详细报价。';
    }
    
    if (input.includes('联系') || input.includes('电话')) {
      return '您可以通过以下方式联系我们：\n📞 客服热线：400-123-4567\n📧 邮箱：support@taishanglaojun.com\n💬 在线客服：就是现在这个窗口\n⏰ 服务时间：7×24小时';
    }
    
    return '感谢您的咨询！我已经记录了您的问题，专业客服人员会尽快为您提供详细解答。如果是紧急问题，建议您拨打客服热线 400-123-4567。';
  };

  // 快捷回复选项
  const quickReplies = [
    '如何开始使用？',
    'AI功能介绍',
    '忘记密码了',
    '联系人工客服',
    '价格咨询'
  ];

  const handleQuickReply = (reply: string) => {
    setInputValue(reply);
  };

  const formatTime = (date: Date) => {
    return date.toLocaleTimeString('zh-CN', { 
      hour: '2-digit', 
      minute: '2-digit' 
    });
  };

  return (
    <Modal
      title={
        <div className="flex items-center justify-between">
          <Space>
            <CustomerServiceOutlined className="text-blue-500" />
            <span>在线客服</span>
            {isConnected && (
              <Tag color="green" size="small">在线</Tag>
            )}
          </Space>
          <Button 
            type="text" 
            icon={<CloseOutlined />} 
            onClick={onClose}
            size="small"
          />
        </div>
      }
      open={visible}
      onCancel={onClose}
      footer={null}
      width={480}
      style={{ top: 20 }}
      styles={{ body: { padding: 0, height: '600px', display: 'flex', flexDirection: 'column' } }}
      closable={false}
    >
      {/* 消息列表 */}
      <div className="flex-1 p-4 overflow-y-auto bg-gray-50">
        <List
          dataSource={messages}
          renderItem={(message) => (
            <div
              className={`mb-4 flex ${
                message.sender === 'user' ? 'justify-end' : 'justify-start'
              }`}
            >
              <div
                className={`max-w-xs px-3 py-2 rounded-lg ${
                  message.sender === 'user'
                    ? 'bg-blue-500 text-white'
                    : 'bg-white border shadow-sm'
                }`}
              >
                <div className="whitespace-pre-wrap">{message.content}</div>
                <div
                  className={`text-xs mt-1 ${
                    message.sender === 'user' ? 'text-blue-100' : 'text-gray-500'
                  }`}
                >
                  {formatTime(message.timestamp)}
                </div>
              </div>
            </div>
          )}
        />
        
        {/* 正在输入提示 */}
        {isTyping && (
          <div className="flex justify-start mb-4">
            <div className="bg-white border shadow-sm px-3 py-2 rounded-lg">
              <div className="flex items-center space-x-1">
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"></div>
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.1s' }}></div>
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.2s' }}></div>
              </div>
            </div>
          </div>
        )}
        
        <div ref={messagesEndRef} />
      </div>

      {/* 快捷回复 */}
      <div className="px-4 py-2 border-t bg-white">
        <div className="mb-2">
          <Text type="secondary" className="text-xs">快捷回复：</Text>
        </div>
        <Space wrap size="small">
          {quickReplies.map((reply, index) => (
            <Button
              key={index}
              size="small"
              type="default"
              onClick={() => handleQuickReply(reply)}
              className="text-xs"
            >
              {reply}
            </Button>
          ))}
        </Space>
      </div>

      {/* 输入区域 */}
      <div className="p-4 border-t bg-white">
        <div className="flex space-x-2">
          <TextArea
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            placeholder="请输入您的问题..."
            autoSize={{ minRows: 1, maxRows: 3 }}
            onPressEnter={(e) => {
              if (!e.shiftKey) {
                e.preventDefault();
                sendMessage();
              }
            }}
            className="flex-1"
          />
          <Button
            type="primary"
            icon={<SendOutlined />}
            onClick={sendMessage}
            disabled={!inputValue.trim()}
          >
            发送
          </Button>
        </div>
        
        <div className="flex justify-between items-center mt-2">
          <Space size="small">
            <Tooltip title="表情">
              <Button type="text" icon={<SmileOutlined />} size="small" />
            </Tooltip>
            <Tooltip title="附件">
              <Button type="text" icon={<PaperClipOutlined />} size="small" />
            </Tooltip>
          </Space>
          
          <Space size="small">
            <Tooltip title="电话客服">
              <Button 
                type="text" 
                icon={<PhoneOutlined />} 
                size="small"
                onClick={() => message.info('客服热线：400-123-4567')}
              />
            </Tooltip>
            <Tooltip title="邮件客服">
              <Button 
                type="text" 
                icon={<MailOutlined />} 
                size="small"
                onClick={() => message.info('客服邮箱：support@taishanglaojun.com')}
              />
            </Tooltip>
          </Space>
        </div>
      </div>
    </Modal>
  );
};

export default CustomerService;