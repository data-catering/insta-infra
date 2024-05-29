#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
LIGHT_BLUE='\033[1;34m'
NC='\033[0m'

if [ -z "$1" ]
then
  echo -e "${RED}Error: No service name passed as argument${NC}"
  exit 1
fi

connection_commands="
cassandra=cqlsh
elasticsearch=elasticsearch-sql-cli http://elastic:elasticsearch@localhost:9200
mysql=mysql -u root -proot
postgres=PGPASSWORD=postgres psql -Upostgres
"

echo -e "${GREEN}Connecting to $1...${NC}"
connection_command=$(echo "$connection_commands" | grep "$1=" | sed -nr 's/.*=(.*)/\1/p')
if [ -z "$connection_command" ]
then
  echo -e "${RED}Error: Failed to find connection command for $1${NC}"
  exit 1
fi

docker exec -it "$1" bash -c "$connection_command"
