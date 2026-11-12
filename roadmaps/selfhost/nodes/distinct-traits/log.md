# Log: distinct-traits

- Branch `feat/distinct-traits`, reset to `feat/batch-11` (generic-implement landed).
- Parser reads the trait lists after the base; `DistinctDecl` has `implementations` and `derived`; astdump, fmt and query signature print them.
- `declareTraits` makes a `TypeTraits` entry per distinct type; `typeTraitsKey` knows `TDistinct`; `distinctImplements` answers blocks, derives and the automatic Equatable and Hashable.
- E1012 in `checkListed` for a derive whose base lacks the trait (Shape included).
- Emitter: `stripDistinct` stops at a distinct type that has its own block (`lookupImplementation`), so `<T: Show>` with `T = UserId` runs the user's text.
- Shape: the loader writes two functions per distinct type; the unwrap uses the hidden name `__unwrap`.
- Not done: no seed update, no use of the feature in selfhost/src or std/src.
