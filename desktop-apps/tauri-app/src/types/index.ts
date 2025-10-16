// User types
export interface User {
  id: string;
  username: string;
  email: string;
  avatar?: string;
  role: 'admin' | 'user';
  created_at: string;
  last_login?: string;
}

// Authentication types
export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

export interface AuthResponse {
  success: boolean;
  user?: User;
  token?: string;
  message?: string;
}

// Chat types
export interface Message {
  id: string;
  content: string;
  role: 'user' | 'assistant' | 'system';
  timestamp: string;
  session_id: string;
  metadata?: Record<string, any>;
}

export interface ChatSession {
  id: string;
  title: string;
  created_at: string;
  updated_at: string;
  message_count: number;
  chat_type: ChatType;
}

export type ChatType = 'general' | 'code' | 'creative' | 'analysis';

// Document types
export interface DocumentInfo {
  name: string;
  size: number;
  type: string;
  path: string;
  last_modified: string;
}

export interface DocumentOperation {
  id: string;
  name: string;
  description: string;
  icon: string;
}

export interface ProcessingResult {
  operation: string;
  result: string;
  metadata?: Record<string, any>;
  timestamp: string;
}

// Image types
export interface ImageGenerationRequest {
  prompt: string;
  style?: string;
  width?: number;
  height?: number;
  quality?: 'standard' | 'hd';
  n?: number;
}

export interface ImageAnalysisResult {
  description: string;
  objects: string[];
  colors: string[];
  mood: string;
  confidence: number;
}

export interface ImageEditRequest {
  image_path: string;
  prompt: string;
  mask_path?: string;
}

export interface GeneratedImage {
  id: string;
  url: string;
  prompt: string;
  style: string;
  dimensions: string;
  created_at: string;
}

// System types
export interface SystemStatus {
  cpu_usage: number;
  memory_usage: number;
  disk_usage: number;
  network_status: 'connected' | 'disconnected';
  ai_service_status: 'running' | 'stopped';
  database_status: 'connected' | 'disconnected';
}

export interface AppSettings {
  language: string;
  autoSave: boolean;
  notifications: boolean;
  aiModel: string;
  maxTokens: number;
  temperature: number;
  apiEndpoint: string;
  encryptionEnabled: boolean;
  backupEnabled: boolean;
  backupInterval: number;
}

// Storage types
export interface FileInfo {
  id: string;
  name: string;
  path: string;
  size: number;
  mime_type: string;
  created_at: string;
  updated_at: string;
  checksum: string;
}

export interface StorageStats {
  total_files: number;
  total_size: number;
  available_space: number;
  backup_count: number;
  last_backup: string;
}

// Security types
export interface SessionInfo {
  token: string;
  user_id: string;
  expires_at: string;
  permissions: string[];
}

export interface EncryptionResult {
  encrypted_data: string;
  iv: string;
  salt: string;
}

// API Response types
export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

// Theme types
export type Theme = 'light' | 'dark' | 'system';

// Component props types
export interface BaseComponentProps {
  className?: string;
  children?: React.ReactNode;
}

// Error types
export interface AppError {
  code: string;
  message: string;
  details?: Record<string, any>;
  timestamp: string;
}

// Navigation types
export interface NavItem {
  id: string;
  label: string;
  path: string;
  icon: React.ComponentType<any>;
  badge?: string | number;
}

// Form types
export interface FormField {
  name: string;
  label: string;
  type: 'text' | 'email' | 'password' | 'number' | 'select' | 'textarea' | 'checkbox';
  placeholder?: string;
  required?: boolean;
  options?: { value: string; label: string }[];
  validation?: {
    min?: number;
    max?: number;
    pattern?: string;
    message?: string;
  };
}

// Notification types
export interface Notification {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  message: string;
  timestamp: string;
  read: boolean;
  actions?: {
    label: string;
    action: () => void;
  }[];
}

// All types are already exported above with their definitions