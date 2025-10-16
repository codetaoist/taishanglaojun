import React, { useState } from 'react';
import { Card, Timeline, Tag, Space, Typography, Row, Col, Select, Input, Button, Modal, Descriptions, Divider, List, Avatar } from 'antd';
import { BranchesOutlined, ClockCircleOutlined, CheckCircleOutlined, ExclamationCircleOutlined, EyeOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';

const { Title, Text, Paragraph } = Typography;
const { Search } = Input;

interface VersionChange {
  id: string;
  type: 'added' | 'modified' | 'deprecated' | 'removed';
  description: string;
  impact: 'breaking' | 'compatible' | 'minor';
}

interface APIVersion {
  id: string;
  version: string;
  apiName: string;
  apiPath: string;
  releaseDate: string;
  status: 'current' | 'deprecated' | 'beta' | 'alpha';
  changes: VersionChange[];
  developer: string;
  notes?: string;
  migrationGuide?: string;
}

const APIVersions: React.FC = () => {
  const { t } = useTranslation();
  const [selectedAPI, setSelectedAPI] = useState<string>('all');
  const [searchText, setSearchText] = useState('');
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [selectedVersion, setSelectedVersion] = useState<APIVersion | null>(null);

  // 模拟版本数据
  const [versionData] = useState<APIVersion[]>([
    {
      id: '1',
      version: 'v2.0.0',
      apiName: '多模态分析',
      apiPath: '/api/ai/multimodal',
      releaseDate: '2024-01-20',
      status: 'beta',
      developer: '赵六',
      changes: [
        {
          id: '1-1',
          type: 'added',
          description: '新增视频分析功能',
          impact: 'compatible'
        },
        {
          id: '1-2',
          type: 'modified',
          description: '优化图像识别算法，提升准确率',
          impact: 'compatible'
        },
        {
          id: '1-3',
          type: 'deprecated',
          description: '废弃旧版本的文本分析接口',
          impact: 'breaking'
        }
      ],
      notes: '这是一个重大版本更新，引入了全新的多模态分析能力',
      migrationGuide: '请参考迁移指南文档进行升级'
    },
    {
      id: '2',
      version: 'v1.5.0',
      apiName: '图像生成',
      apiPath: '/api/ai/image/generate',
      releaseDate: '2024-01-15',
      status: 'current',
      developer: '钱七',
      changes: [
        {
          id: '2-1',
          type: 'added',
          description: '支持高分辨率图像生成',
          impact: 'compatible'
        },
        {
          id: '2-2',
          type: 'modified',
          description: '优化生成速度，减少50%处理时间',
          impact: 'compatible'
        }
      ],
      notes: '性能优化版本，大幅提升用户体验'
    },
    {
      id: '3',
      version: 'v1.1.0',
      apiName: '创建项目',
      apiPath: '/api/projects',
      releaseDate: '2024-01-10',
      status: 'current',
      developer: '王五',
      changes: [
        {
          id: '3-1',
          type: 'added',
          description: '新增项目模板功能',
          impact: 'compatible'
        },
        {
          id: '3-2',
          type: 'modified',
          description: '增强项目权限控制',
          impact: 'minor'
        }
      ]
    },
    {
      id: '4',
      version: 'v1.0.0',
      apiName: '用户登录',
      apiPath: '/api/auth/login',
      releaseDate: '2024-01-01',
      status: 'current',
      developer: '张三',
      changes: [
        {
          id: '4-1',
          type: 'added',
          description: '初始版本发布',
          impact: 'compatible'
        }
      ]
    },
    {
      id: '5',
      version: 'v0.9.0',
      apiName: '用户登录',
      apiPath: '/api/auth/login',
      releaseDate: '2023-12-20',
      status: 'deprecated',
      developer: '张三',
      changes: [
        {
          id: '5-1',
          type: 'deprecated',
          description: '旧版本登录接口，已废弃',
          impact: 'breaking'
        }
      ]
    }
  ]);

  const getStatusColor = (status: string) => {
    const colors = {
      current: 'green',
      deprecated: 'red',
      beta: 'blue',
      alpha: 'orange'
    };
    return colors[status as keyof typeof colors] || 'default';
  };

  const getStatusText = (status: string) => {
    const texts = {
      current: '当前版本',
      deprecated: '已废弃',
      beta: '测试版',
      alpha: '内测版'
    };
    return texts[status as keyof typeof texts] || status;
  };

  const getChangeTypeColor = (type: string) => {
    const colors = {
      added: 'green',
      modified: 'blue',
      deprecated: 'orange',
      removed: 'red'
    };
    return colors[type as keyof typeof colors] || 'default';
  };

  const getChangeTypeText = (type: string) => {
    const texts = {
      added: '新增',
      modified: '修改',
      deprecated: '废弃',
      removed: '删除'
    };
    return texts[type as keyof typeof texts] || type;
  };

  const getImpactColor = (impact: string) => {
    const colors = {
      breaking: 'red',
      compatible: 'green',
      minor: 'blue'
    };
    return colors[impact as keyof typeof colors] || 'default';
  };

  const getImpactText = (impact: string) => {
    const texts = {
      breaking: '破坏性变更',
      compatible: '向后兼容',
      minor: '小版本更新'
    };
    return texts[impact as keyof typeof texts] || impact;
  };

  // 获取唯一的API列表
  const apiList = Array.from(new Set(versionData.map(item => item.apiName)));

  // 过滤数据
  const filteredData = versionData.filter(item => {
    const matchesAPI = selectedAPI === 'all' || item.apiName === selectedAPI;
    const matchesSearch = item.apiName.toLowerCase().includes(searchText.toLowerCase()) ||
                         item.version.toLowerCase().includes(searchText.toLowerCase()) ||
                         item.apiPath.toLowerCase().includes(searchText.toLowerCase());
    
    return matchesAPI && matchesSearch;
  });

  const showVersionDetail = (version: APIVersion) => {
    setSelectedVersion(version);
    setDetailModalVisible(true);
  };

  return (
    <div style={{ padding: '24px' }}>
      <Title level={2}>
        <BranchesOutlined /> 版本管理
      </Title>
      <Paragraph type="secondary">
        记录接口变更历史，追踪版本演进过程，便于版本管理和回滚操作
      </Paragraph>

      {/* 过滤器 */}
      <Card size="small" style={{ marginBottom: '16px' }}>
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} sm={8}>
            <Search
              placeholder="搜索接口名称、版本或路径..."
              allowClear
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
            />
          </Col>
          <Col xs={12} sm={6}>
            <Select
              style={{ width: '100%' }}
              placeholder="选择接口"
              value={selectedAPI}
              onChange={setSelectedAPI}
            >
              <Select.Option value="all">全部接口</Select.Option>
              {apiList.map(api => (
                <Select.Option key={api} value={api}>{api}</Select.Option>
              ))}
            </Select>
          </Col>
          <Col xs={12} sm={10}>
            <Space>
              <Button type="primary" icon={<PlusOutlined />}>
                创建新版本
              </Button>
              <Button icon={<EditOutlined />}>
                批量管理
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* 版本时间线 */}
      <Card>
        <Timeline mode="left">
          {filteredData.map((version, index) => (
            <Timeline.Item
              key={version.id}
              dot={
                version.status === 'current' ? <CheckCircleOutlined style={{ color: '#52c41a' }} /> :
                version.status === 'deprecated' ? <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} /> :
                <ClockCircleOutlined style={{ color: '#1890ff' }} />
              }
              color={
                version.status === 'current' ? 'green' :
                version.status === 'deprecated' ? 'red' : 'blue'
              }
            >
              <Card size="small" style={{ marginBottom: '16px' }}>
                <Row justify="space-between" align="top">
                  <Col span={18}>
                    <Space direction="vertical" size="small" style={{ width: '100%' }}>
                      <div>
                        <Space>
                          <Text strong style={{ fontSize: '16px' }}>
                            {version.apiName} {version.version}
                          </Text>
                          <Tag color={getStatusColor(version.status)}>
                            {getStatusText(version.status)}
                          </Tag>
                        </Space>
                      </div>
                      
                      <Text code>{version.apiPath}</Text>
                      
                      <div>
                        <Text type="secondary">发布时间：{version.releaseDate}</Text>
                        <Divider type="vertical" />
                        <Text type="secondary">开发者：{version.developer}</Text>
                      </div>

                      {version.notes && (
                        <Paragraph style={{ margin: 0 }}>{version.notes}</Paragraph>
                      )}

                      <div>
                        <Text strong>主要变更：</Text>
                        <List
                          size="small"
                          dataSource={version.changes.slice(0, 2)}
                          renderItem={(change) => (
                            <List.Item style={{ padding: '4px 0', border: 'none' }}>
                              <Space>
                                <Tag color={getChangeTypeColor(change.type)} size="small">
                                  {getChangeTypeText(change.type)}
                                </Tag>
                                <Text>{change.description}</Text>
                                <Tag color={getImpactColor(change.impact)} size="small">
                                  {getImpactText(change.impact)}
                                </Tag>
                              </Space>
                            </List.Item>
                          )}
                        />
                        {version.changes.length > 2 && (
                          <Text type="secondary">还有 {version.changes.length - 2} 项变更...</Text>
                        )}
                      </div>
                    </Space>
                  </Col>
                  
                  <Col span={6} style={{ textAlign: 'right' }}>
                    <Button 
                      type="link" 
                      icon={<EyeOutlined />}
                      onClick={() => showVersionDetail(version)}
                    >
                      查看详情
                    </Button>
                  </Col>
                </Row>
              </Card>
            </Timeline.Item>
          ))}
        </Timeline>
      </Card>

      {/* 版本详情弹窗 */}
      <Modal
        title={selectedVersion ? `${selectedVersion.apiName} ${selectedVersion.version} 详情` : '版本详情'}
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={null}
        width={800}
      >
        {selectedVersion && (
          <div>
            <Descriptions bordered column={2}>
              <Descriptions.Item label="接口名称">{selectedVersion.apiName}</Descriptions.Item>
              <Descriptions.Item label="版本号">{selectedVersion.version}</Descriptions.Item>
              <Descriptions.Item label="接口路径" span={2}>
                <Text code>{selectedVersion.apiPath}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="发布时间">{selectedVersion.releaseDate}</Descriptions.Item>
              <Descriptions.Item label="开发者">{selectedVersion.developer}</Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={getStatusColor(selectedVersion.status)}>
                  {getStatusText(selectedVersion.status)}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="变更数量">{selectedVersion.changes.length} 项</Descriptions.Item>
            </Descriptions>

            {selectedVersion.notes && (
              <div style={{ marginTop: '16px' }}>
                <Title level={5}>版本说明</Title>
                <Paragraph>{selectedVersion.notes}</Paragraph>
              </div>
            )}

            <div style={{ marginTop: '16px' }}>
              <Title level={5}>详细变更</Title>
              <List
                dataSource={selectedVersion.changes}
                renderItem={(change) => (
                  <List.Item>
                    <List.Item.Meta
                      avatar={
                        <Avatar 
                          style={{ 
                            backgroundColor: getChangeTypeColor(change.type),
                            fontSize: '12px'
                          }}
                          size="small"
                        >
                          {getChangeTypeText(change.type)[0]}
                        </Avatar>
                      }
                      title={
                        <Space>
                          <Tag color={getChangeTypeColor(change.type)}>
                            {getChangeTypeText(change.type)}
                          </Tag>
                          <Tag color={getImpactColor(change.impact)}>
                            {getImpactText(change.impact)}
                          </Tag>
                        </Space>
                      }
                      description={change.description}
                    />
                  </List.Item>
                )}
              />
            </div>

            {selectedVersion.migrationGuide && (
              <div style={{ marginTop: '16px' }}>
                <Title level={5}>迁移指南</Title>
                <Paragraph>{selectedVersion.migrationGuide}</Paragraph>
              </div>
            )}
          </div>
        )}
      </Modal>
    </div>
  );
};

export default APIVersions;