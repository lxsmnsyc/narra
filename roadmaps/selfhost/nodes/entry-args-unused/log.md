# Log: entry-args-unused

Base: feat/batch-22 (b373c7b4), narrac built with `sh selfhost/bootstrap.sh <scratch>` (fixed point, seed current).

The exemption already exists: `selfhost/src/check/check.nra` (`entryParameter`, since 2026-11-07, unused-names) skips parameter 0 of an exported one-parameter `func` body that matches `checker.entryTargets`, which `recordEntryTargets` fills from every `[[binary]] entry`.

Commands on the base binary, all with no output (no warning):

- `narrac check a.nra`, `narrac run a.nra`, `narrac check --json`, `--all-targets` on `export function run(args: Array<string>): int { return 0 }`
- the same with `export async function run`, with `await Async.idle()` in the body, with a `test` block
- a package with `entry = "main.run"` and `entry = "app/main.run"`, a workspace with one member
- the file inside the repo root and inside `docs/`
- every `narra` fence of docs/**/*.md, README.md and SPEC.md (975 fences) with `narrac check`: 6 "never read" reports, none about `args`
- every testdata/*.nra with `narrac check`: no `args` report
- `narrac run` of the guide chapters' samples: no `args` report

Only case that warns: a package whose binary has `entry = "run"` (no module part). `entryModule` returns module "" for it, which matches no module of a package, so the entry is not recognized:

    $ narrac check <pkg>      # narra.toml: entry = "run"; src/main.nra: export function run(args...)
    src/main.nra:1:21: parameter "args" is never read; remove it, or name it "_args" ...

That entry form does not name a module of a package (`checkEntries` finds no function either), so it is not a real configuration, and no change was made for it.
