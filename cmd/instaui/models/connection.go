package models

// ServiceConnectionInfo holds connection details for a service
type ServiceConnectionInfo struct {
	ServiceName       string `json:"serviceName"`
	HasWebUI          bool   `json:"hasWebUI"`
	WebURL            string `json:"webURL,omitempty"`
	HostPort          string `json:"hostPort,omitempty"`
	ContainerPort     string `json:"containerPort,omitempty"`
	Username          string `json:"username,omitempty"`
	Password          string `json:"password,omitempty"`
	ConnectionCommand string `json:"connectionCommand,omitempty"`
	ConnectionString  string `json:"connectionString,omitempty"`
	Available         bool   `json:"available"`
	Error             string `json:"error,omitempty"`
}
