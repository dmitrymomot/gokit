#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HOOKS_DIR="$SCRIPT_DIR/git-hooks"
GIT_HOOKS_DIR="$(git rev-parse --git-dir)/hooks"

echo "Setting up Git hooks..."

# Check if git-hooks directory exists
if [ ! -d "$HOOKS_DIR" ]; then
  echo "Error: Git hooks directory not found at $HOOKS_DIR"
  exit 1
fi

# Create the hooks directory if it doesn't exist
mkdir -p "$GIT_HOOKS_DIR"

# Copy all hooks
for hook in "$HOOKS_DIR"/*; do
  if [ -f "$hook" ]; then
    hook_name=$(basename "$hook")
    target="$GIT_HOOKS_DIR/$hook_name"
    
    echo "Installing $hook_name hook..."
    cp "$hook" "$target"
    chmod +x "$target"
    echo "$hook_name hook installed successfully."
  fi
done

echo "Git hooks setup complete. The following hooks are now active:"
ls -la "$GIT_HOOKS_DIR"

# Check if goimports is installed
if ! command -v goimports >/dev/null 2>&1; then
  echo "Warning: goimports not found."
  echo "For best results, install goimports with:"
  echo "  go install golang.org/x/tools/cmd/goimports@latest"
fi

# Check if golangci-lint is installed
if ! command -v golangci-lint >/dev/null 2>&1; then
  echo "Warning: golangci-lint not found."
  echo "For pre-push validation, install golangci-lint with:"
  echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
fi

echo "Setup complete!"
