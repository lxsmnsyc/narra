# Log: snippet-warnings

## Finding

Narra has no warnings. Every diagnostic is an error with a code (an unused name is E0921, E0922 or E0947 and fails `narrac check` and `narrac run`). So a runnable sample cannot print a stray diagnostic and still pass the guide suite. The site shows diagnostics only on `narra error` fences (`showErrors` in `annotate`); `narra` and `narra file` fences are queried for hovers and their diagnostics are dropped.

## Inventory before (compiler built from this branch, twoslash cache cleared)

Queried every fence of `docs/guide/*.md`, `docs/reference/**`, `docs/install.md`, `README.md`, `SPEC.md` (`narrac query --json hovers`, the plugin's own call), and built the site.

- Shown inline on the site (`tw-diag` spans in the built pages): 172, all on `narra error` fences of the guide. 145 are the diagnostic the fence teaches. 27 are extra.
- Hidden, on 4 `narra file` fences of chapter 4 (lines 634, 674, 725, 735): 23. Cause: the twoslash manifest named every package `sample`, so `import ... "timeline/span"` did not resolve. The guide suite runs them clean.
- `docs/reference/**`, `docs/install.md`, `README.md`, home page samples (`website/src/home/*.md`): 0.
- `SPEC.md`: 135 on its `narra` fences. They are fragments (the spec suite checks them with its own context), the site shows none of them. Left out of the enforcement.
- The `args` warning of the first report is gone since entry-args-unused (the plugin declares the binary).

The 27 extra diagnostics, by kind:

| Kind | Count | Fences | Decision |
| --- | --- | --- | --- |
| Same message at a second place that uses the name (`rule` twice, `players` twice, `score`/`count`/`fetched` captured at each use, `number` twice, `Weight` map key twice) | 12 | 03:488, 20:699, 09:281, 09:612, 20:617, 14:301, 20:96, 07:457 | Related to the lesson. Kept; each is now a `// error:` line. |
| Cascade of the taught error that the sample can hold back | 9 | 03:295, 04:491, 04:545, 05:468 (order), 07:568, 20:535, 20:845 | Noise. Sample changed so only the taught error remains. |
| Cascade that cannot be separated from the taught error (`unknown name "quantity"`, the generic argument after `cannot tell the type of None`) | 2 | 05:468, 06:259 | Related. Kept as `// error:` lines. |
| Unused parameter or local in the sample | 6 | 03:445, 07:166, 20:517, 20:548, 20:825, 03:295 | Noise. The parameter is read now. |

Totals: 12 kept as expected, 15 removed (the per-kind counts above are by fence and are approximate where one fence had several kinds).

## How a fence marks an expected diagnostic

A `narra error` fence starts with one `// error: LINE:COLUMN: MESSAGE` line per expected diagnostic (before, only the first line was read, and the check was "the output contains it"). The guide suite now also fails on any diagnostic of the sample that no line names (a `: note:` line belongs to the diagnostic before it). The site shows each of them inline as before; an expected diagnostic is the point of the sample, so it stays visible. Nothing else gets a new look: unexpected ones fail the build instead.

Positions count the comment lines, so adding a line moves the lines below it by one.

## Enforcement

- Guide suite: `runErrorSample` in `selfhost/tests/src/guide.nra`. `narra` and `narra file` samples already fail on any diagnostic (the compile fails).
- Site: `findUnexpected` in `website/plugins/twoslash.ts`, called by `checkDiagnostics` in `website/plugins/markdown.ts` for every page except `SPEC.md`. It throws, which fails `pnpm build`. Same rule: an error fence announces its diagnostics with `// error:` lines; any other fence announces none.
- Twoslash manifest: a `narra file` package without a `narra.toml` block is named after its fence and gets the `main.run` binary, as the guide suite writes it.

## Proof

A stray `const stray = 1` in the first sample of chapter 1 (restored after):

- guide suite: `FAIL guide/01-getting-started.md:23/debug` and `/release`, `"stray" is never read`.
- `pnpm --dir website build`: exit 1, `a sample shows a diagnostic its fence does not announce: docs/guide/01-getting-started.md, fence narra, 2:3: "stray" is never read ...`.

For the error fences: before the new rule the first guide run with the stricter check failed on 22 fences (24 diagnostics listed); after the fixes it passes, 622 cases.

## Chapter 2

- `Str.padStart("café", 7, '0')` prints `00café` (added to the padding sample and its output).
- `Char.isLetter('é')` is `false` (the sample already printed `é:`; the sentence says it by name).
- `Float.parse("1e-400")` gives `0.0`, `Float.parse("1e400")` is `Err` (`number does not fit in 64 bits`), checked with narrac run.
- The brief says Unicode classes come later in `std/unicode`; the chapter already says `std/unicode` has no letter classes, which is the current truth, so it is not changed.
