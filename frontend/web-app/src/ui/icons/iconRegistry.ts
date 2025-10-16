import React from 'react';
import * as AntIcons from '@ant-design/icons';

export type IconComponent = React.ComponentType<any>;

export type IconMeta = {
  name: string; // 唯一标识（用于后端/本地存储）
  title: string; // 中文标题
  keywords?: string[]; // 搜索关键词
  component: IconComponent;
  category: 'system' | 'business' | 'common';
};

export const defaultIconName = 'MenuOutlined';

const allowedSuffixes = ['Outlined', 'Filled', 'TwoTone'];

function toKeywords(name: string): string[] {
  return name
    .replace(/([a-z])([A-Z])/g, '$1 $2')
    .split(/\s+/)
    .filter(Boolean)
    .map((s) => s.toLowerCase());
}

const registryList: IconMeta[] = Object.keys(AntIcons)
  .filter((key) => {
    const comp = (AntIcons as any)[key];
    if (!comp) return false;
    // 仅收集标准命名的图标组件
    return allowedSuffixes.some((suf) => key.endsWith(suf));
  })
  .map((key) => ({
    name: key,
    title: key, // 大规模库以原名为标题，避免维护成本
    keywords: toKeywords(key),
    component: (AntIcons as any)[key] as IconComponent,
    category: 'common',
  }));

const registryMap: Record<string, IconMeta> = registryList.reduce((acc, meta) => {
  acc[meta.name] = meta;
  return acc;
}, {} as Record<string, IconMeta>);

export function listIcons(category?: IconMeta['category']): IconMeta[] {
  return registryList.filter((m) => (category ? m.category === category : true));
}

export function searchIcons(query: string, category?: IconMeta['category']): IconMeta[] {
  const q = query.trim().toLowerCase();
  const base = listIcons(category);
  if (!q) return base;
  return base.filter((m) => {
    const hay = [m.name, m.title, ...(m.keywords || [])].join('|').toLowerCase();
    return hay.includes(q);
  });
}

export function getIconNode(name?: string, style?: React.CSSProperties): React.ReactNode {
  const n = (name || defaultIconName).trim();
  // 支持自定义SVG: 前缀 svg: 原始内容
  if (n.startsWith('svg:')) {
    const svgContent = n.slice(4);
    // 注意：此处使用 dangerouslySetInnerHTML，仅用于受信任来源
    return React.createElement('span', { style, dangerouslySetInnerHTML: { __html: svgContent } });
  }
  const meta = registryMap[n] || registryMap[defaultIconName];
  const Comp = meta.component;
  return React.createElement(Comp, { style });
}

// 收藏与最近使用（本地持久化）
const FAVORITES_KEY = 'iconFavorites';
const RECENTS_KEY = 'iconRecents';

function readLS<T>(key: string, fallback: T): T {
  try {
    const v = localStorage.getItem(key);
    return v ? (JSON.parse(v) as T) : fallback;
  } catch {
    return fallback;
  }
}

function writeLS<T>(key: string, value: T): void {
  try {
    localStorage.setItem(key, JSON.stringify(value));
  } catch {}
}

export function getFavorites(): string[] {
  return readLS<string[]>(FAVORITES_KEY, []);
}

export function toggleFavorite(name: string): string[] {
  const favs = getFavorites();
  const idx = favs.indexOf(name);
  if (idx >= 0) {
    favs.splice(idx, 1);
  } else {
    favs.push(name);
  }
  writeLS(FAVORITES_KEY, favs);
  return favs;
}

export function getRecents(): string[] {
  return readLS<string[]>(RECENTS_KEY, []);
}

export function addRecent(name: string): string[] {
  const recents = getRecents().filter((n) => n !== name);
  recents.unshift(name);
  const capped = recents.slice(0, 5);
  writeLS(RECENTS_KEY, capped);
  return capped;
}

export function isValidIconName(name?: string): boolean {
  if (!name) return false;
  if (name.startsWith('svg:')) return true;
  return !!registryMap[name];
}

export const iconCategories = {
  // 动态库目前统一归类为 common；如需细分可在后续增加分类规则
  common: listIcons('common'),
};