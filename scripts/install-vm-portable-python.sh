#!/bin/bash

# VM Installation Script for NIVAI Project (Portable Python Version)
# This script uses portable Python builds that don't require compilation

set -e  # Exit on any error

echo "===================================="
echo "NIVAI VM Portable Installation (No Sudo)"
echo "===================================="

# Set up local installation directories
export LOCAL_PREFIX="$HOME/.local"
export LOCAL_BIN="$LOCAL_PREFIX/bin"
mkdir -p "$LOCAL_BIN"

# Detect architecture
ARCH=$(uname -m)
if [[ "$ARCH" == "x86_64" ]]; then
    ARCH="x86_64"
elif [[ "$ARCH" == "aarch64" ]]; then
    ARCH="aarch64"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

echo "Detected architecture: $ARCH"

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Update PATH for local installations
export PATH="$LOCAL_BIN:$PATH"

# Add to .bashrc if not already present
if ! grep -q "$LOCAL_BIN" ~/.bashrc 2>/dev/null; then
    echo "export PATH=\"$LOCAL_BIN:\$PATH\"" >> ~/.bashrc
fi

# Install Go 1.22.5 locally
echo -e "\n1. Installing Go 1.22.5..."
GO_INSTALL_DIR="$HOME/.local/go"
if [[ ! -f "$GO_INSTALL_DIR/bin/go" ]]; then
    echo "Downloading and installing Go 1.22.5..."
    # Fix architecture naming for Go downloads
    GO_ARCH=$ARCH
    if [[ "$ARCH" == "x86_64" ]]; then
        GO_ARCH="amd64"
    elif [[ "$ARCH" == "aarch64" ]]; then
        GO_ARCH="arm64"
    fi

    wget "https://go.dev/dl/go1.22.5.linux-${GO_ARCH}.tar.gz" -O /tmp/go1.22.5.tar.gz
    mkdir -p "$HOME/.local"
    tar -C "$HOME/.local" -xzf /tmp/go1.22.5.tar.gz
    rm /tmp/go1.22.5.tar.gz
else
    echo "Go is already installed"
fi

# Add Go to PATH
export PATH="$GO_INSTALL_DIR/bin:$PATH"
export GOPATH="$HOME/go"
export PATH="$PATH:$GOPATH/bin"

# Update .bashrc with Go paths
if ! grep -q "$GO_INSTALL_DIR/bin" ~/.bashrc 2>/dev/null; then
    {
        echo "export PATH=\"$GO_INSTALL_DIR/bin:\$PATH\""
        echo "export GOPATH=\"$HOME/go\""
        echo "export PATH=\"\$PATH:\$GOPATH/bin\""
    } >> ~/.bashrc
fi

# Install Node.js 20.15.1 using prebuilt binaries
echo -e "\n2. Installing Node.js 20.15.1..."
NODE_DIR="$HOME/.local/node"
if [[ ! -f "$NODE_DIR/bin/node" ]]; then
    echo "Downloading Node.js 20.15.1 prebuilt binaries..."
    NODE_ARCH=$ARCH
    if [[ "$ARCH" == "x86_64" ]]; then
        NODE_ARCH="x64"
    elif [[ "$ARCH" == "aarch64" ]]; then
        NODE_ARCH="arm64"
    fi

    wget "https://nodejs.org/dist/v20.15.1/node-v20.15.1-linux-${NODE_ARCH}.tar.xz" -O /tmp/node.tar.xz
    mkdir -p "$NODE_DIR"
    tar -xJf /tmp/node.tar.xz -C "$NODE_DIR" --strip-components=1
    rm /tmp/node.tar.xz
else
    echo "Node.js is already installed"
fi

# Add Node to PATH
export PATH="$NODE_DIR/bin:$PATH"
if ! grep -q "$NODE_DIR/bin" ~/.bashrc 2>/dev/null; then
    echo "export PATH=\"$NODE_DIR/bin:\$PATH\"" >> ~/.bashrc
fi

# Install portable Python using python-build-standalone
echo -e "\n3. Installing Python 3.13 (portable)..."
PYTHON_DIR="$HOME/.local/python"
if [[ ! -f "$PYTHON_DIR/bin/python3.13" ]]; then
    echo "Downloading portable Python 3.13..."

    # Use python-build-standalone project for portable Python
    PYTHON_URL="https://github.com/indygreg/python-build-standalone/releases/download/20241016/cpython-3.13.0+20241016-${ARCH}-unknown-linux-gnu-install_only.tar.gz"

    # Try to download, fall back to Python 3.12 if 3.13 is not available
    if ! wget "$PYTHON_URL" -O /tmp/python.tar.gz 2>/dev/null; then
        echo "Python 3.13 not available for this architecture, trying Python 3.12..."
        PYTHON_URL="https://github.com/indygreg/python-build-standalone/releases/download/20241016/cpython-3.12.7+20241016-${ARCH}-unknown-linux-gnu-install_only.tar.gz"
        wget "$PYTHON_URL" -O /tmp/python.tar.gz
        PYTHON_VERSION="3.12"
    else
        PYTHON_VERSION="3.13"
    fi

    mkdir -p "$PYTHON_DIR"
    tar -xzf /tmp/python.tar.gz -C "$PYTHON_DIR" --strip-components=1
    rm /tmp/python.tar.gz

    # Create symlinks for convenience
    ln -sf "$PYTHON_DIR/bin/python${PYTHON_VERSION}" "$PYTHON_DIR/bin/python3"
    ln -sf "$PYTHON_DIR/bin/python${PYTHON_VERSION}" "$PYTHON_DIR/bin/python"
    ln -sf "$PYTHON_DIR/bin/pip${PYTHON_VERSION}" "$PYTHON_DIR/bin/pip3"
    ln -sf "$PYTHON_DIR/bin/pip${PYTHON_VERSION}" "$PYTHON_DIR/bin/pip"
else
    echo "Python is already installed"
fi

# Add Python to PATH
export PATH="$PYTHON_DIR/bin:$PATH"
if ! grep -q "$PYTHON_DIR/bin" ~/.bashrc 2>/dev/null; then
    echo "export PATH=\"$PYTHON_DIR/bin:\$PATH\"" >> ~/.bashrc
fi

# Upgrade pip
echo "Upgrading pip..."
"$PYTHON_DIR/bin/python" -m pip install --upgrade pip

# Install Poetry 2.1.1
echo -e "\n4. Installing Poetry 2.1.1..."
if ! "$PYTHON_DIR/bin/pip" show poetry | grep -q "Version: 2.1.1"; then
    echo "Installing Poetry 2.1.1..."
    "$PYTHON_DIR/bin/pip" install poetry==2.1.1
else
    echo "Poetry 2.1.1 is already installed"
fi

# Create convenience scripts
echo -e "\n5. Creating convenience scripts..."

# Create wrapper scripts in LOCAL_BIN for easier access
cat > "$LOCAL_BIN/go" << EOF
#!/bin/bash
exec "$GO_INSTALL_DIR/bin/go" "\$@"
EOF
chmod +x "$LOCAL_BIN/go"

cat > "$LOCAL_BIN/node" << EOF
#!/bin/bash
exec "$NODE_DIR/bin/node" "\$@"
EOF
chmod +x "$LOCAL_BIN/node"

cat > "$LOCAL_BIN/npm" << EOF
#!/bin/bash
exec "$NODE_DIR/bin/npm" "\$@"
EOF
chmod +x "$LOCAL_BIN/npm"

cat > "$LOCAL_BIN/python3" << EOF
#!/bin/bash
exec "$PYTHON_DIR/bin/python3" "\$@"
EOF
chmod +x "$LOCAL_BIN/python3"

cat > "$LOCAL_BIN/python3.13" << EOF
#!/bin/bash
exec "$PYTHON_DIR/bin/python3" "\$@"
EOF
chmod +x "$LOCAL_BIN/python3.13"

cat > "$LOCAL_BIN/pip" << EOF
#!/bin/bash
exec "$PYTHON_DIR/bin/pip" "\$@"
EOF
chmod +x "$LOCAL_BIN/pip"

cat > "$LOCAL_BIN/poetry" << EOF
#!/bin/bash
exec "$PYTHON_DIR/bin/poetry" "\$@"
EOF
chmod +x "$LOCAL_BIN/poetry"

# Verify installations
echo -e "\n===================================="
echo "Verification of installed versions:"
echo "===================================="

echo -n "Go: "
"$GO_INSTALL_DIR/bin/go" version || echo "Not installed"

echo -n "Node.js: "
"$NODE_DIR/bin/node" --version || echo "Not installed"

echo -n "npm: "
"$NODE_DIR/bin/npm" --version || echo "Not installed"

echo -n "Python: "
"$PYTHON_DIR/bin/python" --version || echo "Not installed"

echo -n "Poetry: "
"$PYTHON_DIR/bin/poetry" --version || echo "Not installed"

echo -e "\n===================================="
echo "Installation complete!"
echo "===================================="

echo -e "\nIMPORTANT: Run this command to update your current shell:"
echo "source ~/.bashrc"
echo ""
echo "Or add this line to your current shell:"
echo "export PATH=\"$LOCAL_BIN:\$PATH\""
echo ""
echo -e "\nNext steps:"
echo "1. Source your .bashrc: source ~/.bashrc"
echo "2. Clone the NIVAI repository"
echo "3. Navigate to each component directory and install dependencies:"
echo "   - Backend: cd backend && go mod download"
echo "   - Frontend: cd frontend && npm install"
echo "   - Python API: cd python_api && poetry install"
echo ""
echo "All tools are installed in:"
echo "  - Go: $GO_INSTALL_DIR"
echo "  - Node.js: $NODE_DIR"
echo "  - Python: $PYTHON_DIR"
echo "  - Convenience scripts: $LOCAL_BIN"