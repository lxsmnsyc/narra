# 15. Threads and shared state

The async tasks of [chapter 14](14-async.md) take turns on one thread. That suits work that waits. Work that computes, such as counting, parsing or compressing, needs more than one core, and that takes threads. This chapter splits a calculation over several threads. It passes values between them through channels and keeps a total that every thread changes. Last, it says which values may cross from one thread to another, and when to pick threads or tasks.

The running example counts the prime numbers below a limit. It does nothing but compute, so on two cores two threads can finish it in about half the time of one.

## Run work on several cores

`std/thread` starts threads. Every thread belongs to a scope, and the scope waits for all its threads before it returns:

```narra
import * as Int from "std/int"
import * as Thread from "std/thread"

function isPrime(number: int): bool {
  if (number < 2) {
    return false
  }
  for (let divisor = 2; divisor * divisor <= number; divisor++) {
    if (divisor != 0 && number % divisor == 0) {
      return false
    }
  }
  return true
}

function countPrimes(low: int, high: int): int {
  let count = 0
  for (let number = low; number < high; number++) {
    if (isPrime(number)) {
      count++
    }
  }
  return count
}

export function run(args: Array<string>): int {
  const total = Thread.scope((scope) => {
    const lower = Thread.start(scope, () => countPrimes(0, 50000))
    const upper = Thread.start(scope, () => countPrimes(50000, 100000))
    return Thread.join(lower) + Thread.join(upper)
  })
  print(`${Int.toString(total)} primes below 100000`)
  return 0
}
```

```text
9592 primes below 100000
```

The `divisor != 0` test in `isPrime` is the check the compiler asks for before `%` on a name ([chapter 2](02-values-and-types.md#dividing-by-a-name)). What happened in `run`:

- `Thread.scope(body)` calls `body` with a `Scope` and gives what `body` returns, or nothing when `body` returns nothing. Before it returns, it waits for every thread started in the scope.
- `Thread.start(scope, work)` runs `work` on a new thread and gives a `Thread<int>` at once, a handle to the running thread. The two counts run at the same time, on two cores when the machine has them.
- `Thread.join(thread)` waits until the thread has finished and gives the value `work` returned. A second `join` on the same handle gives the same value without waiting.

This is the difference from an async task. A task runs only while the other tasks wait, so a long calculation in one task holds up all of them. Threads run at the same moment. The price is that the operating system decides when each one runs, so the order of their steps changes from run to run.

A scope waits for its threads even when nobody joins them. A thread whose work returns nothing gives a `Thread<void>`, and a call to `Thread.start` may then stand as a statement:

```narra
import * as Thread from "std/thread"

export function run(args: Array<string>): int {
  Thread.scope((scope) => {
    Thread.start(scope, () => print("one thread"))
  })
  print("the scope has ended")
  return 0
}
```

```text
one thread
the scope has ended
```

With two printing threads, either line could come first. So the examples after this one print only from the thread that called `Thread.scope`, after a `join` or after the scope. Their output is the same on every run.

`Thread.start` has the effects of `work`, plus `alloc` ([chapter 10](10-effects-and-testing.md)). `Thread.scope` and `Thread.join` add none: a join gives the same value on every run, whatever order the threads ran in.

Each thread gets a stack of 8 MiB, as large as the main thread's. A panic on any thread prints its message and ends the whole program with exit code 101.

## What a thread function can use

A thread runs while the code that started it goes on. So its function captures values the way a stored closure does ([chapter 3](03-functions-and-closures.md)). It gets a copy of each name it uses, and it cannot capture a `let`:

```narra error
// error: 7:46: a closure cannot capture "limit" because it is declared with let
import * as Thread from "std/thread"

export function run(args: Array<string>): int {
  let limit = 100
  return Thread.scope((scope) => {
    const worker = Thread.start(scope, () => limit * 2)
    return Thread.join(worker)
  })
}
```

Copy the value into a `const` first, as chapter 3 shows. A thread function has a few more rules:

- It is an arrow function written in the call, or the name of a top-level function. A function kept in a variable is an error, because the compiler has to see what the function captures.
- It cannot be `async`.
- It cannot capture the `Scope`. A `Scope` can be passed to a function, but not stored, returned or captured by a closure that outlives the call. So only the code inside `Thread.scope` starts threads in it.
- Each value it captures, and the value it returns, must be `Sendable`.

`Sendable` is a trait the compiler gives a type by itself. Numbers, strings, arrays, structs and enums of such values are `Sendable`. Types that would let two threads reach the same changing object are not. A `Cell` from `std/cell` ([chapter 9](09-memory-and-mutable.md#state-that-closures-share-cell)) is one of them, so this program is an error:

```narra error
// error: 8:40: capture "visits" of the thread function: Cell<int> is not Sendable
import * as Cells from "std/cell"
import * as Thread from "std/thread"

export function run(args: Array<string>): int {
  const visits = Cells.create(0)
  return Thread.scope((scope) => {
    const worker = Thread.start(scope, () => Cells.get(visits))
    return Thread.join(worker)
  })
}
```

The section on values across threads lists the types that are not `Sendable`. The section on shared state shows `Shared`, which takes the place of a `Cell` between threads.

## Map an array on every core

Splitting the work by hand gets long when there are many parts. `Parallel.map(items, change)` from `std/parallel` splits an array into one part per core and runs `change` on each element on those threads. It gives the results in the order of the items:

```narra
import * as Int from "std/int"
import * as Parallel from "std/parallel"

function isPrime(number: int): bool {
  if (number < 2) {
    return false
  }
  for (let divisor = 2; divisor * divisor <= number; divisor++) {
    if (divisor != 0 && number % divisor == 0) {
      return false
    }
  }
  return true
}

function countPrimes(low: int, high: int): int {
  let count = 0
  for (let number = low; number < high; number++) {
    if (isPrime(number)) {
      count++
    }
  }
  return count
}

export function run(args: Array<string>): int {
  const size = 25000
  const counts = Parallel.map([0, 25000, 50000, 75000], (low) => countPrimes(low, low + size))
  for (const count of counts) {
    print(Int.toString(count))
  }
  return 0
}
```

```text
2762
2371
2260
2199
```

`change` follows the rules of a thread function: written in the call or named, and capturing only `Sendable` constants such as `size`. `Parallel.map` needs no scope, since it waits for its threads before it returns.

## Pass values through a channel

A thread that gives one result at its end can return it to `join`. A thread that gives values as it goes uses a channel from `std/channel`. `Channel.create<T>(capacity)` gives a pair: a `Sender<T>` to put values in and a `Receiver<T>` to take them out, in the order they went in:

```narra
import * as Channel from "std/channel"
import * as Int from "std/int"
import * as Thread from "std/thread"

function isPrime(number: int): bool {
  if (number < 2) {
    return false
  }
  for (let divisor = 2; divisor * divisor <= number; divisor++) {
    if (divisor != 0 && number % divisor == 0) {
      return false
    }
  }
  return true
}

export function run(args: Array<string>): int {
  const #[sender, receiver] = Channel.create<int>(4)
  Thread.scope((scope) => {
    Thread.start(scope, () => {
      for (let number = 90; number < 120; number++) {
        if (isPrime(number)) {
          Channel.send(sender, number)
        }
      }
      Channel.close(sender)
    })
    for (const prime of Channel.values(receiver)) {
      print(`found ${Int.toString(prime)}`)
    }
  })
  return 0
}
```

```text
found 97
found 101
found 103
found 107
found 109
found 113
```

- The capacity, 4 here, is how many values the channel holds that no one has taken yet. `Channel.send(sender, value)` waits while the channel is full, so a fast sender cannot run far ahead of a slow receiver. A capacity below 1 counts as 1.
- `Channel.receive(receiver)` waits for the next value and gives `Some(value)`, or `None` once the channel is closed and empty. `Channel.values(receiver)` is a generator over the same values ([chapter 12](12-generators.md)), so a `for ... of` loop reads until the end.
- `Channel.close(sender)` closes the channel. The values already in it can still be taken. Dropping the last copy of the sender closes it too.
- `Channel.send` gives `true` when it queued the value. It gives `false`, and drops the value, once the channel is closed or no receiver is left.

The calls that wait have the `io` effect. The element type must be `Sendable`, as a capture must.

### Several workers on one channel

Copies of a `Sender` or a `Receiver` are handles to the same channel. When several threads take from copies of one receiver, each value goes to one of them. This gives a pool of workers: one channel carries the jobs, a second carries the results back:

```narra
import * as Channel from "std/channel"
import * as Int from "std/int"
import * as Thread from "std/thread"
import { Receiver, Sender } from "std/channel"

function isPrime(number: int): bool {
  if (number < 2) {
    return false
  }
  for (let divisor = 2; divisor * divisor <= number; divisor++) {
    if (divisor != 0 && number % divisor == 0) {
      return false
    }
  }
  return true
}

function countPrimes(low: int, high: int): int {
  let count = 0
  for (let number = low; number < high; number++) {
    if (isPrime(number)) {
      count++
    }
  }
  return count
}

function work(jobs: Receiver<int>, results: Sender<int>) uses io, alloc {
  for (const low of Channel.values(jobs)) {
    Channel.send(results, countPrimes(low, low + 10000))
  }
}

export function run(args: Array<string>): int {
  const #[jobSender, jobReceiver] = Channel.create<int>(10)
  const #[resultSender, resultReceiver] = Channel.create<int>(10)
  for (let low = 0; low < 100000; low += 10000) {
    Channel.send(jobSender, low)
  }
  Channel.close(jobSender)

  const total = Thread.scope((scope) => {
    for (let worker = 0; worker < 3; worker++) {
      Thread.start(scope, () => work(jobReceiver, resultSender))
    }
    let sum = 0
    for (let job = 0; job < 10; job++) {
      if (Channel.receive(resultReceiver) is Some(count)) {
        sum += count
      }
    }
    return sum
  })
  print(`${Int.toString(total)} primes below 100000`)
  return 0
}
```

```text
9592 primes below 100000
```

The three workers take the ten jobs in whatever order they get to them, so the results arrive in a different order on each run. The sum does not depend on that order.

A request can also carry the `Sender` of its own reply channel, so a worker answers each request on the channel the request names. A `Sender` sent through its own channel is a mistake, though. The channel then holds itself, so it is never freed, and dropping the other senders does not close it. Close such a channel with `Channel.close`.

## Share a value with a lock

A thread gets its own copy of everything it captures, and a change it makes reaches only that copy. A few types are handles instead: every copy of one reaches the same object. `Sender` and `Receiver` are two. `Shared<T>` from `std/shared` is a value that several threads read and change, guarded by a lock:

```narra
import * as Int from "std/int"
import * as Shared from "std/shared"
import * as Thread from "std/thread"

function isPrime(number: int): bool {
  if (number < 2) {
    return false
  }
  for (let divisor = 2; divisor * divisor <= number; divisor++) {
    if (divisor != 0 && number % divisor == 0) {
      return false
    }
  }
  return true
}

function countInto(found: Shared.Shared<int>, low: int, high: int) {
  for (let number = low; number < high; number++) {
    if (isPrime(number)) {
      Shared.update(found, (count) => count + 1)
    }
  }
}

export function run(args: Array<string>): int {
  const found = Shared.create(0)
  Thread.scope((scope) => {
    for (let low = 0; low < 100000; low += 25000) {
      const part = low
      Thread.start(scope, () => countInto(found, part, part + 25000))
    }
  })
  print(`${Int.toString(Shared.get(found))} primes below 100000`)

  Shared.set(found, 0)
  const doubled = Shared.read(found, (count) => count * 2)
  print(`after set: ${Int.toString(doubled)}`)
  return 0
}
```

```text
9592 primes below 100000
after set: 0
```

- `Shared.create(value)` makes one. The value must be `Sendable`.
- `Shared.get(shared)` gives a copy of the value, and `Shared.set(shared, value)` replaces it.
- `Shared.update(shared, change)` stores `change(value)`. The callback gets the current value and returns the new one, so four threads that each add one never lose a step.
- `Shared.read(shared, look)` gives `look(value)`, such as one field of a large value, without copying the rest.

Each of these calls takes the lock, does its work and releases it before it returns. No thread can leave the lock held. While one thread runs an `update` callback, the others wait at their next call on the same `Shared`. Keep callbacks short. A callback cannot be `async`, so a task never waits while holding a lock.

`Cell` from `std/cell` has the same functions except `read`, for one thread and without a lock.

Two rules of locks remain to the programmer:

- A callback that reaches its own `Shared`, as in `Shared.read(found, (count) => count + Shared.get(found))`, could never get the lock. The program stops with exit code 101 and a message that names the call, as a panic does, instead of waiting forever.
- Two threads whose callbacks take two `Shared` locks in opposite orders can each wait for the other, forever. Nothing reports this yet. Take nested locks in one order everywhere, or avoid nesting them.

A `Shared` whose value holds a handle to the same `Shared` is never freed. The debug leak check reports it at the end of the program.

## Count with an atomic

A lock is more than a counter needs. `Atomic<T>` from `std/atomic` holds an integer or a `bool`, and its operations change it in one step that no other thread can split:

```narra
import * as Atomic from "std/atomic"
import * as Int from "std/int"
import * as Thread from "std/thread"

function isPrime(number: int): bool {
  if (number < 2) {
    return false
  }
  for (let divisor = 2; divisor * divisor <= number; divisor++) {
    if (divisor != 0 && number % divisor == 0) {
      return false
    }
  }
  return true
}

function countInto(found: Atomic.Atomic<int>, first: Atomic.Atomic<bool>, low: int, high: int): int {
  let won = 0
  for (let number = low; number < high; number++) {
    if (isPrime(number)) {
      Atomic.add(found, 1)
      if (Atomic.compareExchange(first, false, true)) {
        won = 1
      }
    }
  }
  return won
}

export function run(args: Array<string>): int {
  const found = Atomic.create(0)
  const first = Atomic.create(false)
  const winners = Thread.scope((scope) => {
    const lower = Thread.start(scope, () => countInto(found, first, 0, 50000))
    const upper = Thread.start(scope, () => countInto(found, first, 50000, 100000))
    return Thread.join(lower) + Thread.join(upper)
  })
  print(`${Int.toString(Atomic.load(found))} primes below 100000`)
  print(`${Int.toString(winners)} thread found a prime first`)
  return 0
}
```

```text
9592 primes below 100000
1 thread found a prime first
```

- `Atomic.create(value)` makes one, `Atomic.load` reads it and `Atomic.store` writes it.
- `Atomic.add(atomic, amount)` adds and gives the value from before. It wraps past the largest value, like `+`. It takes integers only.
- `Atomic.exchange(atomic, value)` stores a value and gives the one it replaced.
- `Atomic.compareExchange(atomic, expected, replacement)` stores `replacement` only when the value is `expected`, and gives `true` when it stored. Above, only one thread can turn `first` from `false` to `true`.

The type must be `bool` or an integer type of at most 64 bits:

```narra error
// error: 5:30: Atomic<f64> is not allowed: an Atomic holds bool or an integer type up to 64 bits
import * as Atomic from "std/atomic"

export function run(args: Array<string>): int {
  const ratio = Atomic.create(0.5)
  return 0
}
```

`==` on two `Shared`, `Atomic`, `Thread`, `Sender` or `Receiver` handles is a compile error, and none of them can be a map key. Compare the values they hold instead.

## Values across threads

The compiler checks `Sendable` at each place a value crosses to another thread. Those are a capture of a thread function, the result of one, and the type of a channel, a `Shared` or an `Atomic`. A type is `Sendable` unless it holds, at any depth:

- a function, such as a closure kept in a field,
- a `Dynamic` trait value ([chapter 7](07-traits.md#values-of-different-types-in-one-array)),
- a `Cell`,
- an `Async` handle, a generator or an async generator, which belong to the thread that runs them,
- a `Scope` or a `Thread` handle,
- a raw pointer ([chapter 16](16-unsafe-code-and-c.md)), or a network connection of `std/net`.

`Shared`, `Atomic`, `Sender` and `Receiver` are `Sendable`, since they are made for this. A type cannot be made `Sendable` by hand: `implement Sendable` is a compile error.

A generic function that starts a thread with a value of type `T` states the bound `<T: Sendable>`, as for any trait:

```narra
import * as Int from "std/int"
import * as Thread from "std/thread"

function onThread<T: Sendable>(value: T): T uses alloc {
  return Thread.scope((scope) => {
    const worker = Thread.start(scope, () => copy(value))
    return Thread.join(worker)
  })
}

export function run(args: Array<string>): int {
  print(Int.toString(onThread(41) + 1))
  print(onThread("text from a thread"))
  return 0
}
```

```text
42
text from a thread
```

A value that two threads hold is not copied when it crosses. Both point at the same value, as two names do on one thread ([chapter 9](09-memory-and-mutable.md)). A thread that changes its copy in place copies the value first, so the other thread never sees the change:

```narra
import * as Str from "std/string"
import * as Thread from "std/thread"

function shout(words: Array<string>): string uses alloc {
  let mine = copy(words)
  mine[0] = "HELLO"
  return Str.join(mine, " ")
}

export function run(args: Array<string>): int {
  const words = ["hello", "there"]
  const changed = Thread.scope((scope) => {
    const worker = Thread.start(scope, () => shout(words))
    return Thread.join(worker)
  })
  print(`thread: ${changed}`)
  print(`main: ${Str.join(words, " ")}`)
  return 0
}
```

```text
thread: HELLO there
main: hello there
```

So threads share nothing that changes, except through the handle types of this chapter. Two threads cannot write the same memory at the same moment. A program that imports none of the thread modules has no values shared between threads. Its reference counts ([chapter 9](09-memory-and-mutable.md)) need no atomic operations.

## Threads or async tasks

Pick by where the program spends its time:

- Async tasks wait. Use them when the program spends its time on the network or timers, and the work between the waits is short. Tasks cost little, since they share one thread.
- Threads compute. Use them when the program spends its time on the processor and the work splits into parts. More threads than cores do not make it faster: `Parallel.map` makes one part per core.

The two do not mix freely. Each thread has its own event loop, and an `Async` handle is not `Sendable`, so a handle never moves to another thread. A thread function is a plain function, so it cannot `await`. It may start tasks. When it returns, its thread cancels them, as a program does at its end (chapter 14), and only then does `join` give the result.

`Thread.join`, `Channel.send` and `Channel.receive` block, like `std/fs` and the other plain calls of chapter 14. In async code they stop every task of the thread until they return.

Next: [16. Unsafe code and C](16-unsafe-code-and-c.md).
