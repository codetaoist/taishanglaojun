import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Typography, 
  Table, 
  Input, 
  Button, 
  Space, 
  Tag, 
  Modal, 
  DatePicker, 
  Select, 
  Form, 
  Row, 
  Col,
  Tooltip,
  Descriptions
} from 'antd';
import { 
  SearchOutlined, 
  ReloadOutlined, 
  InfoCircleOutlined,
  ExportOutlined
} from '@ant-design/icons';
import { auditLogApi, AuditLog } from '../services/laojunApi';
import dayjs from 'dayjs';

const { Title } = Typography;
const { RangePicker } = DatePicker;
const { Option } = Select;

const AuditLogs: React.FC = () => {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });
  const [filters, setFilters] = useState({
    userId: '',
    action: '',
    resource: '',
    dateRange: null as any,
  });

  // 获取审计日志列表
  const fetchLogs = async (page = 1, pageSize = 10) => {
    setLoading(true);
    try {
      const response = await auditLogApi.getAll(page, pageSize);
      if (response.code === 200) {
        setLogs(response.data?.logs || []);
        setPagination({
          current: page,
          pageSize,
          total: response.data?.total || 0,
        });
      } else {
        console.error('获取审计日志失败:', response.message);
      }
    } catch (error) {
      console.error('获取审计日志失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 初始化加载日志列表
  useEffect(() => {
    fetchLogs(pagination.current, pagination.pageSize);
  }, []);

  // 处理表格分页变化
  const handleTableChange = (pagination: any) => {
    fetchLogs(pagination.current, pagination.pageSize);
  };

  // 搜索处理
  const handleSearch = () => {
    // 这里应该根据filters参数进行过滤，简化处理直接刷新
    fetchLogs(1, pagination.pageSize);
  };

  // 重置搜索条件
  const handleReset = () => {
    setFilters({
      userId: '',
      action: '',
      resource: '',
      dateRange: null,
    });
    fetchLogs(1, pagination.pageSize);
  };

  // 查看日志详情
  const handleViewLogDetail = (log: AuditLog) => {
    setSelectedLog(log);
    setDetailModalVisible(true);
  };

  // 导出日志
  const handleExportLogs = () => {
    // 这里可以实现导出功能，简化处理
    console.log('导出日志功能待实现');
  };

  // 获取操作类型标签颜色
  const getActionColor = (action: string) => {
    switch (action.toLowerCase()) {
      case 'create':
      case 'add':
        return 'success';
      case 'update':
      case 'modify':
        return 'processing';
      case 'delete':
      case 'remove':
        return 'error';
      case 'login':
        return 'default';
      case 'logout':
        return 'warning';
      default:
        return 'default';
    }
  };

  // 格式化时间
  const formatTime = (time: string) => {
    return dayjs(time).format('YYYY-MM-DD HH:mm:ss');
  };

  // 表格列定义
  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 200,
      ellipsis: true,
    },
    {
      title: '用户ID',
      dataIndex: 'userId',
      key: 'userId',
      width: 150,
      ellipsis: true,
    },
    {
      title: '操作',
      dataIndex: 'action',
      key: 'action',
      width: 120,
      render: (action: string) => (
        <Tag color={getActionColor(action)}>
          {action}
        </Tag>
      ),
    },
    {
      title: '资源',
      dataIndex: 'resource',
      key: 'resource',
      width: 150,
      ellipsis: true,
    },
    {
      title: '资源ID',
      dataIndex: 'resourceId',
      key: 'resourceId',
      width: 150,
      ellipsis: true,
      render: (resourceId: string) => resourceId || '-',
    },
    {
      title: 'IP地址',
      dataIndex: 'ipAddress',
      key: 'ipAddress',
      width: 130,
      render: (ip: string) => ip || '-',
    },
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 180,
      render: (timestamp: string) => formatTime(timestamp),
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      render: (_: any, record: AuditLog) => (
        <Tooltip title="查看详情">
          <Button 
            type="link" 
            icon={<InfoCircleOutlined />} 
            onClick={() => handleViewLogDetail(record)}
          />
        </Tooltip>
      ),
    },
  ];

  return (
    <div>
      <Title level={2}>审计日志</Title>
      
      <Card>
        {/* 搜索区域 */}
        <div style={{ marginBottom: 16 }}>
          <Form layout="inline">
            <Row gutter={[16, 16]}>
              <Col>
                <Form.Item label="用户ID">
                  <Input
                    placeholder="请输入用户ID"
                    value={filters.userId}
                    onChange={(e) => setFilters({ ...filters, userId: e.target.value })}
                    allowClear
                  />
                </Form.Item>
              </Col>
              <Col>
                <Form.Item label="操作">
                  <Select
                    placeholder="请选择操作类型"
                    value={filters.action}
                    onChange={(value) => setFilters({ ...filters, action: value })}
                    allowClear
                    style={{ width: 120 }}
                  >
                    <Option value="create">创建</Option>
                    <Option value="update">更新</Option>
                    <Option value="delete">删除</Option>
                    <Option value="login">登录</Option>
                    <Option value="logout">登出</Option>
                  </Select>
                </Form.Item>
              </Col>
              <Col>
                <Form.Item label="资源">
                  <Input
                    placeholder="请输入资源名称"
                    value={filters.resource}
                    onChange={(e) => setFilters({ ...filters, resource: e.target.value })}
                    allowClear
                  />
                </Form.Item>
              </Col>
              <Col>
                <Form.Item label="时间范围">
                  <RangePicker
                    value={filters.dateRange}
                    onChange={(dates) => setFilters({ ...filters, dateRange: dates })}
                  />
                </Form.Item>
              </Col>
              <Col>
                <Form.Item>
                  <Space>
                    <Button 
                      type="primary" 
                      icon={<SearchOutlined />}
                      onClick={handleSearch}
                    >
                      搜索
                    </Button>
                    <Button onClick={handleReset}>
                      重置
                    </Button>
                    <Button 
                      icon={<ReloadOutlined />}
                      onClick={() => fetchLogs(pagination.current, pagination.pageSize)}
                      loading={loading}
                    >
                      刷新
                    </Button>
                    <Button 
                      icon={<ExportOutlined />}
                      onClick={handleExportLogs}
                    >
                      导出
                    </Button>
                  </Space>
                </Form.Item>
              </Col>
            </Row>
          </Form>
        </div>
        
        {/* 日志表格 */}
        <Table
          columns={columns}
          dataSource={logs}
          rowKey="id"
          loading={loading}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
          }}
          onChange={handleTableChange}
        />
      </Card>

      {/* 日志详情模态框 */}
      <Modal
        title="审计日志详情"
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailModalVisible(false)}>
            关闭
          </Button>
        ]}
        width={800}
      >
        {selectedLog && (
          <div>
            <Descriptions bordered column={2}>
              <Descriptions.Item label="ID">{selectedLog.id}</Descriptions.Item>
              <Descriptions.Item label="租户ID">{selectedLog.tenantId}</Descriptions.Item>
              <Descriptions.Item label="用户ID">{selectedLog.userId}</Descriptions.Item>
              <Descriptions.Item label="操作">
                <Tag color={getActionColor(selectedLog.action)}>
                  {selectedLog.action}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="资源">{selectedLog.resource}</Descriptions.Item>
              <Descriptions.Item label="资源ID">{selectedLog.resourceId || '-'}</Descriptions.Item>
              <Descriptions.Item label="IP地址">{selectedLog.ipAddress || '-'}</Descriptions.Item>
              <Descriptions.Item label="时间">{formatTime(selectedLog.timestamp)}</Descriptions.Item>
              <Descriptions.Item label="用户代理" span={2}>
                {selectedLog.userAgent || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="详情" span={2}>
                {selectedLog.details ? (
                  <pre style={{ 
                    background: '#f5f5f5', 
                    padding: '10px', 
                    borderRadius: '4px',
                    overflow: 'auto',
                    maxHeight: '200px'
                  }}>
                    {JSON.stringify(selectedLog.details, null, 2)}
                  </pre>
                ) : '-'}
              </Descriptions.Item>
            </Descriptions>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default AuditLogs;