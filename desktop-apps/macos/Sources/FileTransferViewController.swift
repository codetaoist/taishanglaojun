import Cocoa
import Foundation

// MARK: - File Transfer Models
struct FileTransferItem: Identifiable {
    let id: UUID
    var name: String
    var size: Int64
    var type: FileType
    var status: TransferStatus
    var progress: Double
    var speed: Int64
    var remainingTime: TimeInterval
    var localPath: String
    var remotePath: String
    var createdAt: Date
    var updatedAt: Date
    
    init(name: String, size: Int64, localPath: String, remotePath: String) {
        self.id = UUID()
        self.name = name
        self.size = size
        self.type = FileType.fromExtension(URL(fileURLWithPath: name).pathExtension)
        self.status = .pending
        self.progress = 0.0
        self.speed = 0
        self.remainingTime = 0
        self.localPath = localPath
        self.remotePath = remotePath
        self.createdAt = Date()
        self.updatedAt = Date()
    }
}

enum FileType: String, CaseIterable {
    case document = "文档"
    case image = "图片"
    case video = "视频"
    case audio = "音频"
    case archive = "压缩包"
    case code = "代码"
    case other = "其他"
    
    static func fromExtension(_ ext: String) -> FileType {
        let lowercased = ext.lowercased()
        switch lowercased {
        case "txt", "doc", "docx", "pdf", "rtf":
            return .document
        case "jpg", "jpeg", "png", "gif", "bmp", "svg":
            return .image
        case "mp4", "avi", "mov", "mkv", "wmv":
            return .video
        case "mp3", "wav", "flac", "aac", "ogg":
            return .audio
        case "zip", "rar", "7z", "tar", "gz":
            return .archive
        case "swift", "js", "py", "java", "cpp", "c", "h":
            return .code
        default:
            return .other
        }
    }
    
    var icon: String {
        switch self {
        case .document: return "📄"
        case .image: return "🖼️"
        case .video: return "🎬"
        case .audio: return "🎵"
        case .archive: return "📦"
        case .code: return "💻"
        case .other: return "📁"
        }
    }
}

enum TransferStatus: String, CaseIterable {
    case pending = "等待中"
    case uploading = "上传中"
    case downloading = "下载中"
    case completed = "已完成"
    case failed = "失败"
    case paused = "已暂停"
    case cancelled = "已取消"
    
    var color: NSColor {
        switch self {
        case .pending: return .systemYellow
        case .uploading, .downloading: return .systemBlue
        case .completed: return .systemGreen
        case .failed, .cancelled: return .systemRed
        case .paused: return .systemOrange
        }
    }
}

struct TransferSession: Identifiable {
    let id: UUID
    var name: String
    var items: [FileTransferItem]
    var totalSize: Int64
    var transferredSize: Int64
    var overallProgress: Double
    var createdAt: Date
    
    init(name: String) {
        self.id = UUID()
        self.name = name
        self.items = []
        self.totalSize = 0
        self.transferredSize = 0
        self.overallProgress = 0.0
        self.createdAt = Date()
    }
}

// MARK: - File Transfer View Controller
class FileTransferViewController: NSViewController {
    
    // MARK: - Properties
    private var transferSessions: [TransferSession] = []
    private var currentSession: TransferSession?
    private var transferItems: [FileTransferItem] = []
    private var filteredItems: [FileTransferItem] = []
    
    // MARK: - UI Components
    private lazy var splitView: NSSplitView = {
        let splitView = NSSplitView()
        splitView.isVertical = true
        splitView.dividerStyle = .thin
        return splitView
    }()
    
    private lazy var sessionSidebar: NSView = {
        let view = NSView()
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.controlBackgroundColor.cgColor
        return view
    }()
    
    private lazy var transferArea: NSView = {
        let view = NSView()
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.textBackgroundColor.cgColor
        return view
    }()
    
    private lazy var newSessionButton: NSButton = {
        let button = NSButton()
        button.title = "新建传输"
        button.bezelStyle = .rounded
        button.target = self
        button.action = #selector(newTransferSession(_:))
        return button
    }()
    
    private lazy var sessionTableView: NSTableView = {
        let tableView = NSTableView()
        tableView.delegate = self
        tableView.dataSource = self
        tableView.target = self
        tableView.action = #selector(selectSession(_:))
        
        let column = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("session"))
        column.title = "传输会话"
        tableView.addTableColumn(column)
        tableView.headerView = nil
        
        return tableView
    }()
    
    private lazy var sessionScrollView: NSScrollView = {
        let scrollView = NSScrollView()
        scrollView.documentView = sessionTableView
        scrollView.hasVerticalScroller = true
        scrollView.hasHorizontalScroller = false
        scrollView.autohidesScrollers = true
        return scrollView
    }()
    
    private lazy var toolbar: NSView = {
        let view = NSView()
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.controlBackgroundColor.cgColor
        return view
    }()
    
    private lazy var uploadButton: NSButton = {
        let button = NSButton()
        button.title = "上传文件"
        button.bezelStyle = .rounded
        button.target = self
        button.action = #selector(uploadFiles(_:))
        return button
    }()
    
    private lazy var downloadButton: NSButton = {
        let button = NSButton()
        button.title = "下载文件"
        button.bezelStyle = .rounded
        button.target = self
        button.action = #selector(downloadFiles(_:))
        return button
    }()
    
    private lazy var pauseAllButton: NSButton = {
        let button = NSButton()
        button.title = "暂停全部"
        button.bezelStyle = .rounded
        button.target = self
        button.action = #selector(pauseAllTransfers(_:))
        return button
    }()
    
    private lazy var resumeAllButton: NSButton = {
        let button = NSButton()
        button.title = "恢复全部"
        button.bezelStyle = .rounded
        button.target = self
        button.action = #selector(resumeAllTransfers(_:))
        return button
    }()
    
    private lazy var clearCompletedButton: NSButton = {
        let button = NSButton()
        button.title = "清除已完成"
        button.bezelStyle = .rounded
        button.target = self
        button.action = #selector(clearCompleted(_:))
        return button
    }()
    
    private lazy var filterPopup: NSPopUpButton = {
        let popup = NSPopUpButton()
        popup.addItem(withTitle: "全部文件")
        for status in TransferStatus.allCases {
            popup.addItem(withTitle: status.rawValue)
        }
        popup.target = self
        popup.action = #selector(filterItems(_:))
        return popup
    }()
    
    private lazy var transferTableView: NSTableView = {
        let tableView = NSTableView()
        tableView.delegate = self
        tableView.dataSource = self
        
        // 配置列
        let nameColumn = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("name"))
        nameColumn.title = "文件名"
        nameColumn.width = 200
        tableView.addTableColumn(nameColumn)
        
        let sizeColumn = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("size"))
        sizeColumn.title = "大小"
        sizeColumn.width = 80
        tableView.addTableColumn(sizeColumn)
        
        let statusColumn = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("status"))
        statusColumn.title = "状态"
        statusColumn.width = 80
        tableView.addTableColumn(statusColumn)
        
        let progressColumn = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("progress"))
        progressColumn.title = "进度"
        progressColumn.width = 120
        tableView.addTableColumn(progressColumn)
        
        let speedColumn = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("speed"))
        speedColumn.title = "速度"
        speedColumn.width = 80
        tableView.addTableColumn(speedColumn)
        
        let timeColumn = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("time"))
        timeColumn.title = "剩余时间"
        timeColumn.width = 80
        tableView.addTableColumn(timeColumn)
        
        return tableView
    }()
    
    private lazy var transferScrollView: NSScrollView = {
        let scrollView = NSScrollView()
        scrollView.documentView = transferTableView
        scrollView.hasVerticalScroller = true
        scrollView.hasHorizontalScroller = false
        scrollView.autohidesScrollers = true
        return scrollView
    }()
    
    private lazy var statusBar: NSView = {
        let view = NSView()
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.controlBackgroundColor.cgColor
        return view
    }()
    
    private lazy var overallProgressBar: NSProgressIndicator = {
        let progress = NSProgressIndicator()
        progress.style = .bar
        progress.isIndeterminate = false
        progress.minValue = 0
        progress.maxValue = 100
        return progress
    }()
    
    private lazy var statusLabel: NSTextField = {
        let label = NSTextField(labelWithString: "就绪")
        label.font = NSFont.systemFont(ofSize: 12)
        return label
    }()
    
    private lazy var speedLabel: NSTextField = {
        let label = NSTextField(labelWithString: "0 KB/s")
        label.font = NSFont.systemFont(ofSize: 12)
        label.textColor = NSColor.secondaryLabelColor
        return label
    }()
    
    // MARK: - Lifecycle
    override func loadView() {
        view = NSView()
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.windowBackgroundColor.cgColor
    }
    
    override func viewDidLoad() {
        super.viewDidLoad()
        setupUI()
        loadTransferSessions()
        createDefaultSession()
        setupDragAndDrop()
    }
    
    // MARK: - UI Setup
    private func setupUI() {
        setupSplitView()
        setupSessionSidebar()
        setupTransferArea()
        setupConstraints()
    }
    
    private func setupSplitView() {
        view.addSubview(splitView)
        splitView.addArrangedSubview(sessionSidebar)
        splitView.addArrangedSubview(transferArea)
        
        // 设置分割比例
        splitView.setPosition(200, ofDividerAt: 0)
    }
    
    private func setupSessionSidebar() {
        sessionSidebar.addSubview(newSessionButton)
        sessionSidebar.addSubview(sessionScrollView)
    }
    
    private func setupTransferArea() {
        transferArea.addSubview(toolbar)
        transferArea.addSubview(transferScrollView)
        transferArea.addSubview(statusBar)
        
        toolbar.addSubview(uploadButton)
        toolbar.addSubview(downloadButton)
        toolbar.addSubview(pauseAllButton)
        toolbar.addSubview(resumeAllButton)
        toolbar.addSubview(clearCompletedButton)
        toolbar.addSubview(filterPopup)
        
        statusBar.addSubview(overallProgressBar)
        statusBar.addSubview(statusLabel)
        statusBar.addSubview(speedLabel)
    }
    
    private func setupConstraints() {
        // 禁用自动调整大小
        [splitView, newSessionButton, sessionScrollView, toolbar, transferScrollView, statusBar,
         uploadButton, downloadButton, pauseAllButton, resumeAllButton, clearCompletedButton,
         filterPopup, overallProgressBar, statusLabel, speedLabel].forEach {
            $0.translatesAutoresizingMaskIntoConstraints = false
        }
        
        NSLayoutConstraint.activate([
            // Split View
            splitView.topAnchor.constraint(equalTo: view.topAnchor),
            splitView.leadingAnchor.constraint(equalTo: view.leadingAnchor),
            splitView.trailingAnchor.constraint(equalTo: view.trailingAnchor),
            splitView.bottomAnchor.constraint(equalTo: view.bottomAnchor),
            
            // Session Sidebar
            newSessionButton.topAnchor.constraint(equalTo: sessionSidebar.topAnchor, constant: 20),
            newSessionButton.leadingAnchor.constraint(equalTo: sessionSidebar.leadingAnchor, constant: 10),
            newSessionButton.trailingAnchor.constraint(equalTo: sessionSidebar.trailingAnchor, constant: -10),
            
            sessionScrollView.topAnchor.constraint(equalTo: newSessionButton.bottomAnchor, constant: 20),
            sessionScrollView.leadingAnchor.constraint(equalTo: sessionSidebar.leadingAnchor, constant: 10),
            sessionScrollView.trailingAnchor.constraint(equalTo: sessionSidebar.trailingAnchor, constant: -10),
            sessionScrollView.bottomAnchor.constraint(equalTo: sessionSidebar.bottomAnchor, constant: -20),
            
            // Transfer Area
            toolbar.topAnchor.constraint(equalTo: transferArea.topAnchor),
            toolbar.leadingAnchor.constraint(equalTo: transferArea.leadingAnchor),
            toolbar.trailingAnchor.constraint(equalTo: transferArea.trailingAnchor),
            toolbar.heightAnchor.constraint(equalToConstant: 50),
            
            transferScrollView.topAnchor.constraint(equalTo: toolbar.bottomAnchor),
            transferScrollView.leadingAnchor.constraint(equalTo: transferArea.leadingAnchor),
            transferScrollView.trailingAnchor.constraint(equalTo: transferArea.trailingAnchor),
            transferScrollView.bottomAnchor.constraint(equalTo: statusBar.topAnchor),
            
            statusBar.leadingAnchor.constraint(equalTo: transferArea.leadingAnchor),
            statusBar.trailingAnchor.constraint(equalTo: transferArea.trailingAnchor),
            statusBar.bottomAnchor.constraint(equalTo: transferArea.bottomAnchor),
            statusBar.heightAnchor.constraint(equalToConstant: 40),
            
            // Toolbar
            uploadButton.leadingAnchor.constraint(equalTo: toolbar.leadingAnchor, constant: 20),
            uploadButton.centerYAnchor.constraint(equalTo: toolbar.centerYAnchor),
            
            downloadButton.leadingAnchor.constraint(equalTo: uploadButton.trailingAnchor, constant: 10),
            downloadButton.centerYAnchor.constraint(equalTo: toolbar.centerYAnchor),
            
            pauseAllButton.leadingAnchor.constraint(equalTo: downloadButton.trailingAnchor, constant: 20),
            pauseAllButton.centerYAnchor.constraint(equalTo: toolbar.centerYAnchor),
            
            resumeAllButton.leadingAnchor.constraint(equalTo: pauseAllButton.trailingAnchor, constant: 10),
            resumeAllButton.centerYAnchor.constraint(equalTo: toolbar.centerYAnchor),
            
            clearCompletedButton.leadingAnchor.constraint(equalTo: resumeAllButton.trailingAnchor, constant: 20),
            clearCompletedButton.centerYAnchor.constraint(equalTo: toolbar.centerYAnchor),
            
            filterPopup.trailingAnchor.constraint(equalTo: toolbar.trailingAnchor, constant: -20),
            filterPopup.centerYAnchor.constraint(equalTo: toolbar.centerYAnchor),
            
            // Status Bar
            overallProgressBar.leadingAnchor.constraint(equalTo: statusBar.leadingAnchor, constant: 20),
            overallProgressBar.centerYAnchor.constraint(equalTo: statusBar.centerYAnchor),
            overallProgressBar.widthAnchor.constraint(equalToConstant: 200),
            
            statusLabel.leadingAnchor.constraint(equalTo: overallProgressBar.trailingAnchor, constant: 20),
            statusLabel.centerYAnchor.constraint(equalTo: statusBar.centerYAnchor),
            
            speedLabel.trailingAnchor.constraint(equalTo: statusBar.trailingAnchor, constant: -20),
            speedLabel.centerYAnchor.constraint(equalTo: statusBar.centerYAnchor)
        ])
    }
    
    private func setupDragAndDrop() {
        transferTableView.registerForDraggedTypes([.fileURL])
    }
    
    // MARK: - Data Management
    private func loadTransferSessions() {
        transferSessions = []
        sessionTableView.reloadData()
    }
    
    private func createDefaultSession() {
        let session = TransferSession(name: "默认传输")
        transferSessions.append(session)
        currentSession = session
        sessionTableView.reloadData()
        sessionTableView.selectRowIndexes(IndexSet(integer: 0), byExtendingSelection: false)
        updateTransferView()
    }
    
    private func updateTransferView() {
        guard let session = currentSession else {
            transferItems = []
            filteredItems = []
            transferTableView.reloadData()
            return
        }
        
        transferItems = session.items
        applyFilter()
        updateStatusBar()
    }
    
    private func applyFilter() {
        let selectedIndex = filterPopup.indexOfSelectedItem
        if selectedIndex == 0 {
            filteredItems = transferItems
        } else {
            let selectedStatus = TransferStatus.allCases[selectedIndex - 1]
            filteredItems = transferItems.filter { $0.status == selectedStatus }
        }
        transferTableView.reloadData()
    }
    
    private func updateStatusBar() {
        guard let session = currentSession else {
            statusLabel.stringValue = "就绪"
            speedLabel.stringValue = "0 KB/s"
            overallProgressBar.doubleValue = 0
            return
        }
        
        let activeItems = session.items.filter { $0.status == .uploading || $0.status == .downloading }
        let completedItems = session.items.filter { $0.status == .completed }
        
        if activeItems.isEmpty {
            statusLabel.stringValue = "就绪"
            speedLabel.stringValue = "0 KB/s"
        } else {
            statusLabel.stringValue = "传输中 (\(activeItems.count) 个文件)"
            let totalSpeed = activeItems.reduce(0) { $0 + $1.speed }
            speedLabel.stringValue = formatSpeed(totalSpeed)
        }
        
        if session.items.isEmpty {
            overallProgressBar.doubleValue = 0
        } else {
            let progress = Double(completedItems.count) / Double(session.items.count) * 100
            overallProgressBar.doubleValue = progress
        }
    }
    
    private func formatSize(_ bytes: Int64) -> String {
        let formatter = ByteCountFormatter()
        formatter.allowedUnits = [.useKB, .useMB, .useGB]
        formatter.countStyle = .file
        return formatter.string(fromByteCount: bytes)
    }
    
    private func formatSpeed(_ bytesPerSecond: Int64) -> String {
        let formatter = ByteCountFormatter()
        formatter.allowedUnits = [.useKB, .useMB, .useGB]
        formatter.countStyle = .file
        return formatter.string(fromByteCount: bytesPerSecond) + "/s"
    }
    
    private func formatTime(_ seconds: TimeInterval) -> String {
        if seconds < 60 {
            return String(format: "%.0f秒", seconds)
        } else if seconds < 3600 {
            return String(format: "%.0f分钟", seconds / 60)
        } else {
            return String(format: "%.1f小时", seconds / 3600)
        }
    }
    
    // MARK: - Actions
    @objc private func newTransferSession(_ sender: NSButton) {
        let alert = NSAlert()
        alert.messageText = "新建传输会话"
        alert.addButton(withTitle: "创建")
        alert.addButton(withTitle: "取消")
        
        let inputField = NSTextField(frame: NSRect(x: 0, y: 0, width: 200, height: 20))
        inputField.placeholderString = "会话名称"
        alert.accessoryView = inputField
        
        if alert.runModal() == .alertFirstButtonReturn {
            let name = inputField.stringValue.isEmpty ? "新传输会话" : inputField.stringValue
            let session = TransferSession(name: name)
            transferSessions.append(session)
            currentSession = session
            sessionTableView.reloadData()
            sessionTableView.selectRowIndexes(IndexSet(integer: transferSessions.count - 1), byExtendingSelection: false)
            updateTransferView()
        }
    }
    
    @objc private func selectSession(_ sender: NSTableView) {
        let selectedRow = sender.selectedRow
        guard selectedRow >= 0 && selectedRow < transferSessions.count else { return }
        
        currentSession = transferSessions[selectedRow]
        updateTransferView()
    }
    
    @objc private func uploadFiles(_ sender: NSButton) {
        let openPanel = NSOpenPanel()
        openPanel.allowsMultipleSelection = true
        openPanel.canChooseDirectories = true
        openPanel.canChooseFiles = true
        
        if openPanel.runModal() == .OK {
            for url in openPanel.urls {
                addFileToTransfer(url: url, isUpload: true)
            }
        }
    }
    
    @objc private func downloadFiles(_ sender: NSButton) {
        // 这里应该显示远程文件选择器
        // 暂时模拟添加下载任务
        let mockFiles = [
            ("document.pdf", Int64(1024 * 1024 * 5)),
            ("image.jpg", Int64(1024 * 512)),
            ("video.mp4", Int64(1024 * 1024 * 100))
        ]
        
        for (name, size) in mockFiles {
            let item = FileTransferItem(
                name: name,
                size: size,
                localPath: NSHomeDirectory() + "/Downloads/" + name,
                remotePath: "/remote/" + name
            )
            addTransferItem(item)
        }
    }
    
    @objc private func pauseAllTransfers(_ sender: NSButton) {
        guard let session = currentSession else { return }
        
        for i in 0..<session.items.count {
            if session.items[i].status == .uploading || session.items[i].status == .downloading {
                session.items[i].status = .paused
            }
        }
        
        updateTransferView()
    }
    
    @objc private func resumeAllTransfers(_ sender: NSButton) {
        guard let session = currentSession else { return }
        
        for i in 0..<session.items.count {
            if session.items[i].status == .paused {
                session.items[i].status = session.items[i].localPath.contains("Downloads") ? .downloading : .uploading
                simulateTransfer(item: session.items[i])
            }
        }
        
        updateTransferView()
    }
    
    @objc private func clearCompleted(_ sender: NSButton) {
        guard let session = currentSession else { return }
        
        session.items.removeAll { $0.status == .completed }
        updateTransferView()
    }
    
    @objc private func filterItems(_ sender: NSPopUpButton) {
        applyFilter()
    }
    
    private func addFileToTransfer(url: URL, isUpload: Bool) {
        do {
            let attributes = try FileManager.default.attributesOfItem(atPath: url.path)
            let size = attributes[.size] as? Int64 ?? 0
            
            let remotePath = "/remote/" + url.lastPathComponent
            let item = FileTransferItem(
                name: url.lastPathComponent,
                size: size,
                localPath: url.path,
                remotePath: remotePath
            )
            
            addTransferItem(item)
            
            if isUpload {
                startUpload(item: item)
            }
        } catch {
            print("无法获取文件信息: \(error)")
        }
    }
    
    private func addTransferItem(_ item: FileTransferItem) {
        guard let session = currentSession else { return }
        
        session.items.append(item)
        updateTransferView()
    }
    
    private func startUpload(item: FileTransferItem) {
        guard let session = currentSession,
              let index = session.items.firstIndex(where: { $0.id == item.id }) else { return }
        
        session.items[index].status = .uploading
        simulateTransfer(item: session.items[index])
        updateTransferView()
    }
    
    private func simulateTransfer(item: FileTransferItem) {
        // 模拟文件传输进度
        Timer.scheduledTimer(withTimeInterval: 0.5, repeats: true) { timer in
            guard let session = self.currentSession,
                  let index = session.items.firstIndex(where: { $0.id == item.id }) else {
                timer.invalidate()
                return
            }
            
            var currentItem = session.items[index]
            
            if currentItem.status != .uploading && currentItem.status != .downloading {
                timer.invalidate()
                return
            }
            
            currentItem.progress += 0.1
            currentItem.speed = Int64.random(in: 1024*100...1024*1024*2) // 100KB - 2MB/s
            currentItem.remainingTime = Double(currentItem.size) * (1 - currentItem.progress) / Double(currentItem.speed)
            
            if currentItem.progress >= 1.0 {
                currentItem.progress = 1.0
                currentItem.status = .completed
                currentItem.speed = 0
                currentItem.remainingTime = 0
                timer.invalidate()
            }
            
            session.items[index] = currentItem
            
            DispatchQueue.main.async {
                self.updateTransferView()
            }
        }
    }
}

// MARK: - Table View Data Source
extension FileTransferViewController: NSTableViewDataSource {
    func numberOfRows(in tableView: NSTableView) -> Int {
        if tableView == sessionTableView {
            return transferSessions.count
        } else {
            return filteredItems.count
        }
    }
}

// MARK: - Table View Delegate
extension FileTransferViewController: NSTableViewDelegate {
    func tableView(_ tableView: NSTableView, viewFor tableColumn: NSTableColumn?, row: Int) -> NSView? {
        if tableView == sessionTableView {
            guard row < transferSessions.count else { return nil }
            
            let session = transferSessions[row]
            
            let cellView = NSTableCellView()
            let textField = NSTextField()
            textField.isBordered = false
            textField.isEditable = false
            textField.backgroundColor = NSColor.clear
            textField.stringValue = session.name
            textField.font = NSFont.systemFont(ofSize: 13)
            
            cellView.addSubview(textField)
            cellView.textField = textField
            
            textField.translatesAutoresizingMaskIntoConstraints = false
            NSLayoutConstraint.activate([
                textField.leadingAnchor.constraint(equalTo: cellView.leadingAnchor, constant: 5),
                textField.trailingAnchor.constraint(equalTo: cellView.trailingAnchor, constant: -5),
                textField.centerYAnchor.constraint(equalTo: cellView.centerYAnchor)
            ])
            
            return cellView
        } else {
            guard row < filteredItems.count else { return nil }
            
            let item = filteredItems[row]
            let identifier = tableColumn?.identifier
            
            let cellView = NSTableCellView()
            
            switch identifier?.rawValue {
            case "name":
                let textField = NSTextField()
                textField.isBordered = false
                textField.isEditable = false
                textField.backgroundColor = NSColor.clear
                textField.stringValue = "\(item.type.icon) \(item.name)"
                textField.font = NSFont.systemFont(ofSize: 13)
                
                cellView.addSubview(textField)
                cellView.textField = textField
                
                textField.translatesAutoresizingMaskIntoConstraints = false
                NSLayoutConstraint.activate([
                    textField.leadingAnchor.constraint(equalTo: cellView.leadingAnchor, constant: 5),
                    textField.trailingAnchor.constraint(equalTo: cellView.trailingAnchor, constant: -5),
                    textField.centerYAnchor.constraint(equalTo: cellView.centerYAnchor)
                ])
                
            case "size":
                let textField = NSTextField()
                textField.isBordered = false
                textField.isEditable = false
                textField.backgroundColor = NSColor.clear
                textField.stringValue = formatSize(item.size)
                textField.font = NSFont.systemFont(ofSize: 12)
                
                cellView.addSubview(textField)
                cellView.textField = textField
                
                textField.translatesAutoresizingMaskIntoConstraints = false
                NSLayoutConstraint.activate([
                    textField.leadingAnchor.constraint(equalTo: cellView.leadingAnchor, constant: 5),
                    textField.trailingAnchor.constraint(equalTo: cellView.trailingAnchor, constant: -5),
                    textField.centerYAnchor.constraint(equalTo: cellView.centerYAnchor)
                ])
                
            case "status":
                let textField = NSTextField()
                textField.isBordered = false
                textField.isEditable = false
                textField.backgroundColor = NSColor.clear
                textField.stringValue = item.status.rawValue
                textField.font = NSFont.systemFont(ofSize: 12)
                textField.textColor = item.status.color
                
                cellView.addSubview(textField)
                cellView.textField = textField
                
                textField.translatesAutoresizingMaskIntoConstraints = false
                NSLayoutConstraint.activate([
                    textField.leadingAnchor.constraint(equalTo: cellView.leadingAnchor, constant: 5),
                    textField.trailingAnchor.constraint(equalTo: cellView.trailingAnchor, constant: -5),
                    textField.centerYAnchor.constraint(equalTo: cellView.centerYAnchor)
                ])
                
            case "progress":
                let progressBar = NSProgressIndicator()
                progressBar.style = .bar
                progressBar.isIndeterminate = false
                progressBar.minValue = 0
                progressBar.maxValue = 100
                progressBar.doubleValue = item.progress * 100
                
                cellView.addSubview(progressBar)
                
                progressBar.translatesAutoresizingMaskIntoConstraints = false
                NSLayoutConstraint.activate([
                    progressBar.leadingAnchor.constraint(equalTo: cellView.leadingAnchor, constant: 5),
                    progressBar.trailingAnchor.constraint(equalTo: cellView.trailingAnchor, constant: -5),
                    progressBar.centerYAnchor.constraint(equalTo: cellView.centerYAnchor),
                    progressBar.heightAnchor.constraint(equalToConstant: 16)
                ])
                
            case "speed":
                let textField = NSTextField()
                textField.isBordered = false
                textField.isEditable = false
                textField.backgroundColor = NSColor.clear
                textField.stringValue = item.speed > 0 ? formatSpeed(item.speed) : "-"
                textField.font = NSFont.systemFont(ofSize: 12)
                
                cellView.addSubview(textField)
                cellView.textField = textField
                
                textField.translatesAutoresizingMaskIntoConstraints = false
                NSLayoutConstraint.activate([
                    textField.leadingAnchor.constraint(equalTo: cellView.leadingAnchor, constant: 5),
                    textField.trailingAnchor.constraint(equalTo: cellView.trailingAnchor, constant: -5),
                    textField.centerYAnchor.constraint(equalTo: cellView.centerYAnchor)
                ])
                
            case "time":
                let textField = NSTextField()
                textField.isBordered = false
                textField.isEditable = false
                textField.backgroundColor = NSColor.clear
                textField.stringValue = item.remainingTime > 0 ? formatTime(item.remainingTime) : "-"
                textField.font = NSFont.systemFont(ofSize: 12)
                
                cellView.addSubview(textField)
                cellView.textField = textField
                
                textField.translatesAutoresizingMaskIntoConstraints = false
                NSLayoutConstraint.activate([
                    textField.leadingAnchor.constraint(equalTo: cellView.leadingAnchor, constant: 5),
                    textField.trailingAnchor.constraint(equalTo: cellView.trailingAnchor, constant: -5),
                    textField.centerYAnchor.constraint(equalTo: cellView.centerYAnchor)
                ])
                
            default:
                break
            }
            
            return cellView
        }
    }
}