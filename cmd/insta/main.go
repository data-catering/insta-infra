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

	"github.com/data-catering/insta-infra/cmd/insta/container"
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
	dataDir string
	tempDir string
	runtime container.Runtime
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

	// Expand the default data directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Use INSTA_DATA_DIR if set, otherwise use default
	dataDir := os.Getenv("INSTA_DATA_DIR")
	if dataDir == "" {
		dataDir = filepath.Join(homeDir, ".insta/data")
	}

	// Create a temporary directory for docker-compose files
	tempDir, err := os.MkdirTemp("", "insta-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// Extract embedded docker-compose files to temp directory
	composeContent, err := embedFS.ReadFile("resources/docker-compose.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded docker-compose.yaml: %w", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "docker-compose.yaml"), composeContent, 0644); err != nil {
		return nil, fmt.Errorf("failed to write docker-compose.yaml: %w", err)
	}

	persistContent, err := embedFS.ReadFile("resources/docker-compose-persist.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded docker-compose-persist.yaml: %w", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "docker-compose-persist.yaml"), persistContent, 0644); err != nil {
		return nil, fmt.Errorf("failed to write docker-compose-persist.yaml: %w", err)
	}

	// Extract data directory files
	if err := extractDataFiles(tempDir, embedFS); err != nil {
		return nil, fmt.Errorf("failed to extract data files: %w", err)
	}

	return &App{
		dataDir: dataDir,
		tempDir: tempDir,
		runtime: provider.SelectedRuntime(),
	}, nil
}

func extractDataFiles(tempDir string, embedFS embed.FS) error {
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
		relPath, err := filepath.Rel("resources", path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		if d.IsDir() {
			targetDir := filepath.Join(tempDir, relPath)
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
			}
			return nil
		}

		content, err := embedFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		targetFile := filepath.Join(tempDir, relPath)
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
	return os.RemoveAll(a.tempDir)
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

	composeFiles := []string{filepath.Join(a.tempDir, "docker-compose.yaml")}
	if persist {
		composeFiles = append(composeFiles, filepath.Join(a.tempDir, "docker-compose-persist.yaml"))
	}

	fmt.Printf("%sStarting up services...%s\n", colorGreen, colorReset)
	if err := a.runtime.ComposeUp(composeFiles, services, true); err != nil {
		return fmt.Errorf("%sError: Failed to start up services%s", colorRed, colorReset)
	}

	return nil
}

func (a *App) stopServices(services []string) error {
	composeFiles := []string{
		filepath.Join(a.tempDir, "docker-compose.yaml"),
		filepath.Join(a.tempDir, "docker-compose-persist.yaml"),
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
    %s -d                   Bring Postgres down
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
		defer app.Cleanup()

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
		defer app.Cleanup()

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
		defer app.Cleanup()

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
		defer app.Cleanup()

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
		defer app.Cleanup()

		if err := app.update(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}
}
