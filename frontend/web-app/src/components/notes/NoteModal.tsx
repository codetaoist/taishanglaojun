import React, { useState, useEffect } from 'react';
import { Modal, Form, Input, Switch, Select, message, Button, Space } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import { apiClient } from '../../services/api';

const { TextArea } = Input;
const { Option } = Select;

interface NoteModalProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
  wisdomId: string;
  wisdomTitle: string;
  existingNote?: {
    id: string;
    title: string;
    content: string;
    is_private: boolean;
    tags: string[];
  } | null;
}

const NoteModal: React.FC<NoteModalProps> = ({
  visible,
  onCancel,
  onSuccess,
  wisdomId,
  wisdomTitle,
  existingNote
}) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [tags, setTags] = useState<string[]>([]);
  const [newTag, setNewTag] = useState('');

  useEffect(() => {
    if (visible) {
      if (existingNote) {
        form.setFieldsValue({
          title: existingNote.title,
          content: existingNote.content,
          is_private: existingNote.is_private,
        });
        setTags(existingNote.tags || []);
      } else {
        form.resetFields();
        form.setFieldsValue({
          title: `关于《${wisdomTitle}》的笔记`,
          is_private: true,
        });
        setTags([]);
      }
    }
  }, [visible, existingNote, wisdomTitle, form]);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setLoading(true);

      const noteData = {
        title: values.title,
        content: values.content,
        is_private: values.is_private,
        tags: tags,
      };

      if (existingNote) {
        await apiClient.updateNote(wisdomId, noteData);
        message.success('笔记更新成功');
      } else {
        await apiClient.createNote(wisdomId, noteData);
        message.success('笔记创建成功');
      }

      onSuccess();
      onCancel();
    } catch (error) {
      message.error('操作失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteNote = async () => {
    if (!existingNote) return;

    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这条笔记吗？此操作不可恢复。',
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          setLoading(true);
          await apiClient.deleteNote(wisdomId);
          message.success('笔记删除成功');
          onSuccess();
          onCancel();
        } catch (error) {
          message.error('删除失败，请重试');
        } finally {
          setLoading(false);
        }
      },
    });
  };

  const addTag = () => {
    if (newTag && !tags.includes(newTag)) {
      setTags([...tags, newTag]);
      setNewTag('');
    }
  };

  const removeTag = (tagToRemove: string) => {
    setTags(tags.filter(tag => tag !== tagToRemove));
  };

  return (
    <Modal
      title={existingNote ? '编辑笔记' : '添加笔记'}
      open={visible}
      onCancel={onCancel}
      width={600}
      footer={[
        <Button key="cancel" onClick={onCancel}>
          取消
        </Button>,
        existingNote && (
          <Button
            key="delete"
            type="primary"
            danger
            icon={<DeleteOutlined />}
            onClick={handleDeleteNote}
            loading={loading}
          >
            删除笔记
          </Button>
        ),
        <Button
          key="submit"
          type="primary"
          onClick={handleSubmit}
          loading={loading}
        >
          {existingNote ? '更新笔记' : '保存笔记'}
        </Button>,
      ].filter(Boolean)}
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{
          is_private: true,
        }}
      >
        <Form.Item
          name="title"
          label="笔记标题"
          rules={[{ required: true, message: '请输入笔记标题' }]}
        >
          <Input placeholder="为这条笔记起个标题..." />
        </Form.Item>

        <Form.Item
          name="content"
          label="笔记内容"
          rules={[{ required: true, message: '请输入笔记内容' }]}
        >
          <TextArea
            rows={8}
            placeholder="记录你的思考、感悟或疑问..."
            showCount
            maxLength={2000}
          />
        </Form.Item>

        <Form.Item name="is_private" label="隐私设置" valuePropName="checked">
          <Switch
            checkedChildren="私密"
            unCheckedChildren="公开"
            defaultChecked
          />
        </Form.Item>

        <Form.Item label="标签">
          <div style={{ marginBottom: 8 }}>
            {tags.map(tag => (
              <span
                key={tag}
                style={{
                  display: 'inline-block',
                  padding: '2px 8px',
                  margin: '2px 4px 2px 0',
                  backgroundColor: '#f0f0f0',
                  borderRadius: '4px',
                  fontSize: '12px',
                  cursor: 'pointer',
                }}
                onClick={() => removeTag(tag)}
              >
                {tag} ×
              </span>
            ))}
          </div>
          <Space.Compact style={{ width: '100%' }}>
            <Input
              placeholder="添加标签..."
              value={newTag}
              onChange={(e) => setNewTag(e.target.value)}
              onPressEnter={addTag}
            />
            <Button type="primary" icon={<PlusOutlined />} onClick={addTag}>
              添加
            </Button>
          </Space.Compact>
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default NoteModal;