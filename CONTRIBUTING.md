# Contributing to FastApp

We love your input! We want to make contributing to FastApp as easy and transparent as possible, whether it's:

- Reporting a bug
- Discussing the current state of the code
- Submitting a fix
- Proposing new features
- Becoming a maintainer

## Development Process

We use GitHub to host code, to track issues and feature requests, as well as accept pull requests.

## Pull Requests

Pull requests are the best way to propose changes to the codebase. We actively welcome your pull requests:

1. Fork the repo and create your branch from `main`.
2. If you've added code that should be tested, add tests.
3. If you've changed APIs, update the documentation.
4. Ensure the test suite passes.
5. Make sure your code lints.
6. Issue that pull request!

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git

### Setup

```bash
# Clone your fork
git clone https://github.com/yourusername/fast-app.git
cd fast-app

# Add upstream remote
git remote add upstream https://github.com/katalabut/fast-app.git

# Install dependencies
go mod download

# Run tests
go test ./...
```

### Running Examples

```bash
# Basic example
cd example/basic
go run .

# Simple example with multiple services
cd example/simple
go run .

# Advanced example with database
cd example/advanced
go run .
```

## Code Style

### Go Code Style

We follow standard Go conventions:

- Use `gofmt` to format your code
- Use `golint` and `go vet` to check for issues
- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use meaningful variable and function names
- Add comments for exported functions and types

### Commit Messages

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Types:
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that do not affect the meaning of the code
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `perf`: A code change that improves performance
- `test`: Adding missing tests or correcting existing tests
- `chore`: Changes to the build process or auxiliary tools

Examples:
```
feat(health): add weighted health check strategy
fix(config): handle missing environment variables gracefully
docs: update README with health check examples
test(health): add tests for HTTP health checks
```

## Testing

### Writing Tests

- Write unit tests for all new functionality
- Use table-driven tests where appropriate
- Mock external dependencies
- Aim for high test coverage (>80%)

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run specific package tests
go test ./health/...
```

### Test Structure

```go
func TestFeatureName(t *testing.T) {
    t.Run("SpecificScenario", func(t *testing.T) {
        // Arrange
        // Act
        // Assert
    })
}
```

## Documentation

### Code Documentation

- Add godoc comments for all exported functions, types, and constants
- Include examples in godoc comments where helpful
- Keep comments up to date with code changes

### README Updates

- Update README.md if you add new features
- Add examples for new functionality
- Update the feature list if applicable

## Issue Reporting

### Bug Reports

When filing an issue, make sure to answer these questions:

1. What version of Go are you using (`go version`)?
2. What operating system and processor architecture are you using?
3. What did you do?
4. What did you expect to see?
5. What did you see instead?

### Feature Requests

We welcome feature requests! Please provide:

1. A clear description of the feature
2. The motivation/use case for the feature
3. Any implementation ideas you might have

## Code Review Process

1. All submissions require review before merging
2. We use GitHub pull requests for this purpose
3. Maintainers will review your code and provide feedback
4. Address any feedback and update your PR
5. Once approved, a maintainer will merge your PR

## Community

### Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code.

### Getting Help

- Check existing [issues](https://github.com/katalabut/fast-app/issues)
- Start a [discussion](https://github.com/katalabut/fast-app/discussions)
- Read the [documentation](./docs)

## Recognition

Contributors will be recognized in:
- The project's README
- Release notes for significant contributions
- The project's contributors page

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Questions?

Don't hesitate to reach out if you have questions about contributing. We're here to help!

Thank you for contributing to FastApp! 🚀
