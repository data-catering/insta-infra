# insta-infra

![insta-infra services](docs/img/insta-infra.gif)

Spin up any service straight away on your local laptop. Tells you how to connect to it.

- Simple commands
- Add custom data (i.e startup SQL scripts)
- Works anywhere
- Don't worry about startup configuration
- Don't think about what hostname, port, or credentials to use to connect

## Installation

### Prerequisites
- Docker and Docker Compose

### Option 1: Quick Install (Recommended)
```shell
curl -fsSL https://raw.githubusercontent.com/data-catering/insta-infra/main/install.sh | bash
```
This will install insta-infra to `~/.insta-infra` and create a symbolic link in `/usr/local/bin` if possible.

### Option 2: With npm
```shell
npm install -g insta-infra
```
This will install the `insta` command globally on your system.

### Option 3: With Homebrew (macOS)
```shell
brew tap data-catering/insta-infra
brew install insta-infra
```

### Option 4: Using Docker
```shell
docker run -it --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $PWD/data:/app/data \
  data-catering/insta-infra postgres
```
Or create an alias in your shell configuration:
```shell
alias insta='docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock -v $PWD/data:/app/data data-catering/insta-infra'
```

### Option 5: Manual Install
```shell
# Clone the repository
git clone https://github.com/data-catering/insta-infra.git
cd insta-infra
chmod +x run.sh
```

### Option 6: As a Shell Alias (Any OS)
In your `.bashrc, .zshrc, ...`, add:
```shell
alias insta=/path/to/insta-infra/run.sh
```
Then run `source ~/.bashrc` or `source ~/.zshrc` or open a new terminal session.

## How

After installation, you can use insta-infra with the `insta` command (or `./run.sh` if manually installed).

### Basic Commands

```shell
# List available services
insta -l

# Start a service
insta postgres                  # Start PostgreSQL
insta mysql redis               # Start multiple services

# Connect to a service
insta -c postgres              # Connect to PostgreSQL

# Stop services
insta -d                       # Stop all services
insta -d postgres              # Stop specific service

# Run with persisted data
insta -p postgres             # Data will persist across restarts

# Remove persisted data
insta -r                      # Remove all persisted data
insta -r postgres             # Remove specific service data
```

### Using with Docker Installation

If you're using the Docker installation method, prefix your commands with `docker run`:

```shell
docker run -it --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $PWD/data:/app/data \
  data-catering/insta-infra -l  # List services

# Or use the alias if you set it up:
insta -l
```

### Custom Data

You can add custom initialization data for services in the `data` directory:

```
data/
├── postgres/
│   ├── init.sql
│   └── data.sql
├── mysql/
│   └── init.sql
└── ...
```

These files will be automatically executed when the service starts.

### Authentication

By default, services use their standard authentication. You can override these using environment variables:

```shell
# Example: Custom PostgreSQL credentials
POSTGRES_USER=my-user POSTGRES_PASSWORD=my-password insta postgres

# Example: Custom MySQL credentials
MYSQL_USER=my-user MYSQL_PASSWORD=my-password insta mysql
```

### Version Selection

You can specify a particular version of a service:

```shell
POSTGRES_VERSION=14.0 insta postgres
MYSQL_VERSION=8.0 insta mysql
```

### Data Persistence

Data can be persisted to the host machine using the `-p` flag:

```shell
insta -p postgres              # Data will be saved in data/postgres/persist/
```

The data will survive container restarts and removals.

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

## Testing

The project includes comprehensive tests to ensure reliability and correctness. To run the tests:

```shell
# Run all unit tests
./tests/run.sh

# Run integration tests (requires Docker)
./tests/test_integration.sh
```

### Test Categories

1. **Core Tests**: Validate the main script functionality
2. **Docker Compose Tests**: Ensure docker-compose files are valid
3. **Installation Tests**: Verify the installation script works properly
4. **Package Tests**: Check npm package configuration
5. **Integration Tests**: Start and stop a real service to ensure everything works end-to-end

### Continuous Integration

Tests are automatically run on GitHub Actions for every pull request and push to main branch.

