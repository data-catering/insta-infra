package main

// Service represents a supported service and its connection details
type Service struct {
	Name             string
	ConnectionCmd    string
	DefaultUser      string
	DefaultPassword  string
	RequiresPassword bool
}

// Services defines all supported services and their connection details
var Services = map[string]Service{
	"activemq": {
		Name:             "activemq",
		ConnectionCmd:    "/var/lib/artemis-instance/bin/artemis shell --user ${ARTEMIS_USER:-artemis} --password ${ARTEMIS_PASSWORD:-artemis}",
		DefaultUser:      "artemis",
		DefaultPassword:  "artemis",
		RequiresPassword: true,
	},
	"airflow": {
		Name:             "airflow",
		ConnectionCmd:    "airflow",
		DefaultUser:      "airflow",
		DefaultPassword:  "airflow",
		RequiresPassword: true,
	},
	"amundsen": {
		Name:          "amundsen",
		ConnectionCmd: "bash",
	},
	"blazer": {
		Name:          "blazer",
		ConnectionCmd: "bash",
	},
	"cassandra": {
		Name:          "cassandra",
		ConnectionCmd: "cqlsh",
	},
	"clickhouse": {
		Name:          "clickhouse",
		ConnectionCmd: "clickhouse-client",
	},
	"cockroachdb": {
		Name:          "cockroachdb",
		ConnectionCmd: "./cockroach sql --insecure",
	},
	"datahub": {
		Name:             "datahub",
		ConnectionCmd:    "bash",
		RequiresPassword: false,
	},
	"debezium": {
		Name:          "debezium",
		ConnectionCmd: "bash",
	},
	"doris": {
		Name:          "doris",
		ConnectionCmd: "mysql -uroot -P9030 -h127.0.0.1",
	},
	"druid": {
		Name:          "druid",
		ConnectionCmd: "bash",
	},
	"duckdb": {
		Name:          "duckdb",
		ConnectionCmd: "./duckdb",
	},
	"elasticsearch": {
		Name:             "elasticsearch",
		ConnectionCmd:    "elasticsearch-sql-cli http://elastic:${ELASTICSEARCH_PASSWORD:-elasticsearch}@localhost:9200",
		DefaultUser:      "elastic",
		DefaultPassword:  "elasticsearch",
		RequiresPassword: true,
	},
	"evidence": {
		Name:          "evidence",
		ConnectionCmd: "bash",
	},
	"feast": {
		Name:          "feast",
		ConnectionCmd: "bash",
	},
	"flight-sql": {
		Name:             "flight-sql",
		ConnectionCmd:    "flight_sql_client --command Execute --host localhost --port 31337 --use-tls --tls-skip-verify --username ${FLIGHT_SQL_USER:-flight_username} --password ${FLIGHT_SQL_PASSWORD:-flight_password} --query 'SELECT version()'",
		DefaultUser:      "flight_username",
		DefaultPassword:  "flight_password",
		RequiresPassword: true,
	},
	"flink": {
		Name:          "flink",
		ConnectionCmd: "bash",
	},
	"flink-jobmanager": {
		Name:          "flink-jobmanager",
		ConnectionCmd: "bash",
	},
	"grafana": {
		Name:             "grafana",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPassword:  "admin",
		RequiresPassword: true,
	},
	"influxdb": {
		Name:             "influxdb",
		ConnectionCmd:    "influx",
		DefaultUser:      "admin",
		DefaultPassword:  "admin",
		RequiresPassword: true,
	},
	"istio": {
		Name:          "istio",
		ConnectionCmd: "istioctl proxy-status",
	},
	"jaeger": {
		Name:          "jaeger",
		ConnectionCmd: "bash",
	},
	"jenkins": {
		Name:             "jenkins",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		RequiresPassword: true,
	},
	"kafka": {
		Name:          "kafka",
		ConnectionCmd: "kafka-topics --bootstrap-server localhost:9092 --list",
	},
	"keycloak": {
		Name:             "keycloak",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPassword:  "admin",
		RequiresPassword: true,
	},
	"kibana": {
		Name:             "kibana",
		ConnectionCmd:    "bash",
		DefaultUser:      "kibana_system",
		DefaultPassword:  "password",
		RequiresPassword: true,
	},
	"kong": {
		Name:          "kong",
		ConnectionCmd: "bash",
	},
	"lakekeeper": {
		Name:            "lakekeeper",
		ConnectionCmd:   "bash",
		DefaultUser:     "peter",
		DefaultPassword: "iceberg",
	},
	"lakekeeper-jupyter": {
		Name:          "lakekeeper-jupyter",
		ConnectionCmd: "bash",
	},
	"logstash": {
		Name:             "logstash",
		ConnectionCmd:    "bash",
		DefaultUser:      "logstash_internal",
		DefaultPassword:  "password",
		RequiresPassword: true,
	},
	"loki": {
		Name:          "loki",
		ConnectionCmd: "bash",
	},
	"maestro": {
		Name:          "maestro",
		ConnectionCmd: "bash",
	},
	"mage-ai": {
		Name:          "mage-ai",
		ConnectionCmd: "bash",
	},
	"mariadb": {
		Name:             "mariadb",
		ConnectionCmd:    "mariadb --user=${MARIADB_USER:-user} --password=${MARIADB_PASSWORD:-password}",
		DefaultUser:      "user",
		DefaultPassword:  "password",
		RequiresPassword: true,
	},
	"marquez": {
		Name:          "marquez",
		ConnectionCmd: "bash",
	},
	"memcached": {
		Name:          "memcached",
		ConnectionCmd: "bash",
	},
	"metabase": {
		Name:          "metabase",
		ConnectionCmd: "bash",
	},
	"milvus": {
		Name:          "milvus",
		ConnectionCmd: "bash",
	},
	"minio": {
		Name:             "minio",
		ConnectionCmd:    "mc alias set local http://localhost:9000 ${MINIO_USER:-minioadmin} ${MINIO_PASSWORD:-minioadmin}",
		DefaultUser:      "minioadmin",
		DefaultPassword:  "minioadmin",
		RequiresPassword: true,
	},
	"mlflow": {
		Name:          "mlflow",
		ConnectionCmd: "bash",
	},
	"mlflow-serve": {
		Name:          "mlflow-serve",
		ConnectionCmd: "bash",
	},
	"mongodb": {
		Name:             "mongodb",
		ConnectionCmd:    "mongosh mongodb://${MONGODB_USER:-root}:${MONGODB_PASSWORD:-root}@mongodb",
		DefaultUser:      "root",
		DefaultPassword:  "root",
		RequiresPassword: true,
	},
	"mssql": {
		Name:             "mssql",
		ConnectionCmd:    "/opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P \"${MSSQL_PASSWORD:-yourStrong(!)Password}\"",
		DefaultUser:      "sa",
		DefaultPassword:  "yourStrong(!)Password",
		RequiresPassword: true,
	},
	"mysql": {
		Name:             "mysql",
		ConnectionCmd:    "mysql -u ${MYSQL_USER:-root} -p${MYSQL_PASSWORD:-root}",
		DefaultUser:      "root",
		DefaultPassword:  "root",
		RequiresPassword: true,
	},
	"nats": {
		Name:          "nats",
		ConnectionCmd: "nats pub test.subject 'Hello NATS!'",
	},
	"neo4j": {
		Name:             "neo4j",
		ConnectionCmd:    "cypher-shell -u neo4j -p test",
		DefaultUser:      "neo4j",
		DefaultPassword:  "test",
		RequiresPassword: true,
	},
	"openmetadata": {
		Name:             "openmetadata",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPassword:  "admin",
		RequiresPassword: true,
	},
	"opensearch": {
		Name:             "opensearch",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPassword:  "!BigData#1",
		RequiresPassword: true,
	},
	"pinot": {
		Name:          "pinot",
		ConnectionCmd: "bash",
	},
	"polaris": {
		Name:          "polaris",
		ConnectionCmd: "bash",
	},
	"postgres": {
		Name:             "postgres",
		ConnectionCmd:    "PGPASSWORD=${POSTGRES_PASSWORD:-postgres} psql -U${POSTGRES_USER:-postgres}",
		DefaultUser:      "postgres",
		DefaultPassword:  "postgres",
		RequiresPassword: true,
	},
	"prefect-data": {
		Name:          "prefect-data",
		ConnectionCmd: "bash",
	},
	"presto": {
		Name:          "presto",
		ConnectionCmd: "presto-cli",
	},
	"prometheus": {
		Name:          "prometheus",
		ConnectionCmd: "promtool query instant http://localhost:9090 up",
	},
	"pulsar": {
		Name:          "pulsar",
		ConnectionCmd: "pulsar-client produce test-topic -m 'Hello Pulsar!'",
	},
	"qdrant": {
		Name:          "qdrant",
		ConnectionCmd: "bash",
	},
	"rabbitmq": {
		Name:             "rabbitmq",
		ConnectionCmd:    "rabbitmqctl --host localhost --username ${RABBITMQ_USER:-guest} --password ${RABBITMQ_PASSWORD:-guest} list_queues",
		DefaultUser:      "guest",
		DefaultPassword:  "guest",
		RequiresPassword: true,
	},
	"ray": {
		Name:          "ray",
		ConnectionCmd: "bash",
	},
	"redash": {
		Name:          "redash",
		ConnectionCmd: "bash",
	},
	"redis": {
		Name:          "redis",
		ConnectionCmd: "redis-cli",
	},
	"solace": {
		Name:             "solace",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPassword:  "admin",
		RequiresPassword: true,
	},
	"sonarqube": {
		Name:             "sonarqube",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPassword:  "admin",
		RequiresPassword: true,
	},
	"spanner": {
		Name:          "spanner",
		ConnectionCmd: "bash",
	},
	"superset": {
		Name:             "superset",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPassword:  "admin",
		RequiresPassword: true,
	},
	"temporal": {
		Name:          "temporal",
		ConnectionCmd: "temporal workflow list",
	},
	"timescaledb": {
		Name:             "timescaledb",
		ConnectionCmd:    "PGPASSWORD=${TIMESCALEDB_PASSWORD:-postgres} psql -U${TIMESCALEDB_USER:-postgres}",
		DefaultUser:      "postgres",
		DefaultPassword:  "postgres",
		RequiresPassword: true,
	},
	"traefik": {
		Name:          "traefik",
		ConnectionCmd: "bash",
	},
	"trino": {
		Name:          "trino",
		ConnectionCmd: "trino",
	},
	"unitycatalog": {
		Name:          "unitycatalog",
		ConnectionCmd: "bash",
	},
	"vault": {
		Name:             "vault",
		ConnectionCmd:    "vault status",
		DefaultUser:      "root",
		RequiresPassword: true,
	},
	"weaviate": {
		Name:          "weaviate",
		ConnectionCmd: "bash",
	},
	"zookeeper": {
		Name:          "zookeeper",
		ConnectionCmd: "zkCli.sh",
	},
}
