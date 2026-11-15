# Installing Narra

You install Narra by building its compiler, `narrac`, from source and copying the one file it produces to a folder on your `PATH`. The build takes about a minute on a recent Mac.

## What you need

- A C compiler that supports C11: `cc`, `clang` or `gcc`.
- `sh`, the standard shell.
- `git`, to download the source.

Nothing else is needed. The standard library is built into `narrac`, so the single `narrac` file is the whole install. `narrac` also calls the C compiler each time it builds a program, so keep the C compiler installed.

On macOS, `xcode-select --install` installs `cc` and `git`. On Debian or Ubuntu, `sudo apt-get install gcc libc6-dev git` installs them.

## Build

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

### Choose a C compiler

The script uses `cc` unless you set `CC`. `CC` must name one program, with no flags:

```sh
CC=clang sh selfhost/bootstrap.sh out
```

`narrac` reads `CC` in the same way when it builds your programs.

### Exit codes

| Code | Meaning |
| --- | --- |
| 0 | The build finished and the check passed. |
| 1 | The check failed: the two compilers turned the source into different C. |
| 2 | Wrong usage, such as more than one folder name. |
| other | A step failed, for example `127` when the C compiler is not found. The line above the error names the step. |

## Install

Copy `out/narrac` to a folder on your `PATH`. `~/.local/bin` is a common choice:

```sh
mkdir -p ~/.local/bin
cp out/narrac ~/.local/bin/
```

Check that your shell finds it:

```sh
command -v narrac
```

This prints the path of `narrac`, such as `/Users/you/.local/bin/narrac`. If it prints nothing, `~/.local/bin` is not on your `PATH` yet. Add this line to your shell's startup file, `~/.zshrc` for zsh (the macOS default) or `~/.bashrc` for bash, then open a new terminal:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

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

## Update

From inside the `narra` folder, download the new source, build it, and copy the result over the old file:

```sh
git pull
sh selfhost/bootstrap.sh out
cp out/narrac ~/.local/bin/
```

## Uninstall

Delete the file:

```sh
rm ~/.local/bin/narrac
```

You can also delete the `narra` folder you cloned. The install puts no other files on your system.

## Platforms

| System | Processor | Status |
| --- | --- | --- |
| macOS | arm64 | tested |
| macOS | x64 | not tested |
| Linux | arm64, x64 | tested |
| Windows | any | not supported yet |
| other systems | any | not supported yet |

The repository's tests run on Linux in a container with `sh scripts/linux-check.sh`, which needs Docker.

## Prebuilt binaries

Prebuilt binaries are not published yet. Build from source as shown above.
