# Wave 7

## Merged

- selfhost-check-bodies: [#37](https://github.com/lxsmnsyc/narra/pull/37), merged by the maintainer. L. The token count was lost to a context clear, so the brief has no `spent`.
- selfhost-names-frontend: [#38](https://github.com/lxsmnsyc/narra/pull/38), merged by the maintainer. Estimated L, actual M, about 288k tokens. One fresh review (one should-fix: abbreviated locals) and one fix round.
- selfhost-names-checker: [#39](https://github.com/lxsmnsyc/narra/pull/39), merged by the maintainer. Estimated L, actual M, about 309k tokens. One fresh review (clean, three comment nits) and one comment-only fix round.

## Plan changes

- The selfhost-check-patterns brief now names `findTupleType` and `listFields`, the Narra names of Go's `tupleOf` and `fieldsOf`.
- Exported AST fields (`id`, `op`, `col`, `params`, `args`, `decls`, `stmts`, `funcs` and others), idcheck's `funcIds`, `blockIds`, `exprIds`, and `Diag`/`diagText` keep their short names. The selfhost-names-frontend outcome lists them. No node owns that rename yet; it touches every module, so it waits for a quiet point between nodes.
- Checker struct fields of three or more letters (`Checker.cur`, `Checker.errs`, `BodyInfo.targs`) were kept too.

## State

- Active: selfhost-check-calls (L, opus) and selfhost-check-patterns (M, sonnet). They share several files; merge calls first, then patterns rebases.

## Process lessons

- A rename pass is checked well by diffing every string, char and template literal against main. It caught one template text change in the front end.
- Narra rejects shadowing even in pattern bindings; the rule is in CLAUDE.md now.
