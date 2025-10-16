//
//  DataExportView.swift
//  TaishanglaojunTracker
//
//  Created by Taishanglaojun Team
//

import SwiftUI
import Combine

/// 数据导出视图
struct DataExportView: View {
    
    // MARK: - State Properties
    @StateObject private var exportService = DataExportService.shared
    @State private var showingShareSheet = false
    @State private var showingImportPicker = false
    @State private var showingAlert = false
    @State private var alertMessage = ""
    @State private var selectedExportType: ExportType = .all
    
    // MARK: - Private Properties
    private var cancellables = Set<AnyCancellable>()
    
    var body: some View {
        NavigationView {
            ScrollView {
                VStack(spacing: 20) {
                    
                    // 导出选项
                    exportOptionsSection
                    
                    // 导出按钮
                    exportButtonSection
                    
                    // 进度显示
                    if exportService.isExporting {
                        progressSection
                    }
                    
                    // 最近导出
                    if let lastExportURL = exportService.lastExportURL {
                        lastExportSection(url: lastExportURL)
                    }
                    
                    // 导入数据
                    importSection
                    
                    Spacer()
                }
                .padding()
            }
            .navigationTitle("数据导出")
            .navigationBarTitleDisplayMode(.large)
            .alert("提示", isPresented: $showingAlert) {
                Button("确定") { }
            } message: {
                Text(alertMessage)
            }
            .sheet(isPresented: $showingShareSheet) {
                if let url = exportService.lastExportURL {
                    ShareSheet(items: [url])
                }
            }
            .sheet(isPresented: $showingImportPicker) {
                DocumentPicker(onDocumentPicked: handleImportFile)
            }
            .onReceive(exportService.$exportError) { error in
                if let error = error {
                    alertMessage = error
                    showingAlert = true
                    exportService.clearError()
                }
            }
        }
    }
    
    // MARK: - View Components
    
    /// 导出选项部分
    private var exportOptionsSection: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text("选择导出内容")
                .font(.headline)
                .foregroundColor(.primary)
            
            VStack(spacing: 12) {
                ExportOptionRow(
                    title: "全部数据",
                    description: "包含聊天记录、轨迹数据和用户设置",
                    icon: "doc.on.doc",
                    isSelected: selectedExportType == .all
                ) {
                    selectedExportType = .all
                }
                
                ExportOptionRow(
                    title: "聊天记录",
                    description: "仅导出对话和消息数据",
                    icon: "message",
                    isSelected: selectedExportType == .chat
                ) {
                    selectedExportType = .chat
                }
                
                ExportOptionRow(
                    title: "轨迹数据",
                    description: "仅导出位置和轨迹信息",
                    icon: "location",
                    isSelected: selectedExportType == .trajectory
                ) {
                    selectedExportType = .trajectory
                }
                
                ExportOptionRow(
                    title: "用户设置",
                    description: "仅导出应用配置信息",
                    icon: "gearshape",
                    isSelected: selectedExportType == .settings
                ) {
                    selectedExportType = .settings
                }
            }
        }
        .padding()
        .background(Color(.systemGray6))
        .cornerRadius(12)
    }
    
    /// 导出按钮部分
    private var exportButtonSection: some View {
        Button(action: performExport) {
            HStack {
                Image(systemName: "square.and.arrow.up")
                    .font(.title2)
                
                Text("开始导出")
                    .font(.headline)
            }
            .foregroundColor(.white)
            .frame(maxWidth: .infinity)
            .padding()
            .background(
                LinearGradient(
                    gradient: Gradient(colors: [Color.blue, Color.purple]),
                    startPoint: .leading,
                    endPoint: .trailing
                )
            )
            .cornerRadius(12)
        }
        .disabled(exportService.isExporting)
    }
    
    /// 进度显示部分
    private var progressSection: some View {
        VStack(spacing: 12) {
            Text("正在导出数据...")
                .font(.subheadline)
                .foregroundColor(.secondary)
            
            ProgressView(value: exportService.exportProgress)
                .progressViewStyle(LinearProgressViewStyle(tint: .blue))
            
            Text("\(Int(exportService.exportProgress * 100))%")
                .font(.caption)
                .foregroundColor(.secondary)
        }
        .padding()
        .background(Color(.systemGray6))
        .cornerRadius(12)
    }
    
    /// 最近导出部分
    private func lastExportSection(url: URL) -> some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("最近导出")
                .font(.headline)
                .foregroundColor(.primary)
            
            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    Text(url.lastPathComponent)
                        .font(.subheadline)
                        .foregroundColor(.primary)
                    
                    Text("大小: \(formatFileSize(exportService.getExportFileSize(url)))")
                        .font(.caption)
                        .foregroundColor(.secondary)
                    
                    Text("创建时间: \(formatDate(getFileCreationDate(url)))")
                        .font(.caption)
                        .foregroundColor(.secondary)
                }
                
                Spacer()
                
                HStack(spacing: 12) {
                    Button(action: { showingShareSheet = true }) {
                        Image(systemName: "square.and.arrow.up")
                            .foregroundColor(.blue)
                    }
                    
                    Button(action: { exportService.deleteExportFile(url) }) {
                        Image(systemName: "trash")
                            .foregroundColor(.red)
                    }
                }
            }
            .padding()
            .background(Color(.systemGray6))
            .cornerRadius(8)
        }
        .padding()
        .background(Color(.systemGray6))
        .cornerRadius(12)
    }
    
    /// 导入数据部分
    private var importSection: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text("导入数据")
                .font(.headline)
                .foregroundColor(.primary)
            
            Button(action: { showingImportPicker = true }) {
                HStack {
                    Image(systemName: "square.and.arrow.down")
                        .font(.title2)
                    
                    Text("选择文件导入")
                        .font(.headline)
                }
                .foregroundColor(.white)
                .frame(maxWidth: .infinity)
                .padding()
                .background(Color.green)
                .cornerRadius(12)
            }
            
            Text("支持导入之前导出的JSON格式数据文件")
                .font(.caption)
                .foregroundColor(.secondary)
        }
        .padding()
        .background(Color(.systemGray6))
        .cornerRadius(12)
    }
    
    // MARK: - Actions
    
    /// 执行导出
    private func performExport() {
        let publisher: AnyPublisher<URL, Error>
        
        switch selectedExportType {
        case .all:
            publisher = exportService.exportAllData()
        case .chat:
            publisher = exportService.exportChatData()
        case .trajectory:
            publisher = exportService.exportTrajectoryData()
        case .settings:
            publisher = exportService.exportUserSettings()
        }
        
        publisher
            .receive(on: DispatchQueue.main)
            .sink(
                receiveCompletion: { completion in
                    if case .finished = completion {
                        alertMessage = "数据导出成功！"
                        showingAlert = true
                    }
                },
                receiveValue: { url in
                    print("✅ 导出成功: \(url)")
                }
            )
            .store(in: &cancellables)
    }
    
    /// 处理导入文件
    private func handleImportFile(_ url: URL) {
        exportService.importData(from: url)
            .receive(on: DispatchQueue.main)
            .sink(
                receiveCompletion: { completion in
                    switch completion {
                    case .finished:
                        alertMessage = "数据导入成功！"
                        showingAlert = true
                    case .failure(let error):
                        alertMessage = "导入失败: \(error.localizedDescription)"
                        showingAlert = true
                    }
                },
                receiveValue: { _ in }
            )
            .store(in: &cancellables)
    }
    
    // MARK: - Helper Methods
    
    /// 格式化文件大小
    private func formatFileSize(_ bytes: Int64) -> String {
        let formatter = ByteCountFormatter()
        formatter.allowedUnits = [.useKB, .useMB, .useGB]
        formatter.countStyle = .file
        return formatter.string(fromByteCount: bytes)
    }
    
    /// 格式化日期
    private func formatDate(_ date: Date?) -> String {
        guard let date = date else { return "未知" }
        let formatter = DateFormatter()
        formatter.dateStyle = .medium
        formatter.timeStyle = .short
        return formatter.string(from: date)
    }
    
    /// 获取文件创建日期
    private func getFileCreationDate(_ url: URL) -> Date? {
        do {
            let attributes = try FileManager.default.attributesOfItem(atPath: url.path)
            return attributes[.creationDate] as? Date
        } catch {
            return nil
        }
    }
}

// MARK: - Supporting Views

/// 导出选项行
struct ExportOptionRow: View {
    let title: String
    let description: String
    let icon: String
    let isSelected: Bool
    let action: () -> Void
    
    var body: some View {
        Button(action: action) {
            HStack(spacing: 12) {
                Image(systemName: icon)
                    .font(.title2)
                    .foregroundColor(isSelected ? .white : .blue)
                    .frame(width: 30)
                
                VStack(alignment: .leading, spacing: 2) {
                    Text(title)
                        .font(.subheadline)
                        .fontWeight(.medium)
                        .foregroundColor(isSelected ? .white : .primary)
                    
                    Text(description)
                        .font(.caption)
                        .foregroundColor(isSelected ? .white.opacity(0.8) : .secondary)
                }
                
                Spacer()
                
                if isSelected {
                    Image(systemName: "checkmark.circle.fill")
                        .foregroundColor(.white)
                }
            }
            .padding()
            .background(isSelected ? Color.blue : Color(.systemGray5))
            .cornerRadius(8)
        }
        .buttonStyle(PlainButtonStyle())
    }
}

/// 分享表单
struct ShareSheet: UIViewControllerRepresentable {
    let items: [Any]
    
    func makeUIViewController(context: Context) -> UIActivityViewController {
        UIActivityViewController(activityItems: items, applicationActivities: nil)
    }
    
    func updateUIViewController(_ uiViewController: UIActivityViewController, context: Context) {}
}

/// 文档选择器
struct DocumentPicker: UIViewControllerRepresentable {
    let onDocumentPicked: (URL) -> Void
    
    func makeUIViewController(context: Context) -> UIDocumentPickerViewController {
        let picker = UIDocumentPickerViewController(forOpeningContentTypes: [.json, .data])
        picker.delegate = context.coordinator
        return picker
    }
    
    func updateUIViewController(_ uiViewController: UIDocumentPickerViewController, context: Context) {}
    
    func makeCoordinator() -> Coordinator {
        Coordinator(onDocumentPicked: onDocumentPicked)
    }
    
    class Coordinator: NSObject, UIDocumentPickerDelegate {
        let onDocumentPicked: (URL) -> Void
        
        init(onDocumentPicked: @escaping (URL) -> Void) {
            self.onDocumentPicked = onDocumentPicked
        }
        
        func documentPicker(_ controller: UIDocumentPickerViewController, didPickDocumentsAt urls: [URL]) {
            guard let url = urls.first else { return }
            onDocumentPicked(url)
        }
    }
}

// MARK: - Export Types
enum ExportType: CaseIterable {
    case all
    case chat
    case trajectory
    case settings
    
    var title: String {
        switch self {
        case .all: return "全部数据"
        case .chat: return "聊天记录"
        case .trajectory: return "轨迹数据"
        case .settings: return "用户设置"
        }
    }
}

// MARK: - Preview
struct DataExportView_Previews: PreviewProvider {
    static var previews: some View {
        DataExportView()
    }
}