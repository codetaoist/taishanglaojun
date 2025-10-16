package automation

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// NewScheduler ?
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
	
	// ?
	for i := 0; i < config.MaxConcurrent; i++ {
		worker := &Worker{
			id:        i,
			scheduler: scheduler,
		}
		scheduler.workers = append(scheduler.workers, worker)
	}
	
	return scheduler
}

// Start ?
func (s *Scheduler) Start() error {
	// ?
	for _, worker := range s.workers {
		s.wg.Add(1)
		go worker.run()
	}
	
	// 
	s.wg.Add(1)
	go s.scheduleLoop()
	
	return nil
}

// Stop ?
func (s *Scheduler) Stop() error {
	s.cancel()
	s.wg.Wait()
	close(s.workflows)
	return nil
}

// ScheduleWorkflow ?
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

// GetStats 
func (s *Scheduler) GetStats() *SchedulerStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	stats := *s.stats
	return &stats
}

// scheduleLoop 
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

// updateActiveWorkers ?
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

// run ?
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

// processWorkflow ?
func (w *Worker) processWorkflow(workflow *Workflow) {
	workflow.mutex.Lock()
	workflow.Status = WorkflowStatusRunning
	workflow.StartedAt = time.Now()
	workflow.mutex.Unlock()
	
	// 
	w.scheduler.mutex.Lock()
	w.scheduler.stats.ProcessedWorkflows++
	w.scheduler.mutex.Unlock()
	
	// ?
	err := w.executeWorkflowTasks(workflow)
	
	// ?
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

// executeWorkflowTasks ?
func (w *Worker) executeWorkflowTasks(workflow *Workflow) error {
	// ?
	sortedTasks, err := w.sortTasksByDependencies(workflow)
	if err != nil {
		return fmt.Errorf("failed to sort tasks: %w", err)
	}
	
	// 
	for _, task := range sortedTasks {
		if err := w.executeTask(task); err != nil {
			return fmt.Errorf("task %s failed: %w", task.Name, err)
		}
	}
	
	return nil
}

// sortTasksByDependencies ?
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
		
		// 
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
	
	// ?
	for taskName := range workflow.Tasks {
		if err := visit(taskName); err != nil {
			return nil, err
		}
	}
	
	return sortedTasks, nil
}

// executeTask 
func (w *Worker) executeTask(task *Task) error {
	task.mutex.Lock()
	task.Status = TaskStatusRunning
	task.StartedAt = time.Now()
	task.mutex.Unlock()
	
	// ?
	if !w.checkTaskConditions(task) {
		task.mutex.Lock()
		task.Status = TaskStatusSkipped
		task.CompletedAt = time.Now()
		task.mutex.Unlock()
		return nil
	}
	
	// 
	err := w.executeTaskWithExecutor(task)
	
	// ?
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

// checkTaskConditions ?
func (w *Worker) checkTaskConditions(task *Task) bool {
	for _, condition := range task.Definition.Conditions {
		if !w.evaluateCondition(condition) {
			return false
		}
	}
	return true
}

// evaluateCondition 
func (w *Worker) evaluateCondition(condition ConditionDefinition) bool {
	// 
	switch condition.Type {
	case "expression":
		// ?
		return true
	case "script":
		// 
		return true
	case "webhook":
		// webhook
		return true
	default:
		return true
	}
}

// executeTaskWithExecutor ?
func (w *Worker) executeTaskWithExecutor(task *Task) error {
	// ?
	executorConfig := ExecutorConfig{
		Type:     task.Type,
		Enabled:  true,
		Settings: make(map[string]interface{}),
	}
	
	executor, err := CreateExecutor(executorConfig)
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}
	
	// ?
	if err := executor.Start(); err != nil {
		return fmt.Errorf("failed to start executor: %w", err)
	}
	defer executor.Stop()
	
	// 
	return executor.ExecuteTask(task, task.Inputs)
}

// isActive 鹤
func (w *Worker) isActive() bool {
	return w.ctx != nil && w.ctx.Err() == nil
}

// AutoScaler 
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

// Scale ?
func (a *AutoScaler) Scale(currentMetrics map[string]float64) error {
	if !a.config.Enabled {
		return nil
	}
	
	a.mutex.Lock()
	defer a.mutex.Unlock()
	
	// ?
	desiredReplicas := a.calculateDesiredReplicas(currentMetrics)
	
	// 
	if desiredReplicas == a.stats.CurrentReplicas {
		return nil
	}
	
	// ?
	now := time.Now()
	if desiredReplicas > a.stats.CurrentReplicas {
		// 
		if now.Sub(a.stats.LastScaleUp) < a.config.ScaleUpCooldown {
			return nil
		}
		a.stats.LastScaleUp = now
		a.stats.ScaleUpCount++
	} else {
		// 
		if now.Sub(a.stats.LastScaleDown) < a.config.ScaleDownCooldown {
			return nil
		}
		a.stats.LastScaleDown = now
		a.stats.ScaleDownCount++
	}
	
	// ?
	a.stats.DesiredReplicas = desiredReplicas
	a.stats.CurrentReplicas = desiredReplicas
	
	return nil
}

// calculateDesiredReplicas ?
func (a *AutoScaler) calculateDesiredReplicas(metrics map[string]float64) int {
	cpuUsage := metrics["cpu"]
	memoryUsage := metrics["memory"]
	
	// CPU?
	cpuReplicas := int(float64(a.stats.CurrentReplicas) * cpuUsage / a.config.TargetCPU)
	
	// ?
	memoryReplicas := int(float64(a.stats.CurrentReplicas) * memoryUsage / a.config.TargetMemory)
	
	// ?
	desiredReplicas := cpuReplicas
	if memoryReplicas > desiredReplicas {
		desiredReplicas = memoryReplicas
	}
	
	// 
	if desiredReplicas < a.config.MinReplicas {
		desiredReplicas = a.config.MinReplicas
	}
	if desiredReplicas > a.config.MaxReplicas {
		desiredReplicas = a.config.MaxReplicas
	}
	
	return desiredReplicas
}

// GetStats 
func (a *AutoScaler) GetStats() *AutoScalerStats {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	
	stats := *a.stats
	return &stats
}

