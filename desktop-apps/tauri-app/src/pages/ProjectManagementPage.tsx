import React, { useState, useEffect } from 'react';
import {
  Plus,
  Edit,
  Trash2,
  MoreVertical,
  Search,
  RefreshCw,
  Folder,
  FileText,
  Users,
  User,
  Calendar,
  Target,
  CheckCircle,
  XCircle,
  AlertCircle,
  X,
  Save,
  Eye,
  Star,
  BarChart3,
} from 'lucide-react';
import { invoke } from '@tauri-apps/api/core';

interface Project {
  id: string;
  name: string;
  description: string;
  status: 'Active' | 'Completed' | 'Paused' | 'Cancelled';
  priority: 'Low' | 'Medium' | 'High' | 'Critical';
  project_type: 'Personal' | 'Team' | 'Client' | 'Internal';
  owner_id: string;
  owner_name: string;
  start_date: string;
  end_date?: string;
  progress: number;
  budget?: number;
  tags: string[];
  created_at: string;
  updated_at: string;
}

interface ProjectIssue {
  id: string;
  project_id: string;
  title: string;
  description: string;
  status: 'Open' | 'InProgress' | 'Resolved' | 'Closed' | 'Reopened';
  priority: 'Low' | 'Medium' | 'High' | 'Critical';
  issue_type: 'Bug' | 'Feature' | 'Task' | 'Epic' | 'Story';
  assignee_id?: string;
  assignee_name?: string;
  reporter_id: string;
  reporter_name: string;
  labels: string[];
  due_date?: string;
  estimated_hours?: number;
  actual_hours?: number;
  created_at: string;
  updated_at: string;
}

interface ProjectMember {
  id: string;
  project_id: string;
  user_id: string;
  username: string;
  display_name: string;
  email: string;
  role: 'Owner' | 'Admin' | 'Developer' | 'Tester' | 'Viewer';
  permissions: string[];
  joined_at: string;
}

interface ProjectMilestone {
  id: string;
  project_id: string;
  title: string;
  description: string;
  due_date: string;
  status: 'Pending' | 'InProgress' | 'Completed' | 'Overdue';
  progress: number;
  created_at: string;
  updated_at: string;
}

interface ProjectResponse {
  success: boolean;
  message: string;
  projects?: Project[];
  project?: Project;
  issues?: ProjectIssue[];
  issue?: ProjectIssue;
  members?: ProjectMember[];
  member?: ProjectMember;
  milestones?: ProjectMilestone[];
  milestone?: ProjectMilestone;
}

interface CreateProjectRequest {
  name: string;
  description: string;
  project_type: 'Personal' | 'Team' | 'Client' | 'Internal';
  priority: 'Low' | 'Medium' | 'High' | 'Critical';
  start_date: string;
  end_date?: string;
  budget?: number;
  tags: string[];
}

interface CreateIssueRequest {
  project_id: string;
  title: string;
  description: string;
  issue_type: 'Bug' | 'Feature' | 'Task' | 'Epic' | 'Story';
  priority: 'Low' | 'Medium' | 'High' | 'Critical';
  assignee_id?: string;
  labels: string[];
  due_date?: string;
  estimated_hours?: number;
}

const ProjectManagementPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'projects' | 'issues' | 'members' | 'milestones'>('projects');
  const [projects, setProjects] = useState<Project[]>([]);
  const [issues, setIssues] = useState<ProjectIssue[]>([]);
  const [members, setMembers] = useState<ProjectMember[]>([]);
  const [milestones, setMilestones] = useState<ProjectMilestone[]>([]);
  const [selectedProject, setSelectedProject] = useState<Project | null>(null);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error' | 'info'; text: string } | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterStatus, setFilterStatus] = useState<string>('all');
  const [filterPriority, setFilterPriority] = useState<string>('all');
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [showProjectDetails, setShowProjectDetails] = useState(false);
  const [itemToDelete, setItemToDelete] = useState<string | null>(null);

  // 创建项目表单
  const [createProjectForm, setCreateProjectForm] = useState<CreateProjectRequest>({
    name: '',
    description: '',
    project_type: 'Personal',
    priority: 'Medium',
    start_date: '',
    end_date: '',
    budget: 0,
    tags: [],
  });

  // 创建问题表单
  const [createIssueForm, setCreateIssueForm] = useState<CreateIssueRequest>({
    project_id: '',
    title: '',
    description: '',
    issue_type: 'Task',
    priority: 'Medium',
    assignee_id: '',
    labels: [],
    due_date: '',
    estimated_hours: 0,
  });

  useEffect(() => {
    loadProjects();
  }, []);

  useEffect(() => {
    if (selectedProject) {
      loadProjectData(selectedProject.id);
    }
  }, [selectedProject]);

  const loadProjects = async () => {
    setLoading(true);
    try {
      const response = await invoke<ProjectResponse>('project_get_all');
      if (response.success && response.projects) {
        setProjects(response.projects);
        if (response.projects.length > 0 && !selectedProject) {
          setSelectedProject(response.projects[0]);
        }
      } else {
        setMessage({ type: 'error', text: response.message });
      }
    } catch (error) {
      setMessage({ type: 'error', text: `加载项目失败: ${error}` });
    } finally {
      setLoading(false);
    }
  };

  const loadProjectData = async (projectId: string) => {
    try {
      // 加载项目问题
      const issuesResponse = await invoke<ProjectResponse>('project_get_issues', { projectId });
      if (issuesResponse.success && issuesResponse.issues) {
        setIssues(issuesResponse.issues);
      }

      // 加载项目成员
      const membersResponse = await invoke<ProjectResponse>('project_get_members', { projectId });
      if (membersResponse.success && membersResponse.members) {
        setMembers(membersResponse.members);
      }

      // 加载项目里程碑
      const milestonesResponse = await invoke<ProjectResponse>('project_get_milestones', { projectId });
      if (milestonesResponse.success && milestonesResponse.milestones) {
        setMilestones(milestonesResponse.milestones);
      }
    } catch (error) {
      setMessage({ type: 'error', text: `加载项目数据失败: ${error}` });
    }
  };

  const handleCreateProject = async () => {
    if (!createProjectForm.name || !createProjectForm.description) {
      setMessage({ type: 'error', text: '请填写项目名称和描述' });
      return;
    }

    setLoading(true);
    try {
      const response = await invoke<ProjectResponse>('project_create', {
        request: createProjectForm,
      });

      if (response.success) {
        setMessage({ type: 'success', text: '项目创建成功' });
        setShowCreateDialog(false);
        setCreateProjectForm({
          name: '',
          description: '',
          project_type: 'Personal',
          priority: 'Medium',
          start_date: '',
          end_date: '',
          budget: 0,
          tags: [],
        });
        loadProjects();
      } else {
        setMessage({ type: 'error', text: response.message });
      }
    } catch (error) {
      setMessage({ type: 'error', text: `创建项目失败: ${error}` });
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteProject = async () => {
    if (!itemToDelete) return;

    setLoading(true);
    try {
      const response = await invoke<ProjectResponse>('project_delete', {
        projectId: itemToDelete,
      });

      if (response.success) {
        setMessage({ type: 'success', text: '项目删除成功' });
        setShowDeleteDialog(false);
        setItemToDelete(null);
        if (selectedProject?.id === itemToDelete) {
          setSelectedProject(null);
        }
        loadProjects();
      } else {
        setMessage({ type: 'error', text: response.message });
      }
    } catch (error) {
      setMessage({ type: 'error', text: `删除项目失败: ${error}` });
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'Active':
      case 'Open':
      case 'InProgress':
        return 'bg-green-100 text-green-800';
      case 'Completed':
      case 'Resolved':
        return 'bg-blue-100 text-blue-800';
      case 'Paused':
      case 'Pending':
        return 'bg-yellow-100 text-yellow-800';
      case 'Cancelled':
      case 'Closed':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'Critical':
        return 'bg-red-100 text-red-800';
      case 'High':
        return 'bg-orange-100 text-orange-800';
      case 'Medium':
        return 'bg-yellow-100 text-yellow-800';
      case 'Low':
        return 'bg-green-100 text-green-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'Bug':
        return <XCircle className="h-4 w-4 text-red-500" />;
      case 'Feature':
        return <Star className="h-4 w-4 text-blue-500" />;
      case 'Task':
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'Epic':
        return <Target className="h-4 w-4 text-purple-500" />;
      case 'Story':
        return <FileText className="h-4 w-4 text-indigo-500" />;
      default:
        return <FileText className="h-4 w-4 text-gray-500" />;
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('zh-CN');
  };

  const getMessageIcon = (type: string) => {
    switch (type) {
      case 'success': return <CheckCircle className="h-5 w-5 text-green-500" />;
      case 'error': return <XCircle className="h-5 w-5 text-red-500" />;
      case 'info': return <AlertCircle className="h-5 w-5 text-blue-500" />;
      default: return <AlertCircle className="h-5 w-5 text-gray-500" />;
    }
  };

  const filteredProjects = projects.filter(project => {
    const matchesSearch = project.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         project.description.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus = filterStatus === 'all' || project.status === filterStatus;
    const matchesPriority = filterPriority === 'all' || project.priority === filterPriority;
    return matchesSearch && matchesStatus && matchesPriority;
  });

  const filteredIssues = issues.filter(issue => {
    const matchesSearch = issue.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         issue.description.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus = filterStatus === 'all' || issue.status === filterStatus;
    const matchesPriority = filterPriority === 'all' || issue.priority === filterPriority;
    return matchesSearch && matchesStatus && matchesPriority;
  });

  return (
    <div className="min-h-screen bg-background p-6">
      <div className="max-w-7xl mx-auto">
        {/* 头部 */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-foreground mb-2">项目管理</h1>
          <p className="text-muted-foreground">管理项目、任务、成员和里程碑</p>
        </div>

        {/* 消息提示 */}
        {message && (
          <div className={`mb-6 p-4 rounded-lg border flex items-center space-x-3 ${
            message.type === 'success' ? 'bg-green-50 border-green-200 text-green-800' :
            message.type === 'error' ? 'bg-red-50 border-red-200 text-red-800' :
            'bg-blue-50 border-blue-200 text-blue-800'
          }`}>
            {getMessageIcon(message.type)}
            <span>{message.text}</span>
            <button
              onClick={() => setMessage(null)}
              className="ml-auto text-current hover:opacity-70"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
          {/* 左侧项目列表 */}
          <div className="lg:col-span-1">
            <div className="bg-card border border-border rounded-lg p-4">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-semibold text-foreground">项目列表</h2>
                <button
                  onClick={() => setShowCreateDialog(true)}
                  className="bg-primary text-primary-foreground p-2 rounded-lg hover:bg-primary/90 transition-colors"
                >
                  <Plus className="h-4 w-4" />
                </button>
              </div>

              {/* 搜索框 */}
              <div className="relative mb-4">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <input
                  type="text"
                  placeholder="搜索项目..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="w-full pl-10 pr-4 py-2 border border-border rounded-lg bg-background text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                />
              </div>

              {/* 过滤器 */}
              <div className="flex space-x-2 mb-4">
                <select
                  value={filterStatus}
                  onChange={(e) => setFilterStatus(e.target.value)}
                  className="px-3 py-2 border border-border rounded-lg bg-background text-foreground text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                >
                  <option value="all">所有状态</option>
                  <option value="Active">进行中</option>
                  <option value="Completed">已完成</option>
                  <option value="Paused">暂停</option>
                  <option value="Cancelled">已取消</option>
                </select>
                <select
                  value={filterPriority}
                  onChange={(e) => setFilterPriority(e.target.value)}
                  className="px-3 py-2 border border-border rounded-lg bg-background text-foreground text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                >
                  <option value="all">所有优先级</option>
                  <option value="Low">低</option>
                  <option value="Medium">中</option>
                  <option value="High">高</option>
                  <option value="Critical">紧急</option>
                </select>
              </div>

              <div className="space-y-2">
                {filteredProjects.map((project) => (
                  <div
                    key={project.id}
                    onClick={() => setSelectedProject(project)}
                    className={`p-3 rounded-lg cursor-pointer transition-colors ${
                      selectedProject?.id === project.id
                        ? 'bg-primary/10 border border-primary'
                        : 'bg-secondary/50 hover:bg-secondary/70 border border-transparent'
                    }`}
                  >
                    <div className="flex items-center space-x-2 mb-2">
                      <Folder className="h-4 w-4 text-primary" />
                      <span className="font-medium text-foreground truncate">{project.name}</span>
                    </div>
                    <div className="flex items-center space-x-2">
                      <span className={`px-2 py-1 text-xs rounded-full ${getStatusColor(project.status)}`}>
                        {project.status}
                      </span>
                      <span className={`px-2 py-1 text-xs rounded-full ${getPriorityColor(project.priority)}`}>
                        {project.priority}
                      </span>
                    </div>
                    <div className="mt-2">
                      <div className="w-full bg-gray-200 rounded-full h-2">
                        <div
                          className="bg-primary h-2 rounded-full transition-all"
                          style={{ width: `${project.progress}%` }}
                        ></div>
                      </div>
                      <span className="text-xs text-muted-foreground">{project.progress}% 完成</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* 右侧主要内容 */}
          <div className="lg:col-span-3">
            {selectedProject ? (
              <div className="bg-card border border-border rounded-lg">
                {/* 项目头部信息 */}
                <div className="p-6 border-b border-border">
                  <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center space-x-3">
                      <Folder className="h-8 w-8 text-primary" />
                      <div>
                        <h2 className="text-2xl font-bold text-foreground">{selectedProject.name}</h2>
                        <p className="text-muted-foreground">{selectedProject.description}</p>
                      </div>
                    </div>
                    <div className="flex space-x-2">
                      <button
                        onClick={() => setShowProjectDetails(true)}
                        className="bg-secondary text-secondary-foreground p-2 rounded-lg hover:bg-secondary/80 transition-colors"
                      >
                        <Eye className="h-4 w-4" />
                      </button>
                      <button
                        onClick={() => setShowEditDialog(true)}
                        className="bg-secondary text-secondary-foreground p-2 rounded-lg hover:bg-secondary/80 transition-colors"
                      >
                        <Edit className="h-4 w-4" />
                      </button>
                    </div>
                  </div>

                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                    <div className="bg-secondary/50 rounded-lg p-3">
                      <div className="flex items-center space-x-2 mb-1">
                        <BarChart3 className="h-4 w-4 text-blue-500" />
                        <span className="text-sm font-medium text-foreground">进度</span>
                      </div>
                      <span className="text-lg font-bold text-foreground">{selectedProject.progress}%</span>
                    </div>
                    <div className="bg-secondary/50 rounded-lg p-3">
                      <div className="flex items-center space-x-2 mb-1">
                        <Users className="h-4 w-4 text-green-500" />
                        <span className="text-sm font-medium text-foreground">成员</span>
                      </div>
                      <span className="text-lg font-bold text-foreground">{members.length}</span>
                    </div>
                    <div className="bg-secondary/50 rounded-lg p-3">
                      <div className="flex items-center space-x-2 mb-1">
                        <FileText className="h-4 w-4 text-orange-500" />
                        <span className="text-sm font-medium text-foreground">问题</span>
                      </div>
                      <span className="text-lg font-bold text-foreground">{issues.length}</span>
                    </div>
                    <div className="bg-secondary/50 rounded-lg p-3">
                      <div className="flex items-center space-x-2 mb-1">
                        <Target className="h-4 w-4 text-purple-500" />
                        <span className="text-sm font-medium text-foreground">里程碑</span>
                      </div>
                      <span className="text-lg font-bold text-foreground">{milestones.length}</span>
                    </div>
                  </div>
                </div>

                {/* 标签页导航 */}
                <div className="border-b border-border">
                  <nav className="flex space-x-8 px-6">
                    {[
                      { id: 'projects', label: '项目概览', icon: Folder },
                      { id: 'issues', label: '问题跟踪', icon: FileText },
                      { id: 'members', label: '团队成员', icon: Users },
                      { id: 'milestones', label: '里程碑', icon: Target },
                    ].map((tab) => (
                      <button
                        key={tab.id}
                        onClick={() => setActiveTab(tab.id as any)}
                        className={`flex items-center space-x-2 py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                          activeTab === tab.id
                            ? 'border-primary text-primary'
                            : 'border-transparent text-muted-foreground hover:text-foreground hover:border-gray-300'
                        }`}
                      >
                        <tab.icon className="h-4 w-4" />
                        <span>{tab.label}</span>
                      </button>
                    ))}
                  </nav>
                </div>

                {/* 标签页内容 */}
                <div className="p-6">
                  {/* 项目概览 */}
                  {activeTab === 'projects' && (
                    <div className="space-y-6">
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <div>
                          <h3 className="text-lg font-semibold text-foreground mb-4">项目信息</h3>
                          <div className="space-y-3">
                            <div>
                              <label className="text-sm font-medium text-muted-foreground">项目类型</label>
                              <p className="text-foreground">{selectedProject.project_type}</p>
                            </div>
                            <div>
                              <label className="text-sm font-medium text-muted-foreground">负责人</label>
                              <p className="text-foreground">{selectedProject.owner_name}</p>
                            </div>
                            <div>
                              <label className="text-sm font-medium text-muted-foreground">开始日期</label>
                              <p className="text-foreground">{formatDate(selectedProject.start_date)}</p>
                            </div>
                            {selectedProject.end_date && (
                              <div>
                                <label className="text-sm font-medium text-muted-foreground">结束日期</label>
                                <p className="text-foreground">{formatDate(selectedProject.end_date)}</p>
                              </div>
                            )}
                          </div>
                        </div>
                        <div>
                          <h3 className="text-lg font-semibold text-foreground mb-4">项目标签</h3>
                          <div className="flex flex-wrap gap-2">
                            {selectedProject.tags.map((tag, index) => (
                              <span
                                key={index}
                                className="px-3 py-1 bg-primary/10 text-primary text-sm rounded-full"
                              >
                                {tag}
                              </span>
                            ))}
                          </div>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* 问题跟踪 */}
                  {activeTab === 'issues' && (
                    <div>
                      <div className="flex items-center justify-between mb-6">
                        <h3 className="text-lg font-semibold text-foreground">问题列表</h3>
                        <button
                          onClick={() => {
                            setCreateIssueForm({ ...createIssueForm, project_id: selectedProject.id });
                            setShowCreateDialog(true);
                          }}
                          className="bg-primary text-primary-foreground py-2 px-4 rounded-lg hover:bg-primary/90 transition-colors flex items-center space-x-2"
                        >
                          <Plus className="h-4 w-4" />
                          <span>创建问题</span>
                        </button>
                      </div>

                      <div className="space-y-4">
                        {filteredIssues.map((issue) => (
                          <div key={issue.id} className="bg-secondary/50 rounded-lg p-4">
                            <div className="flex items-start justify-between">
                              <div className="flex items-start space-x-3">
                                {getTypeIcon(issue.issue_type)}
                                <div className="flex-1">
                                  <h4 className="font-medium text-foreground">{issue.title}</h4>
                                  <p className="text-sm text-muted-foreground mt-1">{issue.description}</p>
                                  <div className="flex items-center space-x-4 mt-3">
                                    <span className={`px-2 py-1 text-xs rounded-full ${getStatusColor(issue.status)}`}>
                                      {issue.status}
                                    </span>
                                    <span className={`px-2 py-1 text-xs rounded-full ${getPriorityColor(issue.priority)}`}>
                                      {issue.priority}
                                    </span>
                                    {issue.assignee_name && (
                                      <div className="flex items-center space-x-1">
                                        <User className="h-3 w-3 text-muted-foreground" />
                                        <span className="text-xs text-muted-foreground">{issue.assignee_name}</span>
                                      </div>
                                    )}
                                    {issue.due_date && (
                                      <div className="flex items-center space-x-1">
                                        <Calendar className="h-3 w-3 text-muted-foreground" />
                                        <span className="text-xs text-muted-foreground">{formatDate(issue.due_date)}</span>
                                      </div>
                                    )}
                                  </div>
                                </div>
                              </div>
                              <button className="text-muted-foreground hover:text-foreground">
                                <MoreVertical className="h-4 w-4" />
                              </button>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* 团队成员 */}
                  {activeTab === 'members' && (
                    <div>
                      <div className="flex items-center justify-between mb-6">
                        <h3 className="text-lg font-semibold text-foreground">团队成员</h3>
                        <button className="bg-primary text-primary-foreground py-2 px-4 rounded-lg hover:bg-primary/90 transition-colors flex items-center space-x-2">
                          <Plus className="h-4 w-4" />
                          <span>邀请成员</span>
                        </button>
                      </div>

                      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        {members.map((member) => (
                          <div key={member.id} className="bg-secondary/50 rounded-lg p-4">
                            <div className="flex items-center space-x-3 mb-3">
                              <div className="w-10 h-10 bg-primary/10 rounded-full flex items-center justify-center">
                                <User className="h-5 w-5 text-primary" />
                              </div>
                              <div>
                                <h4 className="font-medium text-foreground">{member.display_name}</h4>
                                <p className="text-sm text-muted-foreground">@{member.username}</p>
                              </div>
                            </div>
                            <div className="space-y-2">
                              <div>
                                <span className="text-xs font-medium text-muted-foreground">角色</span>
                                <p className="text-sm text-foreground">{member.role}</p>
                              </div>
                              <div>
                                <span className="text-xs font-medium text-muted-foreground">加入时间</span>
                                <p className="text-sm text-foreground">{formatDate(member.joined_at)}</p>
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* 里程碑 */}
                  {activeTab === 'milestones' && (
                    <div>
                      <div className="flex items-center justify-between mb-6">
                        <h3 className="text-lg font-semibold text-foreground">项目里程碑</h3>
                        <button className="bg-primary text-primary-foreground py-2 px-4 rounded-lg hover:bg-primary/90 transition-colors flex items-center space-x-2">
                          <Plus className="h-4 w-4" />
                          <span>创建里程碑</span>
                        </button>
                      </div>

                      <div className="space-y-4">
                        {milestones.map((milestone) => (
                          <div key={milestone.id} className="bg-secondary/50 rounded-lg p-4">
                            <div className="flex items-start justify-between">
                              <div className="flex items-start space-x-3">
                                <Target className="h-5 w-5 text-purple-500 mt-1" />
                                <div className="flex-1">
                                  <h4 className="font-medium text-foreground">{milestone.title}</h4>
                                  <p className="text-sm text-muted-foreground mt-1">{milestone.description}</p>
                                  <div className="flex items-center space-x-4 mt-3">
                                    <span className={`px-2 py-1 text-xs rounded-full ${getStatusColor(milestone.status)}`}>
                                      {milestone.status}
                                    </span>
                                    <div className="flex items-center space-x-1">
                                      <Calendar className="h-3 w-3 text-muted-foreground" />
                                      <span className="text-xs text-muted-foreground">{formatDate(milestone.due_date)}</span>
                                    </div>
                                  </div>
                                  <div className="mt-3">
                                    <div className="w-full bg-gray-200 rounded-full h-2">
                                      <div
                                        className="bg-purple-500 h-2 rounded-full transition-all"
                                        style={{ width: `${milestone.progress}%` }}
                                      ></div>
                                    </div>
                                    <span className="text-xs text-muted-foreground">{milestone.progress}% 完成</span>
                                  </div>
                                </div>
                              </div>
                              <button className="text-muted-foreground hover:text-foreground">
                                <MoreVertical className="h-4 w-4" />
                              </button>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            ) : (
              <div className="bg-card border border-border rounded-lg p-12 text-center">
                <Folder className="h-16 w-16 text-muted-foreground mx-auto mb-4" />
                <h3 className="text-lg font-semibold text-foreground mb-2">选择一个项目</h3>
                <p className="text-muted-foreground">从左侧列表中选择一个项目来查看详细信息</p>
              </div>
            )}
          </div>
        </div>

        {/* 创建项目对话框 */}
        {showCreateDialog && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-card border border-border rounded-lg p-6 w-full max-w-2xl max-h-[80vh] overflow-y-auto">
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-lg font-semibold text-foreground">创建新项目</h2>
                <button
                  onClick={() => setShowCreateDialog(false)}
                  className="text-muted-foreground hover:text-foreground"
                >
                  <X className="h-5 w-5" />
                </button>
              </div>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">项目名称</label>
                  <input
                    type="text"
                    value={createProjectForm.name}
                    onChange={(e) => setCreateProjectForm({ ...createProjectForm, name: e.target.value })}
                    className="w-full px-4 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                    placeholder="输入项目名称"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">项目描述</label>
                  <textarea
                    value={createProjectForm.description}
                    onChange={(e) => setCreateProjectForm({ ...createProjectForm, description: e.target.value })}
                    className="w-full px-4 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                    rows={3}
                    placeholder="输入项目描述"
                  />
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-2">项目类型</label>
                    <select
                      value={createProjectForm.project_type}
                      onChange={(e) => setCreateProjectForm({ ...createProjectForm, project_type: e.target.value as any })}
                      className="w-full px-4 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                    >
                      <option value="Personal">个人项目</option>
                      <option value="Team">团队项目</option>
                      <option value="Client">客户项目</option>
                      <option value="Internal">内部项目</option>
                    </select>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-foreground mb-2">优先级</label>
                    <select
                      value={createProjectForm.priority}
                      onChange={(e) => setCreateProjectForm({ ...createProjectForm, priority: e.target.value as any })}
                      className="w-full px-4 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                    >
                      <option value="Low">低</option>
                      <option value="Medium">中</option>
                      <option value="High">高</option>
                      <option value="Critical">紧急</option>
                    </select>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-2">开始日期</label>
                    <input
                      type="date"
                      value={createProjectForm.start_date}
                      onChange={(e) => setCreateProjectForm({ ...createProjectForm, start_date: e.target.value })}
                      className="w-full px-4 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-foreground mb-2">结束日期（可选）</label>
                    <input
                      type="date"
                      value={createProjectForm.end_date}
                      onChange={(e) => setCreateProjectForm({ ...createProjectForm, end_date: e.target.value })}
                      className="w-full px-4 py-2 border border-border rounded-lg bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                    />
                  </div>
                </div>

                <div className="flex justify-end space-x-3">
                  <button
                    onClick={() => setShowCreateDialog(false)}
                    className="px-4 py-2 text-muted-foreground hover:text-foreground transition-colors"
                  >
                    取消
                  </button>
                  <button
                    onClick={handleCreateProject}
                    disabled={loading}
                    className="bg-primary text-primary-foreground py-2 px-4 rounded-lg hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center space-x-2"
                  >
                    {loading ? (
                      <RefreshCw className="h-4 w-4 animate-spin" />
                    ) : (
                      <Save className="h-4 w-4" />
                    )}
                    <span>{loading ? '创建中...' : '创建项目'}</span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* 编辑项目对话框 */}
        {showEditDialog && selectedProject && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-background border border-border rounded-lg p-6 w-full max-w-md">
              <div className="flex items-center justify-between mb-4">
                <div>
                  <h2 className="text-lg font-semibold text-foreground">编辑项目</h2>
                  <p className="text-sm text-muted-foreground">修改项目信息</p>
                </div>
                <button
                  onClick={() => setShowEditDialog(false)}
                  className="text-muted-foreground hover:text-foreground"
                >
                  <X className="h-5 w-5" />
                </button>
              </div>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">项目名称</label>
                  <input
                    type="text"
                    value={selectedProject.name}
                    readOnly
                    className="w-full px-4 py-2 border border-border rounded-lg bg-secondary text-foreground"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">项目描述</label>
                  <textarea
                    value={selectedProject.description}
                    readOnly
                    rows={3}
                    className="w-full px-4 py-2 border border-border rounded-lg bg-secondary text-foreground resize-none"
                  />
                </div>
              </div>

              <div className="flex justify-end space-x-3 mt-6">
                <button
                  onClick={() => setShowEditDialog(false)}
                  className="px-4 py-2 text-muted-foreground hover:text-foreground transition-colors"
                >
                  关闭
                </button>
              </div>
            </div>
          </div>
        )}

        {/* 项目详情对话框 */}
        {showProjectDetails && selectedProject && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-background border border-border rounded-lg p-6 w-full max-w-2xl max-h-[80vh] overflow-y-auto">
              <div className="flex items-center justify-between mb-4">
                <div>
                  <h2 className="text-lg font-semibold text-foreground">项目详情</h2>
                  <p className="text-sm text-muted-foreground">{selectedProject.name}</p>
                </div>
                <button
                  onClick={() => setShowProjectDetails(false)}
                  className="text-muted-foreground hover:text-foreground"
                >
                  <X className="h-5 w-5" />
                </button>
              </div>

              <div className="space-y-6">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">项目类型</label>
                    <p className="text-foreground">{selectedProject.project_type}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">优先级</label>
                    <p className="text-foreground">{selectedProject.priority}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">状态</label>
                    <p className="text-foreground">{selectedProject.status}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">进度</label>
                    <p className="text-foreground">{selectedProject.progress}%</p>
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-foreground mb-1">项目描述</label>
                  <p className="text-foreground">{selectedProject.description}</p>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">开始日期</label>
                    <p className="text-foreground">{selectedProject.start_date}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">结束日期</label>
                    <p className="text-foreground">{selectedProject.end_date || '未设置'}</p>
                  </div>
                </div>

                {selectedProject.tags && selectedProject.tags.length > 0 && (
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-2">标签</label>
                    <div className="flex flex-wrap gap-2">
                      {selectedProject.tags.map((tag, index) => (
                        <span
                          key={index}
                          className="px-2 py-1 bg-primary/10 text-primary text-xs rounded-full"
                        >
                          {tag}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>

              <div className="flex justify-end space-x-3 mt-6">
                <button
                  onClick={() => setShowProjectDetails(false)}
                  className="px-4 py-2 text-muted-foreground hover:text-foreground transition-colors"
                >
                  关闭
                </button>
              </div>
            </div>
          </div>
        )}

        {/* 删除确认对话框 */}
        {showDeleteDialog && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-card border border-border rounded-lg p-6 w-full max-w-md">
              <div className="flex items-center space-x-3 mb-4">
                <div className="w-10 h-10 bg-red-100 rounded-full flex items-center justify-center">
                  <Trash2 className="h-5 w-5 text-red-600" />
                </div>
                <div>
                  <h2 className="text-lg font-semibold text-foreground">确认删除</h2>
                  <p className="text-sm text-muted-foreground">此操作无法撤销</p>
                </div>
              </div>

              <p className="text-foreground mb-6">
                您确定要删除这个项目吗？这将同时删除所有相关的问题、成员和里程碑数据。
              </p>

              <div className="flex justify-end space-x-3">
                <button
                  onClick={() => setShowDeleteDialog(false)}
                  className="px-4 py-2 text-muted-foreground hover:text-foreground transition-colors"
                >
                  取消
                </button>
                <button
                  onClick={handleDeleteProject}
                  disabled={loading}
                  className="bg-red-500 text-white py-2 px-4 rounded-lg hover:bg-red-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center space-x-2"
                >
                  {loading ? (
                    <RefreshCw className="h-4 w-4 animate-spin" />
                  ) : (
                    <Trash2 className="h-4 w-4" />
                  )}
                  <span>{loading ? '删除中...' : '确认删除'}</span>
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default ProjectManagementPage;