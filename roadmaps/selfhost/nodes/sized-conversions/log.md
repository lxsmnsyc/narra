# Log: sized-conversions

- Rows, checks and emitters of the seven sized modules and the three `std/int` conversions are in `selfhost/src/stdsized.nra`, built from one table (`SIZED_ROWS`).
- `parse` uses a new runtime function `narra_sized_parse` in `std_int.c` (64 bit result, range in the message).
- Full `sh scripts/test.sh --update` rewrote only the unknown-module goldens (the module list in the message grew).
