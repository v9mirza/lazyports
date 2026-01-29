#!/bin/bash

set -e

REPO="v9mirza/lazyports"
BINARY="lazyports"

# ANSI Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}===============================${NC}"
echo -e "${BLUE}   Installing Lazyports...     ${NC}"
echo -e "${BLUE}===============================${NC}"

# Check if Go is installed
if command -v go >/dev/null 2>&1; then
    echo -e "${GREEN}[+] Go detected. Installing via go install...${NC}"
    
    # Run go install
    if go install github.com/$REPO@latest; then
        echo -e "${GREEN}[SUCCESS] Lazyports installed successfully!${NC}"
        
        # Check if GOBIN is in PATH
        GOBIN=$(go env GOPATH)/bin
        if [[ ":$PATH:" != *":$GOBIN:"* ]]; then
             echo -e "${RED}[WARNING] $GOBIN is not in your PATH.${NC}"
             echo "Please add it to your shell config (e.g. ~/.bashrc):"
             echo "  export PATH=\$PATH:$GOBIN"
             echo "Then try running: lazyports"
        else
             echo "Run 'lazyports' to start the application."
        fi
    else
        echo -e "${RED}[ERROR] Installation failed.${NC}"
        exit 1
    fi
else
    echo -e "${RED}[!] Go is not installed on this system.${NC}"
    echo "This version of Lazyports requires Go to be installed to build from source."
    echo "Please install Go: https://go.dev/doc/install"
    echo "Or check back later for pre-compiled binary releases."
    exit 1
fi
