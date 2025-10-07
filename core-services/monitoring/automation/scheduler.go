package automation

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// NewScheduler 创建调度器
func NewScheduler(config SchedulerConfig) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	
	scheduler := &Scheduler{
		config:    config,
		workflows: make(chan *Workflow, config.QueueSize),
		workers:   make([]*Worker, 0, config.MaxConcurrent),
		stats: &SchedulerStats{
			QueuedWorkflows:    0,
			ProcessedWorkflows: 0,
			ActiveWorkers:      0,
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	// 创建工作器
	for i := 0; i < config.MaxConcurrent; i++ {
		worker := &Worker{
			id:        i,
			scheduler: scheduler,
		}
		scheduler.workers = append(scheduler.workers, worker)
	}
	
	return scheduler
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	// 启动工作器
	for _, worker := range s.workers {
		s.wg.Add(1)
		go worker.run()
	}
	
	// 启动调度循环
	s.wg.Add(1)
	go s.scheduleLoop()
	
	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() error {
	s.cancel()
	s.wg.Wait()
	close(s.workflows)
	return nil
}

// ScheduleWorkflow 调度工作流
func (s *Scheduler) ScheduleWorkflow(workflow *Workflow) error {
	select {
	case s.workflows <- workflow:
		s.mutex.Lock()
		s.stats.QueuedWorkflows++
		s.mutex.Unlock()
		return nil
	default:
		return fmt.Errorf("workflow queue is full")
	}
}

// GetStats 获取统计信息
func (s *Scheduler) GetStats() *SchedulerStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	stats := *s.stats
	return &stats
}

// scheduleLoop 调度循环
func (s *Scheduler) scheduleLoop() {
	defer s.wg.Done()
	
	ticker := time.NewTicker(s.config.Interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.updateActiveWorkers()
		}
	}
}

// updateActiveWorkers 更新活跃工作器数量
func (s *Scheduler) updateActiveWorkers() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	activeCount := 0
	for _, worker := range s.workers {
		if worker.isActive() {
			activeCount++
		}
	}
	s.stats.ActiveWorkers = activeCount
}

// run 工作器运行
func (w *Worker) run() {
	defer w.scheduler.wg.Done()
	
	ctx, cancel := context.WithCancel(w.scheduler.ctx)
	w.ctx = ctx
	w.cancel = cancel
	
	for {
		select {
		case <-w.ctx.Done():
			return
		case workflow := <-w.scheduler.workflows:
			if workflow != nil {
				w.processWorkflow(workflow)
			}
		}
	}
}

// processWorkflow 处理工作流
func (w *Worker) processWorkflow(workflow *Workflow) {
	workflow.mutex.Lock()
	workflow.Status = WorkflowStatusRunning
	workflow.StartedAt = time.Now()
	workflow.mutex.Unlock()
	
	// 更新统计信息
	w.scheduler.mutex.Lock()
	w.scheduler.stats.ProcessedWorkflows++
	w.scheduler.mutex.Unlock()
	
	// 执行工作流任务
	err := w.executeWorkflowTasks(workflow)
	
	// 更新工作流状态
	workflow.mutex.Lock()
	workflow.CompletedAt = time.Now()
	if err != nil {
		workflow.Status = WorkflowStatusFailed
		workflow.Error = err.Error()
	} else {
		workflow.Status = WorkflowStatusCompleted
	}
	workflow.mutex.Unlock()
}

// executeWorkflowTasks 执行工作流任务
func (w *Worker) executeWorkflowTasks(workflow *Workflow) error {
	// 按依赖关系排序任务
	sortedTasks, err := w.sortTasksByDependencies(workflow)
	if err != nil {
		return fmt.Errorf("failed to sort tasks: %w", err)
	}
	
	// 执行任务
	for _, task := range sortedTasks {
		if err := w.executeTask(task); err != nil {
			return fmt.Errorf("task %s failed: %w", task.Name, err)
		}
	}
	
	return nil
}

// sortTasksByDependencies 按依赖关系排序任务
func (w *Worker) sortTasksByDependencies(workflow *Workflow) ([]*Task, error) {
	var sortedTasks []*Task
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	
	var visit func(taskName string) error
	visit = func(taskName string) error {
		if visiting[taskName] {
			return fmt.Errorf("circular dependency detected: %s", taskName)
		}
		if visited[taskName] {
			return nil
		}
		
		visiting[taskName] = true
		
		task, exists := workflow.Tasks[taskName]
		if !exists {
			return fmt.Errorf("task not found: %s", taskName)
		}
		
		// 访问依赖
		for _, dep := range task.Definition.Dependencies {
			if err := visit(dep); err != nil {
				return err
			}
		}
		
		visiting[taskName] = false
		visited[taskName] = true
		sortedTasks = append(sortedTasks, task)
		
		return nil
	}
	
	// 访问所有任务
	for taskName := range workflow.Tasks {
		if err := visit(taskName); err != nil {
			return nil, err
		}
	}
	
	return sortedTasks, nil
}

// executeTask 执行任务
func (w *Worker) executeTask(task *Task) error {
	task.mutex.Lock()
	task.Status = TaskStatusRunning
	task.StartedAt = time.Now()
	task.mutex.Unlock()
	
	// 检查条件
	if !w.checkTaskConditions(task) {
		task.mutex.Lock()
		task.Status = TaskStatusSkipped
		task.CompletedAt = time.Now()
		task.mutex.Unlock()
		return nil
	}
	
	// 执行任务（这里需要与执行器集成）
	err := w.executeTaskWithExecutor(task)
	
	// 更新任务状态
	task.mutex.Lock()
	task.CompletedAt = time.Now()
	if err != nil {
		task.Status = TaskStatusFailed
		task.Error = err.Error()
	} else {
		task.Status = TaskStatusCompleted
	}
	task.mutex.Unlock()
	
	return err
}

// checkTaskConditions 检查任务条件
func (w *Worker) checkTaskConditions(task *Task) bool {
	for _, condition := range task.Definition.Conditions {
		if !w.evaluateCondition(condition) {
			return false
		}
	}
	return true
}

// evaluateCondition 评估条件
func (w *Worker) evaluateCondition(condition ConditionDefinition) bool {
	// 简单的条件评估实现
	switch condition.Type {
	case "expression":
		// 这里可以集成表达式引擎
		return true
	case "script":
		// 这里可以执行脚本
		return true
	case "webhook":
		// 这里可以调用webhook
		return true
	default:
		return true
	}
}

// executeTaskWithExecutor 使用执行器执行任务
func (w *Worker) executeTaskWithExecutor(task *Task) error {
	// 创建执行器
	executorConfig := ExecutorConfig{
		Type:     task.Type,
		Enabled:  true,
		Settings: make(map[string]interface{}),
	}
	
	executor, err := CreateExecutor(executorConfig)
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}
	
	// 启动执行器
	if err := executor.Start(); err != nil {
		return fmt.Errorf("failed to start executor: %w", err)
	}
	defer executor.Stop()
	
	// 执行任务
	return executor.ExecuteTask(task, task.Inputs)
}

// isActive 检查工作器是否活跃
func (w *Worker) isActive() bool {
	return w.ctx != nil && w.ctx.Err() == nil
}

// AutoScaler 自动扩缩容器实现
func NewAutoScaler(config AutoScalerConfig) *AutoScaler {
	return &AutoScaler{
		config: config,
		stats: &AutoScalerStats{
			CurrentReplicas: config.MinReplicas,
			DesiredReplicas: config.MinReplicas,
			LastScaleUp:     time.Time{},
			LastScaleDown:   time.Time{},
			ScaleUpCount:    0,
			ScaleDownCount:  0,
		},
	}
}

// Scale 执行扩缩容
func (a *AutoScaler) Scale(currentMetrics map[string]float64) error {
	if !a.config.Enabled {
		return nil
	}
	
	a.mutex.Lock()
	defer a.mutex.Unlock()
	
	// 计算目标副本数
	desiredReplicas := a.calculateDesiredReplicas(currentMetrics)
	
	// 检查是否需要扩缩容
	if desiredReplicas == a.stats.CurrentReplicas {
		return nil
	}
	
	// 检查冷却时间
	now := time.Now()
	if desiredReplicas > a.stats.CurrentReplicas {
		// 扩容
		if now.Sub(a.stats.LastScaleUp) < a.config.ScaleUpCooldown {
			return nil
		}
		a.stats.LastScaleUp = now
		a.stats.ScaleUpCount++
	} else {
		// 缩容
		if now.Sub(a.stats.LastScaleDown) < a.config.ScaleDownCooldown {
			return nil
		}
		a.stats.LastScaleDown = now
		a.stats.ScaleDownCount++
	}
	
	// 更新副本数
	a.stats.DesiredReplicas = desiredReplicas
	a.stats.CurrentReplicas = desiredReplicas
	
	return nil
}

// calculateDesiredReplicas 计算目标副本数
func (a *AutoScaler) calculateDesiredReplicas(metrics map[string]float64) int {
	cpuUsage := metrics["cpu"]
	memoryUsage := metrics["memory"]
	
	// 基于CPU使用率计算
	cpuReplicas := int(float64(a.stats.CurrentReplicas) * cpuUsage / a.config.TargetCPU)
	
	// 基于内存使用率计算
	memoryReplicas := int(float64(a.stats.CurrentReplicas) * memoryUsage / a.config.TargetMemory)
	
	// 取最大值
	desiredReplicas := cpuReplicas
	if memoryReplicas > desiredReplicas {
		desiredReplicas = memoryReplicas
	}
	
	// 限制在最小和最大副本数之间
	if desiredReplicas < a.config.MinReplicas {
		desiredReplicas = a.config.MinReplicas
	}
	if desiredReplicas > a.config.MaxReplicas {
		desiredReplicas = a.config.MaxReplicas
	}
	
	return desiredReplicas
}

// GetStats 获取统计信息
func (a *AutoScaler) GetStats() *AutoScalerStats {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	
	stats := *a.stats
	return &stats
}