/**
 * 对话列表组件
 * 显示历史对话列表，支持搜索、删除、归档等操作
 */

import React, { useState, useMemo } from 'react';
import {
  List,
  Card,
  Input,
  Button,
  Typography,
  Space,
  Dropdown,
  Modal,
  message,
  Empty,
  Badge,
  Tooltip,
  Tag
} from 'antd';
import type { MenuProps } from 'antd';
import {
  SearchOutlined,
  MessageOutlined,
  DeleteOutlined,
  MoreOutlined,
  PlusOutlined,
  ExportOutlined,
  ImportOutlined,
  ClockCircleOutlined
} from '@ant-design/icons';
import type { ConversationSummary } from '../../types';

const { Search } = Input;
const { Text, Paragraph } = Typography;
const { confirm } = Modal;

export interface ConversationListProps {
  conversations: ConversationSummary[];
  currentConversationId?: string;
  loading?: boolean;
  onSelectConversation: (conversationId: string) => void;
  onCreateConversation: () => void;
  onDeleteConversation: (conversationId: string) => void;
  onArchiveConversation: (conversationId: string) => void;
  onSearchConversations: (query: string) => ConversationSummary[];
  onExportConversations: () => string;
  onImportConversations: (data: string) => boolean;
}

const ConversationList: React.FC<ConversationListProps> = ({
  conversations,
  currentConversationId,
  loading = false,
  onSelectConversation,
  onCreateConversation,
  onDeleteConversation,
  onArchiveConversation,
  onSearchConversations,
  onExportConversations,
  onImportConversations,
}) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [showImportModal, setShowImportModal] = useState(false);
  const [importData, setImportData] = useState('');

  // 过滤后的对话列表
  const filteredConversations = useMemo(() => {
    if (!searchQuery.trim()) {
      return conversations;
    }
    return onSearchConversations(searchQuery.trim());
  }, [conversations, searchQuery, onSearchConversations]);

  /**
   * 处理删除对话
   */
  const handleDeleteConversation = (conversationId: string, title: string) => {
    confirm({
      title: '确认删除对话',
      content: `确定要删除对话"${title}"吗？此操作不可撤销。`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk() {
        onDeleteConversation(conversationId);
        message.success('对话已删除');
      },
    });
  };

  /**
   * 处理归档对话
   */
  const handleArchiveConversation = (conversationId: string, title: string) => {
    confirm({
      title: '确认归档对话',
      content: `确定要归档对话"${title}"吗？归档后的对话将不会在列表中显示。`,
      okText: '归档',
      cancelText: '取消',
      onOk() {
        onArchiveConversation(conversationId);
        message.success('对话已归档');
      },
    });
  };

  /**
   * 处理导出对话
   */
  const handleExportConversations = () => {
    try {
      const data = onExportConversations();
      const blob = new Blob([data], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `conversations_${new Date().toISOString().split('T')[0]}.json`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
      message.success('对话数据已导出');
    } catch (error) {
      message.error('导出失败');
    }
  };

  /**
   * 处理导入对话
   */
  const handleImportConversations = () => {
    if (!importData.trim()) {
      message.error('请输入要导入的数据');
      return;
    }

    const success = onImportConversations(importData.trim());
    if (success) {
      setShowImportModal(false);
      setImportData('');
    }
  };

  /**
   * 处理文件导入
   */
  const handleFileImport = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (e) => {
      const content = e.target?.result as string;
      setImportData(content);
    };
    reader.readAsText(file);
  };

  /**
   * 获取对话操作菜单
   */
  const getConversationMenu = (conversation: ConversationSummary): MenuProps => ({
    items: [
      {
        key: 'archive',
        label: '归档对话',
        icon: <InboxOutlined />,
        onClick: () => handleArchiveConversation(conversation.id, conversation.title),
      },
      {
        key: 'delete',
        label: '删除对话',
        icon: <DeleteOutlined />,
        danger: true,
        onClick: () => handleDeleteConversation(conversation.id, conversation.title),
      },
    ],
  });

  /**
   * 格式化时间显示
   */
  const formatTime = (timestamp: string): string => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

    if (diffDays === 0) {
      return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
    } else if (diffDays === 1) {
      return '昨天';
    } else if (diffDays < 7) {
      return `${diffDays}天前`;
    } else {
      return date.toLocaleDateString('zh-CN');
    }
  };

  return (
    <Card
      title="对话历史"
      size="small"
      extra={
        <Space>
          <Tooltip title="导出对话">
            <Button
              type="text"
              size="small"
              icon={<ExportOutlined />}
              onClick={handleExportConversations}
            />
          </Tooltip>
          <Tooltip title="导入对话">
            <Button
              type="text"
              size="small"
              icon={<ImportOutlined />}
              onClick={() => setShowImportModal(true)}
            />
          </Tooltip>
          <Button
            type="primary"
            size="small"
            icon={<PlusOutlined />}
            onClick={onCreateConversation}
          >
            新对话
          </Button>
        </Space>
      }
      styles={{ body: { padding: '8px' } }}
    >
      {/* 搜索框 */}
      <Search
        placeholder="搜索对话..."
        allowClear
        size="small"
        value={searchQuery}
        onChange={(e) => setSearchQuery(e.target.value)}
        style={{ marginBottom: 8 }}
        prefix={<SearchOutlined />}
      />

      {/* 对话列表 */}
      <List
        size="small"
        loading={loading}
        dataSource={filteredConversations}
        locale={{
          emptyText: (
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description="暂无对话记录"
              style={{ margin: '20px 0' }}
            >
              <Button type="primary" onClick={onCreateConversation}>
                开始新对话
              </Button>
            </Empty>
          ),
        }}
        renderItem={(conversation) => (
          <List.Item
            key={conversation.id}
            style={{
              padding: '8px 12px',
              cursor: 'pointer',
              backgroundColor: conversation.id === currentConversationId ? '#f0f8ff' : 'transparent',
              borderRadius: '4px',
              marginBottom: '4px',
              border: conversation.id === currentConversationId ? '1px solid #1890ff' : '1px solid transparent',
            }}
            onClick={() => onSelectConversation(conversation.id)}
            actions={[
              <Dropdown
                key="more"
                menu={getConversationMenu(conversation)}
                trigger={['click']}
                placement="bottomRight"
              >
                <Button
                  type="text"
                  size="small"
                  icon={<MoreOutlined />}
                  onClick={(e) => e.stopPropagation()}
                />
              </Dropdown>,
            ]}
          >
            <List.Item.Meta
              avatar={
                <Badge count={conversation.messageCount} size="small" showZero={false}>
                  <MessageOutlined style={{ fontSize: '16px', color: '#1890ff' }} />
                </Badge>
              }
              title={
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <Text
                    ellipsis={{ tooltip: conversation.title }}
                    style={{
                      fontWeight: conversation.id === currentConversationId ? 'bold' : 'normal',
                      maxWidth: '150px',
                    }}
                  >
                    {conversation.title}
                  </Text>
                  <Space size={4}>
                    <Tag color="blue" size="small">
                      {conversation.messageCount}
                    </Tag>
                    <Text type="secondary" style={{ fontSize: '11px' }}>
                      <ClockCircleOutlined style={{ marginRight: 2 }} />
                      {formatTime(conversation.updatedAt)}
                    </Text>
                  </Space>
                </div>
              }
              description={
                <Paragraph
                  ellipsis={{ rows: 2, tooltip: conversation.lastMessage }}
                  style={{ margin: 0, fontSize: '12px', color: '#666' }}
                >
                  {conversation.lastMessage}
                </Paragraph>
              }
            />
          </List.Item>
        )}
      />

      {/* 导入对话模态框 */}
      <Modal
        title="导入对话数据"
        open={showImportModal}
        onOk={handleImportConversations}
        onCancel={() => {
          setShowImportModal(false);
          setImportData('');
        }}
        okText="导入"
        cancelText="取消"
        width={600}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <div>
            <Text>选择文件导入：</Text>
            <input
              type="file"
              accept=".json"
              onChange={handleFileImport}
              style={{ marginLeft: 8 }}
            />
          </div>
          <div>
            <Text>或直接粘贴JSON数据：</Text>
            <Input.TextArea
              value={importData}
              onChange={(e) => setImportData(e.target.value)}
              placeholder="请粘贴对话数据的JSON格式..."
              rows={10}
              style={{ marginTop: 8 }}
            />
          </div>
        </Space>
      </Modal>
    </Card>
  );
};

export default ConversationList;