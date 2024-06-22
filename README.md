# insta-infra

Spin up any service straight away on your local laptop. Tells you how to connect to it.

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

### Remove persisted data

```shell
./run.sh [remove|-r] <services>
./run.sh -r #remove all service persisted data
./run.sh remove postgres
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

If any data is persisted from the services to carry across sessions, it gets pushed to folder:

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

| Service Type                | Service       | Supported |
|-----------------------------|---------------|-----------|
| Change Data Capture         | debezium      | ✅         |
| Database                    | cassandra     | ✅         |
| Database                    | cockroachdb   | ✅         |
| Database                    | elasticsearch | ✅         |
| Database                    | mariadb       | ✅         |
| Database                    | mongodb       | ✅         |
| Database                    | mysql         | ✅         |
| Database                    | neo4j         | ✅         |
| Database                    | postgres      | ✅         |
| Database                    | spanner       | ✅         |
| Database                    | sqlite        | ✅         |
| Database                    | opensearch    | ❌         |
| Data Catalog                | marquez       | ✅         |
| Data Catalog                | unitycatalog  | ✅         |
| Data Catalog                | amundsen      | ❌         |
| Data Catalog                | datahub       | ❌         |
| Data Catalog                | openmetadata  | ❌         |
| Distributed Coordination    | zookeeper     | ✅         |
| Distributed Data Processing | flink         | ✅         |
| HTTP                        | httpbin       | ✅         |
| Identity Management         | keycloak      | ✅         |
| Job Orchestrator            | airflow       | ✅         |
| Job Orchestrator            | dagster       | ✅         |
| Job Orchestrator            | mage-ai       | ✅         |
| Job Orchestrator            | mlflow        | ✅         |
| Job Orchestrator            | prefect       | ✅         |
| Messaging                   | activemq      | ✅         |
| Messaging                   | kafka         | ✅         |
| Messaging                   | rabbitmq      | ✅         |
| Messaging                   | solace        | ✅         |
| Object Storage              | minio         | ✅         |
| Query Engine                | duckdb        | ✅         |
| Query Engine                | flight-sql    | ✅         |
| Query Engine                | presto        | ✅         |
| Query Engine                | trino         | ✅         |
| Real-time OLAP              | clickhouse    | ✅         |
| Real-time OLAP              | doris         | ✅         |
| Real-time OLAP              | druid         | ✅         |
| Real-time OLAP              | pinot         | ✅         |
| Test Data Management        | data-caterer  | ✅         |
| Workflow                    | temporal      | ✅         |
