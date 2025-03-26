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
go install github.com/data-catering/insta-infra/cmd/insta@latest
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
│   └── insta/          # Main CLI application
│       ├── container/  # Container runtime implementations
│       ├── resources/  # Embedded resources
│       │   ├── data/   # Service configuration files
│       │   └── *.yaml  # Docker compose files
│       ├── models.go   # Service definitions
│       └── main.go     # CLI entry point
├── tests/              # Integration tests
├── Makefile            # Build and development tasks
└── README.md           # Documentation
```

### Development Workflow

1. Clone the repository
2. Make changes
3. Run tests: `make test`
4. Build: `make build`
5. Run: `./insta`

### Adding a New Service

1. Add service configuration to [`docker-compose.yaml`](cmd/insta/resources/docker-compose.yaml)
2. Add service definition to [`models.go`](cmd/insta/models.go)
3. Add any necessary initialization scripts to [`cmd/insta/resources/data/<service_name>/`](cmd/insta/resources/data/)
4. Update tests

## Services

| Service Type                | Service                   | Supported  |
|-----------------------------|---------------------------|------------|
| Api Gateway                 | kong                      | ✅         |
| Cache                       | redis                     | ✅         |
| Change Data Capture         | debezium                  | ✅         |
| Code Analysis               | sonarqube                 | ✅         |
| Database                    | cassandra                 | ✅         |
| Database                    | cockroachdb               | ✅         |
| Database                    | elasticsearch             | ✅         |
| Database                    | mariadb                   | ✅         |
| Database                    | mongodb                   | ✅         |
| Database                    | mssql                     | ✅         |
| Database                    | mysql                     | ✅         |
| Database                    | neo4j                     | ✅         |
| Database                    | opensearch                | ✅         |
| Database                    | postgres                  | ✅         |
| Database                    | spanner                   | ✅         |
| Database                    | sqlite                    | ✅         |
| Data Catalog                | amundsen                  | ✅         |
| Data Catalog                | datahub                   | ✅         |
| Data Catalog                | marquez                   | ✅         |
| Data Catalog                | openmetadata              | ✅         |
| Data Catalog                | polaris                   | ✅         |
| Data Catalog                | unitycatalog              | ✅         |
| Data Collector              | fluentd                   | ✅         |
| Data Collector              | logstash                  | ✅         |
| Data Visualisation          | blazer                    | ✅         |
| Data Visualisation          | evidence                  | ✅         |
| Data Visualisation          | metabase                  | ✅         |
| Data Visualisation          | redash                    | ✅         |
| Data Visualisation          | superset                  | ✅         |
| Distributed Coordination    | zookeeper                 | ✅         |
| Distributed Data Processing | flink                     | ✅         |
| Identity Management         | keycloak                  | ✅         |
| Job Orchestrator            | airflow                   | ✅         |
| Job Orchestrator            | dagster                   | ✅         |
| Job Orchestrator            | mage-ai                   | ✅         |
| Job Orchestrator            | prefect                   | ✅         |
| Messaging                   | activemq                  | ✅         |
| Messaging                   | kafka                     | ✅         |
| Messaging                   | rabbitmq                  | ✅         |
| Messaging                   | solace                    | ✅         |
| Notebook                    | jupyter                   | ✅         |
| Object Storage              | minio                     | ✅         |
| Query Engine                | duckdb                    | ✅         |
| Query Engine                | flight-sql                | ✅         |
| Query Engine                | presto                    | ✅         |
| Query Engine                | trino                     | ✅         |
| Real-time OLAP              | clickhouse                | ✅         |
| Real-time OLAP              | doris                     | ✅         |
| Real-time OLAP              | druid                     | ✅         |
| Real-time OLAP              | pinot                     | ✅         |
| Schema Registry             | confluent-schema-registry | ✅         |
| Test Data Management        | data-caterer              | ✅         |
| Web Server                  | httpbin                   | ✅         |
| Web Server                  | httpd                     | ✅         |
| Workflow                    | maestro                   | ✅         |
| Workflow                    | temporal                  | ✅         | 

## Updating

### Using the CLI

The easiest way to update is using the built-in update command:

```bash
insta -u
```

This will automatically:
1. Check for the latest version
2. Download the appropriate binary for your platform
3. Install the update
4. Create a backup of the old version

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
