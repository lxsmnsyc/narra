# Log: std-regex-const

## Proof first

- base ae877384 (feat/batch-14), compiler built with `narrac0 build -release` from the base tree.
- `base/narrac run testdata/regex_const.nra`: `regex_const.nra:9:22: module "std/regex" has no export "checked"` (five such lines, then type errors), exit 1.

## Design

- `checked` is `@const` in `std/src/regex.nra`: `compile`, then `panic` with `bad pattern "<pattern>" at <position>: <reason>` on `Err`. The evaluator turns the panic into E1053 with the call chain.
- The evaluator runs `compile` and its helpers as Narra bodies: `std/regex` joins `CONST_STD_MODULES`, with `std/char`, `std/option` and `std/result`. The compile path calls only `Char.fromCode` of those; rows added: `Char.code`, `Char.fromCode`, `Result.ok`, `Result.err`, `Result.withDefault`. `std/option` is source, so its bodies run. `Result.map`, `andThen`, `mapErr` go to `LEFT_OUT_FUNCTIONS` (they take a function); `Options.map` and friends are caught by `takesFunction`.
- E1360: std/regex has no registry row per function (the module is source), so the check sits in `stdregex.nra` (`checkRegexCheckedUse`) and runs where `checkTargetUse` runs: every function call (`calls.nra`) and every function value (`funcvalues.nra`). A call is fine only when the current body is a `const` body whose constant is being checked (`CONST_BUSY`, same name and scope) and is not lazy. The constant rules (E0192/E0193) already keep calls of a constant value in its constant parts, so no further position test is needed. A function value is always E1360: `const F = Regexes.checked` would run at run time.
- The E1052 message lists `CONST_STD_MODULES`, so five existing goldens changed (`const_eval_arrays_map`, `const_eval_arrow`, `const_eval_math_sin`, `const_eval_plain_function`, `const_eval_user_equals`).

## Steps of a pattern

A scratch copy of the compiler reported `state.steps` after each constant (`/private/tmp/claude-501/std-regex-const/steps`):

| pattern | length | steps | of 10M |
| --- | --- | --- | --- |
| `^[A-Za-z_][A-Za-z0-9_]{0,30}\.(nra\|toml)$` | 41 | 14,281 | 0.14% |
| `(\d{4})-(\d\d)-(\d\d)T(\d\d):(\d\d):(\d\d)Z?` | 44 | 9,271 | 0.09% |
| `(?<user>[\w.+-]+)@(?<host>[a-z0-9-]+\.)+[a-z]{2,6}` | 50 | 9,879 | 0.10% |
| `[a-z]{1,40}` | 11 | 9,902 | 0.10% |

## Emit check

`narrac emit selfhost` of the base tree (`git archive feat/batch-14`) with the base and the branch compiler: byte-identical, 224,288 lines.
