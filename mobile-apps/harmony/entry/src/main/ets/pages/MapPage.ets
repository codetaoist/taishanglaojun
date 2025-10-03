import { map } from '@kit.MapKit';
import { LocationPoint } from '../data/LocationPoint';
import { Trajectory } from '../data/Trajectory';
import { LocationViewModel } from '../viewmodel/LocationViewModel';

/**
 * 地图页面
 * 负责显示轨迹和位置点的可视化
 */
@Entry
@Component
struct MapPage {
  @State private mapController: map.MapController | null = null;
  @State private locationViewModel: LocationViewModel = new LocationViewModel();
  
  // 地图状态
  @State private mapType: map.MapType = map.MapType.NORMAL;
  @State private showTrajectories: boolean = true;
  @State private isTracking: boolean = false;
  @State private currentLocation: LocationPoint | null = null;
  @State private trajectories: Trajectory[] = [];
  
  // 可视化元素
  private trajectoryPolylines: Map<string, map.MapPolyline> = new Map();
  private locationMarkers: Map<string, map.MapMarker> = new Map();
  private currentLocationMarker: map.MapMarker | null = null;
  
  // 轨迹颜色配置
  private readonly trajectoryColors: string[] = [
    '#2196F3', '#F44336', '#4CAF50', '#FF9800',
    '#9C27B0', '#00BCD4', '#FFEB3B', '#795548'
  ];
  
  aboutToAppear() {
    this.initializeViewModel();
    this.loadTrajectories();
  }
  
  /**
   * 初始化ViewModel
   */
  private initializeViewModel(): void {
    // 监听位置更新
    this.locationViewModel.addStateChangeListener((state) => {
      this.isTracking = state.isTracking;
      this.currentLocation = state.currentLocation;
      this.trajectories = state.trajectories;
      
      // 更新地图显示
      this.updateMapDisplay();
    });
  }
  
  /**
   * 加载轨迹数据
   */
  private async loadTrajectories(): Promise<void> {
    try {
      await this.locationViewModel.loadTrajectories();
    } catch (error) {
      console.error('Failed to load trajectories:', error);
    }
  }
  
  /**
   * 更新地图显示
   */
  private updateMapDisplay(): void {
    if (!this.mapController) return;
    
    // 更新当前位置
    this.updateCurrentLocationMarker();
    
    // 更新轨迹显示
    this.updateTrajectoriesDisplay();
  }
  
  /**
   * 更新当前位置标记
   */
  private updateCurrentLocationMarker(): void {
    if (!this.mapController || !this.currentLocation) return;
    
    const position: map.LatLng = {
      latitude: this.currentLocation.latitude,
      longitude: this.currentLocation.longitude
    };
    
    if (this.currentLocationMarker) {
      // 更新现有标记位置
      this.currentLocationMarker.setPosition(position);
    } else {
      // 创建新的当前位置标记
      const markerOptions: map.MapMarkerOptions = {
        position: position,
        markerId: 'current_location',
        icon: {
          src: $r('app.media.ic_current_location'),
          width: 32,
          height: 32
        }
      };
      
      this.currentLocationMarker = this.mapController.addMarker(markerOptions);
    }
    
    // 如果正在跟踪，移动相机到当前位置
    if (this.isTracking) {
      this.mapController.moveCamera({
        target: position,
        zoom: 16,
        duration: 1000
      });
    }
  }
  
  /**
   * 更新轨迹显示
   */
  private updateTrajectoriesDisplay(): void {
    if (!this.mapController) return;
    
    // 清除现有轨迹
    this.clearTrajectories();
    
    // 绘制所有轨迹
    this.trajectories.forEach((trajectory, index) => {
      this.drawTrajectory(trajectory, index);
    });
  }
  
  /**
   * 绘制轨迹
   */
  private async drawTrajectory(trajectory: Trajectory, colorIndex: number): Promise<void> {
    if (!this.mapController) return;
    
    try {
      // 获取轨迹点
      const points = await this.locationViewModel.getTrajectoryPoints(trajectory.id);
      if (points.length === 0) return;
      
      // 转换为地图坐标
      const latLngPoints: map.LatLng[] = points.map(point => ({
        latitude: point.latitude,
        longitude: point.longitude
      }));
      
      // 选择颜色
      const color = this.trajectoryColors[colorIndex % this.trajectoryColors.length];
      const isCurrentTrajectory = this.locationViewModel.getCurrentTrajectory()?.id === trajectory.id;
      
      // 创建轨迹线
      const polylineOptions: map.MapPolylineOptions = {
        polylineId: trajectory.id,
        points: latLngPoints,
        color: color,
        width: isCurrentTrajectory ? 8 : 5,
        pattern: isCurrentTrajectory ? [] : [10, 5], // 虚线模式
        clickable: true
      };
      
      const polyline = this.mapController.addPolyline(polylineOptions);
      this.trajectoryPolylines.set(trajectory.id, polyline);
      
      // 添加起点标记
      if (latLngPoints.length > 0) {
        const startMarkerOptions: map.MapMarkerOptions = {
          position: latLngPoints[0],
          markerId: `start_${trajectory.id}`,
          icon: {
            src: $r('app.media.ic_start_point'),
            width: 24,
            height: 24
          }
        };
        
        const startMarker = this.mapController.addMarker(startMarkerOptions);
        this.locationMarkers.set(`start_${trajectory.id}`, startMarker);
      }
      
      // 添加终点标记（如果轨迹已结束）
      if (trajectory.endTime && latLngPoints.length > 1) {
        const endMarkerOptions: map.MapMarkerOptions = {
          position: latLngPoints[latLngPoints.length - 1],
          markerId: `end_${trajectory.id}`,
          icon: {
            src: $r('app.media.ic_end_point'),
            width: 24,
            height: 24
          }
        };
        
        const endMarker = this.mapController.addMarker(endMarkerOptions);
        this.locationMarkers.set(`end_${trajectory.id}`, endMarker);
      }
      
    } catch (error) {
      console.error('Failed to draw trajectory:', error);
    }
  }
  
  /**
   * 清除轨迹
   */
  private clearTrajectories(): void {
    if (!this.mapController) return;
    
    // 移除轨迹线
    this.trajectoryPolylines.forEach((polyline) => {
      this.mapController?.removePolyline(polyline);
    });
    this.trajectoryPolylines.clear();
    
    // 移除标记
    this.locationMarkers.forEach((marker) => {
      this.mapController?.removeMarker(marker);
    });
    this.locationMarkers.clear();
  }
  
  /**
   * 切换地图类型
   */
  private toggleMapType(): void {
    if (!this.mapController) return;
    
    switch (this.mapType) {
      case map.MapType.NORMAL:
        this.mapType = map.MapType.SATELLITE;
        break;
      case map.MapType.SATELLITE:
        this.mapType = map.MapType.TERRAIN;
        break;
      default:
        this.mapType = map.MapType.NORMAL;
        break;
    }
    
    this.mapController.setMapType(this.mapType);
  }
  
  /**
   * 切换轨迹可见性
   */
  private toggleTrajectoriesVisibility(): void {
    this.showTrajectories = !this.showTrajectories;
    
    // 切换轨迹线可见性
    this.trajectoryPolylines.forEach((polyline) => {
      polyline.setVisible(this.showTrajectories);
    });
    
    // 切换标记可见性
    this.locationMarkers.forEach((marker) => {
      marker.setVisible(this.showTrajectories);
    });
  }
  
  /**
   * 居中到当前位置
   */
  private centerToCurrentLocation(): void {
    if (!this.mapController || !this.currentLocation) return;
    
    const position: map.LatLng = {
      latitude: this.currentLocation.latitude,
      longitude: this.currentLocation.longitude
    };
    
    this.mapController.moveCamera({
      target: position,
      zoom: 16,
      duration: 1000
    });
  }
  
  /**
   * 适应所有轨迹
   */
  private fitAllTrajectories(): void {
    if (!this.mapController || this.trajectories.length === 0) return;
    
    // 计算所有轨迹的边界
    let minLat = 90, maxLat = -90, minLng = 180, maxLng = -180;
    let hasPoints = false;
    
    this.trajectories.forEach(async (trajectory) => {
      const points = await this.locationViewModel.getTrajectoryPoints(trajectory.id);
      points.forEach((point) => {
        minLat = Math.min(minLat, point.latitude);
        maxLat = Math.max(maxLat, point.latitude);
        minLng = Math.min(minLng, point.longitude);
        maxLng = Math.max(maxLng, point.longitude);
        hasPoints = true;
      });
    });
    
    if (hasPoints) {
      const bounds: map.LatLngBounds = {
        southwest: { latitude: minLat, longitude: minLng },
        northeast: { latitude: maxLat, longitude: maxLng }
      };
      
      this.mapController.fitBounds(bounds, 100); // 100px padding
    }
  }
  
  /**
   * 清除地图
   */
  private clearMap(): void {
    this.clearTrajectories();
    
    if (this.currentLocationMarker) {
      this.mapController?.removeMarker(this.currentLocationMarker);
      this.currentLocationMarker = null;
    }
  }
  
  /**
   * 格式化时间
   */
  private formatTime(timestamp?: number): string {
    if (!timestamp) return '未知';
    
    const date = new Date(timestamp);
    return `${date.getFullYear()}-${(date.getMonth() + 1).toString().padStart(2, '0')}-${date.getDate().toString().padStart(2, '0')} ${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`;
  }
  
  build() {
    Column() {
      // 地图容器
      Stack() {
        // 地图组件
        MapComponent({
          mapOptions: {
            position: {
              target: {
                latitude: 39.9042,
                longitude: 116.4074
              },
              zoom: 10
            },
            mapType: this.mapType
          },
          mapCallback: (controller: map.MapController) => {
            this.mapController = controller;
            this.updateMapDisplay();
          }
        })
          .width('100%')
          .height('100%')
        
        // 跟踪状态指示器
        Row() {
          Circle({ width: 12, height: 12 })
            .fill(this.isTracking ? Color.Green : Color.Gray)
            .margin({ right: 8 })
          
          Text(this.isTracking ? '正在跟踪' : '未跟踪')
            .fontSize(12)
            .fontColor(Color.Black)
        }
        .padding(8)
        .backgroundColor(Color.White)
        .borderRadius(16)
        .margin(16)
        .alignSelf(ItemAlign.End)
        .alignRules({
          top: { anchor: '__container__', align: VerticalAlign.Top },
          right: { anchor: '__container__', align: HorizontalAlign.End }
        })
      }
      .layoutWeight(1)
      
      // 控制面板
      Row() {
        Button('地图')
          .type(ButtonType.Normal)
          .fontSize(12)
          .layoutWeight(1)
          .margin({ right: 4 })
          .onClick(() => this.toggleMapType())
        
        Button(this.showTrajectories ? '隐藏轨迹' : '显示轨迹')
          .type(ButtonType.Normal)
          .fontSize(12)
          .layoutWeight(1)
          .margin({ left: 4, right: 4 })
          .onClick(() => this.toggleTrajectoriesVisibility())
        
        Button('定位')
          .type(ButtonType.Normal)
          .fontSize(12)
          .layoutWeight(1)
          .margin({ left: 4, right: 4 })
          .onClick(() => this.centerToCurrentLocation())
        
        Button('全览')
          .type(ButtonType.Normal)
          .fontSize(12)
          .layoutWeight(1)
          .margin({ left: 4, right: 4 })
          .onClick(() => this.fitAllTrajectories())
        
        Button('清除')
          .type(ButtonType.Normal)
          .fontSize(12)
          .layoutWeight(1)
          .margin({ left: 4 })
          .onClick(() => this.clearMap())
      }
      .width('100%')
      .padding(8)
      .backgroundColor(Color.White)
      .shadow({
        radius: 8,
        color: Color.Gray,
        offsetY: -2
      })
    }
    .width('100%')
    .height('100%')
    .backgroundColor('#F5F5F5')
  }
}

/**
 * 地图组件
 */
@Component
struct MapComponent {
  @Prop mapOptions: map.MapOptions;
  @Prop mapCallback: (controller: map.MapController) => void;
  
  build() {
    MapContainer(this.mapOptions)
      .width('100%')
      .height('100%')
      .onMapReady((controller: map.MapController) => {
        this.mapCallback(controller);
      })
  }
}