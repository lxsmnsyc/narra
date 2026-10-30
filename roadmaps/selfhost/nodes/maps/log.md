# Log: maps

- Checked first: no `Map` in the compiler, so the whole node was needed.
- Added `MapType`, `MapLit`, `std_map.go`, `runtime_map.h`; map C functions go after the trait functions in `Emit`.
- `maps.nra` passed under ASan and the leak check on the first full run. The float key error test needed its expected text changed to name the key type.
- Fast check passes.
