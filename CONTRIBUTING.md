# Contributing to aegis

## Development Setup

```bash
# Clone the repository
git clone https://github.com/zoobzio/aegis.git
cd aegis

# Install development tools
make install-tools

# Install git hooks
make install-hooks

# Run checks
make check
```

## Available Commands

Run `make help` to see all available commands.

## Workflow

1. Fork the repository
2. Create a feature branch from `main`
3. Make your changes
4. Ensure `make check` passes
5. Submit a pull request

## Code Standards

- All code must pass linting (`make lint`)
- All tests must pass (`make test`)
- New code requires tests (80% patch coverage)
- Follow existing code conventions

## Commit Messages

Use conventional commits:

- `feat:` new features
- `fix:` bug fixes
- `docs:` documentation
- `test:` test changes
- `refactor:` code refactoring
- `perf:` performance improvements
