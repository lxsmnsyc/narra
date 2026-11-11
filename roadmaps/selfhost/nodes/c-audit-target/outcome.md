# Outcome: c-audit-target

## Built

- `narrac audit [path]` (`selfhost/src/audit.nra`, `runAudit`, dispatched from `main.nra`). It loads the package (`Modules.load`) and walks the parsed files of every non-std module: functions, trait and impl functions, test blocks, and every statement and expression (blocks in match arms, arrow bodies and block expressions included). Output, sorted by file, line and column: `src/x.nra:12:3: unsafe block in function name` (at the `unsafe` keyword; `in test name` inside a test) and `src/x.nra:4:1: @trusted function name` (at the attribute). Exit 0; a package that does not load prints the load errors and exits 1; more than one operand is exit 2. It does not type-check.
- `@target` on an `extern { } from` block. `parseFile` reads the attribute list first (`parseAttributeList`), then `parseExternBlock(parser, attributes, startIndex)` (or `parseDeclWithAttributes` for other declarations; `parseDecl` is gone). Only `@target` is allowed before a block (other attributes: E1000). It is stored in the new `ExternBlock.attributes` and copied in front of each member's `FuncDecl.attributes`, so `resolveFunctionTarget`, `checkTargetUse` (E0991) and `check --all-targets` apply per member unchanged, and a member the target lacks is not reached, so its library is not linked. The block's trivia record starts at the first `@`.
- fmt: `printExternBlock` prints the block's attributes before `extern` and drops the copied attributes from the members (matched by position). astdump shows the copy on each member.
- SPEC section 12: the `GetTickCount64` sample is the block form with an import, checked by the spec suite; its skip line is removed from `selfhost/tests/spec-skip.txt`. SPEC section 9 describes `narra audit` output. `compiler/README.md` updated.
- Tests: `testdata/audit/basic` (package, `expected.txt`) via emit case `audit/basic`; `testdata/c_interop_target.nra` (+ `.out`: a Windows block of a missing library, used in a guaranteeing branch, builds and runs on the host); `testdata/build-targets/extern_outside_branch.nra` (+ `.err`, emit case `build-targets/extern-outside-branch`, E0991 on every target); `testdata/errors/c_interop_block_attribute`; fmt `testdata/fmt/comment_extern_target`.

## Deviations

- Edited outside `shared:`/`owns:`: `selfhost/tests/src/emitchecks.nra` (two cases and `checkAudit`), as build-targets did. No new diagnostic code; E1003 was not needed (E1000 for a wrong attribute).
- `parseDecl` was folded into `parseDeclWithAttributes` (it had no other caller).

## Stubs left

- None. `@layout` and C-callable functions stay with c-abi-export. The audit does not list nested `Shared.update` calls (SPEC section 7 mentions them); that was not in the brief.

## Gotchas

- Members of a `@target` block carry the attribute copy with the same `pos` as the block's; fmt tells them apart by position.
- `narrac audit` reads source only, so it works on a package that fails type checking.

## Proposed learnings

- None.

## Verification

- `sh selfhost/bootstrap.sh /private/tmp/claude-501/c-audit-target/b2`: stage 1 and stage 2 emit identical C.
- `narrac audit testdata/audit/basic`, `narrac run testdata/c_interop_target.nra` (prints 3.0), `narrac check --all-targets testdata/build-targets/extern_outside_branch.nra` (E0991 on four targets).
- `sh scripts/test.sh` first run failed only the SPEC sample (missing import); after the fix PASS in 173 s, exit code 0 .

## Size

- Estimated S, actual S.

Landed on feat/batch-8 as c8b0889e (reviewed inline by the orchestrator; the sample function was renamed to computeAbsolute).
