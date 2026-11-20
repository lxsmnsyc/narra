# 16. Unsafe code and C

A Narra program cannot read past the end of an array or use memory after it is freed. A read out of range panics, and memory is freed only when nothing uses it (chapter 9). A C library has none of those checks, and neither does a raw pointer. This chapter calls functions from the C libraries every system has. It shows how Narra marks that code, how to keep it small and how to find it again. The running example measures the long side of a right triangle with the C math library, then moves on to C strings and memory.

## Calling a C function

The C math library, `libm`, has `hypot`, which gives the long side of a right triangle from the two short ones. An `extern` block declares the C functions a program uses, and `from` names the library they come from:

```narra error
// error: 9:29: hypot is unsafe; call it inside an `unsafe { }` block
import * as Float from "std/float"

extern {
  function hypot(x: f64, y: f64): f64
} from "m"

export function run(args: Array<string>): int {
  print(Float.toString(hypot(3.0, 4.0)))
  return 0
}
```

Each line in the block is a C function's signature in Narra types, with no body. The compiler cannot see inside the C code, and it cannot check that the signature matches the real function. A wrong signature, or a C function that writes past the end of its memory, can crash the program or corrupt its data. So every call to a C function must sit inside an `unsafe` block, where you state that you checked it yourself:

```narra
import * as Float from "std/float"

extern {
  function hypot(x: f64, y: f64): f64
} from "m"

export function run(args: Array<string>): int {
  unsafe {
    print(Float.toString(hypot(3.0, 4.0)))
  }
  return 0
}
```

```text
5.0
```

`unsafe { ... }` is a statement, not an expression, so it has no value. To get a result out of it, declare a `let` before the block and assign to it inside, as the next section does.

## The `unsafe` effect

Move the call into a function of its own, and the caller needs a block too:

```narra error
// error: 17:32: longSide is unsafe; call it inside an `unsafe { }` block
import * as Float from "std/float"

extern {
  function hypot(x: f64, y: f64): f64
} from "m"

function longSide(a: f64, b: f64): f64 {
  let side = 0.0
  unsafe {
    side = hypot(a, b)
  }
  return side
}

export function run(args: Array<string>): int {
  print(Float.toString(longSide(3.0, 4.0)))
  return 0
}
```

`unsafe` is an effect, like `alloc` and `io` in [chapter 10](10-effects-and-testing.md). A function that holds an `unsafe` block has it, and so does every function that calls one. The difference is at the call. A call to a function with `alloc` needs nothing around it. A call to a function with `unsafe` needs an `unsafe` block, even in `run`. So the unchecked code is marked wherever it can be reached from.

As with the other effects, a private function gets `unsafe` without writing it, and an exported one must list it:

```narra error
// error: 9:17: Function `longSide` uses `unsafe` (through `hypot`) but its signature does not declare it; write `uses unsafe`
extern {
  function hypot(x: f64, y: f64): f64
} from "m"

export function longSide(a: f64, b: f64): f64 {
  let side = 0.0
  unsafe {
    side = hypot(a, b)
  }
  return side
}

export function run(args: Array<string>): int {
  return 0
}
```

Writing `uses unsafe` there compiles, and passes the job of checking on to every caller. A function value of a type with `uses unsafe`, such as `(x: f64) => f64 uses unsafe`, needs a block at each call too. So does a `for` loop over a generator whose body has the effect, since the loop runs the body.

## Stopping the effect with `@trusted`

`hypot` with two plain numbers cannot go wrong: it reads no memory, and the signature matches the C one. A caller should not have to say `unsafe` for it. Write `@trusted` on its own line before the function. It states that the `unsafe` blocks inside were checked by hand, and the effect stops there:

```narra
import * as Float from "std/float"

extern {
  function hypot(x: f64, y: f64): f64
} from "m"

@trusted
export function longSide(a: f64, b: f64): f64 {
  let side = 0.0
  unsafe {
    side = hypot(a, b)
  }
  return side
}

export function run(args: Array<string>): int {
  print(Float.toString(longSide(3.0, 4.0)))
  print(Float.toString(longSide(5.0, 12.0)))
  return 0
}
```

```text
5.0
13.0
```

`longSide` has no `unsafe` effect, so its signature lists none, and `run` calls it like any other function. Keep `@trusted` functions small, so that the code someone has to check by hand is short. `@trusted` on a function that has no `unsafe` block is an error, and so is `@trusted` together with `uses unsafe`:

```narra error
// error: 2:1: @trusted on double, which has no `unsafe` block; remove @trusted
@trusted
function double(value: int): int {
  return value * 2
}

export function run(args: Array<string>): int {
  return double(0)
}
```

A call through a trait (chapter 7) sees only the trait's signature. So a function in an `implement` block may use `unsafe` only when the trait function declares `uses unsafe`. Otherwise the implementation wraps its unsafe code with `@trusted`, the one attribute an `implement` block member takes.

## C types

An `extern` function takes and returns numbers, `bool` and pointers (`Ptr<T>`, below), or returns nothing. C uses integers of several sizes, and so does Narra. [Chapter 2](02-values-and-types.md#numbers-of-other-sizes) teaches the sized integers `i8` to `u64` and their modules, such as `std/i32`.

Each Narra type stands for one C type:

| Narra | C |
| --- | --- |
| `int`, `i64` | `int64_t`, and `long` on the supported systems |
| `i32` | `int` |
| `i8`, `i16`, `u8` to `u64` | `int8_t`, `int16_t`, `uint8_t` to `uint64_t`; `size_t` is `u64` |
| `f64`, `f32` | `double`, `float` |
| `bool` | `bool` |
| `Ptr<T>` | `T*`; `char*`, `const char*` and `void*` are `Ptr<u8>` |

Narra checks none of this against the C header. Read the C signature and pick the type from the table. Some C functions, such as `isalpha`, return an `int` that means yes or no. Declare them with `i32`, not `bool`, because some C libraries return values such as 1024 for yes. A string, array or any other Narra value with a reference count cannot cross to C:

```narra error
// error: 3:17: text of C function puts has type string; a C function takes numbers, bool and Ptr<T>
extern {
  function puts(text: string): i32
} from "c"

export function run(args: Array<string>): int {
  return 0
}
```

`from "c"` is the C library itself, `libc`, which every program links. The name in `from` is the one `cc` takes after `-l`: `"m"` is `libm`, `"z"` is zlib.

Each function links to the C function of its own name. When that name is taken in your module, or reads badly, give the C name with `@extern` on the line before it:

```narra
import * as I32 from "std/i32"

extern {
  @extern("abs")
  function absolute(value: i32): i32
} from "c"

export function run(args: Array<string>): int {
  unsafe {
    print(I32.toString(absolute(-5)))
  }
  return 0
}
```

```text
5
```

C functions that take a variable number of arguments, such as `printf`, cannot be declared. There is no way to write C's `...`. A fixed signature for one reads garbage on some systems, because C passes those arguments in a different way. Use a C function with a fixed signature, or the Narra functions that do the same job.

## C strings

A C function that takes text wants a pointer to bytes that end with a zero byte. A Narra string has no zero at its end, and it may hold a zero in the middle, which C would read as the end. `std/cstring` has `CString`, a string that holds no zero byte. The examples import it as `import * as CStrings from "std/cstring"`:

- `CStrings.from(text)` returns `Some` with a `CString`, or `None` when `text` holds a zero byte.
- `CStrings.pointer(value)` is the `Ptr<u8>` to pass for `char*` or `const char*`. It is valid while the `CString` is, and C must not write through it.
- `CStrings.toString(value)` gives the text back.
- `CStrings.read(pointer)` copies a zero-terminated string that a C function returned into a Narra string. It reads memory through a pointer, so it is unsafe, and it returns `Err` when the bytes are not valid UTF-8.

The C library's `strtol` reads a number from text in any base, and `strerror` returns the message for an error code:

```narra
import * as CStrings from "std/cstring"
import * as Int from "std/int"
import * as Ptr from "std/ptr"

extern {
  function strtol(text: Ptr<u8>, end: Ptr<Ptr<u8>>, base: i32): int
  function strerror(code: i32): Ptr<u8>
} from "c"

@trusted
function parseHex(text: string): Option<int> {
  if (CStrings.from(text) is Some(input)) {
    let value = 0
    unsafe {
      value = strtol(CStrings.pointer(input), Ptr.null(), 16)
    }
    return Some(value)
  }
  return None
}

@trusted
function describeError(code: i32): string uses alloc {
  let message = ""
  unsafe {
    if (CStrings.read(strerror(code)) is Ok(text)) {
      message = text
    }
  }
  return message
}

export function run(args: Array<string>): int {
  if (parseHex("ff") is Some(value)) {
    print(Int.toString(value))
  }
  if (parseHex("a\0b") is None) {
    print("no CString holds a zero byte")
  }
  print(describeError(2))
  return 0
}
```

```text
255
no CString holds a zero byte
No such file or directory
```

`strtol` can store where the number ended through its second parameter, a `char**`. The program does not need that, and C takes a null pointer to mean "do not store it", so it passes `Ptr.null()`.

## Pointers

`Ptr<T>` is a raw pointer: the address of a `T` in memory, as C's `T*`. It has no reference count, and nothing checks what it points at. Using `Ptr` needs `import * as Ptr from "std/ptr"`, and the type is written `Ptr<T>` under any alias. `std/ptr` has these functions:

| Function | Does | Unsafe |
| --- | --- | --- |
| `Ptr.alloc(count)` | allocates `count` elements set to zero | yes |
| `Ptr.free(pointer)` | frees what `Ptr.alloc` returned | yes |
| `Ptr.read(pointer)` | reads the element it points at | yes |
| `Ptr.write(pointer, value)` | writes the element it points at | yes |
| `Ptr.offset(pointer, count)` | moves `count` elements on | yes |
| `Ptr.cast(pointer)` | reads the address as a pointer to another type | yes |
| `Ptr.null()` | the null pointer | no |
| `Ptr.isNull(pointer)` | whether it is null | no |

`Ptr.alloc`, `Ptr.null` and `Ptr.cast` take the element type from the type the result must have, so give the `const` a type:

```narra
import * as Int from "std/int"
import * as Ptr from "std/ptr"

@trusted
function sumSquares(count: int): int {
  let total = 0
  unsafe {
    const items: Ptr<int> = Ptr.alloc(count)
    for (let i = 0; i < count; i++) {
      Ptr.write(Ptr.offset(items, i), i * i)
    }
    for (let i = 0; i < count; i++) {
      total += Ptr.read(Ptr.offset(items, i))
    }
    Ptr.free(items)
  }
  return total
}

export function run(args: Array<string>): int {
  print(Int.toString(sumSquares(4)))
  const nothing: Ptr<int> = Ptr.null()
  if (Ptr.isNull(nothing)) {
    print("null")
  }
  return 0
}
```

```text
14
null
```

Nothing is checked when the program runs. Each of these is undefined behavior, so the program may crash or carry on with wrong data:

- a read or write past the end of the memory
- a read or write after `Ptr.free`
- a second `Ptr.free` of the same pointer

Memory from `Ptr.alloc` that is never freed shows up in the leak check of chapter 9.

A pointer is a plain value. You can store it in a struct or an array, copy it and pass it to any function. Only reading or writing through it is unsafe. It cannot be compared with `==` or printed, so test it with `Ptr.isNull`. It may point at numbers, `bool`, `char`, other pointers, and structs whose fields are all of those. A pointer to a string, array or anything else with a reference count is an error, since writing through it would get around the count:

```narra error
// error: 5:17: Ptr<string> is not allowed: string is reference counted or has no fixed layout
import * as Ptr from "std/ptr"

export function run(args: Array<string>): int {
  const _names: Ptr<string> = Ptr.null()
  return 0
}
```

### Memory that C allocates

Memory a C function allocates goes back through the C library's own `free`, declared in an `extern` block like any other function. `Ptr.free` frees only what `Ptr.alloc` returned. Here `strdup` copies a string into memory it allocates. The program writes a new first letter into the copy before reading it back:

```narra
import * as CStrings from "std/cstring"
import * as Ptr from "std/ptr"

extern {
  function strdup(text: Ptr<u8>): Ptr<u8>
  function free(memory: Ptr<u8>)
} from "c"

@trusted
function withJ(text: string): string uses alloc {
  let result = ""
  if (CStrings.from(text) is Some(input)) {
    unsafe {
      const copied = strdup(CStrings.pointer(input))
      if (!Ptr.isNull(copied)) {
        // 74 is the byte of the letter J.
        Ptr.write(copied, 74)
        if (CStrings.read(copied) is Ok(changed)) {
          result = changed
        }
        free(copied)
      }
    }
  }
  return result
}

export function run(args: Array<string>): int {
  print(withJ("hello"))
  return 0
}
```

```text
Jello
```

The program writes through `copied`, never through `CStrings.pointer(input)`, because the bytes of a `CString` must not change.

## Linking a library

`narrac build` and `narrac run` pass `-l<library>` to the C compiler for each `extern` block whose functions the program calls. A block whose functions nothing calls links nothing. `libm` is always linked, and `libc` needs no flag.

The C compiler looks for libraries in a few system folders. A library kept elsewhere, for example under Homebrew's `/opt/homebrew/lib` or in the package, needs its folder in `narra.toml`. `[link] search` lists folders, and each becomes a `-L` flag. A relative folder is relative to the package. Here is the triangle as a package, with `longSide` in a module of its own:

```narra file triangle narra.toml
[package]
name = "triangle"

[[binary]]
name = "triangle"
entry = "main.run"

[link]
search = ["vendor/lib"]
```

```narra file triangle src/geometry.nra
extern {
  function hypot(x: f64, y: f64): f64
} from "m"

@trusted
export function longSide(a: f64, b: f64): f64 {
  let side = 0.0
  unsafe {
    side = hypot(a, b)
  }
  return side
}
```

```narra file triangle src/main.nra
import * as Float from "std/float"
import * as Geometry from "triangle/geometry"

export function run(args: Array<string>): int {
  print(Float.toString(Geometry.longSide(5.0, 12.0)))
  return 0
}
```

```text
13.0
```

`libm` needs no search folder. The line shows where a library of your own would go: with `from "shapes"` and `vendor/lib/libshapes.a` in the package, the build finds it. `search` is the only key `[link]` takes.

### A library for one system

Some libraries exist on one operating system only. `@target` before an `extern` block declares its functions, and links its library, only when building for that system. Every call must then sit inside a branch that holds only there, as [chapter 19](19-building-for-other-systems.md) explains for `@target` functions:

```narra
import * as U64 from "std/u64"
import * as Build from "std/build"
import { Os } from "std/build"

@target(os = Os.Windows)
extern {
  function GetTickCount64(): u64
} from "kernel32"

export function run(args: Array<string>): int {
  if (Build.os == Os.Windows) {
    unsafe {
      print(`up for ${U64.toString(GetTickCount64())} ms`)
    }
  } else {
    print("not Windows, so kernel32 is not linked")
  }
  return 0
}
```

```text
not Windows, so kernel32 is not linked
```

On macOS and Linux the Windows branch is dropped, so the build does not look for `kernel32`.

## Finding the unsafe code

`narrac audit` lists every `unsafe` block and every `@trusted` function in a package, one line each, sorted by file and position. Run it on the triangle package:

```sh
narrac audit triangle
```

```text
src/geometry.nra:5:1: @trusted function longSide
src/geometry.nra:8:3: unsafe block in function longSide
```

The list is the code to review by hand when a program crashes, or before you depend on a package. It reads the source only and always exits with code 0.

## Not built yet

C cannot call Narra functions yet, so a C function that takes a callback, such as `qsort`, cannot be used. There is also no way yet to give a Narra struct the field order and padding of a C struct. So a C function that takes a struct cannot be declared. Both are planned.

Next: [17. Proofs and compile-time constants](17-proofs-and-compile-time.md).
