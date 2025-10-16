import React, { useEffect, useState } from 'react';
import { Card, Table, Button, Space, Tag, Typography, message, Select, InputNumber } from 'antd';
import { AlertOutlined, RadarChartOutlined, ReloadOutlined } from '@ant-design/icons';
import { apiClient } from '../../services/api';
import { useNavigate } from 'react-router-dom';

const { Title, Text } = Typography;

interface IssueItem {
  id?: string;
  issue_id?: string;
  title?: string;
  description?: string;
  severity?: 'critical' | 'high' | 'medium' | 'low';
  priority?: number;
  source?: string;
  timestamp?: string;
  suggestions?: string[];
}

const severityColor = (s?: string) => {
  switch (s) {
    case 'critical': return 'red';
    case 'high': return 'volcano';
    case 'medium': return 'orange';
    case 'low': return 'green';
    default: return 'blue';
  }
};

const IssueTracking: React.FC = () => {
  const navigate = useNavigate();
  const [issues, setIssues] = useState<IssueItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [severity, setSeverity] = useState<string | undefined>();
  const [lookback, setLookback] = useState<number>(120);

  const detect = async () => {
    setLoading(true);
    try {
      const payload: any = {};
      if (typeof lookback === 'number') payload.lookback_minutes = lookback;
      if (severity) payload.severity = severity;
      const res = await apiClient.detectIssues(payload);
      setIssues(res.data || []);
      if ((res.data || []).length === 0) {
        message.info('未检测到新的问题');
      }
    } catch (err: any) {
      if (err?.response?.status === 401) {
        message.warning('请先登录后再访问');
        navigate('/login');
      } else {
        message.error('问题检测失败');
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { detect(); }, []);

  const triggerAlert = async (issue: IssueItem) => {
    try {
      const id = issue.issue_id || issue.id || '';
      if (!id) {
        message.warning('该问题缺少唯一标识，无法告警');
        return;
      }
      const res = await apiClient.triggerIssueAlert(id, ['internal']);
      if (res.data.sent) message.success('已触发告警');
    } catch (err: any) {
      if (err?.response?.status === 401) {
        message.warning('请先登录后再访问');
        navigate('/login');
      } else {
        message.error('触发告警失败');
      }
    }
  };

  return (
    <div style={{ padding: 16 }}>
      <Space align="center" style={{ marginBottom: 12 }}>
        <RadarChartOutlined />
        <Title level={4} style={{ margin: 0 }}>智能问题跟踪</Title>
      </Space>

      <Card style={{ marginBottom: 16 }}>
        <Space wrap>
          <Text>回溯范围（分钟）：</Text>
          <InputNumber value={lookback} min={5} max={10080} onChange={(v) => setLookback(Number(v))} />
          <Select
            placeholder="问题严重程度"
            value={severity}
            allowClear
            onChange={setSeverity}
            options={[
              { value: 'critical', label: '致命' },
              { value: 'high', label: '高' },
              { value: 'medium', label: '中' },
              { value: 'low', label: '低' },
            ]}
            style={{ width: 160 }}
          />
          <Button type="primary" icon={<ReloadOutlined />} loading={loading} onClick={detect}>检测问题</Button>
        </Space>
      </Card>

      <Card>
        <Table
          size="small"
          rowKey={(r) => String(r.issue_id || r.id)}
          loading={loading}
          dataSource={issues}
          pagination={{ pageSize: 10 }}
          columns={[
            { title: '时间', dataIndex: 'timestamp', render: (v) => v || '-' },
            { title: '标题', dataIndex: 'title', render: (v) => v || '-' },
            { title: '来源', dataIndex: 'source', render: (v) => v || '-' },
            { title: '严重程度', dataIndex: 'severity', render: (v) => <Tag color={severityColor(v)}>{v || '-'}</Tag> },
            { title: '建议', dataIndex: 'suggestions', render: (arr: string[]) => (arr && arr.length ? arr.join('；') : '-') },
            { title: '操作', key: 'actions', render: (_: any, r: IssueItem) => (
              <Space>
                <Button icon={<AlertOutlined />} onClick={() => triggerAlert(r)}>触发告警</Button>
              </Space>
            ) },
          ]}
        />
      </Card>
    </div>
  );
};

export default IssueTracking;