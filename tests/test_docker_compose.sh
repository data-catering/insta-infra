#!/usr/bin/env bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PARENT_DIR="$(dirname "$SCRIPT_DIR")"
COMPOSE_FILE="$PARENT_DIR/docker-compose.yaml"
PERSIST_COMPOSE_FILE="$PARENT_DIR/docker-compose-persist.yaml"

# Test if the main docker-compose file exists
test_compose_file_exists() {
  if [ -f "$COMPOSE_FILE" ]; then
    echo -e "${GREEN}✓ docker-compose.yaml exists${NC}"
    return 0
  else
    echo -e "${RED}✗ docker-compose.yaml does not exist${NC}"
    return 1
  fi
}

# Test if the persist docker-compose file exists
test_persist_compose_file_exists() {
  if [ -f "$PERSIST_COMPOSE_FILE" ]; then
    echo -e "${GREEN}✓ docker-compose-persist.yaml exists${NC}"
    return 0
  else
    echo -e "${RED}✗ docker-compose-persist.yaml does not exist${NC}"
    return 1
  fi
}

# Validate docker-compose file syntax
test_compose_syntax() {
  if command -v docker &>/dev/null; then
    if docker compose -f "$COMPOSE_FILE" config > /dev/null 2>&1; then
      echo -e "${GREEN}✓ docker-compose.yaml syntax is valid${NC}"
      return 0
    else
      echo -e "${RED}✗ docker-compose.yaml syntax is invalid${NC}"
      return 1
    fi
  else
    echo -e "${YELLOW}! Skipping docker-compose syntax check (Docker not installed)${NC}"
    return 0
  fi
}

# Validate persist docker-compose file syntax
test_persist_compose_syntax() {
  if command -v docker &>/dev/null; then
    if docker compose -f "$COMPOSE_FILE" -f "$PERSIST_COMPOSE_FILE" config > /dev/null 2>&1; then
      echo -e "${GREEN}✓ docker-compose-persist.yaml syntax is valid${NC}"
      return 0
    else
      echo -e "${RED}✗ docker-compose-persist.yaml syntax is invalid${NC}"
      return 1
    fi
  else
    echo -e "${YELLOW}! Skipping docker-compose-persist syntax check (Docker not installed)${NC}"
    return 0
  fi
}

# Check that README services match docker-compose services
test_readme_services_match() {
  # Get services from docker-compose.yaml
  if ! command -v docker &>/dev/null; then
    echo -e "${YELLOW}! Skipping README services match check (Docker not installed)${NC}"
    return 0
  fi
  
  compose_services=$(docker compose -f "$COMPOSE_FILE" config --services | sort)
  
  # Get services from README.md
  readme_services=$(awk '/## Services/{y=1;next}y' "$PARENT_DIR/README.md" | grep '✅' | awk -F'|' '{print $3}' | tr -d ' ' | sort)
  
  # Check if any service in README is not in docker-compose
  missing_services=()
  for service in $readme_services; do
    if ! echo "$compose_services" | grep -q "^$service$"; then
      missing_services+=("$service")
    fi
  done
  
  if [ ${#missing_services[@]} -eq 0 ]; then
    echo -e "${GREEN}✓ All README services are in docker-compose.yaml${NC}"
    return 0
  else
    echo -e "${RED}✗ Some README services are missing from docker-compose.yaml:${NC}"
    for service in "${missing_services[@]}"; do
      echo -e "${RED}  - $service${NC}"
    done
    return 1
  fi
}

# Run all tests
run_tests() {
  failures=0
  
  test_compose_file_exists || ((failures++))
  test_persist_compose_file_exists || ((failures++))
  
  # Only run these tests if Docker is installed
  if command -v docker &>/dev/null; then
    test_compose_syntax || ((failures++))
    test_persist_compose_syntax || ((failures++))
    test_readme_services_match || ((failures++))
  else
    echo -e "${YELLOW}! Skipping docker-compose validation tests (Docker not installed)${NC}"
  fi
  
  if [ "$failures" -eq 0 ]; then
    echo -e "\n${GREEN}All docker-compose tests passed!${NC}"
    return 0
  else
    echo -e "\n${RED}$failures docker-compose tests failed!${NC}"
    return 1
  fi
}

run_tests 