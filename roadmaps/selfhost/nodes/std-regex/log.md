# Log: std-regex

## Proof on the base

- Base: feat/batch-14 at 183eec94. Compiler: `sh selfhost/bootstrap.sh /private/tmp/claude-501/std-regex/base` (release build).
- Command: `/private/tmp/claude-501/std-regex/base/narrac run testdata/regex.nra`.
- Result: `regex.nra:2:1: unknown module "std/regex" (this compiler supports std/int, ...)`, exit 1.

## Work

- Fast loop: `/private/tmp/claude-501/std-regex/try.sh` copies `std/src/regex.nra` into a scratch package as `t/regex` and runs a sample with the base compiler, so the engine is tested without rebuilding narrac.
- Fixed while testing: the search stopped when the thread list was empty before any match, so assertions after position 0 never ran (`\bfoo\b`, `(?m)^b`).
- `test` is a keyword: `Regexes.test(...)` fails to parse (`expected a name, found "test"`); renamed to `matches`.
- Timings (`narrac run` build): `(a*)*b` on 10,000 `a` 27 ms, `(a|aa)*c` on 5,000 20 ms; `a{0,1000}a{1000}` on 1,000 `a` 644 ms (left out of the sample).
- Branch compiler: base `narrac emit selfhost` then `cc -O2 -DNARRA_RELEASE`, after `--suite generated --update`.
- Emit: `narrac emit selfhost` of `git archive feat/batch-14` with base and branch compilers: byte-identical, 224,032 lines.
- `--filter regex`: PASS, 28 cases. guide 309, sources 1, `--filter unknown_module` 15: PASS.

## Review fixes

- Rebased on feat/batch-14 9d3fe64b (std-terminal landed): kept both sides of stdlib.nra, SPEC, README and the guide; regenerated stdtext and the module-list goldens.
- Repeats: each pass emits `Reset` for the capture slots of its groups; each pass beyond the lower count is `Save(slot)`, the body, `Progress(slot)`, with one slot per repeat node. Checked against `node`: `(?:(a)|b)+` on "ab" gives group 1 undefined, `(|a)+` on "aa" gives 0..2 with "a", and nine more repeat cases agree.
- Memory: a program whose instructions times slots is above 1,000,000 is a `RegexError` (3,000 groups in the errors sample). The addThread stack grows on demand.
- `😀` is one code point; `split(/x*/, "")` is `[]` as in `node`; `Match.names` is shared through `from regex`.
- Guide: the base had the terminal section inside the toml sample (guide suite failed at `12-standard-library.md:274`); fixed.
- Emit: `narrac emit selfhost` of `git archive feat/batch-14` (9d3fe64b) with base and branch compilers: byte-identical, 224,243 lines.

## Full check

- `sh scripts/test.sh` on the commit (base feat/batch-14 183eec94): `test.sh: PASS, 8691 cases in 626s`.
- After the review fixes, rebased on feat/batch-14 9d3fe64b: `test.sh: PASS, 8715 cases in 304s`.
