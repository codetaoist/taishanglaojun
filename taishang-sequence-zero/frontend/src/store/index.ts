import { configureStore, combineReducers } from '@reduxjs/toolkit';
import { persistStore, persistReducer, FLUSH, REHYDRATE, PAUSE, PERSIST, PURGE, REGISTER } from 'redux-persist';
import storage from 'redux-persist/lib/storage';
import { PersistConfig } from 'redux-persist';

// 导入所有slice
import authSlice from './slices/authSlice';
import uiSlice from './slices/uiSlice';
import consciousnessSlice from './slices/consciousnessSlice';
import culturalSlice from './slices/culturalSlice';
import adminSlice from './slices/adminSlice';
import notificationSlice from './slices/notificationSlice';

// 根reducer
const rootReducer = combineReducers({
  auth: authSlice,
  ui: uiSlice,
  consciousness: consciousnessSlice,
  cultural: culturalSlice,
  admin: adminSlice,
  notification: notificationSlice,
});

// 持久化配置
const persistConfig: PersistConfig<RootState> = {
  key: 'taishang-sequence-zero',
  version: 1,
  storage,
  // 指定需要持久化的reducer
  whitelist: ['auth', 'ui'],
  // 指定不需要持久化的reducer
  blacklist: ['notification'],
  // 迁移配置
  migrate: (state: any) => {
    // 版本迁移逻辑
    if (state && state._persist && state._persist.version < 1) {
      // 执行迁移操作
      return Promise.resolve({
        ...state,
        _persist: {
          ...state._persist,
          version: 1,
        },
      });
    }
    return Promise.resolve(state);
  },
};

// 创建持久化reducer
const persistedReducer = persistReducer(persistConfig, rootReducer);

// 配置store
export const store = configureStore({
  reducer: persistedReducer,
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        // 忽略redux-persist的action类型
        ignoredActions: [FLUSH, REHYDRATE, PAUSE, PERSIST, PURGE, REGISTER],
        // 忽略特定的路径
        ignoredPaths: ['register'],
      },
      // 开发环境启用不可变性检查
      immutableCheck: process.env.NODE_ENV === 'development',
    }),
  // 开发环境启用Redux DevTools
  devTools: process.env.NODE_ENV === 'development',
  // 预加载状态
  preloadedState: undefined,
});

// 创建persistor
export const persistor = persistStore(store);

// 导出类型
export type RootState = ReturnType<typeof rootReducer>;
export type AppDispatch = typeof store.dispatch;
export type AppStore = typeof store;

// 开发环境热重载支持
if (process.env.NODE_ENV === 'development' && (module as any).hot) {
  (module as any).hot.accept('./slices', () => {
    const newRootReducer = require('./slices').default;
    store.replaceReducer(persistReducer(persistConfig, newRootReducer));
  });
}

// Store增强器
const storeEnhancers = [];

// 开发环境添加调试工具
if (process.env.NODE_ENV === 'development') {
  // 可以添加其他开发工具
}

// 导出store实例
export default store;

// 工具函数：清除持久化数据
export const clearPersistedData = () => {
  persistor.purge();
  localStorage.removeItem('persist:taishang-sequence-zero');
};

// 工具函数：重置store状态
export const resetStore = () => {
  store.dispatch({ type: 'RESET_STORE' });
};

// 工具函数：获取当前状态
export const getCurrentState = (): RootState => store.getState();

// 工具函数：订阅状态变化
export const subscribeToStore = (listener: () => void) => {
  return store.subscribe(listener);
};

// 类型守卫：检查是否为有效的RootState
export const isValidRootState = (state: any): state is RootState => {
  return (
    state &&
    typeof state === 'object' &&
    'auth' in state &&
    'ui' in state &&
    'consciousness' in state &&
    'cultural' in state &&
    'admin' in state &&
    'notification' in state
  );
};

// 中间件：日志记录
const loggerMiddleware = (store: any) => (next: any) => (action: any) => {
  if (process.env.NODE_ENV === 'development') {
    console.group(`Action: ${action.type}`);
    console.log('Previous State:', store.getState());
    console.log('Action:', action);
    const result = next(action);
    console.log('Next State:', store.getState());
    console.groupEnd();
    return result;
  }
  return next(action);
};

// 中间件：错误处理
const errorMiddleware = (store: any) => (next: any) => (action: any) => {
  try {
    return next(action);
  } catch (error) {
    console.error('Redux Error:', error);
    // 可以在这里添加错误上报逻辑
    throw error;
  }
};

// 导出中间件（如果需要在其他地方使用）
export { loggerMiddleware, errorMiddleware };