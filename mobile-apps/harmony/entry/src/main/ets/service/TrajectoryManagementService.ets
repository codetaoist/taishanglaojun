import { Trajectory } from '../model/Trajectory';
import { LocationPoint } from '../model/LocationPoint';
import { DataService } from './DataService';
import fs from '@ohos.file.fs';
import fileio from '@ohos.fileio';
import share from '@ohos.share';
import bundleManager from '@ohos.bundle.bundleManager';

/**
 * 导出格式枚举
 */
export enum ExportFormat {
  GPX = 'gpx',
  KML = 'kml',
  JSON = 'json',
  CSV = 'csv'
}

/**
 * 导出结果接口
 */
export interface ExportResult {
  success: boolean;
  filePath?: string;
  fileName?: string;
  fileSize?: number;
  error?: string;
}

/**
 * 分享选项接口
 */
export interface ShareOptions {
  title?: string;
  text?: string;
  includeMap?: boolean;
  format?: ExportFormat;
}

/**
 * 轨迹管理服务
 */
export class TrajectoryManagementService {
  private static instance: TrajectoryManagementService;
  private dataService: DataService;

  private constructor() {
    this.dataService = DataService.getInstance();
  }

  public static getInstance(): TrajectoryManagementService {
    if (!TrajectoryManagementService.instance) {
      TrajectoryManagementService.instance = new TrajectoryManagementService();
    }
    return TrajectoryManagementService.instance;
  }

  /**
   * 删除单个轨迹
   */
  async deleteTrajectory(trajectoryId: string): Promise<boolean> {
    try {
      // 删除轨迹点
      await this.dataService.deleteLocationPointsByTrajectory(trajectoryId);
      
      // 删除轨迹记录
      await this.dataService.deleteTrajectory(trajectoryId);
      
      return true;
    } catch (error) {
      console.error('Failed to delete trajectory:', error);
      return false;
    }
  }

  /**
   * 批量删除轨迹
   */
  async deleteTrajectories(trajectoryIds: string[]): Promise<{ success: number, failed: number }> {
    let success = 0;
    let failed = 0;

    for (const id of trajectoryIds) {
      const result = await this.deleteTrajectory(id);
      if (result) {
        success++;
      } else {
        failed++;
      }
    }

    return { success, failed };
  }

  /**
   * 导出轨迹
   */
  async exportTrajectory(trajectory: Trajectory, format: ExportFormat): Promise<ExportResult> {
    try {
      // 获取轨迹点数据
      const points = await this.dataService.getLocationPointsByTrajectory(trajectory.id);
      
      // 根据格式生成文件内容
      let content: string;
      let fileName: string;
      let mimeType: string;

      switch (format) {
        case ExportFormat.GPX:
          content = this.generateGPX(trajectory, points);
          fileName = `trajectory_${trajectory.id}.gpx`;
          mimeType = 'application/gpx+xml';
          break;
        case ExportFormat.KML:
          content = this.generateKML(trajectory, points);
          fileName = `trajectory_${trajectory.id}.kml`;
          mimeType = 'application/vnd.google-earth.kml+xml';
          break;
        case ExportFormat.JSON:
          content = this.generateJSON(trajectory, points);
          fileName = `trajectory_${trajectory.id}.json`;
          mimeType = 'application/json';
          break;
        case ExportFormat.CSV:
          content = this.generateCSV(trajectory, points);
          fileName = `trajectory_${trajectory.id}.csv`;
          mimeType = 'text/csv';
          break;
        default:
          throw new Error('Unsupported export format');
      }

      // 保存文件
      const filePath = await this.saveFile(fileName, content);
      const fileSize = content.length;

      return {
        success: true,
        filePath,
        fileName,
        fileSize
      };

    } catch (error) {
      console.error('Failed to export trajectory:', error);
      return {
        success: false,
        error: error.message
      };
    }
  }

  /**
   * 批量导出轨迹
   */
  async exportTrajectories(trajectories: Trajectory[], format: ExportFormat): Promise<ExportResult> {
    try {
      if (trajectories.length === 1) {
        return await this.exportTrajectory(trajectories[0], format);
      }

      // 多个轨迹合并导出
      const allPoints: { trajectory: Trajectory, points: LocationPoint[] }[] = [];
      
      for (const trajectory of trajectories) {
        const points = await this.dataService.getLocationPointsByTrajectory(trajectory.id);
        allPoints.push({ trajectory, points });
      }

      let content: string;
      let fileName: string;

      switch (format) {
        case ExportFormat.GPX:
          content = this.generateMultiGPX(allPoints);
          fileName = `trajectories_${Date.now()}.gpx`;
          break;
        case ExportFormat.KML:
          content = this.generateMultiKML(allPoints);
          fileName = `trajectories_${Date.now()}.kml`;
          break;
        case ExportFormat.JSON:
          content = this.generateMultiJSON(allPoints);
          fileName = `trajectories_${Date.now()}.json`;
          break;
        case ExportFormat.CSV:
          content = this.generateMultiCSV(allPoints);
          fileName = `trajectories_${Date.now()}.csv`;
          break;
        default:
          throw new Error('Unsupported export format');
      }

      const filePath = await this.saveFile(fileName, content);
      const fileSize = content.length;

      return {
        success: true,
        filePath,
        fileName,
        fileSize
      };

    } catch (error) {
      console.error('Failed to export trajectories:', error);
      return {
        success: false,
        error: error.message
      };
    }
  }

  /**
   * 分享轨迹
   */
  async shareTrajectory(trajectory: Trajectory, options: ShareOptions = {}): Promise<boolean> {
    try {
      const format = options.format || ExportFormat.GPX;
      const exportResult = await this.exportTrajectory(trajectory, format);
      
      if (!exportResult.success || !exportResult.filePath) {
        throw new Error('Failed to export trajectory for sharing');
      }

      // 准备分享内容
      const shareData = {
        title: options.title || `轨迹分享 - ${this.formatDate(trajectory.startTime)}`,
        text: options.text || this.generateShareText(trajectory),
        files: [exportResult.filePath]
      };

      // 调用系统分享
      await share.share(shareData);
      
      return true;
    } catch (error) {
      console.error('Failed to share trajectory:', error);
      return false;
    }
  }

  /**
   * 生成GPX格式内容
   */
  private generateGPX(trajectory: Trajectory, points: LocationPoint[]): string {
    const startDate = new Date(trajectory.startTime).toISOString();
    
    let gpx = `<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1" creator="TrajectoryApp" xmlns="http://www.topografix.com/GPX/1/1">
  <metadata>
    <name>Trajectory ${trajectory.id}</name>
    <time>${startDate}</time>
  </metadata>
  <trk>
    <name>Trajectory ${trajectory.id}</name>
    <trkseg>`;

    for (const point of points) {
      const time = new Date(point.timestamp).toISOString();
      gpx += `
      <trkpt lat="${point.latitude}" lon="${point.longitude}">`;
      
      if (point.altitude) {
        gpx += `
        <ele>${point.altitude}</ele>`;
      }
      
      gpx += `
        <time>${time}</time>`;
      
      if (point.speed) {
        gpx += `
        <extensions>
          <speed>${point.speed}</speed>
        </extensions>`;
      }
      
      gpx += `
      </trkpt>`;
    }

    gpx += `
    </trkseg>
  </trk>
</gpx>`;

    return gpx;
  }

  /**
   * 生成KML格式内容
   */
  private generateKML(trajectory: Trajectory, points: LocationPoint[]): string {
    const startDate = new Date(trajectory.startTime).toISOString();
    
    let kml = `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
  <Document>
    <name>Trajectory ${trajectory.id}</name>
    <description>Created on ${startDate}</description>
    <Placemark>
      <name>Trajectory ${trajectory.id}</name>
      <LineString>
        <coordinates>`;

    for (const point of points) {
      kml += `${point.longitude},${point.latitude}`;
      if (point.altitude) {
        kml += `,${point.altitude}`;
      }
      kml += ' ';
    }

    kml += `
        </coordinates>
      </LineString>
    </Placemark>
  </Document>
</kml>`;

    return kml;
  }

  /**
   * 生成JSON格式内容
   */
  private generateJSON(trajectory: Trajectory, points: LocationPoint[]): string {
    const data = {
      trajectory: {
        id: trajectory.id,
        startTime: trajectory.startTime,
        endTime: trajectory.endTime,
        totalDistance: trajectory.totalDistance,
        averageSpeed: trajectory.averageSpeed,
        maxSpeed: trajectory.maxSpeed
      },
      points: points.map(point => ({
        latitude: point.latitude,
        longitude: point.longitude,
        altitude: point.altitude,
        timestamp: point.timestamp,
        speed: point.speed,
        accuracy: point.accuracy,
        bearing: point.bearing
      }))
    };

    return JSON.stringify(data, null, 2);
  }

  /**
   * 生成CSV格式内容
   */
  private generateCSV(trajectory: Trajectory, points: LocationPoint[]): string {
    let csv = 'timestamp,latitude,longitude,altitude,speed,accuracy,bearing\n';
    
    for (const point of points) {
      csv += `${point.timestamp},${point.latitude},${point.longitude},${point.altitude || ''},${point.speed || ''},${point.accuracy || ''},${point.bearing || ''}\n`;
    }

    return csv;
  }

  /**
   * 生成多轨迹GPX
   */
  private generateMultiGPX(trajectoryData: { trajectory: Trajectory, points: LocationPoint[] }[]): string {
    let gpx = `<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1" creator="TrajectoryApp" xmlns="http://www.topografix.com/GPX/1/1">
  <metadata>
    <name>Multiple Trajectories</name>
    <time>${new Date().toISOString()}</time>
  </metadata>`;

    for (const { trajectory, points } of trajectoryData) {
      gpx += `
  <trk>
    <name>Trajectory ${trajectory.id}</name>
    <trkseg>`;

      for (const point of points) {
        const time = new Date(point.timestamp).toISOString();
        gpx += `
      <trkpt lat="${point.latitude}" lon="${point.longitude}">`;
        
        if (point.altitude) {
          gpx += `
        <ele>${point.altitude}</ele>`;
        }
        
        gpx += `
        <time>${time}</time>
      </trkpt>`;
      }

      gpx += `
    </trkseg>
  </trk>`;
    }

    gpx += `
</gpx>`;

    return gpx;
  }

  /**
   * 生成多轨迹KML
   */
  private generateMultiKML(trajectoryData: { trajectory: Trajectory, points: LocationPoint[] }[]): string {
    let kml = `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
  <Document>
    <name>Multiple Trajectories</name>`;

    for (const { trajectory, points } of trajectoryData) {
      kml += `
    <Placemark>
      <name>Trajectory ${trajectory.id}</name>
      <LineString>
        <coordinates>`;

      for (const point of points) {
        kml += `${point.longitude},${point.latitude}`;
        if (point.altitude) {
          kml += `,${point.altitude}`;
        }
        kml += ' ';
      }

      kml += `
        </coordinates>
      </LineString>
    </Placemark>`;
    }

    kml += `
  </Document>
</kml>`;

    return kml;
  }

  /**
   * 生成多轨迹JSON
   */
  private generateMultiJSON(trajectoryData: { trajectory: Trajectory, points: LocationPoint[] }[]): string {
    const data = {
      trajectories: trajectoryData.map(({ trajectory, points }) => ({
        trajectory: {
          id: trajectory.id,
          startTime: trajectory.startTime,
          endTime: trajectory.endTime,
          totalDistance: trajectory.totalDistance,
          averageSpeed: trajectory.averageSpeed,
          maxSpeed: trajectory.maxSpeed
        },
        points: points.map(point => ({
          latitude: point.latitude,
          longitude: point.longitude,
          altitude: point.altitude,
          timestamp: point.timestamp,
          speed: point.speed,
          accuracy: point.accuracy,
          bearing: point.bearing
        }))
      }))
    };

    return JSON.stringify(data, null, 2);
  }

  /**
   * 生成多轨迹CSV
   */
  private generateMultiCSV(trajectoryData: { trajectory: Trajectory, points: LocationPoint[] }[]): string {
    let csv = 'trajectory_id,timestamp,latitude,longitude,altitude,speed,accuracy,bearing\n';
    
    for (const { trajectory, points } of trajectoryData) {
      for (const point of points) {
        csv += `${trajectory.id},${point.timestamp},${point.latitude},${point.longitude},${point.altitude || ''},${point.speed || ''},${point.accuracy || ''},${point.bearing || ''}\n`;
      }
    }

    return csv;
  }

  /**
   * 保存文件
   */
  private async saveFile(fileName: string, content: string): Promise<string> {
    try {
      // 获取应用缓存目录
      const context = getContext();
      const cacheDir = context.cacheDir;
      const filePath = `${cacheDir}/${fileName}`;

      // 写入文件
      const file = fs.openSync(filePath, fs.OpenMode.CREATE | fs.OpenMode.WRITE_ONLY);
      fs.writeSync(file.fd, content);
      fs.closeSync(file);

      return filePath;
    } catch (error) {
      console.error('Failed to save file:', error);
      throw error;
    }
  }

  /**
   * 生成分享文本
   */
  private generateShareText(trajectory: Trajectory): string {
    const startDate = this.formatDate(trajectory.startTime);
    const distance = this.formatDistance(trajectory.totalDistance);
    const duration = this.formatDuration(trajectory.endTime - trajectory.startTime);
    
    return `我在${startDate}记录了一条轨迹，总距离${distance}，用时${duration}。`;
  }

  /**
   * 格式化日期
   */
  private formatDate(timestamp: number): string {
    const date = new Date(timestamp);
    return `${date.getFullYear()}年${date.getMonth() + 1}月${date.getDate()}日`;
  }

  /**
   * 格式化距离
   */
  private formatDistance(meters: number): string {
    if (meters >= 1000) {
      return `${(meters / 1000).toFixed(2)}公里`;
    } else {
      return `${meters.toFixed(0)}米`;
    }
  }

  /**
   * 格式化持续时间
   */
  private formatDuration(milliseconds: number): string {
    const seconds = Math.floor(milliseconds / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    
    if (hours > 0) {
      return `${hours}小时${minutes % 60}分钟`;
    } else if (minutes > 0) {
      return `${minutes}分钟`;
    } else {
      return `${seconds}秒`;
    }
  }
}