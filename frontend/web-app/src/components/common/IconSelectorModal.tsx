import React, { useEffect, useMemo, useState } from 'react';
import { Modal, Card, Tabs, Input, Button, Space, Tooltip, Pagination } from 'antd';
import { StarFilled, StarOutlined, SearchOutlined } from '@ant-design/icons';
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

type CategoryKey = 'common' | 'system' | 'business' | 'favorites' | 'recents' | 'all';

type Props = {
  open: boolean;
  onClose: () => void;
  onSelect: (iconName: string) => void;
  initial?: string;
};

const categoryTabs = [
  { key: 'common', label: '常用' },
  { key: 'system', label: '系统' },
  { key: 'business', label: '业务' },
  { key: 'favorites', label: '收藏' },
  { key: 'recents', label: '最近' },
  { key: 'all', label: '全部' },
];

const PAGE_SIZE = 60;

const IconSelectorModal: React.FC<Props> = ({ open, onClose, onSelect, initial }) => {
  const [activeTab, setActiveTab] = useState<CategoryKey>('common');
  const [query, setQuery] = useState<string>('');
  const [favorites, setFavorites] = useState<string[]>(getFavorites());
  const [recents, setRecents] = useState<string[]>(getRecents());
  const [page, setPage] = useState<number>(1);
  const [selected, setSelected] = useState<string>(initial && isValidIconName(initial) ? initial : defaultIconName);

  useEffect(() => {
    // 重置状态于打开时
    if (open) {
      setActiveTab('common');
      setQuery('');
      setPage(1);
      setSelected(initial && isValidIconName(initial) ? initial : defaultIconName);
    }
  }, [open, initial]);

  const allList = useMemo(() => {
    let items: { name: string }[] = [];
    if (activeTab === 'favorites') {
      items = favorites.map((n) => ({ name: n })).filter((i) => isValidIconName(i.name));
    } else if (activeTab === 'recents') {
      items = recents.map((n) => ({ name: n })).filter((i) => isValidIconName(i.name));
    } else if (activeTab === 'all') {
      const common = iconCategories.common.map((m) => ({ name: m.name }));
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
    setPage(1);
  }, [activeTab, query]);

  const handleFavoriteToggle = (name: string) => {
    const next = toggleFavorite(name);
    setFavorites(next);
  };

  const handleClickItem = (name: string) => {
    setSelected(name);
    addRecent(name);
    onSelect(name);
    onClose();
  };

  return (
    <Modal
      title="选择图标"
      open={open}
      onCancel={onClose}
      footer={null}
      width={900}
      styles={{ body: { maxHeight: '70vh', overflowY: 'auto' } }}
      destroyOnClose
    >
      <div className="mb-3 flex items-center justify-between">
        <Input
          placeholder="搜索图标名称或关键词"
          allowClear
          prefix={<SearchOutlined />}
          style={{ width: 360 }}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
        <Tooltip title="当前选择">
          <span style={{ display: 'inline-flex', alignItems: 'center' }}>
            {getIconNode(selected, { fontSize: 20, marginRight: 8 })}
            <code className="bg-gray-100 px-2 py-0.5 rounded text-xs">{selected}</code>
          </span>
        </Tooltip>
      </div>

      <Tabs
        activeKey={activeTab}
        onChange={(k) => setActiveTab(k as CategoryKey)}
        items={categoryTabs.map((t) => ({ key: t.key, label: t.label }))}
      />

      <div className="grid grid-cols-6 gap-3" style={{ minHeight: 360 }}>
        {pageItems.map((item) => {
          const isFav = favorites.includes(item.name);
          return (
            <Card
              key={item.name}
              hoverable
              size="small"
              onClick={() => handleClickItem(item.name)}
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
    </Modal>
  );
};

export default IconSelectorModal;