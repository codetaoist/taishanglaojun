import React, { useState, useEffect, useCallback } from 'react';
import {
  Card,
  Form,
  Input,
  Select,
  Button,
  Space,
  message,
  Row,
  Col,
  Typography,
  Tag,
  Divider,
  Switch,
  DatePicker,
  InputNumber
} from 'antd';
import {
  SaveOutlined,
  ArrowLeftOutlined,
  EyeOutlined,
  PlusOutlined
} from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { apiClient } from '../../services/api';

const { TextArea } = Input;
const { Option } = Select;
const { Title } = Typography;

interface WisdomFormData {
  title: string;
  content: string;
  author?: string;
  dynasty?: string;
  category: string;
  school?: string;
  source?: string;
  tags: string[];
  interpretation?: string;
  background?: string;
  significance?: string;
  is_published: boolean;
  publish_date?: string;
  views?: number;
  likes?: number;
}

const WisdomEditor: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEdit = !!id;
  
  const [form] = Form.useForm();
  const [saving, setSaving] = useState(false);
  const [categories, setCategories] = useState<Array<{ id: string; name: string; description?: string }>>([]);
  const [schools, setSchools] = useState<Array<{ id: string; name: string; description?: string }>>([]);
  const [tags, setTags] = useState<Array<{ id: string; name: string; description?: string }>>([]);
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [newTag, setNewTag] = useState('');

  // 加载智慧内容（编辑模式）
  const loadWisdom = useCallback(async () => {
    if (!id) return;
    
    try {
      const response = await apiClient.getWisdomDetail(id);
      if (response.success && response.data) {
        const wisdom = response.data;
        form.setFieldsValue({
          title: wisdom.title,
          content: wisdom.content,
          author: wisdom.author,
          dynasty: wisdom.dynasty,
          category: wisdom.category,
          school: wisdom.school,
          source: wisdom.source,
          tags: wisdom.tags || [],
          interpretation: wisdom.interpretation,
          background: wisdom.background,
          significance: wisdom.significance,
          is_published: wisdom.is_published !== false,
          publish_date: wisdom.publish_date ? new Date(wisdom.publish_date) : null,
          views: wisdom.views,
          likes: wisdom.likes
        });
        setSelectedTags(wisdom.tags || []);
      } else {
        message.error('获取智慧内容失败');
        navigate('/admin/wisdom');
      }
    } catch (error) {
      console.error('加载智慧内容失败:', error);
      message.error('网络错误，请稍后重试');
      navigate('/admin/wisdom');
    }
  }, [id, form, navigate, setSelectedTags]);

  // 加载选项数据
  const loadOptions = async () => {
    try {
      const [categoriesRes, schoolsRes, tagsRes] = await Promise.all([
        apiClient.getCategories(),
        apiClient.getSchools(),
        apiClient.getTags()
      ]);

      if (categoriesRes.success) {
        setCategories(categoriesRes.data || []);
      }
      
      if (schoolsRes.success) {
        setSchools(schoolsRes.data || []);
      }

      if (tagsRes.success) {
        setTags(tagsRes.data || []);
      }
    } catch (error) {
      console.error('加载选项数据失败:', error);
    }
  };

  // 保存智慧内容
  const handleSave = async (values: WisdomFormData) => {
    setSaving(true);
    try {
      const data = {
        ...values,
        tags: selectedTags,
        publish_date: values.publish_date ? values.publish_date.toISOString() : null
      };

      let response;
      if (isEdit) {
        response = await apiClient.updateWisdom(id!, data);
      } else {
        response = await apiClient.createWisdom(data);
      }

      if (response.success) {
        message.success(isEdit ? '更新成功' : '创建成功');
        navigate('/admin/wisdom');
      } else {
        message.error(isEdit ? '更新失败' : '创建失败');
      }
    } catch {
      message.error('保存失败');
    }
    setSaving(false);
  };

  // 添加标签
  const handleAddTag = () => {
    if (newTag && !selectedTags.includes(newTag)) {
      setSelectedTags([...selectedTags, newTag]);
      setNewTag('');
    }
  };

  // 删除标签
  const handleRemoveTag = (tagToRemove: string) => {
    setSelectedTags(selectedTags.filter(tag => tag !== tagToRemove));
  };

  // 预览
  const handlePreview = () => {
    // 这里可以打开预览模态框或跳转到预览页面
    message.info('预览功能开发中...');
  };

  useEffect(() => {
    loadOptions();
    if (isEdit) {
      loadWisdom();
    }
  }, [id, isEdit, loadWisdom]);

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <Card>
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <Button 
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate('/admin/wisdom')}
            >
              返回
            </Button>
            <div>
              <Title level={2} className="mb-2">
                {isEdit ? '编辑智慧内容' : '添加智慧内容'}
              </Title>
              <p className="text-gray-600 mb-0">
                {isEdit ? '修改现有的文化智慧内容' : '创建新的文化智慧内容'}
              </p>
            </div>
          </div>
          <Space>
            <Button 
              icon={<EyeOutlined />}
              onClick={handlePreview}
            >
              预览
            </Button>
            <Button 
              type="primary" 
              icon={<SaveOutlined />}
              loading={saving}
              onClick={() => form.submit()}
            >
              {isEdit ? '更新' : '保存'}
            </Button>
          </Space>
        </div>
      </Card>

      {/* 编辑表单 */}
      <Form
        form={form}
        layout="vertical"
        onFinish={handleSave}
        initialValues={{
          is_published: true,
          views: 0,
          likes: 0
        }}
      >
        <Row gutter={[24, 0]}>
          {/* 左侧主要内容 */}
          <Col xs={24} lg={16}>
            <Card title="基本信息" className="mb-6">
              <Form.Item
                name="title"
                label="标题"
                rules={[{ required: true, message: '请输入标题' }]}
              >
                <Input placeholder="请输入智慧内容标题" size="large" />
              </Form.Item>

              <Form.Item
                name="content"
                label="内容"
                rules={[{ required: true, message: '请输入内容' }]}
              >
                <TextArea
                  placeholder="请输入智慧内容正文"
                  rows={8}
                  showCount
                  maxLength={5000}
                />
              </Form.Item>

              <Row gutter={[16, 0]}>
                <Col xs={24} md={12}>
                  <Form.Item name="author" label="作者">
                    <Input placeholder="请输入作者" />
                  </Form.Item>
                </Col>
                <Col xs={24} md={12}>
                  <Form.Item name="dynasty" label="朝代">
                    <Input placeholder="请输入朝代" />
                  </Form.Item>
                </Col>
              </Row>

              <Form.Item name="source" label="出处">
                <Input placeholder="请输入出处" />
              </Form.Item>
            </Card>

            <Card title="详细解释" className="mb-6">
              <Form.Item name="interpretation" label="释义">
                <TextArea
                  placeholder="请输入详细释义"
                  rows={4}
                  showCount
                  maxLength={2000}
                />
              </Form.Item>

              <Form.Item name="background" label="背景">
                <TextArea
                  placeholder="请输入历史背景"
                  rows={4}
                  showCount
                  maxLength={2000}
                />
              </Form.Item>

              <Form.Item name="significance" label="意义">
                <TextArea
                  placeholder="请输入现实意义"
                  rows={4}
                  showCount
                  maxLength={2000}
                />
              </Form.Item>
            </Card>
          </Col>

          {/* 右侧设置 */}
          <Col xs={24} lg={8}>
            <Card title="分类设置" className="mb-6">
              <Form.Item
                name="category"
                label="分类"
                rules={[{ required: true, message: '请选择分类' }]}
              >
                <Select placeholder="请选择分类">
                  {categories.map(cat => (
                    <Option key={cat.id} value={cat.name}>
                      {cat.name}
                    </Option>
                  ))}
                </Select>
              </Form.Item>

              <Form.Item name="school" label="学派">
                <Select placeholder="请选择学派" allowClear>
                  {schools.map(school => (
                    <Option key={school.id} value={school.name}>
                      {school.name}
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Card>

            <Card title="标签管理" className="mb-6">
              <div className="space-y-4">
                <div className="flex space-x-2">
                  <Input
                    placeholder="添加新标签"
                    value={newTag}
                    onChange={(e) => setNewTag(e.target.value)}
                    onPressEnter={handleAddTag}
                  />
                  <Button 
                    type="primary" 
                    icon={<PlusOutlined />}
                    onClick={handleAddTag}
                  >
                    添加
                  </Button>
                </div>

                <div className="space-y-2">
                  <div className="text-sm text-gray-600">已选标签：</div>
                  <div className="flex flex-wrap gap-2">
                    {selectedTags.map(tag => (
                      <Tag
                        key={tag}
                        closable
                        onClose={() => handleRemoveTag(tag)}
                        color="blue"
                      >
                        {tag}
                      </Tag>
                    ))}
                  </div>
                </div>

                <div className="space-y-2">
                  <div className="text-sm text-gray-600">推荐标签：</div>
                  <div className="flex flex-wrap gap-2">
                    {tags
                      .filter(tag => !selectedTags.includes(tag.name))
                      .slice(0, 10)
                      .map(tag => (
                        <Tag
                          key={tag.id}
                          className="cursor-pointer"
                          onClick={() => setSelectedTags([...selectedTags, tag.name])}
                        >
                          {tag.name}
                        </Tag>
                      ))}
                  </div>
                </div>
              </div>
            </Card>

            <Card title="发布设置" className="mb-6">
              <Form.Item name="is_published" label="发布状态" valuePropName="checked">
                <Switch checkedChildren="已发布" unCheckedChildren="草稿" />
              </Form.Item>

              <Form.Item name="publish_date" label="发布时间">
                <DatePicker 
                  showTime 
                  className="w-full"
                  placeholder="选择发布时间"
                />
              </Form.Item>

              {isEdit && (
                <>
                  <Divider />
                  <Form.Item name="views" label="阅读量">
                    <InputNumber min={0} className="w-full" />
                  </Form.Item>

                  <Form.Item name="likes" label="点赞数">
                    <InputNumber min={0} className="w-full" />
                  </Form.Item>
                </>
              )}
            </Card>
          </Col>
        </Row>
      </Form>
    </div>
  );
};

export default WisdomEditor;