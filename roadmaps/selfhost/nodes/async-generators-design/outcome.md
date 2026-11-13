# Outcome: async-generators-design

## Built

- D110 in `DECISIONS.md`, "Proposed, waiting for the maintainer". Answers:
  1. `AsyncGenerator<T>` is a builtin type with effects on it (as `Generator<T>`), an identity type (copies share one position), lazy. One step is `Async.next(stream): Async<Option<T>>`, an ordinary handle.
  2. `for await` only in async bodies (E1122), only over `AsyncGenerator<T>` (E1123, with hints for `Generator<T>` and `Array<Async<T>>`); `for ... of` over a stream is E1124.
  3. `yield` and `await` in one body; `yield await f()` works; `yield*` only over another `AsyncGenerator<T>` (E1116); no async coroutine (E1115).
  4. `break` releases the loop's reference; the last release frees the frame at its `yield` state. A cancel of a step passes to the stream when it is the only pending step. No `Async.cancel(stream)`.
  5. Pull-based: the body runs only while a step is pending, inline in `Async.next`; FIFO steps, no buffer.
  6. Effects as D78; `ensures` per yielded value with loop facts; `for await (mutable ...)` is E1105.
  7. `Net.chunks` in the build; `File.lines`, `debounce`, `throttle`, `merge` later.
- SPEC: section 2 table row and Control flow bullet, section 8 table row and new subsection "Async generators", all marked "Not built yet". The sample is a `text` block until for-await lands.
- `selfhost/src/diagnostics.nra` header: `E1112-E1129  async generators and for await`.
- Briefs: `nodes/async-generators/brief.md` refined (first child, L, E1112-E1121); new `nodes/for-await/brief.md` (M, E1122-E1129, hard on async-generators).

## Deviations

- The SPEC sample is a `text` block, not `ts`: a `ts` block needs a `spec-skip.txt` line, which is outside this spike's paths. for-await's acceptance turns it into a checked `ts` block.

## Stubs left

- None.

## Gotchas

- E1107 is retired, not reused: its row and explanation become "retired: ...".
- Other batch-13 spikes append D111-D117 after D109 too; the merge keeps D110 before them.
- `AsyncGenerator<T>` must not use `takeOwnCopy` or `narra_gen_next`: the frame is shared.

## Proposed learnings

- A frame that can sit in a waiter list cannot be cloned | the list and timer heap hold its address | selfhost/runtime | no

## Verification

- `sh scripts/test.sh --suite spec`: PASS, 62 cases. `--suite sources`: PASS. `--suite codes`: PASS.

## Size

- Estimated S, actual S.
