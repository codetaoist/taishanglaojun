import React, { useEffect, useMemo, useState } from 'react';
import { Card, Tabs, Table, Input, Button, Space, Tag, Typography, Divider, Select, message, InputNumber } from 'antd';
import { DatabaseOutlined, SearchOutlined, PlayCircleOutlined, ReloadOutlined, SettingOutlined, MonitorOutlined } from '@ant-design/icons';
import { apiClient } from '../../services/api';
import DatabaseConnectionManager from '../../components/database/DatabaseConnectionManager';
import ConnectionStatusMonitor from '../../components/database/ConnectionStatusMonitor';
import EnhancedDatabaseMonitor from '../../components/database/EnhancedDatabaseMonitor';

const { Title, Text } = Typography;

interface ColumnDef {
  name: string;
  type?: string;
  nullable?: boolean;
}

const DatabaseManagement: React.FC = () => {
  const [tables, setTables] = useState<string[]>([]);
  const [tablesLoading, setTablesLoading] = useState(false);
  const [tableFilter, setTableFilter] = useState('');
  const [selectedTable, setSelectedTable] = useState<string | null>(null);
  const [columns, setColumns] = useState<ColumnDef[]>([]);
  const [columnsLoading, setColumnsLoading] = useState(false);
  const [columnFilter, setColumnFilter] = useState('');

  const [rowCount, setRowCount] = useState<number | null>(null);
  const [rowCountLoading, setRowCountLoading] = useState(false);
  const [previewColumns, setPreviewColumns] = useState<string[]>([]);
  const [previewRows, setPreviewRows] = useState<Array<Record<string, any>>>([]);
  const [previewLoading, setPreviewLoading] = useState(false);

  const [sql, setSql] = useState<string>('SELECT NOW() as current_time');
  const [queryLoading, setQueryLoading] = useState(false);
  const [queryColumns, setQueryColumns] = useState<string[]>([]);
  const [queryRows, setQueryRows] = useState<Array<Record<string, any>>>([]);
  const [sqlTemplate, setSqlTemplate] = useState<string | undefined>(undefined);
  const [queryMaxRows, setQueryMaxRows] = useState<number>(200);

  const [engine, setEngine] = useState<string>('primary-db');
  const [dbStats, setDbStats] = useState<Record<string, any> | null>(null);
  const [statsLoading, setStatsLoading] = useState(false);
  const [optimizeLoading, setOptimizeLoading] = useState(false);

  // schema 与表分页
  const [selectedSchema, setSelectedSchema] = useState<string | undefined>('public');
  const [schemas, setSchemas] = useState<string[]>([]);
  const [schemasLoading, setSchemasLoading] = useState(false);
  const [schemasPage, setSchemasPage] = useState<number>(1);
  const [schemasLimit, setSchemasLimit] = useState<number>(100);
  const [schemasTotal, setSchemasTotal] = useState<number>(0);
  const [schemasPages, setSchemasPages] = useState<number>(1);
  const [tablesPage, setTablesPage] = useState<number>(1);
  const [tablesLimit, setTablesLimit] = useState<number>(50);
  const [tablesTotal, setTablesTotal] = useState<number>(0);
  const [tablesPages, setTablesPages] = useState<number>(1);

  // 备份与恢复状态
  const [backupName, setBackupName] = useState('');
  const [backupDesc, setBackupDesc] = useState('');
  const [backupCreating, setBackupCreating] = useState(false);
  const [backups, setBackups] = useState<Array<Record<string, any>>>([]);
  const [backupsLoading, setBackupsLoading] = useState(false);
  const [backupPage, setBackupPage] = useState(1);
  const [backupLimit, setBackupLimit] = useState(10);
  const [backupTotal, setBackupTotal] = useState(0);

  // 辅助格式化函数
  const formatBytes = (n: any) => {
    const val = typeof n === 'number' ? n : Number(n);
    if (!val || isNaN(val)) return '-';
    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    let u = 0; let v = val;
    while (v >= 1024 && u < units.length - 1) { v /= 1024; u++; }
    return `${v.toFixed(u === 0 ? 0 : 1)} ${units[u]}`;
  };

  const formatDateTime = (s: any) => {
    const d = s ? new Date(s) : null;
    if (!d || isNaN(d.getTime())) return '-';
    const pad = (x: number) => (x < 10 ? `0${x}` : `${x}`);
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
  };

  const renderStatusTag = (st?: string) => {
    const color = st === 'completed' ? 'green' : st === 'failed' ? 'red' : 'orange';
    return <Tag color={color}>{st || '-'}</Tag>;
  };

  const filteredTables = useMemo(() => {
    if (!tableFilter) return tables;
    return tables.filter(t => t.toLowerCase().includes(tableFilter.toLowerCase()));
  }, [tables, tableFilter]);

  const loadTables = async (page = 1, limit = tablesLimit) => {
    try {
      setTablesLoading(true);
      const res = await apiClient.listDatabaseTablesPaged({ page, limit, schema: selectedSchema?.trim() || undefined });
      const payload = res.data || { items: [], total: 0, page, limit, pages: 1 };
      setTables(payload.items || []);
      setTablesTotal(payload.total || 0);
      setTablesPage(payload.page || page);
      setTablesLimit(payload.limit || limit);
      setTablesPages(payload.pages || Math.ceil((payload.total || 0) / Math.max(payload.limit || limit, 1)));
      if (!selectedTable && (payload.items || []).length > 0) {
        setSelectedTable(payload.items[0]);
      }
    } catch (err) {
      message.error('获取数据表列表失败');
    } finally {
      setTablesLoading(false);
    }
  };

  const loadMoreTables = async () => {
    const nextPage = tablesPage + 1;
    try {
      setTablesLoading(true);
      const res = await apiClient.listDatabaseTablesPaged({ page: nextPage, limit: tablesLimit, schema: selectedSchema?.trim() || undefined });
      const payload = res.data || { items: [], total: tablesTotal, page: nextPage, limit: tablesLimit, pages: tablesPages };
      const newItems = payload.items || [];
      setTables(prev => Array.from(new Set([...prev, ...newItems])));
      setTablesTotal(payload.total || tablesTotal);
      setTablesPage(payload.page || nextPage);
      setTablesLimit(payload.limit || tablesLimit);
      setTablesPages(payload.pages || tablesPages);
    } catch (err) {
      message.error('加载更多数据表失败');
    } finally {
      setTablesLoading(false);
    }
  };

  const loadSchemas = async (page = 1, limit = schemasLimit) => {
    try {
      setSchemasLoading(true);
      const res = await apiClient.listDatabaseSchemasPaged({ page, limit });
      const payload = res.data || { items: [], total: 0, page, limit, pages: 1 };
      setSchemas(payload.items || []);
      setSchemasTotal(payload.total || 0);
      setSchemasPage(payload.page || page);
      setSchemasLimit(payload.limit || limit);
      setSchemasPages(payload.pages || Math.ceil((payload.total || 0) / Math.max(payload.limit || limit, 1)));
      // 如果当前未选择 schema，则尝试选择 public 或首个
      if (!selectedSchema) {
        const items = payload.items || [];
        const defaultSchema = items.find((s: string) => s === 'public') || items[0];
        setSelectedSchema(defaultSchema);
      }
    } catch (err) {
      message.error('获取 schema 列表失败');
    } finally {
      setSchemasLoading(false);
    }
  };

  const loadMoreSchemas = async () => {
    const nextPage = schemasPage + 1;
    try {
      setSchemasLoading(true);
      const res = await apiClient.listDatabaseSchemasPaged({ page: nextPage, limit: schemasLimit });
      const payload = res.data || { items: [], total: schemasTotal, page: nextPage, limit: schemasLimit, pages: schemasPages };
      const newItems = payload.items || [];
      setSchemas(prev => Array.from(new Set([...prev, ...newItems])));
      setSchemasTotal(payload.total || schemasTotal);
      setSchemasPage(payload.page || nextPage);
      setSchemasLimit(payload.limit || schemasLimit);
      setSchemasPages(payload.pages || schemasPages);
    } catch (err) {
      message.error('加载更多 schema 失败');
    } finally {
      setSchemasLoading(false);
    }
  };

  const loadColumns = async (table?: string) => {
    const tbl = table || selectedTable;
    if (!tbl) return;
    try {
      setColumnsLoading(true);
      const res = await apiClient.getTableColumns(tbl);
      setColumns(res.data || []);
    } catch (err) {
      message.error('获取表列信息失败');
    } finally {
      setColumnsLoading(false);
    }
  };

  const loadRowCount = async (table?: string) => {
    const tbl = table || selectedTable;
    if (!tbl) return;
    const candidates: string[] = (() => {
      const base = `SELECT COUNT(*) AS count FROM ${tbl}`;
      const variants: string[] = [base];
      if (tbl.includes('.')) {
        const parts = tbl.split('.');
        const schema = parts[0];
        const name = parts.slice(1).join('.');
        variants.push(`SELECT COUNT(*) AS count FROM "${schema}"."${name}"`);
        variants.push(`SELECT COUNT(*) AS count FROM \`${schema}\`.\`${name}\``);
      }
      return Array.from(new Set(variants));
    })();

    try {
      setRowCountLoading(true);
      let lastError: any = null;
      for (const sql of candidates) {
        try {
          const res = await apiClient.runReadOnlyQuery(sql);
          const rows = res.data.rows || [];
          if (rows.length > 0) {
            const first = rows[0];
            const countValue = first.count ?? Object.values(first)[0];
            setRowCount(typeof countValue === 'number' ? countValue : Number(countValue));
            lastError = null;
            break;
          } else {
            setRowCount(0);
            lastError = null;
            break;
          }
        } catch (e) {
          lastError = e;
          continue;
        }
      }
      if (lastError) {
        throw lastError;
      }
    } catch (err) {
      message.error('获取行数统计失败');
      setRowCount(null);
    } finally {
      setRowCountLoading(false);
    }
  };

  const loadPreview = async (table?: string, limit = 100) => {
    const tbl = table || selectedTable;
    if (!tbl) return;
    const candidates: string[] = (() => {
      const base = `SELECT * FROM ${tbl} LIMIT ${limit}`;
      const variants: string[] = [base];
      if (tbl.includes('.')) {
        const parts = tbl.split('.');
        const schema = parts[0];
        const name = parts.slice(1).join('.');
        variants.push(`SELECT * FROM "${schema}"."${name}" LIMIT ${limit}`);
        variants.push(`SELECT * FROM \`${schema}\`.\`${name}\` LIMIT ${limit}`);
      }
      return Array.from(new Set(variants));
    })();

    try {
      setPreviewLoading(true);
      let lastError: any = null;
      for (const sql of candidates) {
        try {
          const res = await apiClient.runReadOnlyQuery(sql);
          const cols = res.data.columns || [];
          const rows: Array<Record<string, any>> = (res.data.rows || []).map((r: Record<string, any>, i: number) => ({ ...r, __key: `${tbl}-${i}` }));
          setPreviewColumns(cols);
          setPreviewRows(rows);
          lastError = null;
          break;
        } catch (e) {
          lastError = e;
          continue;
        }
      }
      if (lastError) {
        throw lastError;
      }
    } catch (err) {
      message.error('获取数据预览失败');
      setPreviewColumns([]);
      setPreviewRows([]);
    } finally {
      setPreviewLoading(false);
    }
  };

  const runQuery = async () => {
    if (!sql.trim()) {
      message.warning('请输入SQL语句');
      return;
    }
    setQueryLoading(true);
    try {
      const res = await apiClient.runReadOnlyQuery(sql, { maxRows: queryMaxRows });
      const cols = res.data.columns || [];
      const rows: Array<Record<string, any>> = (res.data.rows || []).map((r: Record<string, any>, i: number) => ({ ...r, __key: `query-${i}` }));
      setQueryColumns(cols);
      setQueryRows(rows);
    } catch (err) {
      message.error('查询执行失败');
    } finally {
      setQueryLoading(false);
    }
  };

  const exportQueryCSV = () => {
    const cols = queryColumns.length > 0 ? queryColumns : (queryRows[0] ? Object.keys(queryRows[0]).filter(k => k !== '__key') : []);
    if (cols.length === 0 || queryRows.length === 0) {
      message.warning('无可导出的数据');
      return;
    }
    const escape = (val: any) => {
      const s = val === null || val === undefined ? '' : String(val);
      return '"' + s.replace(/"/g, '""') + '"';
    };
    const header = cols.map(escape).join(',');
    const lines = queryRows.map(row => cols.map(col => escape(row[col])).join(','));
    const csv = [header, ...lines].join('\n');
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `query_result_${Date.now()}.csv`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const loadDatabaseStats = async () => {
    try {
      setStatsLoading(true);
      const res = await apiClient.getDatabaseStats();
      setDbStats(res.data || null);
    } catch (err) {
      message.error('获取数据库统计失败');
    } finally {
      setStatsLoading(false);
    }
  };

  const triggerOptimize = async () => {
    try {
      setOptimizeLoading(true);
      const res = await apiClient.optimizeDatabase();
      const msg = (res as any)?.message || (res as any)?.data?.message || '数据库优化已启动';
      message.success(msg);
    } catch (err) {
      message.error('触发数据库优化失败');
    } finally {
      setOptimizeLoading(false);
    }
  };

  const loadBackups = async (page = backupPage, limit = backupLimit) => {
    try {
      setBackupsLoading(true);
      const res = await apiClient.getBackups({ page, limit });
      const payload = (res as any)?.data || (res as any) || {};
      const list = payload.backups || payload.items || [];
      setBackups(Array.isArray(list) ? list : []);
      setBackupTotal(payload.total ?? (Array.isArray(list) ? list.length : 0));
      setBackupPage(payload.page ?? page);
      setBackupLimit(payload.limit ?? limit);
    } catch (err) {
      message.error('获取备份列表失败');
    } finally {
      setBackupsLoading(false);
    }
  };

  const doCreateBackup = async () => {
    if (!backupName.trim()) {
      message.warning('请输入备份名称');
      return;
    }
    try {
      setBackupCreating(true);
      const res = await apiClient.createBackup(backupName.trim(), backupDesc.trim() || undefined);
      const msg = (res as any)?.message || (res as any)?.data?.message || '备份创建成功';
      message.success(msg);
      setBackupName('');
      setBackupDesc('');
      await loadBackups(1, backupLimit);
    } catch (err) {
      message.error('创建备份失败');
    } finally {
      setBackupCreating(false);
    }
  };

  const doRestoreBackup = async (id: number | string) => {
    try {
      const res = await apiClient.restoreBackup(id);
      const msg = (res as any)?.message || (res as any)?.data?.message || '已触发恢复';
      message.success(msg);
      await loadBackups(backupPage, backupLimit);
    } catch (err) {
      message.error('恢复备份失败');
    }
  };

  useEffect(() => {
    loadSchemas();
    loadTables();
  }, []);

  useEffect(() => {
    loadDatabaseStats();
  }, []);

  useEffect(() => {
    loadBackups();
  }, []);

  useEffect(() => {
    if (selectedTable) {
      loadColumns(selectedTable);
      loadRowCount(selectedTable);
      loadPreview(selectedTable);
    }
  }, [selectedTable]);

  useEffect(() => {
    // schema 切换时重载表列表并重置选中表
    setTablesPage(1);
    setSelectedTable(null);
    loadTables(1, tablesLimit);
  }, [selectedSchema]);

  return (
    <div style={{ padding: 16 }}>
      <Space align="center" style={{ marginBottom: 12 }}>
        <DatabaseOutlined />
        <Title level={4} style={{ margin: 0 }}>数据库管理</Title>
      </Space>

      <Card style={{ marginBottom: 16 }}>
        <Space size={12} wrap>
          <Text>当前引擎：</Text>
          <Select
            value={engine}
            onChange={setEngine}
            options={[
              { value: 'primary-db', label: '主数据库（SQL）' },
              { value: 'mysql', label: 'MySQL（待配置）', disabled: true },
              { value: 'postgresql', label: 'PostgreSQL（待配置）', disabled: true },
              { value: 'sqlite', label: 'SQLite（待配置）', disabled: true },
              { value: 'mongodb', label: 'MongoDB（待配置）', disabled: true },
              { value: 'redis', label: 'Redis（待配置）', disabled: true },
            ]}
            style={{ minWidth: 240 }}
          />
          <Select
            value={selectedSchema}
            onChange={(val) => setSelectedSchema(val)}
            loading={schemasLoading}
            options={(schemas || []).map(s => ({ value: s, label: s }))}
            showSearch
            allowClear
            placeholder="选择 schema（默认 public）"
            style={{ minWidth: 240 }}
            filterOption={(input, option) => (option?.label as string).toLowerCase().includes(input.toLowerCase())}
            dropdownRender={(menu) => (
              <div>
                {menu}
                <Divider style={{ margin: '8px 0' }} />
                <Space style={{ padding: '0 8px 4px' }}>
                  <Button size="small" onClick={loadMoreSchemas} disabled={schemasPage >= schemasPages || schemasLoading}>加载更多</Button>
                  <Text type="secondary">共 {schemasTotal} 个</Text>
                </Space>
              </div>
            )}
          />
          <Tag color="blue">连接状态：正常</Tag>
          <Button icon={<ReloadOutlined />} onClick={() => { loadSchemas(1, schemasLimit); setTablesPage(1); loadTables(1, tablesLimit); if (selectedTable) loadColumns(selectedTable); }}>
            刷新
          </Button>
        </Space>
      </Card>

      <Tabs
        items={[
          {
            key: 'connections',
            label: (
              <Space>
                <SettingOutlined />
                连接管理
              </Space>
            ),
            children: <DatabaseConnectionManager />
          },
          {
            key: 'tables',
            label: '数据表',
            children: (
              <Space align="start" style={{ width: '100%' }}>
                <Card style={{ width: 340 }} styles={{ body: { padding: 12 } }} loading={tablesLoading}>
                  <Input
                    placeholder="搜索数据表"
                    prefix={<SearchOutlined />}
                    value={tableFilter}
                    onChange={(e) => setTableFilter(e.target.value)}
                    style={{ marginBottom: 8 }}
                  />
                  <div style={{ maxHeight: 420, overflow: 'auto' }}>
                    {(filteredTables || []).map((t) => (
                      <div
                        key={t}
                        onClick={() => setSelectedTable(t)}
                        style={{
                          padding: '8px 10px',
                          cursor: 'pointer',
                          borderRadius: 6,
                          background: selectedTable === t ? 'rgba(24,144,255,0.1)' : 'transparent'
                        }}
                      >
                        <Text strong={selectedTable === t}>{t}</Text>
                      </div>
                    ))}
                  </div>
                  <Divider style={{ margin: '12px 0' }} />
                  <Space wrap>
                    <Text type="secondary">共 {tablesTotal} 条，已加载 {tables.length} 条</Text>
                    <Button size="small" onClick={loadMoreTables} disabled={tablesPage >= tablesPages || tablesLoading}>加载更多</Button>
                    <Button size="small" onClick={() => { setTablesPage(1); loadTables(1, tablesLimit); }} disabled={tablesLoading}>重新加载</Button>
                  </Space>
                </Card>
                <Card
                  style={{ flex: 1 }}
                  title={<Space><Text>表结构</Text><Tag color="geekblue">{selectedTable || '未选择'}</Tag></Space>}
                  loading={columnsLoading}
                  extra={
                    <Space>
                      <Input
                        allowClear
                        size="small"
                        placeholder="搜索列名"
                        prefix={<SearchOutlined />}
                        value={columnFilter}
                        onChange={(e) => setColumnFilter(e.target.value)}
                        style={{ width: 180 }}
                      />
                      <Button size="small" icon={<ReloadOutlined />} onClick={() => { if (selectedTable) { loadColumns(selectedTable); } }}>刷新结构</Button>
                      <Tag color="blue">总行数：{rowCountLoading ? '加载中...' : (rowCount ?? '-')}</Tag>
                    </Space>
                  }
                >
                  <Table
                    size="small"
                    rowKey={(r) => `${r.name}`}
                    dataSource={columns.filter(c => !columnFilter || c.name.toLowerCase().includes(columnFilter.toLowerCase()))}
                    pagination={{ pageSize: 10 }}
                    columns={[
                      { title: '列名', dataIndex: 'name' },
                      { title: '类型', dataIndex: 'type', render: (v) => v || '-' },
                      { title: '可空', dataIndex: 'nullable', render: (v: boolean) => (v ? '是' : '否') },
                    ]}
                  />
                </Card>

                <Card style={{ flex: 1 }} title={<Space><Text>数据预览</Text><Tag color="geekblue">{selectedTable || '未选择'}</Tag></Space>} extra={<Button size="small" icon={<ReloadOutlined />} onClick={() => { if (selectedTable) { loadPreview(selectedTable); } }}>刷新预览</Button>}>
                  <Table
                    size="small"
                    rowKey="__key"
                    loading={previewLoading}
                    dataSource={previewRows}
                    pagination={{ pageSize: 10 }}
                    columns={(previewColumns || []).map((c) => ({ title: c, dataIndex: c }))}
                  />
                </Card>
              </Space>
            )
          },
          {
            key: 'stats',
            label: '统计与优化',
            children: (
              <Space align="start" style={{ width: '100%' }}>
                <Card
                  style={{ flex: 1 }}
                  title={<Space><Text>数据库统计</Text></Space>}
                  extra={
                    <Space>
                      <Button size="small" icon={<ReloadOutlined />} onClick={loadDatabaseStats} loading={statsLoading}>刷新统计</Button>
                      <Tag color={optimizeLoading ? 'orange' : 'green'}>{optimizeLoading ? '优化进行中...' : '优化状态：空闲'}</Tag>
                    </Space>
                  }
                >
                  <Table
                    size="small"
                    rowKey={(r) => `${r.metric}`}
                    loading={statsLoading}
                    dataSource={Object.entries(dbStats || {}).map(([k, v]) => ({ metric: k, value: String(v ?? '') }))}
                    pagination={false}
                    columns={[
                      { title: '指标', dataIndex: 'metric' },
                      { title: '值', dataIndex: 'value' },
                    ]}
                  />
                </Card>
                <Card style={{ width: 340 }} title={<Space><Text>优化操作</Text></Space>}>
                  <Space direction="vertical" size={12} style={{ width: '100%' }}>
                    <Button type="primary" icon={<PlayCircleOutlined />} onClick={triggerOptimize} loading={optimizeLoading}>优化数据库</Button>
                    <Text type="secondary">优化在后台进行，完成后自动恢复空闲状态。</Text>
                  </Space>
                </Card>
              </Space>
            )
          },
          {
            key: 'backup',
            label: '备份与恢复',
            children: (
              <Space align="start" style={{ width: '100%' }}>
                <Card style={{ width: 360 }} title={<Space><Text>创建备份</Text></Space>}>
                  <Space direction="vertical" size={8} style={{ width: '100%' }}>
                    <Input
                      value={backupName}
                      onChange={(e) => setBackupName(e.target.value)}
                      placeholder="备份名称（必填）"
                    />
                    <Input.TextArea
                      value={backupDesc}
                      onChange={(e) => setBackupDesc(e.target.value)}
                      placeholder="备份描述（可选）"
                      autoSize={{ minRows: 3, maxRows: 6 }}
                    />
                    <Space>
                      <Button type="primary" icon={<PlayCircleOutlined />} loading={backupCreating} onClick={doCreateBackup}>创建备份</Button>
                      <Button onClick={() => { setBackupName(''); setBackupDesc(''); }}>清空</Button>
                    </Space>
                    <Text type="secondary">创建后可在列表中查看并执行恢复操作。</Text>
                  </Space>
                </Card>
                <Card style={{ flex: 1 }} title={<Space><Text>备份列表</Text></Space>} extra={<Space><Button size="small" icon={<ReloadOutlined />} onClick={() => loadBackups(backupPage, backupLimit)} loading={backupsLoading}>刷新列表</Button><Tag color="blue">总数：{backupTotal}</Tag></Space>}>
                  <Table
                    size="small"
                    rowKey={(r) => `${(r as any).id ?? (r as any).name ?? Math.random()}`}
                    loading={backupsLoading}
                    dataSource={backups}
                    pagination={{
                      current: backupPage,
                      pageSize: backupLimit,
                      total: backupTotal,
                      showSizeChanger: true,
                      onChange: (page, pageSize) => { setBackupPage(page); setBackupLimit(pageSize); loadBackups(page, pageSize); }
                    }}
                    columns={[
                      { title: '名称', dataIndex: 'name' },
                      { title: '描述', dataIndex: 'description', render: (v: any) => v || '-' },
                      { title: '状态', dataIndex: 'status', render: (v: any) => renderStatusTag(v) },
                      { title: '文件', dataIndex: 'file_path', render: (v: any) => v || '-' },
                      { title: '大小', dataIndex: 'file_size', render: (v: any) => formatBytes(v) },
                      { title: '创建者', dataIndex: 'created_by', render: (v: any) => v || '-' },
                      { title: '创建时间', dataIndex: 'created_at', render: (v: any) => formatDateTime(v) },
                      { title: '完成时间', dataIndex: 'completed_at', render: (v: any) => formatDateTime(v) },
                      { title: '操作', key: 'action', render: (_: any, r: any) => (
                        <Space>
                          <Button type="link" onClick={() => doRestoreBackup(r.id)} disabled={String(r.status) !== 'completed'}>恢复</Button>
                        </Space>
                      ) },
                    ]}
                  />
                </Card>
              </Space>
            )
          },
          {
            key: 'query',
            label: '只读查询',
            children: (
              <Card title="执行只读SQL查询">
                <Text type="secondary">注意：仅允许 SELECT 等只读语句</Text>
                <Divider style={{ margin: '12px 0' }} />
                <Space style={{ marginBottom: 8 }}>
                  <Select
                    value={sqlTemplate}
                    onChange={(val) => {
                      setSqlTemplate(val);
                      const tbl = selectedTable || 'your_table';
                      const mapped: Record<string, string> = {
                        'select-all': `SELECT * FROM ${tbl} LIMIT 10`,
                        'count-all': `SELECT COUNT(*) AS count FROM ${tbl}`,
                        'current-time': 'SELECT NOW() AS current_time'
                      };
                      setSql(mapped[val] || '');
                    }}
                    placeholder="选择SQL模板"
                    style={{ minWidth: 220 }}
                    options={[
                      { value: 'select-all', label: '模板：查询前10行（当前选择表）' },
                      { value: 'count-all', label: '模板：统计总行数（当前选择表）' },
                      { value: 'current-time', label: '模板：当前时间' },
                    ]}
                  />
                  <Button onClick={() => { setSqlTemplate(undefined); setSql(''); }}>清空输入</Button>
                </Space>
                {/* SQL 引用提示（兼容大小写/特殊字符） */}
                {selectedTable && (
                  <Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
                    引用建议：PostgreSQL 使用 {(() => { const parts = selectedTable.split('.'); return `${parts[0]}."${parts.slice(1).join('.')}"`; })()}；MySQL 使用 {selectedTable}
                  </Text>
                )}
                <Input.TextArea
                  value={sql}
                  onChange={(e) => setSql(e.target.value)}
                  placeholder="输入只读SQL，例如：SELECT * FROM your_table LIMIT 10"
                  autoSize={{ minRows: 4, maxRows: 10 }}
                  style={{ marginBottom: 8 }}
                />
                <Space>
                  <Button type="primary" icon={<PlayCircleOutlined />} loading={queryLoading} onClick={runQuery}>执行</Button>
                  <InputNumber min={1} max={1000} step={50} value={queryMaxRows} onChange={(v) => setQueryMaxRows(Number(v) || 200)} addonBefore="最大行数" style={{ width: 180 }} />
                  <Button onClick={exportQueryCSV} disabled={queryRows.length === 0}>导出CSV</Button>
                  <Button onClick={() => { setQueryRows([]); setQueryColumns([]); }}>清空结果</Button>
                </Space>
                <Divider style={{ margin: '12px 0' }} />
                <Table
                  size="small"
                  rowKey="__key"
                  dataSource={queryRows}
                  pagination={{ pageSize: 10 }}
                  columns={(queryColumns || []).map((c) => ({ title: c, dataIndex: c }))}
                />
              </Card>
            )
          },
          {
            key: 'monitor',
            label: (
              <Space>
                <MonitorOutlined />
                状态监控
              </Space>
            ),
            children: <ConnectionStatusMonitor />
          },
          {
            key: 'enhanced-monitor',
            label: (
              <Space>
                <MonitorOutlined />
                增强监控
              </Space>
            ),
            children: <EnhancedDatabaseMonitor />
          }
        ]}
      />
    </div>
  );
};

export default DatabaseManagement;