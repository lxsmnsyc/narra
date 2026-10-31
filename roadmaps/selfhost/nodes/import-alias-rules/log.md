# Log: import-alias-rules

- Worktree started before the brief commits; fast-forwarded the branch to local main (docs only).
- `ImportDecl.Names` is now `[]ImportName{Name, Local}`. Dumps print `a as b` for a rename, in Go and Narra.
- Alias fix edits the import alias and every `Alias.` identifier found by a token scan (`File.dotUses`). A template string that mentions the name gets no fix, because the parser keeps no positions inside `${}`.
- Alias equal to a type name: prelude namespace types (`namespaceModules` keys), types, generics and traits of the module. The suggestion is the plural of the std path, else the alias plus "s".
- Duplicate alias: an error unless another file of the module binds the same alias to the same path.
- Brief says `import { map as mapArray } from "std/array"`, but std/array has no `map`. The sample uses `std/option` and `std/result` map, plus `length as count` from std/array.
- First migration run touched error tests that only mention the import in a comment, and string literals such as "Array.filled" in selfhost/src/stdlib.nra. Fixed the script (exact import line, no rename after a quote) and reran.
- Narra port: rename in parser, dump, scopes, declsdump and the duplicate-alias error. Not ported: alias equals a type name, name clashes (listed in README).
