# Log: type-aliases

- Branch feat/type-aliases reset to origin/feat/batch-3 (6924411).
- Design: a non-generic alias is resolved lazily (cycle state per alias) and entered in the scope's type table as its resolved type, so constructors, patterns and struct literals work through it with no other change. A generic alias is kept in `ModScope.aliases` and its body is resolved again at each use with the arguments.
- ctors.nra and exprs.nra `exportsEnum` accept an alias declaration, so `Api.Token.Eof` works through a re-export.
- Diagnostics: `Checker.aliasNames` maps a type key to the alias last used for it; `describeType` prints `Alias (Type)` for struct, enum and distinct types. Primitive targets (`type Id = int`) are not decorated, to avoid noise.
- Full `sh scripts/test.sh` passed. Only golden changed: testdata/query/basic/queries.txt (an alias appended to text.nra).
