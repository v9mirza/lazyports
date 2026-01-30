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
    
    # Run go install with GOPROXY=direct to bypass cache lag
    if GOPROXY=direct go install github.com/$REPO@main; then
        echo -e "${GREEN}[SUCCESS] Lazyports installed successfully!${NC}"
        
        # Check if GOBIN is in PATH
        GOBIN=$(go env GOPATH)/bin
        # Always try to install to /usr/local/bin for sudo support
        if [ -w "/usr/local/bin" ] || command -v sudo >/dev/null 2>&1; then
             echo -e "${BLUE}[INFO] Installing to /usr/local/bin for global access (required for sudo)...${NC}"
             if [ -w "/usr/local/bin" ]; then
                 cp "$GOBIN/$BINARY" "/usr/local/bin/$BINARY"
             else
                 echo -e "${BLUE}[sudo] Enter password to move binary to /usr/local/bin:${NC}"
                 sudo cp "$GOBIN/$BINARY" "/usr/local/bin/$BINARY"
             fi
             
             if [ -f "/usr/local/bin/$BINARY" ]; then
                 echo -e "${GREEN}[SUCCESS] Installed to /usr/local/bin/lazyports!${NC}"
             fi
        fi

        if [[ ":$PATH:" != *":$GOBIN:"* ]]; then

             echo -e "${BLUE}[INFO] $GOBIN is not in your PATH.${NC}"
             
             SHELL_CONFIG=""
             if [ -f "$HOME/.bashrc" ]; then
                 SHELL_CONFIG="$HOME/.bashrc"
             elif [ -f "$HOME/.zshrc" ]; then
                 SHELL_CONFIG="$HOME/.zshrc"
             fi

             if [ -n "$SHELL_CONFIG" ]; then
                 echo -e "${BLUE}[+] detected $SHELL_CONFIG. Appending to PATH...${NC}"
                 if grep -q "$GOBIN" "$SHELL_CONFIG"; then
                      echo -e "${GREEN}[NOTE] PATH already configured in $SHELL_CONFIG.${NC}"
                 else
                      echo "" >> "$SHELL_CONFIG"
                      echo "# Added by Lazyports installer" >> "$SHELL_CONFIG"
                      echo "export PATH=\$PATH:$GOBIN" >> "$SHELL_CONFIG"
                      echo -e "${GREEN}[SUCCESS] Added $GOBIN to $SHELL_CONFIG${NC}"
                 fi
                 echo -e "${BLUE}IMPORTANT: Run the following command to load the changes now:${NC}"
                 echo -e "    ${GREEN}source $SHELL_CONFIG${NC}"
             else
                 echo -e "${RED}[WARNING] Could not detect shell config. Manual step required:${NC}"
                 echo "Please add this to your shell config:"
                 echo "  export PATH=\$PATH:$GOBIN"
             fi
        else
             echo "Run 'lazyports' to start the application."
        fi
    else
        echo -e "${RED}[ERROR] Installation failed. Please check your Go installation.${NC}"
        exit 1
    fi
else
    echo -e "${RED}[!] Go is not installed on this system.${NC}"
    echo "This version of Lazyports requires Go to be installed to build from source."
    echo "Please install Go: https://go.dev/doc/install"
    echo "Or check back later for pre-compiled binary releases."
    exit 1
fi
