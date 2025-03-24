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

### Start

```shell
./run.sh <services>
./run.sh postgres mysql
```

#### Example Output

```shell
How to connect:
Service   Container To Container  Host To Container  Container To Host
postgres  postgres:5432           localhost:5432     host.docker.internal:5432
mysql     mysql:3306              localhost:3306     host.docker.internal:3306
```

### Connect

```shell
./run.sh [connect|-c] <service>
./run.sh -c postgres
./run.sh connect postgres
```

### Shutdown

```shell
./run.sh [down|-d] <services>
./run.sh -d #bring all services down
./run.sh down postgres
```

### List supported services

```shell
./run.sh -l
./run.sh list
```

### Persist data

```shell
./run.sh -p postgres
```

### Remove persisted data

```shell
./run.sh [remove|-r] <services>
./run.sh -r #remove all service persisted data
./run.sh remove postgres
```

### Run version

```shell
<service>_VERSION=0.1 ./run.sh <services>
POSTGRES_VERSION=14.0 ./run.sh postgres
```

### Run from anywhere

In your `.bashrc, .zshrc, ...`, add:

```shell
alias insta=<checkout directory>/insta-infra/run.sh
```

Run `source ~/.bashrc` or `source ~/.zshrc` or open a new terminal session. Then you can run:

```shell
insta -l
insta postgres
insta -c postgres
insta -d
insta -r postgres
```

### Custom data

Alter data in [`data`](data) folder.
  
You may notice that for some services (such as Cassandra, Postgres, MySQL), they follow the same pattern for custom
data. They have a `data` directory which contains data files with DDL statements and an `init.sh` script that will help
execute them at startup. This allows you to dump all your `.sql` files into the directory, and it will be automatically
run at startup.


### Persisted data

If any data is persisted (via running with `-p`) from the services to carry across sessions, it gets pushed to folder:

`./data/<service>/persist`

### Authentication

By default, users and passwords follow what is default in the service. For those services where the user and password
can be altered at startup, it can be altered using environment variable pattern:
```shell
<service>_USER=...
<service>_PASSWORD=...
```

For example:
```shell
POSTGRES_USER=my-user POSTGRES_PASSWORD=my-password ./run.sh postgres
```

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

