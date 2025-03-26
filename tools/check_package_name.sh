#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Package name to check
PACKAGE_NAME="insta"

echo "Checking package name '$PACKAGE_NAME' across different package managers..."

# Check npm
echo -e "\n${YELLOW}Checking npm...${NC}"
if npm search "$PACKAGE_NAME" --json | grep -q "\"name\": \"$PACKAGE_NAME\""; then
    echo -e "${RED}❌ Package name '$PACKAGE_NAME' is taken on npm${NC}"
else
    echo -e "${GREEN}✅ Package name '$PACKAGE_NAME' is available on npm${NC}"
fi

# Check PyPI
echo -e "\n${YELLOW}Checking PyPI...${NC}"
if curl -s "https://pypi.org/pypi/$PACKAGE_NAME/json" > /dev/null; then
    echo -e "${RED}❌ Package name '$PACKAGE_NAME' is taken on PyPI${NC}"
else
    echo -e "${GREEN}✅ Package name '$PACKAGE_NAME' is available on PyPI${NC}"
fi

# Check RubyGems
echo -e "\n${YELLOW}Checking RubyGems...${NC}"
if curl -s "https://rubygems.org/api/v1/gems/$PACKAGE_NAME.json" > /dev/null; then
    echo -e "${RED}❌ Package name '$PACKAGE_NAME' is taken on RubyGems${NC}"
else
    echo -e "${GREEN}✅ Package name '$PACKAGE_NAME' is available on RubyGems${NC}"
fi

# Check Homebrew
echo -e "\n${YELLOW}Checking Homebrew...${NC}"
if brew search "$PACKAGE_NAME" | grep -q "^$PACKAGE_NAME$"; then
    echo -e "${RED}❌ Package name '$PACKAGE_NAME' is taken on Homebrew${NC}"
else
    echo -e "${GREEN}✅ Package name '$PACKAGE_NAME' is available on Homebrew${NC}"
fi

# Check Chocolatey
echo -e "\n${YELLOW}Checking Chocolatey...${NC}"
if curl -s "https://chocolatey.org/packages/$PACKAGE_NAME" | grep -q "404 Not Found"; then
    echo -e "${GREEN}✅ Package name '$PACKAGE_NAME' is available on Chocolatey${NC}"
else
    echo -e "${RED}❌ Package name '$PACKAGE_NAME' is taken on Chocolatey${NC}"
fi

# Check AUR (Arch User Repository)
echo -e "\n${YELLOW}Checking AUR...${NC}"
if curl -s "https://aur.archlinux.org/rpc/?v=5&type=search&by=name&arg=$PACKAGE_NAME" | grep -q "\"Name\":\"$PACKAGE_NAME\""; then
    echo -e "${RED}❌ Package name '$PACKAGE_NAME' is taken on AUR${NC}"
else
    echo -e "${GREEN}✅ Package name '$PACKAGE_NAME' is available on AUR${NC}"
fi

# Check Debian/Ubuntu repositories
echo -e "\n${YELLOW}Checking Debian/Ubuntu repositories...${NC}"
if curl -s "https://packages.debian.org/search?keywords=$PACKAGE_NAME&searchon=names&suite=stable&section=all" | grep -q "Package: $PACKAGE_NAME"; then
    echo -e "${RED}❌ Package name '$PACKAGE_NAME' is taken in Debian/Ubuntu repositories${NC}"
else
    echo -e "${GREEN}✅ Package name '$PACKAGE_NAME' is available in Debian/Ubuntu repositories${NC}"
fi

# Check Fedora repositories
echo -e "\n${YELLOW}Checking Fedora repositories...${NC}"
if curl -s "https://apps.fedoraproject.org/packages/$PACKAGE_NAME" | grep -q "404 Not Found"; then
    echo -e "${GREEN}✅ Package name '$PACKAGE_NAME' is available in Fedora repositories${NC}"
else
    echo -e "${RED}❌ Package name '$PACKAGE_NAME' is taken in Fedora repositories${NC}"
fi

# Check GitHub repositories
echo -e "\n${YELLOW}Checking GitHub repositories...${NC}"
if curl -s "https://api.github.com/repos/$PACKAGE_NAME" | grep -q "\"name\": \"$PACKAGE_NAME\""; then
    echo -e "${RED}❌ Repository name '$PACKAGE_NAME' is taken on GitHub${NC}"
else
    echo -e "${GREEN}✅ Repository name '$PACKAGE_NAME' is available on GitHub${NC}"
fi

echo -e "\n${YELLOW}Note:${NC} This script checks for exact matches only. You might want to also check for similar names to avoid confusion." 