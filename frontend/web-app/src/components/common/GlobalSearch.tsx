import React, { useEffect, useRef } from 'react';
import { Modal, Input, List, Avatar, Tag, Empty, Spin } from 'antd';
import { SearchOutlined, FileTextOutlined, UserOutlined, BookOutlined, MessageOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useGlobalSearch, useDebounce } from '../../hooks';

const { Search } = Input;

interface SearchResult {
  id: string;
  type: 'wisdom' | 'user' | 'conversation' | 'document';
  title: string;
  description?: string;
  avatar?: string;
  tags?: string[];
  path: string;
  score?: number;
}

const typeIcons = {
  wisdom: <BookOutlined />,
  user: <UserOutlined />,
  conversation: <MessageOutlined />,
  document: <FileTextOutlined />
};

const typeLabels = {
  wisdom: '智慧',
  user: '用户',
  conversation: '对话',
  document: '文档'
};

const typeColors = {
  wisdom: 'blue',
  user: 'green',
  conversation: 'orange',
  document: 'purple'
};

export const GlobalSearch: React.FC = () => {
  const navigate = useNavigate();
  const searchInputRef = useRef<any>(null);
  const { visible, query, results, loading, setVisible, setQuery, search } = useGlobalSearch();
  
  // 防抖搜索
  const debouncedQuery = useDebounce(query, 300);

  // 当防抖查询变化时执行搜索
  useEffect(() => {
    if (debouncedQuery && visible) {
      search(debouncedQuery);
    }
  }, [debouncedQuery, visible, search]);

  // 当模态框打开时聚焦搜索框
  useEffect(() => {
    if (visible && searchInputRef.current) {
      setTimeout(() => {
        searchInputRef.current?.focus();
      }, 100);
    }
  }, [visible]);

  // 处理搜索结果点击
  const handleResultClick = (result: SearchResult) => {
    setVisible(false);
    navigate(result.path);
  };

  // 处理键盘事件
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      setVisible(false);
    }
  };

  // 渲染搜索结果项
  const renderResultItem = (item: SearchResult) => (
    <List.Item
      key={item.id}
      className="cursor-pointer hover:bg-gray-50 transition-colors px-4 py-3"
      onClick={() => handleResultClick(item)}
    >
      <List.Item.Meta
        avatar={
          item.avatar ? (
            <Avatar src={item.avatar} />
          ) : (
            <Avatar icon={typeIcons[item.type]} style={{ backgroundColor: '#f0f0f0', color: '#666' }} />
          )
        }
        title={
          <div className="flex items-center gap-2">
            <span className="font-medium">{item.title}</span>
            <Tag color={typeColors[item.type]} size="small">
              {typeLabels[item.type]}
            </Tag>
            {item.score && (
              <span className="text-xs text-gray-400">
                匹配度: {Math.round(item.score * 100)}%
              </span>
            )}
          </div>
        }
        description={
          <div>
            {item.description && (
              <p className="text-gray-600 text-sm mb-1 line-clamp-2">{item.description}</p>
            )}
            {item.tags && item.tags.length > 0 && (
              <div className="flex flex-wrap gap-1">
                {item.tags.slice(0, 3).map((tag, index) => (
                  <Tag key={index} size="small" color="default">
                    {tag}
                  </Tag>
                ))}
                {item.tags.length > 3 && (
                  <Tag size="small" color="default">
                    +{item.tags.length - 3}
                  </Tag>
                )}
              </div>
            )}
          </div>
        }
      />
    </List.Item>
  );

  // 按类型分组结果
  const groupedResults = results.reduce((acc, item) => {
    if (!acc[item.type]) {
      acc[item.type] = [];
    }
    acc[item.type].push(item);
    return acc;
  }, {} as Record<string, SearchResult[]>);

  return (
    <Modal
      title={null}
      open={visible}
      onCancel={() => setVisible(false)}
      footer={null}
      width={600}
      centered
      className="global-search-modal"
      styles={{
        body: { padding: 0 },
        mask: { backgroundColor: 'rgba(0, 0, 0, 0.3)' }
      }}
      onKeyDown={handleKeyDown}
    >
      <div className="p-4 border-b border-gray-200">
        <Search
          ref={searchInputRef}
          placeholder="搜索智慧、用户、对话、文档..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          size="large"
          prefix={<SearchOutlined className="text-gray-400" />}
          allowClear
          className="global-search-input"
        />
      </div>

      <div className="max-h-96 overflow-y-auto">
        {loading ? (
          <div className="flex justify-center items-center py-8">
            <Spin size="large" />
          </div>
        ) : query && results.length === 0 ? (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description="未找到相关结果"
            className="py-8"
          />
        ) : query ? (
          <div>
            {Object.entries(groupedResults).map(([type, items]) => (
              <div key={type}>
                <div className="px-4 py-2 bg-gray-50 border-b border-gray-100">
                  <span className="text-sm font-medium text-gray-700 flex items-center gap-2">
                    {typeIcons[type as keyof typeof typeIcons]}
                    {typeLabels[type as keyof typeof typeLabels]} ({items.length})
                  </span>
                </div>
                <List
                  dataSource={items}
                  renderItem={renderResultItem}
                  split={false}
                />
              </div>
            ))}
          </div>
        ) : (
          <div className="p-8 text-center text-gray-500">
            <SearchOutlined className="text-4xl mb-4 text-gray-300" />
            <p>输入关键词开始搜索</p>
            <div className="mt-4 text-sm">
              <p>支持搜索：</p>
              <div className="flex justify-center gap-2 mt-2">
                <Tag color="blue">智慧内容</Tag>
                <Tag color="green">用户</Tag>
                <Tag color="orange">对话记录</Tag>
                <Tag color="purple">文档</Tag>
              </div>
            </div>
          </div>
        )}
      </div>

      {query && results.length > 0 && (
        <div className="p-3 border-t border-gray-200 bg-gray-50 text-center text-xs text-gray-500">
          找到 {results.length} 个结果 • 按 ESC 关闭 • 点击结果跳转
        </div>
      )}
    </Modal>
  );
};

export default GlobalSearch;