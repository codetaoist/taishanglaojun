package automation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/monitoring/interfaces"
)

// Orchestrator 自动化编排器
type Orchestrator struct {
	config     OrchestratorConfig
	workflows  map[string]*Workflow
	tasks      map[string]*Task
	triggers   map[string]*Trigger
	executors  map[string]Executor
	scheduler  *Scheduler
	stats      *OrchestratorStats
	mutex      sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// OrchestratorConfig 编排器配?
type OrchestratorConfig struct {
	// 基础配置
	MaxConcurrentWorkflows int           `json:"max_concurrent_workflows" yaml:"max_concurrent_workflows"`
	MaxConcurrentTasks     int           `json:"max_concurrent_tasks" yaml:"max_concurrent_tasks"`
	TaskTimeout            time.Duration `json:"task_timeout" yaml:"task_timeout"`
	WorkflowTimeout        time.Duration `json:"workflow_timeout" yaml:"workflow_timeout"`
	
	// 调度配置
	SchedulerInterval      time.Duration `json:"scheduler_interval" yaml:"scheduler_interval"`
	RetryAttempts          int           `json:"retry_attempts" yaml:"retry_attempts"`
	RetryDelay             time.Duration `json:"retry_delay" yaml:"retry_delay"`
	
	// 存储配置
	StateStorage           string        `json:"state_storage" yaml:"state_storage"`
	HistoryRetention       time.Duration `json:"history_retention" yaml:"history_retention"`
	
	// 通知配置
	NotificationEnabled    bool          `json:"notification_enabled" yaml:"notification_enabled"`
	NotificationChannels   []string      `json:"notification_channels" yaml:"notification_channels"`
	
	// 安全配置
	EnableRBAC             bool          `json:"enable_rbac" yaml:"enable_rbac"`
	AllowedUsers           []string      `json:"allowed_users" yaml:"allowed_users"`
	AllowedRoles           []string      `json:"allowed_roles" yaml:"allowed_roles"`
	
	// 执行器配?
	Executors              map[string]ExecutorConfig `json:"executors" yaml:"executors"`
}

// OrchestratorStats 编排器统计信?
type OrchestratorStats struct {
	// 工作流统?
	TotalWorkflows     int64 `json:"total_workflows"`
	RunningWorkflows   int64 `json:"running_workflows"`
	CompletedWorkflows int64 `json:"completed_workflows"`
	FailedWorkflows    int64 `json:"failed_workflows"`
	
	// 任务统计
	TotalTasks         int64 `json:"total_tasks"`
	RunningTasks       int64 `json:"running_tasks"`
	CompletedTasks     int64 `json:"completed_tasks"`
	FailedTasks        int64 `json:"failed_tasks"`
	
	// 性能统计
	AverageWorkflowDuration time.Duration `json:"average_workflow_duration"`
	AverageTaskDuration     time.Duration `json:"average_task_duration"`
	
	// 资源统计
	CPUUsage           float64 `json:"cpu_usage"`
	MemoryUsage        float64 `json:"memory_usage"`
	
	// 时间统计
	LastExecution      time.Time `json:"last_execution"`
	Uptime             time.Duration `json:"uptime"`
	StartTime          time.Time `json:"start_time"`
}

// NewOrchestrator 创建编排?
func NewOrchestrator(config OrchestratorConfig) (*Orchestrator, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// 设置默认?
	if config.MaxConcurrentWorkflows == 0 {
		config.MaxConcurrentWorkflows = 10
	}
	if config.MaxConcurrentTasks == 0 {
		config.MaxConcurrentTasks = 50
	}
	if config.TaskTimeout == 0 {
		config.TaskTimeout = 30 * time.Minute
	}
	if config.WorkflowTimeout == 0 {
		config.WorkflowTimeout = 2 * time.Hour
	}
	if config.SchedulerInterval == 0 {
		config.SchedulerInterval = 10 * time.Second
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 30 * time.Second
	}
	if config.HistoryRetention == 0 {
		config.HistoryRetention = 30 * 24 * time.Hour // 30?
	}
	
	orchestrator := &Orchestrator{
		config:    config,
		workflows: make(map[string]*Workflow),
		tasks:     make(map[string]*Task),
		triggers:  make(map[string]*Trigger),
		executors: make(map[string]Executor),
		stats: &OrchestratorStats{
			StartTime: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	// 初始化调度器
	scheduler, err := NewScheduler(SchedulerConfig{
		Interval:      config.SchedulerInterval,
		MaxConcurrent: config.MaxConcurrentWorkflows,
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}
	orchestrator.scheduler = scheduler
	
	// 初始化执行器
	if err := orchestrator.initExecutors(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize executors: %w", err)
	}
	
	return orchestrator, nil
}

// Start 启动编排?
func (o *Orchestrator) Start() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	
	// 启动调度?
	if err := o.scheduler.Start(); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}
	
	// 启动执行?
	for name, executor := range o.executors {
		if err := executor.Start(); err != nil {
			return fmt.Errorf("failed to start executor %s: %w", name, err)
		}
	}
	
	// 启动监控循环
	o.wg.Add(1)
	go o.monitorLoop()
	
	// 启动清理循环
	o.wg.Add(1)
	go o.cleanupLoop()
	
	// 启动统计更新循环
	o.wg.Add(1)
	go o.statsUpdateLoop()
	
	return nil
}

// Stop 停止编排?
func (o *Orchestrator) Stop() error {
	o.cancel()
	o.wg.Wait()
	
	o.mutex.Lock()
	defer o.mutex.Unlock()
	
	// 停止调度?
	if err := o.scheduler.Stop(); err != nil {
		return fmt.Errorf("failed to stop scheduler: %w", err)
	}
	
	// 停止执行?
	for _, executor := range o.executors {
		executor.Stop()
	}
	
	return nil
}

// CreateWorkflow 创建工作?
func (o *Orchestrator) CreateWorkflow(definition WorkflowDefinition) (*Workflow, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	
	// 验证工作流定?
	if err := o.validateWorkflowDefinition(definition); err != nil {
		return nil, fmt.Errorf("invalid workflow definition: %w", err)
	}
	
	workflow := &Workflow{
		ID:          generateID(),
		Name:        definition.Name,
		Description: definition.Description,
		Definition:  definition,
		Status:      WorkflowStatusPending,
		CreatedAt:   time.Now(),
		Tasks:       make(map[string]*Task),
		Variables:   make(map[string]interface{}),
	}
	
	// 创建任务
	for _, taskDef := range definition.Tasks {
		task, err := o.createTask(taskDef, workflow.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to create task %s: %w", taskDef.Name, err)
		}
		workflow.Tasks[task.ID] = task
		o.tasks[task.ID] = task
	}
	
	o.workflows[workflow.ID] = workflow
	o.stats.TotalWorkflows++
	
	return workflow, nil
}

// ExecuteWorkflow 执行工作?
func (o *Orchestrator) ExecuteWorkflow(workflowID string, inputs map[string]interface{}) error {
	o.mutex.RLock()
	workflow, exists := o.workflows[workflowID]
	o.mutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}
	
	// 设置输入变量
	workflow.mutex.Lock()
	for key, value := range inputs {
		workflow.Variables[key] = value
	}
	workflow.Status = WorkflowStatusRunning
	workflow.StartedAt = time.Now()
	workflow.mutex.Unlock()
	
	// 提交到调度器
	return o.scheduler.ScheduleWorkflow(workflow)
}

// CreateTask 创建任务
func (o *Orchestrator) CreateTask(definition TaskDefinition) (*Task, error) {
	return o.createTask(definition, "")
}

// ExecuteTask 执行任务
func (o *Orchestrator) ExecuteTask(taskID string, inputs map[string]interface{}) error {
	o.mutex.RLock()
	task, exists := o.tasks[taskID]
	o.mutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}
	
	// 获取执行?
	executor, exists := o.executors[task.Definition.Type]
	if !exists {
		return fmt.Errorf("executor not found for task type: %s", task.Definition.Type)
	}
	
	// 执行任务
	return executor.ExecuteTask(task, inputs)
}

// CreateTrigger 创建触发?
func (o *Orchestrator) CreateTrigger(definition TriggerDefinition) (*Trigger, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	
	trigger := &Trigger{
		ID:         generateID(),
		Name:       definition.Name,
		Type:       definition.Type,
		Definition: definition,
		Status:     TriggerStatusActive,
		CreatedAt:  time.Now(),
	}
	
	o.triggers[trigger.ID] = trigger
	
	return trigger, nil
}

// GetWorkflow 获取工作?
func (o *Orchestrator) GetWorkflow(workflowID string) (*Workflow, error) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	
	workflow, exists := o.workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}
	
	return workflow, nil
}

// ListWorkflows 列出工作?
func (o *Orchestrator) ListWorkflows(filter WorkflowFilter) ([]*Workflow, error) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	
	workflows := make([]*Workflow, 0)
	
	for _, workflow := range o.workflows {
		if o.matchesWorkflowFilter(workflow, filter) {
			workflows = append(workflows, workflow)
		}
	}
	
	return workflows, nil
}

// GetTask 获取任务
func (o *Orchestrator) GetTask(taskID string) (*Task, error) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	
	task, exists := o.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	
	return task, nil
}

// ListTasks 列出任务
func (o *Orchestrator) ListTasks(filter TaskFilter) ([]*Task, error) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	
	tasks := make([]*Task, 0)
	
	for _, task := range o.tasks {
		if o.matchesTaskFilter(task, filter) {
			tasks = append(tasks, task)
		}
	}
	
	return tasks, nil
}

// GetStats 获取统计信息
func (o *Orchestrator) GetStats() *OrchestratorStats {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	
	stats := *o.stats
	stats.Uptime = time.Since(o.stats.StartTime)
	
	return &stats
}

// HealthCheck 健康检?
func (o *Orchestrator) HealthCheck() error {
	// 检查调度器
	if err := o.scheduler.HealthCheck(); err != nil {
		return fmt.Errorf("scheduler health check failed: %w", err)
	}
	
	// 检查执行器
	for name, executor := range o.executors {
		if err := executor.HealthCheck(); err != nil {
			return fmt.Errorf("executor %s health check failed: %w", name, err)
		}
	}
	
	return nil
}

// initExecutors 初始化执行器
func (o *Orchestrator) initExecutors() error {
	for name, config := range o.config.Executors {
		executor, err := CreateExecutor(config)
		if err != nil {
			return fmt.Errorf("failed to create executor %s: %w", name, err)
		}
		o.executors[name] = executor
	}
	
	return nil
}

// createTask 创建任务
func (o *Orchestrator) createTask(definition TaskDefinition, workflowID string) (*Task, error) {
	task := &Task{
		ID:         generateID(),
		Name:       definition.Name,
		Type:       definition.Type,
		Definition: definition,
		WorkflowID: workflowID,
		Status:     TaskStatusPending,
		CreatedAt:  time.Now(),
		Inputs:     make(map[string]interface{}),
		Outputs:    make(map[string]interface{}),
	}
	
	return task, nil
}

// validateWorkflowDefinition 验证工作流定?
func (o *Orchestrator) validateWorkflowDefinition(definition WorkflowDefinition) error {
	if definition.Name == "" {
		return fmt.Errorf("workflow name is required")
	}
	
	if len(definition.Tasks) == 0 {
		return fmt.Errorf("workflow must have at least one task")
	}
	
	// 验证任务依赖
	taskNames := make(map[string]bool)
	for _, task := range definition.Tasks {
		taskNames[task.Name] = true
	}
	
	for _, task := range definition.Tasks {
		for _, dep := range task.Dependencies {
			if !taskNames[dep] {
				return fmt.Errorf("task %s depends on non-existent task %s", task.Name, dep)
			}
		}
	}
	
	return nil
}

// matchesWorkflowFilter 检查工作流是否匹配过滤?
func (o *Orchestrator) matchesWorkflowFilter(workflow *Workflow, filter WorkflowFilter) bool {
	if filter.Status != "" && workflow.Status != filter.Status {
		return false
	}
	
	if filter.Name != "" && workflow.Name != filter.Name {
		return false
	}
	
	if !filter.CreatedAfter.IsZero() && workflow.CreatedAt.Before(filter.CreatedAfter) {
		return false
	}
	
	if !filter.CreatedBefore.IsZero() && workflow.CreatedAt.After(filter.CreatedBefore) {
		return false
	}
	
	return true
}

// matchesTaskFilter 检查任务是否匹配过滤器
func (o *Orchestrator) matchesTaskFilter(task *Task, filter TaskFilter) bool {
	if filter.Status != "" && task.Status != filter.Status {
		return false
	}
	
	if filter.Type != "" && task.Type != filter.Type {
		return false
	}
	
	if filter.WorkflowID != "" && task.WorkflowID != filter.WorkflowID {
		return false
	}
	
	return true
}

// monitorLoop 监控循环
func (o *Orchestrator) monitorLoop() {
	defer o.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-o.ctx.Done():
			return
		case <-ticker.C:
			o.monitorWorkflows()
			o.monitorTasks()
		}
	}
}

// cleanupLoop 清理循环
func (o *Orchestrator) cleanupLoop() {
	defer o.wg.Done()
	
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-o.ctx.Done():
			return
		case <-ticker.C:
			o.cleanup()
		}
	}
}

// statsUpdateLoop 统计更新循环
func (o *Orchestrator) statsUpdateLoop() {
	defer o.wg.Done()
	
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-o.ctx.Done():
			return
		case <-ticker.C:
			o.updateStats()
		}
	}
}

// monitorWorkflows 监控工作?
func (o *Orchestrator) monitorWorkflows() {
	o.mutex.RLock()
	workflows := make([]*Workflow, 0)
	for _, workflow := range o.workflows {
		if workflow.Status == WorkflowStatusRunning {
			workflows = append(workflows, workflow)
		}
	}
	o.mutex.RUnlock()
	
	for _, workflow := range workflows {
		// 检查超?
		if time.Since(workflow.StartedAt) > o.config.WorkflowTimeout {
			workflow.mutex.Lock()
			workflow.Status = WorkflowStatusFailed
			workflow.Error = "workflow timeout"
			workflow.CompletedAt = time.Now()
			workflow.mutex.Unlock()
		}
	}
}

// monitorTasks 监控任务
func (o *Orchestrator) monitorTasks() {
	o.mutex.RLock()
	tasks := make([]*Task, 0)
	for _, task := range o.tasks {
		if task.Status == TaskStatusRunning {
			tasks = append(tasks, task)
		}
	}
	o.mutex.RUnlock()
	
	for _, task := range tasks {
		// 检查超?
		if time.Since(task.StartedAt) > o.config.TaskTimeout {
			task.mutex.Lock()
			task.Status = TaskStatusFailed
			task.Error = "task timeout"
			task.CompletedAt = time.Now()
			task.mutex.Unlock()
		}
	}
}

// cleanup 清理过期数据
func (o *Orchestrator) cleanup() {
	cutoff := time.Now().Add(-o.config.HistoryRetention)
	
	o.mutex.Lock()
	defer o.mutex.Unlock()
	
	// 清理过期工作?
	for id, workflow := range o.workflows {
		if workflow.CompletedAt.Before(cutoff) && 
		   (workflow.Status == WorkflowStatusCompleted || workflow.Status == WorkflowStatusFailed) {
			delete(o.workflows, id)
		}
	}
	
	// 清理过期任务
	for id, task := range o.tasks {
		if task.CompletedAt.Before(cutoff) && 
		   (task.Status == TaskStatusCompleted || task.Status == TaskStatusFailed) {
			delete(o.tasks, id)
		}
	}
}

// updateStats 更新统计信息
func (o *Orchestrator) updateStats() {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	
	var runningWorkflows, completedWorkflows, failedWorkflows int64
	var runningTasks, completedTasks, failedTasks int64
	
	for _, workflow := range o.workflows {
		switch workflow.Status {
		case WorkflowStatusRunning:
			runningWorkflows++
		case WorkflowStatusCompleted:
			completedWorkflows++
		case WorkflowStatusFailed:
			failedWorkflows++
		}
	}
	
	for _, task := range o.tasks {
		switch task.Status {
		case TaskStatusRunning:
			runningTasks++
		case TaskStatusCompleted:
			completedTasks++
		case TaskStatusFailed:
			failedTasks++
		}
	}
	
	o.stats.RunningWorkflows = runningWorkflows
	o.stats.CompletedWorkflows = completedWorkflows
	o.stats.FailedWorkflows = failedWorkflows
	o.stats.RunningTasks = runningTasks
	o.stats.CompletedTasks = completedTasks
	o.stats.FailedTasks = failedTasks
	o.stats.LastExecution = time.Now()
}

// generateID 生成ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

