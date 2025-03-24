#!/usr/bin/env bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PARENT_DIR="$(dirname "$SCRIPT_DIR")"
RUN_SCRIPT="$PARENT_DIR/run.sh"

# Check prerequisites
if ! command -v docker &>/dev/null; then
  echo -e "${YELLOW}! Skipping integration tests (Docker not installed)${NC}"
  exit 0
fi

# Test starting a simple service (httpbin is lightest option)
test_start_httpbin() {
  echo -e "Starting httpbin service..."
  if "$RUN_SCRIPT" httpbin > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Started httpbin service successfully${NC}"
    
    # Wait a bit for service to fully start
    sleep 5
    
    # Check if container is running
    if docker ps | grep -q httpbin; then
      echo -e "${GREEN}✓ httpbin container is running${NC}"
      return 0
    else
      echo -e "${RED}✗ httpbin container is not running${NC}"
      return 1
    fi
  else
    echo -e "${RED}✗ Failed to start httpbin service${NC}"
    return 1
  fi
}

# Test if service is accessible
test_service_accessible() {
  echo -e "Testing if httpbin service is accessible..."
  
  # Get the port
  port=$(docker port http | grep -oE '0.0.0.0:[0-9]+' | cut -d':' -f2)
  
  if [ -z "$port" ]; then
    echo -e "${RED}✗ Could not determine httpbin port${NC}"
    return 1
  fi
  
  # Try to access the service
  if command -v curl &>/dev/null; then
    if curl -s -o /dev/null -w "%{http_code}" "http://localhost:$port/get" | grep -q "200"; then
      echo -e "${GREEN}✓ httpbin service is accessible via HTTP${NC}"
      return 0
    else
      echo -e "${RED}✗ httpbin service is not accessible via HTTP${NC}"
      return 1
    fi
  else
    # If curl is not available, use a simple check without testing HTTP
    if nc -z localhost "$port" &>/dev/null; then
      echo -e "${GREEN}✓ Port $port is open (httpbin service)${NC}"
      return 0
    else
      echo -e "${RED}✗ Port $port is not accessible (httpbin service)${NC}"
      return 1
    fi
  fi
}

# Test stopping the service
test_stop_service() {
  echo -e "Stopping httpbin service..."
  if "$RUN_SCRIPT" -d httpbin > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Stopped httpbin service successfully${NC}"
    
    # Check if container is running
    if docker ps | grep -q httpbin; then
      echo -e "${RED}✗ httpbin container is still running${NC}"
      return 1
    else
      echo -e "${GREEN}✓ httpbin container is no longer running${NC}"
      return 0
    fi
  else
    echo -e "${RED}✗ Failed to stop httpbin service${NC}"
    return 1
  fi
}

# Run all integration tests
run_integration_tests() {
  failures=0
  
  # Start services
  test_start_httpbin || ((failures++))
  
  # Only continue if service started
  if docker ps | grep -q httpbin; then
    test_service_accessible || ((failures++))
    test_stop_service || ((failures++))
  else
    echo -e "${RED}✗ Skipping remaining tests as httpbin failed to start${NC}"
    ((failures++))
  fi
  
  # Cleanup in case tests failed
  "$RUN_SCRIPT" -d httpbin > /dev/null 2>&1 || true
  
  if [ "$failures" -eq 0 ]; then
    echo -e "\n${GREEN}All integration tests passed!${NC}"
    return 0
  else
    echo -e "\n${RED}$failures integration tests failed!${NC}"
    return 1
  fi
}

run_integration_tests 