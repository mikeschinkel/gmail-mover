# Comprehensive Outline: Best Practices for Go Package Design

*This outline captures the complete context from an 8-hour technical interview covering every aspect of Go package design best practices. It includes all specific examples, code samples, qualifications, and nuanced guidance provided, along with the reasoning and experience-based insights behind each principle.*

## I. FOUNDATIONAL PRINCIPLES (11 Core Principles from your first answer)

### A. Domain-Driven vs Technical Architecture
- Domain-driven design rather than technical architecture (MVC is example of latter)
- Explanation of what this means in practical terms (without requiring DDD knowledge)
- Why technical layers fail in package-oriented languages

### B. Package Naming Philosophy
- Names should be relatively unique across Go ecosystem
- **Names should not conflict with variable names developers want to use**
- Use branded/creative names (inspiration from art, culture, history)
- Avoid namespacing (web.Request vs myapp.WebRequest example)
- **Specific bad examples**:
  - `models`, `handlers` - Generic, likely conflicts
  - `websocket` - Not necessarily common variable name, but easily a name some other non-discerning Go dev might use
  - `client` - Conflicts with common variable usage
  - `url` - Go's own `net/url` violates this principle
- **Specific good examples**:
  - `pgutil`, `waapi`, `scout`, `oapigen` - Your actual examples with specific domains
  - **NOT `myapp`** - That was just a placeholder example, not literal recommendation

### C. Cohesion and Coupling Principles
- High cohesion: reasonably intelligent person sees components belong together
- Loose coupling: explicit, minimal connection points
- Minimize imports as indicator of both

### D. Dependency Management Rules
- Minimize number of imports required
- Don't pull in more dependencies than absolutely necessary
- **Critical distinction**: Technical improvement packages vs Feature/Integration packages
  - **Technical improvement packages**: gorilla/websocket, validation libraries, string handling utilities - broad use, functionality independent of each other, like a grab bag
  - **Feature packages**: Stripe, Kubernetes, Docker, Facebook APIs - narrow use, cohesive and focused, typically found in top Go packages (https://github.com/EvanLi/Github-Ranking/blob/master/Top100/Go.md)
  - **External = 3rd party packages**
- One major third-party FEATURE package per package rule
- When 3 feature packages needed → create 4 packages total
- Multiple utility packages are okay per package

### E. Project Structure for Executables
- Single executable: `main` in `/cmd` with own module, tiny code, core in branded package
- Multiple executables: each in `/cmd/appname` with own modules
- Reusable main package requirement for testability
- Integration tests in `/test` that can call executables

## II. PACKAGE NAMING IN DEPTH

### A. The `xxxutil` Pattern
- Pattern: specific domain + util (pgutil, fileutil, jsonutil)
- **Why place in `/internal/`**: Not candidates for standalone repos due to generic names
- **BUT**: Most people don't need `/internal/` at all - only use when you genuinely expect external usage
- When this pattern works vs when it doesn't
- **Key insight**: If everyone followed this convention, we'd never have conflicting package names

### B. Creative Naming Strategy
- **Draw inspiration from art, culture, history for unique names**
- Goal: Names demonstrably unique across Go ecosystem
- **Specific examples with rationale**:
  - **Time utilities**: `timeutil`, `ymdcalc`, `chronos` (Greek god of time), `kairos` (opportune time), `cadence` (rhythm), `biztime` (business time), `ticktock`, `quanta`, `tempus` (Latin), `hora` (hour), `moment`, `janus` (two-faced god), `fiscaltime`, `fycalc`
  - **JSON helpers**: `jsonx`, `jsonutil`, `jsonchk` - but question whether JSON packages are needed at all
  - **URL/Context/Error/String utilities**: `urlx`/`urlutil`, `ctxutil`, `errutil`, `strutil` pattern
- **Key insight**: Most JSON/format utilities belong in domain packages, not generic utils
- **Creative naming examples**: `scout` for MCP Server, `hord`/`stash`/`stockpile` for cache packages

### C. Domain Package Naming
- Demonstrably unique across Go ecosystem
- Branded names no one else is using
- Your `scout` example for MCP Server

### D. Concrete Problem-Solving Examples
- Database package solutions: `dbutil`, `pgutil` in `/internal/`
- **HTTP client naming with full analysis**:
  - `waapi` - Short, unique, self-explanatory. Not likely to conflict. Variable: `wa := waapi.NewClient()`
  - `waclient` - Explicit, avoids `client.Client` issues. Variable: `wa := waclient.New()`
  - `wacloud` - Emphasizes Cloud API rather than on-device. Variable: `wa := wacloud.New()`
  - `wahttp` - Emphasizes HTTP-based API usage. Still WhatsApp-specific. Variable: `wa := wahttp.New()`
  - `wasvc` - "WhatsApp service" – fits if part of larger system of services. Variable: `wa := wasvc.New()`
  - `wabot` - If abstraction feels more bot-oriented. Variable: `bot := wabot.New()`
- Authentication: `oktautil`, `fbauth`, `auth0util`, `ghauth`, `omniauth`, `jwtauth`, `saml2auth`
- Cache: `redisutil`, `cacheutil`, `hord`, `stash`, `stockpile`
- File operations: Why do you need them when stdlib covers it? Name for the specific thing.

### E. Naming Patterns to Absolutely Avoid
- Single-word generics: `utils`, `helpers`, `common` (use prefixed versions)
- Generic-generic combinations
- Length constraints: 2 chars too short, 4+ minimum, 10+ probably too long
- Examples: `acctcheck`/`acctchk` vs `accountingvalidation`

### F. JSON and Format-Focused Anti-Patterns
- **JSON processing belongs in domain packages, not generic utils**
- **ChatGPT analysis you agreed with**:
  - JSON processing should be embedded in domain-specific packages
  - Naming should reflect purpose, not format
  - Instead of: `jsonutil.ExtractQuarters(data []byte)`
  - Prefer: `fiscal.QuarterDataFromJSON(data []byte)`
  - Or better: `fiscal.ParseReport(data)` // JSON is implementation detail
- **Exception cases where JSON packages make sense**:
  - Generic middleware for APIs or message brokers
  - Runtime validation of unknown schemas
  - Custom patching, diffing, template-based generation
  - Developer tools for arbitrary JSON
- **Problem-oriented naming**: `oapigen`, `oacheck` instead of `openapijson` or `jsonopenapi`
- **NEVER use format-focused names** (json, yaml, xml) as package names - they're implementation details
- **Key insight**: In idiomatic Go, data format details are implementation details, not architectural concerns

## III. DIRECTORY ORGANIZATION AND PROJECT STRUCTURE

### A. The Three-Package Starting Pattern
- **Your actual shell script example:**
  ```bash
  #!/usr/bin/env bash
  
  # For an app with only one executable
  go mod init github.com/mikeschinkel/myapp
  touch myapp.go
  
  mkdir cmd
  cd cmd
  go mod init github.com/mikeschinkel/myapp/myappd
  touch main.go
  
  mkdir test
  cd test
  touch main_test.go
  
  mkdir internal/logutil
  cd internal/logutil
  touch logutil.go
  
  # etc
  ```
- Why start with exactly these three packages
- When and how to expand beyond this foundation

### B. Package Splitting Decision Triggers
- Less cohesive (things don't belong together)
- Package getting really large and unwieldy
- Too many imports (with caveats about myapp having many imports from own repo)
- Multiple 3rd party feature packages
- **CRITICAL QUALIFIER**: ONLY when a cohesive package emerges from the larger one
- **Key insight**: Don't split just because package is big - only split when you can extract something that clearly belongs together as its own unit

### C. Directory Organization Patterns
- Multiple files within packages (purely for developer convenience, file names unimportant)
- Subdirectory organization: `/dbdrv/mysqldrv`, `/dbdrv/pgdrv`, `/dbdrv/sqlitedrv`
- Shared code placement in subdirectory hierarchies
- Your MCP Server `mcputil` extraction example

### D. Multi-Executable Structure
- `/cmd/myappd` and `/cmd/myappcli` with separate modules
- Shared code in core reusable package
- Testing structure: `/test/myappdtest` and `/test/myappclitest`
- Why executables should be tiny thin wrappers

### E. Non-Go Files
- Not Go package design concerns
- Organized by general repo best practices or tool requirements
- Go-specific files: `go.mod`, `go.sum`, `go.work`, `go.work.sum`

### F. Generated Code Placement
- Often in own package (`sqlc/`, `/generated`)
- Your examples with sqlc and oapicodegen
- Request for input on best practices you're still developing

## IV. THE `/internal/` DIRECTORY: WHEN AND WHY

### A. Core Decision Criteria
- "I reserve the right to change this"
- When you don't want to export for others to use
- Cross-module sharing limitation (executables can't call internal packages)

### B. When You Actually Need `/internal/`
- Organizational reuse (larger organizations with internal shared libraries)
- Open-source projects with significant expected traction
- Business models requiring backward compatibility commitments
- **Specific example consideration**: Corporate foundation teams building for other departments

### C. When You DON'T Need `/internal/`
- Closed-source apps that **won't be shared** (key qualifier: sharing is what matters)
- Open-source projects without expected external adoption
- Most people's code won't be used by others
- **"Tree falling in forest" analogy**: If nobody uses your code, breaking changes don't matter - no sound is made
- **Reality check**: Most developers overestimate how much their code will be used by others

### D. The Flexibility vs Complexity Trade-off
- Less complexity and more flexibility without `/internal/`
- Only use when genuinely needed

## V. AVOIDING IMPORT CYCLES

### A. Your Admission and Core Insight
- **"Those who can do, those who can't teach"** - you just do it now instinctively
- Import cycles were **"bane of my existence"** when learning Go
- **At this point**: "I am not even sure I can explain the process for how to avoid it, I just do it"
- **Key insight**: Your collection of principles work together to prevent import cycles automatically
- **Integration effect**: High cohesion + low coupling + domain organization + avoiding multiple 3rd party imports = natural cycle prevention

### B. Prevention Through Design
- High cohesion, low coupling
- Domain organization vs MVC (***big one***)
- Avoiding multiple 3rd party imports
- Independent, self-standing packages

### C. Related Types Should Be Together
- accounts/customers/products belong in one package (they're interrelated)
- When package gets too big, extract cohesive sub-packages

### D. Visualization and Tools
- Drawing Mermaid diagrams of package dependencies
- Visual emphasis on minimized imports
- All roads lead to core package pattern

### E. Your Database Driver Example
- **Before: Import cycle problem**
  ```go
  // dbdvr/dbdvr.go
  package dbdvr
  
  import (
      "dbapp/dbdvr/pgdvr"
  )
  
  type Driver struct {
      dbType string
      db     any
  }
  
  func NewDriver(db any) *Driver {
      drv := &Driver{db: db}
      _, ok := db.(pgdvr.DB)
      if ok {
          drv.dbType = "pg"
          goto end
      }
      // ... other db types
  end:
      return drv
  }
  
  func (drv Driver) Query(sql string) (result map[string]any) {
      switch drv.dbType {
      case "pg":
          result = drv.db.(*pgdvr.DB).Query(sql)
      }
      return result
  }
  ```

  ```go
  // dbdvr/pgdvr/pgdvr.go
  package pgdvr
  
  import (
      "dbapp/dbdvr"  // IMPORT CYCLE!
  )
  
  type DB struct {
      dbdvr.Driver
  }
  
  func (drv DB) Query(sql string) map[string]any {
      return nil // This would actually query Postgres
  }
  ```

- **After: Interface solution (no import cycle)**
  ```go
  // dbdvr/dbdvr.go
  package dbdvr
  
  type DBDriver interface {
      Query(query string) (result map[string]any)
  }
  
  type Driver struct {
      dbType string
      db     DBDriver
  }
  
  func NewDriver(db DBDriver) *Driver {
      return &Driver{db: db}
  }
  
  func (drv Driver) Query(sql string) (result map[string]any) {
      return drv.db.Query(sql)
  }
  ```

  ```go
  // dbdvr/pgdvr/pgdvr.go
  package pgdvr
  
  // NO IMPORT OF dbdvr NEEDED!
  
  type DB struct{}
  
  func (drv DB) Query(sql string) map[string]any {
      return nil // This would actually query Postgres
  }
  ```

- How interface in `dbdvr` breaks the cycle
- Complete working example with explanation

## VI. EXPORTED vs PRIVATE DESIGN DECISIONS

### A. Constructor Philosophy
- **ALL types should have one or more New*() constructors** - even simple ones
- **Direct instantiation almost always bad idea long term** - even for simple structs
- Multiple constructor patterns for different parameter sets
- **Reasoning**: Future-proofing for when fields need initialization, validation, or setup
- **Example of "simple" struct that still needs constructor**:
  ```go
  // Even this "simple" struct should have a constructor
  type User struct {
      Name string
      Age  int
  }
  
  // Better than direct instantiation
  func NewUser(name string, age int) User {
      return User{Name: name, Age: age}
  }
  
  // Allows for future validation, defaults, etc.
  ```

### B. When Struct Fields Need Initialization
- Maps, slices, channels require explicit initialization
- Constructor functions prevent runtime panics
- Example: Config struct with map field

### C. Private Types with Exported Constructors
- Pattern and when to use
- Interface returns for abstraction
- Your detailed examples and ChatGPT's 6 conditions for returning interfaces

### D. Export Decision Framework
- **LEAD WITH THE "AVOID" GUIDANCE** - default to private
- **Export when**: Part of intended public contract, needed by users, stable, well-documented
- **Keep Private when**: Implementation details, reserve right to change, internal utilities, mutable shared state
- **Key principle**: Export only when the use case is clear and you're committed to maintaining it
- **Avoid exporting "just in case"** - be intentional about public API surface

## VII. PACKAGE VARIABLES AND CONSTANTS

### A. Package Variables: The Singleton Problem
- **LEAD WITH AVOIDANCE GUIDANCE**
- Package-level variables are singletons with global lifetime - problematic in most cases
- Singleton implications: global state, testing difficulties, hidden dependencies
- When problematic: mutable state, request-scoped data, testing injection points

### B. When Package Variables Are Acceptable
- Shared state meaningful across package/application
- Singletons or long-lived resources initialized at startup
- Internal caching/memoization (usually unexported)
- **Accessor function pattern** instead of direct export
- **Example with slog**:
  ```go
  // Good: Controlled singleton with accessor
  var logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
  
  func DefaultLogger() *slog.Logger {
      return logger  // Accessor function provides control
  }
  
  // Avoid: Direct export of mutable state
  var GlobalCounter int // Tests will be problematic
  ```

### C. Constants Organization and Usage
- Your detailed document content about when to use constants
- Default values, grouped constants with `const` blocks
- **Critical iota warning**: Dangerous for persisted/serialized values
- Explicit values for database/network/external system constants
- Typed symbolic constants pattern

### D. Naming Conventions for Constants
- Clear naming for code completion and call site readability
- Prefix/suffix patterns to disambiguate
- Examples table of different strategies

### E. Constants vs Variables Decision Matrix
- Use `const` when: static, compile-time known, never changes
- Use `var` when: computed dynamically, runtime configuration dependent
- File organization patterns

## VIII. INITIALIZATION AND `init()` FUNCTIONS

### A. Package Variable Initialization Order
- Your detailed document about lexical order within files
- Cross-file compilation order problems
- **Prefer `const` when possible** (initializes before all vars)

### B. Safe Patterns for Variable Dependencies
- **Avoid cross-file dependencies** - Go's compiler file order isn't guaranteed
- **Inline function pattern for complex initialization**:
  ```go
  var complexVar = func() ComplexType {
      // Initialization logic here that ensures proper order
      dependency := someDependency()
      return ComplexType{Field: dependency}
  }()
  ```
- **Same-file consolidation strategies** - keep dependent variables in same file
- **Prefer `const` when possible** - constants initialize before all `var`s across all files

### C. `init()` Function Usage Rules
- Generally no issue with multiple but you've never needed them
- **Very limited scope**: Variable initialization that **cannot generate an error**
- **Embedded file example**: JSON country data loading with `embed.FS`
- **CRITICAL**: Unit tests must verify that init() operations cannot fail
- **Accessing embedded files is okay, but parsing must be guaranteed to succeed**

### D. What NOT to Use `init()` For
- Opening real files or network connections
- Actions that can fail (unless intentional panic)
- Side effects (HTTP handlers, metrics, globals)
- **When package used by others** (even if following rules - users may not read them!)
- Order-dependent operations across packages
- **Key insight**: If package will be imported by others, avoid init() entirely unless absolutely safe

### E. Explicit `Initialize()` Functions
- **Should almost always return error** (even if not needed now)
- Clear controllable startup order
- Explicit error handling
- Better testability
- No hidden side effects

## IX. INTERFACE DESIGN PHILOSOPHY AND PATTERNS

### A. Discovery vs Proactive Creation
- Interfaces should emerge from usage, not be designed upfront
- **Avoid over-engineering** with unnecessary abstractions
- Let genuine abstraction needs drive interface creation

### B. Single-Method vs Multi-Method Reality
- **Single-method**: Serendipitous reuse, optional functionality
- **Multi-method**: Drivers, plugins, grouping similar types for iteration
- Your Property interface example for MCP tools
- **Go standard library violates single-method dogma constantly**

### C. Interface Categories with Your Examples
- **Behavioral**: Your `mcpPropertyOptionsGetter` example
- **Type-Shape**: Your `Property` interface with multiple methods
- **Adapter**: Framework extension points
- **Marker**: Your `Property` interface marker method pattern
- **Internal**: Testability and dependency inversion

### D. The `-er` Suffix Reality Check
- **"Comforting rule but falls on its face in production"**
- Use when applicable (enabling single-method interfaces)
- Don't force when inappropriate - many production cases don't fit
- **Go standard library violates this constantly**: `net.Conn`, `http.ResponseWriter`, `context.Context`, etc.
- Your examples where `-er` doesn't apply: `StringOption`, `NumberOption`, `BoolOption`, `ArrayOption`, `Property` interfaces
- **Reality**: Single-method interfaces good for serendipitous reuse, but multi-method needed for drivers, plugins, type grouping

### E. Interface Location Strategy
- **"Doesn't matter much when no import cycle exists"** - your actual position
- **MUST be in user package when import cycle exists** - only then does location become critical
- **Should be in BOTH packages for documentation** (contrary to common dogma)
- **Reasoning**: Discoverability and recognition importance - developers need to find and understand interfaces
- Common dogma says "define where used" but misses documentation value

### F. Interface Embedding Patterns
- **Your `Property`/`propertyEmbed` example** with full context:
  ```go
  type propertyEmbed interface {
      GetName() string
      Required() Property
      // ... base methods
  }
  
  type Property interface {
      propertyEmbed
      mcpToolOption([]mcp.PropertyOption) mcp.ToolOption
  }
  ```
- **Type checking strategy**: `var _ propertyEmbed = (*property)(nil)` for base behavior
- **Compiler verification**: `var _ Property = (*stringProperty)(nil)` for full interface
- **Avoiding dummy methods** while maintaining compiler checks
- **Your insight**: "I embed interfaces in other interfaces when it makes sense. I know, that is a circular explanation, but I am basically saying that I think it is mostly intuitively obvious when you do that."
- **Question**: "Is 'it is mostly intuitively obvious' an acceptable answer, or is it too much of a dodge?"
- **Answer**: It's honest acknowledgment that these decisions are often contextual and experience-based

### G. `any` vs `any` and Generics
- **"At this point I always prefer using `any` over `any`"**
- **Why**:
  - `any` is shorter
  - Doesn't have braces which make reading and writing awkward
  - They both mean exactly the same thing
  - `any` is much easier to reason about
- **Question**: "Given that, why would you ever want to use `any`?!?"
- **Generics not replacement for `any`**: Can't have generic methods per se
- **When `any` still has value**: Many scenarios where generics don't apply
- **Always prefer `any`** in modern Go code

## X. API DESIGN PHILOSOPHY

### A. Core Guiding Principle
- **"Simply put, what do users need for it to be useful"**
- Cut through theoretical concerns
- Focus on actual user needs
- Avoid over-engineering

### B. Balancing Simplicity vs Flexibility
- Start simple, add complexity only when proven necessary
- User-driven API evolution
- Principle of least surprise

### C. Testing Implications for API Design
- If tests need to modify package-level variables = design smell
- Explicit dependency injection over global state
- Better patterns for testable code

## XI. REAL-WORLD EXAMPLES AND CODE PATTERNS

### A. Your MCP Server Implementation
- **Complete interface hierarchy examples**:
  ```go
  // Property is the main interface for all property types
  type Property interface {
      propertyEmbed
      mcpToolOption([]mcp.PropertyOption) mcp.ToolOption
  }
  
  // propertyEmbed provides base functionality
  type propertyEmbed interface {
      GetName() string
      Required() Property
      Name(string) Property
      Description(string) Property
      mcpPropertyOptions() []mcp.PropertyOption
  }
  
  // Type-specific option interfaces
  type StringOption interface {
      SetStringProperty(Property)
  }
  
  type NumberOption interface {
      SetNumberProperty(Property)
  }
  
  type BoolOption interface {
      SetBoolProperty(Property)
  }
  
  type ArrayOption interface {
      SetArrayProperty(Property)
  }
  ```

- **Property option patterns** with concrete implementations
- **Server interface design** for MCP protocol abstraction
- **Tool request/response handling** with type-safe wrappers
- **Example of interface embedding**: Property/propertyEmbed split allows targeted type checking
- **Type checking strategy**: `var _ propertyEmbed = (*property)(nil)` for compiler verification

### B. Database Driver Abstraction
- Import cycle problem and interface solution
- Complete before/after code examples
- Generic vs specific driver implementations

### C. WhatsApp API Client Naming
- ChatGPT's analysis of naming options
- Justification for each approach
- Variable naming considerations

### D. Package Structure Evolution
- **How your MCP Server grew and split packages**:
  - Started as single package with MCP server implementation
  - As code grew, recognized generic MCP interaction patterns
  - Extracted `mcputil` package for generic `mcp-go` interface improvements
  - **Reasoning**: "I wanted to completely isolate `mcp-go` because I really dislike its API and wanted to create one with much better DX"
  - **Result**: `mcputil` has no knowledge of specific MCP server, only generic interface to `mcp-go`
- **Isolating problematic dependencies**: Using wrapper packages to hide poor APIs
- **Cohesive extraction**: `mcputil` emerged as coherent unit, not arbitrary split

## XII. COMMON ANTI-PATTERNS AND PITFALLS

### A. Naming Anti-Patterns
- Generic package names that conflict
- Over-generic combinations
- Format-focused naming (json, xml, yaml as package names)
- Variable name conflicts

### B. Structural Anti-Patterns
- MVC-style technical architecture
- Multiple feature package imports
- Cross-file initialization dependencies
- Inappropriate `/internal/` usage

### C. Interface Anti-Patterns
- Proactive interface creation
- Dogmatic single-method adherence
- Missing documentation through location

### D. Initialization Anti-Patterns
- Side-effectful `init()` functions
- Global state dependency in tests
- Complex error-prone initialization in `init()`

## XIII. ADVANCED TOPICS AND EDGE CASES

### A. Generated Code Integration
- Where to place generated code
- Package organization with code generation
- Your sqlc and oapicodegen examples

### B. Cross-Module Considerations
- Module boundaries and package design
- `/internal/` limitations across modules
- Executable module organization

### C. Third-Party Package Integration
- Distinguishing technical vs feature packages
- Isolation strategies for problematic dependencies
- Wrapper package patterns

### D. Testing Package Design
- Integration test organization
- Testing executable behavior
- Avoiding global state in tests

## XIV. DECISION FRAMEWORKS AND CHECKLISTS

### A. Package Creation Decision Tree
- When to create new package
- Where to place it
- How to name it
- What to export

### B. Interface Design Checklist
- When interface is needed
- Single vs multi-method decision
- Where to define it
- How to name it

### C. Dependency Management Framework
- Import minimization strategies
- Third-party package evaluation
- Cycle prevention techniques

### D. API Evolution Strategy
- Export decision process
- Backward compatibility considerations
- User need assessment