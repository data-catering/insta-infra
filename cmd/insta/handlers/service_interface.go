package handlers

import "github.com/data-catering/insta-infra/v2/cmd/insta/models"

// ServiceHandlerInterface defines the interface that all service handlers must implement
type ServiceHandlerInterface interface {
	// Basic service operations
	ListServices() []*models.EnhancedService
	GetServiceDependencies(serviceName string) ([]string, error)
	StartService(serviceName string, persist bool) error
	StopService(serviceName string) error
	StopAllServices() error

	// Enhanced service operations
	ListEnhancedServices() []*models.EnhancedService
	GetService(name string) (*models.EnhancedService, bool)

	// Get all operations - updated to use enhanced models
	GetMultipleServiceStatuses(serviceNames []string) (map[string]*models.EnhancedService, error)
	GetAllRunningServices() (map[string]*models.EnhancedService, error)
	GetAllServiceDependencies() (map[string][]string, error)
	GetAllServicesWithStatusAndDependencies() ([]*models.EnhancedService, error)
	GetAllDependencyStatuses() (map[string]*models.EnhancedService, error)

	// Operations with status updates - updated to use enhanced models
	StartServiceWithStatusUpdate(serviceName string, persist bool) (map[string]*models.EnhancedService, error)
	StopServiceWithStatusUpdate(serviceName string) (map[string]*models.EnhancedService, error)
	StopAllServicesWithStatusUpdate() (map[string]*models.EnhancedService, error)

	// Status management - updated to use enhanced models
	RefreshStatusFromContainers() (map[string]*models.EnhancedService, error)
	CheckStartingServicesProgress() (map[string]*models.EnhancedService, error)
}
