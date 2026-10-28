# Log: split-compiler

- Split check.go and emit.go with a script that moves whole top-level declarations by name, so the bodies are byte-identical. Checked emitted C for every file in testdata (normal and test mode, including error text) before and after: identical.
- Dead end: moving existing runtime.h helpers into per-module Runtime would change the emitted C text, so the mechanism exists but the three modules leave Runtime empty.
- Review fix: the module hint keeps its old order through a fixed first list, because init order is by file name.
- Lexer and parser errors are wrapped once, inside Parse (parser.go, outside the owned paths, 8 lines).
