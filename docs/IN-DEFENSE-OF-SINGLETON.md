# In Defense of the Singleton Pattern

## Singleton Defense for CLI Application Config

### 1. CLI Application Context
- **Single Execution Scope**: CLI apps run once per invocation with one command
- **No Concurrency**: Commands don't run concurrently - one config per process lifetime
- **Clear Lifecycle**: Config is born at startup, lives during execution, dies at exit
- **Stateless Between Runs**: Each invocation is independent

### 2. Domain-Appropriate Design
- **Single Source of Truth**: CLI configuration is inherently singular - there's one set of flags, one execution context
- **Natural Fit**: The config represents "the current command invocation" - there's exactly one
- **No Multiple Instances**: Unlike a web server handling multiple requests, CLI handles one command

### 3. Practical Benefits
- **Type Safety**: Direct method references (`c.SetSrcEmail`) vs string-based reflection
- **Compile-Time Checking**: Flag definitions fail fast if config interface changes
- **Zero Boilerplate**: `var c = gmover.GetConfig()` vs dependency injection machinery
- **Clear Intent**: Code explicitly shows which config fields each flag affects

### 4. Anti-Pattern Avoidance
The problematic singleton patterns involve:
- **Hidden Dependencies**: Our usage is explicit at call sites
- **Global Mutable State**: Our config is scoped to single command execution
- **Testing Difficulties**: CLI testing can reset config between tests
- **Thread Safety**: N/A in single-threaded CLI context

### 5. Alternative Complexity
Dependency injection alternatives would require:
- **Constructor Injection**: Pass config to every command constructor
- **Context Threading**: Thread config through multiple layers
- **Interface Abstraction**: Abstract config interface for testability
- **Factory Functions**: More code to maintain

### 6. Precedent in CLI Tools
Many successful CLI tools use global config:
- **Cobra/Viper**: Global flag and config management
- **Go's flag package**: Global FlagSet
- **Docker CLI**: Global configuration state

### Verdict
This is a **legitimate, contextually-appropriate use** of singleton pattern. The domain (CLI app) naturally has singular execution context, the benefits (type safety, simplicity) are real, and the typical singleton problems (hidden dependencies, concurrency issues) don't apply.

**Dogmatic opposition ignores context**. This isn't enterprise software with complex object lifecycles - it's a focused CLI tool where singleton actually fits the problem domain perfectly.