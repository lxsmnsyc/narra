# Outcome: spec-fixes

## Built

Four text fixes in `SPEC.md`. No code changed.

- Lines 329-330: `F64.toString(...)` -> `Float.toString(...)`.
- Line 534: `` `I32.from(x)`, `F64.fromInt(x)`, `Int.truncate(f)` `` -> `` `Float.fromInt(x)`, `Float.truncate(f)`, and for sized numbers (planned) `Int.toI32(x)` ``.
- Line 552: `parsePair` now reads `... : Option<Pair> uses alloc {`.
- Line 578: `Err(IoFailed(e))` -> `Err(ConfigError.IoFailed(e))`.

Line 57 (unused locals are errors) is unchanged, as the brief says.

## Deviations

- None. The brief names no replacement for `I32.from`. I chose `Int.toI32(x)` in `std/int`, since the std list has no per-type modules.

## Stubs left

- Line 625 (`Array<u8>`, `Arrays.withCapacity`) still depends on sized-numbers and std-spec-functions.

## Gotchas

- For sized-numbers: the spec now says conversions to sized types are `Int.toI32(x)` style functions in `std/int`. Pick the real names (and the reverse direction, and `f32`) and change line 534 to match. No `std/i32` or `std/f64` module exists in the std list (line 193).
- The `readConfig` sample needs `enum ConfigError { IoFailed(message: string) ... }` to check; the spec leaves the type out.

## Proposed learnings

- Run every SPEC.md `ts` sample with `narra check` when a feature node closes | names in the spec drifted from the std modules | `SPEC.md` | yes: extract fenced blocks in a test

## Verification

- `narra check` on a scratch file with the fixed samples: clean.
- `cd compiler && go vet ./... && go test ./...`: passes (see below).

## Size

- Estimated S, actual S.
