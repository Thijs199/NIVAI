# Developer Guidelines

This document provides guidelines for developers working on this project to ensure code quality, consistency, and smooth collaboration.

## Table of Contents

- [IDE Integration](#ide-integration)
- [Dependency Management](#dependency-management)
- [Pre-commit Hooks](#pre-commit-hooks)
  - [Frontend](#frontend)
  - [Backend](#backend)
- [Continuous Integration (CI)](#continuous-integration-ci)
  - [Frontend CI Workflow](#frontend-ci-workflow)
  - [Backend CI Workflow](#backend-ci-workflow)
- [Code Review Checklist](#code-review-checklist)

## IDE Integration

To maintain code style and catch errors early, please configure your IDE to use the project's linters and formatters.

### VS Code Recommendations

Install the following extensions:

*   **ESLint:** (dbaeumer.vscode-eslint) - For JavaScript/TypeScript linting in the frontend.
*   **Prettier - Code formatter:** (esbenp.prettier-vscode) - For code formatting in the frontend.
*   **Go:** (golang.go) - For Go language support, including formatting and linting, for the backend.

**Recommended VS Code Settings (add to your `.vscode/settings.json`):**

```json
{
  // Frontend specific
  "[javascript]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode",
    "editor.formatOnSave": true
  },
  "[javascriptreact]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode",
    "editor.formatOnSave": true
  },
  "[typescript]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode",
    "editor.formatOnSave": true
  },
  "[typescriptreact]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode",
    "editor.formatOnSave": true
  },
  "eslint.validate": [
    "javascript",
    "javascriptreact",
    "typescript",
    "typescriptreact"
  ],
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": "explicit"
  },
  // Backend specific
  "[go]": {
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.organizeImports": "explicit"
    },
    "editor.defaultFormatter": "golang.go" // Ensures use of the Go extension's formatter
  },
  "go.lintTool": "golangci-lint",
  "go.lintFlags": [
    "--fast"
  ],
  "go.useLanguageServer": true
}
```

## Dependency Management

### Frontend (Node.js/npm)

*   Always add new dependencies to `frontend/package.json` using `npm install <package-name>` or `npm install --save-dev <package-name>`.
*   Ensure the `frontend/package-lock.json` is updated and committed with any dependency changes.
*   Regularly run `npm audit` (from the `frontend` directory) to check for vulnerabilities and apply fixes if necessary using `npm audit fix`.

### Backend (Go Modules)

*   Backend dependencies are managed using Go Modules. The `backend/go.mod` and `backend/go.sum` files handle this.
*   To check for available updates for direct and indirect dependencies, run `go list -u -m all` from the `backend` directory.
*   Update dependencies using `go get <module-path>@<version>` or `go get -u <module-path>`.

## Pre-commit Hooks

Pre-commit hooks are set up to automatically format and lint your code before you commit.

### Frontend

The frontend uses Husky and lint-staged.

1.  **Installation:** After cloning the repository, navigate to the `frontend` directory and run `npm install`. This should install Husky.
2.  **Activation (if needed):** If hooks are not running, you might need to run `npx husky install` from the `frontend` directory once.
3.  **Functionality:** On commit, `eslint --fix` and `prettier --write` will run on staged JavaScript/TypeScript files.

### Backend

A custom Git hook script is used for the backend.

1.  **Script Location:** `scripts/backend-pre-commit.sh`
2.  **Functionality:** This script runs `gofmt -w` (formats Go code) and `golangci-lint run` (lints Go code) on staged `.go` files within the `backend` directory.
3.  **Manual Setup:** To enable this hook, you need to link or copy it to your local `.git/hooks/` directory:
    ```bash
    # From the root of the repository:
    ln -s -f ../../scripts/backend-pre-commit.sh .git/hooks/pre-commit
    # Or, if your team uses a shared Git hooks management tool, follow its instructions.
    ```
    Ensure the script is executable: `chmod +x .git/hooks/pre-commit` (or `chmod +x scripts/backend-pre-commit.sh`).

## Continuous Integration (CI)

CI pipelines are set up using GitHub Actions to automatically build, lint, and test the code on pushes and pull requests.

**(Note: Due to sandbox environment limitations, the following workflow files could not be created automatically. Please create them manually in your `.github/workflows/` directory with the content provided.)**

### Frontend CI Workflow

**File:** `.github/workflows/frontend-ci.yml`

```yaml
name: Frontend CI

on:
  push:
    branches:
      - main
      - develop # Add other relevant branches
    paths: # Only run if frontend code changes
      - 'frontend/**'
      - '.github/workflows/frontend-ci.yml'
  pull_request:
    branches:
      - main
      - develop
    paths:
      - 'frontend/**'
      - '.github/workflows/frontend-ci.yml'

jobs:
  build-and-test:
    name: Build and Test Frontend
    runs-on: ubuntu-latest

    strategy:
      matrix:
        node-version: [18.x, 20.x] # You can specify Node.js versions to test against

    defaults:
      run:
        working-directory: ./frontend

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js ${{ matrix.node-version }}
        uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node-version }}
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json

      - name: Install dependencies
        run: npm ci

      - name: Run linter
        run: npm run lint

      # If you have tests, add a step to run them:
      # - name: Run tests
      #   run: npm test

      - name: Build application
        run: npm run build
```

### Backend CI Workflow

**File:** `.github/workflows/backend-ci.yml`

```yaml
name: Backend CI

on:
  push:
    branches:
      - main
      - develop # Add other relevant branches
    paths: # Only run if backend code changes
      - 'backend/**'
      - '.github/workflows/backend-ci.yml'
  pull_request:
    branches:
      - main
      - develop
    paths:
      - 'backend/**'
      - '.github/workflows/backend-ci.yml'

jobs:
  build-lint-and-test:
    name: Build, Lint, and Test Backend
    runs-on: ubuntu-latest

    defaults:
      run:
        working-directory: ./backend

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22' # Specify your Go version

      - name: Install golangci-lint
        run: |
          curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.58.1
          # Replace v1.58.1 with your desired or latest golangci-lint version
          echo "$(go env GOPATH)/bin" >> $GITHUB_PATH
        # This installs golangci-lint to a directory that's then added to PATH

      - name: Download Go module dependencies
        run: go mod download

      - name: Run gofmt check
        run: |
          # List files that are not formatted
          FMT_FILES=$(gofmt -l .)
          if [ -n "$FMT_FILES" ]; then
            echo "Go files are not formatted. Please run gofmt:"
            echo "$FMT_FILES"
            exit 1
          fi
          echo "All Go files are formatted."

      - name: Run golangci-lint
        run: golangci-lint run ./... --timeout=5m # Adjust timeout as needed

      # If you have tests, add a step to run them:
      # - name: Run Go tests
      #   run: go test -v ./...

      - name: Build application
        run: go build -v ./cmd/api/... # Build the main application binary/binaries
```

## Code Review Checklist

While automated tools catch many issues, human review is crucial. Consider the following:

*   **Functionality:** Does the code meet the requirements of the task/issue?
*   **Readability:** Is the code clear, concise, and easy to understand? Are variable and function names meaningful?
*   **Error Handling:** Are errors handled gracefully? Are there any unhandled edge cases?
*   **Performance:** Are there any obvious performance bottlenecks?
*   **Security:** Does the code introduce any security vulnerabilities? (e.g., SQL injection, XSS)
*   **Test Coverage:** Are there sufficient tests for the new code? Do existing tests pass?
*   **Adherence to Guidelines:** Does the code follow these developer guidelines and project-specific conventions?
*   **Documentation:** Is new code adequately commented? Is any relevant external documentation (e.g., READMEs, API docs) updated?

---
By following these guidelines, we can build a more robust, maintainable, and high-quality application.
