# Outcome: std-regex-const

## Built

- `Regexes.checked(pattern): Regex` in `std/src/regex.nra`, `@const` and exported: it calls `compile` and panics with `bad pattern "<pattern>" at <position>: <reason>` on `Err`; a pattern over 60 bytes is cut to 57 and `...`. As the value of a top-level `const` the D105 evaluator runs it, so a bad pattern is E1053 at the constant, with the call chain down to the panic in `regex.nra`. The compiled `Regex` struct goes into the C through the D105 writer (D61 getter).
- Evaluator (`selfhost/src/constfunctions.nra`, `consteval.nra`): `CONST_STD_MODULES` adds `std/char`, `std/option`, `std/result` and `std/regex`. New rows `Char.code`, `Char.fromCode`, `Result.ok`, `Result.err`, `Result.withDefault` (`callCharBuiltin`, `callResultBuiltin`); `std/option` and `std/regex` are source, so their bodies run as Narra. `Result.map`, `andThen` and `mapErr` are in `LEFT_OUT_FUNCTIONS` (E1052, "takes a function"); `Options.map`, `andThen`, `orElse` are caught by the existing `takesFunction`.
- E1360 (`diagnostics.nra`, with an explain text): `Regexes.checked can only be the value of a top-level const; use Regexes.compile`, naming the function as written (`wordPattern ... use compile from std/regex` for `import { checked as wordPattern }`). Reported by `checkRegexCheckedUse` in `selfhost/src/stdregex.nra`, called where a call resolves to a function (`checkNamedCall` and `checkModuleCall` in `calls.nra`) and where a function value does (two sites in `exprs.nra`). A call passes only inside the body of the constant being checked (`CONST_BUSY`, same name and scope) when that constant is not lazy; any function value is E1360. The existing constant rules (E0192, E0193) already keep such a call in the constant parts of the value.
- Samples: `testdata/regex_const.nra` (five `checked` constants, each compared with `compile` at run time on matches and on the `ops` and `ranges` arrays; named groups, `(?i)`, a 41-character anchored pattern), `testdata/regex_const_rows.nra` (the new char, option and result rows, each constant also computed at run time). Error tests: `regex_checked_bad` (E1053, `bad pattern "(a" at 0: unclosed group`), `regex_checked_runtime` (E1360 in a function), `regex_checked_lazy` (E1360 in a `lazy` value), `regex_checked_value` (E1360 as a function value), `regex_checked_renamed` (E1360 names `checked as wordPattern`), `regex_checked_long` (E1053 with the cut pattern).
- E1054 and E1055 no longer suggest `lazy`: the hint is `compute a value this large at run time in a function, or generate a source file`, and for a `Regexes.checked` constant `compile this pattern at run time with Regexes.compile` (`writeLimitHint` in `consteval.nra`). The engine's own limits (10,000 instructions) stop a pattern first: a 9,000-instruction program takes 903,316 steps, so no sample reaches the regex hint.
- SPEC: the `checked` bullet moved from "Not built yet" into the `std/regex` entry; section 10 lists the four added modules. `compiler/README.md`: the E1360 hook and the step count, and the D105 not-supported list updated.

## Proof

- base: ae877384, `base/narrac run testdata/regex_const.nra` (base compiler from `narrac0 build -release selfhost`): `module "std/regex" has no export "checked"`, exit 1.
- branch: `NARRAC=<branch> sh scripts/test.sh --filter regex`: PASS (samples under ASan, the leak check and the copy check); `--filter const`: PASS, 593 cases.

## Deviations

- E1360 is not reported from a registry row: `std/regex` is a source module with only a module row, and a source function has no row to hang a check on. The check is a function in `stdregex.nra` plus one import and two call lines each in `selfhost/src/calls.nra` and `selfhost/src/exprs.nra`, two files outside the brief's shared list (the call sites have the name as written; `checkFunctionCall` and `checkFunctionValue` do not).
- The E1052 message lists the allowed modules, so five goldens outside the owned paths changed by one line each: `testdata/errors/const_eval_arrays_map.err`, `const_eval_arrow.err`, `const_eval_math_sin.err`, `const_eval_plain_function.err`, `const_eval_user_equals.err`. The new E1054 and E1055 hint changed five more: `const_eval_steps.err`, `const_eval_scan_steps.err`, `const_eval_memory.err`, `const_eval_copy_memory.err`, `const_eval_shared_memory.err`.
- More rows than the compile path needs (it calls only `Char.fromCode`): `Char.code` and the three non-function `Result` rows, so allowing the modules does not leave a callable row that the evaluator reports as unsupported.
- Two extra error tests (lazy, function value) beyond the brief's two.
- Not done (outside the shared paths): the guide's chapter 12 section on `std/regex` does not mention `checked` yet; DECISIONS D105's "Built in batch-11" paragraph still says the other modules are node const-functions-std.

## Stubs left

- None. `std/math`, `std/bytes`, `std/fixedarray` and the sized number modules are still not allowed in `@const` bodies (node const-functions-std).

## Gotchas

- `regex_checked_bad.err` and `regex_checked_long.err` hold the line of the `panic` in `std/src/regex.nra` (`240:10`), noted from the review and left as is. Moving `checked` in that file means `sh scripts/test.sh --filter regex_checked --update`.
- The evaluator runs the non-`@const` source bodies of allowed std modules without the `@const` body scan, so a later edit to `compile` that uses a closure or a generator shows up as an "unsupported" E1052 at the constant, not where the edit is. `regex_const.nra` covers the current path.
- A function value of `checked` is E1360 even inside a constant, because the value could be called at run time.

## Proposed learnings

- A std function that may only run at compile time needs its check at the call and function-value hooks (`checkTargetUse`'s sites) | source modules have no per-function registry row | selfhost/src | yes (the error tests)

## Verification

- Steps per constant (scratch compiler that reports `state.steps`): 41-character pattern 14,281 steps (0.14% of the 10 million budget); 44 characters 9,271; 50 characters 9,879. Table in `log.md`.
- `narrac emit selfhost` of the base tree (`git archive feat/batch-14`) with the base and the branch compiler: byte-identical, 224,288 lines.
- After the review fixes, rebased on 87bba23c: `--filter regex`: PASS, 68. `--filter const`: PASS, 598. `--suite errors`: PASS, 776. `--suite generated`: PASS, 22. `--suite sources`: PASS, 1. `--suite codes`: PASS, 1. Emit comparison on 87bba23c (base and branch compilers): byte-identical, 224,350 lines.
- Full check `sh scripts/test.sh` (seed-built, base ae877384), before the review fixes: PASS, 8745 cases in 254s.

## Size

- Estimated M, actual M.
