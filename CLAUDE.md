# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**runn** is a Go-based scenario testing and automation framework that executes multi-step operations following YAML-based runbooks. It supports HTTP, gRPC, database, Chrome DevTools Protocol (CDP), SSH, and local command execution.

## Build and Development Commands

```bash
# Run tests with coverage
make test

# Run tests with race detection
make race

# Run integration tests
make test-integration

# Run all tests including integration
make test-all

# Build the binary
make build

# Run linting
make lint

# Run load testing
make test-loadt

# Generate documentation
make doc
```

### Key Go Commands

```bash
# Run tests manually
go test ./... -coverprofile=coverage.out -covermode=count

# Run integration tests
go test ./... -tags=integration -count=1

# Build with version info
go build -ldflags="$(BUILD_LDFLAGS)" -o runn cmd/runn/main.go

# Install development dependencies
make depsdev
```

## Architecture Overview

### Core Components

1. **Entry Point**: `cmd/runn/main.go` - CLI entry point with commands in `cmd/` directory
2. **Runbook System**: `runbook.go` - YAML-based scenario definitions with runners, steps, and variables
3. **Execution Engine**: `operator.go` - Core execution engine managing all runners and step execution
4. **Runner Architecture**: Pluggable system for different protocols (HTTP, gRPC, DB, CDP, SSH, exec)
5. **Variable System**: `internal/store/store.go` - Centralized state management with scoping
6. **Expression Engine**: `internal/expr/expr.go` - Dynamic evaluation using expr-lang/expr

### Runner Types

- **HTTP Runner** (`http.go`): REST API testing with OpenAPI validation
- **gRPC Runner** (`grpc.go`): gRPC service testing with protobuf support
- **Database Runner** (`db.go`): SQL operations for MySQL, PostgreSQL, SQLite, Cloud Spanner
- **CDP Runner** (`cdp.go`): Browser automation via Chrome DevTools Protocol
- **SSH Runner** (`ssh.go`): Remote command execution
- **Exec Runner** (`exec.go`): Local command execution
- **Built-in Runners**: test, include, dump, bind, runner

### Key File Structure

- `cmd/`: CLI command implementations
- `internal/`: Internal packages (builtin functions, expression handling, etc.)
- `testdata/`: Test runbooks and test data
- `examples/`: Example runbooks demonstrating features
- `capture/`: Result capture mechanisms
- `testutil/`: Test utilities for setup

## Common Development Patterns

### Testing

- Use `runn.T(t)` for Go test integration
- Runbooks are typically in `testdata/book/*.yml`
- Integration tests use Docker containers via `testutil/`
- Use `runn.Load()` to load multiple runbooks, `runn.New()` for single runbook

### Runbook Structure

```yaml
desc: Description of the scenario
runners:
  req: https://api.example.com  # HTTP runner
  db: mysql://user:pass@host/db # Database runner
vars:
  username: alice
  token: ${SECRET_TOKEN}
steps:
  - desc: Step description
    req:
      /endpoint:
        post:
          body:
            application/json:
              key: value
    test: current.res.status == 200
```

### Variable Scoping

- `vars`: Global variables
- `steps`: Step results (array or map access)
- `current`: Current step result
- `previous`: Previous step result
- `parent`: Parent runbook variables (for included runbooks)
- `env`: Environment variables

### Expression Evaluation

Uses `expr-lang/expr` with additional built-in functions:
- `compare()`, `diff()`: Value comparison
- `pick()`, `omit()`: Map filtering
- `merge()`: Map merging
- `faker.*`: Fake data generation
- `file()`: File reading
- `urlencode()`: URL encoding

## Testing Requirements

- Run integration tests locally with `make test-integration`
- Tests require certificates generated with `make cert`
- Some tests require specific Docker containers (see `testutil/`)
- Use `chmod 600 testdata/sshd/id_rsa` for SSH tests

## Important Notes

- The framework is designed for scenario-based testing, not unit testing
- Runbooks support includes, loops, conditionals, and deferred execution
- All runners implement a consistent `Run(context.Context, *step)` interface
- Results are captured in a structured format accessible via expressions
- Security: Some features require explicit scopes (e.g., `run:exec` for exec runner)