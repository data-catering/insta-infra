package main

// Service represents a supported service and its connection details
type Service struct {
	Name             string
	ConnectionCmd    string
	DefaultUser      string
	DefaultPassword  string
	DefaultPort      int
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
		DefaultPort:      8080,
		RequiresPassword: true,
	},
	"amundsen": {
		Name:          "amundsen",
		ConnectionCmd: "bash",
		DefaultPort:   5003,
	},
	"blazer": {
		Name:          "blazer",
		ConnectionCmd: "bash",
		DefaultPort:   8080,
	},
	"cassandra": {
		Name:          "cassandra",
		ConnectionCmd: "cqlsh",
		DefaultPort:   9042,
	},
	"clickhouse": {
		Name:          "clickhouse",
		ConnectionCmd: "clickhouse-client",
		DefaultPort:   9000,
	},
	"cockroachdb": {
		Name:          "cockroachdb",
		ConnectionCmd: "./cockroach sql --insecure",
		DefaultPort:   26257,
	},
	"datahub": {
		Name:             "datahub",
		ConnectionCmd:    "bash",
		DefaultPort:      9002,
		RequiresPassword: false,
	},
	"debezium": {
		Name:          "debezium",
		ConnectionCmd: "bash",
		DefaultPort:   8080,
	},
	"doris": {
		Name:          "doris",
		ConnectionCmd: "mysql -uroot -P9030 -h127.0.0.1",
		DefaultPort:   9030,
	},
	"druid": {
		Name:          "druid",
		ConnectionCmd: "bash",
		DefaultPort:   8888,
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
		DefaultPort:      9200,
		RequiresPassword: true,
	},
	"evidence": {
		Name:          "evidence",
		ConnectionCmd: "bash",
		DefaultPort:   3000,
	},
	"flight-sql": {
		Name:             "flight-sql",
		ConnectionCmd:    "flight_sql_client --command Execute --host localhost --port 31337 --use-tls --tls-skip-verify --username ${FLIGHT_SQL_USER:-flight_username} --password ${FLIGHT_SQL_PASSWORD:-flight_password} --query 'SELECT version()'",
		DefaultUser:      "flight_username",
		DefaultPassword:  "flight_password",
		DefaultPort:      31337,
		RequiresPassword: true,
	},
	"kafka": {
		Name:          "kafka",
		ConnectionCmd: "kafka-topics --bootstrap-server localhost:9092 --list",
		DefaultPort:   9092,
	},
	"keycloak": {
		Name:             "keycloak",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPassword:  "admin",
		DefaultPort:      8082,
		RequiresPassword: true,
	},
	"kibana": {
		Name:             "kibana",
		ConnectionCmd:    "bash",
		DefaultUser:      "kibana_system",
		DefaultPassword:  "password",
		DefaultPort:      5601,
		RequiresPassword: true,
	},
	"kong": {
		Name:          "kong",
		ConnectionCmd: "bash",
		DefaultPort:   8001,
	},
	"logstash": {
		Name:             "logstash",
		ConnectionCmd:    "bash",
		DefaultUser:      "logstash_internal",
		DefaultPassword:  "password",
		DefaultPort:      9600,
		RequiresPassword: true,
	},
	"maestro": {
		Name:          "maestro",
		ConnectionCmd: "bash",
		DefaultPort:   8081,
	},
	"mage-ai": {
		Name:          "mage-ai",
		ConnectionCmd: "bash",
		DefaultPort:   6789,
	},
	"mariadb": {
		Name:             "mariadb",
		ConnectionCmd:    "mariadb --user=${MARIADB_USER:-user} --password=${MARIADB_PASSWORD:-password}",
		DefaultUser:      "user",
		DefaultPassword:  "password",
		DefaultPort:      3306,
		RequiresPassword: true,
	},
	"marquez": {
		Name:          "marquez",
		ConnectionCmd: "bash",
		DefaultPort:   5002,
	},
	"metabase": {
		Name:          "metabase",
		ConnectionCmd: "bash",
		DefaultPort:   3000,
	},
	"minio": {
		Name:             "minio",
		ConnectionCmd:    "mc alias set local http://localhost:9000 ${MINIO_USER:-minioadmin} ${MINIO_PASSWORD:-minioadmin}",
		DefaultUser:      "minioadmin",
		DefaultPassword:  "minioadmin",
		DefaultPort:      9000,
		RequiresPassword: true,
	},
	"mongodb": {
		Name:             "mongodb",
		ConnectionCmd:    "mongosh mongodb://${MONGODB_USER:-root}:${MONGODB_PASSWORD:-root}@mongodb",
		DefaultUser:      "root",
		DefaultPassword:  "root",
		DefaultPort:      27017,
		RequiresPassword: true,
	},
	"mysql": {
		Name:             "mysql",
		ConnectionCmd:    "mysql -u ${MYSQL_USER:-root} -p${MYSQL_PASSWORD:-root}",
		DefaultUser:      "root",
		DefaultPassword:  "root",
		DefaultPort:      3306,
		RequiresPassword: true,
	},
	"mssql": {
		Name:             "mssql",
		ConnectionCmd:    "/opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P \"${MSSQL_PASSWORD:-yourStrong(!)Password}\"",
		DefaultUser:      "sa",
		DefaultPassword:  "yourStrong(!)Password",
		DefaultPort:      1433,
		RequiresPassword: true,
	},
	"neo4j": {
		Name:             "neo4j",
		ConnectionCmd:    "cypher-shell -u neo4j -p test",
		DefaultUser:      "neo4j",
		DefaultPassword:  "test",
		DefaultPort:      7687,
		RequiresPassword: true,
	},
	"openmetadata": {
		Name:             "openmetadata",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPassword:  "admin",
		DefaultPort:      8585,
		RequiresPassword: true,
	},
	"opensearch": {
		Name:             "opensearch",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPassword:  "!BigData#1",
		DefaultPort:      9200,
		RequiresPassword: true,
	},
	"pinot": {
		Name:          "pinot",
		ConnectionCmd: "bash",
		DefaultPort:   9000,
	},
	"polaris": {
		Name:          "polaris",
		ConnectionCmd: "bash",
		DefaultPort:   8181,
	},
	"postgres": {
		Name:             "postgres",
		ConnectionCmd:    "PGPASSWORD=${POSTGRES_PASSWORD:-postgres} psql -U${POSTGRES_USER:-postgres}",
		DefaultUser:      "postgres",
		DefaultPassword:  "postgres",
		DefaultPort:      5432,
		RequiresPassword: true,
	},
	"prefect-data": {
		Name:          "prefect-data",
		ConnectionCmd: "bash",
	},
	"presto": {
		Name:          "presto",
		ConnectionCmd: "presto-cli",
		DefaultPort:   8080,
	},
	"rabbitmq": {
		Name:             "rabbitmq",
		ConnectionCmd:    "rabbitmqctl --host localhost --username ${RABBITMQ_USER:-guest} --password ${RABBITMQ_PASSWORD:-guest} list_queues",
		DefaultUser:      "guest",
		DefaultPassword:  "guest",
		DefaultPort:      5672,
		RequiresPassword: true,
	},
	"redash": {
		Name:          "redash",
		ConnectionCmd: "bash",
		DefaultPort:   5000,
	},
	"redis": {
		Name:          "redis",
		ConnectionCmd: "redis-cli",
		DefaultPort:   6379,
	},
	"solace": {
		Name:             "solace",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPassword:  "admin",
		DefaultPort:      8080,
		RequiresPassword: true,
	},
	"sonarqube": {
		Name:             "sonarqube",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPassword:  "admin",
		DefaultPort:      9000,
		RequiresPassword: true,
	},
	"spanner": {
		Name:          "spanner",
		ConnectionCmd: "bash",
		DefaultPort:   9020,
	},
	"superset": {
		Name:             "superset",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPassword:  "admin",
		DefaultPort:      8088,
		RequiresPassword: true,
	},
	"temporal": {
		Name:          "temporal",
		ConnectionCmd: "temporal workflow list",
		DefaultPort:   7233,
	},
	"trino": {
		Name:          "trino",
		ConnectionCmd: "trino",
		DefaultPort:   8080,
	},
	"unitycatalog": {
		Name:          "unitycatalog",
		ConnectionCmd: "bash",
	},
	"flink": {
		Name:          "flink",
		ConnectionCmd: "bash",
	},
	"flink-jobmanager": {
		Name:          "flink-jobmanager",
		ConnectionCmd: "bash",
	},
	"zookeeper": {
		Name:          "zookeeper",
		ConnectionCmd: "zkCli.sh",
		DefaultPort:   2181,
	},
	"prometheus": {
		Name:          "prometheus",
		ConnectionCmd: "promtool query instant http://localhost:9090 up",
		DefaultPort:   9090,
	},
	"grafana": {
		Name:             "grafana",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPassword:  "admin",
		DefaultPort:      3000,
		RequiresPassword: true,
	},
	"jaeger": {
		Name:          "jaeger",
		ConnectionCmd: "bash",
		DefaultPort:   16686,
	},
	"vault": {
		Name:             "vault",
		ConnectionCmd:    "vault status",
		DefaultUser:      "root",
		DefaultPort:      8200,
		RequiresPassword: true,
	},
	"nats": {
		Name:          "nats",
		ConnectionCmd: "nats pub test.subject 'Hello NATS!'",
		DefaultPort:   4222,
	},
	"pulsar": {
		Name:          "pulsar",
		ConnectionCmd: "pulsar-client produce test-topic -m 'Hello Pulsar!'",
		DefaultPort:   6650,
	},
	"memcached": {
		Name:          "memcached",
		ConnectionCmd: "bash",
		DefaultPort:   11211,
	},
	"influxdb": {
		Name:             "influxdb",
		ConnectionCmd:    "influx",
		DefaultUser:      "admin",
		DefaultPassword:  "admin",
		DefaultPort:      8086,
		RequiresPassword: true,
	},
	"jenkins": {
		Name:             "jenkins",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPort:      8080,
		RequiresPassword: true,
	},
	"istio": {
		Name:          "istio",
		ConnectionCmd: "istioctl proxy-status",
		DefaultPort:   15010,
	},
	"mlflow": {
		Name:          "mlflow",
		ConnectionCmd: "bash",
		DefaultPort:   5000,
	},
	"milvus": {
		Name:          "milvus",
		ConnectionCmd: "bash",
		DefaultPort:   19530,
	},
	"loki": {
		Name:          "loki",
		ConnectionCmd: "bash",
		DefaultPort:   3100,
	},
	"traefik": {
		Name:          "traefik",
		ConnectionCmd: "bash",
		DefaultPort:   8080,
	},
	"feast": {
		Name:          "feast",
		ConnectionCmd: "feast serve -h 0.0.0.0",
		DefaultPort:   6566,
	},
	"timescaledb": {
		Name:             "timescaledb",
		ConnectionCmd:    "PGPASSWORD=${TIMESCALEDB_PASSWORD:-postgres} psql -U${TIMESCALEDB_USER:-postgres}",
		DefaultUser:      "postgres",
		DefaultPassword:  "postgres",
		DefaultPort:      5432,
		RequiresPassword: true,
	},
	"weaviate": {
		Name:          "weaviate",
		ConnectionCmd: "bash",
		DefaultPort:   8080,
	},
	"qdrant": {
		Name:          "qdrant",
		ConnectionCmd: "bash",
		DefaultPort:   6333,
	},
	"ray": {
		Name:          "ray",
		ConnectionCmd: "ray status",
		DefaultPort:   8265,
	},
}
