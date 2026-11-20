# 14. Async

A program that fetches pages, reads files or talks to other programs spends most of its time waiting. Async functions let it wait on many of those at once, on one thread. This chapter starts several slow jobs together and waits for all of them or for the first. It cancels the ones it no longer needs. Then it uses files and a real network connection from async code.

The standard library has no web client, so the first examples stand a short sleep in for the time a fetch takes.

## Wait for several things at once

Here are three fetches of different lengths. The program starts all three, then prints each result:

```narra
import * as Int from "std/int"
import * as Time from "std/time"

async function fetchPage(name: string, milliseconds: int): string uses io, alloc {
  const waited = await Time.sleep(Time.milliseconds(milliseconds))
  if (waited is Err(_)) {
    return `${name}: cancelled`
  }
  print(`${name} arrived`)
  return `${name} took ${Int.toString(milliseconds)} ms`
}

export async function run(args: Array<string>): int {
  const home = fetchPage("home", 30)
  const news = fetchPage("news", 10)
  const weather = fetchPage("weather", 20)
  print("three fetches started")
  print(await home)
  print(await news)
  print(await weather)
  return 0
}
```

```text
three fetches started
news arrived
weather arrived
home arrived
home took 30 ms
news took 10 ms
weather took 20 ms
```

What happened:

- `async function` declares an async function. Calling it starts a task and gives back a handle at once. The handle's type is `Async<string>`: a string that is on its way. `Async` needs no import.
- The body of `fetchPage` runs until it first has to wait, at `await Time.sleep(...)`. Then the call returns, and `run` starts the next fetch. All three waits overlap, so the program takes about 30 milliseconds, not 60.
- `await home` gives the value of the handle. If the task has not finished, `run` waits for it, and the other tasks keep running in the meantime. That is why "news arrived" prints before "home took 30 ms", although `run` asked for `home` first.
- `export async function run` makes the entry function async, so it can use `await`. The value it returns is the exit code, as for a plain `run`.

`Time.sleep` from `std/time` waits without stopping other tasks. It is an async function, so only async code can call it, with `await` or as a handle. It returns a `Result`, as every function of the standard library that waits does. The section on cancelling explains its `Err`.

`await` works in any expression of an async body, as in `print(await home)` above, and awaiting the same handle twice gives the same value. Outside an async function it is an error:

```narra error
// error: 7:10: await is allowed only in an async function, an async arrow or an async block
async function count(): int {
  return 3
}

export function run(args: Array<string>): int {
  return await count()
}
```

All the tasks run on one thread, taking turns. A task only gives up its turn where it waits on another task, a timer or the network. So two tasks never change the same value at the same moment, and a long calculation in one task holds up all the others. [Chapter 15](15-threads-and-shared-state.md) shows how to run work in parallel on several cores. Each thread has an event loop of its own, and `Thread.startAsync` runs an `async` function on a new thread.

A call of an async function has the effects of the function, plus `alloc` for the task it starts. `await` itself has none ([chapter 10](10-effects-and-testing.md)).

## Every handle is used

A call to an async function only starts the work. A handle that nothing looks at is likely a mistake. So a call left as a statement is an error, like a `Result` nobody checks:

```narra error
// error: 8:3: this Async handle is not used; await it, keep it, pass it on, or let the task run with Async.detach(handle)
async function save(name: string): int uses io, alloc {
  print(`saved ${name}`)
  return 1
}

export async function run(args: Array<string>): int {
  save("notes")
  return 0
}
```

Unlike a `Result`, `const _ = save("notes")` is an error too. Await the handle, keep it in a name, pass it to a function, or say that nobody waits for it with `Async.detach` from `std/async`:

```narra
import * as Async from "std/async"
import * as Time from "std/time"

async function logVisit(page: string) uses io, alloc {
  const waited = await Time.sleep(Time.milliseconds(10))
  if (waited is Ok(_)) {
    print(`logged a visit to ${page}`)
  }
}

export async function run(args: Array<string>): int {
  Async.detach(logVisit("home"))
  print("page shown")
  await Async.idle()
  print("done")
  return 0
}
```

```text
page shown
logged a visit to home
done
```

`logVisit` has no result type, so its handle is an `Async<void>`, and `await` on it gives nothing. A detached task runs to its end without anybody waiting for it. `await Async.idle()` waits until no other task has anything left to do: here, until the visit is logged. Without it, `run` would return first, and the section on the end of a program says what happens to the task then.

## Wait for all of them, or the first

Awaiting each handle in turn works, but `std/async` has two functions for a group of handles:

```narra
import * as Async from "std/async"
import * as Int from "std/int"
import * as Str from "std/string"
import * as Time from "std/time"

async function fetchPage(name: string, milliseconds: int): string uses io, alloc {
  const waited = await Time.sleep(Time.milliseconds(milliseconds))
  if (waited is Err(_)) {
    return `${name}: cancelled`
  }
  return `${name} took ${Int.toString(milliseconds)} ms`
}

export async function run(args: Array<string>): int {
  const pages = await Async.all([fetchPage("home", 30), fetchPage("news", 10)])
  print(Str.join(pages, ", "))

  const first = await Async.race([fetchPage("mirror a", 20), fetchPage("mirror b", 5)])
  if (first is Some(text)) {
    print(`first: ${text}`)
  }
  return 0
}
```

```text
home took 30 ms, news took 10 ms
first: mirror b took 5 ms
```

- `Async.all(handles)` gives an `Async<Array<T>>` with every value, in the order of the handles, once the last task has finished.
- `Async.race(handles)` gives an `Async<Option<T>>` with the value of the first task to finish. It is `None` only for an empty array.

Neither one stops the other tasks. After the race, "mirror a" is still running. The next section stops it.

## Errors and cancel

`Time.sleep` and the functions of `std/net` that wait all return a `Result` whose error is an `AsyncError`, from `std/async`. It has two variants:

- `AsyncError.Failed(message)`: the work failed, with the system's message, such as a connection the other side refused.
- `AsyncError.Cancelled`: someone cancelled the task.

`Async.cancel(handle)` cancels a task. The task does not stop on the spot. Instead the call it waits in gives `Err(AsyncError.Cancelled)` right away, and so does every later call to a waiting function. The body sees the error like any other and returns a value, so `await` on a cancelled handle still gives a value. This is how a timeout looks:

```narra
import * as Async from "std/async"
import * as Int from "std/int"
import * as Time from "std/time"

async function fetchPage(name: string, milliseconds: int): string uses io, alloc {
  const waited = await Time.sleep(Time.milliseconds(milliseconds))
  if (waited is Err(_)) {
    return `${name}: cancelled`
  }
  return `${name} took ${Int.toString(milliseconds)} ms`
}

async function timeout(milliseconds: int): string uses io, alloc {
  const _waited = await Time.sleep(Time.milliseconds(milliseconds))
  return "timed out"
}

export async function run(args: Array<string>): int {
  const slow = fetchPage("slow", 1000)
  const first = await Async.race([slow, timeout(20)])
  if (first is Some(text)) {
    print(text)
  }
  Async.cancel(slow)
  print(await slow)
  return 0
}
```

```text
timed out
slow: cancelled
```

The race ends after 20 milliseconds. `Async.cancel(slow)` then wakes the sleep inside `fetchPage` with `Err(AsyncError.Cancelled)`, and `fetchPage` returns its "cancelled" text. Cancelling a task that has finished does nothing.

A cancel also reaches the task the cancelled one is waiting on, when no other task waits on it. So cancelling a task that awaits a `fetchPage` cancels the fetch too.

## Look at a handle without waiting

`Async.peek` and `Async.then` work on a handle without `await`, so plain functions can use them as well. `Async.idle` gives a handle to await:

```narra
import * as Async from "std/async"
import * as Time from "std/time"

async function fetchPage(name: string, milliseconds: int): string uses io, alloc {
  const waited = await Time.sleep(Time.milliseconds(milliseconds))
  if (waited is Err(_)) {
    return `${name}: cancelled`
  }
  return `${name} page`
}

function describe(handle: Async<string>): string {
  return match (Async.peek(handle)) { Some(text) => `ready: ${text}`, None => "not yet" }
}

export async function run(args: Array<string>): int {
  const news = fetchPage("news", 10)
  Async.then(news, (text) => print(`then sees the ${text}`))
  print(describe(news))
  await Async.idle()
  print(describe(news))
  return 0
}
```

```text
not yet
then sees the news page
ready: news page
```

- `Async.peek(handle)` gives `Some(value)` once the task has finished, and `None` before that. It never waits.
- `Async.then(handle, callback)` runs `callback` with the value when the task finishes. The callback is a plain function, not async, so it cannot `await`. On a handle that has already finished, it runs on the next turn of the loop, not inside the `then` call. The callbacks of a cancelled task never run.
- `Async.idle()` gives an `Async<void>` that finishes once no other task can run: none is ready to go on, sleeping, or waiting for the network.

For an `Async<void>` handle, the `then` callback takes no argument: `Async.then(logVisit("home"), () => print("logged"))`. `Async.peek`, `Async.all` and `Async.race` do not take an `Async<void>`, because they would need an `Option<void>` or an `Array<void>`, and those are not types.

A handle is not a value you can compare. `==` on two handles is a compile error, and a handle cannot be a map key.

## Start tasks from arrows and blocks

An arrow function can be async too. `async (page) => ...` returns a handle each time it is called, so passing it to `Arrays.map` starts one task per element:

```narra
import * as Arrays from "std/array"
import * as Async from "std/async"
import * as Int from "std/int"
import * as Str from "std/string"
import * as Time from "std/time"

struct Page {
  name: string
  milliseconds: int
}

async function fetchPage(name: string, milliseconds: int): string uses io, alloc {
  const waited = await Time.sleep(Time.milliseconds(milliseconds))
  if (waited is Err(_)) {
    return `${name}: cancelled`
  }
  return `${name} took ${Int.toString(milliseconds)} ms`
}

export async function run(args: Array<string>): int {
  const pages = [Page { name: "home", milliseconds: 30 }, Page { name: "news", milliseconds: 10 }]
  const fetches = Arrays.map(pages, async (page) => await fetchPage(page.name, page.milliseconds))
  print(Str.join(await Async.all(fetches), ", "))

  const both = async {
    const home = await fetchPage("home", 5)
    const news = await fetchPage("news", 5)
    return `${home}, then ${news}`
  }
  print(await both)
  return 0
}
```

```text
home took 30 ms, news took 10 ms
home took 5 ms, then news took 5 ms
```

`fetches` is an `Array<Async<string>>`, the type `Async.all` takes. The type of the arrow itself is `(page: Page) => Async<string> uses io, alloc`: the effects of the body are part of it.

`async { ... }` is an expression that starts a task at once and gives its handle. Its value comes from `return`, which ends the block and not the function around it. The block above runs its two fetches one after the other, while `run` could go on with other work.

An async arrow or block may run long after the code that wrote it has moved on. So it captures values the way a stored closure does ([chapter 3](03-functions-and-closures.md)). It gets a copy of each name it uses when the task starts, and it cannot capture a `let`:

```narra error
// error: 6:5: an async arrow or block cannot capture "fetched", which is declared with let
// error: 7:12: an async arrow or block cannot capture "fetched", which is declared with let
export async function run(args: Array<string>): int {
  let fetched = 0
  const counting = async {
    fetched += 1
    return fetched
  }
  return await counting
}
```

Return the new value from the task instead, and add it up after the `await`. The message also names `Cell` from `std/cell` ([chapter 9](09-memory-and-mutable.md#state-that-closures-share-cell)), a slot whose contents tasks can share and change. For the same reason, an async function takes no `mutable` parameter, and a parameter of function type must be `escaping`.

## Files and the network

Each function of the standard library has one form, async or plain. It is async when the operating system can wait for it in the background, as it can for a timer or a network connection. It is plain when the system cannot, as for a file, another program or standard input.

### Files

The operating system cannot wait for a plain file in the background. So `std/fs` ([chapter 11](11-files-and-the-system.md)) has only plain functions, and async code calls them like any other function. Each call stops every task of the thread until it returns. Between calls, `await Async.idle()` lets the other tasks run. This program writes two notes while a task reports its progress, then reads three notes back:

```narra
import * as Async from "std/async"
import * as Env from "std/env"
import * as Fs from "std/fs"
import * as Int from "std/int"
import * as Options from "std/option"
import * as Path from "std/path"
import * as Random from "std/random"
import * as Time from "std/time"

async function reportProgress() uses io, alloc {
  const waited = await Time.sleep(Time.milliseconds(0))
  if (waited is Ok(_)) {
    print("progress: the other tasks get a turn")
  }
}

function readNote(path: string): string uses io, alloc {
  return match (Fs.readText(path)) {
    Ok(text) => text
    Err(_) => `${Path.basename(path)} is missing`
  }
}

export async function run(args: Array<string>): int {
  let random = Random.fromSystem()
  const suffix = Random.nextInt(mutable random, 0, 1000000)
  const temporary = Options.withDefault(Env.get("TMPDIR"), "/tmp")
  const folder = Path.join(temporary, `notes-${Int.toString(Options.withDefault(suffix, 0))}`)
  const _created = Fs.createDir(folder)

  const progress = reportProgress()
  const names = ["monday.txt", "tuesday.txt"]
  for (const name of names) {
    const written = Fs.writeText(Path.join(folder, name), `notes for ${name}`)
    if (written is Ok(_)) {
      print(`wrote ${name}`)
    }
    await Async.idle()
  }
  await progress
  for (const name of ["monday.txt", "tuesday.txt", "friday.txt"]) {
    print(readNote(Path.join(folder, name)))
  }

  for (const name of names) {
    const _removed = Fs.removeFile(Path.join(folder, name))
  }
  const _removedFolder = Fs.removeDir(folder)
  return 0
}
```

```text
wrote monday.txt
progress: the other tasks get a turn
wrote tuesday.txt
notes for monday.txt
notes for tuesday.txt
friday.txt is missing
```

The progress task starts its sleep before the first write. It prints only once `run` waits at `Async.idle()`, since `Fs.writeText` gives no other task a turn.

`Process.run`, standard input and the other plain functions of the earlier chapters work in async code in the same way. They stop every task until they return.

### Network connections: `std/net`

`std/net` opens TCP connections, and its functions that wait are async. `Net.listen(host, port)` gives a `Listener`, `Net.accept(listener)` waits for the next connection, and `Net.connect(host, port)` opens one. `Net.read(connection, maxBytes)` waits for bytes and gives an empty value at the end of the stream, and `Net.write(connection, data)` sends them. Here a server and a client run in one program. Port 0 asks the system for a free port, and `Net.localPort` tells the client which one it got:

```narra
import * as Bytes from "std/bytes"
import * as Net from "std/net"
import { Listener } from "std/net"

async function echoOnce(listener: Listener): int uses io, alloc {
  const accepted = await Net.accept(listener)
  if (accepted is Err(_)) {
    return 0
  }
  const data = await Net.read(accepted.value, 4096)
  if (data is Err(_)) {
    return 0
  }
  const written = await Net.write(accepted.value, data.value)
  if (written is Err(_)) {
    return 0
  }
  return written.value
}

async function ask(port: int, message: string): string uses io, alloc {
  const connection = await Net.connect("127.0.0.1", port)
  if (connection is Err(_)) {
    return "connect failed"
  }
  const sent = await Net.write(connection.value, Bytes.fromString(message))
  if (sent is Err(_)) {
    return "write failed"
  }
  const reply = await Net.read(connection.value, 4096)
  if (reply is Err(_)) {
    return "read failed"
  }
  return match (Bytes.toString(reply.value)) {
    Ok(text) => `reply: ${text}`
    Err(_) => "the reply is not text"
  }
}

export async function run(args: Array<string>): int {
  const listener = Net.listen("127.0.0.1", 0)
  if (listener is Err(_)) {
    print("listen failed")
    return 1
  }
  const server = echoOnce(listener.value)
  print(await ask(Net.localPort(listener.value), "hello"))
  const _echoed = await server
  return 0
}
```

```text
reply: hello
```

`Net.read` gives at most 64 KiB per call, and the network may split a message into several pieces. A program that reads longer messages calls `Net.read` in a loop until it has the whole message. The host is a numeric address such as `127.0.0.1`. A `Connection` closes when the last name that holds it goes away.

## When the program ends

The program ends when the entry function returns, whether other tasks are still running. Those tasks are cancelled then. The call each one waits in gives `Err(AsyncError.Cancelled)`, and the task runs on to its next waiting call, where it stops for good. Nothing it holds leaks. To let the tasks finish first, end the entry with `await Async.idle()`.

The entry can also be a plain `export function run` that starts tasks and returns. The tasks get a turn after it returns, and are cancelled at once:

```narra
import * as Async from "std/async"
import * as Time from "std/time"

async function remind(text: string) uses io, alloc {
  const waited = await Time.sleep(Time.seconds(3600))
  if (waited is Err(_)) {
    print(`not reminded: ${text}`)
    return
  }
  print(text)
}

export function run(args: Array<string>): int {
  Async.detach(remind("water the plants"))
  print("entry returns")
  return 0
}
```

```text
entry returns
not reminded: water the plants
```

The program does not wait an hour: the sleep gives `Err(AsyncError.Cancelled)` as soon as `run` returns.

A plain function cannot wait for a handle, since only async bodies can use `await`. It can start tasks, pass their handles on, detach them, or register a `then` callback.

When every task waits on another task and nothing else can happen, the program can never go on. It stops with the message `async: every task waits on another task` and exit code 101, as a panic does.

## Streams of values: async generators

A generator ([chapter 12](12-generators.md)) produces values one at a time. An async generator may also wait between them. Declare it with `async function*` and the result type `AsyncGenerator<T>`, and read it with `for await`:

```narra
import * as Int from "std/int"
import * as Time from "std/time"

async function* headlines(count: int): AsyncGenerator<string> uses (io, alloc) {
  for (let number = 1; number <= count; number++) {
    const waited = await Time.sleep(Time.milliseconds(5))
    if (waited is Err(_)) {
      return
    }
    print(`fetched headline ${Int.toString(number)}`)
    yield `headline ${Int.toString(number)}`
  }
}

export async function run(args: Array<string>): int {
  for await (const headline of headlines(10)) {
    print(headline)
    if (headline == "headline 2") {
      break
    }
  }
  print("done reading")
  return 0
}
```

```text
fetched headline 1
headline 1
fetched headline 2
headline 2
done reading
```

The call `headlines(10)` runs nothing yet. The body runs only while a reader waits for the next value, so it fetched two headlines, not ten. When the loop ends with `break`, the stream is dropped where it waits at its `yield`, and nothing more of its body runs.

The body is an async body and a generator body at once. It can `await` anywhere, `yield` a value, and `yield*` the values of another async generator. `return` takes no value.

`for await` works only inside an async body, and only on an `AsyncGenerator<T>`. A plain `for ... of` over a stream is an error:

```narra error
// error: 7:31: for ... of cannot read AsyncGenerator<int>, whose values may wait; write `for await (const x of stream)` in an async body
async function* numbers(): AsyncGenerator<int> {
  yield 1
}

export async function run(args: Array<string>): int {
  for (const number of numbers()) {
    print("one")
  }
  return 0
}
```

To read one value at a time, `Async.next(stream)` gives an `Async<Option<T>>`: `Some(value)` at the next `yield`, and `None` once the body has ended. It is a plain handle, so it can be raced against a timer like any other. Copies of a stream share one position, so two loops over one stream each get some of its values. `Net.chunks(connection, maxBytes)` in `std/net` is an async generator of the reads of a connection.

Next: [15. Threads and shared state](15-threads-and-shared-state.md).
