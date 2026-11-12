# Premortem: batch-11

It is two weeks from now and batch-11 is late or reverted. Why?

1. **Hot files turned every landing into a rebase.** Nine nodes edit `diagnostics.nra`, SPEC, the README and DECISIONS, and four or five edit the parser, `ast.nra`, `astdump.nra`, `exprs.nra` and `emitexprs.nra`. Change: logic in each node's own module, shared files only for dispatch rows; nodes land as they finish and the next one rebases before review; diagnostic codes and decision numbers are reserved per node, so the tables never collide.
2. **The trait chain ran long.** generic-implement (L) and dynamic-traits (L) are a hard chain through `traits.nra`, `calls.nra` and `emittraits.nra`, with distinct-traits on the same files. Change: distinct-traits lands first; generic-implement can land its constraints step alone; dynamic-traits starts the moment generic-implement lands.
3. **inline broke reference counting or the leak check.** Copying a body and an arrow argument into the caller moves locals across scopes, and the emitter's borrow rules (`emitstate.nra`) assume one function per body. Change: every inline sample runs under the leak check and sanitizers, including early return, panics, generators as callers, and nested inline calls; the fresh review probes them.
4. **A new feature tripped the D104 rule or the debug copy check.** New emitted code (Dynamic boxes, lazy getters, inlined bodies) can copy on write. Change: the batch acceptance reruns the debug `emit selfhost` with `NARRA_CHECK_COPIES=1`; each node with runtime values runs its samples under it.
5. **A gate stalled.** contract-logic and const-functions wait for D102 and D105. Change: both spikes run in wave 1, so the maintainer sees them early; if a gate is not answered, the follow-up node leaves the batch with a reason and the batch ships without it.

Accepted risk: std-function-values may need its emitter hook in many std emitters; the brief allows splitting the emitter half into its own node.
