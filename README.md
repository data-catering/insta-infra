# insta-infra

![insta-infra services](docs/img/insta-infra.gif)

A simple, fast CLI tool for spinning up data infrastructure services using Docker or Podman.

> [!NOTE]
> [Check out the demo UI](https://data-catering.github.io/insta-infra/demo/ui/index.html)

## Features

- Run data infrastructure services with a single command
- Supports both Docker and Podman container runtimes
- Embed all configuration files in the binary for easy distribution
- Optional data persistence
- Connect to services with pre-configured environment variables

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/data-catering/insta-infra/main/install.sh | sh
```
OR
```bash
wget -q -O - https://raw.githubusercontent.com/data-catering/insta-infra/main/install.sh | sh
```

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
go install github.com/data-catering/insta-infra/v2/cmd/insta@v3.0.0
```

### Manual Installation

If you prefer to install manually from release archives:

1. Visit the [GitHub releases page](https://github.com/data-catering/insta-infra/releases)
2. Download the appropriate archive for your system:
   - For macOS ARM64: `insta-v3.0.0-darwin-arm64.tar.gz`
   - For macOS Intel: `insta-v3.0.0-darwin-amd64.tar.gz`
   - For Linux ARM64: `insta-v3.0.0-linux-arm64.tar.gz`
   - For Linux Intel: `insta-v3.0.0-linux-amd64.tar.gz`
   - For Windows ARM64: `insta-v3.0.0-windows-arm64.zip`
   - For Windows Intel: `insta-v3.0.0-windows-amd64.zip`

3. Extract the archive:
   ```bash
   # For .tar.gz files
   tar -xzf insta-v3.0.0-<os>-<arch>.tar.gz
   
   # For .zip files (Windows)
   unzip insta-v3.0.0-windows-<arch>.zip
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

```bash
insta --ui
```

[Check out the demo UI](https://data-catering.github.io/insta-infra/demo/ui/index.html).

### Features

- **Visual Service Management**: See all available services in an intuitive grid layout
- **One-Click Actions**: Start, stop, and manage services with simple button clicks
- **Real-Time Status**: Live updates of service status with color-coded indicators
- **Connection Details**: Easy access to connection strings, credentials, and web UIs
- **Data Persistence**: Toggle data persistence with checkboxes
- **Browser Integration**: Direct "Open" buttons for web-based services
- **Dependency Visualization**: Clear display of service dependencies

## Configuration

### Custom Container Runtime Paths

If Docker or Podman is installed in a non-standard location, you can specify custom paths using environment variables:

```bash
# Custom Docker path
export INSTA_DOCKER_PATH="/path/to/docker"
insta postgres

# Custom Podman path  
export INSTA_PODMAN_PATH="/path/to/podman"
insta -r podman postgres
```

This is particularly useful for:
- **macOS GUI applications**: When Docker/Podman isn't in the standard PATH
- **Custom installations**: When using alternative installation methods
- **Enterprise environments**: When binaries are in non-standard locations
- **Development setups**: When testing with different container runtime versions

### Supported Installation Paths

insta-infra automatically searches for Docker and Podman in these common locations:

#### Docker
- **macOS**: `/usr/local/bin/docker`, `/opt/homebrew/bin/docker`, `/Applications/Docker.app/Contents/Resources/bin/docker`
- **Linux**: `/usr/bin/docker`, `/usr/local/bin/docker`, `/opt/docker/bin/docker`, `/snap/bin/docker`, `/var/lib/flatpak/exports/bin/docker`
- **Windows**: `C:\Program Files\Docker\Docker\resources\bin\docker.exe`, `C:\ProgramData\chocolatey\bin\docker.exe`, `C:\tools\docker\docker.exe`

#### Podman
- **macOS**: `/usr/local/bin/podman`, `/opt/homebrew/bin/podman`
- **Linux**: `/usr/bin/podman`, `/usr/local/bin/podman`, `/opt/podman/bin/podman`, `/snap/bin/podman`, `/var/lib/flatpak/exports/bin/podman`
- **Windows**: `C:\Program Files\RedHat\Podman\podman.exe`, `C:\ProgramData\chocolatey\bin\podman.exe`, `C:\tools\podman\podman.exe`

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
│   └── insta/          # Web UI application (Browser-based)
│       ├── frontend/   # React frontend
│       │   ├── src/    # React components and styles
│       │   └── dist/   # Built frontend assets
│       ├── webserver.go # HTTP API server
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
2. Install dependencies: `cd cmd/insta/frontend && npm install`
3. Start development mode: `make dev-web`
4. Make changes to:
   - **Go backend**: `cmd/insta/webserver.go`
   - **React frontend**: `cmd/insta/frontend/src/`
   - **Shared logic**: `internal/core/`
5. Build for production: `make build-web`

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
