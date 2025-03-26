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
		ConnectionCmd:    "/var/lib/artemis-instance/bin/artemis shell",
		DefaultUser:      "artemis",
		DefaultPassword:  "artemis",
		RequiresPassword: true,
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
	"doris": {
		Name:          "doris",
		ConnectionCmd: "mysql -uroot -P9030 -h127.0.0.1",
		DefaultPort:   9030,
	},
	"duckdb": {
		Name:          "duckdb",
		ConnectionCmd: "./duckdb",
	},
	"elasticsearch": {
		Name:             "elasticsearch",
		ConnectionCmd:    "elasticsearch-sql-cli",
		DefaultUser:      "elastic",
		DefaultPassword:  "elasticsearch",
		DefaultPort:      9200,
		RequiresPassword: true,
	},
	"flight-sql": {
		Name:             "flight-sql",
		ConnectionCmd:    "flight_sql_client --command Execute --host localhost --port 31337 --use-tls --tls-skip-verify",
		DefaultUser:      "flight_username",
		DefaultPassword:  "flight_password",
		DefaultPort:      31337,
		RequiresPassword: true,
	},
	"mariadb": {
		Name:             "mariadb",
		ConnectionCmd:    "mariadb",
		DefaultUser:      "user",
		DefaultPassword:  "password",
		DefaultPort:      3306,
		RequiresPassword: true,
	},
	"mongodb": {
		Name:             "mongodb",
		ConnectionCmd:    "mongosh",
		DefaultUser:      "root",
		DefaultPassword:  "root",
		DefaultPort:      27017,
		RequiresPassword: true,
	},
	"mysql": {
		Name:             "mysql",
		ConnectionCmd:    "mysql",
		DefaultUser:      "root",
		DefaultPassword:  "root",
		DefaultPort:      3306,
		RequiresPassword: true,
	},
	"mssql": {
		Name:             "mssql",
		ConnectionCmd:    "/opt/mssql-tools/bin/sqlcmd -S localhost",
		DefaultUser:      "sa",
		DefaultPassword:  "yourStrong(!)Password",
		DefaultPort:      1433,
		RequiresPassword: true,
	},
	"neo4j": {
		Name:             "neo4j",
		ConnectionCmd:    "cypher-shell",
		DefaultUser:      "neo4j",
		DefaultPassword:  "test",
		DefaultPort:      7687,
		RequiresPassword: true,
	},
	"postgres": {
		Name:             "postgres",
		ConnectionCmd:    "psql",
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
}
