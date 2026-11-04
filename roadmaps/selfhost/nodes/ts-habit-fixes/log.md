# Log: ts-habit-fixes

- The work did not exist on origin/main (98a517e). Started from the brief.
- Codes: E0923 class, E0924 interface, E0925 new, E0926 try/catch, E0927 throw, E0928 this. `class`, `interface`, `new`, `try`, `throw` and `this` stay ordinary identifiers in the lexer; the parser recognizes them in `decl()` (declaration start) and `primary()` (expression start), so a variable with such a name in a legal position is unaffected.
- `try` is reported only before `{`; `throw` only before a name, literal or template; `new` only before a name.
- The self-hosted parser has no fixes (messages and positions only), so the fixes live in Go. The parse-error condition is the same in both parsers.
- `for ... in`: the fix always writes `Maps.keys(expr)`; the parser cannot know the type. Fix speculates an expression parse (`speculateExpr`) to find its end.
- Selfhost keeps parsing after a failure and keeps the first one, so the first-error rule matches Go's panic.
