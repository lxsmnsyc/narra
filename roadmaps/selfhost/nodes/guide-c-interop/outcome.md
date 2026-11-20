# Outcome: guide-c-interop

## Built

- `docs/guide/16-unsafe-code-and-c.md` replaces the stub and keeps its `Next:` line. The running example finds the long side of a right triangle with `hypot` from `libm`, then moves on to C strings and memory.
- Sections: Calling a C function, The `unsafe` effect, Stopping the effect with `@trusted`, C types, C strings, Pointers (with "Memory that C allocates"), Linking a library (with "A library for one system"), Finding the unsafe code, Not built yet.
- Audit rows covered, each with a sample the `guide` suite runs:
  - 142, 159: `unsafe { }` blocks and the `unsafe` effect. The first sample is the E0996 error, then E0800 on an exported function and E0997 on a `@trusted` function with no block. Traits, generators and function values get one sentence each.
  - 160: `Ptr<T>` and `std/ptr`, with a table of the nine functions, a `Ptr.alloc` sample and the E0998 error for `Ptr<string>`.
  - 161: `extern { } from`, the C type table, E1002 for a `string` parameter, `@extern("abs")`, linking and the variadic limit.
  - 162: `std/cstring` with `strtol`, `strerror`, and `strdup` plus the C `free`.
  - 163: `narrac audit`, shown as a `sh` block with its output for the chapter's `triangle` package.
  - 182: `[link] search` in the `triangle` package's `narra.toml`.
  - `@target` on an extern block, with a Windows-only `kernel32` block that runs on macOS and Linux.
  - 164: `@repr` and C-callable functions are not built, so they get one prose paragraph and no sample.

## Deviations

- The chapter briefly teaches the sized integer types (`i8` to `u64`, `Int.from`, `Int.wrap`). No chapter from 1 to 15 teaches them, and C signatures need them.
- `Ptr.element` and `FixedArray` are left out. No earlier chapter teaches `FixedArray`.
- No sample uses a C file. Narra cannot compile C sources in a package yet, so the samples use `libc` and `libm` only.
- The `[link] search` sample names `vendor/lib`, a folder that does not exist. `cc` accepts a missing `-L` folder on macOS and Linux, and `libm` needs no search folder.

## Stubs left

- None in this chapter.

## Gotchas

- The site highlighter (`website/src/highlight/narra.ts`) and the TextMate grammar (`editors/textmate/narra.tmLanguage.json`) disagree on some capitalized names, and `pnpm --dir website build` fails on any such name in a fence:
  - The site marks `I32`, `U8` and `U64` as constants, and TextMate marks them as class names.
  - TextMate marks `CStrings` and `CString` as identifiers, and the site marks them as class names.
- The chapter works around both. It imports `std/cstring` as `Cstrings`, never writes the `CString` type, and converts with `Int.from` and `Int.wrap` in place of the `I32` and `U8` modules. A grammar fix would let the chapter use the names from SPEC section 9.
- `const x = unsafe { ... }` gives `expected the end of the statement, found "{"`. That matches the spec (the block is not an expression), but the message does not say so.

## Proposed learnings

- Check guide fences with `pnpm --dir website build` before the `guide` suite | the highlighter check rejects aliases such as `I32` that compile and run | `docs/guide` | yes, the website build already does

## Verification

- Every sample was checked by hand with the prebuilt `narrac` (`check`, `run -sanitize` with `NARRA_CHECK_LEAKS=1`, `run -release`). `narrac audit triangle` printed the two lines shown and exited 0.
- `JOBS=2 NARRAC=<prebuilt narrac> sh scripts/test.sh --suite guide`: PASS, 356 cases.
- `pnpm --dir website build`: exit 0, highlighter check 0 problems.
- `check_prose.py docs/guide/16-unsafe-code-and-c.md`: ok.
- Not run, as the orchestrator asked: `selfhost/bootstrap.sh` and the full `scripts/test.sh`.

## Size

- Estimated M, actual M.
