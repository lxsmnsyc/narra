# Outcome: fix-async-arrow-capture

## Built

- `selfhost/src/emit/emitasync.nra`: `emitTaskCreation` takes `takesPointers`. The creation function of an async arrow or `async {}` block takes every input by value, so it no longer writes `*name` for a large parameter. The creation function of an async function still does.
- `selfhost/src/emit/emitasyncgenerators.nra`: the release function of an async generator is `W<owner>`, no longer `w<owner>`, which is the name of the function thunk (`thunkName`).
- Samples: `testdata/async-arrow-struct-param`, `testdata/async-arrow-struct-capture`, `testdata/async-generator-alias` (`.nra` and `.out`).
- No checker change was needed.

## Proof

- base: feat/batch-26 at 218a578c, `out/narrac run testdata/async-arrow-struct-param.nra`: C error `indirection requires pointer operand` at `f->v_block_10 = *v_block_10;`. `out/narrac run testdata/async-arrow-struct-capture.nra`: the same error at `f->v_block_9 = *v_block_9;`. `out/narrac run testdata/async-generator-alias.nra`: `conflicting types for 'w_b77665f8_counter'`. Exact lines are in `log.md`.
- branch: the same commands with the rebuilt compiler print `9`, `35` and `1 2 3`, and the samples pass in `sh scripts/test.sh` (AddressSanitizer and leak check).

## Deviations

- The struct must be large (eight ints) to reproduce: a four-int struct is passed by value and works on the base.
- Only a captured large parameter of an enclosing function fails; a captured local `const` struct works on the base.
- I used one `sed -i.bak` on `emitasync.nra` to add a placeholder argument (backup removed), against the no-`sed -i` rule.

## Stubs left

- None.

## Gotchas

- The `websocket` workarounds (plain arrow calling a named async function) may stay or be simplified later.
- Seed `seed/narrac.c` still emits the old code; a seed update makes it current.

## Proposed learnings

- Name each C helper kind with a letter that no other helper uses | `w` was both the thunk and the async generator release | selfhost/src/emit | no

## Verification

- `sh selfhost/bootstrap.sh out2`: stage 1 and stage 2 emit identical C.
- `sh scripts/test.sh` with the rebuilt compiler: see the reply.

## Size

- Estimated S, actual S.
- Risk estimated medium, measured low. Two small emitter changes; the whole suite covers every other async form.
