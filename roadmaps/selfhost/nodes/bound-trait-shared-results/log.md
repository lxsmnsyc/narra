# Log

- Proof on base (feat/batch-12 af35fdcb): `narrac build -o bin main.nra; NARRA_CHECK_COPIES=1 ./bin` of the bound-call program (`grow<T: Items>` with `let got = Items.items(item); Arrays.push(mutable got, 3)`) compiles and prints "copy check: copies on write that no copy(...) explains: array 1", exit 103.
- Cause: a call through a bound has Ref.Unresolved and no traitSelf entry, so findSharedArgument saw no `from`. The fix reads the trait from the callee's `Trait.function` name.
- The fix exposed 17 sites in selfhost/src and std/src/shape.nra whose results shared a parameter without a marker (or with another one). Fixed with markers and copy(...); stdtext.nra regenerated.
