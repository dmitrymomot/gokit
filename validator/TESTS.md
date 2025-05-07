# Validator Package Testing Guidelines

This document describes the standards and patterns for writing tests for validator rules in the `github.com/dmitrymomot/gokit/validator` package.

## 1. Test File and Package Structure

- Place tests in a file named `<rule>_test.go` or in `data_test.go` if grouped.
- Use the test package suffix: `package validator_test`.
- Import `github.com/stretchr/testify/require` for assertions.

## 2. Test Function Naming

- Use the pattern: `func Test<RuleName>Validator(t *testing.T)`.
    - Example: `TestEmailValidator`, `TestUUIDValidator`

## 3. Test Table Structure

For each rule, define a table-driven test using a struct with these fields:

```go
tests := []struct {
    name            string
    structWithTag   any
    wantErrContains string // "" means expect no error
}{
    // ... test cases
}
```

- `name`: Describes the test case.
- `structWithTag`: Inline struct with the field and validation tag set.
- `wantErrContains`: If not empty, the error message must contain this substring.

## 4. Defining Test Cases

For each rule, cover these scenarios:

- **Valid input**: Should not return an error.
- **Invalid input**: Should return an error containing the rule name or a specific error substring.
- **Empty string**: If the rule should not allow empty, expect an error.
- **Type mismatch**: Use a non-string type to ensure type errors are caught.
- **Edge cases**: Any other boundary or special cases for the rule.

**Example for an `email` rule:**

```go
{
    name: "valid email",
    structWithTag: struct{Email string `validate:"email"`}{Email: "user@example.com"},
},
{
    name: "invalid email",
    structWithTag: struct{Email string `validate:"email"`}{Email: "not-an-email"},
    wantErrContains: "email",
},
{
    name: "empty string",
    structWithTag: struct{Email string `validate:"email"`}{Email: ""},
    wantErrContains: "email",
},
{
    name: "type mismatch (int)",
    structWithTag: struct{Email int `validate:"email"`}{Email: 123},
    wantErrContains: "validator.type_mismatch",
},
```

## 5. Test Runner Pattern

Use this loop for all test tables:

```go
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        err := validator.NewValidator(nil).ValidateStruct(tt.structWithTag)
        if tt.wantErrContains == "" {
            require.NoError(t, err)
        } else {
            require.Error(t, err)
            require.Contains(t, err.Error(), tt.wantErrContains)
        }
    })
}
```

## 6. Additional Guidelines

- Each rule should have its own focused test function.
- No skipped or commented-out tests.
- Use `require` for assertions.
- No `examples.go` or unrelated files created.
- Keep test cases short and focused: one assertion per case.

## 7. Type Mismatch Cases

- Always include a test where the validated field is of a clearly wrong type (e.g., `int`, `bool`, or a struct).
- Use a non-zero value for non-string types to ensure the validator is called.

## 8. Error Message Checking

- Use substrings that are unique to the rule (e.g., `"email"`, `"uuid"`, `"type_mismatch"`).
- Avoid checking for the entire error message to keep tests robust against message changes.

---

## Template for a New Rule Test

```go
func Test<RuleName>Validator(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name            string
        structWithTag   any
        wantErrContains string
    }{
        // Add cases here
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.NewValidator(nil).ValidateStruct(tt.structWithTag)
            if tt.wantErrContains == "" {
                require.NoError(t, err)
            } else {
                require.Error(t, err)
                require.Contains(t, err.Error(), tt.wantErrContains)
            }
        })
    }
}
```

---

## Checklist Before Submitting Tests

- [ ] Each rule has its own `Test<RuleName>Validator` function.
- [ ] Table-driven structure with `structWithTag` and `wantErrContains`.
- [ ] All relevant cases (valid, invalid, empty, type mismatch, edge) are covered.
- [ ] No skipped or commented-out tests.
- [ ] Uses `require` for assertions.
- [ ] No `examples.go` or unrelated files created.
