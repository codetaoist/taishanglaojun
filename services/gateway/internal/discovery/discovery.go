package discovery

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/consul/api"
)

// Service represents a service instance
type Service struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`
}

// ServiceDiscovery interface defines methods for service discovery
type ServiceDiscovery interface {
	Register(service *Service) error
	Deregister(serviceID string) error
	GetService(serviceName string) ([]*Service, error)
	WatchService(serviceName string) (<-chan []*Service, error)
	Close() error
}

// ConsulDiscovery implements ServiceDiscovery using Consul
type ConsulDiscovery struct {
	client *api.Client
}

// Config holds configuration for service discovery
type Config struct {
	Type       string `mapstructure:"type"`
	Address    string `mapstructure:"address"`
	Datacenter string `mapstructure:"datacenter"`
	Token      string `mapstructure:"token"`
}

// NewClient creates a new service discovery client based on the configuration type
func NewClient(cfg Config) (ServiceDiscovery, error) {
	switch cfg.Type {
	case "mock":
		return NewMockDiscovery(), nil
	case "consul":
		config := api.DefaultConfig()
		config.Address = cfg.Address
		config.Datacenter = cfg.Datacenter
		config.Token = cfg.Token

		client, err := api.NewClient(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create consul client: %w", err)
		}

		return &ConsulDiscovery{client: client}, nil
	default:
		return nil, fmt.Errorf("unsupported discovery type: %s", cfg.Type)
	}
}

// Register registers a service with Consul
func (d *ConsulDiscovery) Register(service *Service) error {
	registration := &api.AgentServiceRegistration{
		ID:      service.ID,
		Name:    service.Name,
		Address: service.Address,
		Port:    service.Port,
		Tags:    service.Tags,
		Meta:    service.Metadata,
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/health", service.Address, service.Port),
			Interval: "10s",
			Timeout:  "3s",
		},
	}

	return d.client.Agent().ServiceRegister(registration)
}

// Deregister removes a service from Consul
func (d *ConsulDiscovery) Deregister(serviceID string) error {
	return d.client.Agent().ServiceDeregister(serviceID)
}

// GetService returns all instances of a service
func (d *ConsulDiscovery) GetService(serviceName string) ([]*Service, error) {
	services, _, err := d.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get service %s: %w", serviceName, err)
	}

	result := make([]*Service, 0, len(services))
	for _, service := range services {
		result = append(result, &Service{
			ID:       service.Service.ID,
			Name:     service.Service.Service,
			Address:  service.Service.Address,
			Port:     service.Service.Port,
			Tags:     service.Service.Tags,
			Metadata: service.Service.Meta,
		})
	}

	return result, nil
}

// WatchService watches for changes in a service
func (d *ConsulDiscovery) WatchService(serviceName string) (<-chan []*Service, error) {
	ch := make(chan []*Service, 10)

	go func() {
		defer close(ch)
		var lastIndex uint64

		for {
			queryOptions := &api.QueryOptions{
				WaitIndex: lastIndex,
				WaitTime:  30 * time.Second,
			}

			services, meta, err := d.client.Health().Service(serviceName, "", true, queryOptions)
			if err != nil {
				log.Printf("Error watching service %s: %v", serviceName, err)
				time.Sleep(5 * time.Second)
				continue
			}

			lastIndex = meta.LastIndex

			result := make([]*Service, 0, len(services))
			for _, service := range services {
				result = append(result, &Service{
					ID:       service.Service.ID,
					Name:     service.Service.Service,
					Address:  service.Service.Address,
					Port:     service.Service.Port,
					Tags:     service.Service.Tags,
					Metadata: service.Service.Meta,
				})
			}

			select {
			case ch <- result:
			default:
				// Channel is full, skip this update
			}
		}
	}()

	return ch, nil
}

// Close closes the Consul client
func (d *ConsulDiscovery) Close() error {
	// Consul client doesn't need explicit closing
	return nil
}

// MockDiscovery implements ServiceDiscovery for testing
type MockDiscovery struct {
	services map[string][]*Service
	watchers map[string]chan []*Service
}

// NewMockDiscovery creates a new mock service discovery
func NewMockDiscovery() ServiceDiscovery {
	md := &MockDiscovery{
		services: make(map[string][]*Service),
		watchers: make(map[string]chan []*Service),
	}
	
	// Add default services
	md.services["api"] = []*Service{
		{
			ID:      "api-1",
			Name:    "api",
			Address: "localhost",
			Port:    8082,
			Tags:    []string{"api"},
		},
	}
	
	md.services["auth"] = []*Service{
		{
			ID:      "auth-1",
			Name:    "auth",
			Address: "localhost",
			Port:    8081,
			Tags:    []string{"auth"},
		},
	}
	
	md.services["notification"] = []*Service{
		{
			ID:      "notification-1",
			Name:    "notification",
			Address: "localhost",
			Port:    8083,
			Tags:    []string{"notification"},
		},
	}
	
	return md
}

// Register registers a service in the mock discovery
func (m *MockDiscovery) Register(service *Service) error {
	if m.services[service.Name] == nil {
		m.services[service.Name] = make([]*Service, 0)
	}
	m.services[service.Name] = append(m.services[service.Name], service)

	// Notify watchers
	if ch, ok := m.watchers[service.Name]; ok {
		select {
		case ch <- m.services[service.Name]:
		default:
			// Channel is full, skip this update
		}
	}

	return nil
}

// Deregister removes a service from the mock discovery
func (m *MockDiscovery) Deregister(serviceID string) error {
	for name, services := range m.services {
		for i, service := range services {
			if service.ID == serviceID {
				m.services[name] = append(services[:i], services[i+1:]...)
				break
			}
		}
	}

	return nil
}

// GetService returns all instances of a service from the mock discovery
func (m *MockDiscovery) GetService(serviceName string) ([]*Service, error) {
	if services, ok := m.services[serviceName]; ok {
		return services, nil
	}
	return nil, fmt.Errorf("service %s not found", serviceName)
}

// WatchService watches for changes in a service in the mock discovery
func (m *MockDiscovery) WatchService(serviceName string) (<-chan []*Service, error) {
	ch := make(chan []*Service, 10)
	m.watchers[serviceName] = ch

	// Send initial state
	if services, ok := m.services[serviceName]; ok {
		go func() {
			ch <- services
		}()
	}

	// In a real implementation, we would watch for changes
	// For the mock, we'll just keep the channel open
	return ch, nil
}

// Close closes the mock discovery
func (m *MockDiscovery) Close() error {
	for _, ch := range m.watchers {
		close(ch)
	}
	return nil
}