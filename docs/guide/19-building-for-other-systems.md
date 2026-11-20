# 19. Building for other systems

A Narra program is compiled to C, and the C compiler builds it for one system: an operating system and a processor. By default that is the system you work on. This chapter shows how to build for another one, such as Linux from a Mac, or a WebAssembly module. It also shows how to write code that differs between systems. It also covers what the debug and release builds do, and what the compiler's choices cost at run time. The running example is a small tool that tells the user how to list the files in a folder. The command has a different name on Windows.

## Code that depends on the system

The module `std/build` holds constants that the compiler sets for each build:

- `Build.os` is the operating system, a value of the enum `Os`: `Os.MacOS`, `Os.Linux`, `Os.Windows` or `Os.Wasi`.
- `Build.arch` is the processor, a value of the enum `Arch`: `Arch.Arm64`, `Arch.X86_64` or `Arch.Wasm32`.
- `Build.debug` is `true` in a debug build and `false` with `-release`.

Both enums are exported from `std/build`, so they are imported by name, as `Os` is here:

```narra
import * as Build from "std/build"
import { Os } from "std/build"

function listCommand(): string {
  if (Build.os == Os.Windows) {
    return "dir"
  } else {
    return "ls"
  }
}

export function run(args: Array<string>): int {
  print(`list files with ${listCommand()}`)
  return 0
}
```

```text
list files with ls
```

`Build.debug` works the same way. The guide's examples run in both builds, so this one is a fragment:

```narra fragment
if (Build.debug) {
  print("debug build")
} else {
  print("release build")
}
```

With `narrac run` it prints `debug build`, and with `narrac run -release` it prints `release build`.

An `if` whose condition uses only these constants is decided while compiling. The C holds the branch that matches and nothing of the other, with no test at run time. The condition can combine the constants with `==`, `!=`, `is`, `!`, `&&` and `||`, such as `Build.os == Os.Linux && Build.arch == Arch.Arm64`, and every `else if` is decided on its own. A condition that mixes them with an ordinary value, or a `match` on them, is a plain test that runs every time.

The dropped branch is still checked. A Windows branch must type-check when you build on a Mac, so code for another system cannot break without anyone noticing. Its effects count too, so a function has the same effects on every system.

### Asking the machine

`Build.hostTarget()` is a function, not a constant. It returns an `Option<Target>`, where `Target` has the fields `os` and `arch`. It reads the C compiler's predefined macros. It gives `None` when no variant of `Os` and `Arch` fits:

```narra
import * as Build from "std/build"

export function run(args: Array<string>): int {
  if (Build.hostTarget() is Some(_)) {
    print("hostTarget names a system")
  } else {
    print("hostTarget has no answer")
  }
  return 0
}
```

```text
hostTarget names a system
```

Use `Build.os` and `Build.arch` to choose code. Their branches are dropped while compiling, and `hostTarget()` leaves everything in the program.

## Functions that exist on one system

`@target` on the line before a top-level function makes the function exist only for that system. Here `listCommand` is the Windows one:

```narra file listing narra.toml
[package]
name = "listing"

[[binary]]
name = "listing"
entry = "main.run"
```

```narra file listing src/main.nra
import * as Build from "std/build"
import { Os } from "std/build"

@target(os = Os.Windows)
function listCommand(): string {
  return "dir"
}

function describe(): string {
  if (Build.os == Os.Windows) {
    return `list files with ${listCommand()}`
  } else {
    return "list files with ls"
  }
}

export function run(args: Array<string>): int {
  print(describe())
  return 0
}
```

```text
list files with ls
```

`@target` takes `os =` a variant of `Os`, `arch =` a variant of `Arch`, or both, each once: `@target(os = Os.Linux, arch = Arch.Arm64)`. A part it leaves out matches every value. The function is type-checked on every system, like a dropped branch. A build for another system leaves it out of the C when nothing that system compiles calls it.

The compiler checks each use. Where the function does not exist, a call, or the name used as a value, is an error. A use is accepted in two places. One is a branch that holds only where the function exists, as `listCommand()` does above. The other is a `@target` function with the same or a narrower target. An ordinary `if`, a `match` on `Build.os`, an early `return` and a test of `Build.debug` guarantee nothing. A file that calls `listCommand()` with no condition around the call fails:

```narra fragment
import { Os } from "std/build"

@target(os = Os.Windows)
function listCommand(): string {
  return "dir"
}

export function run(args: Array<string>): int {
  print(listCommand())
  return 0
}
```

I saved the file as `list.nra` and checked it on a Mac:

```sh
narrac check list.nra
```

```text
list.nra:9:20: listCommand exists only on @target(os = Os.Windows), but this use is compiled for MacOS-Arm64; use it inside an if on Build.os or Build.arch that keeps it only on that target
```

On Linux the message names the Linux row instead. To see the mistake from one machine, check every system:

```sh
narrac check --all-targets list.nra
```

```text
list.nra:9:20: listCommand exists only on @target(os = Os.Windows), but this use is compiled for MacOS-Arm64, MacOS-X86_64, Linux-X86_64, Linux-Arm64, Linux-X86_64-musl, Linux-Arm64-musl, Wasi-Wasm32; use it inside an if on Build.os or Build.arch that keeps it only on that target
```

Plain `narrac check` checks a debug build for your own system only, so a use that is wrong only on another system shows only with `--all-targets`. The option checks every system of `narrac targets`, each as a debug and as a release build. A test of `Build.debug` can keep a use in one of them only. Each error is reported once. The names after "compiled for" are the systems where it fails, and `(debug)` or `(release)` follows a name that fails in one build only. A program that reads `Build` is the one that needs this check before a release.

`@target` also goes before an `extern` block, to link a C library on one system only. [Chapter 16](16-unsafe-code-and-c.md#a-library-for-one-system) shows it. Functions and `extern` blocks take `@target`. Other declarations do not.

### Looking at what the compiler kept

`narrac build -emit-c FILE` writes the C ([chapter 18](18-tools.md#building-running-and-looking-at-the-c) shows the option), and `--target` below picks the system it is for. I built `src/main.nra` for Windows and for the Mac I work on, and searched each C file for the text of `listCommand`:

```sh
narrac build --target Windows-X86_64 -emit-c windows.c -o listing.exe .
narrac emit . > mac.c
grep -c '"dir"' windows.c mac.c
```

```text
mac.c:0
windows.c:1
```

The Mac build holds no trace of `dir`. Its `if` kept the `ls` branch, and the build left out the function that nothing reaches. `narrac emit` always writes the C of a debug build for your own system. `NARRA_HOST_TARGET=Windows-X86_64 narrac emit .` writes the C that a Windows build would hold for `std/build`, as far as the conditions go.

## Building for another system

`narrac targets` lists the systems that the compiler can build for ([chapter 18](18-tools.md#building-running-and-looking-at-the-c) shows the list). Its first column is the name for `--target`, and `check --target NAME` checks as that system:

```sh
narrac build --target Linux-Arm64-musl -o listing .
```

That is all Narra needs, but the C compiler has to build for the other system too. `narrac build` runs the compiler named by the environment variable `CC` (`cc` when it is not set) and gives it the target triple of the row. With your own system's C compiler, a build for another system stops at the first missing header. `zig cc` is a C compiler that builds for all eight rows from one machine. `CC` names one program, so a `CC="zig cc"` is not enough. Make a script that calls it:

```sh
#!/bin/sh
exec zig cc "$@"
```

I saved it as `zigcc` and ran these builds on an arm64 Mac, and `file` says what each made:

```sh
export CC=$PWD/zigcc
narrac build --target Windows-X86_64 -o listing.exe .
file listing.exe
narrac build --target Linux-Arm64-musl -o listing .
file listing
```

```text
listing.exe: PE32+ executable (console) x86-64, for MS Windows
listing: ELF 64-bit LSB executable, ARM aarch64, version 1 (SYSV), statically linked, with debug_info, not stripped
```

The result is for the other system. Copy it there and run it, or run it where a runner exists. The names that end in `-musl` link the C library statically, so one file runs on any Linux of that processor. The names without it link the `glibc` that most distributions have. `Build.os` and `Build.arch` do not tell the two apart.

Some commands reject a target. A name that is not in the list is an error that names the rows. `build -sanitize --target ...` is an error, because the sanitizer's runtime cannot be linked across systems. `narrac test` takes no `--target`, so run the tests on the system you build on. Each of these exits with code 2.

The repository's `scripts/cross-check.sh` builds the sample programs for every row this way and runs each wherever a runner exists. It is how the compiler's own authors test the rows.

## WebAssembly with WASI

`Wasi-Wasm32` builds a WebAssembly module for a WASI host. The host gives the module files, arguments and the standard streams, as an operating system does. Build it with `zig cc` as above:

```sh
narrac build --target Wasi-Wasm32 -o listing.wasm .
wasmtime run -W exceptions=y listing.wasm
```

```text
list files with ls
```

I ran this with `wasmtime`. The flag turns on WebAssembly exceptions, which the runtime's panic handler needs. The run prints `ls` and not `dir`, because `Build.os` is `Os.Wasi` there and the condition tested for Windows only.

A WASI host cannot do everything that an operating system does. Some standard library items have no backing there. Using one is an error on this target, as a `@target` function is where it does not exist. They are `Process.run`, every function of `std/thread`, `std/parallel`, `std/channel`, `std/shared` and `std/atomic`, the functions of `std/async/net`, and `Terminal.columns`. Their signatures stay the same, so the same source checks on every system. A program that uses `Process.run` fails the check for this target only:

```narra fragment
import * as Process from "std/process"

export function run(args: Array<string>): int uses io, alloc {
  match (Process.run("echo", ["hello"])) {
    Ok(#[_, output, _]) => print(output)
    Err(message) => print(message)
  }
  return 0
}
```

```sh
narrac check --target Wasi-Wasm32 echo.nra
```

```text
echo.nra:4:21: Process.run is not available on Wasi, but this use is compiled for Wasi-Wasm32; use it inside an if on Build.os that drops Os.Wasi
```

The fix is the one for a `@target` function: put the call in a branch that drops `Os.Wasi`, such as `if (Build.os != Os.Wasi)`. The timers of `std/async` work, and the loop sleeps until the nearest one, because it cannot wait on descriptors.

## Windows

A program built for `Windows-X86_64` runs the standard library like the others, with these differences:

- `Path` also reads `\` as a separator there, and a drive such as `C:` stays in front of the result. `Path.join("C:\\a", "..\\b")` is `C:/b` on Windows. Everywhere else `\` is an ordinary character in a name, and the same call gives `C:\a/..\b`. `Path.join` always writes `/`, which the Windows file functions accept. Paths of the form `\\server\share` are not supported.
- Files are opened in binary mode, as are the standard streams, so `print` writes `\n`, never `\r\n`.
- `Process.run` does not start a shell, so `echo` and `dir` need `cmd /c`. An exit code after a crash is a number such as `3221225477`, where other systems report the signal.
- `narrac test` runs each test in a binary of its own, and a crash is reported as `crashed with status 0x...`.
- `std/async` waits on sockets only, with `WSAPoll`.
- `Env.args()` reads arguments as UTF-8, and the `args` that `run` receives use the system's code page. Use `Env.args()` for a name that is not ASCII ([chapter 11](11-files-and-the-system.md)).

I could not run a Windows program on the machine that wrote this chapter, so the `listing.exe` above is only checked with `file`. The repository's Windows check runs the rows on a Windows runner before a pull request merges.

## Debug and release builds

`narrac build` and `narrac run` make a debug build, and `-release` makes a release build. They differ in three ways:

- A debug build turns on checks. AddressSanitizer stops at the first bad memory access. A leak check runs at exit. A copy check finds a value shared by two names that should have been copied. A release build has none of them.
- A release build is compiled with `-O2`, and the C compiler is clang or GCC.
- `Build.debug` is `true` in the debug build and `false` in the release build, so a branch on it changes what the program holds.

The debug build is for working and for tests, and the release build is the one to ship. A correct program gives the same results in both. For the module above, built for WASI by hand, the sizes are:

```sh
narrac build --target Wasi-Wasm32 -o debug.wasm .
narrac build --target Wasi-Wasm32 -release -o release.wasm .
ls -lh debug.wasm release.wasm | awk '{print $5, $9}'
```

```text
1.5M debug.wasm
954K release.wasm
```

Both builds fix how floats round. The C starts with `#pragma STDC FP_CONTRACT OFF`, so `a * b + c` rounds twice at every optimization level and gives the same digits on every system.

### What the code costs

The compiler's choices are few, and each is one you can predict from the source:

- **A generic function is copied for each set of types it is used with.** There is no boxing: an `Array<int>` holds plain integers, not references to them. The copy is as fast as a function written for the type, and the program grows with the number of sets.
- **A `mutable` parameter is a pointer.** The function writes the caller's value, and nothing is copied (chapter 9).
- **A parameter larger than two machine words is passed as a pointer too.** Parameters are constant, so the program cannot tell.
- **A closure that does not escape keeps its captures on the stack. One that escapes, by being returned or stored, keeps them in a counted block on the heap** (chapter 3).
- **Reading a field or an element through a name copies nothing.** `items[i].field` reads in place.

You can see the first two in the C. This program has a generic function, used for two types, and a `mutable` parameter:

```narra
import * as Arrays from "std/array"
import * as Int from "std/int"

struct Counter {
  value: int
}

function countOf<T>(items: Array<T>): int {
  return Arrays.length(items)
}

function bump(mutable counter: Counter) {
  counter.value = counter.value + 1
}

export function run(args: Array<string>): int {
  const counter = Counter { value: 4 }
  bump(mutable counter)
  print(Int.toString(counter.value))
  print(Int.toString(countOf([1, 2, 3])))
  print(Int.toString(countOf(["a", "b"])))
  return 0
}
```

```text
5
3
2
```

I saved it as `cost.nra` and searched the C for the two functions:

```sh
narrac emit cost.nra | grep -E '^static (void|int64_t) f_[0-9a-f]+_(bump|countOf).*[{]$'
```

```text
static void f_128b81d4_bump(S_128b81d4_Counter* v_counter_19) {
static int64_t f_128b81d4_countOf_8025fb22(A_Array_e7fe1e37* v_items_22) {
static int64_t f_128b81d4_countOf_d3887ce9(A_Array_5fce1d7c* v_items_23) {
```

The numbers in the names differ from build to build. `bump` takes a `Counter*`, a pointer, and `countOf` appears twice: one copy for `Array<int>` and one for `Array<string>`.

## Vectors with std/simd

A processor can add four numbers in one instruction. `std/simd` has types for that. A vector holds a fixed number of numbers of one type, and an operator works on all of them at once. `F32x4` is four `f32` values. The twenty types come in two widths, 128 bits, such as `I32x4`, `U8x16`, `F32x4` and `F64x2`, and 256 bits, such as `I32x8` and `F64x4`.

```narra
import * as Simd from "std/simd"

export function run(args: Array<string>): int {
  const prices: F32x4 = [10.0, 20.0, 30.0, 40.0]
  const discount: F32x4 = Simd.splat(0.5)
  const halved = prices * discount
  print(Show.show(halved))
  print(Show.show(Simd.sum(halved)))
  print(Show.show(halved[2]))
  return 0
}
```

```text
[5.0, 10.0, 15.0, 20.0]
50.0
15.0
```

An array literal with a vector type before it gives one number for each lane. `Simd.splat` fills every lane with one number, and a vector and a plain number cannot meet in one operator, so `prices * 0.5` is an error. `Simd.sum` adds the lanes. `halved[2]` reads one lane, and the compiler must prove the index in range, as for a fixed-size array. Integer lanes wrap around on overflow, like integers do. `Simd.load` and `Simd.store` copy lanes from and to an `Array`, and `Simd.select`, `Simd.shuffle` and the comparisons `Simd.less` and `Simd.equal` cover masks and reordering.

The C compiler turns the types into its vector extensions. The build passes no flag for a particular processor, so x86-64 code uses SSE2 and arm64 code uses NEON. The same source builds for every row above. A float sum is the same on all of them. Most programs never need vectors. They are for loops over large arrays of numbers.

Next: [20. Narra for TypeScript programmers](20-narra-for-typescript-programmers.md).
