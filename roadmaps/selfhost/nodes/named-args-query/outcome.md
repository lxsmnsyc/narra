# Outcome: named-args-query

## Built

- `narrac query refs FUNCTION.PARAMETER` (for example `refs app/shapes.box.width`, `refs box.width`): one line per named argument that passes the parameter, at the label's position (`name` start to end), with the enclosing body, in file and position order. `--json` gives the usual `uses` list; `name` is the qualified function plus `.PARAMETER`. A label inside a template string reports the template's range, as other sites do.
- Errors: `unknown name`, `X is a <kind>, not a function`, `X has no parameter "p"`.
- `callers FUNCTION` and `refs FUNCTION` already list labelled calls at the callee name; goldens cover them.
- `type-of` on a label has no answer (`no expression at ...`); the argument's value still answers.
- New site kind `SiteKind.LabelSite` in `querysites.nra` (nodeId is the call, name the label).

## Tests

`testdata/query/named_args/` (package with `queries.txt`): refs of parameters (qualified, bare, JSON, missing parameter, label in a template), refs and callers of a labelled function.

## Gate decisions

- Syntax for a parameter target: `FUNCTION.PARAMETER`, tried only when the whole name finds no declaration. Alternative: a new `param` command or `FUNCTION(PARAMETER)`; rejected because `refs` already takes dotted names (`Shape.Circle`) and the fallback changes no existing answer.
- Scope: refs lists labels only, not reads of the parameter inside the function body. Alternative: also list body uses; deferred because a rename needs the labels from this query and body uses come from the local `refs` of the declaring site.
- Only `Ref.Function` callees (plain, generic, imported); trait and implementation functions are not targets of `refs` today.

## New diagnostic codes

None.
