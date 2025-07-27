# Go Lessons Learned: CLI Architecture Evolution

This document captures the key insights, design patterns, and architectural lessons learned during the development of a robust, maintainable CLI application architecture in Go.

## Table of Contents

1. [Architectural Evolution](#architectural-evolution)
2. [Coding Style Guidelines](#coding-style-guidelines)
3. [Design Pattern Insights](#design-pattern-insights)
4. [Flag System Architecture](#flag-system-architecture)
5. [Command Registry Design](#command-registry-design)
6. [Dependency Management](#dependency-management)
7. [Error Handling Patterns](#error-handling-patterns)
8. [Testing Considerations](#testing-considerations)
9. [Performance and Scalability](#performance-and-scalability)
10. [Anti-Patterns Avoided](#anti-patterns-avoided)

## Architectural Evolution

### From Monolithic to Modular

**Initial State**: Single main.go with flag-based mode switching
```go
// Anti-pattern: Convoluted switch logic
switch mode {
case "list-labels":
    // 50 lines of procedural code
case "move-emails":
    // 100 lines of procedural code
}
```

**Final State**: Self-contained command architecture
```go
// Clean pattern: Commands register themselves
func init() {
    var c = gmover.GetConfig()
    cliutil.RegisterCmd(&MoveCmd{
        CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
            Name: "move",
            Flags: []cliutil.FlagDef{
                {SetString: c.SetSrcEmail, Name: "src", Required: true, Usage: "Source Gmail address"},
            },
        }),
    })
}
```

**Key Insight**: Self-contained commands with declarative flag definitions are infinitely more maintainable than procedural mode switching.

### MCP Tools Pattern Adoption

**Inspiration**: The scout-mcp/mcptools pattern provided the blueprint for self-contained commands.

**Implementation**: Each command exists in its own file with:
- Embedded `*CmdBase` for common functionality
- `init()` function for self-registration
- Declarative flag definitions
- Minimal `Handle()` method focused on business logic

**Benefit**: Adding new commands requires zero changes to existing code - just create the new command file.

## Coding Style Guidelines

### Clear Path Coding Style

**Principle**: Single exit point, named return variables, declare all variables before goto statements.

```go
// Correct Clear Path style
func ProcessData(input string) (result string, err error) {
    var processed bool
    var data []byte
    
    if input == "" {
        err = fmt.Errorf("input required")
        goto end
    }
    
    data, err = process(input)
    if err != nil {
        goto end
    }
    
    processed = true
    result = string(data)

end:
    if processed {
        logger.Info("Processing completed")
    }
    return result, err
}
```

**Key Benefits**:
- Consistent error handling patterns
- Clear resource cleanup opportunities
- Single point of return for debugging
- Reduced cognitive load when reading code

### Single-Use Variable Guidelines

**Anti-Pattern**: Declaring variables for single-use struct literals
```go
// Wrong: Unnecessary variable declaration
cmd := &MoveCmd{
    CmdBase: NewCmdBase("move", "Move emails", "Moves emails between accounts"),
}
cliutil.RegisterCmd(cmd)
```

**Correct Pattern**: Pass struct literals directly
```go
// Right: Direct struct literal usage
cliutil.RegisterCmd(&MoveCmd{
    CmdBase: NewCmdBase("move", "Move emails", "Moves emails between accounts"),
})
```

**Rationale**: Reduces boilerplate and eliminates unnecessary intermediate variables that serve no purpose beyond parameter passing.

## Design Pattern Insights

### Singleton Pattern Defense

**Context**: CLI applications have inherently singular execution contexts.

**Justified Usage**:
```go
// Singleton for CLI configuration
var globalConfig *Config

func GetConfig() *Config {
    if globalConfig == nil {
        config := NewConfig(ShowHelp)
        globalConfig = &config
    }
    return globalConfig
}
```

**Defense Arguments**:
1. **Domain Appropriateness**: CLI apps run once per invocation with one command
2. **No Concurrency Issues**: Commands don't run concurrently
3. **Clear Lifecycle**: Config is born at startup, lives during execution, dies at exit
4. **Precedent**: Go's flag package, Docker CLI, Cobra/Viper all use global state
5. **Testing Support**: Easy to reset with `ResetConfig()` function

**Key Lesson**: **Dogmatic opposition ignores context**. Patterns that are problematic in enterprise software can be perfect solutions in focused CLI tools.

### Declarative vs Procedural Patterns

**Transformation**: From procedural flag parsing to declarative flag definitions.

**Before**: Procedural flag handling in every command
```go
func (c *MoveCmd) Handle(args []string) error {
    flagSet := flag.NewFlagSet("move", flag.ExitOnError)
    srcEmail := flagSet.String("src", "", "Source Gmail address")
    dstEmail := flagSet.String("dst", "", "Destination Gmail address")
    // ... more boilerplate
    
    err := flagSet.Parse(args)
    // ... validation logic
    
    config.SetSrcEmail(*srcEmail)
    config.SetDstEmail(*dstEmail)
    // ... more manual config setting
}
```

**After**: Declarative flag definitions with automatic config population
```go
func init() {
    var c = gmover.GetConfig()
    cliutil.RegisterCmd(&MoveCmd{
        CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
            Flags: []cliutil.FlagDef{
                {SetString: c.SetSrcEmail, Name: "src", Required: true, Usage: "Source Gmail address"},
                {SetString: c.SetDstEmail, Name: "dst", Required: true, Usage: "Destination Gmail address"},
            },
        }),
    })
}

func (c *MoveCmd) Handle(ctx context.Context, config cliutil.ConfigProvider, args []string) error {
    // Config already populated, just execute business logic
    return gmover.RunWithApproval(config.(*gmover.Config), nil)
}
```

**Benefits**:
- **Zero Boilerplate**: No manual flag parsing or config setting
- **Generic Help**: Help text generated automatically from flag definitions
- **Generic Validation**: Required flags validated automatically
- **Consistency**: All commands follow the same pattern

## Flag System Architecture

### Setter Function Pointers

**Innovation**: Using function pointers to automatically populate configuration.

```go
type FlagDef struct {
    Name      string
    Required  bool
    Usage     string
    SetString func(string)  // Direct pointer to config setter
    SetBool   func(bool)
    SetInt64  func(int64)
}
```

**Automatic Population**:
```go
// During flag parsing, setters are called automatically
for _, flagDef := range c.flags {
    switch flagDef.Type {
    case StringFlag:
        value := *flagPtrs[flagDef.Name].(*string)
        if flagDef.SetString != nil {
            flagDef.SetString(value)  // Config automatically updated
        }
    }
}
```

**Key Insight**: This eliminates the need for manual `config.Set*()` calls in every command, while maintaining compile-time type safety.

### Global Flags Architecture

**Challenge**: Flags like `--auto-confirm` should be global, not per-command.

**Solution**: Global flag parsing before command-specific parsing.

```go
// Global flag handler interface (keeps cliutil generic)
type GlobalFlagHandler interface {
    SetAutoConfirm(bool)
}

// Parse global flags first, then command-specific flags
args, err = parseGlobalFlags(args)  // Extracts --auto-confirm
cmdPath, args = findBestCmdMatch(args)  // Find command
cmdBase.ParseFlags(args)  // Parse command-specific flags
```

**Benefits**:
- Global flags work with any command
- No duplication of global flag definitions
- Clean separation between global and command-specific concerns

## Command Registry Design

### Recursive Registry vs Flat Maps

**Problem**: Flat registry design doesn't scale to multiple command levels.

**Anti-Pattern**: Multiple maps for different levels
```go
// Bad: Doesn't scale beyond 2 levels
var cmdRegistry = make(map[string]CmdHandler)
var subCmdRegistry = make(map[string]map[string]CmdHandler)
// What about sub-sub-commands?
```

**Solution**: Recursive command structure
```go
// Good: Infinitely scalable
type Cmd struct {
    Handler CmdHandler
    SubCmds map[string]*Cmd
}

var cmds = make(map[string]*Cmd)
```

**Dot Notation Support**: Commands can register at any depth using "job.run.special" notation.

**Key Lesson**: Recursive data structures are always better than nested flat structures when depth is unknown.

### Naming Improvements

**Problem**: JavaEsque names like `Registry` and `Manager` focus on mechanisms rather than domain concepts.

**Before**: `cmdRegistry`, `ValidateRegistry()`
**After**: `cmds`, `ValidateCmds()`

**Principle**: Name things for what they represent in the domain, not what pattern they implement.

## Dependency Management

### Explicit Dependencies vs Hidden Globals

**Evolution**: From hidden global access to explicit dependency injection.

**Before**: Hidden global access
```go
func (c *MoveCmd) Handle(args []string) error {
    config := gmover.GetConfig()  // Hidden dependency
    return gmover.RunWithApproval(config, nil)
}
```

**After**: Explicit dependency injection
```go
func (c *MoveCmd) Handle(ctx context.Context, config cliutil.ConfigProvider, args []string) error {
    gmoverConfig := config.(*gmover.Config)  // Explicit dependency
    return gmover.RunWithApproval(gmoverConfig, nil)
}
```

**Benefits**:
- **Testability**: Easy to pass test configs
- **Clarity**: Dependencies are obvious at call sites
- **Type Safety**: Compiler ensures dependencies are provided

### Context Integration

**Addition**: `context.Context` parameter for cancellation support.

```go
// Signal handling for graceful shutdown
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

go func() {
    <-sigChan
    logger.Info("Received interrupt signal, shutting down...")
    cancel()
}()

// Context passed to all commands
err = cliutil.RunCmd(ctx, gmover.GetConfig(), os.Args[1:])
```

**Future Benefit**: Long-running operations (like Gmail API calls) can be cancelled gracefully on Ctrl-C.

## Error Handling Patterns

### Import Cycle Resolution

**Problem**: Generic utilities importing app-specific packages.

**Anti-Pattern**: cliutil importing gmover
```go
// Bad: Creates import cycle
package cliutil
import "github.com/mikeschinkel/gmail-mover/gmover"

func (c *CmdBase) GetConfig() *gmover.Config {
    return gmover.GetConfig()
}
```

**Solution**: Interface-based dependency injection
```go
// Good: Generic interface
type ConfigProvider interface{}

// Application sets up the connection
cliutil.SetGlobalFlagHandler(gmover.GetConfig())
```

**Principle**: Generic utilities should never import application-specific packages. Use interfaces and dependency injection instead.

### Goto Statement Best Practices

**Clear Path Pattern**: Declare all variables before any goto statements.

```go
// Wrong: Variable declared after potential goto
func BadExample() (err error) {
    if condition {
        goto end  // This will fail compilation
    }
    
    var result string  // Declared after goto - compilation error
    
end:
    return err
}

// Correct: All variables declared first
func GoodExample() (err error) {
    var result string  // All variables declared before any goto
    
    if condition {
        goto end  // This works fine
    }
    
    result = "success"
    
end:
    return err
}
```

## Testing Considerations

### Singleton Reset Support

**Testing Pattern**: Provide reset functions for stateful singletons.

```go
// Production code
func GetConfig() *Config {
    if globalConfig == nil {
        config := NewConfig(ShowHelp)
        globalConfig = &config
    }
    return globalConfig
}

// Testing support
func ResetConfig() {
    globalConfig = nil
}

// Test usage
func TestMoveCommand(t *testing.T) {
    gmover.ResetConfig()  // Clean state for each test
    config := gmover.GetConfig()
    // test with clean config...
}
```

**Key Insight**: Singletons in CLI applications are actually easier to test than dependency injection when reset functions are provided.

## Performance and Scalability

### Command Registration Performance

**Self-Registration Pattern**: Commands register themselves during package initialization.

```go
// Each command file includes:
func init() {
    cliutil.RegisterCmd(&MoveCmd{...})
}

// Main just imports the package
import _ "github.com/mikeschinkel/gmail-mover/gmcmds"
```

**Benefits**:
- **Fast Startup**: All commands registered during compilation/linking
- **Zero Dynamic Discovery**: No runtime filesystem scanning
- **Minimal Memory**: Only active command structures loaded

### Flag Parsing Optimization

**Two-Phase Parsing**: Global flags parsed separately from command flags.

1. **Global Phase**: Extract `--auto-confirm` and similar
2. **Command Phase**: Parse command-specific flags

**Benefit**: Avoids complex flag precedence logic while maintaining clean separation.

## Anti-Patterns Avoided

### Builder Pattern (GARBAJE)

**Definition**: GARBAJE = "Go As wRitten By A Java Engineer"

**Anti-Pattern**: Unnecessary fluent interfaces
```go
// GARBAJE: Over-engineered for Go
FlagBuilder().
    StringFlag("src", "SrcEmail", true, "Source Gmail address").
    StringFlag("dst", "DstEmail", true, "Destination Gmail address").
    Build()
```

**Go Idiomatic**: Simple struct literals
```go
// Idiomatic Go: Clear and direct
[]cliutil.FlagDef{
    {SetString: c.SetSrcEmail, Name: "src", Required: true, Usage: "Source Gmail address"},
    {SetString: c.SetDstEmail, Name: "dst", Required: true, Usage: "Destination Gmail address"},
}
```

**Lesson**: Go favors explicit, simple constructs over complex fluent APIs.

### Premature Abstraction

**Anti-Pattern**: Creating interfaces before they're needed.

**Principle**: Start with concrete implementations. Extract interfaces only when you have multiple implementations or need to break import cycles.

**Example**: The `ConfigProvider interface{}` was introduced only when needed to break the import cycle between cliutil and gmover.

### Reflection Over Type Safety

**Anti-Pattern**: String-based field references
```go
// Bad: Runtime errors, no compile-time checking
{SetString: "SrcEmail", Name: "src", ...}
```

**Type-Safe Alternative**: Function pointers
```go
// Good: Compile-time checking, refactoring-safe
{SetString: c.SetSrcEmail, Name: "src", ...}
```

**Lesson**: Prefer compile-time safety over runtime flexibility in CLI applications.

## Key Takeaways

1. **Context Matters**: Design patterns must fit the domain. Singleton is perfect for CLI apps.

2. **Declarative > Procedural**: Declarative flag definitions eliminate boilerplate and enable generic functionality.

3. **Self-Contained Components**: Commands that register themselves are easier to maintain and extend.

4. **Explicit Dependencies**: Pass dependencies as parameters rather than accessing globals in business logic.

5. **Recursive Design**: Use recursive data structures for unknown-depth hierarchies.

6. **Type Safety First**: Prefer compile-time checking over runtime flexibility.

7. **Go Idioms**: Embrace Go's simplicity. Avoid over-engineering from other languages.

8. **Import Cycles**: Keep generic utilities generic. Use interfaces and dependency injection to break cycles.

9. **Testing Support**: Design stateful components with reset capabilities for clean testing.

10. **Evolution Over Revolution**: Refactor incrementally. Each step should leave the code in a working state.

## Conclusion

This CLI architecture evolution demonstrates that thoughtful, incremental refactoring can transform a procedural, monolithic CLI into a clean, maintainable, and extensible system. The key is to:

- **Question assumptions** (like "singletons are always bad")
- **Favor simplicity** over complexity
- **Learn from the domain** (CLI apps have different constraints than web services)
- **Embrace Go idioms** rather than importing patterns from other languages
- **Focus on maintainability** and developer experience

The result is a CLI architecture that's easy to understand, test, extend, and maintain while following Go best practices and idioms.