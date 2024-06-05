# insta-infra

Spin up any tool or service straight away on your local laptop. Tells you how to connect to it.

- Simple commands
- Add custom data (i.e startup SQL scripts)
- Works anywhere
- Don't worry about startup configuration
- Don't think about what hostname, port, or credentials to use to connect

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
./connect.sh <service>
./connect.sh postgres
```

### Shutdown

```shell
docker-compose down
```

### Custom data

Alter data in [`data`](data) folder.

### Persisted data

If any data is persisted from the services to carry across sessions, it gets pushed to folder:

`${HOME}/data/<service>`

## Services

| Service Type         | Service       | Supported |
|----------------------|---------------|-----------|
| Database             | cassandra     | ✅         |
| Database             | elasticsearch | ✅         |
| Database             | mariadb       | ✅         |
| Database             | mongodb       | ✅         |
| Database             | mysql         | ✅         |
| Database             | postgres      | ✅         |
| Data Catalog         | marquez       | ✅         |
| Data Catalog         | amundsen      | ❌         |
| Data Catalog         | datahub       | ❌         |
| Data Catalog         | openmetadata  | ❌         |
| HTTP                 | httpbin       | ✅         |
| Identity Management  | keycloak      | ✅         |
| Job Orchestrator     | airflow       | ✅         |
| Job Orchestrator     | dagster       | ✅         |
| Job Orchestrator     | mage-ai       | ✅         |
| Job Orchestrator     | prefect       | ✅         |
| Messaging            | kafka         | ✅         |
| Messaging            | solace        | ✅         |
| Messaging            | activemq      | ❌         |
| Messaging            | rabbitmq      | ❌         |
| Object Storage       | minio         | ✅         |
| Query Engine         | duckdb        | ❌         |
| Query Engine         | presto        | ✅         |
| Query Engine         | trino         | ✅         |
| Real-time OLAP       | clickhouse    | ❌         |
| Real-time OLAP       | doris         | ❌         |
| Real-time OLAP       | druid         | ❌         |
| Real-time OLAP       | pinot         | ❌         |
| Test Data Management | data-caterer  | ✅         |
