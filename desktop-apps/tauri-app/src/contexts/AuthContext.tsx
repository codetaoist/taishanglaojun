import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { invoke } from '@tauri-apps/api/core';

interface User {
  id: string;
  username: string;
  email: string;
  role: string;
}

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  loading: boolean;
  login: (username: string, password: string) => Promise<boolean>;
  logout: () => Promise<void>;
  register: (username: string, email: string, password: string) => Promise<boolean>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check if user is already logged in
    checkAuthStatus();
  }, []);

  const checkAuthStatus = async () => {
    try {
      // Check if we're in a Tauri environment
      if (typeof window !== 'undefined' && window.__TAURI__) {
        // This would typically check for a stored session token
        const sessionValid = await invoke('validate_session');
        if (sessionValid) {
          // Get user info if session is valid
          const userInfo = await invoke('get_user_info');
          setUser(userInfo as User);
        }
      } else {
        // In browser environment, simulate authentication for development
        console.log('Running in browser mode - simulating authentication');
        setUser({ id: '1', username: 'demo', email: 'demo@example.com' });
      }
    } catch (error) {
      console.error('Auth check failed:', error);
    } finally {
      setLoading(false);
    }
  };

  const login = async (username: string, password: string): Promise<boolean> => {
    try {
      setLoading(true);
      
      if (typeof window !== 'undefined' && window.__TAURI__) {
        const response = await invoke('login', { username, password });
        
        if (response) {
          const userInfo = await invoke('get_user_info');
          setUser(userInfo as User);
          return true;
        }
      } else {
        // Browser mode - simulate login
        if (username === 'demo' && password === 'demo') {
          setUser({ id: '1', username: 'demo', email: 'demo@example.com' });
          return true;
        }
      }
      return false;
    } catch (error) {
      console.error('Login failed:', error);
      return false;
    } finally {
      setLoading(false);
    }
  };

  const logout = async (): Promise<void> => {
    try {
      if (typeof window !== 'undefined' && window.__TAURI__) {
        await invoke('logout');
      }
      setUser(null);
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  const register = async (username: string, email: string, password: string): Promise<boolean> => {
    try {
      setLoading(true);
      const response = await invoke('register', { username, email, password });
      return response as boolean;
    } catch (error) {
      console.error('Registration failed:', error);
      return false;
    } finally {
      setLoading(false);
    }
  };

  const value: AuthContextType = {
    user,
    isAuthenticated: !!user,
    loading,
    login,
    logout,
    register,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}