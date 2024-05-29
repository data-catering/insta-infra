#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
LIGHT_BLUE='\033[1;34m'
NC='\033[0m'

echo -e "${GREEN}Checking for docker and docker-compose...${NC}"
if ! command -v docker &>/dev/null; then
  echo -e "${RED}Error: docker could not be found${NC}"
  exit 1
fi
if ! command -v docker-compose &>/dev/null; then
  echo -e "${RED}Error: docker-compose could not be found${NC}"
  exit 1
fi

all_services=("$@")
echo -e "${GREEN}Starting up services...${NC}"
docker-compose up -d "$@"
sleep 2

echo -e "${GREEN}How to connect:${NC}"
connect_result=("${YELLOW}Service,${YELLOW}Container To Container,Host To Container,Container To Host")
for service in "${all_services[@]}"; do
  ports=$(docker inspect "$service" | grep HostPort | sed -nr 's/.*\: "([0-9]+)"/\1/p' | sort -u)
  for port in $ports; do
    current_service="${RED}$service,${LIGHT_BLUE}$service:$port,localhost:$port,host.docker.internal:$port"
    connect_result+=("$current_service")
  done
done

for value in "${connect_result[@]}"; do
    echo -e "$value"
done | column -t -s ','
