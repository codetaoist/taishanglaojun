/**
 * 位置点数据模型
 */
export class LocationPoint {
  id?: string;
  latitude: number;
  longitude: number;
  timestamp: number;
  accuracy?: number;
  altitude?: number;
  speed?: number;
  bearing?: number;
  trajectoryId?: string;
  provider?: string;
  isMock?: boolean;
  batteryLevel?: number;
  networkType?: string;

  constructor(
    latitude: number,
    longitude: number,
    timestamp: number = Date.now(),
    options?: {
      id?: string;
      accuracy?: number;
      altitude?: number;
      speed?: number;
      bearing?: number;
      trajectoryId?: string;
      provider?: string;
      isMock?: boolean;
      batteryLevel?: number;
      networkType?: string;
    }
  ) {
    this.latitude = latitude;
    this.longitude = longitude;
    this.timestamp = timestamp;
    
    if (options) {
      this.id = options.id;
      this.accuracy = options.accuracy;
      this.altitude = options.altitude;
      this.speed = options.speed;
      this.bearing = options.bearing;
      this.trajectoryId = options.trajectoryId;
      this.provider = options.provider;
      this.isMock = options.isMock;
      this.batteryLevel = options.batteryLevel;
      this.networkType = options.networkType;
    }
  }

  /**
   * 格式化时间戳
   */
  getFormattedTime(): string {
    const date = new Date(this.timestamp);
    return date.toLocaleString();
  }

  /**
   * 格式化坐标
   */
  getFormattedCoordinates(): string {
    return `${this.latitude.toFixed(6)}, ${this.longitude.toFixed(6)}`;
  }

  /**
   * 计算到另一个点的距离（米）
   */
  distanceTo(other: LocationPoint): number {
    const R = 6371000; // 地球半径（米）
    const lat1Rad = this.latitude * Math.PI / 180;
    const lat2Rad = other.latitude * Math.PI / 180;
    const deltaLatRad = (other.latitude - this.latitude) * Math.PI / 180;
    const deltaLonRad = (other.longitude - this.longitude) * Math.PI / 180;

    const a = Math.sin(deltaLatRad / 2) * Math.sin(deltaLatRad / 2) +
      Math.cos(lat1Rad) * Math.cos(lat2Rad) *
      Math.sin(deltaLonRad / 2) * Math.sin(deltaLonRad / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));

    return R * c;
  }

  /**
   * 计算到另一个点的方位角（度）
   */
  bearingTo(other: LocationPoint): number {
    const lat1Rad = this.latitude * Math.PI / 180;
    const lat2Rad = other.latitude * Math.PI / 180;
    const deltaLonRad = (other.longitude - this.longitude) * Math.PI / 180;

    const y = Math.sin(deltaLonRad) * Math.cos(lat2Rad);
    const x = Math.cos(lat1Rad) * Math.sin(lat2Rad) -
      Math.sin(lat1Rad) * Math.cos(lat2Rad) * Math.cos(deltaLonRad);

    const bearing = Math.atan2(y, x) * 180 / Math.PI;
    return (bearing + 360) % 360;
  }

  /**
   * 验证位置点是否有效
   */
  isValid(): boolean {
    return this.latitude >= -90 && this.latitude <= 90 &&
      this.longitude >= -180 && this.longitude <= 180 &&
      this.timestamp > 0;
  }

  /**
   * 检查精度是否足够好
   */
  hasGoodAccuracy(threshold: number = 50): boolean {
    return this.accuracy !== undefined && this.accuracy <= threshold;
  }

  /**
   * 获取位置质量等级
   */
  getQuality(): 'excellent' | 'good' | 'fair' | 'poor' {
    if (!this.accuracy) return 'poor';
    
    if (this.accuracy <= 5) return 'excellent';
    if (this.accuracy <= 15) return 'good';
    if (this.accuracy <= 50) return 'fair';
    return 'poor';
  }

  /**
   * 转换为JSON对象
   */
  toJSON(): Record<string, any> {
    return {
      id: this.id,
      latitude: this.latitude,
      longitude: this.longitude,
      timestamp: this.timestamp,
      accuracy: this.accuracy,
      altitude: this.altitude,
      speed: this.speed,
      bearing: this.bearing,
      trajectoryId: this.trajectoryId,
      provider: this.provider,
      isMock: this.isMock,
      batteryLevel: this.batteryLevel,
      networkType: this.networkType
    };
  }

  /**
   * 从JSON对象创建LocationPoint
   */
  static fromJSON(json: Record<string, any>): LocationPoint {
    return new LocationPoint(
      json.latitude,
      json.longitude,
      json.timestamp,
      {
        id: json.id,
        accuracy: json.accuracy,
        altitude: json.altitude,
        speed: json.speed,
        bearing: json.bearing,
        trajectoryId: json.trajectoryId,
        provider: json.provider,
        isMock: json.isMock,
        batteryLevel: json.batteryLevel,
        networkType: json.networkType
      }
    );
  }

  /**
   * 创建测试用的位置点
   */
  static createTestPoint(lat: number = 39.9042, lon: number = 116.4074): LocationPoint {
    return new LocationPoint(lat, lon, Date.now(), {
      accuracy: 10,
      provider: 'gps',
      isMock: false
    });
  }
}