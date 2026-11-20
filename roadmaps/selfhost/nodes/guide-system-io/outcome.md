# Outcome: guide-system-io

## Built

- `docs/guide/11-files-and-the-system.md`, about 560 lines, 10 runnable samples. Sections: read a file, build paths, folders, standard input and output, the command line and the environment, the `lines` program, other programs and exit codes, bytes, measure time.
- Audit rows covered: 59 (`std/fs`), 60 (`std/io`), 61 (`std/env`, with the Windows note on `Env.args()`), 62 (`std/path`), 63 (`Process.run`, `Process.exit`), 121 (`bytes`, `std/bytes`, `Fs.readBytes`, `Fs.writeBytes`), 67 (`std/time`).
- The running example is `lines`, a line counter over files or standard input.

## Deviations

- The guide suite starts samples with no arguments and an empty stdin, so the file samples write their own files in a scratch folder under `TMPDIR`. The `lines` program prints `0 (standard input)` there; the run over named files is an unchecked `sh` block, which I ran by hand.
- Samples name the scratch folder with `Time.unixMillis()`, so the debug and release runs of one sample cannot collide.
- `Fs.listDir` order is not promised and the std has no sort, so the sample prints the entry count.
- Not covered: `std/async/file` (chapter 14), `Process.run` signal forwarding (named in one line only), UNC paths.

## Stubs left

- None.

## Gotchas

- `narrac run` closes stdin unless given `-stdin FILE`, and shows all stdout before stderr. The chapter says so.
- The `Next:` line points to chapter 12.

## Verification

- `NARRAC=<prebuilt> sh scripts/test.sh --suite guide`: PASS, 354 cases.
- `pnpm --dir website build`: passes.
- plain-writing checker on the chapter: ok.

## Size

- Estimated M, actual M.
