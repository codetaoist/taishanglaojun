import React, { useState, useEffect, useCallback, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Card,
  Table,
  Button,
  Space,
  Modal,
  Form,
  Input,
  Select,
  Tree,
  Switch,
  Popconfirm,
  Tag,
  Tooltip,
  Row,
  Col,
  InputNumber,
  TreeSelect,
  Spin,
  App
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  ReloadOutlined,
  SearchOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  HomeOutlined,
  DashboardOutlined,
  SettingOutlined,
  MenuOutlined,
  ProjectOutlined,
  DesktopOutlined
} from '@ant-design/icons';
import {
  FolderOutlined,
  FileOutlined,
  BarsOutlined,
  ApiOutlined
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import type { TreeDataNode } from 'antd/es/tree';
import { menuService } from '../../services/menuService';
import { PERMISSIONS } from '../../hooks/usePermissions';
import type { MenuItem, CreateMenuRequest, UpdateMenuRequest, MenuPermission } from '../../services/menuService';
import { testPermissionAPI } from '../../utils/testPermissionAPI';
import { getIconNode, defaultIconName, isValidIconName } from '../../ui/icons/iconRegistry';
import IconSelectorModal from '../../components/common/IconSelectorModal';
import { formatDateTime, getStatusColor } from '../../utils/display';

const { Option } = Select;
const { Search } = Input;

interface MenuFormData {
  name: string;
  path: string;
  icon?: string;
  component?: string;
  parentId?: string;
  sort: number;
  status: 'active' | 'inactive';
  type: 'menu' | 'button' | 'api';
  permissions: string[];
}

const MenuManagement: React.FC = () => {
  const { message } = App.useApp();
  const { t } = useTranslation();
  const [iconModalOpen, setIconModalOpen] = useState(false);
  const [menus, setMenus] = useState<MenuItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingMenu, setEditingMenu] = useState<MenuItem | null>(null);
  const [form] = Form.useForm<MenuFormData>();
  const [searchText, setSearchText] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [typeFilter, setTypeFilter] = useState<string>('');
  const [permissions, setPermissions] = useState<MenuPermission[]>([]);
  const [treeData, setTreeData] = useState<TreeDataNode[]>([]);
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);
  const [selectedKeys, setSelectedKeys] = useState<React.Key[]>([]);
  const [selectedMenu, setSelectedMenu] = useState<MenuItem | null>(null);
  const [menuTreeItems, setMenuTreeItems] = useState<MenuItem[]>([]);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0
  });
  const [collapsed, setCollapsed] = useState(false);

  // 加载菜单数据
  const loadMenus = useCallback(async (page?: number, pageSize?: number) => {
    setLoading(true);
    try {
      const currentPage = page ?? pagination.current;
      const currentPageSize = pageSize ?? pagination.pageSize;
      
      const response = await menuService.getMenuList({
        page: currentPage,
        pageSize: currentPageSize,
        name: searchText || undefined,
        status: statusFilter || undefined,
        type: typeFilter || undefined
      });
      
      setMenus(response.items ?? []);
      setPagination(prev => ({
        ...prev,
        current: currentPage,
        pageSize: currentPageSize,
        total: response.total
      }));
      
      // 同时加载树形数据
      const treeResponse = await menuService.getMenuTree();
      setMenuTreeItems(treeResponse.items);
      setTreeData(convertToTreeData(treeResponse.items));
    } catch (error) {
      message.error(t('adminMenu.messages.loadMenusFailure'));
    } finally {
      setLoading(false);
    }
  }, [searchText, statusFilter, typeFilter]);

  // 加载权限数据
  const loadPermissions = async () => {
    try {
      console.log('🔄 开始加载权限数据...');
      const permissionList = await menuService.getAvailablePermissions();
      console.log('✅ 权限数据加载成功:', permissionList);
      setPermissions(permissionList ?? []);
    } catch (error) {
      console.error('❌ 权限数据加载失败:', error);
      console.error('错误详情:', {
        message: error instanceof Error ? error.message : '未知错误',
        response: (error as any)?.response?.data,
        status: (error as any)?.response?.status,
        statusText: (error as any)?.response?.statusText
      });
      
      // 回退到本地权限常量，避免页面不可用
      const fallbackPermissions: MenuPermission[] = Object.values(PERMISSIONS).map(code => ({
        id: code,
        name: code,
        code,
        description: ''
      }));
      setPermissions(fallbackPermissions);
      
      const errorMessage = error instanceof Error ? error.message : 'Unknown Error';
      message.error({ 
        content: t('adminMenu.messages.loadPermissionsFailureFallback', { error: errorMessage }), 
        key: 'load-permissions-error',
        duration: 8
      });
    }
  };

  // 转换为树形数据（兼容嵌套 children 与扁平 parentId 两种结构）
  const convertToTreeData = (items: MenuItem[] = []): TreeDataNode[] => {
    // 如果后端已返回嵌套 children 结构，则直接递归映射
    const hasChildrenStructure = items.some(item => Array.isArray(item.children));

    // 使用统一图标注册中心进行渲染

    const renderTypeIcon = (type: MenuItem['type']) => {
      switch (type) {
        case 'menu':
          return <FolderOutlined />;
        case 'button':
          return <BarsOutlined />;
        case 'api':
        default:
          return <ApiOutlined />;
      }
    };

    const renderItemIcon = (item: MenuItem) => {
      const node = getIconNode(item.icon || defaultIconName);
      return React.cloneElement(node as React.ReactElement, { style: { marginRight: collapsed ? 0 : 8 } });
    };

    const renderTitle = (item: MenuItem) => {
      const displayName = item.title || item.name; // 仅显示中文名称，避免重复
      if (collapsed) {
        return (
          <span className="flex items-center">
            {renderItemIcon(item)}
          </span>
        );
      }
      return (
        <span className="flex items-center py-1">
          {renderItemIcon(item)}
          <span className="font-medium">{displayName}</span>
        </span>
      );
    };

    const toNode = (item: MenuItem): TreeDataNode => ({
      key: item.id,
      title: renderTitle(item),
      children: Array.isArray(item.children) && item.children.length
        ? item.children.map(toNode)
        : undefined
    });

    if (hasChildrenStructure) {
      return items.map(toNode);
    }

    // 否则，按扁平结构以 parentId 归并树（兼容 null/undefined 作为根）
    const buildTree = (parentId?: string | null): TreeDataNode[] => {
      return items
        .filter(item => (item.parentId ?? null) === (parentId ?? null))
        .sort((a, b) => (b.sort ?? 0) - (a.sort ?? 0))
        .map(item => ({
          key: item.id,
          title: renderTitle(item),
          children: buildTree(item.id)
        }));
    };
    return buildTree(null);
  };

  // 根据选中节点，计算可见菜单（选中节点及其所有后代）
  const getDescendantIds = (items: MenuItem[], targetId: string): Set<string> => {
    const ids = new Set<string>();
    const dfs = (list: MenuItem[]) => {
      for (const it of list) {
        if (it.id === targetId) {
          const collect = (node: MenuItem) => {
            ids.add(node.id);
            if (Array.isArray(node.children)) {
              node.children.forEach(collect);
            }
          };
          collect(it);
        }
        if (Array.isArray(it.children) && it.children.length) {
          dfs(it.children);
        }
      }
    };
    dfs(items);
    return ids;
  };

  const visibleMenus = selectedMenu && menuTreeItems.length
    ? menus.filter(m => getDescendantIds(menuTreeItems, selectedMenu.id).has(m.id))
    : menus;

  // 初始化数据
  const didInitPermissionsRef = useRef(false);
  useEffect(() => {
    if (didInitPermissionsRef.current) return;
    didInitPermissionsRef.current = true;
    
    // 测试权限API调用
    testPermissionAPI().then(result => {
      console.log('🧪 权限API测试结果:', result);
    });
    
    loadPermissions();
  }, []);

  // 搜索和过滤变化时重新加载数据
  useEffect(() => {
    loadMenus(1); // 重置到第一页
  }, [loadMenus]);

  // 收缩状态变化时重建树数据以应用仅图标或中文名称渲染
  useEffect(() => {
    setTreeData(convertToTreeData(menuTreeItems));
  }, [collapsed, menuTreeItems]);

  // 表格列定义
  const columns: ColumnsType<MenuItem> = [
    {
      title: t('adminMenu.table.name'),
      dataIndex: 'name',
      key: 'name',
      render: (text, record) => (
        <div className="flex items-center">
          {record.icon && <span className="mr-2">{record.icon}</span>}
          {text}
        </div>
      )
    },
    {
      title: t('adminMenu.table.path'),
      dataIndex: 'path',
      key: 'path',
      render: (text) => <code className="bg-gray-100 px-2 py-1 rounded">{text}</code>
    },
    {
      title: t('adminMenu.table.type'),
      dataIndex: 'type',
      key: 'type',
      render: (type) => (
        <Tag color={type === 'menu' ? 'blue' : type === 'button' ? 'orange' : 'purple'}>
          {type === 'menu' ? t('adminMenu.filters.typeOptions.menu') : type === 'button' ? t('adminMenu.filters.typeOptions.button') : t('adminMenu.filters.typeOptions.api')}
        </Tag>
      )
    },
    {
      title: t('adminMenu.table.status'),
      dataIndex: 'status',
      key: 'status',
      render: (status) => (
        <Tag color={getStatusColor(status)}>
          {status === 'active' ? t('adminMenu.filters.statusOptions.enabled') : t('adminMenu.filters.statusOptions.disabled')}
        </Tag>
      )
    },
    {
      title: t('adminMenu.table.sort'),
      dataIndex: 'sort',
      key: 'sort',
      width: 80
    },
    {
      title: t('adminMenu.table.permissions'),
      dataIndex: 'permissions',
      key: 'permissions',
      render: (permissions: MenuPermission[]) => (
        <div>
          {permissions.slice(0, 2).map(permission => (
            <Tag key={permission.id} size="small">{permission.name}</Tag>
          ))}
          {permissions.length > 2 && (
            <Tooltip title={permissions.slice(2).map(p => p.name).join(', ')}>
              <Tag size="small">+{permissions.length - 2}</Tag>
            </Tooltip>
          )}
        </div>
      )
    },
    {
      title: t('adminMenu.table.actions'),
      key: 'action',
      width: 200,
      render: (_, record) => (
        <Space.Compact>
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => handleView(record)}
            size="small"
          >
            {t('adminMenu.actions.viewDetails')}
          </Button>
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
            size="small"
          >
            {t('adminMenu.actions.editMenu')}
          </Button>
          <Popconfirm
            title={t('adminMenu.confirm.deleteTitle')}
            onConfirm={() => handleDelete(record.id)}
            okText={t('adminMenu.confirm.ok')}
            cancelText={t('adminMenu.confirm.cancel')}
          >
            <Button
              type="text"
              icon={<DeleteOutlined />}
              danger
              size="small"
            >
              {t('adminMenu.actions.deleteMenu')}
            </Button>
          </Popconfirm>
        </Space.Compact>
      )
    }
  ];

  // 处理搜索
  const handleSearch = (value: string) => {
    setSearchText(value);
    setPagination(prev => ({ ...prev, current: 1 }));
  };

  // 处理筛选
  const handleStatusFilter = (value: string) => {
    setStatusFilter(value);
    setPagination(prev => ({ ...prev, current: 1 }));
  };

  const handleTypeFilter = (value: string) => {
    setTypeFilter(value);
    setPagination(prev => ({ ...prev, current: 1 }));
  };

  // 处理新增
  const handleAdd = () => {
    setEditingMenu(null);
    setModalVisible(true);
  };

  // 处理编辑
  const handleEdit = (menu: MenuItem) => {
    setEditingMenu(menu);
    setModalVisible(true);
  };

  // 处理查看
  const handleView = (menu: MenuItem) => {
    Modal.info({
      title: t('adminMenu.detail.title'),
      width: 600,
      content: (
        <div className="space-y-4">
          <div><strong>{t('adminMenu.detail.name')}:</strong> {menu.name}</div>
          <div><strong>{t('adminMenu.detail.path')}:</strong> <code>{menu.path}</code></div>
          <div><strong>{t('adminMenu.detail.icon')}:</strong> {menu.icon || t('common.none')}</div>
          <div><strong>{t('adminMenu.detail.component')}:</strong> {menu.component || t('common.none')}</div>
          <div><strong>{t('adminMenu.detail.type')}:</strong> {menu.type}</div>
          <div><strong>{t('adminMenu.detail.status')}:</strong> {menu.status === 'active' ? t('adminMenu.filters.statusOptions.enabled') : t('adminMenu.filters.statusOptions.disabled')}</div>
          <div><strong>{t('adminMenu.detail.sort')}:</strong> {menu.sort}</div>
          <div><strong>{t('adminMenu.detail.permissions')}:</strong></div>
          <div className="ml-4">
            {menu.permissions.map(permission => (
              <Tag key={permission.id} className="mb-1">
                {permission.name} ({permission.code})
              </Tag>
            ))}
          </div>
          <div><strong>{t('common.created')}:</strong> {formatDateTime(menu.createdAt)}</div>
          <div><strong>{t('common.updated')}:</strong> {formatDateTime(menu.updatedAt)}</div>
        </div>
      )
    });
  };

  // 处理删除
  const handleDelete = async (id: string) => {
    try {
      await menuService.deleteMenu(id);
      message.success(t('adminMenu.messages.deleteSuccess'));
      await loadMenus();
    } catch (error) {
      message.error(t('adminMenu.messages.deleteFailure'));
    }
  };

  // 处理表单提交
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      
      if (editingMenu) {
        // 更新菜单
        await menuService.updateMenu({
          id: editingMenu.id,
          ...values
        });
        message.success(t('adminMenu.messages.updateSuccess'));
      } else {
        // 创建菜单
        await menuService.createMenu(values);
        message.success(t('adminMenu.messages.createSuccess'));
      }
      
      setModalVisible(false);
      await loadMenus();
    } catch (error) {
      message.error(editingMenu ? t('adminMenu.messages.updateFailure') : t('adminMenu.messages.createFailure'));
    }
  };

  // 打开弹窗后再进行表单赋值，避免未挂载警告
  useEffect(() => {
    if (!modalVisible) return;
    if (editingMenu) {
      form.setFieldsValue({
        name: editingMenu.name,
        path: editingMenu.path,
        icon: editingMenu.icon,
        component: editingMenu.component,
        parentId: editingMenu.parentId,
        sort: editingMenu.sort,
        status: editingMenu.status,
        type: editingMenu.type,
        permissions: editingMenu.permissions.map(p => p.id)
      });
    } else {
      form.resetFields();
    }
  }, [modalVisible, editingMenu]);

  const openIconSelector = () => {
    setIconModalOpen(true);
  };

  // 获取父菜单选项
  const getParentMenuOptions = () => {
    const buildOptions = (items: MenuItem[] = [], level = 0): any[] => {
      return items
        .filter(item => item.type === 'menu')
        .map(item => ({
          value: item.id,
          title: (
            <span style={{ display: 'flex', alignItems: 'center', paddingLeft: level * 12 }}>
              {level > 0 && <span style={{ color: '#999', marginRight: 6 }}>{'└─'}</span>}
              <span style={{ marginRight: 6 }}>{getIconNode(item.icon || defaultIconName)}</span>
              <span>{item.name}</span>
            </span>
          ),
          children: Array.isArray(item.children) && item.children.length
            ? buildOptions(item.children, level + 1)
            : undefined
        }));
    };
    // 使用已构建的菜单树，确保层级一致
    return buildOptions(menuTreeItems);
  };

  return (
    <div className="p-6">
      <Card>
        <div className="mb-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">{t('adminMenu.title')}</h2>
            <Space>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={handleAdd}
              >
                {t('adminMenu.actions.addMenu')}
              </Button>
              <Button
                icon={<ReloadOutlined />}
                onClick={() => loadMenus()}
              >
                {t('adminMenu.actions.refresh')}
              </Button>
            </Space>
          </div>

          {/* 搜索和筛选 */}
          <Row gutter={16} className="mb-4">
            <Col span={8}>
              <Search
                placeholder={t('adminMenu.search.placeholder')}
                allowClear
                onSearch={handleSearch}
                style={{ width: '100%' }}
              />
            </Col>
            <Col span={4}>
              <Select
                placeholder={t('adminMenu.filters.statusPlaceholder')}
                allowClear
                onChange={handleStatusFilter}
                style={{ width: '100%' }}
              >
                <Option value="active">{t('adminMenu.filters.statusOptions.enabled')}</Option>
                <Option value="inactive">{t('adminMenu.filters.statusOptions.disabled')}</Option>
              </Select>
            </Col>
            <Col span={4}>
              <Select
                placeholder={t('adminMenu.filters.typePlaceholder')}
                allowClear
                onChange={handleTypeFilter}
                style={{ width: '100%' }}
              >
                <Option value="menu">{t('adminMenu.filters.typeOptions.menu')}</Option>
                <Option value="button">{t('adminMenu.filters.typeOptions.button')}</Option>
                <Option value="api">{t('adminMenu.filters.typeOptions.api')}</Option>
              </Select>
            </Col>
          </Row>
        </div>

        {/* 菜单树形视图 */}
        <Row gutter={24}>
          <Col span={collapsed ? 4 : 8}>
            <Card size="small" title={
              <div className="flex items-center justify-between">
                <span>{t('adminMenu.tree.title')}</span>
                <Button type="text" onClick={() => setCollapsed(prev => !prev)} icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}>{t('adminMenu.tree.collapse')}</Button>
              </div>
            }>
              <Tree
                treeData={treeData}
                expandedKeys={expandedKeys}
                selectedKeys={selectedKeys}
                onExpand={setExpandedKeys}
                onSelect={(keys) => {
                  setSelectedKeys(keys);
                  const id = String(keys?.[0] ?? '');
                  if (id) {
                    // 在树数据中查找选中菜单
                    const findById = (list: MenuItem[]): MenuItem | null => {
                      for (const it of list) {
                        if (it.id === id) return it;
                        if (Array.isArray(it.children)) {
                          const found = findById(it.children);
                          if (found) return found;
                        }
                      }
                      return null;
                    };
                    const found = findById(menuTreeItems);
                    setSelectedMenu(found);
                    // 展开选中的节点
                    setExpandedKeys(prev => Array.from(new Set([id, ...prev])));
                  } else {
                    setSelectedMenu(null);
                  }
                }}
                showLine
                showIcon={false}
              />
            </Card>
          </Col>
          
          <Col span={16}>
            {selectedMenu && (
              <Card className="mb-3" size="small" title={
                <div className="flex items-center">
                  <span className="font-semibold" style={{ marginRight: 8 }}>{selectedMenu.name}</span>
                  <Tag size="small" color={selectedMenu.status === 'active' ? 'green' : 'red'}>
                    {selectedMenu.status === 'active' ? t('adminMenu.filters.statusOptions.enabled') : t('adminMenu.filters.statusOptions.disabled')}
                  </Tag>
                  <Tag size="small" color={selectedMenu.type === 'menu' ? 'blue' : selectedMenu.type === 'button' ? 'orange' : 'purple'} style={{ marginLeft: 8 }}>
                    {selectedMenu.type === 'menu' ? t('adminMenu.filters.typeOptions.menu') : selectedMenu.type === 'button' ? t('adminMenu.filters.typeOptions.button') : t('adminMenu.filters.typeOptions.api')}
                  </Tag>
                </div>
              }>
                <div className="flex items-center justify-between">
                  <div className="flex items-center">
                    <code className="bg-gray-100 px-2 py-0.5 rounded text-xs">{selectedMenu.path}</code>
                  </div>
                  <Space>
                    <Button type="text" size="small" icon={<EyeOutlined />} onClick={() => handleView(selectedMenu)}>{t('adminMenu.actions.viewDetails')}</Button>
                    <Button type="text" size="small" icon={<EditOutlined />} onClick={() => handleEdit(selectedMenu)}>{t('adminMenu.actions.editMenu')}</Button>
                    <Button type="text" size="small" icon={<PlusOutlined />} onClick={() => {
                      setEditingMenu(null);
                      setModalVisible(true);
                      // 预设父菜单为当前选中项
                      setTimeout(() => {
                        form.setFieldsValue({ parentId: selectedMenu.id });
                      }, 0);
                    }}>{t('adminMenu.actions.addSubmenu')}</Button>
                  </Space>
                </div>
              </Card>
            )}
            {/* 菜单表格 */}
            <Table
              columns={columns}
              dataSource={visibleMenus}
              rowKey="id"
              loading={loading}
              pagination={{
                current: pagination.current,
                pageSize: pagination.pageSize,
                total: pagination.total,
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (total, range) => t('adminMenu.pagination.total', { start: range[0], end: range[1], total }),
                onChange: (page, pageSize) => {
                  loadMenus(page, pageSize);
                }
              }}
            />
          </Col>
        </Row>

        {/* 新增/编辑菜单弹窗 */}
        <Modal
          title={editingMenu ? t('adminMenu.modal.editTitle') : t('adminMenu.modal.addTitle')}
          open={modalVisible}
          onOk={handleSubmit}
          onCancel={() => setModalVisible(false)}
          width={900}
          styles={{ body: { maxHeight: '70vh', overflowY: 'auto' } }}
          destroyOnHidden
          forceRender
        >
          <Form
            form={form}
            layout="vertical"
            initialValues={{
              status: 'active',
              type: 'menu',
              sort: 1,
              permissions: []
            }}
          >
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  label={t('adminMenu.form.nameLabel')}
                  name="name"
                  rules={[{ required: true, message: t('adminMenu.form.nameRequired') }]}
                >
                  <Input placeholder={t('adminMenu.form.namePlaceholder')} />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  label={t('adminMenu.form.pathLabel')}
                  name="path"
                  dependencies={['parentId','type']}
                  rules={[
                    {
                      validator: async (_, value) => {
                        const { parentId, type } = form.getFieldsValue(['parentId','type']);
                        // 子级菜单且类型为菜单，路径必填并需合法
                        if (type === 'menu' && parentId) {
                          if (!value) throw new Error(t('adminMenu.form.pathRequired'));
                          const pattern = /^\/[a-z0-9-]+(\/[a-z0-9-]+)*$/;
                          if (!pattern.test(value)) {
                            throw new Error(t('adminMenu.form.pathPattern'));
                          }
                        }
                        // 顶级菜单路径可为空，表示目录节点（不可点击）
                        return Promise.resolve();
                      }
                    }
                  ]}
                >
                  <Input placeholder={t('adminMenu.form.pathPlaceholder')} />
                </Form.Item>
              </Col>
            </Row>

            <Row gutter={16}>
              <Col span={12}>
                <Form.Item label={t('adminMenu.form.iconLabel')} name="icon">
                  <Space>
                    <Button onClick={openIconSelector}>{t('adminMenu.form.iconSelect')}</Button>
                    <Tooltip title={String(form.getFieldValue('icon') || defaultIconName)}>
                      <span style={{ display: 'inline-flex', alignItems: 'center', cursor: 'pointer' }} onClick={openIconSelector}>
                        {getIconNode(form.getFieldValue('icon') || defaultIconName, { fontSize: 18, marginRight: 6 })}
                        <code className="bg-gray-100 px-2 py-0.5 rounded text-xs">{String(form.getFieldValue('icon') || defaultIconName)}</code>
                      </span>
                    </Tooltip>
                  </Space>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label={t('adminMenu.form.componentLabel')} name="component">
                  <Input placeholder={t('adminMenu.form.componentPlaceholder')} />
                </Form.Item>
              </Col>
            </Row>

            <Row gutter={16}>
              <Col span={8}>
                <Form.Item label={t('adminMenu.form.typeLabel')} name="type">
                  <Select>
                    <Option value="menu">{t('adminMenu.filters.typeOptions.menu')}</Option>
                    <Option value="button">{t('adminMenu.filters.typeOptions.button')}</Option>
                    <Option value="api">{t('adminMenu.filters.typeOptions.api')}</Option>
                  </Select>
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item label={t('adminMenu.form.statusLabel')} name="status">
                  <Select>
                    <Option value="active">{t('adminMenu.filters.statusOptions.enabled')}</Option>
                    <Option value="inactive">{t('adminMenu.filters.statusOptions.disabled')}</Option>
                  </Select>
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item label={t('adminMenu.form.sortLabel')} name="sort">
                  <InputNumber min={1} style={{ width: '100%' }} />
                </Form.Item>
              </Col>
            </Row>

            <Form.Item label={t('adminMenu.form.parentLabel')} name="parentId">
              <TreeSelect
                placeholder={t('adminMenu.form.parentSelectPlaceholder')}
                allowClear
                treeData={getParentMenuOptions()}
                treeDefaultExpandAll
              />
            </Form.Item>

            <Form.Item label={t('adminMenu.form.permissionsLabel')} name="permissions">
              <Select
                mode="multiple"
                placeholder={t('adminMenu.form.permissionsSelectPlaceholder')}
                allowClear
              >
                {permissions.map(permission => (
                  <Option key={permission.id} value={permission.id}>
                    {permission.name} ({permission.code})
                  </Option>
                ))}
              </Select>
            </Form.Item>
          </Form>
        </Modal>
      </Card>
      <IconSelectorModal
        open={iconModalOpen}
        onClose={() => setIconModalOpen(false)}
        initial={form.getFieldValue('icon')}
        onSelect={(name) => {
          if (isValidIconName(name)) {
            form.setFieldsValue({ icon: name });
          }
        }}
      />
    </div>
  );
};

export default MenuManagement;