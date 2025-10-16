import { LocationPoint } from './LocationPoint';

/**
 * 轨迹记录状态
 */
export enum TrajectoryStatus {
  RECORDING = 'recording',
  PAUSED = 'paused',
  FINISHED = 'finished'
}

/**
 * 轨迹数据模型
 */
export class Trajectory {
  id: string;
  name: string;
  description?: string;
  startTime: number;
  endTime?: number;
  status: TrajectoryStatus;
  totalDistance: number;
  duration: number;
  maxSpeed: number;
  avgSpeed: number;
  pointCount: number;
  minLatitude: number;
  maxLatitude: number;
  minLongitude: number;
  maxLongitude: number;
  createdAt: number;
  updatedAt: number;
  isSynced: boolean;
  userId?: string;
  tags: string[];
  color: string;
  activityType?: string; // 活动类型：walking, running, cycling, driving
  trajectoryType?: string; // 轨迹类型：exercise, commute, travel, leisure

  constructor(
    name: string,
    description?: string,
    options?: {
      id?: string;
      userId?: string;
      tags?: string[];
      color?: string;
      activityType?: string;
      trajectoryType?: string;
    }
  ) {
    this.id = options?.id || this.generateId();
    this.name = name;
    this.description = description;
    this.startTime = Date.now();
    this.status = TrajectoryStatus.RECORDING;
    this.totalDistance = 0;
    this.duration = 0;
    this.maxSpeed = 0;
    this.avgSpeed = 0;
    this.pointCount = 0;
    this.minLatitude = 90;
    this.maxLatitude = -90;
    this.minLongitude = 180;
    this.maxLongitude = -180;
    this.createdAt = Date.now();
    this.updatedAt = Date.now();
    this.isSynced = false;
    this.userId = options?.userId;
    this.tags = options?.tags || [];
    this.color = options?.color || '#007AFF';
    this.activityType = options?.activityType;
    this.trajectoryType = options?.trajectoryType;
  }

  /**
   * 生成唯一ID
   */
  private generateId(): string {
    return 'traj_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
  }

  /**
   * 计算持续时间
   */
  getDuration(): number {
    if (this.endTime) {
      return this.endTime - this.startTime;
    }
    return this.status === TrajectoryStatus.RECORDING ? Date.now() - this.startTime : this.duration;
  }

  /**
   * 格式化持续时间
   */
  getFormattedDuration(): string {
    const duration = this.getDuration();
    const hours = Math.floor(duration / 3600000);
    const minutes = Math.floor((duration % 3600000) / 60000);
    const seconds = Math.floor((duration % 60000) / 1000);

    if (hours > 0) {
      return `${hours}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
    }
    return `${minutes}:${seconds.toString().padStart(2, '0')}`;
  }

  /**
   * 格式化距离
   */
  getFormattedDistance(): string {
    if (this.totalDistance >= 1000) {
      return `${(this.totalDistance / 1000).toFixed(2)} km`;
    }
    return `${this.totalDistance.toFixed(0)} m`;
  }

  /**
   * 格式化最大速度
   */
  getFormattedMaxSpeed(): string {
    return `${(this.maxSpeed * 3.6).toFixed(1)} km/h`;
  }

  /**
   * 格式化平均速度
   */
  getFormattedAvgSpeed(): string {
    return `${(this.avgSpeed * 3.6).toFixed(1)} km/h`;
  }

  /**
   * 格式化开始时间
   */
  getFormattedStartTime(): string {
    return new Date(this.startTime).toLocaleString();
  }

  /**
   * 格式化结束时间
   */
  getFormattedEndTime(): string {
    return this.endTime ? new Date(this.endTime).toLocaleString() : '';
  }

  /**
   * 获取轨迹日期
   */
  getDate(): string {
    return new Date(this.startTime).toLocaleDateString();
  }

  /**
   * 获取状态显示文本
   */
  getStatusText(): string {
    switch (this.status) {
      case TrajectoryStatus.RECORDING:
        return '记录中';
      case TrajectoryStatus.PAUSED:
        return '已暂停';
      case TrajectoryStatus.FINISHED:
        return '已完成';
      default:
        return '未知';
    }
  }

  /**
   * 评估轨迹质量
   */
  getQuality(): 'excellent' | 'good' | 'fair' | 'poor' {
    if (this.pointCount < 10) return 'poor';
    if (this.totalDistance < 100) return 'poor';
    
    const avgPointDistance = this.totalDistance / this.pointCount;
    if (avgPointDistance > 100) return 'poor';
    if (avgPointDistance > 50) return 'fair';
    if (avgPointDistance > 20) return 'good';
    return 'excellent';
  }

  /**
   * 验证轨迹数据
   */
  isValid(): boolean {
    return this.name.length > 0 &&
      this.startTime > 0 &&
      this.totalDistance >= 0 &&
      this.pointCount >= 0;
  }

  /**
   * 获取中心点
   */
  getCenterPoint(): { latitude: number; longitude: number } {
    return {
      latitude: (this.minLatitude + this.maxLatitude) / 2,
      longitude: (this.minLongitude + this.maxLongitude) / 2
    };
  }

  /**
   * 获取边界框
   */
  getBoundingBox(): {
    minLatitude: number;
    maxLatitude: number;
    minLongitude: number;
    maxLongitude: number;
  } {
    return {
      minLatitude: this.minLatitude,
      maxLatitude: this.maxLatitude,
      minLongitude: this.minLongitude,
      maxLongitude: this.maxLongitude
    };
  }

  /**
   * 根据位置点更新统计信息
   */
  updateStats(points: LocationPoint[]): void {
    if (points.length === 0) return;

    this.pointCount = points.length;
    this.totalDistance = 0;
    let totalSpeed = 0;
    let speedCount = 0;
    this.maxSpeed = 0;

    // 重置边界
    this.minLatitude = 90;
    this.maxLatitude = -90;
    this.minLongitude = 180;
    this.maxLongitude = -180;

    for (let i = 0; i < points.length; i++) {
      const point = points[i];

      // 更新边界
      this.minLatitude = Math.min(this.minLatitude, point.latitude);
      this.maxLatitude = Math.max(this.maxLatitude, point.latitude);
      this.minLongitude = Math.min(this.minLongitude, point.longitude);
      this.maxLongitude = Math.max(this.maxLongitude, point.longitude);

      // 计算距离
      if (i > 0) {
        this.totalDistance += point.distanceTo(points[i - 1]);
      }

      // 更新速度统计
      if (point.speed !== undefined && point.speed >= 0) {
        this.maxSpeed = Math.max(this.maxSpeed, point.speed);
        totalSpeed += point.speed;
        speedCount++;
      }
    }

    // 计算平均速度
    this.avgSpeed = speedCount > 0 ? totalSpeed / speedCount : 0;
    this.updatedAt = Date.now();
  }

  /**
   * 完成轨迹记录
   */
  finish(): void {
    this.status = TrajectoryStatus.FINISHED;
    this.endTime = Date.now();
    this.duration = this.endTime - this.startTime;
    this.updatedAt = Date.now();
  }

  /**
   * 暂停轨迹记录
   */
  pause(): void {
    if (this.status === TrajectoryStatus.RECORDING) {
      this.status = TrajectoryStatus.PAUSED;
      this.duration += Date.now() - this.startTime;
      this.updatedAt = Date.now();
    }
  }

  /**
   * 恢复轨迹记录
   */
  resume(): void {
    if (this.status === TrajectoryStatus.PAUSED) {
      this.status = TrajectoryStatus.RECORDING;
      this.startTime = Date.now();
      this.updatedAt = Date.now();
    }
  }

  /**
   * 转换为JSON对象
   */
  toJSON(): Record<string, any> {
    return {
      id: this.id,
      name: this.name,
      description: this.description,
      startTime: this.startTime,
      endTime: this.endTime,
      status: this.status,
      totalDistance: this.totalDistance,
      duration: this.duration,
      maxSpeed: this.maxSpeed,
      avgSpeed: this.avgSpeed,
      pointCount: this.pointCount,
      minLatitude: this.minLatitude,
      maxLatitude: this.maxLatitude,
      minLongitude: this.minLongitude,
      maxLongitude: this.maxLongitude,
      createdAt: this.createdAt,
      updatedAt: this.updatedAt,
      isSynced: this.isSynced,
      userId: this.userId,
      tags: this.tags,
      color: this.color
    };
  }

  /**
   * 从JSON对象创建Trajectory
   */
  static fromJSON(json: Record<string, any>): Trajectory {
    const trajectory = new Trajectory(json.name, json.description, {
      id: json.id,
      userId: json.userId,
      tags: json.tags,
      color: json.color
    });

    trajectory.startTime = json.startTime;
    trajectory.endTime = json.endTime;
    trajectory.status = json.status;
    trajectory.totalDistance = json.totalDistance;
    trajectory.duration = json.duration;
    trajectory.maxSpeed = json.maxSpeed;
    trajectory.avgSpeed = json.avgSpeed;
    trajectory.pointCount = json.pointCount;
    trajectory.minLatitude = json.minLatitude;
    trajectory.maxLatitude = json.maxLatitude;
    trajectory.minLongitude = json.minLongitude;
    trajectory.maxLongitude = json.maxLongitude;
    trajectory.createdAt = json.createdAt;
    trajectory.updatedAt = json.updatedAt;
    trajectory.isSynced = json.isSynced;

    return trajectory;
  }

  /**
   * 创建新轨迹
   */
  static createNew(name?: string): Trajectory {
    const defaultName = `轨迹 ${new Date().toLocaleDateString()}`;
    return new Trajectory(name || defaultName);
  }
}