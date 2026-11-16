Written samples (not compiled; no compiler was edited). Snippet: let prices = [100, 200]; const synonym = prices; Arrays.push(mutable prices, 300).

A. Local alias (P9). Compiles. After the push, synonym is the same array: synonym.length is 3, synonym[2] is 300.
   `let saved = prices; prices = []` still leaves saved at the old array (reassigning a name never changes the object), as in TypeScript.
   `return synonym`, `box.items = synonym`, `Cells.set(cell, synonym)`, `spawn(f, synonym)` then a write to prices: E0978,
   "synonym may leave this function while prices changes; write copy(prices) where it is taken".
B. Move (P12). Error at the push: "prices was moved into synonym at line 6; write const synonym = copy(prices) to keep both".
   Also errors on every later read of prices, e.g. `print(prices.length)` after the binding.
C. Reworded error (P13). Error at the push: "synonym is another name for prices' array and is read at line 8; prices cannot change until then".
D. Marker (P14). `const synonym = prices` is E0978 at the push with fix "write alias prices (same array) or copy(prices) (own array)". `const synonym = alias prices` compiles as in A.

Chapter 9 today (line 13): "Both names refer to the same array, and its count goes up to 2" for copy(...). Under A the count story is replaced by: a plain
`const b = a` makes b another name for the same array; changes through either name show in both; copy(a) makes b its own array.
Home section 02 under A: TypeScript and Narra both print 70 70 for `const saved = hero; damage(mutable hero, 30)`; Narra adds `copy(hero)` for 70 100.
That removes the headline contrast "Values never change behind your back" for local names; it stays true for function arguments, results, threads.
