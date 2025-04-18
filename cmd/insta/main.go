package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/data-catering/insta-infra/v2/cmd/insta/container"
)

// Version information - these will be set during build via ldflags
var (
	version   = "dev"
	buildTime = "unknown"
)

//go:embed resources/docker-compose.yaml resources/docker-compose-persist.yaml
//go:embed all:resources/data
var embedFS embed.FS

// ANSI color codes
const (
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[1;33m"
	colorBlue   = "\033[1;34m"
	colorReset  = "\033[0m"
)

// Default data directory for persisted data
const defaultDataDir = "~/.insta/data"

type App struct {
	dataDir    string
	instaDir   string
	runtime    container.Runtime
}

func NewApp(runtimeName string) (*App, error) {
	// Initialize container runtime
	provider := container.NewProvider()
	if err := provider.DetectRuntime(); err != nil {
		return nil, fmt.Errorf(`failed to detect container runtime: %w

Please ensure one of the following is available:
1. Docker (20.10+) with Docker Compose plugin
2. Podman (3.0+) with Podman Compose plugin or podman-compose

For Docker:
  - Install Docker Desktop or Docker Engine
  - Install Docker Compose plugin

For Podman:
  - Install Podman
  - Install Podman Compose plugin or podman-compose
  - On macOS, ensure podman machine is running (podman machine start)`, err)
	}

	// If runtime is explicitly specified, try to use it
	if runtimeName != "" {
		if err := provider.SetRuntime(runtimeName); err != nil {
			return nil, fmt.Errorf("failed to set runtime to %s: %w", runtimeName, err)
		}
	}

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Use INSTA_HOME if set, otherwise use ~/.insta
	instaDir := os.Getenv("INSTA_HOME")
	if instaDir == "" {
		instaDir = filepath.Join(homeDir, ".insta")
	}

	// Create insta directory
	if err := os.MkdirAll(instaDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create insta directory: %w", err)
	}

	// Data directory is always relative to insta directory
	dataDir := filepath.Join(instaDir, "data")

	// Check if docker-compose.yaml exists
	composePath := filepath.Join(instaDir, "docker-compose.yaml")
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		// Extract docker-compose.yaml only if it doesn't exist
		composeContent, err := embedFS.ReadFile("resources/docker-compose.yaml")
		if err != nil {
			return nil, fmt.Errorf("failed to read embedded docker-compose.yaml: %w", err)
		}
		if err := os.WriteFile(composePath, composeContent, 0644); err != nil {
			return nil, fmt.Errorf("failed to write docker-compose.yaml: %w", err)
		}
	}

	// Check if docker-compose-persist.yaml exists
	persistPath := filepath.Join(instaDir, "docker-compose-persist.yaml")
	if _, err := os.Stat(persistPath); os.IsNotExist(err) {
		// Extract docker-compose-persist.yaml only if it doesn't exist
		persistContent, err := embedFS.ReadFile("resources/docker-compose-persist.yaml")
		if err != nil {
			return nil, fmt.Errorf("failed to read embedded docker-compose-persist.yaml: %w", err)
		}
		if err := os.WriteFile(persistPath, persistContent, 0644); err != nil {
			return nil, fmt.Errorf("failed to write docker-compose-persist.yaml: %w", err)
		}
	}

	// Check if data directory exists and extract files if needed
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		// Extract data directory files only if the directory doesn't exist
		if err := extractDataFiles(instaDir, embedFS); err != nil {
			return nil, fmt.Errorf("failed to extract data files: %w", err)
		}
	}

	return &App{
		dataDir:    dataDir,
		instaDir:   instaDir,
		runtime:    provider.SelectedRuntime(),
	}, nil
}

func extractDataFiles(tempDir string, embedFS embed.FS) error {
	// Create data directory in temp dir
	dataDir := filepath.Join(tempDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	return fs.WalkDir(embedFS, "resources/data", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip all persist directories and files within them
		if strings.Contains(path, "persist") {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		// Get relative path from resources/data to use for target
		relPath, err := filepath.Rel("resources/data", path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		if d.IsDir() {
			targetDir := filepath.Join(dataDir, relPath)
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
			}
			return nil
		}

		content, err := embedFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		targetFile := filepath.Join(dataDir, relPath)
		targetDir := filepath.Dir(targetFile)

		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
		}

		if err := os.WriteFile(targetFile, content, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", targetFile, err)
		}

		return nil
	})
}

func (a *App) Cleanup() error {
	if a.instaDir != "" {
		return os.RemoveAll(a.instaDir)
	}
	return nil
}

func (a *App) checkRuntime() error {
	if err := a.runtime.CheckAvailable(); err != nil {
		return fmt.Errorf("%sError: %s not available: %v%s", colorRed, a.runtime.Name(), err, colorReset)
	}
	return nil
}

func (a *App) listServices() error {
	var serviceNames []string
	for name := range Services {
		serviceNames = append(serviceNames, name)
	}

	sort.Strings(serviceNames)
	fmt.Printf("%sSupported services: %s%s%s\n",
		colorGreen,
		strings.Join(serviceNames, " "),
		colorReset,
		colorReset)

	return nil
}

func (a *App) startServices(services []string, persist bool) error {
	if persist {
		fmt.Printf("%sPersisting data to %s%s\n", colorYellow, a.dataDir, colorReset)

		// Create data directory
		if err := os.MkdirAll(a.dataDir, 0755); err != nil {
			return fmt.Errorf("failed to create data directory: %w", err)
		}

		// Create persist directories for each service
		for _, service := range services {
			persistDir := filepath.Join(a.dataDir, "data", service, "persist")
			if err := os.MkdirAll(persistDir, 0755); err != nil {
				return fmt.Errorf("failed to create persist directory for %s: %w", service, err)
			}
		}
	}

	composeFiles := []string{filepath.Join(a.instaDir, "docker-compose.yaml")}
	if persist {
		composeFiles = append(composeFiles, filepath.Join(a.instaDir, "docker-compose-persist.yaml"))
	}

	fmt.Printf("%sStarting up services...%s\n", colorGreen, colorReset)
	if err := a.runtime.ComposeUp(composeFiles, services, true); err != nil {
		return fmt.Errorf("%sError: Failed to start up services: %v%s", colorRed, err, colorReset)
	}

	// Get the expanded list of services including recursive dependencies
	expandedServices := make(map[string]bool)
	
	// Function to recursively collect dependencies
	var collectDependencies func(service string) error
	collectDependencies = func(service string) error {
		if expandedServices[service] {
			return nil // Already processed this service
		}
		
		expandedServices[service] = true
		
		// Get dependencies for this service
		deps, err := a.runtime.GetDependencies(service, composeFiles)
		if err != nil {
			return fmt.Errorf("failed to get dependencies for %s: %w", service, err)
		}
		
		// Recursively process each dependency
		for _, dep := range deps {
			if err := collectDependencies(dep); err != nil {
				// Just log the error and continue
				fmt.Printf("%sWarning: %v%s\n", colorYellow, err, colorReset)
			}
		}
		
		return nil
	}
	
	// Process each requested service
	for _, service := range services {
		if err := collectDependencies(service); err != nil {
			fmt.Printf("%sWarning: Failed to collect all dependencies: %v%s\n", colorYellow, err, colorReset)
		}
	}

	// Extract all services to display
	var servicesToDisplay []string
	for service := range expandedServices {
		servicesToDisplay = append(servicesToDisplay, service)
	}
	sort.Strings(servicesToDisplay)

	// Display connection information for all services in a single table
	fmt.Printf("\n%sConnection Information Table%s\n", colorBlue, colorReset)
	fmt.Printf("%s┌─────────────────────────┬──────────────────────────────┬──────────────────────┬──────────────────────────────┬────────────┬────────────┐%s\n", colorYellow, colorReset)
	fmt.Printf("%s│ Service                 │ Container to Container       │ Host to Container    │ Container to Host            │ Username   │ Password   │%s\n", colorYellow, colorReset)
	fmt.Printf("%s├─────────────────────────┼──────────────────────────────┼──────────────────────┼──────────────────────────────┼────────────┼────────────┤%s\n", colorYellow, colorReset)

	// Track if any services with ports were displayed
	servicesDisplayed := false

	// Print each service row
	for _, serviceName := range servicesToDisplay {
		// Get port information from the container runtime
		portMappings, err := a.runtime.GetPortMappings(serviceName)
		// Skip services without any port mappings
		if err != nil || len(portMappings) == 0 {
			continue
		}
		
		// Use the first port mapping
		portInfo := "N/A"
		for _, hostPort := range portMappings {
			portInfo = hostPort
			break
		}
		
		servicesDisplayed = true
		
		if service, exists := Services[serviceName]; exists {
			// Get username and password, defaulting to empty string if not set
			username := ""
			if service.DefaultUser != "" {
				username = service.DefaultUser
			}
			password := ""
			if service.DefaultPassword != "" {
				password = service.DefaultPassword
			}

			fmt.Printf("%s│ %-23s │ %-28s │ %-20s │ %-28s │ %-10s │ %-10s │%s\n",
				colorYellow,
				serviceName,
				fmt.Sprintf("%s:%s", serviceName, portInfo),
				fmt.Sprintf("localhost:%s", portInfo),
				fmt.Sprintf("host.docker.internal:%s", portInfo),
				username,
				password,
				colorReset)
		} else {
			// For services not in the Services map, still display what we know
			fmt.Printf("%s│ %-23s │ %-28s │ %-20s │ %-28s │ %-10s │ %-10s │%s\n",
				colorYellow,
				serviceName,
				fmt.Sprintf("%s:%s", serviceName, portInfo),
				fmt.Sprintf("localhost:%s", portInfo),
				fmt.Sprintf("host.docker.internal:%s", portInfo),
				"N/A",
				"N/A",
				colorReset)
		}
	}

	// If no services were displayed, show a message
	if !servicesDisplayed {
		fmt.Printf("%s│ %-23s │ %-28s │ %-20s │ %-28s │ %-10s │ %-10s │%s\n",
			colorYellow,
			"No services with ports",
			"N/A",
			"N/A",
			"N/A",
			"N/A",
			"N/A",
			colorReset)
	}

	// Print footer
	fmt.Printf("%s└─────────────────────────┴──────────────────────────────┴──────────────────────┴──────────────────────────────┴────────────┴────────────┘%s\n", colorYellow, colorReset)
	fmt.Println()

	return nil
}

func (a *App) stopServices(services []string) error {
	composeFiles := []string{
		filepath.Join(a.instaDir, "docker-compose.yaml"),
		filepath.Join(a.instaDir, "docker-compose-persist.yaml"),
	}

	if len(services) > 0 {
		fmt.Printf("Shutting down services: %s...\n", strings.Join(services, " "))
	} else {
		fmt.Println("Shutting down all services...")
	}

	return a.runtime.ComposeDown(composeFiles, services)
}

func (a *App) connectToService(serviceName string) error {
	if serviceName == "" {
		return fmt.Errorf("%sError: No service name passed as argument%s", colorRed, colorReset)
	}

	service, exists := Services[serviceName]
	if !exists {
		return fmt.Errorf("%sError: Unknown service %s%s", colorRed, serviceName, colorReset)
	}

	fmt.Printf("%sConnecting to %s...%s\n", colorGreen, serviceName, colorReset)

	var cmd string
	if service.ConnectionCmd == "bash" {
		return a.runtime.ExecInContainer(serviceName, "", true)
	}

	// Check for command arguments after --
	args := os.Args[2:]
	for i, arg := range args {
		if arg == "--" && i+1 < len(args) {
			cmd = strings.Join(args[i+1:], " ")
			break
		}
	}

	if cmd == "" {
		if service.RequiresPassword {
			cmd = fmt.Sprintf("export %s_USER=%s && export %s_PASSWORD=%s && %s",
				strings.ToUpper(serviceName), service.DefaultUser,
				strings.ToUpper(serviceName), service.DefaultPassword,
				service.ConnectionCmd)
		} else {
			cmd = service.ConnectionCmd
		}
	}

	return a.runtime.ExecInContainer(serviceName, cmd, true)
}

func usage() {
	fmt.Printf(`insta-infra %s (built: %s)
Usage: %s [options...] [services...]

    <services>                Name of services to run
    -c, connect [service]     Connect to service
    -d, down [services...]    Shutdown services (if empty, shutdown all services)
    -h, help                  Show this help message
    -r, runtime [name]        Specify container runtime (docker or podman)
    -u, update                Check for and install updates
    -v, version               Show version information

Examples:
    %s -l                   List supported services
    %s postgres             Spin up Postgres
    %s -c postgres          Connect to Postgres
    %s -d postgres          Bring Postgres down
    %s -p postgres          Run Postgres with persisted data
    %s -r docker postgres   Run Postgres using Docker
    %s -r podman postgres   Run Postgres using Podman
`, version, buildTime, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}

func main() {
	// Define flags
	connectCmd := flag.NewFlagSet("connect", flag.ExitOnError)
	downCmd := flag.NewFlagSet("down", flag.ExitOnError)

	// Add update and runtime flags
	update := flag.Bool("update", false, "Check for and install updates")
	runtime := flag.String("runtime", "", "Explicitly set container runtime (docker/podman)")

	if len(os.Args) < 2 {
		usage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "-h", "--help", "help":
		usage()
		return

	case "-v", "--version", "version":
		fmt.Printf("insta-infra %s (built: %s)\n", version, buildTime)
		return

	case "-l", "list":
		app, err := NewApp(*runtime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%sError: %v%s\n", colorRed, err, colorReset)
			os.Exit(1)
		}

		if err := app.listServices(); err != nil {
			fmt.Fprintf(os.Stderr, "%sError: %v%s\n", colorRed, err, colorReset)
			os.Exit(1)
		}

	case "-c", "connect":
		app, err := NewApp(*runtime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%sError: %v%s\n", colorRed, err, colorReset)
			os.Exit(1)
		}

		connectCmd.Parse(os.Args[2:])
		if connectCmd.NArg() < 1 {
			fmt.Fprintf(os.Stderr, "%sError: No service specified%s\n", colorRed, colorReset)
			os.Exit(1)
		}
		if err := app.connectToService(connectCmd.Arg(0)); err != nil {
			fmt.Fprintf(os.Stderr, "%sError: %v%s\n", colorRed, err, colorReset)
			os.Exit(1)
		}

	case "-d", "down":
		app, err := NewApp(*runtime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%sError: %v%s\n", colorRed, err, colorReset)
			os.Exit(1)
		}

		downCmd.Parse(os.Args[2:])
		if err := app.stopServices(downCmd.Args()); err != nil {
			fmt.Fprintf(os.Stderr, "%sError: %v%s\n", colorRed, err, colorReset)
			os.Exit(1)
		}

	default:
		app, err := NewApp(*runtime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%sError: %v%s\n", colorRed, err, colorReset)
			os.Exit(1)
		}

		// Check for persist flag
		startArgs := os.Args[1:]
		persistIndex := -1
		for i, arg := range startArgs {
			if arg == "-p" {
				persistIndex = i
				break
			}
		}

		persist := persistIndex >= 0
		if persist {
			// Remove the -p flag from arguments
			startArgs = append(startArgs[:persistIndex], startArgs[persistIndex+1:]...)
		}

		if err := app.checkRuntime(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		if err := app.startServices(startArgs, persist); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	}

	// Handle update command
	if *update {
		app, err := NewApp(*runtime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if err := app.update(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}
}
