# Outcome: install-guide

## Built

- `docs/install.md`, "Installing Narra". It covers prerequisites (a C11 compiler, `sh`, `git`), the build with `sh selfhost/bootstrap.sh out` and its real output, what the stages and the fixed point check mean, `CC=...`, and the exit codes. It also covers the install to `~/.local/bin` with a `PATH` line, a `narrac run hello.nra` check, update, uninstall, a platform table, and a "Prebuilt binaries" header with one line. Node release-binaries fills that last section.
- `README.md`: one line linking to `docs/install.md`.
- `docs/guide/01-getting-started.md`: keeps the three-line build and links to `../install.md` for the `PATH` setup and the build output.

## Deviations

- Platform table: the brief says macOS and Linux on arm64 and x64 are tested. I could run macOS arm64, Linux aarch64 (native container) and Linux x86_64 (emulated container), but not macOS x64. The table lists macOS x64 as "not tested". The orchestrator can change that row if macOS x64 is covered somewhere.
- Prerequisites list `git` as well as `cc` and `sh`, because the install starts with `git clone`.
- Exit codes: `selfhost/bootstrap.sh` says "1 it does not, or a step failed". In practice a failed step passes its own code through `set -e`, for example 127 for a missing `CC`. The guide documents 0, 1, 2 and "other". The script comment is outside this node's paths and is left as is.

## Stubs left

- "Prebuilt binaries" holds one line. release-binaries replaces it with the `curl -fsSL <url>/install.sh | sh` installer.

## Gotchas

- `narrac` has no `--version` or `help` command. The guide checks the install with `command -v narrac`. A `--version` would give release-binaries and the install guide a better check.
- The `bootstrap.sh` stage lines say `cc builds ...` even when `CC=clang`.
- The guide suite does not read `docs/install.md`, so its `narra` block is checked by hand only.
- colima does not mount /private/tmp. I piped a `git archive` tar into `docker run -i` instead of mounting the clone.

## Proposed learnings

- Bootstrap exit codes beyond 0, 1 and 2 come from the failed step | `set -eu` passes the step's code through | selfhost/bootstrap.sh | yes

## Verification

- macOS arm64, fresh clone of feat/batch-16 in /private/tmp/claude-501/install-guide/narra: `sh selfhost/bootstrap.sh out` exit 0 in 1:19, "ok: stage 1 and stage 2 emit identical C (246458 lines)", "seed is current". `CC=clang` gives the same result. Two arguments give exit 2, and `CC=nosuchcc` gives exit 127. Install to `~/.local/bin` in a scratch home, then `narrac run hello.nra` prints `hello, narra` and a `std/int` program prints `3`. `git pull`, `rm ~/.local/bin/narrac` and `git clone https://github.com/lxsmnsyc/narra` all exit 0.
- Linux x86_64, `docker run --platform linux/amd64 debian`, batch branch: bootstrap exit 0, fixed point, seed current. Install and hello pass.
- Linux x86_64, the same image with the guide's literal commands (`apt-get install gcc libc6-dev git`, clone of main from GitHub): build, install, hello and uninstall pass.
- Linux aarch64, `docker run --platform linux/arm64 debian`, batch branch: build, install and hello pass.
- Not run: `xcode-select --install` (the tools are already installed here, and the command opens a dialog). In the containers `apt-get` ran as root without `sudo`.
- Self-review, round 1: two claims made exact ("took 80 seconds on an arm64 Mac", "the install puts no other files on your system"). Round 2: no findings.
- `sh scripts/test.sh --suite guide`: PASS, 328 cases.
- plain-writing `check_prose.py` on docs/install.md, the chapter 1 file and README.md: no findings.

## Size

- Estimated S, actual S.
