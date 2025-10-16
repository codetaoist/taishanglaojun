// User related type definitions

export interface User {
  id: string;
  username: string;
  email: string;
  display_name?: string;
  first_name?: string;
  last_name?: string;
  avatar?: string;
  avatar_url?: string;
  role: string;
  roles?: string[];
  permissions?: string[];
  status?: 'active' | 'inactive' | 'suspended' | 'deleted';
  phone?: string;
  birth_date?: string;
  gender?: 'male' | 'female' | 'other' | 'prefer_not_to_say';
  language?: string;
  timezone?: string;
  bio?: string;
  location?: UserLocation;
  preferences?: UserPreferences;
  statistics?: UserStatistics;
  email_verified?: boolean;
  phone_verified?: boolean;
  two_factor_enabled?: boolean;
  last_login?: string;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
}

export interface UserLocation {
  country?: string;
  province?: string;
  city?: string;
  latitude?: number;
  longitude?: number;
}

export interface UserPreferences {
  theme?: 'light' | 'dark' | 'auto';
  notifications?: NotificationPreferences;
  privacy?: PrivacyPreferences;
  learning?: LearningPreferences;
}

export interface NotificationPreferences {
  email?: boolean;
  push?: boolean;
  sms?: boolean;
  learning_reminders?: boolean;
  social_updates?: boolean;
  system_announcements?: boolean;
}

export interface PrivacyPreferences {
  profile_visibility?: 'public' | 'friends' | 'private';
  activity_visibility?: 'public' | 'friends' | 'private';
  search_visibility?: boolean;
  data_sharing?: boolean;
}

export interface LearningPreferences {
  difficulty_preference?: 'beginner' | 'intermediate' | 'advanced';
  learning_style?: string;
  preferred_topics?: string[];
}

export interface UserStatistics {
  total_learning_hours?: number;
  completed_courses?: number;
  achievement_points?: number;
  social_connections?: number;
  login_count?: number;
  message_count?: number;
  image_count?: number;
  document_count?: number;
  storage_used?: number;
  api_calls_today?: number;
  api_calls_month?: number;
}

export interface AuthUser extends User {
  token?: string;
  refresh_token?: string;
  expires_at?: string;
}

export interface LoginRequest {
  username: string;
  password: string;
  remember_me?: boolean;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  display_name?: string;
  first_name?: string;
  last_name?: string;
}

export interface UpdateUserRequest {
  username?: string;
  display_name?: string;
  first_name?: string;
  last_name?: string;
  avatar?: string;
  bio?: string;
  phone?: string;
  location?: UserLocation;
  preferences?: UserPreferences;
}

export type UserRole = 'GUEST' | 'USER' | 'PREMIUM' | 'ADMIN' | 'SUPER_ADMIN';

export interface UserPermission {
  id: string;
  name: string;
  description?: string;
  resource: string;
  action: string;
}

export interface UserResponse {
  success: boolean;
  data: User;
  message?: string;
}

export interface UsersResponse {
  success: boolean;
  data: User[];
  pagination?: {
    page: number;
    limit: number;
    total: number;
    pages: number;
  };
  message?: string;
}
