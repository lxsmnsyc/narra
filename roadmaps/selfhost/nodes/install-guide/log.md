# Log: install-guide

- 2026-11-15: branch docs/install-guide reset to feat/batch-16 (2d9f2b89). docs/install.md did not exist. README and chapter 1 had no link. Not already fulfilled.
- Fresh clone in /private/tmp/claude-501/install-guide/narra. macOS arm64: `sh selfhost/bootstrap.sh out` exit 0 in 1:19, fixed point, "seed is current". `CC=clang` same. Too many arguments exit 2. `CC=nosuchcc` exits 127, not 1: a failed step passes its own exit code through `set -e`, so the script's "1" comment is only part of the story.
- Install into a scratch HOME (`~/.local/bin`), `narrac run hello.nra` prints `hello, narra`.
- narrac has no `--version` or `help` command. The guide uses `command -v narrac` to check the install.
- Linux: debian under `docker run --platform linux/amd64`, source piped in as a `git archive` tar (colima does not mount /private/tmp).
- Linux x86_64 (debian, emulated): batch branch tar, bootstrap exit 0, fixed point, seed current, install, `narrac run hello.nra` prints `hello, narra`. A second x86_64 run used the guide's literal commands: `apt-get install gcc libc6-dev git`, `git clone https://github.com/lxsmnsyc/narra` (main), bootstrap, install, hello, uninstall. All passed.
- Linux aarch64 (debian, native): batch branch tar, same steps, all passed.
- In dash, `command -v narrac` still printed the path right after `rm`, from the shell's command hash. A new shell does not. The guide does not use `command -v` to check the uninstall.
- Guide suite: PASS, 328 cases.
