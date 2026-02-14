# Go Subcommand Evaluation (v0.0.17)

This document details the evaluation of `go-subcommand` (specifically `gosubc` v0.0.17) for generating code in this repository.

## Findings

The current version of `gosubc` has several critical bugs and limitations that prevent it from being used to replace the existing CLI implementation without significant manual intervention or workarounds.

### Bugs

#### 1. Missing Package Qualifier in Generated Calls

When generating code for a subcommand defined in a package (e.g., `cli`), the generated `root.go` file correctly imports the package but fails to use the package qualifier when calling the subcommand function.

**Example:**
Definition in package `cli`:
```go
package cli
// JsonSimplePath is a subcommand ...
func JsonSimplePath(...) {}
```

Generated `root.go` (package `main`):
```go
import "example.com/test" // imported as cli

// ...
JsonSimplePath(...) // Error: undefined: JsonSimplePath
// Should be: cli.JsonSimplePath(...)
```

#### 2. Missing Embedded Templates

The generated code includes an `embed` directive for `*.txt` files in a `templates` directory, but `gosubc` does not generate any matching files, causing compilation failure.

**Error:**
`pattern *.txt: no matching files found`

**Solution:** `gosubc` should either generate default templates or allow users to provide them easily (and document where).

#### 3. Unescaped Quotes in Descriptions

If a parameter description contains double quotes (e.g., in a default value description), `gosubc` fails to escape them in the generated Go string literal, causing syntax errors.

**Example:**
Comment: `// delim: -d output delimiter (default "\n")`
Generated code: `c.StringVar(..., "output delimiter default "\n"")`
**Error:** `illegal character U+005C` or syntax error.

#### 4. Module Root Requirement / Non-Recursive Scan

`gosubc` requires the target directory to contain a `go.mod` file and does not appear to scan subdirectories recursively. This makes it difficult to use in a repository with multiple commands or where commands are defined in subpackages, unless each is a module.

## Conclusion

Upgrading to `go-subcommand` v0.0.17 is not recommended at this time due to the issues listed above. The tool requires bug fixes regarding import handling, template generation, and string escaping before it can be reliably used.
