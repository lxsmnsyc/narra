# Log: std-args

## Proof on the base

- Base: feat/batch-14 at 4403befe. Compiler: `sh selfhost/bootstrap.sh /private/tmp/claude-501/std-args/base` (release build).
- Command: `/private/tmp/claude-501/std-args/base/narrac run testdata/args.nra`.
- Result: exit 1, `args.nra:2:1: unknown module "std/args" (this compiler supports std/int, ...)`.

## Full check

- `sh scripts/test.sh` on 9ef2d6b5 (rebased on feat/batch-14 4403befe): `test.sh: PASS, 8576 cases in 347s`.
- Emit: `narrac emit` of the base `selfhost/` tree with the base and branch compilers is byte-identical (223,852 lines).
- Seed: narrac0 from `seed/narrac.c` compiles args.nra as a local module; the three samples print their goldens.

## Review fixes

- Fixed: flag-name clash with an enclosing command is BadSpec; hints compare names without dashes and need 3 bytes; MissingValue and UnknownFlag name the flag as typed. Noted: help with an unknown path prints the deepest matched command.
- Rebased on feat/batch-14 cb834b6e (std-iterator landed; kept both sides of stdlib.nra, compiler/README.md, the guide chapter and its README line; module-list goldens rewritten). `sh scripts/test.sh`: `test.sh: PASS, 8611 cases in 514s`.
