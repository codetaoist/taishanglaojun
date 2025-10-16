use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use tauri::State;
use reqwest::Client;
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ProjectStatus {
    Planning,
    Active,
    OnHold,
    Completed,
    Cancelled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ProjectPriority {
    Low,
    Medium,
    High,
    Critical,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ProjectType {
    Development,
    Research,
    Marketing,
    Operations,
    Other,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum IssueStatus {
    Open,
    InProgress,
    Review,
    Testing,
    Closed,
    Reopened,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum IssuePriority {
    Low,
    Medium,
    High,
    Critical,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum IssueType {
    Bug,
    Feature,
    Task,
    Epic,
    Story,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ProjectRole {
    Owner,
    Manager,
    Developer,
    Tester,
    Viewer,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Project {
    pub id: String,
    pub name: String,
    pub description: String,
    pub status: ProjectStatus,
    pub priority: ProjectPriority,
    pub project_type: ProjectType,
    pub owner_id: String,
    pub manager_id: Option<String>,
    pub organization_id: Option<String>,
    pub start_date: Option<String>,
    pub end_date: Option<String>,
    pub budget: Option<f64>,
    pub progress: f64,
    pub tags: Vec<String>,
    pub labels: Vec<String>,
    pub metadata: HashMap<String, String>,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProjectIssue {
    pub id: String,
    pub project_id: String,
    pub title: String,
    pub description: String,
    pub issue_type: IssueType,
    pub status: IssueStatus,
    pub priority: IssuePriority,
    pub assignee_id: Option<String>,
    pub reporter_id: String,
    pub milestone_id: Option<String>,
    pub labels: Vec<String>,
    pub tags: Vec<String>,
    pub estimated_hours: Option<f64>,
    pub actual_hours: Option<f64>,
    pub due_date: Option<String>,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProjectMember {
    pub id: String,
    pub project_id: String,
    pub user_id: String,
    pub username: String,
    pub email: String,
    pub role: ProjectRole,
    pub joined_at: String,
    pub permissions: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProjectMilestone {
    pub id: String,
    pub project_id: String,
    pub name: String,
    pub description: String,
    pub due_date: String,
    pub completed_at: Option<String>,
    pub progress: f64,
    pub created_by: String,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IssueComment {
    pub id: String,
    pub issue_id: String,
    pub author_id: String,
    pub author_username: String,
    pub content: String,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IssueAttachment {
    pub id: String,
    pub issue_id: String,
    pub filename: String,
    pub file_url: String,
    pub file_size: u64,
    pub content_type: String,
    pub uploaded_by: String,
    pub uploaded_at: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct CreateProjectRequest {
    pub name: String,
    pub description: String,
    pub project_type: ProjectType,
    pub priority: ProjectPriority,
    pub organization_id: Option<String>,
    pub manager_id: Option<String>,
    pub start_date: Option<String>,
    pub end_date: Option<String>,
    pub budget: Option<f64>,
    pub tags: Vec<String>,
    pub labels: Vec<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct UpdateProjectRequest {
    pub name: Option<String>,
    pub description: Option<String>,
    pub status: Option<ProjectStatus>,
    pub priority: Option<ProjectPriority>,
    pub manager_id: Option<String>,
    pub start_date: Option<String>,
    pub end_date: Option<String>,
    pub budget: Option<f64>,
    pub tags: Option<Vec<String>>,
    pub labels: Option<Vec<String>>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct CreateIssueRequest {
    pub project_id: String,
    pub title: String,
    pub description: String,
    pub issue_type: IssueType,
    pub priority: IssuePriority,
    pub assignee_id: Option<String>,
    pub milestone_id: Option<String>,
    pub labels: Vec<String>,
    pub tags: Vec<String>,
    pub estimated_hours: Option<f64>,
    pub due_date: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct UpdateIssueRequest {
    pub title: Option<String>,
    pub description: Option<String>,
    pub status: Option<IssueStatus>,
    pub priority: Option<IssuePriority>,
    pub assignee_id: Option<String>,
    pub milestone_id: Option<String>,
    pub labels: Option<Vec<String>>,
    pub tags: Option<Vec<String>>,
    pub estimated_hours: Option<f64>,
    pub actual_hours: Option<f64>,
    pub due_date: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct CreateMilestoneRequest {
    pub project_id: String,
    pub name: String,
    pub description: String,
    pub due_date: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct AddMemberRequest {
    pub project_id: String,
    pub user_id: String,
    pub role: ProjectRole,
    pub permissions: Vec<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct AddCommentRequest {
    pub issue_id: String,
    pub content: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProjectResponse {
    pub success: bool,
    pub message: String,
    pub projects: Option<Vec<Project>>,
    pub project: Option<Project>,
    pub issues: Option<Vec<ProjectIssue>>,
    pub issue: Option<ProjectIssue>,
    pub members: Option<Vec<ProjectMember>>,
    pub milestones: Option<Vec<ProjectMilestone>>,
    pub comments: Option<Vec<IssueComment>>,
    pub attachments: Option<Vec<IssueAttachment>>,
}

#[derive(Debug, Clone)]
pub struct ProjectManagerState {
    pub projects: HashMap<String, Project>,
    pub issues: HashMap<String, Vec<ProjectIssue>>, // project_id -> issues
    pub members: HashMap<String, Vec<ProjectMember>>, // project_id -> members
    pub milestones: HashMap<String, Vec<ProjectMilestone>>, // project_id -> milestones
    pub comments: HashMap<String, Vec<IssueComment>>, // issue_id -> comments
    pub current_user_id: Option<String>,
}

impl Default for ProjectManagerState {
    fn default() -> Self {
        Self {
            projects: HashMap::new(),
            issues: HashMap::new(),
            members: HashMap::new(),
            milestones: HashMap::new(),
            comments: HashMap::new(),
            current_user_id: None,
        }
    }
}

pub struct ProjectManager {
    pub state: Arc<Mutex<ProjectManagerState>>,
    pub client: Client,
    pub api_base_url: String,
}

impl ProjectManager {
    pub fn new() -> Self {
        Self {
            state: Arc::new(Mutex::new(ProjectManagerState::default())),
            client: Client::new(),
            api_base_url: "http://localhost:8082".to_string(),
        }
    }

    pub async fn create_project(&self, auth_token: &str, request: CreateProjectRequest) -> Result<ProjectResponse, String> {
        let url = format!("{}/api/projects", self.api_base_url);
        
        let response = self.client
            .post(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .json(&request)
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("创建项目失败: HTTP {}", response.status()));
        }

        let project_response: ProjectResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if project_response.success {
            if let Some(project) = &project_response.project {
                let mut state = self.state.lock().unwrap();
                state.projects.insert(project.id.clone(), project.clone());
            }
        }

        Ok(project_response)
    }

    pub async fn update_project(&self, auth_token: &str, project_id: &str, request: UpdateProjectRequest) -> Result<ProjectResponse, String> {
        let url = format!("{}/api/projects/{}", self.api_base_url, project_id);
        
        let response = self.client
            .put(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .json(&request)
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("更新项目失败: HTTP {}", response.status()));
        }

        let project_response: ProjectResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if project_response.success {
            if let Some(project) = &project_response.project {
                let mut state = self.state.lock().unwrap();
                state.projects.insert(project.id.clone(), project.clone());
            }
        }

        Ok(project_response)
    }

    pub async fn delete_project(&self, auth_token: &str, project_id: &str) -> Result<ProjectResponse, String> {
        let url = format!("{}/api/projects/{}", self.api_base_url, project_id);
        
        let response = self.client
            .delete(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("删除项目失败: HTTP {}", response.status()));
        }

        let project_response: ProjectResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 从本地状态中移除
        if project_response.success {
            let mut state = self.state.lock().unwrap();
            state.projects.remove(project_id);
            state.issues.remove(project_id);
            state.members.remove(project_id);
            state.milestones.remove(project_id);
        }

        Ok(project_response)
    }

    pub async fn get_project(&self, auth_token: &str, project_id: &str) -> Result<ProjectResponse, String> {
        let url = format!("{}/api/projects/{}", self.api_base_url, project_id);
        
        let response = self.client
            .get(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("获取项目失败: HTTP {}", response.status()));
        }

        let project_response: ProjectResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if project_response.success {
            if let Some(project) = &project_response.project {
                let mut state = self.state.lock().unwrap();
                state.projects.insert(project.id.clone(), project.clone());
            }
        }

        Ok(project_response)
    }

    pub async fn list_projects(&self, auth_token: &str, limit: Option<i32>, offset: Option<i32>) -> Result<ProjectResponse, String> {
        let mut url = format!("{}/api/projects", self.api_base_url);
        
        let mut params = Vec::new();
        if let Some(limit) = limit {
            params.push(format!("limit={}", limit));
        }
        if let Some(offset) = offset {
            params.push(format!("offset={}", offset));
        }
        
        if !params.is_empty() {
            url.push('?');
            url.push_str(&params.join("&"));
        }
        
        let response = self.client
            .get(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("获取项目列表失败: HTTP {}", response.status()));
        }

        let project_response: ProjectResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if project_response.success {
            if let Some(projects) = &project_response.projects {
                let mut state = self.state.lock().unwrap();
                for project in projects {
                    state.projects.insert(project.id.clone(), project.clone());
                }
            }
        }

        Ok(project_response)
    }

    pub async fn create_issue(&self, auth_token: &str, request: CreateIssueRequest) -> Result<ProjectResponse, String> {
        let url = format!("{}/api/projects/{}/issues", self.api_base_url, request.project_id);
        
        let response = self.client
            .post(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .json(&request)
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("创建问题失败: HTTP {}", response.status()));
        }

        let project_response: ProjectResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if project_response.success {
            if let Some(issue) = &project_response.issue {
                let mut state = self.state.lock().unwrap();
                let project_issues = state.issues.entry(issue.project_id.clone()).or_insert_with(Vec::new);
                project_issues.push(issue.clone());
            }
        }

        Ok(project_response)
    }

    pub async fn update_issue(&self, auth_token: &str, issue_id: &str, request: UpdateIssueRequest) -> Result<ProjectResponse, String> {
        let url = format!("{}/api/issues/{}", self.api_base_url, issue_id);
        
        let response = self.client
            .put(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .json(&request)
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("更新问题失败: HTTP {}", response.status()));
        }

        let project_response: ProjectResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        Ok(project_response)
    }

    pub async fn list_issues(&self, auth_token: &str, project_id: &str, limit: Option<i32>, offset: Option<i32>) -> Result<ProjectResponse, String> {
        let mut url = format!("{}/api/projects/{}/issues", self.api_base_url, project_id);
        
        let mut params = Vec::new();
        if let Some(limit) = limit {
            params.push(format!("limit={}", limit));
        }
        if let Some(offset) = offset {
            params.push(format!("offset={}", offset));
        }
        
        if !params.is_empty() {
            url.push('?');
            url.push_str(&params.join("&"));
        }
        
        let response = self.client
            .get(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("获取问题列表失败: HTTP {}", response.status()));
        }

        let project_response: ProjectResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if project_response.success {
            if let Some(issues) = &project_response.issues {
                let mut state = self.state.lock().unwrap();
                state.issues.insert(project_id.to_string(), issues.clone());
            }
        }

        Ok(project_response)
    }

    pub async fn add_member(&self, auth_token: &str, request: AddMemberRequest) -> Result<ProjectResponse, String> {
        let url = format!("{}/api/projects/{}/members", self.api_base_url, request.project_id);
        
        let response = self.client
            .post(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .json(&request)
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("添加成员失败: HTTP {}", response.status()));
        }

        let project_response: ProjectResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        Ok(project_response)
    }

    pub async fn list_members(&self, auth_token: &str, project_id: &str) -> Result<ProjectResponse, String> {
        let url = format!("{}/api/projects/{}/members", self.api_base_url, project_id);
        
        let response = self.client
            .get(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("获取成员列表失败: HTTP {}", response.status()));
        }

        let project_response: ProjectResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if project_response.success {
            if let Some(members) = &project_response.members {
                let mut state = self.state.lock().unwrap();
                state.members.insert(project_id.to_string(), members.clone());
            }
        }

        Ok(project_response)
    }

    pub async fn create_milestone(&self, auth_token: &str, request: CreateMilestoneRequest) -> Result<ProjectResponse, String> {
        let url = format!("{}/api/projects/{}/milestones", self.api_base_url, request.project_id);
        
        let response = self.client
            .post(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .json(&request)
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("创建里程碑失败: HTTP {}", response.status()));
        }

        let project_response: ProjectResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        Ok(project_response)
    }

    pub async fn list_milestones(&self, auth_token: &str, project_id: &str) -> Result<ProjectResponse, String> {
        let url = format!("{}/api/projects/{}/milestones", self.api_base_url, project_id);
        
        let response = self.client
            .get(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("获取里程碑列表失败: HTTP {}", response.status()));
        }

        let project_response: ProjectResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if project_response.success {
            if let Some(milestones) = &project_response.milestones {
                let mut state = self.state.lock().unwrap();
                state.milestones.insert(project_id.to_string(), milestones.clone());
            }
        }

        Ok(project_response)
    }

    pub async fn add_comment(&self, auth_token: &str, request: AddCommentRequest) -> Result<ProjectResponse, String> {
        let url = format!("{}/api/issues/{}/comments", self.api_base_url, request.issue_id);
        
        let response = self.client
            .post(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .json(&request)
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("添加评论失败: HTTP {}", response.status()));
        }

        let project_response: ProjectResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        Ok(project_response)
    }

    pub async fn get_comments(&self, auth_token: &str, issue_id: &str) -> Result<ProjectResponse, String> {
        let url = format!("{}/api/issues/{}/comments", self.api_base_url, issue_id);
        
        let response = self.client
            .get(&url)
            .header("Authorization", format!("Bearer {}", auth_token))
            .send()
            .await
            .map_err(|e| format!("网络请求失败: {}", e))?;

        if !response.status().is_success() {
            return Err(format!("获取评论失败: HTTP {}", response.status()));
        }

        let project_response: ProjectResponse = response
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        // 更新本地状态
        if project_response.success {
            if let Some(comments) = &project_response.comments {
                let mut state = self.state.lock().unwrap();
                state.comments.insert(issue_id.to_string(), comments.clone());
            }
        }

        Ok(project_response)
    }

    pub fn get_cached_projects(&self) -> Vec<Project> {
        let state = self.state.lock().unwrap();
        state.projects.values().cloned().collect()
    }

    pub fn get_cached_issues(&self, project_id: &str) -> Vec<ProjectIssue> {
        let state = self.state.lock().unwrap();
        state.issues.get(project_id).cloned().unwrap_or_default()
    }

    pub fn get_cached_members(&self, project_id: &str) -> Vec<ProjectMember> {
        let state = self.state.lock().unwrap();
        state.members.get(project_id).cloned().unwrap_or_default()
    }

    pub fn get_cached_milestones(&self, project_id: &str) -> Vec<ProjectMilestone> {
        let state = self.state.lock().unwrap();
        state.milestones.get(project_id).cloned().unwrap_or_default()
    }

    pub fn get_cached_comments(&self, issue_id: &str) -> Vec<IssueComment> {
        let state = self.state.lock().unwrap();
        state.comments.get(issue_id).cloned().unwrap_or_default()
    }

    pub fn set_current_user_id(&self, user_id: String) {
        let mut state = self.state.lock().unwrap();
        state.current_user_id = Some(user_id);
    }

    pub fn set_api_base_url(&mut self, url: String) {
        self.api_base_url = url;
    }
}

// Tauri 命令
#[tauri::command]
pub async fn project_create(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    request: CreateProjectRequest,
) -> Result<ProjectResponse, String> {
    project_manager.create_project(&auth_token, request).await
}

#[tauri::command]
pub async fn project_update(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    project_id: String,
    request: UpdateProjectRequest,
) -> Result<ProjectResponse, String> {
    project_manager.update_project(&auth_token, &project_id, request).await
}

#[tauri::command]
pub async fn project_delete(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    project_id: String,
) -> Result<ProjectResponse, String> {
    project_manager.delete_project(&auth_token, &project_id).await
}

#[tauri::command]
pub async fn project_get(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    project_id: String,
) -> Result<ProjectResponse, String> {
    project_manager.get_project(&auth_token, &project_id).await
}

#[tauri::command]
pub async fn project_list(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    limit: Option<i32>,
    offset: Option<i32>,
) -> Result<ProjectResponse, String> {
    project_manager.list_projects(&auth_token, limit, offset).await
}

#[tauri::command]
pub async fn issue_create(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    request: CreateIssueRequest,
) -> Result<ProjectResponse, String> {
    project_manager.create_issue(&auth_token, request).await
}

#[tauri::command]
pub async fn issue_update(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    issue_id: String,
    request: UpdateIssueRequest,
) -> Result<ProjectResponse, String> {
    project_manager.update_issue(&auth_token, &issue_id, request).await
}

#[tauri::command]
pub async fn issue_list(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    project_id: String,
    limit: Option<i32>,
    offset: Option<i32>,
) -> Result<ProjectResponse, String> {
    project_manager.list_issues(&auth_token, &project_id, limit, offset).await
}

#[tauri::command]
pub async fn issue_delete(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    issue_id: String,
) -> Result<ProjectResponse, String> {
    // 暂时返回成功，实际实现需要添加到ProjectManager
    Ok(ProjectResponse {
        success: true,
        message: "Issue deleted successfully".to_string(),
        projects: None,
        project: None,
        issues: None,
        issue: None,
        members: None,
        milestones: None,
        comments: None,
        attachments: None,
    })
}

#[tauri::command]
pub async fn issue_get(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    issue_id: String,
) -> Result<ProjectResponse, String> {
    // 暂时返回成功，实际实现需要添加到ProjectManager
    Ok(ProjectResponse {
        success: true,
        message: "Issue retrieved successfully".to_string(),
        projects: None,
        project: None,
        issues: None,
        issue: None,
        members: None,
        milestones: None,
        comments: None,
        attachments: None,
    })
}

#[tauri::command]
pub async fn member_add(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    request: AddMemberRequest,
) -> Result<ProjectResponse, String> {
    project_manager.add_member(&auth_token, request).await
}

#[tauri::command]
pub async fn member_list(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    project_id: String,
) -> Result<ProjectResponse, String> {
    project_manager.list_members(&auth_token, &project_id).await
}

#[tauri::command]
pub async fn member_remove(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    project_id: String,
    user_id: String,
) -> Result<ProjectResponse, String> {
    // 暂时返回成功，实际实现需要添加到ProjectManager
    Ok(ProjectResponse {
        success: true,
        message: "Member removed successfully".to_string(),
        projects: None,
        project: None,
        issues: None,
        issue: None,
        members: None,
        milestones: None,
        comments: None,
        attachments: None,
    })
}

#[tauri::command]
pub async fn member_update_role(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    project_id: String,
    user_id: String,
    role: String,
) -> Result<ProjectResponse, String> {
    // 暂时返回成功，实际实现需要添加到ProjectManager
    Ok(ProjectResponse {
        success: true,
        message: "Member role updated successfully".to_string(),
        projects: None,
        project: None,
        issues: None,
        issue: None,
        members: None,
        milestones: None,
        comments: None,
        attachments: None,
    })
}

#[tauri::command]
pub async fn milestone_create(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    request: CreateMilestoneRequest,
) -> Result<ProjectResponse, String> {
    project_manager.create_milestone(&auth_token, request).await
}

#[tauri::command]
pub async fn milestone_list(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    project_id: String,
) -> Result<ProjectResponse, String> {
    project_manager.list_milestones(&auth_token, &project_id).await
}

#[tauri::command]
pub async fn milestone_update(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    milestone_id: String,
    title: String,
    description: Option<String>,
) -> Result<ProjectResponse, String> {
    // 暂时返回成功，实际实现需要添加到ProjectManager
    Ok(ProjectResponse {
        success: true,
        message: "Milestone updated successfully".to_string(),
        projects: None,
        project: None,
        issues: None,
        issue: None,
        members: None,
        milestones: None,
        comments: None,
        attachments: None,
    })
}

#[tauri::command]
pub async fn milestone_delete(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    milestone_id: String,
) -> Result<ProjectResponse, String> {
    // 暂时返回成功，实际实现需要添加到ProjectManager
    Ok(ProjectResponse {
        success: true,
        message: "Milestone deleted successfully".to_string(),
        projects: None,
        project: None,
        issues: None,
        issue: None,
        members: None,
        milestones: None,
        comments: None,
        attachments: None,
    })
}

#[tauri::command]
pub async fn comment_add(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    request: AddCommentRequest,
) -> Result<ProjectResponse, String> {
    project_manager.add_comment(&auth_token, request).await
}

#[tauri::command]
pub async fn comment_list(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    issue_id: String,
) -> Result<ProjectResponse, String> {
    project_manager.get_comments(&auth_token, &issue_id).await
}

#[tauri::command]
pub async fn comment_update(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    comment_id: String,
    content: String,
) -> Result<ProjectResponse, String> {
    // 暂时返回成功，实际实现需要添加到ProjectManager
    Ok(ProjectResponse {
        success: true,
        message: "Comment updated successfully".to_string(),
        projects: None,
        project: None,
        issues: None,
        issue: None,
        members: None,
        milestones: None,
        comments: None,
        attachments: None,
    })
}

#[tauri::command]
pub async fn comment_delete(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    comment_id: String,
) -> Result<ProjectResponse, String> {
    // 暂时返回成功，实际实现需要添加到ProjectManager
    Ok(ProjectResponse {
        success: true,
        message: "Comment deleted successfully".to_string(),
        projects: None,
        project: None,
        issues: None,
        issue: None,
        members: None,
        milestones: None,
        comments: None,
        attachments: None,
    })
}

#[tauri::command]
pub async fn comment_get(
    project_manager: State<'_, ProjectManager>,
    auth_token: String,
    comment_id: String,
) -> Result<ProjectResponse, String> {
    // 暂时返回成功，实际实现需要添加到ProjectManager
    Ok(ProjectResponse {
        success: true,
        message: "Comment retrieved successfully".to_string(),
        projects: None,
        project: None,
        issues: None,
        issue: None,
        members: None,
        milestones: None,
        comments: None,
        attachments: None,
    })
}

#[tauri::command]
pub fn project_get_cached_list(project_manager: State<'_, ProjectManager>) -> Vec<Project> {
    project_manager.get_cached_projects()
}

#[tauri::command]
pub fn issue_get_cached_list(
    project_manager: State<'_, ProjectManager>,
    project_id: String,
) -> Vec<ProjectIssue> {
    project_manager.get_cached_issues(&project_id)
}

#[tauri::command]
pub fn project_get_cached_issues(
    project_manager: State<'_, ProjectManager>,
    project_id: String,
) -> Vec<ProjectIssue> {
    project_manager.get_cached_issues(&project_id)
}

#[tauri::command]
pub fn project_get_cached_members(
    project_manager: State<'_, ProjectManager>,
    project_id: String,
) -> Vec<ProjectMember> {
    project_manager.get_cached_members(&project_id)
}

#[tauri::command]
pub fn project_get_cached_milestones(
    project_manager: State<'_, ProjectManager>,
    project_id: String,
) -> Vec<ProjectMilestone> {
    project_manager.get_cached_milestones(&project_id)
}

#[tauri::command]
pub fn project_get_cached_comments(
    project_manager: State<'_, ProjectManager>,
    issue_id: String,
) -> Vec<IssueComment> {
    project_manager.get_cached_comments(&issue_id)
}