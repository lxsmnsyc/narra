# Log: index-assign-order

- Proof on base 6c616a46 (feat/batch-15): `narrac run testdata/index_assign_order.nra` with a compiler built from the base prints `value p` before `index p`, `value r` before `index r`, `value k` before `key k`, and in the async cases the `await` runs before `index a`, `index c` and `key m`. The expected output has each index or key line before the value line.
- Fix: `selfhost/src/emitassignorder.nra` holds each index expression that runs code (call, await, match, block) with `holdArgument` when the right side also runs code; the emitter then reads the held values when it builds the target pointer. The map write emits its key before the right side by the same rule.
- Simple targets (literal or variable indexes) and pure right sides take the old path.
