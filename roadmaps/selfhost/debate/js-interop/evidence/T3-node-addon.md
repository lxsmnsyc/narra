# Proof T3: emitted Narra C inside a Node-API addon

Run after the maintainer's answer to Q2 ("ideally it should be either"). Scratch folder: /private/tmp/claude-501/js-interop-design/t3 (thrown away). Capped at 30 minutes, done in about 10. Nothing installed. Node-API headers come from the Node 26.12.1 install (`include/node/node_api.h`).

## Input

`lib.nra`:

```ts
export function greet(name: string): string uses alloc {
  return `Hello, ${name}`
}

export function run(args: Array<string>): int {
  print(greet("world"))
  return 0
}
```

`narrac emit lib.nra > lib.c`. The function is `f_436f9208_greet(lstr*)`, a hashed static name, which is what c-export-functions replaces with an unhashed exported symbol.

## Glue

`addon.c` (hand-written, about 30 lines) does `#define main narra_program_main` before `#include "lib.c"`. One `napi_value` wrapper copies the JS string into a new `lstr` with `napi_get_value_string_utf8`, calls `f_436f9208_greet`, returns the result with `napi_create_string_utf8`, and releases both strings. `NAPI_MODULE_INIT` registers `greet`.

```sh
cc -O2 -std=c11 -w -shared -fPIC -undefined dynamic_lookup \
  -I$NODE/include/node -o greet.node addon.c -lm
node -e 'const a = require("./greet.node"); console.log(a.greet("Node"))'
```

## Result

- `greet.node` is 51,352 bytes. `require` loads it, and `greet("Node")` returns `Hello, Node`.
- One million calls of `greet("x")` from a JS loop take 176 ms, about 0.18 µs per call.
- The runtime needed no initialization call: the allocator starts lazily. Not tested: a panic inside the call (it would end the Node process today), worker threads, and the leak count at unload.
- Not tested: the wasm fallback through emnapi. That is the first acceptance item of node node-addon-wasm.
