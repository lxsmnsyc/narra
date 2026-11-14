# Log: uses-list-design

- D117 did not exist in DECISIONS.md (last record D109 on this base).
- Reproduced at `feat/batch-13` with a stage 2 narrac: `Map<Generator<int> uses io, int>` dumps `uses=io,int`; `#[Generator<int> uses io, Array<int>]` gives `expected "]", found "<"`; a trailing comma after `g: Generator<int> uses io` gives `expected a name, found ")"`; `(x: () => int uses io, int) => int` dumps `uses=io,int`; `function f(): () => int uses io, alloc` puts both effects on the result type, none on the function.
- `parseUsesList` has three callers: the generator branch of `parseType`, `parseFuncType`, `parseFuncDecl`. Only `isParamAfterComma` limits the type lists today.
- Counts with `narrac parse` over every parsing `.nra` file (script in /private/tmp/claude-501/uses-list-design/count.py): declarations 2991 (447 with more than one effect), types 250 (27 with more than one: 0 in selfhost/src, 0 in std/src, 2 in selfhost/tests registry.nra, 25 in testdata). No seed step needed.
- fmt works around the trailing-comma bug: `endsWithUses` drops the trailing comma of a broken list whose last item ends with `uses` (5 call sites).
- Type display goes through `usesSuffix` (check.nra), shared by generator names, declsdump and query type-of; about 19 goldens print a two-effect type.
- Found, not in scope: `Generator<int> uses /* c */ io` formats as `Generator<int /* c */> uses io` (the `uses` keyword has no trivia node).
- E1240 first had an automatic fix; dropped it, since the roadmap rule forbids fixes for old forms. Whether E1240 stays at all is a maintainer question.
- spec suite PASS 62 cases, sources suite PASS 1 case (no fenced SPEC samples changed).
