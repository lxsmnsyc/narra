# Outcome: spec-samples-test

## Built

- Suite `spec` in `selfhost/tests/src/spec.nra`, registered in `selfhost/tests/src/registry.nra`. It runs `narrac check --json` on every fenced `ts` block of `SPEC.md`: 46 blocks, 28 checked, 18 skipped.
- A block that starts with a declaration (import, export, struct, enum, function, trait, implement, test, type, opaque, distinct, `@`) is a module; a module without `run` gets an empty one. Any other block is wrapped in the body of `export function run(...) uses io, alloc`.
- Unused-name diagnostics (E0921, E0922, E0946, E0947) are not judged, because samples bind names they do not read.
- Markers are HTML comments right before a block, so the rendered spec is unchanged:
  - `<!-- error: E0119 -->`: the block must fail with that code (used for the distinct-type, `mutable` parameter and `bump` errors).
  - `<!-- context ... -->`: lines that declare the names a block uses. `import` lines go to the top. For statement blocks a `---` line splits module-level lines (before) from lines placed in `run` before the statements (after); without `---` all go into `run`. A module takes them after its code.
- Skip list `selfhost/tests/spec-skip.txt`: `SECTION | FIRST LINE | REASON`. A skipped sample must still fail; when it starts to pass its case fails and says to remove the line. The `skip-list` case fails for a line that matches no sample. Each fix node removes its own lines.
- Samples fixed in `SPEC.md` (wrong, not unsupported):
  - Opaque types, `create`: now `uses alloc` (it returns `Err(...)`).
  - Traits: the example declared its own `trait Show`, which is already the standard trait ("declared twice"). It is now `trait Label` with `label`, and the prose line "Show.show(price)" became `Label.label(price)`.
  - Blocks that mixed declarations with statements were split in two or three blocks: Structs (`makeEntry`), Distinct types, Traits (the call), Mutable parameters (three blocks), Generators (the `for` loop).

## Proof

Not a fix node.

## Deviations

- The brief asks for a skip entry of "section, first line, node or reason"; entries use a `|` separator and the first line may be shortened to its start.
- Context comments were added to 20 blocks; they hide declarations that the prose assumes.

## Stubs left

- 18 skipped samples (see `spec-skip.txt`): std-spec-functions (3), type-aliases, top-level-const, inline, regions, threads (3), async, c-interop, build-conditions (2), contracts, the @unique attribute (audit F15), and three that cannot be checked alone (a signature without a body, `Config.parse` of undefined modules, `app/text`).
- No sample exposed a compiler bug, so no skip says "bug:".

## Gotchas

- The `closure-borrow` skip was not needed: the `each(numbers, (item) => { total = total + item })` sample checks once `each` is declared in its context.
- Changing a sample's first line changes its skip key; the `skip-list` case reports it.

## Proposed learnings

- Mark a SPEC.md sample that needs names with a `<!-- context -->` comment, never by editing the visible code | the rendered spec stays the reader's text | SPEC.md | yes (the spec suite)

## Verification

- `NARRAC=<built narrac> sh scripts/test.sh --suite spec`: 47 cases pass (46 samples plus `skip-list`).
- Full `sh scripts/test.sh`: PASS (8 shards of 695 cases, plus the bootstrap job).

## Size

- Estimated M, actual M.
