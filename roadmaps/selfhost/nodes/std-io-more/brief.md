---
id: std-io-more
title: stdin, directories and paths
status: done
priority: should
type: feature
story:
early: false
outline: false
milestone: m10-std
target:
ship: auto
from:
size: M
spike: false
hypothesis:
verify: test -f compiler/testdata/io_stdin.nra && test -f compiler/testdata/path.nra && cd compiler && go test -count=1 ./internal/narra -run 'TestPrograms|TestCompileErrors|TestSelfhostBuiltins|TestSelfhostEmit|TestSelfhostRuntimeText'
exclusive: []
fulfilled_by:
gate: false
added: 2026-11-01
spent:
owns:
  - compiler/internal/narra/std_io.go
  - compiler/internal/narra/std_path.go
  - compiler/testdata/io_stdin*
  - compiler/testdata/io_dirs*
  - compiler/testdata/path*
  - compiler/testdata/errors/io_*
  - compiler/testdata/errors/path_*
  - roadmaps/selfhost/nodes/std-io-more/**
shared:
  - compiler/internal/narra/std_fs.go
  - compiler/internal/narra/runtime_io.h
  - compiler/internal/narra/run_test.go
  - selfhost/src/stdlib.nra
  - selfhost/src/stdcalls.nra
  - selfhost/src/emitstd.nra
  - selfhost/src/runtimetext.nra
  - compiler/README.md
  - SPEC.md
reserves: []
provides: []
depends_on:
  - on: selfhost-bootstrap
    kind: hard
    why: std grows on the self-hosted compiler, so nothing is built in Go and ported
---

# stdin, directories and paths

## Goal

Fill the I/O gaps that real tools hit, as std builtins in both compilers: read stdin (lines, exact byte counts for LSP framing, all of it), write to stdout without a newline and to stderr, flush stdout, create and remove directories and remove files, and a new `std/path` module. Fallible functions return `Result<T, string>` with the `<subject>: <reason>` message, as in `std/fs` (D57).

## Context

- Pattern to copy: `compiler/internal/narra/std_fs.go` (Check through `checkFixed`, Emit through `ioResult` / `ioEither`, C in `runtime_io.h` sections) and its port in `selfhost/src/stdlib.nra` (`STD_FUNCS` rows), `selfhost/src/emitstd.nra` (`FixedCall` rows and the `Fs.readText` style emitters).
- Std builtins live in both compilers. `TestSelfhostBuiltins` fails when the registries differ, `TestSelfhostEmit` compares the C of both compilers on every `testdata/*.nra`, and `TestSelfhostRuntimeText` regenerates `runtimetext.nra` from the Go runtime files (`NARRA_UPDATE=1 go test -run TestSelfhostRuntimeText`; do not edit that file by hand).
- These are primitives (syscalls), so they stay builtins whatever std-source decides. `std/path` is pure string work; see the notes.
- The language service (`language-service`) reads framed JSON-RPC messages from stdin and writes them to stdout. That is the main user.
- Source: GOAL.md decision 2026-11-01 (m10-std), D57, README "Supported" line on `std/fs`.

## API

New module `std/io`, imported as `import * as Io from "std/io"`:

| Function | Result | Effects | Meaning |
| --- | --- | --- | --- |
| `Io.readLine()` | `Result<Option<string>, string>` | io, alloc | next line of stdin without the `\n`; `None` at end of input; Err on a read error or invalid UTF-8 |
| `Io.readBytes(count: int)` | `Result<string, string>` | io, alloc | exactly `count` bytes; Err at an early end of input, a negative count or invalid UTF-8 |
| `Io.readAll()` | `Result<string, string>` | io, alloc | the rest of stdin |
| `Io.write(text: string)` | `void` | io | stdout, no newline |
| `Io.writeError(text: string)` | `void` | io | stderr, no newline |
| `Io.flush()` | `void` | io | flushes stdout |

Additions to `std/fs`, in its existing descriptive style. The changing functions return `Result<bool, string>`: `Ok(true)` when the call changed the filesystem, `Ok(false)` when there was nothing to do, and Err for a missing path or any other failure. This follows `writeText`, whose Ok holds a meaningful value.

| Function | Result | Effects | Meaning |
| --- | --- | --- | --- |
| `Fs.isDir(path: string)` | `bool` | io | true for an existing directory |
| `Fs.createDir(path: string)` | `Result<bool, string>` | io, alloc | creates one directory (`mkdir`, mode 0777 before umask); `Ok(true)`; Err when it exists or the parent is missing |
| `Fs.createDirAll(path: string)` | `Result<bool, string>` | io, alloc | creates the directory and missing parents; `Ok(true)` when it created any, `Ok(false)` when the directory already exists; Err when a part is a file |
| `Fs.removeFile(path: string)` | `Result<bool, string>` | io, alloc | `unlink`; `Ok(true)`; Err when the path is missing or a directory |
| `Fs.removeDir(path: string)` | `Result<bool, string>` | io, alloc | `rmdir`, empty directories only; `Ok(true)`; Err when missing or not empty |

New module `std/path`, imported as `import * as Path from "std/path"`. Names and results follow Node.js `path` (POSIX). Pure (alloc only), `/` separators only. Narra has no optional parameters (SPEC section 2), so `basename` takes only the path; Node's suffix argument is left out (decided 2026-11-05: callers strip `extname` with the string helpers).

| Function | Result | Meaning (Node.js behavior) |
| --- | --- | --- |
| `Path.join(left: string, right: string)` | `string` | joins with `/` and normalizes `.` and `..` segments; an absolute `right` is not special; `""` result is `"."` |
| `Path.dirname(path: string)` | `string` | everything before the last segment, trailing `/` ignored; `"."` when there is none; `"/"` for `"/a"` and `"/"` |
| `Path.basename(path: string)` | `string` | the last segment, trailing `/` ignored; `""` for `"/"` and `""` |
| `Path.extname(path: string)` | `string` | the last segment's suffix from its last `.`, with the dot (`".nra"`); `""` when there is no dot or the only dot is the first character |

C runtime: new sections in `runtime_io.h`: `io` (`narra_io_read_line`, `narra_io_read_exact`, `narra_io_read_all` on `stdin` with `getline` / `fread`; `narra_io_write`, `narra_io_write_error`, `narra_io_flush`), additions to `fs` (`narra_fs_is_dir`, `narra_fs_mkdir`, `narra_fs_mkdir_all`, `narra_fs_unlink`, `narra_fs_rmdir`), and `path` helpers (or inline emitted C, whichever keeps both compilers simplest). Reuse `narra_io_utf8`, `narra_io_fail` and `narra_io_no_nul`.

## Acceptance

- [ ] Sample programs, clean under ASan and the leak check:
  - `io_stdin.nra` with `io_stdin.in`: `readLine` until `None`, a line with `\r\n`, an empty line, the last line without `\n`, `readBytes` on a framed message (`Content-Length: N` header then N bytes, including multibyte characters), `readAll` of the rest, and `readBytes` past the end giving Err.
  - `io_dirs.nra`: create, list, `isDir`, remove a file and a directory in a temp directory (pattern of `io_files.nra`); `createDirAll` gives `Ok(true)` then `Ok(false)` on the same path; `removeDir` on a non-empty directory, `createDir` on an existing one and `removeFile` on a missing path give Err with the path in the message.
  - `path.nra`: the Node.js results below, each checked with `Test.eq` or printed (taken from Node 26 `path.posix`):
    - `extname`: `".bashrc"` -> `""`, `"index."` -> `"."`, `"a.b/"` -> `".b"`, `"/a/b.nra"` -> `".nra"`, `"a/b.tar.gz"` -> `".gz"`, `"/"` -> `""`, `""` -> `""`.
    - `basename`: `".bashrc"` -> `".bashrc"`, `"a.b/"` -> `"a.b"`, `"/a/b.nra"` -> `"b.nra"`, `"/"` -> `""`, `""` -> `""`.
    - `dirname`: `".bashrc"` -> `"."`, `"a.b/"` -> `"."`, `"/a/b.nra"` -> `"/a"`, `"a/b.tar.gz"` -> `"a"`, `"/"` -> `"/"`, `""` -> `"."`.
    - `join`: `("a", "/b")` -> `"a/b"`, `("a/", "b")` -> `"a/b"`, `("a", "../b")` -> `"b"`, `("", "")` -> `"."`, `("/", "a")` -> `"/a"`, `("a", "./b/")` -> `"a/b/"`.
  - `Io.write`, `Io.writeError` and `Io.flush` in one sample whose `.out` shows stdout without the extra newline.
- [ ] `run_test.go` feeds `<name>.in` to the program's stdin when that file exists; other samples keep `/dev/null`.
- [ ] Error tests: calling `Io.readLine` without the import (existing unknown-module error), and an exported function that calls `Fs.removeFile` without `uses io` (existing effect error). No new diagnostic codes are expected; if one is needed, ask the orchestrator for a code from E0935 up.
- [ ] `TestSelfhostBuiltins`, `TestSelfhostEmit` and `TestSelfhostBootstrap` pass: both compilers register and emit the same.
- [ ] SPEC.md line listing the std modules (section 7, "Standard modules") names `std/io` and `std/path`; `compiler/README.md` "Supported" describes the new functions.
- [ ] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

Uncertainty low (same pattern as `std/fs`), coupling low (one dependency, done), blast radius medium (shared registry tables in both compilers, `runtime_io.h`, `run_test.go`), verifiability low (automated tests), size medium (about 10 files, near 800 lines across Go, C and Narra), context medium. M.

## Notes for the worker

- Mixing stdio and raw `read(0)` loses buffered bytes. Read stdin only through `FILE* stdin` (`getline`, `fread`). `print` writes through `stdout` (`narra_print`), so `Io.write` must use the same stream; when stdout is a pipe it is fully buffered, which is why `Io.flush` exists.
- `Result<Option<string>, string>` nests an enum payload in an enum. `ioEither` builds one level; check that the Ok branch allocates the `Some` and that the leak check stays clean on `None` and Err paths.
- `readLine` keeps the `\r` of a `\r\n` line (the caller trims). LSP headers end in `\r\n`; document it in the README line.
- Runtime sections are emitted only for imported modules (`ioRuntime(section)`); `std/io` and `std/path` need their own section names. After changing `runtime_io.h`, regenerate `runtimetext.nra` with the update command above.
- Keep every new `STD_FUNCS` row in the same position as Go's builtin dump order (prelude, then `moduleList` order, functions by export name); `TestSelfhostBuiltins` prints the first difference.
- `Path.join` is Node's `join`, not `resolve`: `("a", "/b")` is `"a/b"`. Normalizing `..` past the start of a relative path keeps the `..` (`join("..", "a")` is `"../a"`), as Node does; check any other case against `node -e 'console.log(require("path").posix.join(...))'`.
- `std/path` is pure, so it is a candidate for Narra source if std-source answers yes. If std-source has landed with an approved answer when this node starts, write `std/path` as source in the form it chose and say so in the outcome; otherwise build it as builtins like the rest.
- `selfhost/src/modules.nra` has private `baseName` and `dirName`; leave them alone (moving the compiler to `std/path` is a separate change).


Merged as PR #91.
