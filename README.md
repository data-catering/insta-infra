# insta-infra

Spin up any tool or service straight away. Tells you how to connect to it.

- Single command
- Add custom data (i.e startup SQL scripts)
- Works anywhere
- Don't think about where to get image or version from
- Don't think about what hostname and port to use to connect

## How

```shell
./run.sh <services>
./run.sh postgres mysql
```

### Custom data

Alter data in [`data`](data) folder.

## Services

| Service Type     | Service       | Supported |
|------------------|---------------|-----------|
| Database         | cassandra     | ✅         |
| Database         | elasticsearch | ❌         |
| Database         | mariadb       | ❌         |
| Database         | mongodb       | ❌         |
| Database         | mysql         | ✅         |
| Database         | postgres      | ✅         |
| Data Catalog     | marquez       | ❌         |
| Data Catalog     | openmetadata  | ❌         |
| Job Orchestrator | airflow       | ❌         |
| Messaging        | kafka         | ✅         |
| Messaging        | solace        | ✅         |
