# 11. Files and the system

Until now the programs in this guide have printed and finished. A real program also reads files, takes its input from the command line, and tells the system whether it worked. This chapter builds `lines`, a small tool that counts the lines in the files named on its command line. When no file is named, it counts its standard input. On the way it covers `std/fs`, `std/path`, `std/env`, `std/io`, `std/process`, `std/bytes` and `std/time`.

Every function in this chapter that touches the outside world has the `io` effect from chapter 10, and most of them also have `alloc`. An exported function of yours that calls one has to declare both: `uses io, alloc`. Private functions and `run` may leave the list out. The samples write it there too, so that each signature shows what the function does.

## Read a file

`Fs.readText` reads a whole file as a string. `Fs.writeText` replaces a file with a string, creating it when it does not exist, and gives back the number of bytes it wrote. Both return a `Result`, since a file may be missing or locked, and the error is a string that names the file and the reason. Import `std/fs` as `Fs`.

The first piece of `lines` reads a file and counts its lines. The sample writes the file it counts, so that it runs the same on every machine:

```narra
import * as Arrays from "std/array"
import * as Env from "std/env"
import * as Fs from "std/fs"
import * as Int from "std/int"
import * as Options from "std/option"
import * as Path from "std/path"
import * as Str from "std/string"
import * as Time from "std/time"

function countLines(path: string): Result<int, string> uses io, alloc {
  const read = Fs.readText(path)
  if (read is Err(message)) {
    return Err(message)
  }
  const lines: Array<string> = Options.withDefault(Str.split(read.value, "\n"), [])
  let count = Arrays.length(lines)
  // A file that ends with a newline has an empty piece after the last line.
  if (Str.endsWith(read.value, "\n") || read.value == "") {
    count--
  }
  return Ok(count)
}

function scratchFolder(name: string): string uses io, alloc {
  const temporary = Options.withDefault(Env.get("TMPDIR"), "/tmp")
  return Path.join(temporary, `narra-guide-${name}-${Int.toString(Time.unixMillis())}`)
}

export function run(args: Array<string>): int uses io, alloc {
  const folder = scratchFolder("read")
  const _created = Fs.createDirAll(folder)
  const path = Path.join(folder, "notes.txt")

  const written = Fs.writeText(path, "one\ntwo\nthree\n")
  if (written is Ok(size)) {
    print(`wrote ${Int.toString(size)} bytes`)
  }

  match (countLines(path)) {
    Ok(count) => print(`${Int.toString(count)} lines`)
    Err(message) => print(`cannot read: ${message}`)
  }
  match (countLines(Path.join(folder, "missing.txt"))) {
    Ok(count) => print(`${Int.toString(count)} lines`)
    Err(message) => print(`cannot read: ${Str.replace(message, folder, "<folder>")}`)
  }

  const _removed = Fs.removeFile(path)
  const _gone = Fs.removeDir(folder)
  return 0
}
```

```text
wrote 14 bytes
3 lines
cannot read: <folder>/missing.txt: No such file or directory
```

`countLines` passes the error up as it is. `Str.split` returns an `Option` because the separator could be empty, which is why `Options.withDefault` is there. The last piece of a text that ends with `\n` is empty, so the count drops by one.

The rest of the sample is setup, and later samples repeat it. `scratchFolder` builds a folder name inside the system's temporary folder, which `Env.get("TMPDIR")` reads (see "The command line and the environment" below). The name ends with the time in milliseconds, so two runs never share a folder. `Fs.createDirAll` makes it, and the last two lines remove what the sample made. A sample that leaves files behind would not give the same output twice.

`Fs.writeText` replaces the whole file. There is no function that appends: read the text, join the new text to it and write it back. The read and write functions treat the text as UTF-8; "Bytes" below reads files that are not.

## Build paths

A path is a string, and `std/path` takes it apart and puts it together without touching the disk. Import it as `Path`:

```narra
import * as Path from "std/path"

export function run(args: Array<string>): int uses alloc {
  print(Path.join("notes", "today.txt"))
  print(Path.join("/home/ada", "../grace/./a.txt"))

  const file = "/home/ada/archive.tar.gz"
  print(Path.dirname(file))
  print(Path.basename(file))
  print(Path.extname(file))

  print(Path.dirname("today.txt"))
  print(`[${Path.extname("Makefile")}]`)
  return 0
}
```

```text
notes/today.txt
/home/grace/a.txt
/home/ada
archive.tar.gz
.gz
.
[]
```

- `Path.join(left, right)` puts one `/` between the two and tidies the result: `..` removes a folder, `.` disappears.
- `Path.dirname(path)` is everything before the last name, and `Path.basename(path)` is the last name.
- `Path.extname(path)` is the part of the last name from its last dot, such as `.gz`. A name without a dot gives an empty string.
- The folder of a bare name such as `today.txt` is `.`, the current folder.

Always build a path with `Path.join` instead of writing `${folder}/${name}`. On Windows, a program also reads `\` as a separator, and `Path.join` still writes `/`, which the Windows file functions accept. Path functions only compute: they do not check that the path exists.

## Folders

`std/fs` also makes, lists and removes folders:

```narra
import * as Arrays from "std/array"
import * as Env from "std/env"
import * as Fs from "std/fs"
import * as Int from "std/int"
import * as Options from "std/option"
import * as Path from "std/path"
import * as Str from "std/string"
import * as Time from "std/time"

function scratchFolder(name: string): string uses io, alloc {
  const temporary = Options.withDefault(Env.get("TMPDIR"), "/tmp")
  return Path.join(temporary, `narra-guide-${name}-${Int.toString(Time.unixMillis())}`)
}

function yesNo(value: bool): string {
  if (value) {
    return "yes"
  }
  return "no"
}

export function run(args: Array<string>): int uses io, alloc {
  const folder = scratchFolder("folders")
  print(`exists: ${yesNo(Fs.exists(folder))}`)

  const deep = Path.join(folder, "notes/2026")
  if (Fs.createDir(deep) is Err(message)) {
    print(`createDir: ${Str.replace(message, folder, "<folder>")}`)
  }
  print(`createDirAll: ${Fs.createDirAll(deep) is Ok(true) ? "made" : "failed"}`)
  print(`createDirAll again: ${Fs.createDirAll(deep) is Ok(false) ? "already there" : "failed"}`)
  print(`exists: ${yesNo(Fs.exists(folder))}`)
  print(`isDir: ${yesNo(Fs.isDir(deep))}`)

  const _first = Fs.writeText(Path.join(folder, "notes/a.txt"), "a")
  const _second = Fs.writeText(Path.join(folder, "notes/b.txt"), "b")
  if (Fs.listDir(Path.join(folder, "notes")) is Ok(names)) {
    let sorted = copy(names)
    Arrays.sort(mutable sorted)
    print(`entries: ${Str.join(sorted, ", ")}`)
  }

  if (Fs.removeDir(Path.join(folder, "notes")) is Err(message)) {
    print(`removeDir: ${Str.replace(message, folder, "<folder>")}`)
  }

  // Remove the files, then the folders, deepest first.
  const _a = Fs.removeFile(Path.join(folder, "notes/a.txt"))
  const _b = Fs.removeFile(Path.join(folder, "notes/b.txt"))
  const _year = Fs.removeDir(deep)
  const _notes = Fs.removeDir(Path.join(folder, "notes"))
  const _root = Fs.removeDir(folder)
  print(`exists: ${yesNo(Fs.exists(folder))}`)
  return 0
}
```

```text
exists: no
createDir: <folder>/notes/2026: No such file or directory
createDirAll: made
createDirAll again: already there
exists: yes
isDir: yes
entries: 2026, a.txt, b.txt
removeDir: <folder>/notes: Directory not empty
exists: no
```

What each function does:

- `Fs.exists(path)` and `Fs.isDir(path)` return a `bool`. They need `io`, but not `alloc`, and they never fail: a path they cannot reach is "not there".
- `Fs.createDir(path)` makes one folder and fails when the parent is missing or the folder exists, as the second line shows.
- `Fs.createDirAll(path)` makes the folder and every missing parent. Its result is `Ok(true)` when it made something and `Ok(false)` when the folder was already there, so it is safe to call twice.
- `Fs.listDir(path)` returns the names inside a folder, without the folder's own path. The order is the one the system gives, so the sample sorts the names with `Arrays.sort` (chapter 2); `.` and `..` are not in the list.
- `Fs.removeFile(path)` and `Fs.removeDir(path)` remove a file and an empty folder. A folder with anything inside is an error, as the `removeDir` line shows.

Every function that can fail returns a `Result`, and ignoring one is a compile error (chapter 6). The sample gives the results it does not need to names that start with `_`, which discards them on purpose.

## Standard input and output

`print` writes a line to standard output. `std/io`, imported as `Io`, is the rest of the standard streams:

| Function | Result |
| --- | --- |
| `Io.readLine()` | `Result<Option<string>, string>`: `Ok(Some(line))` for each line without its `\n`, and `Ok(None)` at the end of the input |
| `Io.readAll()` | `Result<string, string>`: everything that is left |
| `Io.readBytes(count)` | `Result<string, string>`: exactly `count` bytes, or an error when the input ends first |
| `Io.write(text)` | writes `text` to standard output, with no newline added |
| `Io.writeError(text)` | writes `text` to standard error, with no newline added |
| `Io.flush()` | writes out what standard output holds |

Standard error is the second stream every program has. A shell can send it to a different place than standard output. A tool writes its messages about problems there and keeps its results on standard output, where the next program can read them.

Counting the lines of the input is a loop that reads until it gets `Ok(None)`:

```narra
import * as Int from "std/int"
import * as Io from "std/io"

function countInput(): Result<int, string> uses io, alloc {
  let count = 0
  while (true) {
    const next = Io.readLine()
    if (next is Err(message)) {
      return Err(message)
    }
    if (next.value is None) {
      return Ok(count)
    }
    count++
  }
}

export function run(args: Array<string>): int uses io, alloc {
  Io.writeError("reading standard input\n")
  Io.write("lines: ")
  match (countInput()) {
    Ok(count) => print(Int.toString(count))
    Err(message) => Io.writeError(`cannot read: ${message}\n`)
  }
  Io.flush()
  return 0
}
```

```text
lines: 0
```

The guide's test gives the program an empty input, so it counts no lines. `narrac run` closes the standard input of the program unless you give it a file: `narrac run -stdin notes.txt count.nra` feeds `notes.txt` to it. For a pipe, build an executable with `narrac build count.nra` and run `cat notes.txt | ./count`. The note "reading standard input" goes to standard error, so it does not show above and does not mix with the result.

`Io.write` writes no newline, and standard output may be buffered. A program that asks a question with `Io.write("name: ")` calls `Io.flush()` before it waits for a line. Then the question shows before the program waits. `Io.readAll` is shorter when the whole input fits in memory.

## The command line and the environment

The `args` that `run` receives holds the command line. The first entry is the path of the program, and the rest are the arguments, as chapter 1 showed. `std/env`, imported as `Env`, has two functions:

- `Env.args()` returns the same list, read from the system again.
- `Env.get(name)` returns the value of an environment variable as an `Option<string>`, `None` when it is not set.

```narra
import * as Arrays from "std/array"
import * as Env from "std/env"
import * as Int from "std/int"
import * as Options from "std/option"

export function run(args: Array<string>): int uses io, alloc {
  print(`arguments: ${Int.toString(Arrays.length(args))}`)
  print(`from Env.args: ${Int.toString(Arrays.length(Env.args()))}`)

  const editor = Options.withDefault(Env.get("NARRA_GUIDE_NO_SUCH_VARIABLE"), "none")
  print(`editor: ${editor}`)
  return 0
}
```

```text
arguments: 1
from Env.args: 1
editor: none
```

The guide's test starts the program with no arguments, so only the program path is there. `narrac run lines.nra one.txt two.txt` would give three entries, and `args[1]` would be `one.txt`.

Use `Env.args()` when your program may run on Windows and has to read arguments that are not ASCII, such as a file named `résumé.txt`. On Windows, the `args` given to `run` come through the system's old character set and may be wrong for such names. `Env.args()` reads them as UTF-8. On the other systems the two are equal.

## The `lines` program

Now put the pieces together. The program takes its files from the command line, and counts standard input when there is none. A file that cannot be read goes to standard error, and the exit code says whether every file was fine. The `run` function's result is the exit code, as chapter 1 said, so returning `1` ends the program with that code:

```narra
import * as Arrays from "std/array"
import * as Int from "std/int"
import * as Io from "std/io"
import * as Fs from "std/fs"
import * as Options from "std/option"
import * as Ranges from "std/range"
import * as Str from "std/string"

function countText(text: string): int uses alloc {
  const lines: Array<string> = Options.withDefault(Str.split(text, "\n"), [])
  let count = Arrays.length(lines)
  if (Str.endsWith(text, "\n") || text == "") {
    count--
  }
  return count
}

function countFile(path: string): Result<int, string> uses io, alloc {
  const read = Fs.readText(path)
  if (read is Err(message)) {
    return Err(message)
  }
  return Ok(countText(read.value))
}

export function run(args: Array<string>): int uses io, alloc {
  // args[0] is the program, so the files start at 1.
  if (Arrays.length(args) < 2) {
    const input = Io.readAll()
    if (input is Err(message)) {
      Io.writeError(`lines: ${message}\n`)
      return 1
    }
    print(`${Int.toString(countText(input.value))} (standard input)`)
    return 0
  }

  let failed = false
  let total = 0
  for (const index of Ranges.between(1, Arrays.length(args))) {
    const path = args[index]
    match (countFile(path)) {
      Ok(count) => {
        total += count
        print(`${Int.toString(count)} ${path}`)
      }
      Err(message) => {
        Io.writeError(`lines: ${message}\n`)
        failed = true
      }
    }
  }
  print(`${Int.toString(total)} total`)
  if (failed) {
    return 1
  }
  return 0
}
```

```text
0 (standard input)
```

With files, the program prints one line per file. Suppose `one.txt` has three lines and `two.txt` has two:

```sh
narrac run lines.nra one.txt two.txt missing.txt
```

```text
3 one.txt
2 two.txt
5 total
lines: missing.txt: No such file or directory
```

`narrac run` shows all of standard output first and standard error after it. The message about `missing.txt` goes to standard error, and the program ends with code `1` after it has counted the other files. In a shell, `echo $?` shows the code of the last command.

## Other programs and exit codes

`std/process`, imported as `Process`, starts other programs. `Process.run(command, arguments)` runs a program, waits for it, and returns `Result<#[int, string, string], string>`. The tuple holds the exit code, what the program wrote to standard output and what it wrote to standard error. The `Err` case means the program could not be started at all, such as a name the system does not know. A program that starts and fails is an `Ok` with a code that is not `0`.

```narra
import * as Int from "std/int"
import * as Process from "std/process"
import * as Str from "std/string"

function runAndReport(command: string, arguments: Array<string>) uses io, alloc {
  match (Process.run(command, arguments)) {
    Ok(#[code, output, errorOutput]) => {
      print(`${command}: code ${Int.toString(code)}`)
      print(`  output: [${Str.trim(output)}]`)
      print(`  errors: [${Str.trim(errorOutput)}]`)
    }
    Err(message) => print(`${command}: not started (${message})`)
  }
}

export function run(args: Array<string>): int uses io, alloc {
  runAndReport("echo", ["hello", "there"])
  runAndReport("sh", ["-c", "echo oops >&2; exit 3"])
  runAndReport("narra-no-such-program", [])
  return 0
}
```

```text
echo: code 0
  output: [hello there]
  errors: []
sh: code 3
  output: []
  errors: [oops]
narra-no-such-program: not started (narra-no-such-program: No such file or directory)
```

The command is not a line for a shell. It is a program name, and each argument is passed as it is, with no quoting and no wildcards. To use shell features, run the shell: `Process.run("sh", ["-c", "ls *.txt"])`. `Process.run` looks the name up on the system's `PATH`. On Windows, `echo` and `dir` are part of the command shell, so run them as `Process.run("cmd", ["/c", "echo", "hi"])`.

`Process.run` captures all the output before it returns, so it suits tools with small output. When the child is stopped by a signal, the code is 128 plus the signal number.

`Process.exit(code)` ends the program at once with that code, after writing out what standard output holds. It has the `io` effect. Nothing after it runs, so a function that has to return a value may end with it, as it may end with `panic`:

```narra
import * as Arrays from "std/array"
import * as Io from "std/io"
import * as Process from "std/process"

function requireName(args: Array<string>): string uses io, alloc {
  if (Arrays.length(args) < 2) {
    Io.writeError("usage: greet NAME\n")
    Process.exit(2)
  }
  return args[1]
}

export function run(args: Array<string>): int uses io, alloc {
  print(`hello, ${requireName(["greet", "Ada"])}`)
  print("done")
  Process.exit(0)
}
```

```text
hello, Ada
done
```

Prefer returning a code from `run` when you can: every value is released in the normal way. `Process.exit` leaves at once and releases nothing, which is right for a helper deep inside the program that finds a problem it cannot continue from.

## Bytes

Not every file is text. An image or an archive can hold any byte, and `Fs.readText` would reject such a file because it checks for UTF-8. `bytes` is the type for data of any kind: an immutable sequence of bytes. `std/bytes`, imported as `Bytes`, works with it, and `Fs.readBytes` and `Fs.writeBytes` read and write files as `bytes`:

```narra
import * as Bytes from "std/bytes"
import * as Env from "std/env"
import * as Fs from "std/fs"
import * as Int from "std/int"
import * as Options from "std/option"
import * as Path from "std/path"
import * as Time from "std/time"

function scratchFolder(name: string): string uses io, alloc {
  const temporary = Options.withDefault(Env.get("TMPDIR"), "/tmp")
  return Path.join(temporary, `narra-guide-${name}-${Int.toString(Time.unixMillis())}`)
}

export function run(args: Array<string>): int uses io, alloc {
  const folder = scratchFolder("bytes")
  const _created = Fs.createDirAll(folder)
  const path = Path.join(folder, "data.bin")

  const data = Bytes.fromArray([104, 105, 255, 33])
  print(`length: ${Int.toString(Bytes.length(data))}`)
  print(Show.show(data))
  print(Show.show(Bytes.at(data, 2)))
  print(Show.show(Bytes.at(data, 9)))
  print(Show.show(Bytes.slice(data, 0, 2)))
  print(Show.show(Bytes.concat(data, Bytes.fromString("?"))))

  const _written = Fs.writeBytes(path, data)
  if (Fs.readText(path) is Err(_)) {
    print("readText: not valid UTF-8")
  }
  if (Fs.readBytes(path) is Ok(back)) {
    print(`readBytes: ${Int.toString(Bytes.length(back))} bytes, same: ${back == data ? "yes" : "no"}`)
    print(Show.show(Bytes.toString(back)))
  }
  if (Bytes.slice(data, 0, 2) is Some(prefix)) {
    print(Show.show(Bytes.toString(prefix)))
  }

  const _removed = Fs.removeFile(path)
  const _gone = Fs.removeDir(folder)
  return 0
}
```

```text
length: 4
bytes[68 69 ff 21]
Some(255)
None
Some(bytes[68 69])
bytes[68 69 ff 21 3f]
readText: not valid UTF-8
readBytes: 4 bytes, same: yes
Err("invalid UTF-8 at byte 2")
Ok("hi")
```

`bytes` has no `print` form of its own, so the sample uses `Show.show` (chapter 7), which writes each byte as two hex digits. What each function does:

- `Bytes.fromString(text)` and `Bytes.fromArray(items)` make `bytes` from a string or an `Array<u8>`. `Bytes.toArray(data)` goes back to an array.
- `Bytes.length(data)` is the number of bytes.
- `Bytes.at(data, index)` returns the byte at `index` as an `Option<u8>`.
- `Bytes.slice(data, start, end)` returns the bytes from `start` up to but not including `end`, as an `Option<bytes>`. Both give `None` for a position outside the data.
- `Bytes.concat(left, right)` joins two values into a new one.
- `Bytes.toString(data)` turns valid UTF-8 into text. Otherwise it returns an `Err` that names the first bad byte, here byte 2, the `ff`.

`bytes` values compare with `==`, and they work as map keys.

## Measure time

`std/time`, imported as `Time`, has a clock for measuring and a clock for the date. `Time.now()` reads the monotonic clock, which only goes forward and suits measuring: `Time.elapsed(start)` is the time since `start`. Reading a clock has the `io` effect. A duration is built with `Time.milliseconds(count)`, `Time.seconds(count)` or `Time.nanoseconds(count)`, and read back with `Time.asMilliseconds(duration)` or `Time.asNanoseconds(duration)`:

```narra
import * as Int from "std/int"
import * as Time from "std/time"

function yesNo(value: bool): string {
  if (value) {
    return "yes"
  }
  return "no"
}

export function run(args: Array<string>): int uses io, alloc {
  const start = Time.now()
  let total = 0
  for (let number = 1; number <= 1000; number++) {
    total += number
  }
  const spent = Time.elapsed(start)

  print(`sum: ${Int.toString(total)}`)
  print(`one second is ${Int.toString(Time.asMilliseconds(Time.seconds(1)))} ms`)

  // A later reading of the monotonic clock is never earlier, so the span to it covers the loop.
  const later = Time.now()
  const covered = Time.asNanoseconds(Time.between(start, later)) >= Time.asNanoseconds(spent)
  print(`the later span covers the loop: ${yesNo(covered)}`)
  return 0
}
```

```text
sum: 500500
one second is 1000 ms
the later span covers the loop: yes
```

`Time.between(earlier, later)` needs no clock, so it has no `io`. A real program prints the measured time. The sample checks it instead, so that its output stays the same on every run.

`Time.sleep(duration)` waits for a duration. It is an async function, so it needs the async code of [chapter 14](14-async.md).

`Time.unixMillis()` is the other clock: the milliseconds since 1 January 1970 in UTC. It can jump when the system's clock is set, so use it for names and timestamps and use `Time.now()` for measuring. The scratch folders of this chapter's samples use it for their names.

The next chapter writes functions that produce a sequence of values one at a time.

Next: [12. Generators](12-generators.md).
