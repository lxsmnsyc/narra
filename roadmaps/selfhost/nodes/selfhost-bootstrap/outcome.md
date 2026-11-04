# Outcome: selfhost-bootstrap

## Built

- `compiler/internal/narra/selfhost_bootstrap_test.go`: `TestSelfhostBootstrap` builds stage 2 (stage 1 emits `selfhost`, `cc` builds it with the ASan/UBSan flags of `modes[0]`, cached once per run), checks stage 1 C equals Go's, checks the fixed point (stage 2 `emit selfhost` equals stage 1 output), then compares stage 2 with Go on every `emitInputs` entry, with `NARRA_CHECK_LEAKS=1`.
- `selfhost/bootstrap.sh`: Go build, stage 2 build, `cmp` of Go, stage 1 and stage 2 C. Needs `go` and `cc`; exits non-zero on a difference.
- `compiler/README.md`: "Bootstrapping" section.
- `selfhost_test.go`: `TestMain` also removes the stage-2 directory.
- No divergences between stage 2 and Go; no compiler or `selfhost/src` change was needed.

## Deviations

- None.

## Stubs left

- None.

## Gotchas

- Times (release builds, warm): `narrac1 emit selfhost` 0.33 s (first run 0.59 s), `narrac2 emit selfhost` 0.34 s (first run 0.60 s). Stage 2 binary (`-O2`): 1,923,624 bytes. Test runtime of `TestSelfhostBootstrap`: 21 s (includes the stage-1 sanitizer build); `bootstrap.sh`: about 24 s cold, mostly Go build and C compile. Under the test's sanitizer builds, stage 2 emits `selfhost` in about 4 s.
- No `testing.Short()` gate was needed (under 60 s).

## Proposed learnings

- None.

## Verification

- `cd compiler && go test -count=1 ./internal/narra -run TestSelfhostBootstrap`: pass.
- `sh selfhost/bootstrap.sh`: pass (95,033 lines identical).
- `cd compiler && go vet ./... && go test ./...`: pass (full suite 151 s).

## Size

- Estimated M, actual S.
