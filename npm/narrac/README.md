# narrac

The Narra compiler as a command and a `.js` API.

```sh
npx narrac check src/main.nra
```

```js
const { check, format, emit } = require("narrac");

const diagnostics = check(["src/main.nra"]); // [{ code, severity, message, file, range, fixes }]
const text = format("export function  run(): int { return 0 }");
const [{ c }] = emit(["src/main.nra"]);
```

The package picks the native compiler of your platform (`narrac-<os>-<arch>`, an optional
dependency). Without one it runs `narrac-wasm` in the same process, about 7 times slower than
native, and each call starts a fresh instance (93 to 325 ms), so call the API a few times, not in a
hot loop. Set `NARRAC_WASM=1` to use the wasm package even where a native package is installed.

A wasm compiler cannot detect its host, so the loader sets `NARRA_HOST_TARGET` from
`process.platform` and `process.arch`. A value you set yourself wins. On a platform with no row
the variable stays unset, and narrac reports E0989 only for a program that imports `std/build`.
