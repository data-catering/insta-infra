#!/usr/bin/env bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Installing insta-infra...${NC}"

# Check if git is installed
if ! command -v git &>/dev/null; then
  echo -e "${RED}Error: git is required but could not be found${NC}"
  exit 1
fi

# Check if docker is installed
if ! command -v docker &>/dev/null; then
  echo -e "${YELLOW}Warning: docker is required to use insta-infra but could not be found${NC}"
  echo -e "${YELLOW}Please install Docker from https://docs.docker.com/get-docker/${NC}"
fi

# Check if docker-compose is installed
if ! command -v docker compose &>/dev/null; then
  echo -e "${YELLOW}Warning: docker compose is required to use insta-infra but could not be found${NC}"
  echo -e "${YELLOW}Please install Docker Compose from https://docs.docker.com/compose/install/${NC}"
fi

# Default installation directory
DEFAULT_DIR="$HOME/.insta-infra"
INSTALL_DIR=${INSTALL_DIR:-$DEFAULT_DIR}

# Clone the repository
if [ -d "$INSTALL_DIR" ]; then
  echo -e "${YELLOW}Found existing installation at $INSTALL_DIR, updating...${NC}"
  cd "$INSTALL_DIR"
  git pull
else
  echo -e "${GREEN}Cloning insta-infra to $INSTALL_DIR...${NC}"
  git clone https://github.com/data-catering/insta-infra.git "$INSTALL_DIR"
fi

# Make run.sh executable
chmod +x "$INSTALL_DIR/run.sh"

# Create symbolic link in /usr/local/bin if possible
if [ -d "/usr/local/bin" ] && [ -w "/usr/local/bin" ]; then
  echo -e "${GREEN}Creating symbolic link in /usr/local/bin...${NC}"
  ln -sf "$INSTALL_DIR/run.sh" /usr/local/bin/insta
  echo -e "${GREEN}Installation complete! You can now use insta-infra by running 'insta'${NC}"
else
  echo -e "${YELLOW}Could not create symbolic link in /usr/local/bin (permission denied)${NC}"
  echo -e "${YELLOW}Add the following line to your shell configuration file (.bashrc, .zshrc, etc.):${NC}"
  echo -e "${GREEN}alias insta=$INSTALL_DIR/run.sh${NC}"
  echo -e "${YELLOW}Then run 'source ~/.bashrc' or 'source ~/.zshrc' to apply the changes${NC}"
fi

echo -e "${GREEN}To get started, run:${NC}"
echo -e "${GREEN}insta -l${NC} ${YELLOW}# List available services${NC}"
echo -e "${GREEN}insta postgres${NC} ${YELLOW}# Start a PostgreSQL instance${NC}"
echo -e "${GREEN}insta -c postgres${NC} ${YELLOW}# Connect to the PostgreSQL instance${NC}"
echo -e "${GREEN}insta -d${NC} ${YELLOW}# Shut down all services${NC}" 