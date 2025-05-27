# insta-infra

![insta-infra services](docs/img/insta-infra.gif)

A simple, fast CLI tool for spinning up data infrastructure services using Docker or Podman.

## Features

- Run data infrastructure services with a single command
- Supports both Docker and Podman container runtimes
- Embed all configuration files in the binary for easy distribution
- Optional data persistence
- Connect to services with pre-configured environment variables

## Installation

### Using Homebrew

```bash
# Add the tap
brew tap data-catering/insta-infra

# Install insta-infra
brew install insta-infra
```

### From Source

```bash
# Clone the repository
git clone https://github.com/data-catering/insta-infra.git
cd insta-infra

# Build and install
make install
```

### Using Go

```bash
go install github.com/data-catering/insta-infra/v2/cmd/insta@v2.1.3
```

### Manual Installation

If you prefer to install manually from release archives:

1. Visit the [GitHub releases page](https://github.com/data-catering/insta-infra/releases)
2. Download the appropriate archive for your system:
   - For macOS ARM64: `insta-v2.1.3-darwin-arm64.tar.gz`
   - For macOS Intel: `insta-v2.1.3-darwin-amd64.tar.gz`
   - For Linux ARM64: `insta-v2.1.3-linux-arm64.tar.gz`
   - For Linux Intel: `insta-v2.1.3-linux-amd64.tar.gz`
   - For Windows ARM64: `insta-v2.1.3-windows-arm64.zip`
   - For Windows Intel: `insta-v2.1.3-windows-amd64.zip`

3. Extract the archive:
   ```bash
   # For .tar.gz files
   tar -xzf insta-v2.1.3-<os>-<arch>.tar.gz
   
   # For .zip files (Windows)
   unzip insta-v2.1.3-windows-<arch>.zip
   ```

4. Move the binary to a directory in your PATH:
   ```bash
   # For macOS/Linux
   sudo mv insta /usr/local/bin/
   
   # For Windows (PowerShell as Administrator)
   Move-Item insta.exe C:\Windows\System32\
   ```

5. Make the binary executable (macOS/Linux only):
   ```bash
   chmod +x /usr/local/bin/insta
   ```

## Requirements

- Docker (20.10+) or Podman (3.0+)
- For Docker: Docker Compose plugin
- For Podman: Podman Compose plugin or podman-compose

## Usage

```bash
# List available services
insta -l

# Start a service
insta postgres

# Start multiple services
insta postgres mysql elasticsearch

# Start a service with persistent data
insta -p postgres

# Connect to a running service
insta -c postgres

# Shutdown services
insta -d postgres

# Shutdown all services
insta -d

# Explicitly start a service in docker or podman
insta -r docker postgres
insta -r podman postgres

# Show help
insta -h

# Show version
insta -v
```

## Web UI

insta-infra also provides a modern graphical interface built with Wails for users who prefer visual service management.

![Web UI Screenshot](docs/img/web-ui-demo.png)

### Features

- **Visual Service Management**: See all available services in an intuitive grid layout
- **One-Click Actions**: Start, stop, and manage services with simple button clicks
- **Real-Time Status**: Live updates of service status with color-coded indicators
- **Connection Details**: Easy access to connection strings, credentials, and web UIs
- **Data Persistence**: Toggle data persistence with checkboxes
- **Browser Integration**: Direct "Open" buttons for web-based services
- **Dependency Visualization**: Clear display of service dependencies

### Building the Web UI

#### Prerequisites

In addition to the standard requirements, the Web UI requires:
- **Node.js** (16+) and **npm**
- **Wails CLI**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

#### Development Mode

To run the Web UI in development mode with hot reload:

```bash
# Clone and navigate to the project
git clone https://github.com/data-catering/insta-infra.git
cd insta-infra

# Start the development server
cd cmd/instaui
wails dev
```

This will:
1. Start the Go backend
2. Launch the React frontend with hot reload
3. Open the application window automatically

#### Production Build

To build a production version of the Web UI:

```bash
# Build the Web UI binary
make build-ui

# Or build manually
cd cmd/instaui
wails build
```

The built application will be available in the `cmd/instaui/build/bin/` directory.

### Using the Web UI

1. **Launch the application**:
   ```bash
   # If you built with make
   ./insta-infra-ui
   
   # If you built manually
   ./cmd/instaui/build/bin/instaui
   ```

2. **Service Management**:
   - Browse available services in the main grid
   - Click **Start** to launch a service (check "Persist data" for persistence)
   - Click **Stop** to shut down running services
   - Use **Stop All** to shut down all running services at once

3. **Connecting to Services**:
   - **Open**: For web-based services (Grafana, Kibana, etc.), click to open in browser
   - **Connect**: View connection details, URLs, credentials, and CLI commands
   - **Copy**: All connection details have copy-to-clipboard functionality

4. **Status Monitoring**:
   - Green indicators show running services
   - Gray indicators show stopped services
   - Red indicators show error states
   - Real-time updates every 30 seconds

### Cross-Platform Support

The Web UI supports the same platforms as the CLI:
- **macOS**: ARM64 and Intel
- **Linux**: ARM64 and Intel  
- **Windows**: ARM64 and Intel

### Troubleshooting

**Application won't start**:
- Ensure Docker/Podman is running
- Check that required ports aren't already in use
- Try running with `wails dev` for detailed error messages

**Services won't start**:
- Verify container runtime is accessible
- Check system resources (memory, disk space)
- Review Docker/Podman logs for specific errors

**UI not updating**:
- Click the refresh button manually
- Check network connectivity if using remote Docker
- Restart the application if status seems stuck

## Data Persistence

By default, all data is stored in memory and will be lost when the containers are stopped. To enable persistence, use the `-p` flag:

```bash
insta -p postgres
```

This will store data in `~/.insta/data/<service_name>/persist/`.

## Development

### Project Structure

```
.
├── cmd/
│   ├── insta/          # Main CLI application
│   │   ├── container/  # Container runtime implementations
│   │   ├── resources/  # Embedded resources
│   │   │   ├── data/   # Service configuration files
│   │   │   └── *.yaml  # Docker compose files
│   │   ├── models.go   # Service definitions
│   │   └── main.go     # CLI entry point
│   └── instaui/        # Web UI application (Wails)
│       ├── frontend/   # React frontend
│       │   ├── src/    # React components and styles
│       │   └── dist/   # Built frontend assets
│       ├── app.go      # Wails backend methods
│       └── main.go     # Web UI entry point
├── internal/
│   └── core/           # Shared business logic
│       ├── models.go   # Service definitions (shared)
│       └── service.go  # Service management logic
├── tests/              # Integration tests
├── docs/               # Documentation and images
├── Makefile            # Build and development tasks
└── README.md           # Documentation
```

### Development Workflow

#### CLI Development

1. Clone the repository
2. Make changes to CLI code in `cmd/insta/`
3. Run tests: `make test`
4. Build: `make build`
5. Run: `./insta`

#### Web UI Development

1. Clone the repository
2. Install dependencies: `cd cmd/instaui && npm install`
3. Start development mode: `wails dev`
4. Make changes to:
   - **Go backend**: `cmd/instaui/app.go`
   - **React frontend**: `cmd/instaui/frontend/src/`
   - **Shared logic**: `internal/core/`
5. Build for production: `make build-ui`

#### Full Development Environment

```bash
# Install all dependencies
make deps

# Run all tests
make test

# Build both CLI and Web UI
make build-all

# Clean build artifacts
make clean
```

### Adding a New Service

1. Add service configuration to [`docker-compose.yaml`](cmd/insta/resources/docker-compose.yaml)
2. Add service definition to [`internal/core/models.go`](internal/core/models.go)
3. Add any necessary initialization scripts to [`cmd/insta/resources/data/<service_name>/`](cmd/insta/resources/data/)
4. Update tests
5. Test in both CLI and Web UI

## Services

| Service Type                | Services                                                                                                                               |
|-----------------------------|----------------------------------------------------------------------------------------------------------------------------------------|
| Api Gateway                 | kong                                                                                                                                   |
| Cache                       | redis                                                                                                                                  |
| Change Data Capture         | debezium                                                                                                                               |
| Code Analysis               | sonarqube                                                                                                                              |
| Data Annotation             | argilla, cvat, doccano, label-studio                                                                                                   |
| Data Catalog                | amundsen, datahub, lakekeeper, marquez, openmetadata, polaris, unitycatalog                                                            |
| Data Collector              | fluentd, logstash                                                                                                                      |
| Data Visualisation          | blazer, evidence, grafana, metabase, redash, superset                                                                                  |
| Database                    | cassandra, cockroachdb, elasticsearch, influxdb, mariadb, milvus, mongodb, mssql, mysql, neo4j, opensearch, postgres, qdrant, spanner, sqlite, timescaledb, weaviate |
| Distributed Coordination    | zookeeper                                                                                                                              |
| Distributed Data Processing | flink, ray                                                                                                                             |
| Feature Store               | feast                                                                                                                                  |
| Identity Management         | keycloak                                                                                                                               |
| Job Orchestrator            | airflow, dagster, mage-ai, prefect                                                                                                     |
| ML Platform                 | mlflow                                                                                                                                 |
| Messaging                   | activemq, kafka, nats, pulsar, rabbitmq, solace                                                                                        |
| Monitoring                  | loki, prometheus                                                                                                                       |
| Notebook                    | jupyter                                                                                                                                |
| Object Storage              | minio                                                                                                                                  |
| Query Engine                | duckdb, flight-sql, presto, trino                                                                                                      |
| Real-time OLAP              | clickhouse, doris, druid, pinot                                                                                                        |
| Schema Registry             | confluent-schema-registry                                                                                                              |
| Secret Management           | vault                                                                                                                                  |
| Test Data Management        | data-caterer                                                                                                                           |
| Tracing                     | jaeger                                                                                                                                 |
| Web Server                  | httpbin, httpd                                                                                                                         |
| Workflow                    | maestro, temporal                                                                                                                      |

## Updating

### Using Package Managers

If you installed via a package manager, you can update using the standard update commands:

```bash
# Debian/Ubuntu
sudo apt update && sudo apt upgrade

# RHEL/CentOS/Fedora
sudo dnf update
# or
sudo yum update

# Arch Linux
sudo pacman -Syu

# macOS (Homebrew)
brew upgrade

# Windows (Chocolatey)
choco upgrade insta
```

### Manual Update

If you prefer to update manually:

1. Download the latest release from the [GitHub releases page](https://github.com/data-catering/insta-infra/releases)
2. Replace your existing binary with the new one
3. Make sure the binary is executable: `chmod +x insta`
