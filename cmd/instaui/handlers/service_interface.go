package handlers

import "github.com/data-catering/insta-infra/v2/cmd/instaui/models"

// ServiceHandlerInterface defines the interface that all service handlers must implement
type ServiceHandlerInterface interface {
	// Basic service operations
	ListServices() []models.ServiceInfo
	GetServiceStatus(serviceName string) (string, error)
	GetServiceDependencies(serviceName string) ([]string, error)
	StartService(serviceName string, persist bool) error
	StopService(serviceName string) error
	StopAllServices() error

	// Batch operations for efficiency
	GetMultipleServiceStatuses(serviceNames []string) (map[string]models.ServiceStatus, error)
	GetAllRunningServices() (map[string]models.ServiceStatus, error)
	GetAllServiceDependencies() (map[string][]string, error)
	GetAllServicesWithStatusAndDependencies() ([]models.ServiceDetailInfo, error)

	// Operations with status updates
	StartServiceWithStatusUpdate(serviceName string, persist bool) (map[string]models.ServiceStatus, error)
	StopServiceWithStatusUpdate(serviceName string) (map[string]models.ServiceStatus, error)
	StopAllServicesWithStatusUpdate() (map[string]models.ServiceStatus, error)

	// Status management
	RefreshStatusFromContainers() (map[string]models.ServiceStatus, error)
	CheckStartingServicesProgress() (map[string]models.ServiceStatus, error)
}
