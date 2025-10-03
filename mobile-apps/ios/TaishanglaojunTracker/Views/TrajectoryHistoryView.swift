//
//  TrajectoryHistoryView.swift
//  TaishanglaojunTracker
//
//  Created by Taishanglaojun Team
//

import SwiftUI
import MapKit

struct TrajectoryHistoryView: View {
    @EnvironmentObject var locationViewModel: LocationViewModel
    @StateObject private var dataService = DataService.shared
    @State private var selectedTrajectory: Trajectory?
    @State private var showingTrajectoryDetail = false
    @State private var searchText = ""
    @State private var selectedDateRange: DateRange = .all
    @State private var showingDatePicker = false
    @State private var customStartDate = Date()
    @State private var customEndDate = Date()
    
    var body: some View {
        NavigationView {
            VStack(spacing: 0) {
                // 搜索和筛选栏
                SearchAndFilterBar(
                    searchText: $searchText,
                    selectedDateRange: $selectedDateRange,
                    showingDatePicker: $showingDatePicker
                )
                
                // 轨迹列表
                if dataService.isLoading {
                    LoadingView()
                } else if filteredTrajectories.isEmpty {
                    EmptyStateView()
                } else {
                    TrajectoryList(
                        trajectories: filteredTrajectories,
                        selectedTrajectory: $selectedTrajectory,
                        showingDetail: $showingTrajectoryDetail
                    )
                }
            }
            .navigationTitle("轨迹历史")
            .navigationBarTitleDisplayMode(.large)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Menu {
                        Button("刷新", action: refreshTrajectories)
                        Button("导出全部", action: exportAllTrajectories)
                        Button("清理缓存", action: clearCache)
                    } label: {
                        Image(systemName: "ellipsis.circle")
                    }
                }
            }
            .sheet(isPresented: $showingTrajectoryDetail) {
                if let trajectory = selectedTrajectory {
                    TrajectoryDetailView(trajectory: trajectory)
                }
            }
            .sheet(isPresented: $showingDatePicker) {
                DateRangePickerView(
                    startDate: $customStartDate,
                    endDate: $customEndDate,
                    isPresented: $showingDatePicker
                )
            }
        }
    }
    
    // MARK: - Computed Properties
    
    private var filteredTrajectories: [Trajectory] {
        var trajectories = dataService.trajectories
        
        // 文本搜索
        if !searchText.isEmpty {
            trajectories = trajectories.filter { trajectory in
                trajectory.name.localizedCaseInsensitiveContains(searchText)
            }
        }
        
        // 日期筛选
        trajectories = trajectories.filter { trajectory in
            isTrajectoryInDateRange(trajectory)
        }
        
        return trajectories.sorted { $0.startTime > $1.startTime }
    }
    
    private func isTrajectoryInDateRange(_ trajectory: Trajectory) -> Bool {
        let calendar = Calendar.current
        let now = Date()
        
        switch selectedDateRange {
        case .all:
            return true
        case .today:
            return calendar.isDate(trajectory.startTime, inSameDayAs: now)
        case .thisWeek:
            let weekAgo = calendar.date(byAdding: .weekOfYear, value: -1, to: now) ?? now
            return trajectory.startTime >= weekAgo
        case .thisMonth:
            let monthAgo = calendar.date(byAdding: .month, value: -1, to: now) ?? now
            return trajectory.startTime >= monthAgo
        case .custom:
            return trajectory.startTime >= customStartDate && trajectory.startTime <= customEndDate
        }
    }
    
    // MARK: - Actions
    
    private func refreshTrajectories() {
        // 触发数据刷新
        dataService.syncWithServer()
    }
    
    private func exportAllTrajectories() {
        // 导出所有轨迹的逻辑
        print("导出所有轨迹")
    }
    
    private func clearCache() {
        // 清理缓存的逻辑
        print("清理缓存")
    }
}

// MARK: - Search and Filter Bar
struct SearchAndFilterBar: View {
    @Binding var searchText: String
    @Binding var selectedDateRange: DateRange
    @Binding var showingDatePicker: Bool
    
    var body: some View {
        VStack(spacing: 12) {
            // 搜索框
            HStack {
                Image(systemName: "magnifyingglass")
                    .foregroundColor(.gray)
                
                TextField("搜索轨迹...", text: $searchText)
                    .textFieldStyle(PlainTextFieldStyle())
                
                if !searchText.isEmpty {
                    Button(action: { searchText = "" }) {
                        Image(systemName: "xmark.circle.fill")
                            .foregroundColor(.gray)
                    }
                }
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 8)
            .background(Color(.systemGray6))
            .cornerRadius(10)
            
            // 日期筛选
            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 12) {
                    ForEach(DateRange.allCases, id: \.self) { range in
                        DateRangeButton(
                            range: range,
                            isSelected: selectedDateRange == range,
                            action: {
                                if range == .custom {
                                    showingDatePicker = true
                                } else {
                                    selectedDateRange = range
                                }
                            }
                        )
                    }
                }
                .padding(.horizontal)
            }
        }
        .padding()
        .background(Color(.systemBackground))
    }
}

struct DateRangeButton: View {
    let range: DateRange
    let isSelected: Bool
    let action: () -> Void
    
    var body: some View {
        Button(action: action) {
            Text(range.displayName)
                .font(.caption)
                .fontWeight(.medium)
                .foregroundColor(isSelected ? .white : .primary)
                .padding(.horizontal, 12)
                .padding(.vertical, 6)
                .background(isSelected ? Color.blue : Color(.systemGray5))
                .cornerRadius(15)
        }
    }
}

// MARK: - Trajectory List
struct TrajectoryList: View {
    let trajectories: [Trajectory]
    @Binding var selectedTrajectory: Trajectory?
    @Binding var showingDetail: Bool
    
    var body: some View {
        List {
            ForEach(trajectories, id: \.id) { trajectory in
                TrajectoryRowView(trajectory: trajectory)
                    .onTapGesture {
                        selectedTrajectory = trajectory
                        showingDetail = true
                    }
                    .swipeActions(edge: .trailing, allowsFullSwipe: false) {
                        Button("删除", role: .destructive) {
                            deleteTrajectory(trajectory)
                        }
                        
                        Button("分享") {
                            shareTrajectory(trajectory)
                        }
                        .tint(.blue)
                    }
            }
        }
        .listStyle(PlainListStyle())
    }
    
    private func deleteTrajectory(_ trajectory: Trajectory) {
        DataService.shared.deleteTrajectory(trajectory)
    }
    
    private func shareTrajectory(_ trajectory: Trajectory) {
        if let url = DataService.shared.exportTrajectoryAsGPX(trajectory) {
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

// MARK: - Trajectory Row View
struct TrajectoryRowView: View {
    let trajectory: Trajectory
    
    var body: some View {
        HStack(spacing: 12) {
            // 轨迹缩略图
            TrajectoryThumbnail(trajectory: trajectory)
                .frame(width: 60, height: 60)
            
            // 轨迹信息
            VStack(alignment: .leading, spacing: 4) {
                Text(trajectory.name)
                    .font(.headline)
                    .lineLimit(1)
                
                Text(formatDate(trajectory.startTime))
                    .font(.caption)
                    .foregroundColor(.secondary)
                
                HStack(spacing: 16) {
                    Label(trajectory.formattedDistance, systemImage: "ruler")
                    Label(trajectory.formattedDuration, systemImage: "clock")
                }
                .font(.caption)
                .foregroundColor(.secondary)
            }
            
            Spacer()
            
            // 状态指示器
            VStack(spacing: 4) {
                if trajectory.isSynced {
                    Image(systemName: "checkmark.circle.fill")
                        .foregroundColor(.green)
                } else {
                    Image(systemName: "arrow.clockwise.circle")
                        .foregroundColor(.orange)
                }
                
                Text("\(trajectory.points.count)")
                    .font(.caption2)
                    .foregroundColor(.secondary)
            }
        }
        .padding(.vertical, 8)
    }
    
    private func formatDate(_ date: Date) -> String {
        let formatter = DateFormatter()
        formatter.dateStyle = .medium
        formatter.timeStyle = .short
        return formatter.string(from: date)
    }
}

// MARK: - Trajectory Thumbnail
struct TrajectoryThumbnail: View {
    let trajectory: Trajectory
    
    var body: some View {
        ZStack {
            RoundedRectangle(cornerRadius: 8)
                .fill(Color(.systemGray6))
            
            if trajectory.points.count > 1 {
                // 简化的轨迹线
                TrajectoryPath(points: trajectory.points)
                    .stroke(Color.blue, lineWidth: 2)
                    .clipped()
            } else {
                Image(systemName: "location.fill")
                    .foregroundColor(.gray)
            }
        }
    }
}

struct TrajectoryPath: Shape {
    let points: [LocationPoint]
    
    func path(in rect: CGRect) -> Path {
        guard points.count > 1 else { return Path() }
        
        // 计算边界
        let latitudes = points.map { $0.latitude }
        let longitudes = points.map { $0.longitude }
        
        guard let minLat = latitudes.min(),
              let maxLat = latitudes.max(),
              let minLon = longitudes.min(),
              let maxLon = longitudes.max() else {
            return Path()
        }
        
        let latRange = maxLat - minLat
        let lonRange = maxLon - minLon
        
        // 防止除零
        let safeLatRange = latRange > 0 ? latRange : 0.001
        let safeLonRange = lonRange > 0 ? lonRange : 0.001
        
        var path = Path()
        
        for (index, point) in points.enumerated() {
            let x = CGFloat((point.longitude - minLon) / safeLonRange) * rect.width
            let y = CGFloat(1 - (point.latitude - minLat) / safeLatRange) * rect.height
            
            if index == 0 {
                path.move(to: CGPoint(x: x, y: y))
            } else {
                path.addLine(to: CGPoint(x: x, y: y))
            }
        }
        
        return path
    }
}

// MARK: - Loading View
struct LoadingView: View {
    var body: some View {
        VStack(spacing: 16) {
            ProgressView()
                .scaleEffect(1.2)
            
            Text("加载轨迹中...")
                .font(.caption)
                .foregroundColor(.secondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

// MARK: - Empty State View
struct EmptyStateView: View {
    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: "location.slash")
                .font(.system(size: 48))
                .foregroundColor(.gray)
            
            Text("暂无轨迹记录")
                .font(.headline)
                .foregroundColor(.primary)
            
            Text("开始追踪您的第一条轨迹吧！")
                .font(.caption)
                .foregroundColor(.secondary)
                .multilineTextAlignment(.center)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .padding()
    }
}

// MARK: - Date Range Picker
struct DateRangePickerView: View {
    @Binding var startDate: Date
    @Binding var endDate: Date
    @Binding var isPresented: Bool
    
    var body: some View {
        NavigationView {
            VStack(spacing: 20) {
                DatePicker("开始日期", selection: $startDate, displayedComponents: .date)
                    .datePickerStyle(GraphicalDatePickerStyle())
                
                DatePicker("结束日期", selection: $endDate, displayedComponents: .date)
                    .datePickerStyle(GraphicalDatePickerStyle())
                
                Spacer()
            }
            .padding()
            .navigationTitle("选择日期范围")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarLeading) {
                    Button("取消") {
                        isPresented = false
                    }
                }
                
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button("确定") {
                        isPresented = false
                    }
                }
            }
        }
    }
}

// MARK: - Supporting Types
enum DateRange: CaseIterable {
    case all, today, thisWeek, thisMonth, custom
    
    var displayName: String {
        switch self {
        case .all: return "全部"
        case .today: return "今天"
        case .thisWeek: return "本周"
        case .thisMonth: return "本月"
        case .custom: return "自定义"
        }
    }
}

#Preview {
    TrajectoryHistoryView()
        .environmentObject(LocationViewModel())
}