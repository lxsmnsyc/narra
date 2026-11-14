# Outcome: std-terminal

## Built

- `std/src/terminal.nra`, module `std/terminal` (alias `Terminal`), Narra source over two C primitives.
  - Detection: `enum Stream { Stdin, Stdout, Stderr }`, `isTerminal(stream): bool uses io`, `columns(stream): Option<int> uses io` (`None` when the stream is not a terminal or reports width 0), `useColor(stream): bool uses io, alloc` with the D118 order: `NO_COLOR` non-empty is false, `FORCE_COLOR` not `""` or `"0"` is true, `TERM=dumb` is false, else `isTerminal`.
  - Styles, pure with `uses alloc`: `bold`, `dim`, `italic`, `underline`, `inverse`, `strikethrough`, `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`, `gray` and `bgBlack` to `bgWhite`. Each closes with its reset code (22, 23, 24, 27, 29, 39, 49) and replaces an inner close of the same code by close plus open, so the outer style continues. `bold` and `dim` share 22 and reopen both.
  - `style(text, Style)`, `struct Style { foreground: Option<Color> = None, background: Option<Color> = None, bold, dim, italic, underline, inverse, strikethrough: bool = false }`, `enum Color { Black ... White, BrightBlack ... BrightWhite, Ansi256(index: u8), Rgb(red: u8, green: u8, blue: u8) }`. `stripStyles(text)` removes `ESC [ digits/; m` sequences and keeps every other escape.
- `std/internal/terminal`: `Tty.isTerminal(fd: int): bool` and `Tty.columns(fd: int): int` (0 when unknown), `uses io`. C in `selfhost/runtime/std_terminal.c` (`isatty`, `ioctl` with `TIOCGWINSZ`), linked only when the module is imported. Rows `INTERNAL_TERMINAL_FUNCS` and two emit cases in `selfhost/src/stdinternal.nra`; `Tty.` dispatch and runtime-prefix rows in `emitstd.nra`; case in `runtimetext.nra`; generated `runtimestdterminal.nra`.
- `selfhost/src/stdterminal.nra`: the module row (`check: "source"`), registered last in `stdlib.nra`. `std/terminal` is last in the unknown-module list (`testdata/errors/unknown_module.err`, `testdata/declsdump/imports/err_unknown_module.dump`). `stdtext.nra` regenerated.
- Samples: `testdata/terminal_styles.nra` (every kind of style, nesting, `style` with `Rgb`, `Ansi256` and named colors, `stripStyles` of each), `testdata/terminal_detect.nra` (runs itself through `env` with eight combinations of `NO_COLOR`, `FORCE_COLOR`, `TERM`, stdout on a pipe). Error: `testdata/errors/terminal_internal_import.nra`.
- SPEC (standard modules list, D118 bullet), `compiler/README.md`, guide section "Terminals" in `docs/guide/12-standard-library.md`.

## Deviations

- `Style.foreground` and `Style.background` are `Option<Color>` with default `None`, and `Color` has no `Default` variant. D118 says `Color = Color.Default`; the checker rejects a variant as a field default ("a field default must be a constant"). When constructor-field-defaults lands, the type can go back to `Color` with a `Default` variant.
- The brief's example "bold around text with a red part reopens the bold" does not hold in the codes: red closes with 39, which does not touch bold. The sample shows the reopen for `bold` around `dim` and `red` around `red`.
- The error test has no code: the checker prints the internal-module message without one, as for `std/internal/probe`.
- No codes from E1370-E1379 used. The seed is not regenerated.

## Stubs left

- None.

## Gotchas

- Std text is embedded in narrac: each edit of `std/src/terminal.nra` needs `--suite generated --update` and a bootstrap before a sample sees it.
- The runtime text changed, so the seed needs two updates (the orchestrator does them at the end of the batch). narrac does not use `std/terminal` yet.
- `columns` is `None` for a terminal that reports width 0 (a bare pty).
- `useColor` reads `NO_COLOR` before `FORCE_COLOR`, as D118 says; `FORCE_COLOR=1` with `TERM=dumb` is true.

## Proposed learnings

- Test a std module on a stubbed single file before the embedded rebuild | each std edit costs a regenerate and a bootstrap | std/src | no

## Verification

- Proof: base 6c8c5998 (the merge base; first measured on cb834b6e, which has no `std/terminal` either), `narrac run testdata/terminal_styles.nra`: `unknown module "std/terminal"`. Branch: prints `testdata/terminal_styles.out`, clean under `-sanitize`.
- `sh scripts/test.sh` before the last rebase: `test.sh: PASS, 8635 cases in 562s`. `sh scripts/linux-check.sh`: `linux-check: PASS in 1647s`. After the rebase on 6c8c5998 and the colon fix: `--filter terminal`, `--suite errors`, `--suite generated`, `--suite sources` (lines in the worker report).
- `stripStyles` also removes the colon form (`ESC[38:5:1m`); `terminal_styles` checks it.
- Code that does not use the feature emits the same C: `narrac emit selfhost` from the base and the branch compilers, byte-identical for `narrac emit selfhost` (base and branch compilers) and for `emit testdata/iterator.nra`.

## Size

- Estimated M, actual S to M.
