# Installing Narra

## Install

On macOS and Linux, run this in a terminal:

```sh
curl -fsSL https://github.com/lxsmnsyc/narra/releases/latest/download/install.sh | sh
```

On Windows, run this in PowerShell:

```powershell
irm https://github.com/lxsmnsyc/narra/releases/latest/download/install.ps1 | iex
```

The command downloads the archive for your system, checks its SHA-256 sum against the sum published with the release, and installs `narrac`. It then tells you the next step. The sum catches a damaged or incomplete download. It does not catch a compromised release: the sum is published next to the archive, so whoever can replace one can replace the other. `narrac` calls a C compiler each time it builds a program, so keep one installed (`cc`, `clang` or `gcc`).

### Where it installs

| System | Folder |
| --- | --- |
| macOS and Linux | `~/.local/bin` |
| Windows | `%LOCALAPPDATA%\narra\bin`, added to your user `PATH` |

Open a new terminal, then check that your shell finds `narrac`:

```sh
narrac targets
```

This lists the systems `narrac` can build for. If your shell says `narrac` is not found on macOS or Linux, `~/.local/bin` is not on your `PATH` yet. Add this line to your shell's startup file, `~/.zshrc` for zsh (the macOS default) or `~/.bashrc` for bash, then open a new terminal:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

### Prebuilt systems

| System | Processor |
| --- | --- |
| macOS | arm64, x64 |
| Linux (glibc) | arm64, x64 |
| Linux (musl, such as Alpine) | arm64, x64 |
| Windows | x64 |

The shell script picks the musl archive on a musl system. On another system, build from source (below).

### Options

| Variable | Meaning | Default |
| --- | --- | --- |
| `NARRA_VERSION` | The version to install, such as `0.1.0`. | The latest release. |
| `NARRA_INSTALL` | The folder that receives `narrac`. | See the table above. |

Set them before `sh` in the shell command, or in the environment before the PowerShell command. For example, this installs version 0.1.0 into `/usr/local/bin`:

```sh
curl -fsSL https://github.com/lxsmnsyc/narra/releases/latest/download/install.sh | NARRA_VERSION=0.1.0 NARRA_INSTALL=/usr/local/bin sh
```

When a release is marked as a prerelease, GitHub's `latest` link skips it. The scripts then look for the newest release of any kind.

`NARRA_BASE_URL` is a testing override. It names a folder that holds the archive and its `.sha256` file instead of the GitHub release, and it must start with `https://`. Only `http://localhost` and `http://127.0.0.1` are allowed without TLS, and the shell script also takes a `file://` folder. The script needs `HOME` or `NARRA_INSTALL` to know where to install.

### Update and uninstall

Run the same command again to update. To uninstall, delete the `narrac` file (`narrac.exe` on Windows). On Windows, also remove the folder from your user `PATH` in the system settings.

## Check that it works

Create a file named `hello.nra`:

```narra
export function run(args: Array<string>): int {
  print("hello, narra")
  return 0
}
```

Run it:

```sh
narrac run hello.nra
```

```text
hello, narra
```

[Chapter 1 of the guide](guide/01-getting-started.md) explains this program and goes on from there.

## Build from source

You can build the compiler yourself instead. The build takes a few minutes.

### What you need

- A C compiler that supports C11: `cc`, `clang` or `gcc`.
- `sh`, the standard shell.
- `git`, to download the source.

Nothing else is needed. The standard library is built into `narrac`, so the single `narrac` file is the whole install. `narrac` also calls the C compiler each time it builds a program, so keep the C compiler installed.

On macOS, `xcode-select --install` installs `cc` and `git`. On Debian or Ubuntu, `sudo apt-get install gcc libc6-dev git` installs them.

### Build

```sh
git clone https://github.com/lxsmnsyc/narra
cd narra
sh selfhost/bootstrap.sh out
```

The last command prints its progress and ends like this:

```text
stage 0: cc builds the seed
stage 1: the seed compiler emits selfhost, cc builds it
stage 2: the stage 1 compiler emits selfhost
stage 1 binary: 5080728 bytes
ok: stage 1 and stage 2 emit identical C (246458 lines)
seed is current
```

The sizes in your output can differ. The new compiler is now in `out/narrac`.

The script builds the compiler in steps:

- **Stage 0.** The C compiler builds `seed/narrac.c`, a C copy of the compiler that is kept in the repository.
- **Stage 1.** That compiler turns the Narra source of the compiler, in `selfhost/`, into C, and the C compiler builds it.
- **Stage 2.** The new compiler turns the same source into C again.

The line that starts with `ok:` is the check that the build is correct. Two compilers, built in different ways, turned the same source into the same C. So the new compiler compiles itself the way the old one did. If the two outputs differ, the script runs one more stage and compares again. It fails only if that comparison also differs.

`seed is current` means `seed/narrac.c` matches the source. When the source is newer, the script prints a note instead. The note is not an error, and the build still works.

#### Choose a C compiler

The script uses `cc` unless you set `CC`. `CC` must name one program, with no flags:

```sh
CC=clang sh selfhost/bootstrap.sh out
```

`narrac` reads `CC` in the same way when it builds your programs.

#### Exit codes

| Code | Meaning |
| --- | --- |
| 0 | The build finished and the check passed. |
| 1 | The check failed: the two compilers turned the source into different C. |
| 2 | Wrong usage, such as more than one folder name. |
| other | A step failed, for example `127` when the C compiler is not found. The line above the error names the step. |

### Install the build

Copy `out/narrac` to a folder on your `PATH`. `~/.local/bin` is a common choice:

```sh
mkdir -p ~/.local/bin
cp out/narrac ~/.local/bin/
```

### Update a source build

From inside the `narra` folder, download the new source, build it, and copy the result over the old file:

```sh
git pull
sh selfhost/bootstrap.sh out
cp out/narrac ~/.local/bin/
```

The repository's tests run on Linux in a container with `sh scripts/linux-check.sh`, which needs Docker.
