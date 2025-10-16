import Cocoa
import Foundation

// MARK: - Project Model
struct Project: Codable, Identifiable {
    let id: UUID
    var name: String
    var description: String
    var type: ProjectType
    var status: ProjectStatus
    var createdAt: Date
    var updatedAt: Date
    var members: [ProjectMember]
    var tags: [String]
    var progress: Double
    var deadline: Date?
    
    init(name: String, description: String, type: ProjectType) {
        self.id = UUID()
        self.name = name
        self.description = description
        self.type = type
        self.status = .active
        self.createdAt = Date()
        self.updatedAt = Date()
        self.members = []
        self.tags = []
        self.progress = 0.0
        self.deadline = nil
    }
}

enum ProjectType: String, CaseIterable, Codable {
    case software = "软件开发"
    case research = "研究项目"
    case design = "设计项目"
    case education = "教育培训"
    case business = "商业项目"
    case personal = "个人项目"
}

enum ProjectStatus: String, CaseIterable, Codable {
    case active = "进行中"
    case completed = "已完成"
    case paused = "已暂停"
    case cancelled = "已取消"
}

struct ProjectMember: Codable, Identifiable {
    let id: UUID
    var name: String
    var email: String
    var role: String
    var avatar: String?
    var joinedAt: Date
}

// MARK: - Project View Controller
class ProjectViewController: NSViewController {
    
    // MARK: - Properties
    private var projects: [Project] = []
    private var filteredProjects: [Project] = []
    private var selectedProject: Project?
    
    // MARK: - UI Components
    private lazy var splitView: NSSplitView = {
        let splitView = NSSplitView()
        splitView.isVertical = true
        splitView.dividerStyle = .thin
        return splitView
    }()
    
    private lazy var sidebarView: NSView = {
        let view = NSView()
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.controlBackgroundColor.cgColor
        return view
    }()
    
    private lazy var contentView: NSView = {
        let view = NSView()
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.textBackgroundColor.cgColor
        return view
    }()
    
    private lazy var searchField: NSSearchField = {
        let field = NSSearchField()
        field.placeholderString = "搜索项目..."
        field.target = self
        field.action = #selector(searchProjects(_:))
        return field
    }()
    
    private lazy var filterPopup: NSPopUpButton = {
        let popup = NSPopUpButton()
        popup.addItem(withTitle: "所有项目")
        for type in ProjectType.allCases {
            popup.addItem(withTitle: type.rawValue)
        }
        popup.target = self
        popup.action = #selector(filterProjects(_:))
        return popup
    }()
    
    private lazy var addButton: NSButton = {
        let button = NSButton()
        button.title = "新建项目"
        button.bezelStyle = .rounded
        button.target = self
        button.action = #selector(addProject(_:))
        return button
    }()
    
    private lazy var projectTableView: NSTableView = {
        let tableView = NSTableView()
        tableView.delegate = self
        tableView.dataSource = self
        tableView.target = self
        tableView.doubleAction = #selector(openProject(_:))
        
        // 配置列
        let nameColumn = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("name"))
        nameColumn.title = "项目名称"
        nameColumn.width = 200
        tableView.addTableColumn(nameColumn)
        
        let typeColumn = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("type"))
        typeColumn.title = "类型"
        typeColumn.width = 100
        tableView.addTableColumn(typeColumn)
        
        let statusColumn = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("status"))
        statusColumn.title = "状态"
        statusColumn.width = 80
        tableView.addTableColumn(statusColumn)
        
        let progressColumn = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("progress"))
        progressColumn.title = "进度"
        progressColumn.width = 100
        tableView.addTableColumn(progressColumn)
        
        return tableView
    }()
    
    private lazy var scrollView: NSScrollView = {
        let scrollView = NSScrollView()
        scrollView.documentView = projectTableView
        scrollView.hasVerticalScroller = true
        scrollView.hasHorizontalScroller = false
        scrollView.autohidesScrollers = true
        return scrollView
    }()
    
    private lazy var detailView: NSView = {
        let view = NSView()
        view.wantsLayer = true
        view.layer?.backgroundColor = NSColor.textBackgroundColor.cgColor
        return view
    }()
    
    private lazy var projectNameLabel: NSTextField = {
        let label = NSTextField(labelWithString: "选择一个项目")
        label.font = NSFont.systemFont(ofSize: 24, weight: .bold)
        return label
    }()
    
    private lazy var projectDescriptionLabel: NSTextField = {
        let label = NSTextField(labelWithString: "")
        label.font = NSFont.systemFont(ofSize: 14)
        label.textColor = NSColor.secondaryLabelColor
        label.maximumNumberOfLines = 0
        label.lineBreakMode = .byWordWrapping
        return label
    }()
    
    private lazy var progressIndicator: NSProgressIndicator = {
        let indicator = NSProgressIndicator()
        indicator.style = .bar
        indicator.isIndeterminate = false
        return indicator
    }()
    
    private lazy var membersLabel: NSTextField = {
        let label = NSTextField(labelWithString: "团队成员")
        label.font = NSFont.systemFont(ofSize: 16, weight: .semibold)
        return label
    }()
    
    private lazy var membersStackView: NSStackView = {
        let stackView = NSStackView()
        stackView.orientation = .vertical
        stackView.spacing = 8
        return stackView
    }()
    
    private lazy var editButton: NSButton = {
        let button = NSButton()
        button.title = "编辑项目"
        button.bezelStyle = .rounded
        button.target = self
        button.action = #selector(editProject(_:))
        return button
    }()
    
    private lazy var deleteButton: NSButton = {
        let button = NSButton()
        button.title = "删除项目"
        button.bezelStyle = .rounded
        button.target = self
        button.action = #selector(deleteProject(_:))
        return button
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
        loadProjects()
    }
    
    // MARK: - UI Setup
    private func setupUI() {
        setupSplitView()
        setupSidebar()
        setupDetailView()
        setupConstraints()
    }
    
    private func setupSplitView() {
        view.addSubview(splitView)
        splitView.addArrangedSubview(sidebarView)
        splitView.addArrangedSubview(contentView)
        
        // 设置分割比例
        splitView.setPosition(300, ofDividerAt: 0)
    }
    
    private func setupSidebar() {
        sidebarView.addSubview(searchField)
        sidebarView.addSubview(filterPopup)
        sidebarView.addSubview(addButton)
        sidebarView.addSubview(scrollView)
    }
    
    private func setupDetailView() {
        contentView.addSubview(detailView)
        detailView.addSubview(projectNameLabel)
        detailView.addSubview(projectDescriptionLabel)
        detailView.addSubview(progressIndicator)
        detailView.addSubview(membersLabel)
        detailView.addSubview(membersStackView)
        detailView.addSubview(editButton)
        detailView.addSubview(deleteButton)
    }
    
    private func setupConstraints() {
        // 禁用自动调整大小
        [splitView, searchField, filterPopup, addButton, scrollView, detailView,
         projectNameLabel, projectDescriptionLabel, progressIndicator,
         membersLabel, membersStackView, editButton, deleteButton].forEach {
            $0.translatesAutoresizingMaskIntoConstraints = false
        }
        
        NSLayoutConstraint.activate([
            // Split View
            splitView.topAnchor.constraint(equalTo: view.topAnchor),
            splitView.leadingAnchor.constraint(equalTo: view.leadingAnchor),
            splitView.trailingAnchor.constraint(equalTo: view.trailingAnchor),
            splitView.bottomAnchor.constraint(equalTo: view.bottomAnchor),
            
            // Sidebar constraints
            searchField.topAnchor.constraint(equalTo: sidebarView.topAnchor, constant: 20),
            searchField.leadingAnchor.constraint(equalTo: sidebarView.leadingAnchor, constant: 20),
            searchField.trailingAnchor.constraint(equalTo: sidebarView.trailingAnchor, constant: -20),
            
            filterPopup.topAnchor.constraint(equalTo: searchField.bottomAnchor, constant: 10),
            filterPopup.leadingAnchor.constraint(equalTo: sidebarView.leadingAnchor, constant: 20),
            filterPopup.widthAnchor.constraint(equalToConstant: 120),
            
            addButton.topAnchor.constraint(equalTo: searchField.bottomAnchor, constant: 10),
            addButton.trailingAnchor.constraint(equalTo: sidebarView.trailingAnchor, constant: -20),
            addButton.leadingAnchor.constraint(equalTo: filterPopup.trailingAnchor, constant: 10),
            
            scrollView.topAnchor.constraint(equalTo: filterPopup.bottomAnchor, constant: 20),
            scrollView.leadingAnchor.constraint(equalTo: sidebarView.leadingAnchor, constant: 20),
            scrollView.trailingAnchor.constraint(equalTo: sidebarView.trailingAnchor, constant: -20),
            scrollView.bottomAnchor.constraint(equalTo: sidebarView.bottomAnchor, constant: -20),
            
            // Detail View constraints
            detailView.topAnchor.constraint(equalTo: contentView.topAnchor),
            detailView.leadingAnchor.constraint(equalTo: contentView.leadingAnchor),
            detailView.trailingAnchor.constraint(equalTo: contentView.trailingAnchor),
            detailView.bottomAnchor.constraint(equalTo: contentView.bottomAnchor),
            
            // Detail content constraints
            projectNameLabel.topAnchor.constraint(equalTo: detailView.topAnchor, constant: 40),
            projectNameLabel.leadingAnchor.constraint(equalTo: detailView.leadingAnchor, constant: 40),
            projectNameLabel.trailingAnchor.constraint(equalTo: detailView.trailingAnchor, constant: -40),
            
            projectDescriptionLabel.topAnchor.constraint(equalTo: projectNameLabel.bottomAnchor, constant: 20),
            projectDescriptionLabel.leadingAnchor.constraint(equalTo: detailView.leadingAnchor, constant: 40),
            projectDescriptionLabel.trailingAnchor.constraint(equalTo: detailView.trailingAnchor, constant: -40),
            
            progressIndicator.topAnchor.constraint(equalTo: projectDescriptionLabel.bottomAnchor, constant: 30),
            progressIndicator.leadingAnchor.constraint(equalTo: detailView.leadingAnchor, constant: 40),
            progressIndicator.trailingAnchor.constraint(equalTo: detailView.trailingAnchor, constant: -40),
            progressIndicator.heightAnchor.constraint(equalToConstant: 20),
            
            membersLabel.topAnchor.constraint(equalTo: progressIndicator.bottomAnchor, constant: 30),
            membersLabel.leadingAnchor.constraint(equalTo: detailView.leadingAnchor, constant: 40),
            
            membersStackView.topAnchor.constraint(equalTo: membersLabel.bottomAnchor, constant: 15),
            membersStackView.leadingAnchor.constraint(equalTo: detailView.leadingAnchor, constant: 40),
            membersStackView.trailingAnchor.constraint(equalTo: detailView.trailingAnchor, constant: -40),
            
            editButton.bottomAnchor.constraint(equalTo: detailView.bottomAnchor, constant: -40),
            editButton.leadingAnchor.constraint(equalTo: detailView.leadingAnchor, constant: 40),
            editButton.widthAnchor.constraint(equalToConstant: 100),
            
            deleteButton.bottomAnchor.constraint(equalTo: detailView.bottomAnchor, constant: -40),
            deleteButton.leadingAnchor.constraint(equalTo: editButton.trailingAnchor, constant: 20),
            deleteButton.widthAnchor.constraint(equalToConstant: 100)
        ])
    }
    
    // MARK: - Data Management
    private func loadProjects() {
        // 模拟加载项目数据
        projects = [
            Project(name: "太上老君AI助手", description: "基于大语言模型的智能助手应用，提供多模态交互和个性化服务", type: .software),
            Project(name: "桌面宠物系统", description: "可爱的桌面宠物，具备AI对话和情感交互功能", type: .software),
            Project(name: "跨平台文件同步", description: "安全高效的多设备文件同步解决方案", type: .software)
        ]
        
        // 添加一些示例成员
        for i in 0..<projects.count {
            projects[i].members = [
                ProjectMember(id: UUID(), name: "张三", email: "zhangsan@example.com", role: "项目经理", avatar: nil, joinedAt: Date()),
                ProjectMember(id: UUID(), name: "李四", email: "lisi@example.com", role: "开发工程师", avatar: nil, joinedAt: Date())
            ]
            projects[i].progress = Double.random(in: 0.1...0.9)
        }
        
        filteredProjects = projects
        projectTableView.reloadData()
    }
    
    private func updateDetailView() {
        guard let project = selectedProject else {
            projectNameLabel.stringValue = "选择一个项目"
            projectDescriptionLabel.stringValue = ""
            progressIndicator.doubleValue = 0
            membersStackView.arrangedSubviews.forEach { $0.removeFromSuperview() }
            editButton.isEnabled = false
            deleteButton.isEnabled = false
            return
        }
        
        projectNameLabel.stringValue = project.name
        projectDescriptionLabel.stringValue = project.description
        progressIndicator.doubleValue = project.progress * 100
        
        // 更新成员列表
        membersStackView.arrangedSubviews.forEach { $0.removeFromSuperview() }
        for member in project.members {
            let memberView = createMemberView(member: member)
            membersStackView.addArrangedSubview(memberView)
        }
        
        editButton.isEnabled = true
        deleteButton.isEnabled = true
    }
    
    private func createMemberView(member: ProjectMember) -> NSView {
        let containerView = NSView()
        
        let nameLabel = NSTextField(labelWithString: member.name)
        nameLabel.font = NSFont.systemFont(ofSize: 14, weight: .medium)
        
        let roleLabel = NSTextField(labelWithString: member.role)
        roleLabel.font = NSFont.systemFont(ofSize: 12)
        roleLabel.textColor = NSColor.secondaryLabelColor
        
        containerView.addSubview(nameLabel)
        containerView.addSubview(roleLabel)
        
        nameLabel.translatesAutoresizingMaskIntoConstraints = false
        roleLabel.translatesAutoresizingMaskIntoConstraints = false
        
        NSLayoutConstraint.activate([
            nameLabel.topAnchor.constraint(equalTo: containerView.topAnchor),
            nameLabel.leadingAnchor.constraint(equalTo: containerView.leadingAnchor),
            
            roleLabel.topAnchor.constraint(equalTo: nameLabel.bottomAnchor, constant: 2),
            roleLabel.leadingAnchor.constraint(equalTo: containerView.leadingAnchor),
            roleLabel.bottomAnchor.constraint(equalTo: containerView.bottomAnchor)
        ])
        
        return containerView
    }
    
    // MARK: - Actions
    @objc private func searchProjects(_ sender: NSSearchField) {
        let searchText = sender.stringValue.lowercased()
        if searchText.isEmpty {
            filteredProjects = projects
        } else {
            filteredProjects = projects.filter { project in
                project.name.lowercased().contains(searchText) ||
                project.description.lowercased().contains(searchText)
            }
        }
        projectTableView.reloadData()
    }
    
    @objc private func filterProjects(_ sender: NSPopUpButton) {
        let selectedIndex = sender.indexOfSelectedItem
        if selectedIndex == 0 {
            filteredProjects = projects
        } else {
            let selectedType = ProjectType.allCases[selectedIndex - 1]
            filteredProjects = projects.filter { $0.type == selectedType }
        }
        projectTableView.reloadData()
    }
    
    @objc private func addProject(_ sender: NSButton) {
        showProjectEditor(project: nil)
    }
    
    @objc private func editProject(_ sender: NSButton) {
        guard let project = selectedProject else { return }
        showProjectEditor(project: project)
    }
    
    @objc private func deleteProject(_ sender: NSButton) {
        guard let project = selectedProject,
              let index = projects.firstIndex(where: { $0.id == project.id }) else { return }
        
        let alert = NSAlert()
        alert.messageText = "删除项目"
        alert.informativeText = "确定要删除项目 \"\(project.name)\" 吗？此操作无法撤销。"
        alert.addButton(withTitle: "删除")
        alert.addButton(withTitle: "取消")
        alert.alertStyle = .warning
        
        if alert.runModal() == .alertFirstButtonReturn {
            projects.remove(at: index)
            filteredProjects = projects
            projectTableView.reloadData()
            selectedProject = nil
            updateDetailView()
        }
    }
    
    @objc private func openProject(_ sender: NSTableView) {
        let selectedRow = sender.selectedRow
        guard selectedRow >= 0 && selectedRow < filteredProjects.count else { return }
        
        let project = filteredProjects[selectedRow]
        // 这里可以打开项目详情窗口或切换到项目工作区
        print("打开项目: \(project.name)")
    }
    
    private func showProjectEditor(project: Project?) {
        // 这里应该显示项目编辑器窗口
        // 暂时使用简单的输入对话框
        let alert = NSAlert()
        alert.messageText = project == nil ? "新建项目" : "编辑项目"
        alert.addButton(withTitle: "保存")
        alert.addButton(withTitle: "取消")
        
        let inputView = NSView(frame: NSRect(x: 0, y: 0, width: 300, height: 100))
        
        let nameField = NSTextField(frame: NSRect(x: 0, y: 60, width: 300, height: 20))
        nameField.placeholderString = "项目名称"
        nameField.stringValue = project?.name ?? ""
        
        let descField = NSTextField(frame: NSRect(x: 0, y: 30, width: 300, height: 20))
        descField.placeholderString = "项目描述"
        descField.stringValue = project?.description ?? ""
        
        let typePopup = NSPopUpButton(frame: NSRect(x: 0, y: 0, width: 150, height: 25))
        for type in ProjectType.allCases {
            typePopup.addItem(withTitle: type.rawValue)
        }
        if let project = project {
            typePopup.selectItem(withTitle: project.type.rawValue)
        }
        
        inputView.addSubview(nameField)
        inputView.addSubview(descField)
        inputView.addSubview(typePopup)
        
        alert.accessoryView = inputView
        
        if alert.runModal() == .alertFirstButtonReturn {
            let name = nameField.stringValue
            let description = descField.stringValue
            let type = ProjectType.allCases[typePopup.indexOfSelectedItem]
            
            if !name.isEmpty {
                if var existingProject = project {
                    // 编辑现有项目
                    existingProject.name = name
                    existingProject.description = description
                    existingProject.type = type
                    existingProject.updatedAt = Date()
                    
                    if let index = projects.firstIndex(where: { $0.id == existingProject.id }) {
                        projects[index] = existingProject
                    }
                } else {
                    // 创建新项目
                    let newProject = Project(name: name, description: description, type: type)
                    projects.append(newProject)
                }
                
                filteredProjects = projects
                projectTableView.reloadData()
            }
        }
    }
}

// MARK: - Table View Data Source
extension ProjectViewController: NSTableViewDataSource {
    func numberOfRows(in tableView: NSTableView) -> Int {
        return filteredProjects.count
    }
}

// MARK: - Table View Delegate
extension ProjectViewController: NSTableViewDelegate {
    func tableView(_ tableView: NSTableView, viewFor tableColumn: NSTableColumn?, row: Int) -> NSView? {
        guard row < filteredProjects.count else { return nil }
        
        let project = filteredProjects[row]
        let identifier = tableColumn?.identifier
        
        let cellView = NSTableCellView()
        let textField = NSTextField()
        textField.isBordered = false
        textField.isEditable = false
        textField.backgroundColor = NSColor.clear
        
        switch identifier?.rawValue {
        case "name":
            textField.stringValue = project.name
            textField.font = NSFont.systemFont(ofSize: 13, weight: .medium)
        case "type":
            textField.stringValue = project.type.rawValue
            textField.font = NSFont.systemFont(ofSize: 12)
        case "status":
            textField.stringValue = project.status.rawValue
            textField.font = NSFont.systemFont(ofSize: 12)
        case "progress":
            textField.stringValue = String(format: "%.0f%%", project.progress * 100)
            textField.font = NSFont.systemFont(ofSize: 12)
        default:
            textField.stringValue = ""
        }
        
        cellView.addSubview(textField)
        cellView.textField = textField
        
        textField.translatesAutoresizingMaskIntoConstraints = false
        NSLayoutConstraint.activate([
            textField.leadingAnchor.constraint(equalTo: cellView.leadingAnchor, constant: 5),
            textField.trailingAnchor.constraint(equalTo: cellView.trailingAnchor, constant: -5),
            textField.centerYAnchor.constraint(equalTo: cellView.centerYAnchor)
        ])
        
        return cellView
    }
    
    func tableViewSelectionDidChange(_ notification: Notification) {
        let selectedRow = projectTableView.selectedRow
        if selectedRow >= 0 && selectedRow < filteredProjects.count {
            selectedProject = filteredProjects[selectedRow]
        } else {
            selectedProject = nil
        }
        updateDetailView()
    }
}