# Pre-commit Hook Setup Guide

This repository uses pre-commit hooks to ensure code quality and consistent formatting across all file types.

## What's Included

The pre-commit hooks automatically format and lint:
- **Go files**: `gofmt` and `golangci-lint`
- **Python files**: `black`, `isort`, and `ruff`
- **TypeScript/JavaScript files**: `ESLint` and `Prettier` (via husky/lint-staged in frontend)

## Installation

### Option 1: Automatic Setup (Recommended)

The repository includes a comprehensive pre-commit hook that's already installed at `.git/hooks/pre-commit`.

### Option 2: Using pre-commit framework

1. Install pre-commit:
   ```bash
   pip install pre-commit
   ```

2. Install the hooks:
   ```bash
   pre-commit install
   ```

### Required Tools

Make sure you have the following tools installed:

#### For Go development

```bash
# Install golangci-lint (optional - for additional linting beyond gofmt)
# Note: If you have version compatibility issues, the hook will skip golangci-lint
# and use only gofmt, which is sufficient for basic formatting
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
```

#### For Python development

```bash
cd python_api
pip install black isort ruff
```

#### For Frontend development

```bash
cd frontend
npm install
```

## How It Works

When you commit files, the pre-commit hook will:

1. Detect which types of files you're committing
2. Run the appropriate formatters and linters
3. Automatically fix formatting issues when possible
4. Re-stage the fixed files
5. Proceed with the commit if all checks pass

## Manual Formatting

You can also manually format files:

### Go files

```bash
gofmt -w backend/
cd backend && golangci-lint run --fix
```

### Python files

```bash
cd python_api
black src/ tests/
isort src/ tests/
ruff check src/ tests/ --fix
```

### TypeScript/JavaScript files

```bash
cd frontend
npm run lint -- --fix
npx prettier --write "src/**/*.{js,jsx,ts,tsx}"
```

## Troubleshooting

If pre-commit hooks are not running:

1. Check that the hook is executable: `chmod +x .git/hooks/pre-commit`
2. Ensure required tools are installed (see Required Tools section)
3. For frontend, ensure `node_modules` is installed: `cd frontend && npm install`

## Disabling Hooks (Not Recommended)

If you need to bypass hooks temporarily:

```bash
git commit --no-verify
```

**Note**: This should only be used in exceptional circumstances. Always ensure your code is properly formatted before pushing.
