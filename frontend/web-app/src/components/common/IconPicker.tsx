import React, { useMemo, useState, useEffect } from 'react';
import { Tabs, Input, Tooltip, Button, Space, Upload, message } from 'antd';
import type { UploadProps } from 'antd';
import { StarFilled, StarOutlined, SearchOutlined, PictureOutlined } from '@ant-design/icons';
import {
  iconCategories,
  searchIcons,
  getFavorites,
  toggleFavorite,
  getRecents,
  addRecent,
  getIconNode,
  defaultIconName,
  isValidIconName,
} from '../../ui/icons/iconRegistry';

type Props = {
  value?: string;
  onChange?: (value: string) => void;
};

const categoryTabs = [
  { key: 'common', label: '常用' },
  { key: 'system', label: '系统' },
  { key: 'business', label: '业务' },
  { key: 'favorites', label: '收藏' },
  { key: 'recents', label: '最近' },
];

export default function IconPicker({ value, onChange }: Props) {
  const [activeTab, setActiveTab] = useState<string>('common');
  const [query, setQuery] = useState<string>('');
  const [favorites, setFavorites] = useState<string[]>(getFavorites());
  const [recents, setRecents] = useState<string[]>(getRecents());
  const selectedName = useMemo(() => {
    if (value && isValidIconName(value)) return value;
    return defaultIconName;
  }, [value]);

  useEffect(() => {
    // 保证选择器初次打开也记录最近项
    setRecents(addRecent(selectedName));
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const allList = useMemo(() => {
    if (activeTab === 'favorites') {
      return favorites
        .map((n) => ({ name: n }))
        .filter((i) => isValidIconName(i.name));
    }
    if (activeTab === 'recents') {
      return recents
        .map((n) => ({ name: n }))
        .filter((i) => isValidIconName(i.name));
    }
    const category = activeTab as 'system' | 'business' | 'common';
    return searchIcons(query, category).map((m) => ({ name: m.name }));
  }, [activeTab, query, favorites, recents]);

  const handlePick = (name: string) => {
    onChange?.(name);
    setRecents(addRecent(name));
  };

  const handleToggleFavorite = (name: string) => {
    const updated = toggleFavorite(name);
    setFavorites(updated);
  };

  const uploadProps: UploadProps = {
    accept: '.svg',
    showUploadList: false,
    beforeUpload: (file) => {
      const isSvg = file.type === 'image/svg+xml' || file.name.toLowerCase().endsWith('.svg');
      if (!isSvg) {
        message.error('仅支持上传SVG格式图标');
        return Upload.LIST_IGNORE;
      }
      const reader = new FileReader();
      reader.onload = (e) => {
        const content = String(e.target?.result || '');
        const iconValue = `svg:${content}`;
        onChange?.(iconValue);
        message.success('SVG图标已加载并选中');
      };
      reader.readAsText(file);
      return Upload.LIST_IGNORE;
    },
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      <Space align="center" style={{ justifyContent: 'space-between' }}>
        <Input
          allowClear
          prefix={<SearchOutlined />}
          placeholder="搜索图标名称或关键词"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          style={{ maxWidth: 280 }}
        />
        <Space>
          <Upload {...uploadProps}>
            <Button icon={<PictureOutlined />}>上传SVG</Button>
          </Upload>
        </Space>
      </Space>

      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={categoryTabs.map((t) => ({ key: t.key, label: t.label }))}
      />

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(64px, 1fr))', gap: 8 }}>
        {allList.map((item) => {
          const node = getIconNode(item.name);
          const picked = selectedName === item.name;
          const isFav = favorites.includes(item.name);
          return (
            <div
              key={item.name}
              onClick={() => handlePick(item.name)}
              style={{
                cursor: 'pointer',
                padding: 8,
                border: picked ? '2px solid #1677ff' : '1px solid #f0f0f0',
                borderRadius: 8,
                textAlign: 'center',
                background: picked ? '#e6f4ff' : '#fff',
              }}
            >
              <Tooltip title={item.name}>
                <div style={{ fontSize: 22, lineHeight: '32px' }}>{node}</div>
              </Tooltip>
              <div style={{ marginTop: 6 }}>
                <Button
                  size="small"
                  type="text"
                  icon={isFav ? <StarFilled style={{ color: '#faad14' }} /> : <StarOutlined />}
                  onClick={(e) => {
                    e.stopPropagation();
                    handleToggleFavorite(item.name);
                  }}
                />
              </div>
            </div>
          );
        })}
      </div>

      <div style={{ marginTop: 12, padding: 12, border: '1px dashed #d9d9d9', borderRadius: 8 }}>
        <Space direction="vertical" style={{ width: '100%' }}>
          <div>预览</div>
          <div style={{ fontSize: 36 }}>{getIconNode(selectedName)}</div>
          <div style={{ color: '#999' }}>已选：{selectedName}</div>
        </Space>
      </div>
    </div>
  );
}