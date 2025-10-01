import { useDispatch, useSelector, TypedUseSelectorHook } from 'react-redux';
import type { RootState, AppDispatch } from '../store';

// 使用类型化的hooks
export const useAppDispatch = () => useDispatch<AppDispatch>();
export const useAppSelector: TypedUseSelectorHook<RootState> = useSelector;

// 自定义hooks
export const useAuth = () => {
  return useAppSelector(state => state.auth);
};

export const useUI = () => {
  return useAppSelector(state => state.ui);
};

export const useConsciousness = () => {
  return useAppSelector(state => state.consciousness);
};

export const useCultural = () => {
  return useAppSelector(state => state.cultural);
};

export const useAdmin = () => {
  return useAppSelector(state => state.admin);
};

export const useNotification = () => {
  return useAppSelector(state => state.notification);
};