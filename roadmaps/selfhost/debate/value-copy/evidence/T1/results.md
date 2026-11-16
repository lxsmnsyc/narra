# P9 non-retaining alias proof (all runs: cc -O0 -g -std=c11 -w -lm, NARRA_CHECK_LEAKS=1 NARRA_CHECK_COPIES=1)
Files: a/b/c/d.nra (compile today, use copy()), *.c emitted, *_p9.c hand-edited.

## A  (a_p9.c)
Edit: delete `NARRA_COPY_GRANTED(..._retain(t2););` at the binding; delete both `_release(v_synonym_20)` at scope end. v_synonym_20 = same pointer as v_prices_19.
Today (a.c): prints 2, then panic index 2 out of range (copy semantics), exit 101.
Edited: prints 3, 300, exit 0, no leak report, no copy (exit not 103).
Verdict: works. push never changes the header pointer (only a->items is realloc'd; cap 2->4 was exercised), so same-pointer alias is enough while rc==1.

## B  (b_p9.c)  let saved = prices; prices = [7]; push(mutable saved, 5)
Edit: remove retain at binding; remove `_release((*t4))` in the reassignment of prices (old array's ownership moves to saved).
Edited: prices len 1, prices[0]=7, saved len 3, saved[2]=5, exit 0, clean.
Negative (keep the reassignment release): saved dangles, len 1 then panic exit 101 (use after free).
Verdict: works; compiler must drop the release at a reassignment whose old value is aliased.

## C  (c_p9.c, d_p9.c)  alias, then `if (cond) { prices = [7] }`, push(saved, 5), later reassign / scope end
(condition changed in C to len(args) > 1 so both paths run)
Edit: no retain; no release in the branch's reassignment (alias known live there); after the merge the state is "maybe aliased", so
  later `prices = [9]` -> `if (*t9 != v_saved_20) release(*t9)`; scope end -> `if (v_prices_19 != v_saved_20) release(v_prices_19)`.
Results: branch not taken: prices len 3, saved len 3 (write-through), exit 0 clean. Branch taken: prices len 1, saved len 3, exit 0 clean. Both c_p9 (with later reassign, prints 9) and d_p9 (merge state reaches scope end).
Verdict: no stored flag needed, but a runtime pointer-equality guard on every release of a possibly-aliased name after a merge point. Without the guard it is a double free or a leak.

## E  (judge)  function reset(mutable items: Array<int>) { items = [9] }
Rejected today: "mutable parameter "items" cannot be assigned whole". A callee cannot replace the header an alias of the argument points at.

## F  (judge, f.diff)  const items = bag.items; reset(mutable bag) where reset does bag.items = [9]
Non-retaining alias: prints 2 and 0 (expected 2 and 100), exit 0, checks silent.
AddressSanitizer: heap-use-after-free at f_p9.c:985. The callee releases the old items array that the alias still names.
Verdict: a `mutable` call on a container of an aliased place must stay E0978 (or the alias must retain, which brings the copy back).
