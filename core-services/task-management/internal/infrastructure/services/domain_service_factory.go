package services

import (
	"task-management/internal/domain"
)

// DomainServiceFactory 领域服务工厂实现
type DomainServiceFactory struct {
	taskRepo    domain.TaskRepository
	projectRepo domain.ProjectRepository
	teamRepo    domain.TeamRepository
}

// NewDomainServiceFactory 创建领域服务工厂
func NewDomainServiceFactory(
	taskRepo domain.TaskRepository,
	projectRepo domain.ProjectRepository,
	teamRepo domain.TeamRepository,
) domain.DomainServiceFactory {
	return &DomainServiceFactory{
		taskRepo:    taskRepo,
		projectRepo: projectRepo,
		teamRepo:    teamRepo,
	}
}

// CreateTaskAllocationService 创建任务分配服务
func (f *DomainServiceFactory) CreateTaskAllocationService() domain.TaskAllocationService {
	return NewTaskAllocationService(f.taskRepo, f.projectRepo, f.teamRepo)
}

// CreateTaskSchedulingService 创建任务调度服务
func (f *DomainServiceFactory) CreateTaskSchedulingService() domain.TaskSchedulingService {
	return NewTaskSchedulingService(f.taskRepo, f.projectRepo, f.teamRepo)
}

// CreatePerformanceAnalysisService 创建性能分析服务
func (f *DomainServiceFactory) CreatePerformanceAnalysisService() domain.PerformanceAnalysisService {
	return NewPerformanceAnalysisService(f.taskRepo, f.projectRepo, f.teamRepo)
}

// CreateNotificationService 创建通知服务
func (f *DomainServiceFactory) CreateNotificationService() domain.NotificationService {
	// 在实际应用中，这里应该注入真实的邮件、短信、推送服务
	// 这里使用模拟服务进行演示
	return NewMockNotificationService()
}