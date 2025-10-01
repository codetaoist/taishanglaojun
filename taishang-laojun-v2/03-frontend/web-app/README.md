# Web前端应用模块

## 🎯 模块目标

构建现代化的Web前端应用，提供优雅的用户界面和流畅的交互体验。

## 📋 主要功能

### 1. 用户界面
- 响应式设计
- 现代化UI组件
- 主题切换支持
- 无障碍访问

### 2. 核心页面
- 首页和导航
- 智慧内容浏览
- 搜索和发现
- 个人中心

### 3. 交互功能
- 智能对话界面
- 内容收藏和分享
- 学习进度跟踪
- 社区互动

### 4. 性能优化
- 代码分割
- 懒加载
- 缓存策略
- SEO优化

## 🚀 开发优先级

**P0 - 立即开始**：
- [ ] 项目脚手架搭建
- [ ] 基础组件库
- [ ] 路由和状态管理

**P1 - 第一周完成**：
- [ ] 核心页面开发
- [ ] API集成
- [ ] 用户认证界面

**P2 - 第二周完成**：
- [ ] 高级交互功能
- [ ] 性能优化
- [ ] 测试和部署

## 🔧 技术栈

- **框架**：React 18 + TypeScript
- **构建工具**：Vite
- **状态管理**：Zustand / Redux Toolkit
- **路由**：React Router v6
- **UI库**：Ant Design / Material-UI
- **样式**：Tailwind CSS / Styled Components
- **HTTP客户端**：Axios / React Query
- **测试**：Jest + React Testing Library

## 📁 目录结构

```
web-app/
├── public/
│   ├── index.html
│   └── favicon.ico
├── src/
│   ├── components/           # 通用组件
│   │   ├── common/          # 基础组件
│   │   ├── layout/          # 布局组件
│   │   └── business/        # 业务组件
│   ├── pages/               # 页面组件
│   │   ├── Home/            # 首页
│   │   ├── Wisdom/          # 智慧内容
│   │   ├── Search/          # 搜索页面
│   │   ├── Chat/            # 对话页面
│   │   └── Profile/         # 个人中心
│   ├── hooks/               # 自定义Hooks
│   │   ├── useAuth.ts       # 认证Hook
│   │   ├── useApi.ts        # API Hook
│   │   └── useLocalStorage.ts
│   ├── services/            # API服务
│   │   ├── api.ts           # API配置
│   │   ├── auth.ts          # 认证服务
│   │   ├── wisdom.ts        # 智慧服务
│   │   └── ai.ts            # AI服务
│   ├── store/               # 状态管理
│   │   ├── authStore.ts     # 认证状态
│   │   ├── wisdomStore.ts   # 智慧状态
│   │   └── uiStore.ts       # UI状态
│   ├── utils/               # 工具函数
│   │   ├── constants.ts     # 常量
│   │   ├── helpers.ts       # 辅助函数
│   │   └── validators.ts    # 验证函数
│   ├── styles/              # 样式文件
│   │   ├── globals.css      # 全局样式
│   │   └── themes.ts        # 主题配置
│   └── types/               # 类型定义
│       ├── api.ts           # API类型
│       ├── user.ts          # 用户类型
│       └── wisdom.ts        # 智慧类型
├── tests/                   # 测试文件
│   ├── components/          # 组件测试
│   └── pages/               # 页面测试
├── package.json
├── vite.config.ts
├── tsconfig.json
└── tailwind.config.js
```

## 🎯 核心组件设计

```typescript
// 智慧内容卡片组件
interface WisdomCardProps {
  wisdom: CulturalWisdom;
  onLike: (id: string) => void;
  onShare: (id: string) => void;
  onCollect: (id: string) => void;
}

// 智能对话组件
interface ChatInterfaceProps {
  sessionId?: string;
  onNewMessage: (message: string) => void;
  messages: ChatMessage[];
  loading: boolean;
}

// 搜索组件
interface SearchComponentProps {
  onSearch: (query: string, filters: SearchFilters) => void;
  results: SearchResult[];
  loading: boolean;
  pagination: PaginationInfo;
}
```

## 🎯 页面路由设计

```typescript
const routes = [
  {
    path: '/',
    element: <HomePage />,
  },
  {
    path: '/wisdom',
    element: <WisdomListPage />,
  },
  {
    path: '/wisdom/:id',
    element: <WisdomDetailPage />,
  },
  {
    path: '/search',
    element: <SearchPage />,
  },
  {
    path: '/chat',
    element: <ChatPage />,
    protected: true,
  },
  {
    path: '/profile',
    element: <ProfilePage />,
    protected: true,
  },
];
```

## 🎯 状态管理设计

```typescript
// 认证状态
interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => void;
  refreshToken: () => Promise<void>;
}

// 智慧内容状态
interface WisdomState {
  wisdoms: CulturalWisdom[];
  currentWisdom: CulturalWisdom | null;
  loading: boolean;
  error: string | null;
  fetchWisdoms: (params: FetchParams) => Promise<void>;
  fetchWisdomById: (id: string) => Promise<void>;
}
```

## 🎯 API集成示例

```typescript
// API服务配置
const apiClient = axios.create({
  baseURL: process.env.REACT_APP_API_BASE_URL,
  timeout: 10000,
});

// 智慧内容API
export const wisdomApi = {
  getWisdoms: (params: GetWisdomsParams) =>
    apiClient.get<WisdomListResponse>('/api/v1/wisdom', { params }),
    
  getWisdomById: (id: string) =>
    apiClient.get<CulturalWisdom>(`/api/v1/wisdom/${id}`),
    
  searchWisdoms: (query: string, filters: SearchFilters) =>
    apiClient.get<SearchResponse>('/api/v1/search', {
      params: { q: query, ...filters }
    }),
};
```

## 🎯 成功标准

- [ ] 项目脚手架搭建完成
- [ ] 基础组件库可用
- [ ] 核心页面功能完整
- [ ] API集成正常工作
- [ ] 响应式设计适配
- [ ] 性能指标达标（LCP < 2.5s）
- [ ] 无障碍访问支持
- [ ] 单元测试覆盖率 > 80%