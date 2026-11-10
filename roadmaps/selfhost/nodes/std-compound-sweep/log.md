# Log

- Count at start (branch feat/batch-6, e8f0b61f): 33 sites of `x = x op y` in std/src: 29 in json.nra (`+`), 3 in math.nra (`*`, `/`, `*`), 1 in random.nra (`+`). The brief's 28 and 5 add up to the same 33.
- Not rewritten: json.nra `code = code * 16 + digit` (not the `x = x op y` form; `code *= 16 + digit` would change the value).
- `remaining /= 2` in math.nra has a constant divisor, so the proven-operands rule accepts it.
- `sh scripts/test.sh --suite generated --update` changed only selfhost/src/stdtext.nra (std sources are embedded there). No emitted-C golden changed; the full run passes without any other update.
