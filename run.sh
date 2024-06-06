#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
LIGHT_BLUE='\033[1;34m'
NC='\033[0m'

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

check_docker_installed() {
  echo -e "${GREEN}Checking for docker and docker-compose...${NC}"
  if ! command -v docker &>/dev/null; then
    echo -e "${RED}Error: docker could not be found${NC}"
    exit 1
  fi
  if ! command -v docker-compose &>/dev/null; then
    echo -e "${RED}Error: docker-compose could not be found${NC}"
    exit 1
  fi
}

startup_services() {
  all_services=("$@")
  echo -e "${GREEN}Starting up services...${NC}"
  docker-compose -f "$SCRIPT_DIR/docker-compose.yaml" up -d "$@"
  if [ $? != 0 ]; then
    echo -e "${RED}Failed to start up services${NC}"
    exit 1
  fi
  sleep 2
}

log_how_to_connect() {
  echo -e "${GREEN}How to connect:${NC}"
  connect_result=("${YELLOW}Service,${YELLOW}Container To Container,Host To Container,Container To Host")
  for service in "${all_services[@]}"; do
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

if [ "$1" == "-h" ] || [ "$1" == "--help" ]; then
  echo "Usage: $(basename "$0") [service...] [-h][--help]"
  echo
  echo "    service: name of service to run"
  exit 0
fi

check_docker_installed
startup_services "$@"
log_how_to_connect

