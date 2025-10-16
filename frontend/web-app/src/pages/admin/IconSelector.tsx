import React, { useMemo, useState, useEffect } from 'react';
import { Card, Tabs, Input, Button, Space, Tooltip, Pagination, App } from 'antd';
import { StarFilled, StarOutlined, SearchOutlined, ArrowLeftOutlined, CheckOutlined } from '@ant-design/icons';
import {
  iconCategories,
  searchIcons,
  getFavorites,
  toggleFavorite,
  getRecents,
  addRecent,
  getIconNode,
  isValidIconName,
  defaultIconName,
} from '../../ui/icons/iconRegistry';
import { useNavigate, useLocation } from 'react-router-dom';

type CategoryKey = 'common' | 'system' | 'business' | 'favorites' | 'recents' | 'all';

const categoryTabs = [
  { key: 'common', label: '常用' },
  { key: 'system', label: '系统' },
  { key: 'business', label: '业务' },
  { key: 'favorites', label: '收藏' },
  { key: 'recents', label: '最近' },
  { key: 'all', label: '全部' },
];

const PAGE_SIZE = 60;

const IconSelectorPage: React.FC = () => {
  const { message } = App.useApp();
  const navigate = useNavigate();
  const location = useLocation() as { state?: { returnTo?: string; field?: string } };
  const returnTo = location.state?.returnTo || '/admin/menus';
  const field = location.state?.field || 'icon';

  const [activeTab, setActiveTab] = useState<CategoryKey>('common');
  const [query, setQuery] = useState<string>('');
  const [favorites, setFavorites] = useState<string[]>(getFavorites());
  const [recents, setRecents] = useState<string[]>(getRecents());
  const [selected, setSelected] = useState<string>(defaultIconName);
  const [page, setPage] = useState<number>(1);

  // 初始从会话中读取之前的选择
  useEffect(() => {
    const current = sessionStorage.getItem('iconPicker:current');
    if (current && isValidIconName(current)) {
      setSelected(current);
    }
  }, []);

  const allList = useMemo(() => {
    let items: { name: string }[] = [];
    if (activeTab === 'favorites') {
      items = favorites.map((n) => ({ name: n })).filter((i) => isValidIconName(i.name));
    } else if (activeTab === 'recents') {
      items = recents.map((n) => ({ name: n })).filter((i) => isValidIconName(i.name));
    } else if (activeTab === 'all') {
      // 聚合所有分类
      const common = iconCategories.common.map((m) => ({ name: m.name }));
      // system/business 分类暂时与 common 合并或可扩展后端规则
      items = common;
      if (query.trim()) {
        items = searchIcons(query).map((m) => ({ name: m.name }));
      }
    } else {
      const category = activeTab as 'system' | 'business' | 'common';
      items = searchIcons(query, category).map((m) => ({ name: m.name }));
    }
    return items;
  }, [activeTab, query, favorites, recents]);

  const total = allList.length;
  const pageItems = useMemo(() => {
    const start = (page - 1) * PAGE_SIZE;
    return allList.slice(start, start + PAGE_SIZE);
  }, [allList, page]);

  useEffect(() => {
    // 切换标签或搜索时重置分页
    setPage(1);
  }, [activeTab, query]);

  const handleSelect = (name: string) => {
    setSelected(name);
  };

  const handleFavoriteToggle = (name: string) => {
    const next = toggleFavorite(name);
    setFavorites(next);
  };

  const handleConfirm = () => {
    if (!selected || !isValidIconName(selected)) {
      message.error('请选择合法图标');
      return;
    }
    addRecent(selected);
    sessionStorage.setItem('iconPicker:selected', selected);
    // 记录当前字段名，便于回填
    sessionStorage.setItem('iconPicker:field', field);
    navigate(returnTo);
  };

  const handleBack = () => {
    navigate(-1);
  };

  return (
    <div className="p-6">
      <Card title={<Space><Button type="text" icon={<ArrowLeftOutlined />} onClick={handleBack}>返回</Button><span>选择图标</span></Space>}>
        <div className="mb-4 flex items-center justify-between">
          <Space>
            <Input
              placeholder="搜索图标名称或关键词"
              allowClear
              prefix={<SearchOutlined />}
              style={{ width: 320 }}
              value={query}
              onChange={(e) => setQuery(e.target.value)}
            />
          </Space>
          <Space>
            <Tooltip title="当前选择">
              <span style={{ display: 'inline-flex', alignItems: 'center' }}>
                {getIconNode(selected, { fontSize: 20, marginRight: 8 })}
                <code className="bg-gray-100 px-2 py-0.5 rounded text-xs">{selected}</code>
              </span>
            </Tooltip>
            <Button type="primary" icon={<CheckOutlined />} onClick={handleConfirm}>确认选择</Button>
          </Space>
        </div>

        <Tabs
          activeKey={activeTab}
          onChange={(k) => setActiveTab(k as CategoryKey)}
          items={categoryTabs.map((t) => ({ key: t.key, label: t.label }))}
        />

        <div className="grid grid-cols-6 gap-3" style={{ minHeight: 360 }}>
          {pageItems.map((item) => {
            const isFav = favorites.includes(item.name);
            const isSelected = selected === item.name;
            return (
              <Card
                key={item.name}
                hoverable
                size="small"
                className={isSelected ? 'border-blue-500' : ''}
                onClick={() => handleSelect(item.name)}
                bodyStyle={{ padding: 12 }}
              >
                <div className="flex items-center justify-between">
                  <Tooltip title={item.name}>
                    <span style={{ display: 'inline-flex', alignItems: 'center' }}>
                      {getIconNode(item.name, { fontSize: 20, marginRight: 8 })}
                      <span style={{ fontSize: 12 }}>{item.name.replace(/(Outlined|Filled|TwoTone)$/,'')}</span>
                    </span>
                  </Tooltip>
                  <Button
                    type="text"
                    size="small"
                    icon={isFav ? <StarFilled style={{ color: '#faad14' }} /> : <StarOutlined />}
                    onClick={(e) => { e.stopPropagation(); handleFavoriteToggle(item.name); }}
                  />
                </div>
              </Card>
            );
          })}
        </div>

        <div className="mt-4 flex justify-end">
          <Pagination
            current={page}
            pageSize={PAGE_SIZE}
            total={total}
            onChange={(p) => setPage(p)}
            showSizeChanger={false}
            showTotal={(t) => `共 ${t} 个图标`}
          />
        </div>
      </Card>
    </div>
  );
};

export default IconSelectorPage;