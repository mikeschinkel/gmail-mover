# ADR-001: Go and Related Best Practices

**Status:** Accepted  
**Date:** 2025-08-08

## Context

Gmail Mover is implemented in Go and follows specific coding standards, architectural patterns, and tooling choices that establish consistency across the codebase and align with Go best practices.

## Decisions

### **Programming Language Choice**
- **Decision**: Use Go as the primary programming language
- **Rationale**: Excellent for CLI tools, strong standard library, cross-platform compilation, good Gmail API support

### **Clear Path Coding Style**
- **Decision**: All functions use `goto end` pattern with single exit point and named return variables
- **Implementation**: Functions declare all variables before any `goto` statements, use single `end:` label
- **Rationale**: Consistent error handling, single point for cleanup logic, easier debugging

### **Domain Types for Type Safety**
- **Decision**: Use domain types like `EmailAddress` and `LabelName` instead of primitive strings
- **Implementation**: Types defined in `gmover/` package with validation and zero-value checking
- **Rationale**: Prevents mixing different string types, provides validation, enhances API clarity

### **Branded Package Naming Convention**
- **Decision**: Use project-specific package names (gmover, gapi, gmcfg, gmcmds) instead of generic names
- **Implementation**: Package directories reflect branded names
- **Rationale**: Avoids conflicts with ecosystem packages, provides clear project identity

### **Command Pattern with Registration**
- **Decision**: Commands implement `CommandHandler` interface and register via `init()` functions
- **Implementation**: Commands in `gmcmds/` package register themselves with `cliutil.RegisterCommand()`
- **Rationale**: Extensible architecture, clean separation of concerns, follows established patterns

### **Explicit Dependency Injection**
- **Decision**: Pass dependencies (context, config) as parameters rather than accessing globals
- **Implementation**: `CommandHandler.Handle(ctx, config, args)` signature
- **Rationale**: Testability, explicit dependencies, proper cancellation support

### **Centralized Configuration Management**
- **Decision**: Use `gmcfg.FileStore` with filesystem sandboxing via `io/fs.Sub`
- **Implementation**: Config operations isolated to `~/.config/gmover/` directory
- **Rationale**: Secure file access, centralized config management, prevents path traversal

### **Type-Safe Database Operations**
- **Decision**: Use SQLC to generate type-safe Go code from SQL queries
- **Implementation**: `sqlc/` directory with schema and query definitions
- **Rationale**: Type safety without ORM complexity, compile-time query validation

### **Go Workspace for Multi-Module Architecture**
- **Decision**: Use Go workspace (`go.work`) with separate CLI module
- **Implementation**: Main project + `cmd/gmover-cli/` module with independent versioning
- **Rationale**: CLI binary can be versioned independently while sharing common packages

### **CLI-Friendly Logging**
- **Decision**: Custom `slog.Handler` providing clean output without timestamps
- **Implementation**: CLI handler in `cmd/gmover-cli/logger.go`
- **Rationale**: User-friendly CLI output while maintaining structured logging for debugging

## Implementation Evidence

These practices are evident throughout the codebase:
- Function patterns in any `.go` file (goto end style)
- Domain types in `gmover/email_address.go`, `gmover/label_name.go`
- Package structure: `gmover/`, `gapi/`, `gmcfg/`, `gmcmds/`
- Command handlers in `gmcmds/*.go` files
- Configuration management in `gmcfg/file_store.go`
- SQLC setup in `sqlc/` directory and `internal/sqlc/`
- Workspace files: `go.work`, `cmd/gmover-cli/go.mod`

## Consequences

**Positive:**
- Consistent codebase that's easy to navigate and understand
- Type safety reduces runtime errors
- Extensible command architecture
- Good separation of concerns
- Testable design with explicit dependencies

**Trade-offs:**
- More verbose than some alternatives (domain types vs strings)
- Requires discipline to maintain patterns consistently
- Learning curve for developers unfamiliar with Clear Path style