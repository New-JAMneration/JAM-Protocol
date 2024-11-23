# JAM Protocol Implemetation

![Go Format Check](https://github.com/New-JAMneration/JAM-Protocol/workflows/Go%20Format%20Check/badge.svg)
[![golangci-lint (multi OS)](https://github.com/New-JAMneration/JAM-Protocol/workflows/golangci-lint%20(multi%20OS)/badge.svg)](https://github.com/New-JAMneration/JAM-Protocol/actions?query=workflow%3A"golangci-lint%20(multi%20OS)")
[![golangci-lint-reusable](https://github.com/New-JAMneration/JAM-Protocol/workflows/golangci-lint-reusable/badge.svg)](https://github.com/New-JAMneration/JAM-Protocol/actions?query=workflow%3Agolangci-lint-reusable)


## Documentation of  development is on [Hackmd](https://hackmd.io/8ckvpUULSp-HqThsxXE3jg)

## Remember to install [golangci-lint](https://github.com/golangci/golangci-lint)

## Style stick to [The Google Go Style Guide](https://google.github.io/styleguide/go/)

## Code Formatting

We use `gofmt` to maintain consistent code formatting. Here are the commands you can use:

### Basic Formatting
To format all Go files in the current directory and subdirectories:

```bash
gofmt -w .
```

### Advanced Formatting
To format and simplify Go files (recommended):

```bash
gofmt -s -w .
```

### Check Format Differences
To preview formatting changes without applying them:

```bash
gofmt -d .
```

> Note: The `-w` flag writes changes directly to the files, while `-d` shows the differences without making changes. The `-s` flag enables additional code simplification.

### CI/CD Integration
This project uses GitHub Actions to automatically check code formatting. The checks will run on every pull request and push to the main branch.

To ensure your code passes the CI checks:
1. Run `gofmt -s -w .` before committing
2. Check for any remaining issues with `gofmt -d .`

## Commit massage stick to [here](semantic-commit-messages.md)
