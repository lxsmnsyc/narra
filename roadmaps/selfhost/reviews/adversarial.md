# Adversarial review (fresh-context agent), with dispositions

1. modules missed emitter, builtins and type files; C names were not module-aware. Accepted: added the files to modules' shared list and the qualified C name rule to module-graph.
2. module-graph contract too vague. Accepted: added `Binary`, `LinkConfig`, entry points, layout rule, name mangling, and the single-file rule. Removed `[build] generics` (layout-tables defines it).
3. Feature nodes missed files (maps, closures, patterns, traits, opaque-distinct, effects). Accepted: added to each node's shared list.
4. No mechanism loads per-module runtime C; strings-std changes the string layout. Accepted: `Runtime` field in the builtins contract, built by split-compiler; strings-std now shares `runtime.h`.
5. builtins contract did not match the code (`where` argument, mutable arguments, prelude and namespaces). Accepted: contract rewritten around the existing `Builtin` struct.
6. diagnostic contract (sorting, file name, parser errors). Accepted: emission order kept, driver fills `File`, lexer and parser errors wrapped.
7. Edges. Accepted: box-recursive now depends on traits; json-diagnostics depends on modules, traits, closures and patterns; strings-std depends hard on split-compiler and split-compiler owns only the std files it creates. Rejected: an edge from effects to io-std, maps and strings-std. The builtins contract now requires every node to set `Effects` when it adds a builtin. Rejected: making the split-compiler edges `contract`. split-compiler is a gate on purpose.
8. trait-model C names invalid for primitives. Accepted: a table of C names per type and the borrowed-argument rule.
