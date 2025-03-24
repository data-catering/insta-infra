#!/usr/bin/env bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PARENT_DIR="$(dirname "$SCRIPT_DIR")"
PACKAGE_JSON="$PARENT_DIR/package.json"

# Test if package.json exists
test_package_json_exists() {
  if [ -f "$PACKAGE_JSON" ]; then
    echo -e "${GREEN}✓ package.json exists${NC}"
    return 0
  else
    echo -e "${RED}✗ package.json does not exist${NC}"
    return 1
  fi
}

# Test package.json syntax
test_package_json_syntax() {
  if command -v node &>/dev/null; then
    if node -e "JSON.parse(require('fs').readFileSync('$PACKAGE_JSON', 'utf8'))" &>/dev/null; then
      echo -e "${GREEN}✓ package.json syntax is valid${NC}"
      return 0
    else
      echo -e "${RED}✗ package.json syntax is invalid${NC}"
      return 1
    fi
  else
    # Try with python if node is not available
    if command -v python &>/dev/null || command -v python3 &>/dev/null; then
      PYTHON_CMD="python"
      if ! command -v python &>/dev/null; then
        PYTHON_CMD="python3"
      fi
      
      if $PYTHON_CMD -c "import json; json.load(open('$PACKAGE_JSON'))" &>/dev/null; then
        echo -e "${GREEN}✓ package.json syntax is valid (validated with Python)${NC}"
        return 0
      else
        echo -e "${RED}✗ package.json syntax is invalid (validated with Python)${NC}"
        return 1
      fi
    else
      echo -e "${YELLOW}! Skipping package.json syntax check (Node.js and Python not installed)${NC}"
      return 0
    fi
  fi
}

# Test required fields in package.json
test_package_json_fields() {
  if ! command -v node &>/dev/null && ! command -v python &>/dev/null && ! command -v python3 &>/dev/null; then
    echo -e "${YELLOW}! Skipping package.json fields check (Node.js and Python not installed)${NC}"
    return 0
  fi
  
  REQUIRED_FIELDS=("name" "version" "bin" "description")
  MISSING_FIELDS=()
  
  if command -v node &>/dev/null; then
    for field in "${REQUIRED_FIELDS[@]}"; do
      if ! node -e "process.exit(require('$PACKAGE_JSON').$field ? 0 : 1)" &>/dev/null; then
        MISSING_FIELDS+=("$field")
      fi
    done
  elif command -v python &>/dev/null || command -v python3 &>/dev/null; then
    PYTHON_CMD="python"
    if ! command -v python &>/dev/null; then
      PYTHON_CMD="python3"
    fi
    
    for field in "${REQUIRED_FIELDS[@]}"; do
      if ! $PYTHON_CMD -c "import json; exit(0 if '$field' in json.load(open('$PACKAGE_JSON')) else 1)" &>/dev/null; then
        MISSING_FIELDS+=("$field")
      fi
    done
  fi
  
  if [ ${#MISSING_FIELDS[@]} -eq 0 ]; then
    echo -e "${GREEN}✓ package.json contains all required fields${NC}"
    return 0
  else
    echo -e "${RED}✗ package.json is missing required fields:${NC}"
    for field in "${MISSING_FIELDS[@]}"; do
      echo -e "${RED}  - $field${NC}"
    done
    return 1
  fi
}

# Test bin field points to run.sh
test_package_bin_field() {
  if ! command -v node &>/dev/null && ! command -v python &>/dev/null && ! command -v python3 &>/dev/null; then
    echo -e "${YELLOW}! Skipping package.json bin field check (Node.js and Python not installed)${NC}"
    return 0
  fi
  
  if command -v node &>/dev/null; then
    bin_value=$(node -e "console.log(typeof require('$PACKAGE_JSON').bin === 'object' ? require('$PACKAGE_JSON').bin.insta : require('$PACKAGE_JSON').bin)")
    if [ "$bin_value" = "./run.sh" ]; then
      echo -e "${GREEN}✓ package.json bin field correctly points to ./run.sh${NC}"
      return 0
    else
      echo -e "${RED}✗ package.json bin field does not correctly point to ./run.sh (got: $bin_value)${NC}"
      return 1
    fi
  elif command -v python &>/dev/null || command -v python3 &>/dev/null; then
    PYTHON_CMD="python"
    if ! command -v python &>/dev/null; then
      PYTHON_CMD="python3"
    fi
    
    bin_value=$($PYTHON_CMD -c "import json; data = json.load(open('$PACKAGE_JSON')); print(data['bin']['insta'] if isinstance(data['bin'], dict) else data['bin'])" 2>/dev/null)
    if [ "$bin_value" = "./run.sh" ]; then
      echo -e "${GREEN}✓ package.json bin field correctly points to ./run.sh${NC}"
      return 0
    else
      echo -e "${RED}✗ package.json bin field does not correctly point to ./run.sh (got: $bin_value)${NC}"
      return 1
    fi
  fi
}

# Run all tests
run_tests() {
  failures=0
  
  test_package_json_exists || ((failures++))
  
  # Only run these tests if package.json exists
  if [ -f "$PACKAGE_JSON" ]; then
    test_package_json_syntax || ((failures++))
    test_package_json_fields || ((failures++))
    test_package_bin_field || ((failures++))
  fi
  
  if [ "$failures" -eq 0 ]; then
    echo -e "\n${GREEN}All npm package tests passed!${NC}"
    return 0
  else
    echo -e "\n${RED}$failures npm package tests failed!${NC}"
    return 1
  fi
}

run_tests 