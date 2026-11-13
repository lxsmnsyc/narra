# Log: review-nits

Behavior on feat/batch-12 (narrac built from the seed with bootstrap.sh):

1. writeArrowHint: a row with no parameters gets `(items) => Name(items)`. Every
   existing row that reaches it with no parameters is a computed one
   (`Arrays.length`, `Maps.size`), where `(items)` is right and tested in
   `std_function_value_generic`. A fixed zero-parameter row (`Host.os`) has no
   obstacle, so it never reaches the hint. No row shows the wrong text: skipped.
2. `const show = print` -> `a.nra:2:13: unknown name "print"`.
3. Arrow loop labels use `_i<tempCount>`, copy labels `_i<inlineCount>`: two
   counters, one prefix (emitinline.nra lines 220 and 290). Loop labels also carry
   line and column, so a clash needs the same loop emitted as arrow and as copy.
4. An exported generic `@inline function countDown<T>` calling itself, never
   called: `narrac check` exits 0 with no output.
5. runBuild returned 1 on the load-error, no-binary and `-o` paths.

After: items 2, 3, 4, 5 changed; item 1 skipped. See outcome.md.
