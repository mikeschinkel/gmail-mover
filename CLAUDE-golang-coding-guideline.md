# CLAUDE Golang Coding Guidelines

## Clear Path Style Overview

The Clear Path style prioritizes code readability and maintainability through minimized nesting, strategic use of `goto`, and consistent patterns. This document serves as a comprehensive reference for all Golang development.

## Core Principles

### 1. Control Flow
- **Minimize nesting wherever possible**
- **Avoid `else` statements** - use helper functions instead of sequential if statements checking opposite conditions
- **Use `goto end` instead of early `return`**
- **Place `end:` label before the only return statement**
- **Use only ONE `end:` label per function**
- **DO NOT use an `end:` label in functions that have no `goto end` statements**
- **Place the sole return on the last line of the function**
- Do not use `break` to exit a loop if `goto end` would work instead, because `goto end` is less brittle when refactoring.

### 2. Function Structure
- If another label is needed beyond `end:`, refactor that logic into a helper function
- For long `if` statements that cannot use `goto end`, refactor the entire `if` into a helper function
- Refactor functions longer than can be viewed without scrolling into multiple helper functions
- Prefer `switch` statements over multiple `if` statements where applicable

## Golang-Specific Rules

### Const Declaration and Assignment
- Close ProperCase or CamelCase instead of snake_case

### Variable Declaration and Assignment
- **Declare ALL variables prior to first `goto`** (Go team requirement)
- **NEVER declare a variable after a `goto`** - this includes using `:=`
- **Do NOT use `:=` after the first `goto end`**
- **Do not shadow any variables**
- **Leverage Go's zero values with return variables where possible**

### Function Signatures and Returns
- **Use named return variables in the `func` signature for most functions**
- **Use the named return variables on the final `return`**
- **Do not use compound expressions in control flow statements like `if`**

### Error Handling
- **Always handle errors, even in `defer` statements**
- Error messages in `errors.New()`, `fmt.Errorf()` etc. should not start with a capital letter unless an initialism or acronym.

### Comments and Documentation
- **Comment for a type MUST start with the type name**
    - Example: `// MyCustomType is used for ...`

### Modern Go Practices
- **Use `any` instead of `interface{}`**

## Formatting and Whitespace

### Critical Formatting Rules
- **DO NOT leave trailing tabs or spaces on a line** - they make diffs very noisy
- **DO NOT leave trailing tabs or spaces on EMPTY lines**
- **Add a trailing newline at the end of a file** - helps with diffs

## Clear Path Pattern Examples

### Basic Function Structure
```go
func ProcessData(input string) (result string, err error) {
    var processed string
    var valid bool
    
    if input == "" {
        err = errors.New("empty input")
        goto end
    }
    
    processed = strings.TrimSpace(input)
    valid = validateInput(processed)
    if !valid {
        err = errors.New("invalid input")
        goto end
    }
    
    result = processValidInput(processed)
    
end:
    return result, err
}
```

### Avoiding Else Statements
```go
// WRONG - uses else
func BadExample(value int) (result string, err error) {
    if value > 0 {
        result = "positive"
    } else {
        result = "non-positive"
    }
    return result, err
}

// CORRECT - uses goto end
func GoodExample(value int) (result string, err error) {
    if value > 0 {
        result = "positive"
        goto end
    }
    
    result = "non-positive"
    
end:
    return result, err
}
```

### Complex Logic Refactoring
```go
// When you need complex if-else logic, refactor to helper function
func MainFunction(data []string) (processed []string, err error) {
    var item string
    
    for _, item = range data {
        processed = append(processed, processItem(item))
    }
    
end:
    return processed, err
}

func processItem(item string) string {
    var result string
    
    if strings.HasPrefix(item, "special_") {
        result = handleSpecialItem(item)
        goto end
    }
    
    if len(item) > 10 {
        result = handleLongItem(item)
        goto end
    }
    
    result = handleNormalItem(item)
    
end:
    return result
}
```

### Switch Over Multiple Ifs
```go
// PREFER this switch pattern
func HandleType(itemType string) (result string, err error) {
    switch itemType {
    case "type1":
        result = "handled type 1"
    case "type2":
        result = "handled type 2"
    case "type3":
        result = "handled type 3"
    default:
        err = errors.New("unknown type")
        goto end
    }
    
end:
    return result, err
}
```

## Common Mistakes to Avoid

1. **Using `:=` after `goto end`** - declare all variables before first goto
2. **Adding trailing whitespace** - especially on empty lines
3. **Using `else` statements** - refactor to helper functions instead
4. **Multiple labels** - use only `end:` label, refactor complex logic to helpers
5. **Compound expressions in control flow** - break them into separate statements
6. **Variable shadowing** - use unique names throughout function scope
7. **Missing trailing newline** - always end files with newline

## Checklist for Code Review

Before submitting Golang code, verify:
- [ ] All variables declared before first `goto`
- [ ] No `:=` used after `goto end`
- [ ] Only one `end:` label per function
- [ ] No `end:` label in functions without `goto end`
- [ ] No `else` statements (refactored to helpers if needed)
- [ ] No trailing whitespace on any lines
- [ ] Trailing newline at end of file
- [ ] Named return variables used
- [ ] All errors handled
- [ ] Type comments start with type name
- [ ] `any` used instead of `interface{}`
- [ ] Functions are reasonable length (viewable without scrolling)

## Remember

The Clear Path style exists to make code more readable and maintainable. When in doubt, favor clarity and simplicity over cleverness. Use helper functions liberally to break down complex logic while maintaining the single-exit-point pattern with `goto end`.
