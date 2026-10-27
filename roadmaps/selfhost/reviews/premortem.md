# Premortem

It is some weeks from now and the self-hosted compiler is late. Why?

1. **Merge conflicts in the shared compiler files stalled wave 2.** Traits, closures, modules and patterns all add cases to `parser.go`, `ast.go`, `clone.go`, `check_expr.go` and `emit_expr.go`. Change: split-compiler goes first and is a gate. Wave 2 merges in the order `plan.py status` prints, and each later branch rebases before review. Accept the remaining risk: the edits are new `case` lines and small additions.
2. **The enum representation change broke match and is.** box-recursive moves enum payloads to the heap, which changes every pattern binding and field read. Change: box-recursive's acceptance includes the existing enum, options and generics programs under the leak check. It runs after patterns lands so it changes the final pattern code once.
3. **The self-hosted port found features missing late.** The spike only runs after the core features. Change: selfhost-spike is a gate, and its outcome adds nodes before the port starts. Port order is lexer and parser first, which uses the fewest features.
4. **Reference counting bugs hid in new features.** Closures, maps and boxes add new heap objects. Change: every node's sample programs run under AddressSanitizer and the leak check, which already caught a missing release in milestone 1.
5. **Traits took longer than planned and blocked maps.** Change: maps depends on traits by `behavior` through the trait-model contract, so it can build against string and integer keys first.
