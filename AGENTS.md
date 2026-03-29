# AGENTS.md

## Project Overview
This repository provides Go utility packages for generic programming, logging, and units handling. It is organised as a monorepo with several focused submodules:

- **Root package (`dvgoutils`)**: Generic helpers (e.g., `FilterSlice`, `MapSlice`, `Must`, `Ptr`).
- **`logging/`**: Context-aware logging utilities, error attribute helpers, and a test handler for structured log testing.
- **`units/`**: Types and helpers for working with bits and bytes, including string formatting and structured logging integration.

## Key Architectural Patterns
- **Generics**: Core helpers use Go generics for type-agnostic operations (see `filter.go`, `map.go`, `must.go`, `ptr.go`).
- **Context-based Logging**: Logging utilities store/retrieve loggers in `context.Context` (see `logging/context.go`). Panics if logger is missing.
- **Structured Logging**: Uses Go's `log/slog` for all logging. Custom attributes and error handling are provided (see `logging/errattr.go`).
- **Test Logging**: `logging/testhandler` provides a `TestHandler` for capturing and asserting logs in tests. Supports handler wrappers for extensibility.
- **Units**: `units/bits.go` and `units/bytes.go` define `Bits` and `Bytes` types with custom string and log value formatting.

## Developer Workflows
- **Testing**: Run all tests with `go test -trimpath ./...`. `-trimpath` is required so source locations in test output are stable across machines and Go installations. All packages are expected to be covered by tests.
- **Adding Utilities**: Place generic helpers in the root. For logging or units, use the respective subdirectory.
- **Extending Logging**: To add log handler wrappers for tests, use `WithHandlerWrapper` and `SetupTestHandler` (see `testhandler.go`).

## Project-Specific Conventions
- **Australian English**: All code, comments, documentation, and identifiers must use Australian English spelling. Update any existing code to fix non-Australian spellings.
- **Panics for Missing Context**: Functions like `FromContext` and `Must` panic on missing values or errors, enforcing fail-fast behaviour.
- **No External Logging Frameworks**: Only Go's standard `log/slog` is used for logging.
- **Test Assertions**: All tests use `github.com/stretchr/testify/require` for assertions.
- **Attribute Filtering**: Logging helpers filter out empty keys and expand attribute groups for structured logs.
- **Handler Extensibility**: Test log handlers can be wrapped for additional behaviour using handler wrapper functions.

## Integration Points
- **External Dependencies**: Minimal, mostly for testing (`testify`). Indirect dependencies are managed in `go.mod`.
- **Cross-Package Usage**: Root utilities are used by submodules (e.g., `FilterSlice` and `MapSlice` in `testhandler`).

## Examples
- **Filtering a slice**: `FilterSlice([]int{1,2,3}, func(i int) bool { return i%2==0 })`
- **Context logger**: `ctx = logging.WithLogger(ctx, logger)`; `logger := logging.FromContext(ctx)`
- **Test handler setup**: `ctx, handler, logger := SetupTestHandler(t, WithHandlerWrapper(...))`

## Key Files/Directories
- `filter.go`, `map.go`, `must.go`, `ptr.go`: Generic helpers
- `logging/context.go`, `logging/errattr.go`: Logging context and error helpers
- `logging/testhandler/testhandler.go`: Test log handler and setup utilities
- `units/bits.go`, `units/bytes.go`: Units types and formatting
