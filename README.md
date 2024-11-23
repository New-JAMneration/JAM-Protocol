# JAM Protocol Implemetation

![Go Format Check](https://github.com/New-JAMneration/JAM-Protocol/workflows/Go%20Format%20Check/badge.svg)
[![golangci-lint (multi OS)](https://github.com/New-JAMneration/JAM-Protocol/workflows/golangci-lint%20(multi%20OS)/badge.svg)](https://github.com/New-JAMneration/JAM-Protocol/actions?query=workflow%3A"golangci-lint%20(multi%20OS)")
[![golangci-lint-reusable](https://github.com/New-JAMneration/JAM-Protocol/workflows/golangci-lint-reusable/badge.svg)](https://github.com/New-JAMneration/JAM-Protocol/actions?query=workflow%3Agolangci-lint-reusable)


## Documentation

Our development documentation is maintained on HackMD for real-time collaboration and easy updates.

### Access Documentation
- Main documentation: [HackMD Development Guide](https://hackmd.io/8ckvpUULSp-HqThsxXE3jg)
- Requires team member access - please contact project maintainers if you need access

## Coding Style

Our codebase follows the [Google Go Style Guide](https://google.github.io/styleguide/go/) for consistent and maintainable code.

We enforce code quality using [golangci-lint](https://github.com/golangci/golangci-lint) in our CI/CD pipeline. To ensure smooth development:

- Install and configure golangci-lint in your local environment
- Run lint checks before committing code
- Fix any lint issues to prevent CI/CD pipeline failures

### Quick Setup for golangci-lint

1. Installation:

```bash
# macOS
brew install golangci-lint

# Windows
scoop install golangci-lint

# Linux/Ubuntu
Golangci-lint is available inside the majority of the package managers.

# Using Go (all platforms)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

2. Run lint check in project root:
```bash
golangci-lint run
```

Following these guidelines helps maintain code quality and ensures consistency across the project.

## Code Formatting

We use `gofmt` to maintain consistent code formatting. [Here](./READMERef/code-formatting.md) are the commands you can use.


## Commit massage
Please stick to [here](./READMERef/semantic-commit-messages.md) when you are going to submit a commit.
