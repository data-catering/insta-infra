#!/usr/bin/env bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PARENT_DIR="$(dirname "$SCRIPT_DIR")"
INSTALL_SCRIPT="$PARENT_DIR/install.sh"

# Test if the install script exists
test_install_script_exists() {
  if [ -f "$INSTALL_SCRIPT" ]; then
    echo -e "${GREEN}✓ install.sh exists${NC}"
    return 0
  else
    echo -e "${RED}✗ install.sh does not exist${NC}"
    return 1
  fi
}

# Test if the install script is executable
test_install_script_executable() {
  if [ -x "$INSTALL_SCRIPT" ]; then
    echo -e "${GREEN}✓ install.sh is executable${NC}"
    return 0
  else
    echo -e "${RED}✗ install.sh is not executable${NC}"
    chmod +x "$INSTALL_SCRIPT"
    echo -e "${YELLOW}! Made install.sh executable${NC}"
    return 0
  fi
}

# Test install script syntax
test_install_script_syntax() {
  if bash -n "$INSTALL_SCRIPT" &>/dev/null; then
    echo -e "${GREEN}✓ install.sh syntax is valid${NC}"
    return 0
  else
    echo -e "${RED}✗ install.sh syntax is invalid${NC}"
    return 1
  fi
}

# Test if install script references valid repository URL
test_install_script_urls() {
  github_url=$(grep -o 'github.com/[^/]*/insta-infra' "$INSTALL_SCRIPT" | head -1)
  
  if [ -z "$github_url" ]; then
    echo -e "${RED}✗ No valid GitHub URL found in install.sh${NC}"
    return 1
  else
    echo -e "${GREEN}✓ install.sh contains GitHub URL: $github_url${NC}"
    
    # Test if repository exists (optional, can be skipped in CI)
    if [ -z "$CI" ] && command -v curl &>/dev/null; then
      http_code=$(curl -s -o /dev/null -w "%{http_code}" "https://$github_url")
      if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✓ GitHub repository exists${NC}"
      else
        echo -e "${RED}✗ GitHub repository doesn't exist or is not accessible (HTTP $http_code)${NC}"
        return 1
      fi
    fi
    
    return 0
  fi
}

# Run all tests
run_tests() {
  failures=0
  
  test_install_script_exists || ((failures++))
  
  # Only run these tests if install script exists
  if [ -f "$INSTALL_SCRIPT" ]; then
    test_install_script_executable || ((failures++))
    test_install_script_syntax || ((failures++))
    test_install_script_urls || ((failures++))
  fi
  
  if [ "$failures" -eq 0 ]; then
    echo -e "\n${GREEN}All installation script tests passed!${NC}"
    return 0
  else
    echo -e "\n${RED}$failures installation script tests failed!${NC}"
    return 1
  fi
}

run_tests 