#!/usr/bin/env bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PARENT_DIR="$(dirname "$SCRIPT_DIR")"
RUN_SCRIPT="$PARENT_DIR/run.sh"

# Test if the script exists
test_script_exists() {
  if [ -f "$RUN_SCRIPT" ]; then
    echo -e "${GREEN}✓ run.sh exists${NC}"
    return 0
  else
    echo -e "${RED}✗ run.sh does not exist${NC}"
    return 1
  fi
}

# Test if the script is executable
test_script_executable() {
  if [ -x "$RUN_SCRIPT" ]; then
    echo -e "${GREEN}✓ run.sh is executable${NC}"
    return 0
  else
    echo -e "${RED}✗ run.sh is not executable${NC}"
    return 1
  fi
}

# Test script help output
test_script_help() {
  output=$("$RUN_SCRIPT" --help 2>&1)
  if echo "$output" | grep -q "Usage:"; then
    echo -e "${GREEN}✓ Help command works${NC}"
    return 0
  else
    echo -e "${RED}✗ Help command failed${NC}"
    return 1
  fi
}

# Test list services command
test_list_services() {
  output=$("$RUN_SCRIPT" -l 2>&1)
  if echo "$output" | grep -q "Supported services:"; then
    echo -e "${GREEN}✓ List services command works${NC}"
    return 0
  else
    echo -e "${RED}✗ List services command failed${NC}"
    return 1
  fi
}

# Test Docker installed check
test_docker_installed() {
  # We need to source the script to test internal functions
  source "$RUN_SCRIPT"
  
  # Save current PATH to restore later
  OLD_PATH="$PATH"
  
  # Mock command function to simulate Docker not being installed
  command() {
    if [ "$2" = "docker" ]; then
      return 1
    fi
    return 0
  }
  
  # Redirect stderr to capture output
  output=$(check_docker_installed 2>&1) || true
  
  # Restore PATH
  PATH="$OLD_PATH"
  
  if echo "$output" | grep -q "Error: docker could not be found"; then
    echo -e "${GREEN}✓ Docker check works correctly${NC}"
    return 0
  else
    echo -e "${RED}✗ Docker check failed${NC}"
    return 1
  fi
}

# Run all tests
run_tests() {
  failures=0
  
  test_script_exists || ((failures++))
  test_script_executable || ((failures++))
  test_script_help || ((failures++))
  test_list_services || ((failures++))
  
  # Only run this test if we're in CI environment, as it mocks system functions
  if [ -n "$CI" ]; then
    test_docker_installed || ((failures++))
  fi
  
  if [ "$failures" -eq 0 ]; then
    echo -e "\n${GREEN}All core tests passed!${NC}"
    return 0
  else
    echo -e "\n${RED}$failures core tests failed!${NC}"
    return 1
  fi
}

run_tests 