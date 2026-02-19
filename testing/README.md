# Testing

This directory contains test infrastructure for aegis.

## Structure

```
testing/
├── helpers.go        # Domain-specific test helpers
├── helpers_test.go   # Tests for helpers themselves
├── integration/      # Integration tests
└── benchmarks/       # Performance benchmarks
```

## Running Tests

```bash
# All tests
make test

# Unit tests only (short mode)
make test-unit

# Integration tests
make test-integration

# Benchmarks
make test-bench

# Coverage report
make coverage
```

## Writing Tests

- Use helpers from this package for common assertions
- Integration tests require the `integration` build tag
- Place benchmarks in `testing/benchmarks/`
- Maintain 1:1 mapping between source files and test files
