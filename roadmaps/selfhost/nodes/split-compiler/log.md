# Log: split-compiler

- Split check.go and emit.go with a script that moves whole top-level declarations by name, so the bodies are byte-identical. Checked emitted C for every file in testdata (normal and test mode, including error text) before and after: identical.
- Dead end: moving existing runtime.h helpers into per-module Runtime would change the emitted C text, so the mechanism exists but the three modules leave Runtime empty.
- Doubt: the "unknown module" hint now lists registered modules sorted ("std/float, std/int and std/string"); it was "std/int, std/float and std/string". No test covers it.
- Lexer and parser errors are wrapped once, inside Parse (parser.go, outside the owned paths, 8 lines).
