import relationalStore from '@ohos.data.relationalStore';
import { BusinessError } from '@ohos.base';
import hilog from '@ohos.hilog';
import { LocationPoint } from '../model/LocationPoint';
import { Trajectory, TrajectoryStatus } from '../model/Trajectory';

/**
 * 数据库配置
 */
const DB_CONFIG = {
  name: 'TaishanglaojunTracker.db',
  version: 1,
  encrypt: false
};

/**
 * 数据服务类
 */
export class DataService {
  private static instance: DataService;
  private rdbStore?: relationalStore.RdbStore;
  private isInitialized: boolean = false;

  private constructor() {}

  /**
   * 获取单例实例
   */
  static getInstance(): DataService {
    if (!DataService.instance) {
      DataService.instance = new DataService();
    }
    return DataService.instance;
  }

  /**
   * 初始化数据库
   */
  async initialize(context: Context): Promise<void> {
    if (this.isInitialized) {
      return;
    }

    try {
      const config: relationalStore.StoreConfig = {
        name: DB_CONFIG.name,
        securityLevel: relationalStore.SecurityLevel.S1,
        encrypt: DB_CONFIG.encrypt
      };

      this.rdbStore = await relationalStore.getRdbStore(context, config);
      await this.createTables();
      this.isInitialized = true;
      
      hilog.info(0x0000, 'DataService', 'Database initialized successfully');
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Database initialization error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 创建数据表
   */
  private async createTables(): Promise<void> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    // 创建轨迹表
    const createTrajectoryTable = `
      CREATE TABLE IF NOT EXISTS trajectories (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        description TEXT,
        start_time INTEGER NOT NULL,
        end_time INTEGER,
        status TEXT NOT NULL,
        total_distance REAL DEFAULT 0,
        duration INTEGER DEFAULT 0,
        max_speed REAL DEFAULT 0,
        avg_speed REAL DEFAULT 0,
        point_count INTEGER DEFAULT 0,
        min_latitude REAL DEFAULT 90,
        max_latitude REAL DEFAULT -90,
        min_longitude REAL DEFAULT 180,
        max_longitude REAL DEFAULT -180,
        created_at INTEGER NOT NULL,
        updated_at INTEGER NOT NULL,
        is_synced INTEGER DEFAULT 0,
        user_id TEXT,
        tags TEXT,
        color TEXT DEFAULT '#007AFF'
      )
    `;

    // 创建位置点表
    const createLocationPointTable = `
      CREATE TABLE IF NOT EXISTS location_points (
        id TEXT PRIMARY KEY,
        latitude REAL NOT NULL,
        longitude REAL NOT NULL,
        timestamp INTEGER NOT NULL,
        accuracy REAL,
        altitude REAL,
        speed REAL,
        bearing REAL,
        trajectory_id TEXT,
        provider TEXT,
        is_mock INTEGER DEFAULT 0,
        battery_level REAL,
        network_type TEXT,
        FOREIGN KEY (trajectory_id) REFERENCES trajectories (id) ON DELETE CASCADE
      )
    `;

    // 创建AI对话表
    const createConversationsTable = `
      CREATE TABLE IF NOT EXISTS conversations (
        conversationId TEXT PRIMARY KEY,
        title TEXT NOT NULL,
        aiPersonality TEXT NOT NULL,
        createdAt INTEGER NOT NULL,
        updatedAt INTEGER NOT NULL,
        messageCount INTEGER DEFAULT 0,
        isArchived INTEGER DEFAULT 0,
        is_synced INTEGER DEFAULT 0
      )
    `;

    // 创建AI消息表
    const createChatMessagesTable = `
      CREATE TABLE IF NOT EXISTS chat_messages (
        messageId TEXT PRIMARY KEY,
        conversationId TEXT NOT NULL,
        content TEXT NOT NULL,
        messageType TEXT NOT NULL,
        sender TEXT NOT NULL,
        status TEXT NOT NULL,
        timestamp INTEGER NOT NULL,
        isRead INTEGER DEFAULT 0,
        metadata TEXT,
        error TEXT,
        is_synced INTEGER DEFAULT 0,
        FOREIGN KEY (conversationId) REFERENCES conversations (conversationId) ON DELETE CASCADE
      )
    `;

    // 创建索引
    const createIndexes = [
      'CREATE INDEX IF NOT EXISTS idx_trajectory_start_time ON trajectories (start_time)',
      'CREATE INDEX IF NOT EXISTS idx_trajectory_status ON trajectories (status)',
      'CREATE INDEX IF NOT EXISTS idx_location_trajectory_id ON location_points (trajectory_id)',
      'CREATE INDEX IF NOT EXISTS idx_location_timestamp ON location_points (timestamp)',
      'CREATE INDEX IF NOT EXISTS idx_conversations_updated_at ON conversations (updatedAt)',
      'CREATE INDEX IF NOT EXISTS idx_conversations_archived ON conversations (isArchived)',
      'CREATE INDEX IF NOT EXISTS idx_conversations_synced ON conversations (is_synced)',
      'CREATE INDEX IF NOT EXISTS idx_chat_messages_conversation_id ON chat_messages (conversationId)',
      'CREATE INDEX IF NOT EXISTS idx_chat_messages_timestamp ON chat_messages (timestamp)',
      'CREATE INDEX IF NOT EXISTS idx_chat_messages_status ON chat_messages (status)',
      'CREATE INDEX IF NOT EXISTS idx_chat_messages_sender ON chat_messages (sender)',
      'CREATE INDEX IF NOT EXISTS idx_chat_messages_content ON chat_messages (content)',
      'CREATE INDEX IF NOT EXISTS idx_chat_messages_synced ON chat_messages (is_synced)'
    ];

    try {
      await this.rdbStore.executeSql(createTrajectoryTable);
      await this.rdbStore.executeSql(createLocationPointTable);
      await this.rdbStore.executeSql(createConversationsTable);
      await this.rdbStore.executeSql(createChatMessagesTable);
      
      for (const indexSql of createIndexes) {
        await this.rdbStore.executeSql(indexSql);
      }
      
      hilog.info(0x0000, 'DataService', 'Database tables created successfully');
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Create tables error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 保存轨迹
   */
  async saveTrajectory(trajectory: Trajectory): Promise<void> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      const valueBucket: relationalStore.ValuesBucket = {
        id: trajectory.id,
        name: trajectory.name,
        description: trajectory.description || null,
        start_time: trajectory.startTime,
        end_time: trajectory.endTime || null,
        status: trajectory.status,
        total_distance: trajectory.totalDistance,
        duration: trajectory.duration,
        max_speed: trajectory.maxSpeed,
        avg_speed: trajectory.avgSpeed,
        point_count: trajectory.pointCount,
        min_latitude: trajectory.minLatitude,
        max_latitude: trajectory.maxLatitude,
        min_longitude: trajectory.minLongitude,
        max_longitude: trajectory.maxLongitude,
        created_at: trajectory.createdAt,
        updated_at: trajectory.updatedAt,
        is_synced: trajectory.isSynced ? 1 : 0,
        user_id: trajectory.userId || null,
        tags: JSON.stringify(trajectory.tags),
        color: trajectory.color
      };

      await this.rdbStore.insert('trajectories', valueBucket);
      hilog.info(0x0000, 'DataService', 'Trajectory saved: %{public}s', trajectory.id);
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Save trajectory error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 更新轨迹
   */
  async updateTrajectory(trajectory: Trajectory): Promise<void> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      trajectory.updatedAt = Date.now();
      
      const valueBucket: relationalStore.ValuesBucket = {
        name: trajectory.name,
        description: trajectory.description || null,
        end_time: trajectory.endTime || null,
        status: trajectory.status,
        total_distance: trajectory.totalDistance,
        duration: trajectory.duration,
        max_speed: trajectory.maxSpeed,
        avg_speed: trajectory.avgSpeed,
        point_count: trajectory.pointCount,
        min_latitude: trajectory.minLatitude,
        max_latitude: trajectory.maxLatitude,
        min_longitude: trajectory.minLongitude,
        max_longitude: trajectory.maxLongitude,
        updated_at: trajectory.updatedAt,
        is_synced: trajectory.isSynced ? 1 : 0,
        tags: JSON.stringify(trajectory.tags),
        color: trajectory.color
      };

      const predicates = new relationalStore.RdbPredicates('trajectories');
      predicates.equalTo('id', trajectory.id);

      await this.rdbStore.update(valueBucket, predicates);
      hilog.info(0x0000, 'DataService', 'Trajectory updated: %{public}s', trajectory.id);
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Update trajectory error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 获取所有轨迹
   */
  async getAllTrajectories(): Promise<Trajectory[]> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      const predicates = new relationalStore.RdbPredicates('trajectories');
      predicates.orderByDesc('start_time');

      const resultSet = await this.rdbStore.query(predicates);
      const trajectories: Trajectory[] = [];

      if (resultSet.goToFirstRow()) {
        do {
          const trajectory = this.parseTrajectoryFromResultSet(resultSet);
          trajectories.push(trajectory);
        } while (resultSet.goToNextRow());
      }

      resultSet.close();
      return trajectories;
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Get all trajectories error: %{public}s', JSON.stringify(error));
      return [];
    }
  }

  /**
   * 根据ID获取轨迹
   */
  async getTrajectoryById(id: string): Promise<Trajectory | null> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      const predicates = new relationalStore.RdbPredicates('trajectories');
      predicates.equalTo('id', id);

      const resultSet = await this.rdbStore.query(predicates);
      let trajectory: Trajectory | null = null;

      if (resultSet.goToFirstRow()) {
        trajectory = this.parseTrajectoryFromResultSet(resultSet);
      }

      resultSet.close();
      return trajectory;
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Get trajectory by id error: %{public}s', JSON.stringify(error));
      return null;
    }
  }

  /**
   * 删除轨迹
   */
  async deleteTrajectory(id: string): Promise<void> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      const predicates = new relationalStore.RdbPredicates('trajectories');
      predicates.equalTo('id', id);

      await this.rdbStore.delete(predicates);
      hilog.info(0x0000, 'DataService', 'Trajectory deleted: %{public}s', id);
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Delete trajectory error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 保存位置点
   */
  async saveLocationPoint(point: LocationPoint): Promise<void> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      if (!point.id) {
        point.id = this.generateLocationPointId();
      }

      const valueBucket: relationalStore.ValuesBucket = {
        id: point.id,
        latitude: point.latitude,
        longitude: point.longitude,
        timestamp: point.timestamp,
        accuracy: point.accuracy || null,
        altitude: point.altitude || null,
        speed: point.speed || null,
        bearing: point.bearing || null,
        trajectory_id: point.trajectoryId || null,
        provider: point.provider || null,
        is_mock: point.isMock ? 1 : 0,
        battery_level: point.batteryLevel || null,
        network_type: point.networkType || null
      };

      await this.rdbStore.insert('location_points', valueBucket);
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Save location point error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 获取轨迹的位置点
   */
  async getLocationPoints(trajectoryId: string): Promise<LocationPoint[]> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      const predicates = new relationalStore.RdbPredicates('location_points');
      predicates.equalTo('trajectory_id', trajectoryId);
      predicates.orderByAsc('timestamp');

      const resultSet = await this.rdbStore.query(predicates);
      const points: LocationPoint[] = [];

      if (resultSet.goToFirstRow()) {
        do {
          const point = this.parseLocationPointFromResultSet(resultSet);
          points.push(point);
        } while (resultSet.goToNextRow());
      }

      resultSet.close();
      return points;
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Get location points error: %{public}s', JSON.stringify(error));
      return [];
    }
  }

  /**
   * 清除所有数据
   */
  async clearAllData(): Promise<void> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      await this.rdbStore.executeSql('DELETE FROM location_points');
      await this.rdbStore.executeSql('DELETE FROM trajectories');
      hilog.info(0x0000, 'DataService', 'All data cleared');
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Clear all data error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 获取统计信息
   */
  async getStatistics(): Promise<{
    totalTrajectories: number;
    totalDistance: number;
    totalDuration: number;
    totalPoints: number;
  }> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      const sql = `
        SELECT 
          COUNT(*) as total_trajectories,
          SUM(total_distance) as total_distance,
          SUM(duration) as total_duration,
          SUM(point_count) as total_points
        FROM trajectories 
        WHERE status = ?
      `;

      const resultSet = await this.rdbStore.querySql(sql, [TrajectoryStatus.FINISHED]);
      
      let stats = {
        totalTrajectories: 0,
        totalDistance: 0,
        totalDuration: 0,
        totalPoints: 0
      };

      if (resultSet.goToFirstRow()) {
        stats = {
          totalTrajectories: resultSet.getLong(resultSet.getColumnIndex('total_trajectories')),
          totalDistance: resultSet.getDouble(resultSet.getColumnIndex('total_distance')),
          totalDuration: resultSet.getLong(resultSet.getColumnIndex('total_duration')),
          totalPoints: resultSet.getLong(resultSet.getColumnIndex('total_points'))
        };
      }

      resultSet.close();
      return stats;
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Get statistics error: %{public}s', JSON.stringify(error));
      return {
        totalTrajectories: 0,
        totalDistance: 0,
        totalDuration: 0,
        totalPoints: 0
      };
    }
  }

  /**
   * 解析轨迹数据
   */
  private parseTrajectoryFromResultSet(resultSet: relationalStore.ResultSet): Trajectory {
    const trajectory = new Trajectory(
      resultSet.getString(resultSet.getColumnIndex('name')),
      resultSet.getString(resultSet.getColumnIndex('description')),
      {
        id: resultSet.getString(resultSet.getColumnIndex('id')),
        userId: resultSet.getString(resultSet.getColumnIndex('user_id')),
        tags: JSON.parse(resultSet.getString(resultSet.getColumnIndex('tags')) || '[]'),
        color: resultSet.getString(resultSet.getColumnIndex('color'))
      }
    );

    trajectory.startTime = resultSet.getLong(resultSet.getColumnIndex('start_time'));
    trajectory.endTime = resultSet.getLong(resultSet.getColumnIndex('end_time')) || undefined;
    trajectory.status = resultSet.getString(resultSet.getColumnIndex('status')) as TrajectoryStatus;
    trajectory.totalDistance = resultSet.getDouble(resultSet.getColumnIndex('total_distance'));
    trajectory.duration = resultSet.getLong(resultSet.getColumnIndex('duration'));
    trajectory.maxSpeed = resultSet.getDouble(resultSet.getColumnIndex('max_speed'));
    trajectory.avgSpeed = resultSet.getDouble(resultSet.getColumnIndex('avg_speed'));
    trajectory.pointCount = resultSet.getLong(resultSet.getColumnIndex('point_count'));
    trajectory.minLatitude = resultSet.getDouble(resultSet.getColumnIndex('min_latitude'));
    trajectory.maxLatitude = resultSet.getDouble(resultSet.getColumnIndex('max_latitude'));
    trajectory.minLongitude = resultSet.getDouble(resultSet.getColumnIndex('min_longitude'));
    trajectory.maxLongitude = resultSet.getDouble(resultSet.getColumnIndex('max_longitude'));
    trajectory.createdAt = resultSet.getLong(resultSet.getColumnIndex('created_at'));
    trajectory.updatedAt = resultSet.getLong(resultSet.getColumnIndex('updated_at'));
    trajectory.isSynced = resultSet.getLong(resultSet.getColumnIndex('is_synced')) === 1;

    return trajectory;
  }

  /**
   * 解析位置点数据
   */
  private parseLocationPointFromResultSet(resultSet: relationalStore.ResultSet): LocationPoint {
    return new LocationPoint(
      resultSet.getDouble(resultSet.getColumnIndex('latitude')),
      resultSet.getDouble(resultSet.getColumnIndex('longitude')),
      resultSet.getLong(resultSet.getColumnIndex('timestamp')),
      {
        id: resultSet.getString(resultSet.getColumnIndex('id')),
        accuracy: resultSet.getDouble(resultSet.getColumnIndex('accuracy')) || undefined,
        altitude: resultSet.getDouble(resultSet.getColumnIndex('altitude')) || undefined,
        speed: resultSet.getDouble(resultSet.getColumnIndex('speed')) || undefined,
        bearing: resultSet.getDouble(resultSet.getColumnIndex('bearing')) || undefined,
        trajectoryId: resultSet.getString(resultSet.getColumnIndex('trajectory_id')) || undefined,
        provider: resultSet.getString(resultSet.getColumnIndex('provider')) || undefined,
        isMock: resultSet.getLong(resultSet.getColumnIndex('is_mock')) === 1,
        batteryLevel: resultSet.getDouble(resultSet.getColumnIndex('battery_level')) || undefined,
        networkType: resultSet.getString(resultSet.getColumnIndex('network_type')) || undefined
      }
    );
  }

  /**
   * 生成位置点ID
   */
  private generateLocationPointId(): string {
    return 'point_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
  }

  /**
   * 关闭数据库连接
   */
  async close(): Promise<void> {
    if (this.rdbStore) {
      // 注意：HarmonyOS的RdbStore没有显式的close方法
      // 数据库连接会在应用退出时自动关闭
      this.rdbStore = undefined;
      this.isInitialized = false;
      hilog.info(0x0000, 'DataService', 'Database connection closed');
    }
  }

  // ==================== AI对话功能数据操作方法 ====================

  /**
   * 创建新对话
   */
  async createConversation(conversation: ConversationEntity): Promise<void> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      const sql = `
        INSERT INTO conversations (
          conversationId, title, aiPersonality, createdAt, updatedAt, messageCount, isArchived
        ) VALUES (?, ?, ?, ?, ?, ?, ?)
      `;

      await this.rdbStore.executeSql(sql, [
        conversation.conversationId,
        conversation.title,
        conversation.aiPersonality,
        conversation.createdAt,
        conversation.updatedAt,
        conversation.messageCount,
        conversation.isArchived ? 1 : 0
      ]);

      hilog.info(0x0000, 'DataService', 'Conversation created: %{public}s', conversation.conversationId);
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Create conversation error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 获取所有对话
   */
  async getConversations(includeArchived: boolean = false): Promise<ConversationEntity[]> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      let sql = 'SELECT * FROM conversations';
      const params: any[] = [];

      if (!includeArchived) {
        sql += ' WHERE isArchived = ?';
        params.push(0);
      }

      sql += ' ORDER BY updatedAt DESC';

      const resultSet = await this.rdbStore.querySql(sql, params);
      const conversations: ConversationEntity[] = [];

      while (resultSet.goToNextRow()) {
        conversations.push(this.parseConversationFromResultSet(resultSet));
      }

      resultSet.close();
      return conversations;
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Get conversations error: %{public}s', JSON.stringify(error));
      return [];
    }
  }

  /**
   * 获取单个对话
   */
  async getConversation(conversationId: string): Promise<ConversationEntity | null> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      const sql = 'SELECT * FROM conversations WHERE conversationId = ?';
      const resultSet = await this.rdbStore.querySql(sql, [conversationId]);

      let conversation: ConversationEntity | null = null;
      if (resultSet.goToFirstRow()) {
        conversation = this.parseConversationFromResultSet(resultSet);
      }

      resultSet.close();
      return conversation;
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Get conversation error: %{public}s', JSON.stringify(error));
      return null;
    }
  }

  /**
   * 更新对话
   */
  async updateConversation(conversation: ConversationEntity): Promise<void> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      const sql = `
        UPDATE conversations SET 
          title = ?, aiPersonality = ?, updatedAt = ?, messageCount = ?, isArchived = ?
        WHERE conversationId = ?
      `;

      await this.rdbStore.executeSql(sql, [
        conversation.title,
        conversation.aiPersonality,
        conversation.updatedAt,
        conversation.messageCount,
        conversation.isArchived ? 1 : 0,
        conversation.conversationId
      ]);

      hilog.info(0x0000, 'DataService', 'Conversation updated: %{public}s', conversation.conversationId);
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Update conversation error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 删除对话
   */
  async deleteConversation(conversationId: string): Promise<void> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      // 先删除相关消息
      await this.rdbStore.executeSql('DELETE FROM chat_messages WHERE conversationId = ?', [conversationId]);
      // 再删除对话
      await this.rdbStore.executeSql('DELETE FROM conversations WHERE conversationId = ?', [conversationId]);

      hilog.info(0x0000, 'DataService', 'Conversation deleted: %{public}s', conversationId);
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Delete conversation error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 归档对话
   */
  async archiveConversation(conversationId: string): Promise<void> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      const sql = 'UPDATE conversations SET isArchived = 1, updatedAt = ? WHERE conversationId = ?';
      await this.rdbStore.executeSql(sql, [Date.now(), conversationId]);

      hilog.info(0x0000, 'DataService', 'Conversation archived: %{public}s', conversationId);
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Archive conversation error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 保存聊天消息
   */
  async saveChatMessage(message: ChatMessageEntity): Promise<void> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      const sql = `
        INSERT INTO chat_messages (
          messageId, conversationId, content, messageType, sender, status, 
          timestamp, isRead, metadata, error
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
      `;

      await this.rdbStore.executeSql(sql, [
        message.messageId,
        message.conversationId,
        message.content,
        message.messageType,
        message.sender,
        message.status,
        message.timestamp,
        message.isRead ? 1 : 0,
        message.metadata ? JSON.stringify(message.metadata) : null,
        message.error
      ]);

      // 更新对话的消息计数和更新时间
      await this.updateConversationMessageCount(message.conversationId);

      hilog.info(0x0000, 'DataService', 'Chat message saved: %{public}s', message.messageId);
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Save chat message error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 获取对话的消息列表
   */
  async getChatMessages(conversationId: string, limit?: number, offset?: number): Promise<ChatMessageEntity[]> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      let sql = 'SELECT * FROM chat_messages WHERE conversationId = ? ORDER BY timestamp ASC';
      const params: any[] = [conversationId];

      if (limit !== undefined) {
        sql += ' LIMIT ?';
        params.push(limit);
        
        if (offset !== undefined) {
          sql += ' OFFSET ?';
          params.push(offset);
        }
      }

      const resultSet = await this.rdbStore.querySql(sql, params);
      const messages: ChatMessageEntity[] = [];

      while (resultSet.goToNextRow()) {
        messages.push(this.parseChatMessageFromResultSet(resultSet));
      }

      resultSet.close();
      return messages;
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Get chat messages error: %{public}s', JSON.stringify(error));
      return [];
    }
  }

  /**
   * 更新消息状态
   */
  async updateMessageStatus(messageId: string, status: string, error?: string): Promise<void> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      const sql = 'UPDATE chat_messages SET status = ?, error = ? WHERE messageId = ?';
      await this.rdbStore.executeSql(sql, [status, error || null, messageId]);

      hilog.info(0x0000, 'DataService', 'Message status updated: %{public}s -> %{public}s', messageId, status);
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Update message status error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 搜索消息
   */
  async searchMessages(query: string, conversationId?: string): Promise<ChatMessageEntity[]> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      let sql = 'SELECT * FROM chat_messages WHERE content LIKE ?';
      const params: any[] = [`%${query}%`];

      if (conversationId) {
        sql += ' AND conversationId = ?';
        params.push(conversationId);
      }

      sql += ' ORDER BY timestamp DESC LIMIT 50';

      const resultSet = await this.rdbStore.querySql(sql, params);
      const messages: ChatMessageEntity[] = [];

      while (resultSet.goToNextRow()) {
        messages.push(this.parseChatMessageFromResultSet(resultSet));
      }

      resultSet.close();
      return messages;
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Search messages error: %{public}s', JSON.stringify(error));
      return [];
    }
  }

  /**
   * 删除消息
   */
  async deleteChatMessage(messageId: string): Promise<void> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      // 获取消息的对话ID
      const messageResult = await this.rdbStore.querySql(
        'SELECT conversationId FROM chat_messages WHERE messageId = ?', 
        [messageId]
      );
      
      let conversationId: string | null = null;
      if (messageResult.goToFirstRow()) {
        conversationId = messageResult.getString(messageResult.getColumnIndex('conversationId'));
      }
      messageResult.close();

      // 删除消息
      await this.rdbStore.executeSql('DELETE FROM chat_messages WHERE messageId = ?', [messageId]);

      // 更新对话的消息计数
      if (conversationId) {
        await this.updateConversationMessageCount(conversationId);
      }

      hilog.info(0x0000, 'DataService', 'Chat message deleted: %{public}s', messageId);
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Delete chat message error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 更新对话的消息计数
   */
  private async updateConversationMessageCount(conversationId: string): Promise<void> {
    try {
      const countResult = await this.rdbStore!.querySql(
        'SELECT COUNT(*) as count FROM chat_messages WHERE conversationId = ?',
        [conversationId]
      );

      let messageCount = 0;
      if (countResult.goToFirstRow()) {
        messageCount = countResult.getLong(countResult.getColumnIndex('count'));
      }
      countResult.close();

      await this.rdbStore!.executeSql(
        'UPDATE conversations SET messageCount = ?, updatedAt = ? WHERE conversationId = ?',
        [messageCount, Date.now(), conversationId]
      );
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Update conversation message count error: %{public}s', JSON.stringify(error));
    }
  }

  /**
   * 解析对话数据
   */
  private parseConversationFromResultSet(resultSet: relationalStore.ResultSet): ConversationEntity {
    return {
      conversationId: resultSet.getString(resultSet.getColumnIndex('conversationId')),
      title: resultSet.getString(resultSet.getColumnIndex('title')),
      aiPersonality: resultSet.getString(resultSet.getColumnIndex('aiPersonality')),
      createdAt: resultSet.getLong(resultSet.getColumnIndex('createdAt')),
      updatedAt: resultSet.getLong(resultSet.getColumnIndex('updatedAt')),
      messageCount: resultSet.getLong(resultSet.getColumnIndex('messageCount')),
      isArchived: resultSet.getLong(resultSet.getColumnIndex('isArchived')) === 1
    };
  }

  /**
   * 解析聊天消息数据
   */
  private parseChatMessageFromResultSet(resultSet: relationalStore.ResultSet): ChatMessageEntity {
    const metadataStr = resultSet.getString(resultSet.getColumnIndex('metadata'));
    let metadata: Record<string, any> | undefined;
    
    if (metadataStr) {
      try {
        metadata = JSON.parse(metadataStr);
      } catch (e) {
        metadata = undefined;
      }
    }

    return {
      messageId: resultSet.getString(resultSet.getColumnIndex('messageId')),
      conversationId: resultSet.getString(resultSet.getColumnIndex('conversationId')),
      content: resultSet.getString(resultSet.getColumnIndex('content')),
      messageType: resultSet.getString(resultSet.getColumnIndex('messageType')),
      sender: resultSet.getString(resultSet.getColumnIndex('sender')),
      status: resultSet.getString(resultSet.getColumnIndex('status')),
      timestamp: resultSet.getLong(resultSet.getColumnIndex('timestamp')),
      isRead: resultSet.getLong(resultSet.getColumnIndex('isRead')) === 1,
      metadata: metadata,
      error: resultSet.getString(resultSet.getColumnIndex('error')) || undefined
    };
  }

  // ==================== 数据同步相关方法 ====================

  /**
   * 获取未同步的对话
   */
  async getUnsyncedConversations(): Promise<ConversationEntity[]> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      const predicates = new relationalStore.RdbPredicates('conversations');
      predicates.equalTo('is_synced', 0);
      predicates.orderByAsc('createdAt');

      const resultSet = await this.rdbStore.query(predicates);
      const conversations: ConversationEntity[] = [];

      if (resultSet.goToFirstRow()) {
        do {
          conversations.push(this.parseConversationFromResultSet(resultSet));
        } while (resultSet.goToNextRow());
      }

      resultSet.close();
      return conversations;
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Get unsynced conversations error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 获取未同步的消息
   */
  async getUnsyncedMessages(): Promise<ChatMessageEntity[]> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      const predicates = new relationalStore.RdbPredicates('chat_messages');
      predicates.equalTo('is_synced', 0);
      predicates.orderByAsc('timestamp');

      const resultSet = await this.rdbStore.query(predicates);
      const messages: ChatMessageEntity[] = [];

      if (resultSet.goToFirstRow()) {
        do {
          messages.push(this.parseChatMessageFromResultSet(resultSet));
        } while (resultSet.goToNextRow());
      }

      resultSet.close();
      return messages;
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Get unsynced messages error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 标记对话为已同步
   */
  async markConversationAsSynced(conversationId: string): Promise<void> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      const valueBucket: relationalStore.ValuesBucket = {
        is_synced: 1,
        updatedAt: Date.now()
      };

      const predicates = new relationalStore.RdbPredicates('conversations');
      predicates.equalTo('conversationId', conversationId);

      await this.rdbStore.update(valueBucket, predicates);
      hilog.info(0x0000, 'DataService', 'Conversation marked as synced: %{public}s', conversationId);
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Mark conversation as synced error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 标记消息为已同步
   */
  async markMessageAsSynced(messageId: string): Promise<void> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      const valueBucket: relationalStore.ValuesBucket = {
        is_synced: 1
      };

      const predicates = new relationalStore.RdbPredicates('chat_messages');
      predicates.equalTo('messageId', messageId);

      await this.rdbStore.update(valueBucket, predicates);
      hilog.info(0x0000, 'DataService', 'Message marked as synced: %{public}s', messageId);
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Mark message as synced error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 批量标记对话为已同步
   */
  async markConversationsAsSynced(conversationIds: string[]): Promise<void> {
    if (!this.rdbStore || conversationIds.length === 0) {
      return;
    }

    try {
      const valueBucket: relationalStore.ValuesBucket = {
        is_synced: 1,
        updatedAt: Date.now()
      };

      const predicates = new relationalStore.RdbPredicates('conversations');
      predicates.in('conversationId', conversationIds);

      await this.rdbStore.update(valueBucket, predicates);
      hilog.info(0x0000, 'DataService', 'Conversations marked as synced: %{public}d', conversationIds.length);
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Mark conversations as synced error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 批量标记消息为已同步
   */
  async markMessagesAsSynced(messageIds: string[]): Promise<void> {
    if (!this.rdbStore || messageIds.length === 0) {
      return;
    }

    try {
      const valueBucket: relationalStore.ValuesBucket = {
        is_synced: 1
      };

      const predicates = new relationalStore.RdbPredicates('chat_messages');
      predicates.in('messageId', messageIds);

      await this.rdbStore.update(valueBucket, predicates);
      hilog.info(0x0000, 'DataService', 'Messages marked as synced: %{public}d', messageIds.length);
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Mark messages as synced error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 获取同步统计信息
   */
  async getSyncStats(): Promise<{ unsyncedConversations: number, unsyncedMessages: number }> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      // 获取未同步对话数量
      const conversationPredicates = new relationalStore.RdbPredicates('conversations');
      conversationPredicates.equalTo('is_synced', 0);
      const conversationResultSet = await this.rdbStore.query(conversationPredicates, ['COUNT(*) as count']);
      
      let unsyncedConversations = 0;
      if (conversationResultSet.goToFirstRow()) {
        unsyncedConversations = conversationResultSet.getLong(0);
      }
      conversationResultSet.close();

      // 获取未同步消息数量
      const messagePredicates = new relationalStore.RdbPredicates('chat_messages');
      messagePredicates.equalTo('is_synced', 0);
      const messageResultSet = await this.rdbStore.query(messagePredicates, ['COUNT(*) as count']);
      
      let unsyncedMessages = 0;
      if (messageResultSet.goToFirstRow()) {
        unsyncedMessages = messageResultSet.getLong(0);
      }
      messageResultSet.close();

      return { unsyncedConversations, unsyncedMessages };
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Get sync stats error: %{public}s', JSON.stringify(error));
      throw error;
    }
  }

  /**
   * 获取所有聊天消息（用于导出）
   */
  async getAllChatMessages(options?: {
    startDate?: number;
    endDate?: number;
    conversationIds?: string[];
  }): Promise<ChatMessageEntity[]> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      let sql = 'SELECT * FROM chat_messages WHERE 1=1';
      const params: any[] = [];

      if (options?.startDate) {
        sql += ' AND timestamp >= ?';
        params.push(options.startDate);
      }

      if (options?.endDate) {
        sql += ' AND timestamp <= ?';
        params.push(options.endDate);
      }

      if (options?.conversationIds && options.conversationIds.length > 0) {
        const placeholders = options.conversationIds.map(() => '?').join(',');
        sql += ` AND conversationId IN (${placeholders})`;
        params.push(...options.conversationIds);
      }

      sql += ' ORDER BY timestamp ASC';

      const resultSet = await this.rdbStore.querySql(sql, params);
      const messages: ChatMessageEntity[] = [];

      while (resultSet.goToNextRow()) {
        messages.push(this.parseChatMessageFromResultSet(resultSet));
      }

      resultSet.close();
      return messages;
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Get all chat messages error: %{public}s', JSON.stringify(error));
      return [];
    }
  }

  /**
   * 获取轨迹数据（用于导出）
   */
  async getTrajectories(options?: {
    startDate?: number;
    endDate?: number;
    trajectoryIds?: string[];
  }): Promise<TrajectoryEntity[]> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      let sql = 'SELECT * FROM trajectories WHERE 1=1';
      const params: any[] = [];

      if (options?.startDate) {
        sql += ' AND start_time >= ?';
        params.push(options.startDate);
      }

      if (options?.endDate) {
        sql += ' AND start_time <= ?';
        params.push(options.endDate);
      }

      if (options?.trajectoryIds && options.trajectoryIds.length > 0) {
        const placeholders = options.trajectoryIds.map(() => '?').join(',');
        sql += ` AND trajectory_id IN (${placeholders})`;
        params.push(...options.trajectoryIds);
      }

      sql += ' ORDER BY start_time DESC';

      const resultSet = await this.rdbStore.querySql(sql, params);
      const trajectories: TrajectoryEntity[] = [];

      while (resultSet.goToNextRow()) {
        trajectories.push(this.parseTrajectoryFromResultSet(resultSet));
      }

      resultSet.close();
      return trajectories;
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Get trajectories error: %{public}s', JSON.stringify(error));
      return [];
    }
  }

  /**
   * 获取位置点数据（用于导出）
   */
  async getLocationPoints(options?: {
    startDate?: number;
    endDate?: number;
    trajectoryIds?: string[];
  }): Promise<LocationPointEntity[]> {
    if (!this.rdbStore) {
      throw new Error('Database not initialized');
    }

    try {
      let sql = 'SELECT * FROM location_points WHERE 1=1';
      const params: any[] = [];

      if (options?.startDate) {
        sql += ' AND timestamp >= ?';
        params.push(options.startDate);
      }

      if (options?.endDate) {
        sql += ' AND timestamp <= ?';
        params.push(options.endDate);
      }

      if (options?.trajectoryIds && options.trajectoryIds.length > 0) {
        const placeholders = options.trajectoryIds.map(() => '?').join(',');
        sql += ` AND trajectory_id IN (${placeholders})`;
        params.push(...options.trajectoryIds);
      }

      sql += ' ORDER BY timestamp ASC';

      const resultSet = await this.rdbStore.querySql(sql, params);
      const locationPoints: LocationPointEntity[] = [];

      while (resultSet.goToNextRow()) {
        locationPoints.push(this.parseLocationPointFromResultSet(resultSet));
      }

      resultSet.close();
      return locationPoints;
    } catch (error) {
      hilog.error(0x0000, 'DataService', 'Get location points error: %{public}s', JSON.stringify(error));
      return [];
    }
  }
}