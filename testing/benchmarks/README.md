# Benchmarks

Performance benchmarks for aegis.

## Running

```bash
make test-bench
```

## Writing Benchmarks

- Use standard Go benchmark format: `func BenchmarkXxx(b *testing.B)`
- Include `-benchmem` results in analysis
- Document baseline expectations in comments
- Compare against previous results when optimizing
