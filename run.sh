#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
LIGHT_BLUE='\033[1;34m'
NC='\033[0m'

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
COMPOSE_FILES="-f $SCRIPT_DIR/docker-compose.yaml"

connection_commands="
activemq='/var/lib/artemis-instance/bin/artemis shell --user ${ARTEMIS_USER:-artemis} --password ${ARTEMIS_PASSWORD:-artemis}'
cassandra='cqlsh'
clickhouse='clickhouse-client'
cockroachdb='./cockroach sql --insecure'
doris='mysql -uroot -P9030 -h127.0.0.1'
duckdb='./duckdb'
elasticsearch='elasticsearch-sql-cli http://elastic:${ELASTICSEARCH_PASSWORD:-elasticsearch}@localhost:9200'
flight-sql='flight_sql_client --command Execute --host localhost --port 31337 --username ${FLIGHT_SQL_USER:-flight_username} --password ${FLIGHT_SQL_PASSWORD:-flight_password} --query 'SELECT version()' --use-tls --tls-skip-verify'
mariadb='mariadb --user=${MARIADB_USER:-user} --password=${MARIADB_PASSWORD:-password}'
mongodb-connect='mongosh mongodb://${MONGODB_USER:-root}:${MONGODB_PASSWORD:-root}@mongodb'
mysql='mysql -u ${MYSQL_USER:-root} -p${MYSQL_PASSWORD:-root}'
mssql='/opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P \"${MSSQL_PASSWORD:-yourStrong(!)Password}\"'
neo4j='cypher-shell -u neo4j -p test'
postgres='PGPASSWORD=${POSTGRES_PASSWORD:-postgres} psql -U${POSTGRES_USER:-postgres}'
prefect-data='bash'
presto='presto-cli'
trino='trino'
unitycatalog='bash'
flink='bash'
flink-jobmanager='bash'
"

usage() {
  echo "Usage: $(basename "$0") [options...] [services...]"
  echo
  echo "    <services>                Name of services to run"
  echo "    -c, connect [service]     Connect to service"
  echo "    -d, down [services...]    Shutdown services (if empty, shutdown all services)"
  echo "    -h, --help, help          Show help"
  echo "    -l, list                  List supported services"
  echo "    -p                        Run service(s) with persisted data"
  echo "    -r, remove [services...]  Remove persisted data (if empty, remove all services persisted data)"
  echo
  echo "Examples:"
  echo "    $(basename "$0") -l"
  echo "    $(basename "$0") postgres           Spin up Postgres"
  echo "    $(basename "$0") -c postgres        Connect to Postgres"
  echo "    $(basename "$0") -d                 Bring Postgres down"
  echo "    $(basename "$0") -p postgres        Run Postgres with persisted data"
  echo "    $(basename "$0") -r postgres        Remove Postgres persisted data"
  exit 0
}

connect_to_service() {
  if [ -z "$1" ]
  then
    echo -e "${RED}Error: No service name passed as argument${NC}"
    exit 1
  fi

  echo -e "${GREEN}Connecting to $1...${NC}"
  base_command=$(echo "$connection_commands" | grep "^$1")
  IFS=$'\t' read -r container_name connection_command \
    < <(sed -nr "s/(.*)='(.*)'/\1\t\2/p" <<< "$base_command")

  if [ -z "$connection_command" ]
  then
    echo -e "${RED}Error: Failed to find connection command for $1${NC}"
    exit 1
  fi

  docker exec -it "$container_name" bash -c "$connection_command"
}

shutdown_service() {
  if [ -z "$1" ]; then
    echo "Shutting down all services..."
    docker compose -f "$SCRIPT_DIR/docker-compose.yaml" down
  else
    echo "Shutting down services: $*..."
    docker compose -f "$SCRIPT_DIR/docker-compose.yaml" down "$@"
  fi
}

list_supported_services() {
  supported_services=$(awk '/## Services/{y=1;next}y' "$SCRIPT_DIR/README.md" | grep '✅' | awk -F'|' '{print $3}' | sort | xargs)
  echo -e "Supported services: ${GREEN}$supported_services${NC}"
}

check_docker_installed() {
  if ! command -v docker &>/dev/null; then
    echo -e "${RED}Error: docker could not be found${NC}"
    exit 1
  fi
  if ! command -v docker compose &>/dev/null; then
    echo -e "${RED}Error: docker compose could not be found${NC}"
    exit 1
  fi
}

check_persist_flag() {
  if [ "$1" = "-p" ]; then
    echo -e "${YELLOW}Persisting data to host"
    COMPOSE_FILES="${COMPOSE_FILES} -f docker-compose-persist.yaml"
    shift
  fi
  original_services=("$@")
  service_array=("$@")
}

startup_services() {
  echo -e "${GREEN}Starting up services...${NC}"
  docker compose $COMPOSE_FILES up -d --quiet-pull "$@"
  if [ $? != 0 ]; then
    echo -e "${RED}Error: Failed to start up services${NC}"
    exit 1
  fi
  sleep 2
}

log_how_to_connect() {
  echo -e "${GREEN}How to connect:${NC}"
  connect_result=("${YELLOW}Service,${YELLOW}Container To Container,Host To Container,Container To Host")
  for service in "${original_services[@]}"; do
    ports=$(docker inspect "$service" | grep HostPort | sed -nr 's/.*\: "([0-9]+)"/\1/p' | sort -u)
    for port in $ports; do
      container_port=$(docker inspect "$service" | grep -B 3 "HostPort\": \"${port}\"" | sed -nr 's/.*\"([0-9]+)\/tcp\".*/\1/p' | head -1)
      current_service="${RED}$service,${LIGHT_BLUE}$service:$container_port,localhost:$port,host.docker.internal:$port"
      connect_result+=("$current_service")
    done
  done

  for value in "${connect_result[@]}"; do
      echo -e "$value"
  done | column -t -s ','
}

remove_persisted_data() {
  if [ -z "$1" ]; then
    read -p "Continue to remove all persisted data? (Y/n)" CONT
    if [ "$CONT" = "Y" ]; then
      echo "Removing all services persisted data..."
      find "${SCRIPT_DIR}/data" -type d -name "persist" -maxdepth 2 -exec rm -r {} \;
    else
      echo "Not removing any persisted data";
    fi
  else
    read -p "Continue to remove persisted data for services: $*? (Y/n)" CONT
    if [ "$CONT" = "Y" ]; then
      echo "Removing persisted data for services: $*..."
      for service in "$@"; do
        rm -r "${SCRIPT_DIR}/data/${service}/persist"
      done
    else
      echo "Not removing any persisted data";
    fi
  fi
}

case $1 in
  "-h"|"--help"|"help")
    usage
    ;;
  "-c"|"connect")
    connect_to_service "$2"
    ;;
  "-d"|"down")
    shutdown_service "${@:2}"
    ;;
  "-l"|"list")
    list_supported_services
    ;;
  "-r"|"remove")
    remove_persisted_data "${@:2}"
    ;;
  *)
    if [ $# -eq 0 ]; then
      usage
    else
      check_docker_installed
      service_array=()
      check_persist_flag "$@"
      startup_services "${service_array[@]}"
      log_how_to_connect
    fi
    ;;
esac

