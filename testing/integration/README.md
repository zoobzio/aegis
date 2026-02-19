# Integration Tests

Integration tests for aegis mesh functionality.

## Running

```bash
make test-integration
```

## Writing Integration Tests

- Place tests in this directory
- Use build tag: `//go:build integration`
- Tests may require network access
- Tests should be self-contained and clean up after themselves
