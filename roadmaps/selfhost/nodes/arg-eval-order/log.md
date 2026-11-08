# Log: arg-eval-order

- Repro on origin/main 6e42877: `cd compiler && go run ./cmd/narra run testdata/arg_eval_order.nra`
  prints `1, 55, first, second, 12, 101001, wx`; expected `3, 15, first, second, 12, 101001, ab`.
  Reproduced: a block argument runs before an earlier `let` read (user call `show`, `pair`, builtin `Str.slice`).
  Calls with side effects (`say`) and the self tail call already follow written order.
