# Frontend Architecture

This document describes the intended layering for building query frontends on top of Lookup.

## Layering

```
frontend syntax/parser
    -> frontend-specific AST/semantics
    -> generic Lookup Runner/Pathor/Scope substrate
    -> user data / custom Pathor implementations
```

## What belongs in Lookup

*   **Execution Primitives:** `Runner`, `Pathor`, `Scope`.
*   **Navigation & Composition:** `Find`, `Index`, `Chain`, `NestChain`.
*   **Data Structures:** Constants and standard arithmetic/comparison runners.
*   **Context Plumbing:** Evaluator context, functions, and error handling structure (`Invalidor`).

These provide the universal substrate that any query language or DSL can target.

## What remains frontend-specific

*   **Syntax parsing and AST generation:** e.g., mapping a DSL to `lookup.Runner` tree.
*   **Language-specific error handling/swallowing:** JSONata, for example, swallows certain missing-field errors and treats them as `Undefined`.
*   **Specialized sequences and truthiness:** JSONata's `Sequence`, `Array`, `Undefined`, sequence flattening rules, `Materialize`, and its custom truthiness semantics remain inside the `jsonata` package.

## Integration

Any new frontend should integrate by translating its abstract syntax tree into a chained sequence of `lookup.Runner` implementations, utilizing `lookup.NestChain` (for hierarchical depth traversal) or `lookup.Chain` (for parallel step progression).

The `examples/frontend_proof` directory contains an example of a small DSL targeting Lookup generic runners without importing JSONata.

## Criteria for Extracting Modules

`jsonata` currently exists as a package within `lookup` because they co-evolved and share a release cycle. Splitting `jsonata` into an independent module or repository could be justified in the future if:

1.  **Independent Release Cadence:** The JSONata feature set and standard function library require updates that should not bump the version of the `lookup` generic API.
2.  **Conformance:** The JSONata test suite and conformance requirements heavily diverge from core `lookup` testing strategies.
3.  **Proven External Consumability:** The generic `lookup` APIs are proven to cleanly support multiple distinct production frontends, cementing the boundary logic.
