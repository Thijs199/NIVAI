#!/bin/sh

# scripts/backend-pre-commit.sh

# Get the list of staged Go files
STAGED_GO_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$')

if [ -z "$STAGED_GO_FILES" ]; then
  exit 0
fi

echo "Running gofmt on staged Go files..."
echo "$STAGED_GO_FILES" | xargs gofmt -w

# Check if gofmt made any changes that need to be re-staged
# git diff --quiet will exit 0 if no changes, 1 if changes
if ! git diff --quiet -- $STAGED_GO_FILES; then
    echo "gofmt made changes. Please review and stage them."
    # Re-add the files modified by gofmt
    echo "$STAGED_GO_FILES" | xargs git add
    exit 1
fi

echo "Running golangci-lint on staged Go files..."
# It's often better to run golangci-lint on the specific files
# if your project is large, but ./... is also common.
# For staged files, you can pass them as arguments.
# However, ensure golangci-lint is installed and configured.
# This script assumes golangci-lint is installed in the environment.
# If not, the CI should handle installation. For local hooks,
# developers might need to install it manually or via a project setup script.

# Filter for files within the 'backend' directory
BACKEND_STAGED_GO_FILES=$(echo "$STAGED_GO_FILES" | grep '^backend/')

if [ -n "$BACKEND_STAGED_GO_FILES" ]; then
    # Temporarily change to backend directory to ensure golangci-lint runs with correct context
    # (e.g., finding .golangci.yml if it exists there)
    CURRENT_DIR=$(pwd)
    cd backend || exit 1 # Exit if cd fails

    echo "Linting files in backend directory:"
    # Use sed to remove 'backend/' prefix for golangci-lint if it's running inside 'backend'
    # and expects paths relative to 'backend'.
    # The path /go/bin/golangci-lint is based on its installation in the Dockerfile.
    # For local use, developers might have it elsewhere.
    # Consider making this path configurable or relying on it being in PATH.
    # Using --issues-exit-code=0 so linting errors don't stop the script here,
    # we capture the exit code and handle it.
    # Using --fix to auto-fix issues.

    # Prepare file list relative to backend directory
    RELATIVE_BACKEND_FILES=$(echo "$BACKEND_STAGED_GO_FILES" | sed 's_backend/ __')

    /go/bin/golangci-lint run --fix --issues-exit-code=1 --out-format=line-number -- $RELATIVE_BACKEND_FILES
    LINT_RESULT=$?

    cd "$CURRENT_DIR" || exit 1 # Return to original directory

    if [ $LINT_RESULT -ne 0 ]; then
        echo "golangci-lint found issues. Please fix them."
        # Check again if linting made changes that need to be staged
        # This checks the original $BACKEND_STAGED_GO_FILES paths from repo root
        if ! git diff --quiet -- $BACKEND_STAGED_GO_FILES; then
             echo "golangci-lint --fix made changes. Please review and stage them."
             echo "$BACKEND_STAGED_GO_FILES" | xargs git add
        fi
        exit 1
    fi
else
    echo "No staged Go files found in the 'backend' directory."
fi

echo "Backend pre-commit checks passed."
exit 0
