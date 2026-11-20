# Outcome: std-header-comments

## Files changed

- `std/src/`: `arena.nra`, `args.nra`, `async.nra`, `atomic.nra`, `channel.nra`, `iterator.nra`, `parallel.nra`, `regex.nra`, `shared.nra`, `template.nra`, `terminal.nra`, `thread.nra`, `toml.nra`, `unicode.nra`. Comments only.
- `selfhost/src/std/stdtext.nra`: regenerated with `sh scripts/test.sh --suite generated --update`. The generated reference pages did not change.

Two citations carried a fact that the sentence did not say, so it is written out:

- `atomic.add`: "The sum wraps, like +, and no atomic operation panics." (was "(D92)").
- `toml.writeTable`: "path is the path from the root used in error messages, such as `.a.b[2]`" (was "the D84 path").

The other citations (SPEC section, D109, D110 item 5, D111, D114, D118) were dropped because the sentence already states the fact, for example "Importing this module builds the program in thread mode".

## Acceptance

- [x] No citation in std comments.
  `grep -nE 'D[0-9]{2,3}|SPEC' std/src/*.nra` now prints only `std/src/random.nra:27`, a hex constant (`0x94D049BB133111EB`) in code, not a comment. With a word boundary (`D[0-9]{2,3}\b`) the grep prints nothing.
- [x] `selfhost/src/std/stdtext.nra` regenerated (`--suite generated --update`: PASS, 26 cases).
- [x] `sh scripts/test.sh`: PASS, 10925 cases in 480s. An earlier run, started while the website build ran, printed `FAIL after 399s`; its output was cut by `tail`, and the rerun alone passed. `pnpm --dir website build` exits 0.
