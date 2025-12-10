# Contributing to Shelly CLI

Thank you for your interest in contributing to Shelly CLI! This document provides guidelines and instructions for contributing.

## Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## How to Contribute

### Reporting Bugs

Before submitting a bug report:
1. Check existing issues to avoid duplicates
2. Use the latest version of the CLI
3. Collect relevant information (OS, Go version, device type, etc.)

When submitting a bug report, include:
- Clear, descriptive title
- Steps to reproduce the issue
- Expected vs actual behavior
- CLI version (`shelly version`)
- Device information (if applicable)
- Relevant logs or error messages

### Suggesting Features

Feature requests are welcome! Please:
1. Check existing issues for similar suggestions
2. Describe the problem the feature would solve
3. Explain the proposed solution
4. Consider alternative approaches

### Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Run linter (`make lint`)
6. Commit with clear messages (see below)
7. Push to your fork
8. Open a Pull Request

## Development Setup

### Prerequisites

- Go 1.25.5 or later
- golangci-lint v2.7.1 or later
- Make

### Building

```bash
# Clone the repository
git clone https://github.com/tj-smith47/shelly-cli.git
cd shelly-cli

# Install dependencies
go mod download

# Build
make build

# Run tests
make test

# Run linter
make lint
```

### Project Structure

```
shelly-cli/
├── cmd/shelly/          # Entry point
├── internal/            # Private packages
│   ├── cli/             # Command implementations
│   ├── tui/             # TUI components
│   ├── config/          # Configuration management
│   ├── output/          # Output formatting
│   ├── plugins/         # Plugin system
│   └── version/         # Version info
├── pkg/api/             # Public API for plugins
├── docs/                # Documentation
└── examples/            # Example configurations
```

## Coding Standards

### Go Style

- Follow standard Go formatting (`gofmt`)
- Pass all golangci-lint checks
- Use meaningful variable and function names
- Write clear comments for exported types and functions
- Keep functions focused and reasonably sized

### Testing

- Write tests for new functionality
- Maintain 90%+ code coverage
- Use table-driven tests where appropriate
- Test edge cases and error conditions

### Commit Messages

Use conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Formatting, no code change
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:
```
feat(discovery): add BLE device discovery
fix(config): handle missing config file gracefully
docs(readme): update installation instructions
```

### Pull Request Guidelines

- Keep PRs focused on a single change
- Update documentation as needed
- Add tests for new functionality
- Ensure CI passes before requesting review
- Respond to review feedback promptly

## Getting Help

- Open an issue for questions
- Check existing documentation
- Join discussions in GitHub Discussions

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
