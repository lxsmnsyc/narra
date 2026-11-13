# Log: std-terminal

- Base cb834b6e (feat/batch-14). Compiler built with `sh selfhost/bootstrap.sh <out>` from a `git archive` of the base.
- Proof: `narrac run testdata/terminal_styles.nra` with the base compiler: `unknown module "std/terminal"`.
- `scripts/test.sh --suite generated --update` can run with the base compiler before the branch compiler exists (the test runner does not import std/terminal), which breaks the loop: runtime text, then bootstrap.
- Std sources are embedded, so every edit of `std/src/terminal.nra` needs a regenerate and a bootstrap (about 8 minutes on a busy machine). I iterated on the style functions in one file, with the two `Tty` calls stubbed, using the compiler of the first build.
- A field default must be a constant: `foreground: Color = Color.Default` fails ("a field default must be a constant"), also through a top-level `const`. `Option<Color> = None` works, so `Style` uses it and `Color` has no `Default` variant.
- Terminal check by hand: the sample binary under `script -q /dev/null` reports `isTerminal` true and `useColor` true; `columns` is `None` there because the pty has width 0.
- `Env.args()[0]` is the program path, so `terminal_detect.nra` runs itself through `env -u NO_COLOR -u FORCE_COLOR -u TERM <set> <program> child` with `Process.run`; the child's stdout is a pipe.
- picocolors only reopens a style after an inner close of the same code, so `bold(red("x") + " done")` has no extra codes: red closes with 39. The reopen shows in `bold` around `dim` (both close with 22) and in `red` around `red`.
