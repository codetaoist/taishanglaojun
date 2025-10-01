// Redux Slices 统一导出
import { combineReducers } from '@reduxjs/toolkit';

// 导入所有 slice
import authSlice from './authSlice';
import uiSlice from './uiSlice';
import notificationSlice from './notificationSlice';
import adminSlice from './adminSlice';
import consciousnessSlice from './consciousnessSlice';
import culturalSlice from './culturalSlice';

// 合并所有 reducer
const rootReducer = combineReducers({
  auth: authSlice,
  ui: uiSlice,
  notification: notificationSlice,
  admin: adminSlice,
  consciousness: consciousnessSlice,
  cultural: culturalSlice,
});

export default rootReducer;

// 导出类型
export type RootState = ReturnType<typeof rootReducer>;