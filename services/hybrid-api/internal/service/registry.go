package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ServiceInfo represents information about a service
type ServiceInfo struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Version  string            `json:"version"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Protocol string            `json:"protocol"` // "http" or "grpc"
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`
	Health   HealthStatus      `json:"health"`
}

// HealthStatus represents the health status of a service
type HealthStatus struct {
	Status    string    `json:"status"` // "healthy", "unhealthy", "unknown"
	LastCheck time.Time `json:"last_check"`
	Message   string    `json:"message"`
}

// ServiceRegistry interface for service discovery and registration
type ServiceRegistry interface {
	Register(ctx context.Context, service *ServiceInfo) error
	Unregister(ctx context.Context, serviceID string) error
	GetService(ctx context.Context, serviceName string) (*ServiceInfo, error)
	GetServices(ctx context.Context, serviceName string) ([]*ServiceInfo, error)
	ListServices(ctx context.Context) ([]*ServiceInfo, error)
	Watch(ctx context.Context, serviceName string) (<-chan []*ServiceInfo, error)
	HealthCheck(ctx context.Context, serviceID string) (*HealthStatus, error)
}

// ServiceConnector interface for connecting to services
type ServiceConnector interface {
	ConnectHTTP(ctx context.Context, service *ServiceInfo) (*http.Client, error)
	ConnectGRPC(ctx context.Context, service *ServiceInfo) (*grpc.ClientConn, error)
}

// InMemoryServiceRegistry is an in-memory implementation of ServiceRegistry
type InMemoryServiceRegistry struct {
	mu       sync.RWMutex
	services map[string][]*ServiceInfo // key: service name, value: list of services
}

// NewInMemoryServiceRegistry creates a new in-memory service registry
func NewInMemoryServiceRegistry() *InMemoryServiceRegistry {
	return &InMemoryServiceRegistry{
		services: make(map[string][]*ServiceInfo),
	}
}

// Register registers a service
func (r *InMemoryServiceRegistry) Register(ctx context.Context, service *ServiceInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.services[service.Name]; !exists {
		r.services[service.Name] = make([]*ServiceInfo, 0)
	}

	// Check if service with same ID already exists
	for i, s := range r.services[service.Name] {
		if s.ID == service.ID {
			// Update existing service
			r.services[service.Name][i] = service
			return nil
		}
	}

	// Add new service
	r.services[service.Name] = append(r.services[service.Name], service)
	return nil
}

// Unregister unregisters a service
func (r *InMemoryServiceRegistry) Unregister(ctx context.Context, serviceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, services := range r.services {
		for i, service := range services {
			if service.ID == serviceID {
				// Remove service from slice
				r.services[name] = append(services[:i], services[i+1:]...)
				
				// Remove empty slice
				if len(r.services[name]) == 0 {
					delete(r.services, name)
				}
				
				return nil
			}
		}
	}

	return fmt.Errorf("service with ID %s not found", serviceID)
}

// GetService gets a service by name (returns the first healthy service)
func (r *InMemoryServiceRegistry) GetService(ctx context.Context, serviceName string) (*ServiceInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services, exists := r.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	// Return the first healthy service
	for _, service := range services {
		if service.Health.Status == "healthy" {
			return service, nil
		}
	}

	// If no healthy service, return the first one
	if len(services) > 0 {
		return services[0], nil
	}

	return nil, fmt.Errorf("no instances of service %s available", serviceName)
}

// GetServices gets all services by name
func (r *InMemoryServiceRegistry) GetServices(ctx context.Context, serviceName string) ([]*ServiceInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services, exists := r.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	// Return a copy of the services slice
	result := make([]*ServiceInfo, len(services))
	copy(result, services)
	return result, nil
}

// ListServices lists all registered services
func (r *InMemoryServiceRegistry) ListServices(ctx context.Context) ([]*ServiceInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*ServiceInfo
	for _, services := range r.services {
		result = append(result, services...)
	}

	return result, nil
}

// Watch watches for changes in services (not implemented for in-memory registry)
func (r *InMemoryServiceRegistry) Watch(ctx context.Context, serviceName string) (<-chan []*ServiceInfo, error) {
	// Not implemented for in-memory registry
	return nil, fmt.Errorf("watching not supported by in-memory registry")
}

// HealthCheck performs a health check on a service
func (r *InMemoryServiceRegistry) HealthCheck(ctx context.Context, serviceID string) (*HealthStatus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, services := range r.services {
		for _, service := range services {
			if service.ID == serviceID {
				return &service.Health, nil
			}
		}
	}

	return nil, fmt.Errorf("service with ID %s not found", serviceID)
}

// DefaultServiceConnector is a default implementation of ServiceConnector
type DefaultServiceConnector struct{}

// NewDefaultServiceConnector creates a new default service connector
func NewDefaultServiceConnector() *DefaultServiceConnector {
	return &DefaultServiceConnector{}
}

// ConnectHTTP connects to an HTTP service
func (c *DefaultServiceConnector) ConnectHTTP(ctx context.Context, service *ServiceInfo) (*http.Client, error) {
	if service.Protocol != "http" {
		return nil, fmt.Errorf("service %s is not an HTTP service", service.ID)
	}

	// Create a simple HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return client, nil
}

// ConnectGRPC connects to a gRPC service
func (c *DefaultServiceConnector) ConnectGRPC(ctx context.Context, service *ServiceInfo) (*grpc.ClientConn, error) {
	if service.Protocol != "grpc" {
		return nil, fmt.Errorf("service %s is not a gRPC service", service.ID)
	}

	// Create gRPC connection
	address := fmt.Sprintf("%s:%d", service.Address, service.Port)
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC service %s: %v", service.ID, err)
	}

	return conn, nil
}

// ServiceManager manages service discovery and connections
type ServiceManager struct {
	registry  ServiceRegistry
	connector ServiceConnector
}

// NewServiceManager creates a new service manager
func NewServiceManager(registry ServiceRegistry, connector ServiceConnector) *ServiceManager {
	return &ServiceManager{
		registry:  registry,
		connector: connector,
	}
}

// RegisterService registers a service
func (m *ServiceManager) RegisterService(ctx context.Context, service *ServiceInfo) error {
	return m.registry.Register(ctx, service)
}

// UnregisterService unregisters a service
func (m *ServiceManager) UnregisterService(ctx context.Context, serviceID string) error {
	return m.registry.Unregister(ctx, serviceID)
}

// GetService gets a service by name
func (m *ServiceManager) GetService(ctx context.Context, serviceName string) (*ServiceInfo, error) {
	return m.registry.GetService(ctx, serviceName)
}

// GetServices gets all services by name
func (m *ServiceManager) GetServices(ctx context.Context, serviceName string) ([]*ServiceInfo, error) {
	return m.registry.GetServices(ctx, serviceName)
}

// ListServices lists all registered services
func (m *ServiceManager) ListServices(ctx context.Context) ([]*ServiceInfo, error) {
	return m.registry.ListServices(ctx)
}

// ConnectHTTP connects to an HTTP service
func (m *ServiceManager) ConnectHTTP(ctx context.Context, serviceName string) (*http.Client, error) {
	service, err := m.registry.GetService(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	return m.connector.ConnectHTTP(ctx, service)
}

// ConnectGRPC connects to a gRPC service
func (m *ServiceManager) ConnectGRPC(ctx context.Context, serviceName string) (*grpc.ClientConn, error) {
	service, err := m.registry.GetService(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	return m.connector.ConnectGRPC(ctx, service)
}

// HealthCheck performs a health check on a service
func (m *ServiceManager) HealthCheck(ctx context.Context, serviceID string) (*HealthStatus, error) {
	return m.registry.HealthCheck(ctx, serviceID)
}

// ServiceHealthChecker periodically checks the health of services
type ServiceHealthChecker struct {
	registry ServiceRegistry
	interval time.Duration
	stopCh   chan struct{}
}

// NewServiceHealthChecker creates a new service health checker
func NewServiceHealthChecker(registry ServiceRegistry, interval time.Duration) *ServiceHealthChecker {
	return &ServiceHealthChecker{
		registry: registry,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start starts the health checker
func (h *ServiceHealthChecker) Start(ctx context.Context) {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.stopCh:
			return
		case <-ticker.C:
			h.checkAllServices(ctx)
		}
	}
}

// Stop stops the health checker
func (h *ServiceHealthChecker) Stop() {
	close(h.stopCh)
}

// checkAllServices checks the health of all registered services
func (h *ServiceHealthChecker) checkAllServices(ctx context.Context) {
	services, err := h.registry.ListServices(ctx)
	if err != nil {
		return
	}

	for _, service := range services {
		go h.checkService(ctx, service)
	}
}

// checkService checks the health of a single service
func (h *ServiceHealthChecker) checkService(ctx context.Context, service *ServiceInfo) {
	var status string
	var message string

	if service.Protocol == "http" {
		status, message = h.checkHTTPService(ctx, service)
	} else if service.Protocol == "grpc" {
		status, message = h.checkGRPCService(ctx, service)
	} else {
		status = "unknown"
		message = "Unsupported protocol"
	}

	// Update health status
	service.Health = HealthStatus{
		Status:    status,
		LastCheck: time.Now(),
		Message:   message,
	}

	// Re-register the service with updated health status
	h.registry.Register(ctx, service)
}

// checkHTTPService checks the health of an HTTP service
func (h *ServiceHealthChecker) checkHTTPService(ctx context.Context, service *ServiceInfo) (string, string) {
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("http://%s:%d/health", service.Address, service.Port)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "unhealthy", fmt.Sprintf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "unhealthy", fmt.Sprintf("Failed to connect: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return "healthy", "OK"
	}

	return "unhealthy", fmt.Sprintf("HTTP status: %d", resp.StatusCode)
}

// checkGRPCService checks the health of a gRPC service
func (h *ServiceHealthChecker) checkGRPCService(ctx context.Context, service *ServiceInfo) (string, string) {
	address := fmt.Sprintf("%s:%d", service.Address, service.Port)
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "unhealthy", fmt.Sprintf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// For now, just check if we can connect
	// In a real implementation, you would call a health check RPC
	return "healthy", "Connected"
}

// ServiceConfig represents configuration for a service
type ServiceConfig struct {
	Name     string            `json:"name"`
	Version  string            `json:"version"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Protocol string            `json:"protocol"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`
}

// NewServiceInfo creates a new ServiceInfo from ServiceConfig
func NewServiceInfo(config ServiceConfig) *ServiceInfo {
	return &ServiceInfo{
		ID:       fmt.Sprintf("%s-%s-%d", config.Name, config.Version, time.Now().Unix()),
		Name:     config.Name,
		Version:  config.Version,
		Address:  config.Address,
		Port:     config.Port,
		Protocol: config.Protocol,
		Tags:     config.Tags,
		Metadata: config.Metadata,
		Health: HealthStatus{
			Status:    "unknown",
			LastCheck: time.Time{},
			Message:   "Not checked yet",
		},
	}
}