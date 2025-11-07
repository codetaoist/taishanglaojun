import React, { useState, useEffect } from 'react';
import { 
  Table, 
  Button, 
  Space, 
  Modal, 
  Form, 
  Input, 
  message, 
  Popconfirm 
} from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { configApi } from '../services/laojunApi';
import type { Config } from '../services/laojunApi';

const ConfigManagement: React.FC = () => {
  const [configs, setConfigs] = useState<Config[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingConfig, setEditingConfig] = useState<Config | null>(null);
  const [form] = Form.useForm();

  // 加载配置列表
  const loadConfigs = async () => {
    setLoading(true);
    try {
      const response = await configApi.getAll();
      if (response.code === 200 && response.data) {
        setConfigs(response.data);
      } else {
        message.error('加载配置失败');
      }
    } catch (error) {
      console.error('加载配置出错:', error);
      message.error('加载配置出错');
    } finally {
      setLoading(false);
    }
  };

  // 添加或更新配置
  const handleSaveConfig = async (values: { key: string; value: string; description?: string }) => {
    try {
      const response = await configApi.set(values.key, values.value, values.description);
      if (response.code === 200) {
        message.success(editingConfig ? '更新配置成功' : '添加配置成功');
        setModalVisible(false);
        form.resetFields();
        setEditingConfig(null);
        loadConfigs();
      } else {
        message.error(editingConfig ? '更新配置失败' : '添加配置失败');
      }
    } catch (error) {
      console.error('保存配置出错:', error);
      message.error('保存配置出错');
    }
  };

  // 删除配置
  const handleDeleteConfig = async (key: string) => {
    try {
      const response = await configApi.delete(key);
      if (response.code === 200) {
        message.success('删除配置成功');
        loadConfigs();
      } else {
        message.error('删除配置失败');
      }
    } catch (error) {
      console.error('删除配置出错:', error);
      message.error('删除配置出错');
    }
  };

  // 打开编辑模态框
  const handleEditConfig = (config: Config) => {
    setEditingConfig(config);
    form.setFieldsValue({
      key: config.key,
      value: config.value,
      description: config.description
    });
    setModalVisible(true);
  };

  // 打开添加模态框
  const handleAddConfig = () => {
    setEditingConfig(null);
    form.resetFields();
    setModalVisible(true);
  };

  // 表格列定义
  const columns = [
    {
      title: '配置键',
      dataIndex: 'key',
      key: 'key',
    },
    {
      title: '配置值',
      dataIndex: 'value',
      key: 'value',
      ellipsis: true,
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record: Config) => (
        <Space size="middle">
          <Button 
            type="primary" 
            icon={<EditOutlined />} 
            size="small"
            onClick={() => handleEditConfig(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除这个配置吗？"
            onConfirm={() => handleDeleteConfig(record.key)}
            okText="确定"
            cancelText="取消"
          >
            <Button 
              type="primary" 
              danger 
              icon={<DeleteOutlined />} 
              size="small"
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  // 组件挂载时加载数据
  useEffect(() => {
    loadConfigs();
  }, []);

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <h1>配置管理</h1>
        <Button 
          type="primary" 
          icon={<PlusOutlined />}
          onClick={handleAddConfig}
        >
          添加配置
        </Button>
      </div>
      
      <Table 
        columns={columns} 
        dataSource={configs} 
        rowKey="key"
        loading={loading}
        pagination={{
          pageSize: 10,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条记录`,
        }}
      />

      <Modal
        title={editingConfig ? '编辑配置' : '添加配置'}
        open={modalVisible}
        onCancel={() => {
          setModalVisible(false);
          form.resetFields();
          setEditingConfig(null);
        }}
        footer={null}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSaveConfig}
        >
          <Form.Item
            name="key"
            label="配置键"
            rules={[{ required: true, message: '请输入配置键' }]}
          >
            <Input disabled={!!editingConfig} />
          </Form.Item>
          
          <Form.Item
            name="value"
            label="配置值"
            rules={[{ required: true, message: '请输入配置值' }]}
          >
            <Input.TextArea rows={4} />
          </Form.Item>
          
          <Form.Item
            name="description"
            label="描述"
          >
            <Input.TextArea rows={2} />
          </Form.Item>
          
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                {editingConfig ? '更新' : '添加'}
              </Button>
              <Button 
                onClick={() => {
                  setModalVisible(false);
                  form.resetFields();
                  setEditingConfig(null);
                }}
              >
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default ConfigManagement;