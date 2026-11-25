# Log

- Design: `Process.run` becomes a Narra function in `std/src/ext__process.nra` (the Narra part of the builtin module) with `options: RunOptions = RunOptions {}`. The builtin it calls is `Spawn.run` in the new module `std/internal/process` (runtime `std_process.c`, now loaded for that module path).
- Wasi: the E0991 rule moved from the builtin name to the function `std/ext/process|run`; the message still says `Process.run`.
- Chicken and egg: the seed does not know the Narra part, so `stdtext.nra` was regenerated with the seed compiler (`NARRAC=<narrac0> ... --suite generated --update`), then a compiler was built from the new source and the docs were regenerated with it.
