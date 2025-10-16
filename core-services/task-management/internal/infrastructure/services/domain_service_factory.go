package services

import (
	"task-management/internal/domain"
)

// DomainServiceFactory 
type DomainServiceFactory struct {
	taskRepo    domain.TaskRepository
	projectRepo domain.ProjectRepository
	teamRepo    domain.TeamRepository
}

// NewDomainServiceFactory 
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

// CreateTaskAllocationService 
func (f *DomainServiceFactory) CreateTaskAllocationService() domain.TaskAllocationService {
	return NewTaskAllocationService(f.taskRepo, f.projectRepo, f.teamRepo)
}

// CreateTaskSchedulingService 
func (f *DomainServiceFactory) CreateTaskSchedulingService() domain.TaskSchedulingService {
	return NewTaskSchedulingService(f.taskRepo, f.projectRepo, f.teamRepo)
}

// CreatePerformanceAnalysisService 
func (f *DomainServiceFactory) CreatePerformanceAnalysisService() domain.PerformanceAnalysisService {
	return NewPerformanceAnalysisService(f.taskRepo, f.projectRepo, f.teamRepo)
}

// CreateNotificationService 
func (f *DomainServiceFactory) CreateNotificationService() domain.NotificationService {
	// ?
	// 
	return NewMockNotificationService()
}

