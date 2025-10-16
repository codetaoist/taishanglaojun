// 数据库连接配置相关类型定义

export enum DatabaseType {
  MYSQL = 'mysql',
  POSTGRESQL = 'postgresql',
  SQLITE = 'sqlite',
  MONGODB = 'mongodb',
  REDIS = 'redis',
  ORACLE = 'oracle',
  SQLSERVER = 'sqlserver',
  MARIADB = 'mariadb'
}

export enum ConnectionStatus {
  CONNECTED = 'connected',
  DISCONNECTED = 'disconnected',
  CONNECTING = 'connecting',
  ERROR = 'error',
  UNKNOWN = 'unknown'
}

export interface DatabaseConnectionConfig {
  id: string;
  name: string;
  type: DatabaseType;
  host: string;
  port: number;
  database: string;
  username: string;
  password: string; // 前端显示时会被加密或隐藏
  ssl?: boolean;
  connectionTimeout?: number;
  maxConnections?: number;
  description?: string;
  tags?: string[];
  isDefault?: boolean;
  createdAt: string;
  updatedAt: string;
  createdBy: string;
  lastConnectedAt?: string;
}

export interface DatabaseConnectionStatus {
  id: string;
  status: ConnectionStatus;
  lastChecked: string;
  responseTime?: number;
  errorMessage?: string;
  serverVersion?: string;
  databaseSize?: string;
  activeConnections?: number;
}

export interface DatabaseConnectionTest {
  success: boolean;
  responseTime: number;
  errorMessage?: string;
  serverInfo?: {
    version: string;
    charset?: string;
    timezone?: string;
  };
}

export interface DatabaseConnectionForm {
  name: string;
  type: DatabaseType;
  host: string;
  port: number;
  database: string;
  username: string;
  password: string;
  ssl?: boolean;
  connectionTimeout?: number;
  maxConnections?: number;
  description?: string;
  tags?: string[];
  isDefault?: boolean;
}

export interface DatabaseConnectionListItem {
  id: string;
  name: string;
  type: DatabaseType;
  host: string;
  port: number;
  database: string;
  description?: string;
  tags?: string[];
  isDefault?: boolean;
  status: ConnectionStatus;
  lastConnectedAt?: string;
  createdAt: string;
  createdBy: string;
}

export interface DatabaseConnectionStats {
  totalConnections: number;
  activeConnections: number;
  connectionsByType: Record<DatabaseType, number>;
  connectionsByStatus: Record<ConnectionStatus, number>;
  averageResponseTime: number;
  lastUpdated: string;
}

// API 响应类型
export interface DatabaseConnectionResponse {
  success: boolean;
  data?: DatabaseConnectionConfig;
  message?: string;
}

export interface DatabaseConnectionListResponse {
  success: boolean;
  data?: {
    connections: DatabaseConnectionListItem[];
    total: number;
    page: number;
    pageSize: number;
  };
  message?: string;
}

export interface DatabaseConnectionStatusResponse {
  success: boolean;
  data?: DatabaseConnectionStatus[];
  message?: string;
}

export interface DatabaseConnectionTestResponse {
  success: boolean;
  data?: DatabaseConnectionTest;
  message?: string;
}

export interface DatabaseConnectionStatsResponse {
  success: boolean;
  data?: DatabaseConnectionStats;
  message?: string;
}

// 查询参数类型
export interface DatabaseConnectionQuery {
  page?: number;
  pageSize?: number;
  search?: string;
  type?: DatabaseType;
  status?: ConnectionStatus;
  tags?: string[];
  sortBy?: 'name' | 'type' | 'createdAt' | 'lastConnectedAt';
  sortOrder?: 'asc' | 'desc';
}

// 数据库连接操作类型
export enum DatabaseConnectionAction {
  CREATE = 'create',
  UPDATE = 'update',
  DELETE = 'delete',
  TEST = 'test',
  CONNECT = 'connect',
  DISCONNECT = 'disconnect',
  REFRESH_STATUS = 'refresh_status'
}

// 数据库连接事件类型
export interface DatabaseConnectionEvent {
  type: DatabaseConnectionAction;
  connectionId: string;
  timestamp: string;
  success: boolean;
  message?: string;
  details?: any;
}

// 数据库类型配置
export interface DatabaseTypeConfig {
  type: DatabaseType;
  name: string;
  icon: string;
  defaultPort: number;
  supportsSsl: boolean;
  supportsConnectionPool: boolean;
  connectionStringTemplate: string;
  description: string;
  documentationUrl?: string;
}

// 预定义的数据库类型配置
export const DATABASE_TYPE_CONFIGS: Record<DatabaseType, DatabaseTypeConfig> = {
  [DatabaseType.MYSQL]: {
    type: DatabaseType.MYSQL,
    name: 'MySQL',
    icon: 'mysql',
    defaultPort: 3306,
    supportsSsl: true,
    supportsConnectionPool: true,
    connectionStringTemplate: 'mysql://{username}:{password}@{host}:{port}/{database}',
    description: 'MySQL 关系型数据库',
    documentationUrl: 'https://dev.mysql.com/doc/'
  },
  [DatabaseType.POSTGRESQL]: {
    type: DatabaseType.POSTGRESQL,
    name: 'PostgreSQL',
    icon: 'postgresql',
    defaultPort: 5432,
    supportsSsl: true,
    supportsConnectionPool: true,
    connectionStringTemplate: 'postgresql://{username}:{password}@{host}:{port}/{database}',
    description: 'PostgreSQL 关系型数据库',
    documentationUrl: 'https://www.postgresql.org/docs/'
  },
  [DatabaseType.SQLITE]: {
    type: DatabaseType.SQLITE,
    name: 'SQLite',
    icon: 'sqlite',
    defaultPort: 0,
    supportsSsl: false,
    supportsConnectionPool: false,
    connectionStringTemplate: 'sqlite://{database}',
    description: 'SQLite 轻量级数据库',
    documentationUrl: 'https://www.sqlite.org/docs.html'
  },
  [DatabaseType.MONGODB]: {
    type: DatabaseType.MONGODB,
    name: 'MongoDB',
    icon: 'mongodb',
    defaultPort: 27017,
    supportsSsl: true,
    supportsConnectionPool: true,
    connectionStringTemplate: 'mongodb://{username}:{password}@{host}:{port}/{database}',
    description: 'MongoDB 文档数据库',
    documentationUrl: 'https://docs.mongodb.com/'
  },
  [DatabaseType.REDIS]: {
    type: DatabaseType.REDIS,
    name: 'Redis',
    icon: 'redis',
    defaultPort: 6379,
    supportsSsl: true,
    supportsConnectionPool: true,
    connectionStringTemplate: 'redis://{username}:{password}@{host}:{port}/{database}',
    description: 'Redis 内存数据库',
    documentationUrl: 'https://redis.io/documentation'
  },
  [DatabaseType.ORACLE]: {
    type: DatabaseType.ORACLE,
    name: 'Oracle',
    icon: 'oracle',
    defaultPort: 1521,
    supportsSsl: true,
    supportsConnectionPool: true,
    connectionStringTemplate: 'oracle://{username}:{password}@{host}:{port}/{database}',
    description: 'Oracle 企业级数据库',
    documentationUrl: 'https://docs.oracle.com/database/'
  },
  [DatabaseType.SQLSERVER]: {
    type: DatabaseType.SQLSERVER,
    name: 'SQL Server',
    icon: 'sqlserver',
    defaultPort: 1433,
    supportsSsl: true,
    supportsConnectionPool: true,
    connectionStringTemplate: 'sqlserver://{username}:{password}@{host}:{port}/{database}',
    description: 'Microsoft SQL Server',
    documentationUrl: 'https://docs.microsoft.com/sql/'
  },
  [DatabaseType.MARIADB]: {
    type: DatabaseType.MARIADB,
    name: 'MariaDB',
    icon: 'mariadb',
    defaultPort: 3306,
    supportsSsl: true,
    supportsConnectionPool: true,
    connectionStringTemplate: 'mariadb://{username}:{password}@{host}:{port}/{database}',
    description: 'MariaDB 关系型数据库',
    documentationUrl: 'https://mariadb.org/documentation/'
  }
};