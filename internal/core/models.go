package core

// PortType represents the type of port mapping for services
type PortType string

const (
	PortTypeWebUI     PortType = "WEB_UI"    // Web user interface
	PortTypeAPI       PortType = "API"       // REST/GraphQL API
	PortTypeDatabase  PortType = "DATABASE"  // Database connection
	PortTypeAdmin     PortType = "ADMIN"     // Admin interface
	PortTypeMetrics   PortType = "METRICS"   // Metrics/monitoring
	PortTypeMessaging PortType = "MESSAGING" // Message queue/broker
	PortTypeOther     PortType = "OTHER"     // Other/unknown
)

// ServicePort defines a known port for a service
type ServicePort struct {
	InternalPort int      `json:"internal_port"`           // Port inside the container
	Type         PortType `json:"type"`                    // Type of port
	Name         string   `json:"name"`                    // Human-readable name
	Description  string   `json:"description"`             // What this port provides
	Path         string   `json:"path,omitempty"`          // Default path for web interfaces
	RequiresAuth bool     `json:"requires_auth,omitempty"` // Whether authentication is required
}

// Service represents a supported service and its connection details
type Service struct {
	Name             string        `json:"name"`
	Type             string        `json:"type"`
	ConnectionCmd    string        `json:"connection_cmd"`
	DefaultUser      string        `json:"default_user,omitempty"`
	DefaultPassword  string        `json:"default_password,omitempty"`
	RequiresPassword bool          `json:"requires_password"`
	Ports            []ServicePort `json:"ports"` // Known ports for this service
}

// Services defines all supported services and their connection details by container name
var Services = map[string]Service{
	"activemq": {
		Name:             "activemq",
		Type:             "Messaging",
		ConnectionCmd:    "/var/lib/artemis-instance/bin/artemis shell --user ${ARTEMIS_USER:-artemis} --password ${ARTEMIS_PASSWORD:-artemis}",
		DefaultUser:      "artemis",
		DefaultPassword:  "artemis",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 8161, Type: PortTypeWebUI, Name: "Web Console", Description: "ActiveMQ Web Console", Path: "/", RequiresAuth: true},
			{InternalPort: 61616, Type: PortTypeMessaging, Name: "JMS Port", Description: "JMS messaging port"},
		},
	},
	"airflow": {
		Name:             "airflow",
		Type:             "Job Orchestrator",
		ConnectionCmd:    "airflow",
		DefaultUser:      "airflow",
		DefaultPassword:  "airflow",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 8080, Type: PortTypeWebUI, Name: "Web UI", Description: "Airflow Web Interface", Path: "/", RequiresAuth: true},
		},
	},
	"amundsen": {
		Name:          "amundsen",
		Type:          "Data Catalog",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 5000, Type: PortTypeWebUI, Name: "Web UI", Description: "Amundsen Data Catalog Interface", Path: "/"},
		},
	},
	"argilla": {
		Name:             "argilla",
		Type:             "Data Annotation",
		ConnectionCmd:    "bash",
		DefaultUser:      "argilla",
		DefaultPassword:  "12345678",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 6900, Type: PortTypeWebUI, Name: "Web UI", Description: "Argilla Annotation Interface", Path: "/", RequiresAuth: true},
		},
	},
	"blazer": {
		Name:          "blazer",
		Type:          "Data Visualisation",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 8080, Type: PortTypeWebUI, Name: "Web UI", Description: "Blazer SQL Explorer", Path: "/"},
		},
	},
	"cassandra": {
		Name:          "cassandra",
		Type:          "Database",
		ConnectionCmd: "cqlsh",
		Ports: []ServicePort{
			{InternalPort: 9042, Type: PortTypeDatabase, Name: "CQL Port", Description: "Cassandra Query Language port"},
			{InternalPort: 7000, Type: PortTypeOther, Name: "Inter-node", Description: "Inter-node communication"},
		},
	},
	"clickhouse": {
		Name:          "clickhouse",
		Type:          "Real-time OLAP",
		ConnectionCmd: "clickhouse-client",
		Ports: []ServicePort{
			{InternalPort: 8123, Type: PortTypeAPI, Name: "HTTP API", Description: "ClickHouse HTTP interface", Path: "/"},
			{InternalPort: 9000, Type: PortTypeDatabase, Name: "Native TCP", Description: "ClickHouse native TCP protocol"},
		},
	},
	"cockroachdb": {
		Name:          "cockroachdb",
		Type:          "Database",
		ConnectionCmd: "./cockroach sql --insecure",
		Ports: []ServicePort{
			{InternalPort: 26257, Type: PortTypeDatabase, Name: "SQL Port", Description: "CockroachDB SQL interface"},
			{InternalPort: 8080, Type: PortTypeWebUI, Name: "Admin UI", Description: "CockroachDB Admin Console", Path: "/"},
		},
	},
	"cvat": {
		Name:             "cvat",
		Type:             "Data Annotation",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPassword:  "admin",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 8080, Type: PortTypeWebUI, Name: "Web UI", Description: "CVAT Annotation Interface", Path: "/", RequiresAuth: true},
		},
	},
	"datahub": {
		Name:             "datahub",
		Type:             "Data Catalog",
		ConnectionCmd:    "bash",
		RequiresPassword: false,
		Ports: []ServicePort{
			{InternalPort: 9002, Type: PortTypeWebUI, Name: "Web UI", Description: "DataHub Data Catalog", Path: "/"},
		},
	},
	"debezium": {
		Name:          "debezium",
		Type:          "Change Data Capture",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 8083, Type: PortTypeAPI, Name: "REST API", Description: "Debezium Connect REST API", Path: "/"},
		},
	},
	"doccano": {
		Name:             "doccano",
		Type:             "Data Annotation",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPassword:  "admin",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 8000, Type: PortTypeWebUI, Name: "Web UI", Description: "Doccano Annotation Interface", Path: "/", RequiresAuth: true},
		},
	},
	"doris": {
		Name:          "doris",
		Type:          "Real-time OLAP",
		ConnectionCmd: "mysql -uroot -P9030 -h127.0.0.1",
		Ports: []ServicePort{
			{InternalPort: 8030, Type: PortTypeWebUI, Name: "Web UI", Description: "Doris Frontend Web Interface", Path: "/"},
			{InternalPort: 9030, Type: PortTypeDatabase, Name: "MySQL Protocol", Description: "MySQL-compatible query port"},
		},
	},
	"druid": {
		Name:          "druid",
		Type:          "Real-time OLAP",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 8888, Type: PortTypeWebUI, Name: "Router UI", Description: "Druid Router Web Console", Path: "/"},
			{InternalPort: 8081, Type: PortTypeWebUI, Name: "Coordinator UI", Description: "Druid Coordinator Console", Path: "/"},
			{InternalPort: 8082, Type: PortTypeAPI, Name: "Broker API", Description: "Druid Broker Query API", Path: "/"},
		},
	},
	"duckdb": {
		Name:          "duckdb",
		Type:          "Query Engine",
		ConnectionCmd: "./duckdb",
		Ports:         []ServicePort{}, // DuckDB typically runs as embedded, no network ports
	},
	"elasticsearch": {
		Name:             "elasticsearch",
		Type:             "Database",
		ConnectionCmd:    "elasticsearch-sql-cli http://elastic:${ELASTICSEARCH_PASSWORD:-elasticsearch}@localhost:9200",
		DefaultUser:      "elastic",
		DefaultPassword:  "elasticsearch",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 9200, Type: PortTypeAPI, Name: "REST API", Description: "Elasticsearch REST API", Path: "/"},
			{InternalPort: 9300, Type: PortTypeOther, Name: "Transport", Description: "Elasticsearch transport port"},
		},
	},
	"evidence": {
		Name:          "evidence",
		Type:          "Data Visualisation",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 3000, Type: PortTypeWebUI, Name: "Web UI", Description: "Evidence BI Interface", Path: "/"},
		},
	},
	"feast": {
		Name:          "feast",
		Type:          "Feature Store",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 6566, Type: PortTypeAPI, Name: "Feature Server", Description: "Feast Feature Server API", Path: "/"},
		},
	},
	"flight-sql": {
		Name:             "flight-sql",
		Type:             "Query Engine",
		ConnectionCmd:    "flight_sql_client --command Execute --host localhost --port 31337 --use-tls --tls-skip-verify --username ${FLIGHT_SQL_USER:-flight_username} --password ${FLIGHT_SQL_PASSWORD:-flight_password} --query 'SELECT version()'",
		DefaultUser:      "flight_username",
		DefaultPassword:  "flight_password",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 31337, Type: PortTypeDatabase, Name: "Flight SQL", Description: "Apache Arrow Flight SQL interface"},
		},
	},
	"flink": {
		Name:          "flink",
		Type:          "Distributed Data Processing",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 8081, Type: PortTypeWebUI, Name: "Web UI", Description: "Flink Dashboard", Path: "/"},
		},
	},
	"flink-jobmanager": {
		Name:          "flink-jobmanager",
		Type:          "Distributed Data Processing",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 8081, Type: PortTypeWebUI, Name: "JobManager UI", Description: "Flink JobManager Dashboard", Path: "/"},
		},
	},
	"grafana": {
		Name:             "grafana",
		Type:             "Data Visualisation",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPassword:  "admin",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 3000, Type: PortTypeWebUI, Name: "Web UI", Description: "Grafana Dashboard", Path: "/", RequiresAuth: true},
		},
	},
	"influxdb": {
		Name:             "influxdb",
		Type:             "Database",
		ConnectionCmd:    "influx",
		DefaultUser:      "admin",
		DefaultPassword:  "admin",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 8086, Type: PortTypeAPI, Name: "HTTP API", Description: "InfluxDB HTTP API", Path: "/"},
		},
	},
	"istio": {
		Name:          "istio",
		Type:          "Other",
		ConnectionCmd: "istioctl proxy-status",
		Ports:         []ServicePort{}, // Istio components have various ports, but no single main port
	},
	"jaeger": {
		Name:          "jaeger",
		Type:          "Tracing",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 16686, Type: PortTypeWebUI, Name: "Web UI", Description: "Jaeger Query Interface", Path: "/"},
			{InternalPort: 14268, Type: PortTypeAPI, Name: "Collector API", Description: "Jaeger Collector HTTP API", Path: "/"},
		},
	},
	"jenkins": {
		Name:             "jenkins",
		Type:             "Other",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 8080, Type: PortTypeWebUI, Name: "Web UI", Description: "Jenkins Dashboard", Path: "/", RequiresAuth: true},
		},
	},
	"kafka": {
		Name:          "kafka",
		Type:          "Messaging",
		ConnectionCmd: "kafka-topics --bootstrap-server localhost:9092 --list",
		Ports: []ServicePort{
			{InternalPort: 9092, Type: PortTypeMessaging, Name: "Kafka Protocol", Description: "Kafka broker port"},
		},
	},
	"keycloak": {
		Name:             "keycloak",
		Type:             "Identity Management",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPassword:  "admin",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 8080, Type: PortTypeWebUI, Name: "Admin Console", Description: "Keycloak Admin Console", Path: "/admin", RequiresAuth: true},
		},
	},
	"kibana": {
		Name:             "kibana",
		Type:             "Data Visualisation",
		ConnectionCmd:    "bash",
		DefaultUser:      "kibana_system",
		DefaultPassword:  "password",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 5601, Type: PortTypeWebUI, Name: "Web UI", Description: "Kibana Dashboard", Path: "/", RequiresAuth: true},
		},
	},
	"kong": {
		Name:          "kong",
		Type:          "Api Gateway",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 8000, Type: PortTypeAPI, Name: "Proxy API", Description: "Kong Proxy API", Path: "/"},
			{InternalPort: 8001, Type: PortTypeAdmin, Name: "Admin API", Description: "Kong Admin API", Path: "/"},
			{InternalPort: 8002, Type: PortTypeWebUI, Name: "Manager UI", Description: "Kong Manager Web Interface", Path: "/"},
		},
	},
	"label-studio": {
		Name:          "label-studio",
		Type:          "Data Annotation",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 8080, Type: PortTypeWebUI, Name: "Web UI", Description: "Label Studio Interface", Path: "/"},
		},
	},
	"lakekeeper": {
		Name:            "lakekeeper",
		Type:            "Data Catalog",
		ConnectionCmd:   "bash",
		DefaultUser:     "peter",
		DefaultPassword: "iceberg",
		Ports: []ServicePort{
			{InternalPort: 8181, Type: PortTypeAPI, Name: "REST API", Description: "Lakekeeper Catalog API", Path: "/"},
		},
	},
	"lakekeeper-jupyter": {
		Name:          "lakekeeper-jupyter",
		Type:          "Data Catalog",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 8888, Type: PortTypeWebUI, Name: "Jupyter Notebook", Description: "Jupyter Notebook Interface", Path: "/"},
		},
	},
	"logstash": {
		Name:             "logstash",
		Type:             "Data Collector",
		ConnectionCmd:    "bash",
		DefaultUser:      "logstash_internal",
		DefaultPassword:  "password",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 5044, Type: PortTypeOther, Name: "Beats Input", Description: "Logstash Beats input"},
			{InternalPort: 9600, Type: PortTypeAPI, Name: "API", Description: "Logstash Monitoring API", Path: "/"},
		},
	},
	"loki": {
		Name:          "loki",
		Type:          "Monitoring",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 3100, Type: PortTypeAPI, Name: "HTTP API", Description: "Loki Query API", Path: "/"},
		},
	},
	"maestro": {
		Name:          "maestro",
		Type:          "Workflow",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 8080, Type: PortTypeWebUI, Name: "Web UI", Description: "Maestro Workflow Interface", Path: "/"},
		},
	},
	"mage-ai": {
		Name:          "mage-ai",
		Type:          "Job Orchestrator",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 6789, Type: PortTypeWebUI, Name: "Web UI", Description: "Mage AI Interface", Path: "/"},
		},
	},
	"mariadb": {
		Name:             "mariadb",
		Type:             "Database",
		ConnectionCmd:    "mariadb --user=${MARIADB_USER:-user} --password=${MARIADB_PASSWORD:-password}",
		DefaultUser:      "user",
		DefaultPassword:  "password",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 3306, Type: PortTypeDatabase, Name: "MySQL", Description: "MySQL database port"},
		},
	},
	"marquez": {
		Name:          "marquez",
		Type:          "Data Catalog",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 3000, Type: PortTypeWebUI, Name: "Web UI", Description: "Marquez Lineage Interface", Path: "/"},
			{InternalPort: 5000, Type: PortTypeAPI, Name: "API", Description: "Marquez REST API", Path: "/"},
		},
	},
	"memcached": {
		Name:          "memcached",
		Type:          "Cache",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 11211, Type: PortTypeOther, Name: "Memcached", Description: "Memcached protocol port"},
		},
	},
	"metabase": {
		Name:          "metabase",
		Type:          "Data Visualisation",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 3000, Type: PortTypeWebUI, Name: "Web UI", Description: "Metabase Dashboard", Path: "/"},
		},
	},
	"milvus": {
		Name:          "milvus",
		Type:          "Database",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 19530, Type: PortTypeDatabase, Name: "gRPC API", Description: "Milvus gRPC interface"},
			{InternalPort: 9091, Type: PortTypeMetrics, Name: "Metrics", Description: "Milvus metrics endpoint", Path: "/metrics"},
		},
	},
	"minio": {
		Name:             "minio",
		Type:             "Object Storage",
		ConnectionCmd:    "mc alias set local http://localhost:9000 ${MINIO_USER:-minioadmin} ${MINIO_PASSWORD:-minioadmin}",
		DefaultUser:      "minioadmin",
		DefaultPassword:  "minioadmin",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 9000, Type: PortTypeAPI, Name: "S3 API", Description: "MinIO S3-compatible API", Path: "/"},
			{InternalPort: 9001, Type: PortTypeWebUI, Name: "Console", Description: "MinIO Console", Path: "/", RequiresAuth: true},
		},
	},
	"mlflow": {
		Name:          "mlflow",
		Type:          "ML Platform",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 5000, Type: PortTypeWebUI, Name: "Web UI", Description: "MLflow Tracking Server", Path: "/"},
		},
	},
	"mlflow-serve": {
		Name:          "mlflow-serve",
		Type:          "ML Platform",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 1234, Type: PortTypeAPI, Name: "Model API", Description: "MLflow Model Serving API", Path: "/"},
		},
	},
	"mongodb": {
		Name:             "mongodb",
		Type:             "Database",
		ConnectionCmd:    "mongosh mongodb://${MONGODB_USER:-root}:${MONGODB_PASSWORD:-root}@mongodb",
		DefaultUser:      "root",
		DefaultPassword:  "root",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 27017, Type: PortTypeDatabase, Name: "MongoDB", Description: "MongoDB database port"},
		},
	},
	"mssql": {
		Name:             "mssql",
		Type:             "Database",
		ConnectionCmd:    "/opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P \"${MSSQL_PASSWORD:-yourStrong(!)Password}\"",
		DefaultUser:      "sa",
		DefaultPassword:  "yourStrong(!)Password",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 1433, Type: PortTypeDatabase, Name: "SQL Server", Description: "Microsoft SQL Server port"},
		},
	},
	"mysql": {
		Name:             "mysql",
		Type:             "Database",
		ConnectionCmd:    "mysql -u ${MYSQL_USER:-root} -p${MYSQL_PASSWORD:-root}",
		DefaultUser:      "root",
		DefaultPassword:  "root",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 3306, Type: PortTypeDatabase, Name: "MySQL", Description: "MySQL database port"},
		},
	},
	"nats": {
		Name:          "nats",
		Type:          "Messaging",
		ConnectionCmd: "nats pub test.subject 'Hello NATS!'",
		Ports: []ServicePort{
			{InternalPort: 4222, Type: PortTypeMessaging, Name: "Client Port", Description: "NATS client connections"},
			{InternalPort: 8222, Type: PortTypeWebUI, Name: "Monitoring", Description: "NATS monitoring interface", Path: "/"},
		},
	},
	"neo4j": {
		Name:             "neo4j",
		Type:             "Database",
		ConnectionCmd:    "cypher-shell -u neo4j -p test",
		DefaultUser:      "neo4j",
		DefaultPassword:  "test",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 7474, Type: PortTypeWebUI, Name: "Browser", Description: "Neo4j Browser", Path: "/", RequiresAuth: true},
			{InternalPort: 7687, Type: PortTypeDatabase, Name: "Bolt", Description: "Neo4j Bolt protocol"},
		},
	},
	"openmetadata": {
		Name:          "openmetadata",
		Type:          "Data Catalog",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 8585, Type: PortTypeWebUI, Name: "Web UI", Description: "OpenMetadata Interface", Path: "/"},
		},
	},
	"opensearch": {
		Name:             "opensearch",
		Type:             "Database",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPassword:  "!BigData#1",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 9200, Type: PortTypeAPI, Name: "REST API", Description: "OpenSearch REST API", Path: "/"},
			{InternalPort: 5601, Type: PortTypeWebUI, Name: "Dashboards", Description: "OpenSearch Dashboards", Path: "/", RequiresAuth: true},
		},
	},
	"pinot": {
		Name:          "pinot",
		Type:          "Real-time OLAP",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 9000, Type: PortTypeWebUI, Name: "Controller", Description: "Pinot Controller Console", Path: "/"},
			{InternalPort: 8000, Type: PortTypeAPI, Name: "Broker API", Description: "Pinot Broker Query API", Path: "/"},
		},
	},
	"polaris": {
		Name:          "polaris",
		Type:          "Data Catalog",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 8181, Type: PortTypeAPI, Name: "REST API", Description: "Polaris Catalog API", Path: "/"},
		},
	},
	"postgres": {
		Name:             "postgres",
		Type:             "Database",
		ConnectionCmd:    "PGPASSWORD=${POSTGRES_PASSWORD:-postgres} psql -U${POSTGRES_USER:-postgres}",
		DefaultUser:      "postgres",
		DefaultPassword:  "postgres",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 5432, Type: PortTypeDatabase, Name: "PostgreSQL", Description: "PostgreSQL database port"},
		},
	},
	"prefect-data": {
		Name:          "prefect-data",
		Type:          "Job Orchestrator",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 4200, Type: PortTypeWebUI, Name: "Web UI", Description: "Prefect UI", Path: "/"},
		},
	},
	"presto": {
		Name:          "presto",
		Type:          "Query Engine",
		ConnectionCmd: "presto-cli",
		Ports: []ServicePort{
			{InternalPort: 8080, Type: PortTypeAPI, Name: "Coordinator", Description: "Presto Coordinator API", Path: "/"},
		},
	},
	"prometheus": {
		Name:          "prometheus",
		Type:          "Monitoring",
		ConnectionCmd: "promtool query instant http://localhost:9090 up",
		Ports: []ServicePort{
			{InternalPort: 9090, Type: PortTypeWebUI, Name: "Web UI", Description: "Prometheus Console", Path: "/"},
		},
	},
	"pulsar": {
		Name:          "pulsar",
		Type:          "Messaging",
		ConnectionCmd: "pulsar-client produce test-topic -m 'Hello Pulsar!'",
		Ports: []ServicePort{
			{InternalPort: 6650, Type: PortTypeMessaging, Name: "Broker", Description: "Pulsar broker port"},
			{InternalPort: 8080, Type: PortTypeWebUI, Name: "Admin Console", Description: "Pulsar Manager", Path: "/"},
		},
	},
	"qdrant": {
		Name:          "qdrant",
		Type:          "Database",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 6333, Type: PortTypeAPI, Name: "REST API", Description: "Qdrant REST API", Path: "/"},
			{InternalPort: 6334, Type: PortTypeAPI, Name: "gRPC API", Description: "Qdrant gRPC API"},
		},
	},
	"rabbitmq": {
		Name:             "rabbitmq",
		Type:             "Messaging",
		ConnectionCmd:    "rabbitmqctl --host localhost --username ${RABBITMQ_USER:-guest} --password ${RABBITMQ_PASSWORD:-guest} list_queues",
		DefaultUser:      "guest",
		DefaultPassword:  "guest",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 5672, Type: PortTypeMessaging, Name: "AMQP", Description: "RabbitMQ AMQP port"},
			{InternalPort: 15672, Type: PortTypeWebUI, Name: "Management", Description: "RabbitMQ Management Console", Path: "/", RequiresAuth: true},
		},
	},
	"ray": {
		Name:          "ray",
		Type:          "Distributed Data Processing",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 8265, Type: PortTypeWebUI, Name: "Dashboard", Description: "Ray Dashboard", Path: "/"},
			{InternalPort: 10001, Type: PortTypeOther, Name: "Client", Description: "Ray client port"},
		},
	},
	"redash": {
		Name:          "redash",
		Type:          "Data Visualisation",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 5000, Type: PortTypeWebUI, Name: "Web UI", Description: "Redash Dashboard", Path: "/"},
		},
	},
	"redis": {
		Name:          "redis",
		Type:          "Cache",
		ConnectionCmd: "redis-cli",
		Ports: []ServicePort{
			{InternalPort: 6379, Type: PortTypeDatabase, Name: "Redis", Description: "Redis database port"},
		},
	},
	"solace": {
		Name:             "solace",
		Type:             "Messaging",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPassword:  "admin",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 8080, Type: PortTypeWebUI, Name: "PubSub+ Manager", Description: "Solace Management Interface", Path: "/", RequiresAuth: true},
			{InternalPort: 55555, Type: PortTypeMessaging, Name: "SMF", Description: "Solace Message Format port"},
		},
	},
	"sonarqube": {
		Name:             "sonarqube",
		Type:             "Code Analysis",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPassword:  "admin",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 9000, Type: PortTypeWebUI, Name: "Web UI", Description: "SonarQube Interface", Path: "/", RequiresAuth: true},
		},
	},
	"spanner": {
		Name:          "spanner",
		Type:          "Database",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 9010, Type: PortTypeDatabase, Name: "gRPC API", Description: "Cloud Spanner Emulator gRPC API"},
			{InternalPort: 9020, Type: PortTypeAPI, Name: "REST API", Description: "Cloud Spanner Emulator REST API", Path: "/"},
		},
	},
	"superset": {
		Name:             "superset",
		Type:             "Data Visualisation",
		ConnectionCmd:    "bash",
		DefaultUser:      "admin",
		DefaultPassword:  "admin",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 8088, Type: PortTypeWebUI, Name: "Web UI", Description: "Apache Superset Dashboard", Path: "/", RequiresAuth: true},
		},
	},
	"temporal": {
		Name:          "temporal",
		Type:          "Workflow",
		ConnectionCmd: "temporal workflow list",
		Ports: []ServicePort{
			{InternalPort: 7233, Type: PortTypeAPI, Name: "Frontend", Description: "Temporal Frontend gRPC API"},
			{InternalPort: 8080, Type: PortTypeWebUI, Name: "Web UI", Description: "Temporal Web Interface", Path: "/"},
		},
	},
	"timescaledb": {
		Name:             "timescaledb",
		Type:             "Database",
		ConnectionCmd:    "PGPASSWORD=${TIMESCALEDB_PASSWORD:-postgres} psql -U${TIMESCALEDB_USER:-postgres}",
		DefaultUser:      "postgres",
		DefaultPassword:  "postgres",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 5432, Type: PortTypeDatabase, Name: "PostgreSQL", Description: "TimescaleDB PostgreSQL-compatible port"},
		},
	},
	"traefik": {
		Name:          "traefik",
		Type:          "Other",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 8080, Type: PortTypeWebUI, Name: "Dashboard", Description: "Traefik Dashboard", Path: "/"},
		},
	},
	"trino": {
		Name:          "trino",
		Type:          "Query Engine",
		ConnectionCmd: "trino",
		Ports: []ServicePort{
			{InternalPort: 8080, Type: PortTypeAPI, Name: "Coordinator", Description: "Trino Coordinator API", Path: "/"},
		},
	},
	"unitycatalog": {
		Name:          "unitycatalog",
		Type:          "Data Catalog",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 8080, Type: PortTypeAPI, Name: "REST API", Description: "Unity Catalog API", Path: "/"},
		},
	},
	"vault": {
		Name:             "vault",
		Type:             "Secret Management",
		ConnectionCmd:    "vault status",
		DefaultUser:      "root",
		RequiresPassword: true,
		Ports: []ServicePort{
			{InternalPort: 8200, Type: PortTypeWebUI, Name: "Web UI", Description: "Vault Web Interface", Path: "/", RequiresAuth: true},
		},
	},
	"weaviate": {
		Name:          "weaviate",
		Type:          "Database",
		ConnectionCmd: "bash",
		Ports: []ServicePort{
			{InternalPort: 8080, Type: PortTypeAPI, Name: "GraphQL API", Description: "Weaviate GraphQL API", Path: "/"},
		},
	},
	"zookeeper": {
		Name:          "zookeeper",
		Type:          "Distributed Coordination",
		ConnectionCmd: "zkCli.sh",
		Ports: []ServicePort{
			{InternalPort: 2181, Type: PortTypeOther, Name: "Client Port", Description: "ZooKeeper client port"},
			{InternalPort: 8080, Type: PortTypeWebUI, Name: "Admin Server", Description: "ZooKeeper Admin Server", Path: "/"},
		},
	},
}
