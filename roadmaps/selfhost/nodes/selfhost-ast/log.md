# Log: selfhost-ast

- Format uses node lines plus slot lines (`.name`) so optional and repeated children stay unambiguous. Single-list nodes (block, template, array, pat-*) print children directly.
- `pat-ctor` prints `bare` when Go's Args is nil, so `Circle` and `Circle()` differ. Narra side stores `bare: bool` plus an array.
- CharLit.value is a string in Narra to avoid char/code conversion.
- Dead end: a parameter list `f: () => U uses E, mutable out: ...` fails to parse in Go (paramAfterComma does not see `mutable`). Not fixed here; the samples avoid it.
- Unreferenced modules in selfhost/src are type-checked by `narra build`, so TestSelfhostLexer already compiles astdump.nra. I also ran it once with a hand-built AST in a scratch main; output matched the Go format.
