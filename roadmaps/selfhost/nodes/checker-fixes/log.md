# Log: checker-fixes

- Source facts the Go AST keeps (header line, source text, alias and name ends, `dotUses`, name positions, call ends) go in `File.facts` (ast.nra), filled by the parser; the checker reads them from `ModScope.facts`.
- Go's `hoist` flag is a counter on the use collector (`blocked`), not a parameter of every walk.
- `throw` fix: the parser parses the expression speculatively (index, id counter and failed flag restored).
- compare.py: found one gap, a call inside a template string (Go's fix text there is odd, `\``); matched by taking the call end as token position + 1 and merging the inner parser's call ends.
- check-codes.py copied with the fix reporters added.
