import React, { useEffect, useMemo, useState } from 'react';
import { Card, Tabs, Table, Input, Button, Space, Tag, Typography, Divider, Select, message, DatePicker, Statistic, Row, Col, Switch } from 'antd';
import { FileSearchOutlined, ReloadOutlined } from '@ant-design/icons';
import dayjs, { Dayjs } from 'dayjs';
import { apiClient } from '../../services/api';
import { useNavigate } from 'react-router-dom';

const { Title, Text } = Typography;
const { RangePicker } = DatePicker;

interface LogItem {
  time?: string;
  timestamp?: string;
  level?: string;
  source?: string;
  message?: string;
  user_id?: string;
  ip?: string;
  user_agent?: string;
  extra?: any;
}

const LogsManagement: React.FC = () => {
  const navigate = useNavigate();
  const [logs, setLogs] = useState<LogItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [level, setLevel] = useState<string | undefined>();
  const [source, setSource] = useState<string | undefined>();
  const [keyword, setKeyword] = useState('');
  const [range, setRange] = useState<[Dayjs, Dayjs] | null>(null);
  const [userIdFilter, setUserIdFilter] = useState<string>('');
  const [ipFilter, setIpFilter] = useState<string>('');

  const [stats, setStats] = useState<Record<string, number>>({});
  const [statsLoading, setStatsLoading] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [refreshInterval, setRefreshInterval] = useState<number>(10000);
  const [lastUpdatedAt, setLastUpdatedAt] = useState<string>('');

  const filtered = useMemo(() => {
    const lower = keyword.toLowerCase();
    return logs.filter(l => {
      const msg = (l.message || '') + ' ' + (l.source || '');
      const baseMatch = !keyword || msg.toLowerCase().includes(lower);
      const userMatch = !userIdFilter || (l.user_id || '').toLowerCase().includes(userIdFilter.toLowerCase());
      const ipMatch = !ipFilter || (l.ip || '').toLowerCase().includes(ipFilter.toLowerCase());
      return baseMatch && userMatch && ipMatch;
    });
  }, [logs, keyword, userIdFilter, ipFilter]);

  const fetchLogs = async () => {
    setLoading(true);
    try {
      const params: any = {};
      if (level) params.level = level;
      if (source) params.source = source;
      if (range) {
        params.start = range[0].toISOString();
        params.end = range[1].toISOString();
      }
      const res = await apiClient.getSystemLogs(params);
      setLogs(res.data || []);
      setLastUpdatedAt(dayjs().format('YYYY-MM-DD HH:mm:ss'));
    } catch (err: any) {
      if (err?.response?.status === 401) {
        message.warning('请先登录后再访问');
        navigate('/login');
      } else {
        message.error('获取日志失败');
      }
    } finally {
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    setStatsLoading(true);
    try {
      const res = await apiClient.getSystemLogStats();
      setStats(res.data || {});
    } catch (err: any) {
      if (err?.response?.status === 401) {
        message.warning('请先登录后再访问');
        navigate('/login');
      }
    } finally {
      setStatsLoading(false);
    }
  };

  useEffect(() => {
    fetchLogs();
    fetchStats();
  }, []);

  useEffect(() => {
    fetchLogs();
  }, [level, source, range]);

  useEffect(() => {
    if (!autoRefresh) return;
    const timer = setInterval(() => {
      fetchLogs();
      fetchStats();
    }, refreshInterval);
    return () => clearInterval(timer);
  }, [autoRefresh, refreshInterval, level, source, range]);

  const columns = [
    { title: '时间', dataIndex: 'timestamp', render: (_: any, r: LogItem) => dayjs(r.timestamp || r.time).format('YYYY-MM-DD HH:mm:ss') },
    { title: '级别', dataIndex: 'level', render: (v: string) => {
      const color = v === 'ERROR' ? 'red' : v === 'WARN' ? 'orange' : v === 'DEBUG' ? 'geekblue' : 'green';
      return <Tag color={color}>{v || 'INFO'}</Tag>;
    } },
    { title: '来源', dataIndex: 'source', render: (v: string) => v || '-' },
    { title: '内容', dataIndex: 'message' },
    { title: '用户', dataIndex: 'user_id', render: (v: string) => v || '-' },
    { title: 'IP', dataIndex: 'ip', render: (v: string) => v || '-' },
  ];

  return (
    <div style={{ padding: 16 }}>
      <Space align="center" style={{ marginBottom: 12 }}>
        <FileSearchOutlined />
        <Title level={4} style={{ margin: 0 }}>日志管理</Title>
      </Space>

      <Card style={{ marginBottom: 16 }}>
        <Space wrap>
          <Select
            placeholder="选择级别"
            value={level}
            allowClear
            onChange={setLevel}
            options={[
              { value: 'DEBUG', label: 'DEBUG' },
              { value: 'INFO', label: 'INFO' },
              { value: 'WARN', label: 'WARN' },
              { value: 'ERROR', label: 'ERROR' },
            ]}
            style={{ width: 160 }}
          />
          <Select
            placeholder="选择来源"
            value={source}
            allowClear
            onChange={setSource}
            options={[
              { value: 'frontend', label: '前端' },
              { value: 'backend', label: '后端' },
              { value: 'database', label: '数据库' },
              { value: 'security', label: '安全' },
              { value: 'system', label: '系统' },
            ]}
            style={{ width: 180 }}
          />
          <RangePicker showTime value={range as any} onChange={(v) => setRange(v as any)} style={{ minWidth: 280 }} />
          <Input placeholder="关键词过滤" value={keyword} onChange={(e) => setKeyword(e.target.value)} style={{ width: 220 }} />
          <Input placeholder="用户ID过滤" value={userIdFilter} onChange={(e) => setUserIdFilter(e.target.value)} style={{ width: 180 }} />
          <Input placeholder="IP过滤" value={ipFilter} onChange={(e) => setIpFilter(e.target.value)} style={{ width: 180 }} />
          <Button icon={<ReloadOutlined />} onClick={() => { fetchLogs(); fetchStats(); }}>刷新</Button>
          <Divider type="vertical" />
          <Space>
            <Text>自动刷新</Text>
            <Switch checked={autoRefresh} onChange={setAutoRefresh} />
            <Select
              value={refreshInterval}
              onChange={setRefreshInterval}
              options={[
                { value: 5000, label: '5秒' },
                { value: 10000, label: '10秒' },
                { value: 30000, label: '30秒' },
                { value: 60000, label: '60秒' },
              ]}
              style={{ width: 120 }}
              disabled={!autoRefresh}
            />
          </Space>
          <Tag color="geekblue">最后更新：{lastUpdatedAt || '-'}</Tag>
        </Space>
      </Card>

      <Row gutter={12} style={{ marginBottom: 12 }}>
        <Col xs={24} md={12} lg={6}><Card loading={statsLoading}><Statistic title="ERROR" value={stats.ERROR || 0} valueStyle={{ color: '#cf1322' }} /></Card></Col>
        <Col xs={24} md={12} lg={6}><Card loading={statsLoading}><Statistic title="WARN" value={stats.WARN || 0} valueStyle={{ color: '#fa8c16' }} /></Card></Col>
        <Col xs={24} md={12} lg={6}><Card loading={statsLoading}><Statistic title="INFO" value={stats.INFO || 0} /></Card></Col>
        <Col xs={24} md={12} lg={6}><Card loading={statsLoading}><Statistic title="DEBUG" value={stats.DEBUG || 0} /></Card></Col>
      </Row>

      <Card>
        <Table
          size="small"
          rowKey={(_, idx) => String(idx)}
          loading={loading}
          dataSource={filtered}
          pagination={{ pageSize: 10 }}
          columns={columns as any}
          expandable={{
            expandedRowRender: (record: LogItem) => (
              <div style={{ padding: 8 }}>
                <Space direction="vertical" style={{ width: '100%' }}>
                  <Text type="secondary">User Agent：{record.user_agent || '-'}</Text>
                  <Divider style={{ margin: '8px 0' }} />
                  <Text strong>额外上下文</Text>
                  <pre style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word', background: '#f6f8fa', padding: 8, borderRadius: 4 }}>
                    {JSON.stringify(record.extra ?? {}, null, 2)}
                  </pre>
                </Space>
              </div>
            ),
          }}
        />
      </Card>
    </div>
  );
};

export default LogsManagement;