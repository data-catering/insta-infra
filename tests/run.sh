#!/usr/bin/env bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PARENT_DIR="$(dirname "$SCRIPT_DIR")"
FAILURES=0
TOTAL=0

log_success() {
  echo -e "${GREEN}✓ $1${NC}"
}

log_failure() {
  echo -e "${RED}✗ $1${NC}"
  FAILURES=$((FAILURES + 1))
}

# Run all test files
run_all_tests() {
  echo "Running all tests..."
  
  # Find all test files
  test_files=($(find "$SCRIPT_DIR" -name "test_*.sh" -type f | sort))
  TOTAL=${#test_files[@]}
  
  # Run each test file
  for test_file in "${test_files[@]}"; do
    echo -e "\n${YELLOW}Running $(basename "$test_file")...${NC}"
    chmod +x "$test_file"
    if "$test_file"; then
      log_success "$(basename "$test_file") passed"
    else
      log_failure "$(basename "$test_file") failed"
    fi
  done
  
  # Print summary
  echo -e "\n${YELLOW}Test Summary:${NC}"
  echo -e "Total: $TOTAL, Failed: $FAILURES"
  
  if [ "$FAILURES" -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
  else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
  fi
}

run_all_tests 