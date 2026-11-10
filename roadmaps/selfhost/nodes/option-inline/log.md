# Log: option-inline

- Base: feat/batch-4 e64faa2. Base compiler built with `sh selfhost/bootstrap.sh` into scratch.
- Base instructions retired (median of 3, load 4): compiler 3.662 G (130 MB), fib 1.213 G, sieve 2.291 G, matmul 2.447 G, strings 1.096 G, interp 1.937 G, particles 2.612 G.
- Layout decision in `check.nra` (`decideHeapEnum`, cached in `Checker.heapEnums` by `finish`). Emitter reads it through `isHeapEnum`.
- No size limit: the self-build has 85 inline enums with fields and 7 heap ones. C sizes of the inline ones: most 16 to 48 bytes, the largest 704 (a Result of a checker-sized struct) and 808 (an Option of the Emitter). The compiler took 3.869 G instructions, 5.6% more than the base: large values are copied by value and each retain touches every counted field.
- Found on the base: `struct Node { next: Link }` declared before `enum Link { Next(node: Node) }` fails with E0109 ("contains itself"). `visitOrder` visits a heap enum's fields while the struct that holds it is still open. Fixed: a heap enum skips fields whose type is still being visited.
- Size limit sweep (compiler instructions on e64faa2, cache of counted types in place): all heap 3.53 G, 8 bytes 3.48, 16 3.46, 32 3.45, 48 3.51, 64 3.56, none 3.67. Chose 32.
- interp: 1.4% more instructions, about 20% more user time; clang if-converts the Add/Sub arms (csel chain on the accumulator). A switch shape and a pointer read did not help.
- Mutants (inline retain off, inline release off): all four samples fail under ASan or the leak check.
- Rebased on feat/batch-4 8de99b4 (conflict in the emitstate imports with cache-type-names). Branch: compiler 3.281 G to 3.084 G.
- Full check: declsdump goldens `nested_generics` and `shadow_prelude` changed order (an inline Option now follows its field type); updated.
