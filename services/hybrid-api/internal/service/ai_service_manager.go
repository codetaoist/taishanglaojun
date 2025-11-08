package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// Import generated protobuf packages
	// pb "github.com/codetaoist/api/proto"
)

// AIServiceClient wraps the gRPC clients for AI services
type AIServiceClient struct {
	vectorServiceClient interface{} // Will be pb.VectorServiceClient
	modelServiceClient  interface{} // Will be pb.ModelServiceClient
	conn                *grpc.ClientConn
	serviceManager      *ServiceManager
}

// NewAIServiceClient creates a new AI service client
func NewAIServiceClient(serviceManager *ServiceManager) *AIServiceClient {
	return &AIServiceClient{
		serviceManager: serviceManager,
	}
}

// Connect connects to the AI services
func (c *AIServiceClient) Connect(ctx context.Context) error {
	// Connect to vector service
	vectorConn, err := c.serviceManager.ConnectGRPC(ctx, "vector-service")
	if err != nil {
		return fmt.Errorf("failed to connect to vector service: %v", err)
	}

	// Connect to model service
	modelConn, err := c.serviceManager.ConnectGRPC(ctx, "model-service")
	if err != nil {
		return fmt.Errorf("failed to connect to model service: %v", err)
	}

	// In a real implementation, we would create the protobuf clients here
	// c.vectorServiceClient = pb.NewVectorServiceClient(vectorConn)
	// c.modelServiceClient = pb.NewModelServiceClient(modelConn)

	// For now, we'll store the connection
	c.conn = vectorConn // Use vector service connection as the primary connection

	log.Printf("Connected to AI services")
	return nil
}

// Close closes the connection to the AI services
func (c *AIServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// HealthCheck performs a health check on the AI services
func (c *AIServiceClient) HealthCheck(ctx context.Context) error {
	// Check vector service health
	vectorService, err := c.serviceManager.GetService(ctx, "vector-service")
	if err != nil {
		return fmt.Errorf("vector service not found: %v", err)
	}

	vectorHealth, err := c.serviceManager.HealthCheck(ctx, vectorService.ID)
	if err != nil {
		return fmt.Errorf("failed to check vector service health: %v", err)
	}

	if vectorHealth.Status != "healthy" {
		return fmt.Errorf("vector service is not healthy: %s", vectorHealth.Message)
	}

	// Check model service health
	modelService, err := c.serviceManager.GetService(ctx, "model-service")
	if err != nil {
		return fmt.Errorf("model service not found: %v", err)
	}

	modelHealth, err := c.serviceManager.HealthCheck(ctx, modelService.ID)
	if err != nil {
		return fmt.Errorf("failed to check model service health: %v", err)
	}

	if modelHealth.Status != "healthy" {
		return fmt.Errorf("model service is not healthy: %s", modelHealth.Message)
	}

	log.Printf("All AI services are healthy")
	return nil
}

// AIServiceManager manages AI services
type AIServiceManager struct {
	serviceManager      *ServiceManager
	aiServiceClient     *AIServiceClient
	healthChecker       *ServiceHealthChecker
	vectorServiceConfig ServiceConfig
	modelServiceConfig  ServiceConfig
}

// NewAIServiceManager creates a new AI service manager
func NewAIServiceManager() *AIServiceManager {
	registry := NewInMemoryServiceRegistry()
	connector := NewDefaultServiceConnector()
	serviceManager := NewServiceManager(registry, connector)
	aiServiceClient := NewAIServiceClient(serviceManager)
	healthChecker := NewServiceHealthChecker(registry, 30*time.Second)

	return &AIServiceManager{
		serviceManager:  serviceManager,
		aiServiceClient: aiServiceClient,
		healthChecker:   healthChecker,
		vectorServiceConfig: ServiceConfig{
			Name:     "vector-service",
			Version:  "1.0.0",
			Address:  "localhost",
			Port:     50051,
			Protocol: "grpc",
			Tags:     []string{"ai", "vector", "database"},
			Metadata: map[string]string{
				"description": "Vector database service",
			},
		},
		modelServiceConfig: ServiceConfig{
			Name:     "model-service",
			Version:  "1.0.0",
			Address:  "localhost",
			Port:     50051,
			Protocol: "grpc",
			Tags:     []string{"ai", "model", "inference"},
			Metadata: map[string]string{
				"description": "Model inference service",
			},
		},
	}
}

// Start starts the AI service manager
func (m *AIServiceManager) Start(ctx context.Context) error {
	// Register vector service
	vectorService := NewServiceInfo(m.vectorServiceConfig)
	if err := m.serviceManager.RegisterService(ctx, vectorService); err != nil {
		return fmt.Errorf("failed to register vector service: %v", err)
	}

	// Register model service
	modelService := NewServiceInfo(m.modelServiceConfig)
	if err := m.serviceManager.RegisterService(ctx, modelService); err != nil {
		return fmt.Errorf("failed to register model service: %v", err)
	}

	// Start health checker
	go m.healthChecker.Start(ctx)

	// Connect to AI services
	if err := m.aiServiceClient.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to AI services: %v", err)
	}

	log.Printf("AI service manager started")
	return nil
}

// Stop stops the AI service manager
func (m *AIServiceManager) Stop() error {
	// Stop health checker
	m.healthChecker.Stop()

	// Close AI service client connection
	if err := m.aiServiceClient.Close(); err != nil {
		return fmt.Errorf("failed to close AI service client: %v", err)
	}

	log.Printf("AI service manager stopped")
	return nil
}

// GetServiceManager returns the service manager
func (m *AIServiceManager) GetServiceManager() *ServiceManager {
	return m.serviceManager
}

// GetAIServiceClient returns the AI service client
func (m *AIServiceManager) GetAIServiceClient() *AIServiceClient {
	return m.aiServiceClient
}

// UpdateServiceConfig updates the configuration for a service
func (m *AIServiceManager) UpdateServiceConfig(ctx context.Context, serviceName string, config ServiceConfig) error {
	// Unregister the existing service
	services, err := m.serviceManager.GetServices(ctx, serviceName)
	if err == nil && len(services) > 0 {
		for _, service := range services {
			if err := m.serviceManager.UnregisterService(ctx, service.ID); err != nil {
				return fmt.Errorf("failed to unregister service %s: %v", service.ID, err)
			}
		}
	}

	// Register the service with new configuration
	newService := NewServiceInfo(config)
	if err := m.serviceManager.RegisterService(ctx, newService); err != nil {
		return fmt.Errorf("failed to register service with new config: %v", err)
	}

	// Update the stored configuration
	if serviceName == "vector-service" {
		m.vectorServiceConfig = config
	} else if serviceName == "model-service" {
		m.modelServiceConfig = config
	}

	// Reconnect to AI services
	if err := m.aiServiceClient.Connect(ctx); err != nil {
		return fmt.Errorf("failed to reconnect to AI services: %v", err)
	}

	log.Printf("Updated configuration for service %s", serviceName)
	return nil
}

// GetServiceHealth returns the health status of a service
func (m *AIServiceManager) GetServiceHealth(ctx context.Context, serviceName string) (*HealthStatus, error) {
	service, err := m.serviceManager.GetService(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("service %s not found: %v", serviceName, err)
	}

	return m.serviceManager.HealthCheck(ctx, service.ID)
}

// ListAllServices returns all registered services
func (m *AIServiceManager) ListAllServices(ctx context.Context) ([]*ServiceInfo, error) {
	return m.serviceManager.ListServices(ctx)
}

// HybridServiceManager manages the hybrid architecture
type HybridServiceManager struct {
	aiServiceManager *AIServiceManager
	// Other service managers can be added here
}

// NewHybridServiceManager creates a new hybrid service manager
func NewHybridServiceManager() *HybridServiceManager {
	return &HybridServiceManager{
		aiServiceManager: NewAIServiceManager(),
	}
}

// Start starts the hybrid service manager
func (m *HybridServiceManager) Start(ctx context.Context) error {
	// Start AI service manager
	if err := m.aiServiceManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start AI service manager: %v", err)
	}

	log.Printf("Hybrid service manager started")
	return nil
}

// Stop stops the hybrid service manager
func (m *HybridServiceManager) Stop() error {
	// Stop AI service manager
	if err := m.aiServiceManager.Stop(); err != nil {
		return fmt.Errorf("failed to stop AI service manager: %v", err)
	}

	log.Printf("Hybrid service manager stopped")
	return nil
}

// GetAIServiceManager returns the AI service manager
func (m *HybridServiceManager) GetAIServiceManager() *AIServiceManager {
	return m.aiServiceManager
}