# Commands

Every `narrac` command and flag, with its exit codes, and the environment variables that `narrac` and the programs it builds read. `narrac --help` prints the same list of commands.

In a command line, `PATH` is a `.nra` file or a package folder (a folder with `narra.toml`). The root of a workspace stands for every member. A flag with a value takes it as the next word or as `-name=value`. The flags of `build`, `run` and `test` accept one or two dashes, so `-release` and `--release` are the same. Exit code 2 always means the command line is wrong.

## Commands

| Command | What it does |
| --- | --- |
| [`narrac build`](#narrac-build) | build the executables of a package or file |
| [`narrac run`](#narrac-run) | build and run a program |
| [`narrac test`](#narrac-test) | run the test blocks |
| [`narrac check`](#narrac-check) | report errors without building |
| [`narrac explain`](#narrac-explain) | explain a diagnostic code |
| [`narrac apply-fix`](#narrac-apply-fix) | print a file with the fix of an error applied |
| [`narrac fmt`](#narrac-fmt) | format `.nra` files |
| [`narrac query`](#narrac-query) | answer a question about a package |
| [`narrac emit`](#narrac-emit) | print the C code of a debug build |
| [`narrac audit`](#narrac-audit) | list the `unsafe` blocks and `@trusted` functions |
| [`narrac lsp`](#narrac-lsp) | run the language server |
| [`narrac targets`](#narrac-targets) | list the systems `-target` accepts |
| [`narrac --version`](#narrac---version) | print the version |
| [`narrac --help`](#narrac---help) | print the usage |

Running `narrac` with no argument prints the usage on the standard error and exits with code 2. The usage lists the commands above. Other commands exist for working on the compiler itself. They print internal dumps, have no stability promise and are not part of this page.

## narrac build

`narrac build [flags] PATH` builds the executables of a package or a file. For a package, it builds every `[[binary]]`, or the one that `-bin` names. For a file, the executable is named after the file.

| Flag | What it does |
| --- | --- |
| `-release` | build with optimizations |
| `-sanitize` | build with AddressSanitizer; not for `Windows-X86_64` (E1601) |
| `-target NAME` | build for another system; the names are the first column of `narrac targets` |
| `-bin NAME` | build one binary of the package |
| `-o PATH` | name the executable; needs one binary |
| `-emit-c FILE` | also write the C code to `FILE` |
| `-dependents` | on a workspace member, also build the members that depend on it |
| `-dependencies` | on a workspace member, also build the members it depends on |

Exit codes: 0 when every executable was built, 1 for errors in the program or a failed C compiler, 2 for a wrong command line.

## narrac run

`narrac run [flags] PATH [ARGS...]` builds the program in a temporary folder and runs it. The words after `PATH` are the program's arguments. The program's standard output is written when it ends, then its standard error.

| Flag | What it does |
| --- | --- |
| `-release` | build with optimizations |
| `-sanitize` | build with AddressSanitizer |
| `-bin NAME` | run this binary; without it, the package must have exactly one |
| `-stdin FILE` | give the program `FILE` as its standard input; without it, the input is closed |

Exit codes: the program's own exit code, 1 when the build failed, 2 for a wrong command line.

## narrac test

`narrac test [flags] PATH` builds the test blocks of a file or of every module of a package, and runs each in a process of its own. A test that crashes or runs longer than `NARRA_TEST_TIMEOUT` seconds fails.

| Flag | What it does |
| --- | --- |
| `-filter TEXT` | run the tests whose name contains `TEXT` |
| `-time` | print how long each test took |
| `-release` | build with optimizations |
| `-sanitize` | build with AddressSanitizer |
| `-stdin FILE` | give the tests `FILE` as their standard input |
| `-dependents` | on a workspace member, also test the members that depend on it |
| `-dependencies` | on a workspace member, also test the members it depends on |

Exit codes: 0 when every test passed, 1 when a test failed, no test matches `-filter`, or the build failed, 2 for a wrong command line.

## narrac check

`narrac check [flags] PATH` checks a file or package and reports its errors. It prints nothing when there are none. It builds no executable.

| Flag | What it does |
| --- | --- |
| `--json` | print one JSON object per error on the standard output, with `code`, `severity`, `message`, `file`, `range` and `fixes` |
| `--target NAME` | check as if building for that system |
| `--all-targets` | check for every system; not with `--target` |
| `-dependents` | on a workspace member, also check the members that depend on it |
| `-dependencies` | on a workspace member, also check the members it depends on |

The text form is one `file:line:column: message` line for each error on the standard error.

Exit codes: 0 when there are no errors, 1 for errors or an unreadable path, 2 for a wrong command line or an unknown target.

## narrac explain

`narrac explain CODE` prints the message and explanation of a diagnostic code such as `E0003`.

| Flag | What it does |
| --- | --- |
| `--list` | print every code; only with `--json` |
| `--json` | print the list as JSON |

`narrac explain --list --json` prints the whole table of codes. [Errors](errors.md) shows the same table as a page.

Exit codes: 0 when it printed an answer, 1 for an unknown code, 2 for a wrong command line, which includes `--list` without `--json`.

## narrac apply-fix

`narrac apply-fix [flags] PATH` prints the file of the first error that has a fix, with the first fix of that error applied. It does not change the file. For a package it prints the file that the fix changes.

| Flag | What it does |
| --- | --- |
| `--diagnostic N` | pick the error by its place in `narrac check --json`, counted from 0 |
| `--fix M` | pick the fix of that error, counted from 0 |

Exit codes: 0 when it printed a file, 1 for `no diagnostics`, `no such diagnostic`, `no such fix` or an unreadable file, 2 for a wrong command line.

## narrac fmt

`narrac fmt [flags] [PATH...]` rewrites `.nra` files in place in the standard layout. A folder means every `.nra` file inside it, at any depth. A workspace root means the files of every member. With no path it formats the current folder. The layout comes from the `[fmt]` table of `narra.toml` (see [`narra.toml`](narra-toml.md)).

| Flag | What it does |
| --- | --- |
| `--check` | change nothing; print the path of each file that is not formatted |
| `--stdin` | read one file from the standard input and write the result to the standard output; `-` is the same. No paths with it. |

Exit codes: 0 when every file is formatted (or was rewritten), 1 for a file that `--check` found unformatted, a file with a syntax error, a missing file or an unreadable `narra.toml`, 2 for a wrong command line.

## narrac query

`narrac query [flags] QUESTION OPERAND` answers a question about a package from the result of checking it. The questions are `type-of FILE:LINE:COL`, `def NAME`, `refs NAME`, `callers FUNCTION`, `interface MODULE`, `recursion MODULE` and `hovers FILE`. [Chapter 18](../guide/18-tools.md#asking-questions-about-a-package) explains each one.

| Flag | What it does |
| --- | --- |
| `--package DIR` | the package to ask about; the current folder by default |
| `--json` | print the answer as one JSON object |

Exit codes: 0 when there is an answer, an empty one included, 1 for an unknown name, a place with no expression or a package with errors, 2 for a wrong command line.

## narrac emit

`narrac emit [flags] PATH` prints the C code of a debug build on the standard output and builds nothing. For the C of an optimized build, use `narrac build -release -emit-c FILE`.

| Flag | What it does |
| --- | --- |
| `--tests` | emit the test blocks, which a build leaves out, and no entry function |
| `--bin NAME` | emit this binary |

Exit codes: 0 when the C is complete, 1 for errors in the program or an unknown binary, 2 for a wrong command line, 3 when the C holds constructs the compiler cannot emit; each is named on the standard error as `unsupported: ...`.

## narrac audit

`narrac audit [PATH]` lists every `unsafe` block and every `@trusted` function of a package, one line each as `file:line:column: text`, sorted by place. It reads the parsed files and does not check the package. The package is the current folder by default.

Exit codes: 0 on success, 1 when the package cannot be loaded, 2 for a wrong command line.

## narrac lsp

`narrac lsp` is a language server. It speaks the Language Server Protocol on the standard input and output. [Chapter 18](../guide/18-tools.md#using-an-editor) shows how to set up an editor.

| Flag | What it does |
| --- | --- |
| `--full` | check the whole package on every change, instead of the modules a change can affect |
| `--trace` | write a line to the standard error for each check, saying what it checked |

Exit codes: 0 after the editor asks it to shut down, 1 when the input ends before that, 2 for a wrong command line.

## narrac targets

`narrac targets` prints the systems that `-target` and `check --target` accept, one per line, as a name and the triple of the C compiler. It takes no flags. [Chapter 19](../guide/19-building-for-other-systems.md) uses it. Exit code: 0.

## narrac --version

`narrac --version` prints the version, `dev` for a build that is not a release. Exit code: 0.

## narrac --help

`narrac --help` and `narrac -h` print the usage on the standard output. Exit code: 0. `-h` or `--help` after `build`, `run` or `test` prints the usage on the standard error and exits with code 2.

## Environment variables

| Variable | Read by | What it does |
| --- | --- | --- |
| `CC` | `build`, `run`, `test` | the C compiler, one program name or path that is never split into words; `cc` when unset or empty |
| `NARRA_HOST_TARGET` | every command that builds or checks | names the host system as `<os>-<arch>`, such as `Linux-X86_64`, in place of the system `narrac` was built for |
| `NARRA_CACHE` | commands that load packages with dependencies | the folder where downloaded dependencies are kept; without it, `$HOME/.cache/narra` |
| `HOME` | the same | the base of the default cache folder |
| `TMPDIR`, `TEMP`, `TMP` | `run`, `test` | the folder for temporary builds, the first one that is set and not empty, else `/tmp` |
| `PWD` | `fmt`, `lsp` | where relative paths start when looking for `narra.toml` |
| `NARRA_CHECK_LEAKS` | programs built by `narrac`, and `narrac` itself | when set, a program that ends with memory still held reports it and exits with code 102; in `narrac test`, the test that leaked fails |
| `NARRA_CHECK_COPIES` | programs built by `narrac`, and `narrac` itself | when set, a program that ends after copies on write that no `copy(...)` explains reports them and exits with code 103 |
| `NARRA_UPDATE_GOLDEN` | tests | when `1`, `Test.golden` rewrites its expected file instead of comparing |
| `NARRA_TEST_TIMEOUT` | tests | seconds a test may run, from 1 to 99999; 10 when unset or invalid |
| `NO_COLOR`, `FORCE_COLOR`, `TERM` | programs that call `Terminal.useColor` | [chapter 13](../guide/13-standard-library.md) has the rules |

`NARRA_CHECK_LEAKS` and `NARRA_CHECK_COPIES` are read when the program runs, so set them on the command that runs it, such as `NARRA_CHECK_LEAKS=1 narrac run .`. They apply to debug builds.
