import React, { useState, useEffect } from 'react';
import { 
  Card, 
  List, 
  Typography, 
  Tag, 
  Space, 
  Button, 
  message,
  Spin,
  Empty,
  Pagination,
  Input,
  Row,
  Col,
  Select,
  Modal
} from 'antd';
import { 
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  BookOutlined,
  CalendarOutlined,
  SearchOutlined,
  TagOutlined,
  LockOutlined,
  UnlockOutlined
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { apiClient } from '../services/api';
import NoteModal from '../components/notes/NoteModal';

const { Title, Paragraph } = Typography;
const { Search } = Input;
const { Option } = Select;

interface NoteItem {
  id: string;
  wisdom_id: string;
  wisdom_title: string;
  title: string;
  content: string;
  is_private: boolean;
  tags: string[];
  created_at: string;
  updated_at: string;
}

const UserNotes: React.FC = () => {
  const navigate = useNavigate();
  const [notes, setNotes] = useState<NoteItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(10);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [allTags, setAllTags] = useState<string[]>([]);
  const [noteModalVisible, setNoteModalVisible] = useState(false);
  const [editingNote, setEditingNote] = useState<NoteItem | null>(null);

  // 加载笔记列表
  const loadNotes = async (page = 1, search = '', tags: string[] = []) => {
    setLoading(true);
    try {
      const response = await apiClient.getUserNotes({
        page,
        limit: pageSize,
        search,
        tags: tags.length > 0 ? tags : undefined,
      });

      if (response.success && response.data) {
        setNotes(response.data.notes);
        setTotal(response.data.total);
        
        // 提取所有标签
        const tagsSet = new Set<string>();
        response.data.notes.forEach(note => {
          note.tags.forEach(tag => tagsSet.add(tag));
        });
        setAllTags(Array.from(tagsSet));
      } else {
        message.error('获取笔记列表失败');
      }
    } catch (error) {
      console.error('加载笔记列表失败:', error);
      message.error('网络错误，请稍后重试');
    }
    setLoading(false);
  };

  // 删除笔记
  const handleDeleteNote = async (wisdomId: string, noteTitle: string) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除笔记"${noteTitle}"吗？此操作不可恢复。`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await apiClient.deleteNote(wisdomId);
          message.success('笔记删除成功');
          loadNotes(currentPage, searchKeyword, selectedTags);
        } catch (error) {
          message.error('删除笔记失败');
        }
      },
    });
  };

  // 编辑笔记
  const handleEditNote = (note: NoteItem) => {
    setEditingNote(note);
    setNoteModalVisible(true);
  };

  // 笔记操作成功回调
  const handleNoteSuccess = () => {
    loadNotes(currentPage, searchKeyword, selectedTags);
    setEditingNote(null);
  };

  // 搜索处理
  const handleSearch = (value: string) => {
    setSearchKeyword(value);
    setCurrentPage(1);
    loadNotes(1, value, selectedTags);
  };

  // 标签筛选
  const handleTagsChange = (tags: string[]) => {
    setSelectedTags(tags);
    setCurrentPage(1);
    loadNotes(1, searchKeyword, tags);
  };

  // 页码变化
  const handlePageChange = (page: number) => {
    setCurrentPage(page);
    loadNotes(page, searchKeyword, selectedTags);
  };

  useEffect(() => {
    loadNotes();
  }, []);

  return (
    <div className="max-w-6xl mx-auto p-6">
      {/* 页面标题 */}
      <div className="mb-6">
        <Title level={2} className="mb-2">
          <EditOutlined className="text-blue-500 mr-2" />
          我的笔记
        </Title>
        <Paragraph className="text-gray-600">
          管理您的学习笔记，记录思考与感悟
        </Paragraph>
      </div>

      {/* 搜索和筛选栏 */}
      <Card className="mb-6">
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} sm={12} md={14}>
            <Search
              placeholder="搜索笔记标题或内容..."
              allowClear
              enterButton={<SearchOutlined />}
              size="large"
              onSearch={handleSearch}
              onChange={(e) => {
                if (!e.target.value) {
                  handleSearch('');
                }
              }}
            />
          </Col>
          <Col xs={24} sm={8} md={6}>
            <Select
              mode="multiple"
              placeholder="按标签筛选"
              style={{ width: '100%' }}
              value={selectedTags}
              onChange={handleTagsChange}
              allowClear
            >
              {allTags.map(tag => (
                <Option key={tag} value={tag}>
                  <TagOutlined className="mr-1" />
                  {tag}
                </Option>
              ))}
            </Select>
          </Col>
          <Col xs={24} sm={4} md={4}>
            <div className="text-right">
              <span className="text-gray-500">
                共 {total} 条笔记
              </span>
            </div>
          </Col>
        </Row>
      </Card>

      {/* 笔记列表 */}
      <Card>
        <Spin spinning={loading}>
          {notes.length === 0 ? (
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description={
                searchKeyword || selectedTags.length > 0 
                  ? '没有找到匹配的笔记' 
                  : '您还没有创建任何笔记'
              }
            >
              {!searchKeyword && selectedTags.length === 0 && (
                <Button type="primary" onClick={() => navigate('/wisdom')}>
                  去创建笔记
                </Button>
              )}
            </Empty>
          ) : (
            <>
              <List
                itemLayout="vertical"
                dataSource={notes}
                renderItem={(item) => (
                  <List.Item
                    key={item.id}
                    className="hover:bg-gray-50 transition-colors duration-200 rounded-lg p-4"
                    actions={[
                      <Button
                        key="view"
                        type="link"
                        icon={<EyeOutlined />}
                        onClick={() => navigate(`/wisdom/${item.wisdom_id}`)}
                      >
                        查看原文
                      </Button>,
                      <Button
                        key="edit"
                        type="link"
                        icon={<EditOutlined />}
                        onClick={() => handleEditNote(item)}
                      >
                        编辑
                      </Button>,
                      <Button
                        key="delete"
                        type="link"
                        danger
                        icon={<DeleteOutlined />}
                        onClick={() => handleDeleteNote(item.wisdom_id, item.title)}
                      >
                        删除
                      </Button>,
                    ]}
                  >
                    <List.Item.Meta
                      title={
                        <div className="flex items-center justify-between">
                          <div className="flex items-center space-x-2">
                            <span className="text-lg font-semibold text-gray-800">
                              {item.title}
                            </span>
                            {item.is_private ? (
                              <LockOutlined className="text-gray-400" title="私密笔记" />
                            ) : (
                              <UnlockOutlined className="text-green-500" title="公开笔记" />
                            )}
                          </div>
                          <div className="flex items-center space-x-2">
                            {item.tags.map(tag => (
                              <Tag key={tag} color="blue" className="text-xs">
                                {tag}
                              </Tag>
                            ))}
                          </div>
                        </div>
                      }
                      description={
                        <div className="space-y-3">
                          <div className="text-sm text-gray-600">
                            <BookOutlined className="mr-1" />
                            关于：
                            <span 
                              className="text-blue-600 hover:text-blue-800 cursor-pointer ml-1"
                              onClick={() => navigate(`/wisdom/${item.wisdom_id}`)}
                            >
                              {item.wisdom_title}
                            </span>
                          </div>
                          
                          <Paragraph
                            className="text-gray-700 bg-gray-50 p-3 rounded-md"
                            ellipsis={{ rows: 3, expandable: true, symbol: '展开' }}
                          >
                            {item.content}
                          </Paragraph>
                          
                          <div className="flex items-center justify-between text-sm text-gray-500">
                            <Space>
                              <span>
                                <CalendarOutlined className="mr-1" />
                                创建于 {new Date(item.created_at).toLocaleDateString()}
                              </span>
                              {item.updated_at !== item.created_at && (
                                <span>
                                  更新于 {new Date(item.updated_at).toLocaleDateString()}
                                </span>
                              )}
                            </Space>
                          </div>
                        </div>
                      }
                    />
                  </List.Item>
                )}
              />

              {/* 分页 */}
              {total > pageSize && (
                <div className="mt-6 text-center">
                  <Pagination
                    current={currentPage}
                    total={total}
                    pageSize={pageSize}
                    onChange={handlePageChange}
                    showSizeChanger={false}
                    showQuickJumper
                    showTotal={(total, range) =>
                      `第 ${range[0]}-${range[1]} 条，共 ${total} 条`
                    }
                  />
                </div>
              )}
            </>
          )}
        </Spin>
      </Card>

      {/* 笔记编辑模态框 */}
      {editingNote && (
        <NoteModal
          visible={noteModalVisible}
          onCancel={() => {
            setNoteModalVisible(false);
            setEditingNote(null);
          }}
          onSuccess={handleNoteSuccess}
          wisdomId={editingNote.wisdom_id}
          wisdomTitle={editingNote.wisdom_title}
          existingNote={{
            id: editingNote.id,
            title: editingNote.title,
            content: editingNote.content,
            is_private: editingNote.is_private,
            tags: editingNote.tags,
          }}
        />
      )}
    </div>
  );
};

export default UserNotes;