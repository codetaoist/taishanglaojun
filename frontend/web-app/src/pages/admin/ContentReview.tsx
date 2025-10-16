import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  Select,
  message,
  Tooltip,
  Typography,
  Row,
  Col,
  Statistic,
  Tabs
} from 'antd';
import type { ColumnType } from 'antd/es/table';
import {
  EyeOutlined,
  CheckOutlined,
  CloseOutlined,
  HistoryOutlined,
  FilterOutlined,
  ExportOutlined,
  UserOutlined,
  FileTextOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined
} from '@ant-design/icons';
import apiClient from '../../services/api';
import dayjs from 'dayjs';

const { TextArea } = Input;
const { Option } = Select;

const { Title, Text, Paragraph } = Typography;

interface ReviewItem {
  id: string;
  title: string;
  content: string;
  type: 'article' | 'comment' | 'image' | 'video';
  author: {
    id: string;
    name: string;
    avatar?: string;
  };
  status: 'pending' | 'approved' | 'rejected';
  priority: 'low' | 'medium' | 'high';
  submitTime: string;
  reviewTime?: string;
  reviewer?: {
    id: string;
    name: string;
  };
  reason?: string;
  tags: string[];
  reportCount: number;
  riskLevel: number;
}

interface ReviewStats {
  total: number;
  pending: number;
  approved: number;
  rejected: number;
  todayReviewed: number;
  avgReviewTime: number;
}

const ContentReview: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [reviewItems, setReviewItems] = useState<ReviewItem[]>([]);
  const [stats, setStats] = useState<ReviewStats>({
    total: 0,
    pending: 0,
    approved: 0,
    rejected: 0,
    todayReviewed: 0,
    avgReviewTime: 0
  });
  const [selectedRowKeys, setSelectedRowKeys] = useState<string[]>([]);
  const [reviewModalVisible, setReviewModalVisible] = useState(false);
  const [detailDrawerVisible, setDetailDrawerVisible] = useState(false);
  const [historyModalVisible, setHistoryModalVisible] = useState(false);
  const [currentItem, setCurrentItem] = useState<ReviewItem | null>(null);
  const [reviewHistory, setReviewHistory] = useState<{
    id: string;
    action: string;
    reason?: string;
    reviewer: string;
    reviewTime: string;
  }[]>([]);
  const [activeTab, setActiveTab] = useState('pending');
  const [form] = Form.useForm();

  useEffect(() => {
    fetchReviewItems();
    fetchStats();
  }, [activeTab]);

  const fetchReviewItems = async () => {
    setLoading(true);
    try {
      const response = await apiClient.getReviewItems({ status: activeTab });
      if (response.success) {
        setReviewItems(response.data.items || []);
      }
    } catch {
      message.error('加载内容失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    try {
      const response = await apiClient.getReviewStats();
      if (response.success) {
        setStats(response.data);
      }
    } catch (error) {
      console.error('获取统计数据失败:', error);
    }
  };

  const handleReview = async (values: {
    action: 'approve' | 'reject';
    reason?: string;
    tags?: string[];
  }) => {
    if (!currentItem) return;

    try {
      const response = await apiClient.reviewContent(currentItem.id, {
        action: values.action,
        reason: values.reason,
        tags: values.tags
      });

      if (response.success) {
        message.success('审核完成');
        setReviewModalVisible(false);
        fetchReviewItems();
        fetchStats();
        form.resetFields();
      }
    } catch {
      message.error('审核失败');
    }
  };

  const handleBatchReview = async (action: 'approve' | 'reject') => {
    if (selectedRowKeys.length === 0) {
      message.warning('请选择要审核的内容');
      return;
    }

    try {
      const response = await apiClient.batchReviewContent({
        ids: selectedRowKeys,
        action,
        reason: action === 'reject' ? '批量审核' : undefined
      });

      if (response.success) {
        message.success(`批量${action === 'approve' ? '通过' : '拒绝'}成功`);
        setSelectedRowKeys([]);
        fetchReviewItems();
        fetchStats();
      }
    } catch {
      message.error('批量审核失败');
    }
  };

  const showReviewModal = (item: ReviewItem) => {
    setCurrentItem(item);
    setReviewModalVisible(true);
  };

  const showDetailDrawer = (item: ReviewItem) => {
    setCurrentItem(item);
    setDetailDrawerVisible(true);
  };

  const showHistory = async (item: ReviewItem) => {
    setCurrentItem(item);
    try {
      const response = await apiClient.getReviewHistory(item.id);
      if (response.success) {
        setReviewHistory(response.data);
        setHistoryModalVisible(true);
      }
    } catch {
      message.error('获取审核历史失败');
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending': return 'orange';
      case 'approved': return 'green';
      case 'rejected': return 'red';
      default: return 'default';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'pending': return '待审核';
      case 'approved': return '已通过';
      case 'rejected': return '已拒绝';
      default: return status;
    }
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high': return 'red';
      case 'medium': return 'orange';
      case 'low': return 'green';
      default: return 'default';
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'article': return <FileTextOutlined />;
      case 'comment': return <UserOutlined />;
      case 'image': return <EyeOutlined />;
      case 'video': return <EyeOutlined />;
      default: return <FileTextOutlined />;
    }
  };

  const columns: ColumnType<ReviewItem>[] = [
    {
      title: '内容',
      dataIndex: 'title',
      key: 'title',
      render: (text, record) => (
        <div>
          <Space>
            {getTypeIcon(record.type)}
            <Text strong>{text}</Text>
            {record.reportCount > 0 && (
              <Badge count={record.reportCount} size="small" />
            )}
          </Space>
          <div style={{ marginTop: 4 }}>
            <Text type="secondary" ellipsis>
              {record.content.substring(0, 100)}...
            </Text>
          </div>
        </div>
      ),
    },
    {
      title: '作者',
      dataIndex: 'author',
      key: 'author',
      width: 120,
      render: (author) => (
        <Space>
          <Avatar src={author.avatar} icon={<UserOutlined />} size="small" />
          <Text>{author.name}</Text>
        </Space>
      ),
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      width: 80,
      render: (type) => {
        const typeMap = {
          article: '文章',
          comment: '评论',
          image: '图片',
          video: '视频'
        };
        return <Tag>{typeMap[type as keyof typeof typeMap]}</Tag>;
      },
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      width: 80,
      render: (priority) => (
        <Tag color={getPriorityColor(priority)}>
          {priority === 'high' ? '高' : priority === 'medium' ? '中' : '低'}
        </Tag>
      ),
    },
    {
      title: '风险等级',
      dataIndex: 'riskLevel',
      key: 'riskLevel',
      width: 100,
      render: (level) => (
        <Rate disabled value={level} count={5} />
      ),
    },
    {
      title: '提交时间',
      dataIndex: 'submitTime',
      key: 'submitTime',
      width: 120,
      render: (time) => dayjs(time).format('MM-DD HH:mm'),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status) => (
        <Tag color={getStatusColor(status)}>
          {getStatusText(status)}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_, record) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => showDetailDrawer(record)}
            />
          </Tooltip>
          {record.status === 'pending' && (
            <>
              <Tooltip title="审核">
                <Button
                  type="text"
                  icon={<CheckOutlined />}
                  onClick={() => showReviewModal(record)}
                />
              </Tooltip>
            </>
          )}
          <Tooltip title="审核历史">
            <Button
              type="text"
              icon={<HistoryOutlined />}
              onClick={() => showHistory(record)}
            />
          </Tooltip>
        </Space>
      ),
    },
  ];

  const rowSelection = {
    selectedRowKeys,
    onChange: (keys: React.Key[]) => {
      setSelectedRowKeys(keys as string[]);
    },
    getCheckboxProps: (record: ReviewItem) => ({
      disabled: record.status !== 'pending',
    }),
  };

  return (
    <div style={{ padding: '24px' }}>
      <Title level={2}>内容审核</Title>

      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={4}>
          <Card>
            <Statistic
              title="总数"
              value={stats.total}
              prefix={<FileTextOutlined />}
            />
          </Card>
        </Col>
        <Col span={4}>
          <Card>
            <Statistic
              title="待审核"
              value={stats.pending}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col span={4}>
          <Card>
            <Statistic
              title="已通过"
              value={stats.approved}
              prefix={<CheckOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={4}>
          <Card>
            <Statistic
              title="已拒绝"
              value={stats.rejected}
              prefix={<CloseOutlined />}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
        <Col span={4}>
          <Card>
            <Statistic
              title="今日审核"
              value={stats.todayReviewed}
              prefix={<ExclamationCircleOutlined />}
            />
          </Card>
        </Col>
        <Col span={4}>
          <Card>
            <Statistic
              title="平均用时"
              value={stats.avgReviewTime}
              suffix="分钟"
              prefix={<ClockCircleOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Card>
        <Tabs 
          activeKey={activeTab} 
          onChange={setActiveTab}
          items={[
            { key: 'pending', label: '待审核' },
            { key: 'approved', label: '已通过' },
            { key: 'rejected', label: '已拒绝' },
            { key: 'all', label: '全部' }
          ]}
        />

        <div style={{ marginBottom: 16 }}>
          <Space>
            {selectedRowKeys.length > 0 && activeTab === 'pending' && (
              <>
                <Popconfirm
                  title="确定批量通过选中的内容吗？"
                  onConfirm={() => handleBatchReview('approve')}
                >
                  <Button type="primary" icon={<CheckOutlined />}>
                    批量通过 ({selectedRowKeys.length})
                  </Button>
                </Popconfirm>
                <Popconfirm
                  title="确定批量拒绝选中的内容吗？"
                  onConfirm={() => handleBatchReview('reject')}
                >
                  <Button danger icon={<CloseOutlined />}>
                    批量拒绝 ({selectedRowKeys.length})
                  </Button>
                </Popconfirm>
              </>
            )}
            <Button icon={<FilterOutlined />}>筛选</Button>
            <Button icon={<ExportOutlined />}>导出</Button>
          </Space>
        </div>

        <Table
          columns={columns}
          dataSource={reviewItems}
          rowKey="id"
          loading={loading}
          rowSelection={activeTab === 'pending' ? rowSelection : undefined}
          pagination={{
            total: reviewItems.length,
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
        />
      </Card>

      {/* 审核模态框 */}
      <Modal
        title="内容审核"
        open={reviewModalVisible}
        onCancel={() => setReviewModalVisible(false)}
        footer={null}
        width={600}
      >
        {currentItem && (
          <Form form={form} onFinish={handleReview} layout="vertical">
            <div style={{ marginBottom: 16 }}>
              <Title level={4}>{currentItem.title}</Title>
              <Paragraph>{currentItem.content}</Paragraph>
            </div>
            
            <Form.Item
              name="action"
              label="审核结果"
              rules={[{ required: true, message: '请选择审核结果' }]}
            >
              <Select placeholder="请选择审核结果">
                <Option value="approve">通过</Option>
                <Option value="reject">拒绝</Option>
              </Select>
            </Form.Item>

            <Form.Item
              name="reason"
              label="审核意见"
              rules={[{ required: true, message: '请输入审核意见' }]}
            >
              <TextArea rows={4} placeholder="请输入审核意见" />
            </Form.Item>

            <Form.Item name="tags" label="标签">
              <Select mode="tags" placeholder="添加标签">
                <Option value="违规内容">违规内容</Option>
                <Option value="垃圾信息">垃圾信息</Option>
                <Option value="优质内容">优质内容</Option>
                <Option value="需要修改">需要修改</Option>
              </Select>
            </Form.Item>

            <Form.Item>
              <Space>
                <Button type="primary" htmlType="submit">
                  提交审核
                </Button>
                <Button onClick={() => setReviewModalVisible(false)}>
                  取消
                </Button>
              </Space>
            </Form.Item>
          </Form>
        )}
      </Modal>

      {/* 详情抽屉 */}
      <Drawer
        title="内容详情"
        placement="right"
        width={600}
        open={detailDrawerVisible}
        onClose={() => setDetailDrawerVisible(false)}
      >
        {currentItem && (
          <div>
            <Title level={4}>{currentItem.title}</Title>
            <Space style={{ marginBottom: 16 }}>
              <Tag color={getStatusColor(currentItem.status)}>
                {getStatusText(currentItem.status)}
              </Tag>
              <Tag color={getPriorityColor(currentItem.priority)}>
                优先级: {currentItem.priority === 'high' ? '高' : currentItem.priority === 'medium' ? '中' : '低'}
              </Tag>
            </Space>
            
            <Divider />
            
            <div style={{ marginBottom: 16 }}>
              <Text strong>作者信息：</Text>
              <div style={{ marginTop: 8 }}>
                <Space>
                  <Avatar src={currentItem.author.avatar} icon={<UserOutlined />} />
                  <Text>{currentItem.author.name}</Text>
                </Space>
              </div>
            </div>

            <div style={{ marginBottom: 16 }}>
              <Text strong>内容：</Text>
              <Paragraph style={{ marginTop: 8 }}>
                {currentItem.content}
              </Paragraph>
            </div>

            <div style={{ marginBottom: 16 }}>
              <Text strong>风险等级：</Text>
              <div style={{ marginTop: 8 }}>
                <Rate disabled value={currentItem.riskLevel} count={5} />
              </div>
            </div>

            {currentItem.tags.length > 0 && (
              <div style={{ marginBottom: 16 }}>
                <Text strong>标签：</Text>
                <div style={{ marginTop: 8 }}>
                  {currentItem.tags.map(tag => (
                    <Tag key={tag}>{tag}</Tag>
                  ))}
                </div>
              </div>
            )}

            <div style={{ marginBottom: 16 }}>
              <Text strong>提交时间：</Text>
              <Text style={{ marginLeft: 8 }}>
                {dayjs(currentItem.submitTime).format('YYYY-MM-DD HH:mm:ss')}
              </Text>
            </div>

            {currentItem.reviewTime && (
              <div style={{ marginBottom: 16 }}>
                <Text strong>审核时间：</Text>
                <Text style={{ marginLeft: 8 }}>
                  {dayjs(currentItem.reviewTime).format('YYYY-MM-DD HH:mm:ss')}
                </Text>
              </div>
            )}

            {currentItem.reviewer && (
              <div style={{ marginBottom: 16 }}>
                <Text strong>审核人：</Text>
                <Text style={{ marginLeft: 8 }}>{currentItem.reviewer.name}</Text>
              </div>
            )}

            {currentItem.reason && (
              <div>
                <Text strong>审核意见：</Text>
                <Paragraph style={{ marginTop: 8 }}>
                  {currentItem.reason}
                </Paragraph>
              </div>
            )}
          </div>
        )}
      </Drawer>

      {/* 审核历史模态框 */}
      <Modal
        title="审核历史"
        open={historyModalVisible}
        onCancel={() => setHistoryModalVisible(false)}
        footer={null}
        width={600}
      >
        <Timeline>
          {reviewHistory.map((item, index) => (
            <Timeline.Item key={index}>
              <div>
                <Text strong>{item.action}</Text>
                <Text type="secondary" style={{ marginLeft: 8 }}>
                  {dayjs(item.time).format('YYYY-MM-DD HH:mm:ss')}
                </Text>
              </div>
              <div>
                <Text>审核人：{item.reviewer}</Text>
              </div>
              {item.reason && (
                <div>
                  <Text type="secondary">{item.reason}</Text>
                </div>
              )}
            </Timeline.Item>
          ))}
        </Timeline>
      </Modal>
    </div>
  );
};

export default ContentReview;