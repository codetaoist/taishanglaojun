//
//  ContentView.swift
//  TaishanglaojunTracker
//
//  Created by Taishanglaojun Team
//

import SwiftUI
import MapKit

struct ContentView: View {
    @StateObject private var locationViewModel = LocationViewModel()
    @State private var selectedTab = 0
    
    var body: some View {
        TabView(selection: $selectedTab) {
            // 地图追踪页面
            MapTrackingView()
                .environmentObject(locationViewModel)
                .tabItem {
                    Image(systemName: "location.fill")
                    Text("追踪")
                }
                .tag(0)
            
            // 轨迹历史页面
            TrajectoryHistoryView()
                .environmentObject(locationViewModel)
                .tabItem {
                    Image(systemName: "list.bullet")
                    Text("历史")
                }
                .tag(1)
            
            // 统计页面
            StatisticsView()
                .environmentObject(locationViewModel)
                .tabItem {
                    Image(systemName: "chart.bar.fill")
                    Text("统计")
                }
                .tag(2)
            
            // 设置页面
            SettingsView()
                .environmentObject(locationViewModel)
                .tabItem {
                    Image(systemName: "gear")
                    Text("设置")
                }
                .tag(3)
        }
        .accentColor(.blue)
        .onAppear {
            setupAppearance()
        }
    }
    
    private func setupAppearance() {
        // 设置TabBar外观
        let appearance = UITabBarAppearance()
        appearance.configureWithOpaqueBackground()
        appearance.backgroundColor = UIColor.systemBackground
        
        UITabBar.appearance().standardAppearance = appearance
        UITabBar.appearance().scrollEdgeAppearance = appearance
    }
}

// MARK: - Map Tracking View
struct MapTrackingView: View {
    @EnvironmentObject var locationViewModel: LocationViewModel
    @State private var region = MKCoordinateRegion(
        center: CLLocationCoordinate2D(latitude: 39.9042, longitude: 116.4074), // 北京
        span: MKCoordinateSpan(latitudeDelta: 0.01, longitudeDelta: 0.01)
    )
    @State private var showingPermissionAlert = false
    
    var body: some View {
        NavigationView {
            ZStack {
                // 地图视图
                MapView(region: $region, trajectory: locationViewModel.currentTrajectory)
                    .ignoresSafeArea(.all, edges: .top)
                
                VStack {
                    Spacer()
                    
                    // 当前轨迹信息卡片
                    if let trajectoryInfo = locationViewModel.currentTrajectoryInfo {
                        TrajectoryInfoCard(info: trajectoryInfo)
                            .padding(.horizontal)
                    }
                    
                    // 控制按钮
                    TrackingControlsView()
                        .environmentObject(locationViewModel)
                        .padding()
                }
            }
            .navigationTitle("位置追踪")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button(action: centerOnCurrentLocation) {
                        Image(systemName: "location.circle.fill")
                            .foregroundColor(.blue)
                    }
                }
            }
        }
        .alert("位置权限", isPresented: $showingPermissionAlert) {
            Button("设置") {
                openAppSettings()
            }
            Button("取消", role: .cancel) { }
        } message: {
            Text("需要位置权限才能追踪您的轨迹。请在设置中允许位置访问。")
        }
        .onChange(of: locationViewModel.authorizationStatus) { status in
            handleAuthorizationStatusChange(status)
        }
        .onChange(of: locationViewModel.currentLocation) { location in
            updateMapRegion(location)
        }
    }
    
    private func centerOnCurrentLocation() {
        if let location = locationViewModel.currentLocation {
            withAnimation(.easeInOut(duration: 0.5)) {
                region.center = location.coordinate
            }
        }
    }
    
    private func handleAuthorizationStatusChange(_ status: CLAuthorizationStatus) {
        switch status {
        case .denied, .restricted:
            showingPermissionAlert = true
        case .authorizedWhenInUse, .authorizedAlways:
            showingPermissionAlert = false
        default:
            break
        }
    }
    
    private func updateMapRegion(_ location: CLLocation?) {
        guard let location = location else { return }
        
        // 只有在追踪时才自动更新地图中心
        if locationViewModel.isTracking {
            withAnimation(.easeInOut(duration: 0.3)) {
                region.center = location.coordinate
            }
        }
    }
    
    private func openAppSettings() {
        if let settingsUrl = URL(string: UIApplication.openSettingsURLString) {
            UIApplication.shared.open(settingsUrl)
        }
    }
}

// MARK: - Map View
struct MapView: UIViewRepresentable {
    @Binding var region: MKCoordinateRegion
    let trajectory: Trajectory?
    
    func makeUIView(context: Context) -> MKMapView {
        let mapView = MKMapView()
        mapView.delegate = context.coordinator
        mapView.showsUserLocation = true
        mapView.userTrackingMode = .none
        mapView.mapType = .standard
        return mapView
    }
    
    func updateUIView(_ mapView: MKMapView, context: Context) {
        mapView.setRegion(region, animated: true)
        
        // 更新轨迹显示
        updateTrajectoryOverlay(mapView)
    }
    
    func makeCoordinator() -> Coordinator {
        Coordinator(self)
    }
    
    private func updateTrajectoryOverlay(_ mapView: MKMapView) {
        // 清除现有覆盖层
        mapView.removeOverlays(mapView.overlays)
        
        guard let trajectory = trajectory,
              trajectory.points.count > 1 else {
            return
        }
        
        // 创建轨迹线
        let coordinates = trajectory.points.map { point in
            CLLocationCoordinate2D(latitude: point.latitude, longitude: point.longitude)
        }
        
        let polyline = MKPolyline(coordinates: coordinates, count: coordinates.count)
        mapView.addOverlay(polyline)
        
        // 添加起点和终点标记
        if let startPoint = trajectory.startLocation {
            let startAnnotation = TrajectoryAnnotation(
                coordinate: CLLocationCoordinate2D(latitude: startPoint.latitude, longitude: startPoint.longitude),
                title: "起点",
                type: .start
            )
            mapView.addAnnotation(startAnnotation)
        }
        
        if let endPoint = trajectory.endLocation, trajectory.isRecording == false {
            let endAnnotation = TrajectoryAnnotation(
                coordinate: CLLocationCoordinate2D(latitude: endPoint.latitude, longitude: endPoint.longitude),
                title: "终点",
                type: .end
            )
            mapView.addAnnotation(endAnnotation)
        }
    }
    
    class Coordinator: NSObject, MKMapViewDelegate {
        var parent: MapView
        
        init(_ parent: MapView) {
            self.parent = parent
        }
        
        func mapView(_ mapView: MKMapView, rendererFor overlay: MKOverlay) -> MKOverlayRenderer {
            if let polyline = overlay as? MKPolyline {
                let renderer = MKPolylineRenderer(polyline: polyline)
                renderer.strokeColor = .systemBlue
                renderer.lineWidth = 4
                return renderer
            }
            return MKOverlayRenderer(overlay: overlay)
        }
        
        func mapView(_ mapView: MKMapView, viewFor annotation: MKAnnotation) -> MKAnnotationView? {
            guard let trajectoryAnnotation = annotation as? TrajectoryAnnotation else {
                return nil
            }
            
            let identifier = "TrajectoryAnnotation"
            var annotationView = mapView.dequeueReusableAnnotationView(withIdentifier: identifier)
            
            if annotationView == nil {
                annotationView = MKAnnotationView(annotation: annotation, reuseIdentifier: identifier)
                annotationView?.canShowCallout = true
            } else {
                annotationView?.annotation = annotation
            }
            
            // 设置图标
            switch trajectoryAnnotation.type {
            case .start:
                annotationView?.image = UIImage(systemName: "play.circle.fill")?.withTintColor(.green, renderingMode: .alwaysOriginal)
            case .end:
                annotationView?.image = UIImage(systemName: "stop.circle.fill")?.withTintColor(.red, renderingMode: .alwaysOriginal)
            }
            
            return annotationView
        }
    }
}

// MARK: - Trajectory Annotation
class TrajectoryAnnotation: NSObject, MKAnnotation {
    let coordinate: CLLocationCoordinate2D
    let title: String?
    let type: AnnotationType
    
    enum AnnotationType {
        case start, end
    }
    
    init(coordinate: CLLocationCoordinate2D, title: String?, type: AnnotationType) {
        self.coordinate = coordinate
        self.title = title
        self.type = type
    }
}

// MARK: - Trajectory Info Card
struct TrajectoryInfoCard: View {
    let info: TrajectoryInfo
    
    var body: some View {
        VStack(spacing: 12) {
            HStack {
                Text(info.name)
                    .font(.headline)
                    .foregroundColor(.primary)
                Spacer()
                Text("\(info.pointCount) 个点")
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
            
            HStack(spacing: 20) {
                InfoItem(title: "距离", value: info.formattedDistance, icon: "ruler")
                InfoItem(title: "时间", value: info.formattedDuration, icon: "clock")
                InfoItem(title: "平均速度", value: info.formattedAverageSpeed, icon: "speedometer")
            }
        }
        .padding()
        .background(Color(.systemBackground))
        .cornerRadius(12)
        .shadow(color: .black.opacity(0.1), radius: 5, x: 0, y: 2)
    }
}

struct InfoItem: View {
    let title: String
    let value: String
    let icon: String
    
    var body: some View {
        VStack(spacing: 4) {
            Image(systemName: icon)
                .foregroundColor(.blue)
                .font(.caption)
            
            Text(value)
                .font(.system(.caption, design: .monospaced))
                .fontWeight(.semibold)
            
            Text(title)
                .font(.caption2)
                .foregroundColor(.secondary)
        }
        .frame(maxWidth: .infinity)
    }
}

// MARK: - Tracking Controls View
struct TrackingControlsView: View {
    @EnvironmentObject var locationViewModel: LocationViewModel
    @State private var showingExportOptions = false
    
    var body: some View {
        HStack(spacing: 16) {
            // 主控制按钮
            Button(action: toggleTracking) {
                HStack {
                    Image(systemName: trackingButtonIcon)
                        .font(.title2)
                    Text(trackingButtonText)
                        .fontWeight(.semibold)
                }
                .foregroundColor(.white)
                .frame(maxWidth: .infinity)
                .frame(height: 50)
                .background(trackingButtonColor)
                .cornerRadius(25)
            }
            
            // 导出按钮（仅在有轨迹时显示）
            if locationViewModel.currentTrajectory != nil {
                Button(action: { showingExportOptions = true }) {
                    Image(systemName: "square.and.arrow.up")
                        .font(.title2)
                        .foregroundColor(.white)
                        .frame(width: 50, height: 50)
                        .background(Color.gray)
                        .cornerRadius(25)
                }
            }
        }
        .actionSheet(isPresented: $showingExportOptions) {
            ActionSheet(
                title: Text("导出轨迹"),
                buttons: [
                    .default(Text("导出为 GPX")) {
                        exportAsGPX()
                    },
                    .default(Text("导出为 JSON")) {
                        exportAsJSON()
                    },
                    .cancel()
                ]
            )
        }
    }
    
    private var trackingButtonIcon: String {
        if locationViewModel.isTracking {
            return "stop.fill"
        } else {
            return "play.fill"
        }
    }
    
    private var trackingButtonText: String {
        if locationViewModel.isTracking {
            return "停止追踪"
        } else {
            return "开始追踪"
        }
    }
    
    private var trackingButtonColor: Color {
        if locationViewModel.isTracking {
            return .red
        } else {
            return .green
        }
    }
    
    private func toggleTracking() {
        if locationViewModel.isTracking {
            locationViewModel.stopTracking()
        } else {
            if locationViewModel.authorizationStatus == .notDetermined {
                locationViewModel.requestLocationPermission()
            } else {
                locationViewModel.startTracking()
            }
        }
    }
    
    private func exportAsGPX() {
        if let url = locationViewModel.exportCurrentTrajectoryAsGPX() {
            shareFile(url: url)
        }
    }
    
    private func exportAsJSON() {
        if let url = locationViewModel.exportCurrentTrajectoryAsJSON() {
            shareFile(url: url)
        }
    }
    
    private func shareFile(url: URL) {
        let activityViewController = UIActivityViewController(activityItems: [url], applicationActivities: nil)
        
        if let windowScene = UIApplication.shared.connectedScenes.first as? UIWindowScene,
           let window = windowScene.windows.first {
            window.rootViewController?.present(activityViewController, animated: true)
        }
    }
}

#Preview {
    ContentView()
}