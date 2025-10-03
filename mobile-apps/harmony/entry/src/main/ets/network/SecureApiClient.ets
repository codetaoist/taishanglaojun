import { http } from '@kit.NetworkKit';
import { LocationPoint } from '../data/LocationPoint';
import { Trajectory } from '../data/Trajectory';
import { DataEncryption, TransportEncryptedData } from '../security/DataEncryption';

/**
 * API响应封装
 */
export class ApiResponse<T> {
  constructor(
    public success: boolean,
    public data?: T,
    public message?: string
  ) {}
  
  static success<T>(data: T): ApiResponse<T> {
    return new ApiResponse(true, data);
  }
  
  static error<T>(message: string): ApiResponse<T> {
    return new ApiResponse(false, undefined, message);
  }
}

/**
 * API响应包装器
 */
interface ApiResponseWrapper {
  success: boolean;
  message?: string;
  data?: any;
}

/**
 * 上传位置点请求
 */
interface UploadLocationPointsRequest {
  transportKey: string;
  encryptedPoints: TransportEncryptedData[];
  deviceId: string;
  timestamp: number;
}

/**
 * 上传轨迹请求
 */
interface UploadTrajectoryRequest {
  transportKey: string;
  encryptedTrajectory: TransportEncryptedData;
  deviceId: string;
  timestamp: number;
}

/**
 * 下载轨迹响应
 */
interface DownloadTrajectoriesResponse {
  transportKey: string;
  encryptedTrajectories: TransportEncryptedData[];
}

/**
 * 下载位置点响应
 */
interface DownloadLocationPointsResponse {
  transportKey: string;
  encryptedPoints: TransportEncryptedData[];
}

/**
 * 同步请求
 */
interface SyncRequest {
  lastSyncTime: number;
  deviceId: string;
  trajectoryIds: string[];
}

/**
 * 同步响应
 */
interface SyncResponse {
  newTrajectories: Trajectory[];
  updatedTrajectories: Trajectory[];
  deletedTrajectoryIds: string[];
  syncTime: number;
}

/**
 * 安全API客户端
 * 负责与后端服务的安全通信
 */
export class SecureApiClient {
  private static instance: SecureApiClient;
  private static readonly BASE_URL = 'https://api.taishanglaojun.com';
  private static readonly CONNECT_TIMEOUT = 30000;
  private static readonly READ_TIMEOUT = 30000;
  
  private dataEncryption: DataEncryption;
  private httpRequest: http.HttpRequest;
  
  private constructor() {
    this.dataEncryption = DataEncryption.getInstance();
    this.httpRequest = http.createHttp();
    this.configureHttpRequest();
  }
  
  /**
   * 获取单例实例
   */
  public static getInstance(): SecureApiClient {
    if (!SecureApiClient.instance) {
      SecureApiClient.instance = new SecureApiClient();
    }
    return SecureApiClient.instance;
  }
  
  /**
   * 配置HTTP请求
   */
  private configureHttpRequest(): void {
    // 配置HTTP请求的默认参数
    this.httpRequest.on('headersReceive', (header) => {
      console.info('Response headers:', JSON.stringify(header));
    });
    
    this.httpRequest.on('dataReceive', (data) => {
      console.info('Response data received');
    });
    
    this.httpRequest.on('dataEnd', () => {
      console.info('Response data end');
    });
  }
  
  /**
   * 上传位置点
   */
  public async uploadLocationPoints(locationPoints: LocationPoint[]): Promise<ApiResponse<string>> {
    try {
      const transportKey = this.dataEncryption.generateTransportKey();
      const encryptedPoints: TransportEncryptedData[] = [];
      
      for (const point of locationPoints) {
        const json = JSON.stringify(point);
        const encryptedData = await this.dataEncryption.encryptForTransport(json, transportKey);
        encryptedPoints.push(encryptedData);
      }
      
      const requestData: UploadLocationPointsRequest = {
        transportKey: transportKey,
        encryptedPoints: encryptedPoints,
        deviceId: await this.getDeviceId(),
        timestamp: Date.now()
      };
      
      const response = await this.makeRequest(
        'POST',
        '/api/v1/location-points',
        requestData
      );
      
      return this.handleResponse<string>(response);
    } catch (error) {
      console.error('Failed to upload location points:', error);
      return ApiResponse.error(`Failed to upload location points: ${error.message}`);
    }
  }
  
  /**
   * 上传轨迹
   */
  public async uploadTrajectory(trajectory: Trajectory): Promise<ApiResponse<string>> {
    try {
      const transportKey = this.dataEncryption.generateTransportKey();
      const json = JSON.stringify(trajectory);
      const encryptedTrajectory = await this.dataEncryption.encryptForTransport(json, transportKey);
      
      const requestData: UploadTrajectoryRequest = {
        transportKey: transportKey,
        encryptedTrajectory: encryptedTrajectory,
        deviceId: await this.getDeviceId(),
        timestamp: Date.now()
      };
      
      const response = await this.makeRequest(
        'POST',
        '/api/v1/trajectories',
        requestData
      );
      
      return this.handleResponse<string>(response);
    } catch (error) {
      console.error('Failed to upload trajectory:', error);
      return ApiResponse.error(`Failed to upload trajectory: ${error.message}`);
    }
  }
  
  /**
   * 下载轨迹列表
   */
  public async downloadTrajectories(userId: string): Promise<ApiResponse<Trajectory[]>> {
    try {
      const response = await this.makeRequest(
        'GET',
        `/api/v1/trajectories?userId=${userId}`
      );
      
      const apiResponse = this.handleResponse<DownloadTrajectoriesResponse>(response);
      
      if (apiResponse.success && apiResponse.data) {
        const trajectories: Trajectory[] = [];
        
        for (const encryptedData of apiResponse.data.encryptedTrajectories) {
          const decryptedJson = await this.dataEncryption.decryptFromTransport(
            encryptedData,
            apiResponse.data.transportKey
          );
          const trajectory = JSON.parse(decryptedJson) as Trajectory;
          trajectories.push(trajectory);
        }
        
        return ApiResponse.success(trajectories);
      } else {
        return ApiResponse.error(apiResponse.message || 'Failed to download trajectories');
      }
    } catch (error) {
      console.error('Failed to download trajectories:', error);
      return ApiResponse.error(`Failed to download trajectories: ${error.message}`);
    }
  }
  
  /**
   * 下载轨迹的位置点
   */
  public async downloadLocationPoints(trajectoryId: string): Promise<ApiResponse<LocationPoint[]>> {
    try {
      const response = await this.makeRequest(
        'GET',
        `/api/v1/trajectories/${trajectoryId}/points`
      );
      
      const apiResponse = this.handleResponse<DownloadLocationPointsResponse>(response);
      
      if (apiResponse.success && apiResponse.data) {
        const locationPoints: LocationPoint[] = [];
        
        for (const encryptedData of apiResponse.data.encryptedPoints) {
          const decryptedJson = await this.dataEncryption.decryptFromTransport(
            encryptedData,
            apiResponse.data.transportKey
          );
          const locationPoint = JSON.parse(decryptedJson) as LocationPoint;
          locationPoints.push(locationPoint);
        }
        
        return ApiResponse.success(locationPoints);
      } else {
        return ApiResponse.error(apiResponse.message || 'Failed to download location points');
      }
    } catch (error) {
      console.error('Failed to download location points:', error);
      return ApiResponse.error(`Failed to download location points: ${error.message}`);
    }
  }
  
  /**
   * 删除轨迹
   */
  public async deleteTrajectory(trajectoryId: string): Promise<ApiResponse<string>> {
    try {
      const response = await this.makeRequest(
        'DELETE',
        `/api/v1/trajectories/${trajectoryId}`
      );
      
      return this.handleResponse<string>(response);
    } catch (error) {
      console.error('Failed to delete trajectory:', error);
      return ApiResponse.error(`Failed to delete trajectory: ${error.message}`);
    }
  }
  
  /**
   * 同步数据
   */
  public async syncData(syncRequest: SyncRequest): Promise<ApiResponse<SyncResponse>> {
    try {
      const response = await this.makeRequest(
        'POST',
        '/api/v1/sync',
        syncRequest
      );
      
      return this.handleResponse<SyncResponse>(response);
    } catch (error) {
      console.error('Failed to sync data:', error);
      return ApiResponse.error(`Failed to sync data: ${error.message}`);
    }
  }
  
  /**
   * 发起HTTP请求
   */
  private async makeRequest(
    method: string,
    path: string,
    data?: any
  ): Promise<http.HttpResponse> {
    const url = `${SecureApiClient.BASE_URL}${path}`;
    
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'X-Device-ID': await this.getDeviceId(),
      'X-Encryption-Version': '1.0',
      'X-Client-Version': '1.0.0'
    };
    
    // 添加认证头
    const token = await this.dataEncryption.getSecureConfig('auth_token');
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    
    const requestOptions: http.HttpRequestOptions = {
      method: method as http.RequestMethod,
      header: headers,
      connectTimeout: SecureApiClient.CONNECT_TIMEOUT,
      readTimeout: SecureApiClient.READ_TIMEOUT,
      usingCache: false,
      usingProtocol: http.HttpProtocol.HTTP2
    };
    
    if (data && (method === 'POST' || method === 'PUT' || method === 'PATCH')) {
      requestOptions.extraData = JSON.stringify(data);
    }
    
    return new Promise((resolve, reject) => {
      this.httpRequest.request(url, requestOptions, (err, data) => {
        if (err) {
          console.error('HTTP request error:', err);
          reject(new Error(`HTTP request failed: ${err.message}`));
        } else {
          resolve(data);
        }
      });
    });
  }
  
  /**
   * 处理HTTP响应
   */
  private handleResponse<T>(response: http.HttpResponse): ApiResponse<T> {
    try {
      if (response.responseCode >= 200 && response.responseCode < 300) {
        const responseBody = response.result as string;
        if (responseBody) {
          const apiResponse: ApiResponseWrapper = JSON.parse(responseBody);
          if (apiResponse.success) {
            return ApiResponse.success(apiResponse.data as T);
          } else {
            return ApiResponse.error(apiResponse.message || 'Unknown error');
          }
        } else {
          return ApiResponse.error('Empty response body');
        }
      } else {
        return ApiResponse.error(`HTTP ${response.responseCode}: ${response.result}`);
      }
    } catch (error) {
      console.error('Failed to parse response:', error);
      return ApiResponse.error(`Failed to parse response: ${error.message}`);
    }
  }
  
  /**
   * 获取设备ID
   */
  private async getDeviceId(): Promise<string> {
    let deviceId = await this.dataEncryption.getSecureConfig('device_id');
    if (!deviceId) {
      deviceId = this.generateUUID();
      await this.dataEncryption.storeSecureConfig('device_id', deviceId);
    }
    return deviceId;
  }
  
  /**
   * 生成UUID
   */
  private generateUUID(): string {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
      const r = Math.random() * 16 | 0;
      const v = c === 'x' ? r : (r & 0x3 | 0x8);
      return v.toString(16);
    });
  }
  
  /**
   * 销毁HTTP请求
   */
  public destroy(): void {
    if (this.httpRequest) {
      this.httpRequest.destroy();
    }
  }
}