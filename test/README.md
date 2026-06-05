# Test Directory

In Go, test files must be located in the same package as the code they test. This is the standard Go testing convention.

## Test File Locations

```
vyzorix-update-server/
├── security/
│   ├── hmac_test.go        # HMAC verification tests
│   ├── jwt_test.go         # JWT tests
│   └── google_token_test.go # Google token verification tests
├── middleware/
│   ├── auth_test.go        # Auth middleware tests
│   ├── cors_test.go        # CORS tests
│   └── rate_limiter_test.go # Rate limiter tests
├── config/
│   └── config_test.go       # Configuration tests
├── controllers/
│   └── *_test.go            # Controller tests (when added)
├── hub/
│   └── *_test.go            # Hub tests (when added)
└── cmd/mockserver/
    └── *_test.go            # Integration tests
```

## Running Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./security/...
go test ./middleware/...
go test ./config/...

# Run with verbose output
go test ./security/... -v
go test ./middleware/... -v

# Run with coverage
go test ./... -cover
```

## Test Categories

1. **Unit Tests**: Test individual functions in isolation
2. **Integration Tests**: Test how components work together
3. **Security Tests**: Verify security properties (HMAC, JWT, CORS, etc.)
4. **Edge Case Tests**: Test boundary conditions and error handling