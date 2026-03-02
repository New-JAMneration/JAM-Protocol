# Code Formatting

## Basic Formatting
To format all Go files in the current directory and subdirectories:

```bash
gofmt -w .
```

## Advanced Formatting
To format and simplify Go files (recommended):

```bash
gofmt -s -w .
```

## Check Format Differences
To preview formatting changes without applying them:

```bash
gofmt -d .
```

> Note: The `-w` flag writes changes directly to the files, while `-d` shows the differences without making changes. The `-s` flag enables additional code simplification.

## CI/CD Integration
This project uses GitHub Actions to automatically check code formatting. The checks will run on every pull request and push to the main branch.

To ensure your code passes the CI checks:
1. Run `gofmt -s -w .` before committing
2. Check for any remaining issues with `gofmt -d .`