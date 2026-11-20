# Log: thread-scope-void

- Base (feat/batch-21 at 77674c71 plus batch-21 merge), compiler built with `sh selfhost/bootstrap.sh`:
  `narrac run main.nra` on the brief's program gave `std/src/thread.nra:42:3: this expression has no value`.
- Cause: the std body of `Thread.scope` held the body's result in `const result = body(opened)`. In the instance with `T = void` the checker rejects a void local (E0225), and the error sits in std.
  `Thread.start` avoids it by storing `Some(work())` in a `Box<Option<T>>` and reading it with `Threads.result`, which the checker and emitter already accept for void.
- Fix: `Thread.scope` keeps the result the same way. Regenerated `selfhost/src/std/stdtext.nra` with `--suite generated --update`.
- Branch: the brief's program prints `done`; `testdata/threads_scope_void.nra` prints `total 3` and `done`; base prints the std error for it.
