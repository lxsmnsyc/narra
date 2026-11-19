# Proposal: structure of the Narra docs

For the maintainer to approve or change. Audited on `feat/batch-20` at `94107ca4` (2026-11-19). Nothing in the docs is changed by this node.

## Decision needed

1. Approve the before-release pass (one node, `docs-release-pass`). It fixes pages that are wrong today for a first user.
2. Approve a Reference section on the site, filled after the release: standard library, commands, `narra.toml`, error codes and the specification.
3. Approve that `narrac` generates the std reference and the error index. The pages are committed as Markdown, and a test fails when they are out of date.

## Current structure

The site (`website/`) renders three sets of repository files at build time (`website/src/pages.ts`). The header has three links: Guide, Install, Spec (`website/src/routes/_layout.brm`). Search indexes the same files (`website/plugins/search.ts`).

| Page | What it holds | Who reads it | What links to it |
| --- | --- | --- | --- |
| `/` (`website/src/routes/index.brm`) | Install command, benchmark chart, value semantics, effects, tools for agents, chapter list | New visitors | Header logo |
| `/guide` (`docs/guide/README.md`, 63 lines) | Chapter list, how to read samples, a "Planned" list and a "built but not taught" list | Learners | Header, sidebar, chapter 13 |
| `docs/guide/01-getting-started.md` (145 lines) | Builds the compiler with `git clone` and `sh selfhost/bootstrap.sh out`, first program, first error, a two-file package, a command table | Learners, from the home page button | Home "Start the guide", guide index, `docs/install.md` |
| `docs/guide/02` to `10` (278 to 614 lines each) | Values, functions, structs, patterns, Option and Result, traits, modules and packages (git dependencies, workspaces), memory and `mutable`, effects and tests | Learners | Previous and next links, sidebar |
| `docs/guide/11-standard-library.md` (472 lines) | Tutorials for 7 source modules: `std/iterator`, `std/args`, `std/arena`, `std/toml`, `std/terminal`, `std/regex`, `std/unicode` | Learners writing tools | Sidebar |
| `docs/guide/12-tools.md` (396 lines) | `check --json`, `explain`, `apply-fix`, `fmt` and `[fmt]`, `query`, `test`, `lsp`, and a command table | Learners, agent users | Sidebar |
| `docs/guide/13-narra-for-typescript-programmers.md` (1098 lines) | Topic-by-topic comparison with TypeScript, and a short Rust and Go section | TypeScript programmers | Sidebar, chapter 12 |
| `/install` (`docs/install.md`, 179 lines) | One-line install, folders, prebuilt systems, options, prerelease caveat, update, hello world, build from source with the bootstrap stages, exit codes, `linux-check.sh` | First users, people on other systems | Header, `README.md`, chapter 1 |
| `/spec` (`SPEC.md`, 2111 lines, 226 KB) | The language rules, the std module rules (section 3), code generation, build model, packages, deferred features | Advanced readers, agents, contributors | Header, home "Read the spec" |
| `README.md` (repo only) | Pitch, "Status: specification draft", contributor commands, install link, skill install, goals, file map, prior art | GitHub visitors | GitHub |
| `DECISIONS.md` (repo only, 122 decisions) | Reasons for each design rule | Contributors | 99 `D` numbers in `SPEC.md`, as plain text |
| `seed/README.md`, `compiler/README.md`, `website/README.md`, `skills/narra/SKILL.md` (repo only) | Contributor docs and the agent skill | Contributors, agents | `README.md` (skill only) |

Links between documents: `README.md` links to `docs/install.md`. `docs/install.md` links to chapter 1. Chapter 1 links to `docs/install.md`. Chapter 13 links to the guide index. Nothing in the guide or the install page links to `SPEC.md`.

## Findings

Each finding was checked against the files or by running `narrac` built from `seed/narrac.c`.

**Outdated**

- Chapter 1, lines 5 to 15, builds the compiler from source. `docs/install.md` and the home page show the one-line install from contract `install-command`. The guide index (`docs/guide/README.md`, line 9) also says "build the compiler".
- `docs/guide/README.md`, "Planned", lists "Workspaces of several packages". Chapter 8 has taught workspaces since batch-19 (`## Workspaces`, line 361).
- `SPEC.md` names the compiler `narra` in 25 lines and `narrac` in 14. Line 5 says "Compiler command: `narra`". The real command is `narrac`.
- `SPEC.md` section 15 is titled "Packages (sketch)". It now holds the full rules for dependencies, the lock file, `[exports]` and workspaces, with error codes.
- `README.md` says "Narra is a draft design" and "Status: specification draft". After the release tag, `narrac` installs with one command.
- `docs/install.md` checks the install with `narrac targets`. Node `narrac-version` changes this to `narrac --version` in the same batch, so this proposal does not repeat it.
- `docs/install.md`, the prerelease paragraph, explains that `releases/latest` returns 404 while only prereleases exist. Whether it stays depends on the first tag (node `first-release`).

**Rule violations in the guide** (the guide assumes no TypeScript outside chapter 13)

- `docs/guide/02-values-and-types.md:66`: "as in TypeScript".
- `docs/guide/11-standard-library.md:360`: names the regex syntax after the browser scripting language that TypeScript compiles to.
- `docs/guide/11-standard-library.md:432`: compares the case functions with `toUpperCase()` of that same language.
- No SPEC or decision citations were found in the guide or `docs/install.md`.

**Repeated content**

- `docs/install.md` explains the three bootstrap stages, the `ok:` line and `seed is current`. `seed/README.md` explains the same stages for contributors. A user who installs from source needs the commands, not the stages.
- `docs/install.md` ends with `sh scripts/linux-check.sh`, which runs the repository's tests. That is contributor information.
- Chapter 1 and chapter 12 each end with a command table. They do not repeat each other: chapter 1 lists `run`, `build`, `check`, `test`; chapter 12 lists `check`, `explain`, `apply-fix`, `fmt`, `query`, `test`, `lsp`. Neither lists every flag.

**Missing**

- A std API reference. Today the function lists of a module are in `SPEC.md` section 3 (13 bullets for some modules) or nowhere. Builtin modules such as `std/string`, `std/array`, `std/map`, `std/fs` and `std/path` have no list a learner can read.
- The brief's lead that `narrac query interface MODULE` can generate it does not hold yet. Measured with the seed build:
  - `narrac query --package p interface std/iterator` prints `error unknown module "std/iterator"`. The same holds for `std/string` and `std/toml`. The header of `selfhost/src/tools/query.nra` says "Only the package's own modules are searched, not std source modules."
  - 18 public std modules are builtin (`std/array`, `std/string`, `std/int`, `std/fs` and others). They have no source file. Their signatures are in the builtin registry, which `narrac builtins` prints in an internal format (950 lines, including `std/internal/*`).
  - The source modules in `std/src/` have no doc comments on their exports, and builtins have no description field. A generated reference holds signatures only until descriptions are written.
- A command reference. `narrac --help` and `narrac help` fail with "No such file or directory" (node `narrac-version` adds `--help` from the existing usage text). Bare `narrac` prints the internal dump commands (`parse`, `pkg`, `decls`, `types`, `builtins`). The usage text in `selfhost/src/driver/driver.nra` covers `build`, `run`, `test`, `check`, `targets` and `explain`, but not `fmt`, `query`, `lsp`, `apply-fix` or `emit`. Flags such as `-emit-c` appear in no guide page.
- A `narra.toml` reference. The tables are spread over five places:
  - `[package]` and `[[binary]]` in chapter 1.
  - `[dependencies]`, `[overrides]`, `[exports]` and `[workspace]` in chapter 8.
  - `[fmt]` in chapter 12, and its keys in `SPEC.md` section 2.
  - `[link]` only in `SPEC.md` section 15.
- An error code index. `selfhost/src/driver/diagnostics.nra` has 479 codes, 96 of them with a longer explanation. `narrac explain CODE` prints one code. No command lists them all.
- Evidence for a home page claim. "an agent passed 19 of 20 Narra tasks, against 15 of 20 without" links to nothing. The numbers are only in `roadmaps/selfhost/nodes/narra-skills/outcome.md`. The benchmark chart does link to `bench/`.
- A `LICENSE` file. The repository has none. This is the maintainer's choice, not a docs structure question, but a release without one cannot be reused by others.
- Untaught features. The guide index lists features that are built but not taught. They include async, threads and channels, SIMD, fixed arrays, `@const`, refinement types, unsafe code and `--target`.

**Kept, with the reason**

- Chapter 11 stays. It teaches 7 modules through programs. A generated reference lists signatures of every module. The readers differ: one learns, one looks up.
- Chapter 12 stays. It teaches tasks (read an error, apply a fix, ask a question). Its command table becomes a link to the command reference once that page exists.
- `SPEC.md` stays whole on the site. It is the source of truth (`GOAL.md`, Constraints), and it is the only place that states some std rules. The site has no visit data, so there is no evidence that readers skip any section. Section 11 (code generation) is the strongest candidate to leave out later, if the maintainer wants a shorter page.
- `DECISIONS.md` stays off the site. The site shows what Narra is. The reasons stay on GitHub for contributors.

## Proposed structure

Header: Install, Guide, Reference. Install comes first because chapter 1 starts from an installed `narrac`.

1. `/`: home page, unchanged except the evidence link for the agent claim.
2. `/install` (`docs/install.md`): install, check with `narrac --version`, where it installs, prebuilt systems, options, update and uninstall, hello world, build from source (requirements, three commands, copy to `PATH`, update). The bootstrap stages, exit codes and `linux-check.sh` move to `seed/README.md`.
3. `/guide` (`docs/guide/README.md`): chapter list, reading samples, the "Planned" list without workspaces.
4. Chapters 1 to 13, same order. Chapter 1 opens with "Install `narrac`" (the one-line command and a link to `/install` for other systems and source builds), then `narrac --version`.
5. `/reference` (new, `docs/reference/README.md`): one line per reference page.
6. `/reference/std` (new, generated): one page per public std module, with every export's signature and effects, from `narrac`. Descriptions are added when `std-doc-comments` lands.
7. `/reference/commands` (new, `docs/reference/commands.md`): every `narrac` command and flag, with exit codes and environment variables (`CC`, `NARRA_CHECK_LEAKS`, `NARRA_CACHE`).
8. `/reference/narra-toml` (new, `docs/reference/narra-toml.md`): every table and key, its type, default and the error code for a wrong value.
9. `/reference/errors` (new, generated): every diagnostic code with its message and, where one exists, its explanation.
10. `/spec` (`SPEC.md`): listed under Reference, URL unchanged. `D` numbers link to `DECISIONS.md` on GitHub.

Repository only: `README.md` (pitch, install, link to the site, a short "Working on the compiler" section), `DECISIONS.md`, `seed/README.md`, `compiler/README.md`, `website/README.md`, `skills/narra/`.

## Changes

| # | Change | Kind | Reason and evidence | When |
| --- | --- | --- | --- | --- |
| 1 | Chapter 1 installs with the one-line command and links to `/install`. The guide index line for chapter 1 says "install the compiler". | moves | Chapter 1, lines 5 to 15, builds from source; the home page and `/install` show the one-liner (contract `install-command`). | before release |
| 2 | Drop "Workspaces of several packages" from the guide's "Planned" list. | dropped | Chapter 8 teaches workspaces (line 361). | before release |
| 3 | Rewrite the three comparisons with TypeScript and its runtime language in chapters 2 and 11 without the comparison. | changed | Guide rule: no assumed TypeScript outside chapter 13. Lines 02:66, 11:360, 11:432. | before release |
| 4 | Bootstrap stages, exit codes and `linux-check.sh` move from `docs/install.md` to `seed/README.md`. `docs/install.md` keeps the commands. | moves | `seed/README.md` already explains the stages; the user only needs the commands. | before release |
| 5 | `SPEC.md`: `narra` becomes `narrac` in 25 lines; section 15 drops "(sketch)". | changed | The command is `narrac`; section 15 holds full rules with error codes. | before release |
| 6 | `README.md`: replace "draft design" and "Status: specification draft" with the release status; put the install command and the site first; move `bootstrap.sh` and `test.sh` to a "Working on the compiler" section. | changed | The README is the first page on GitHub; after the tag it should lead with install. | before release |
| 7 | Home page: link the agent claim (19 of 20 against 15 of 20) to a committed result under `skills/narra/evals/`. | new | Site rule: claims carry evidence. Today the numbers are only in a roadmap outcome. | before release |
| 8 | Prerelease paragraph of `docs/install.md`: keep if the first tag is a prerelease, cut to one line if it is a full release. | changed | Depends on node `first-release`. | with the tag |
| 9 | Reference section: `/reference` index, header link, routes in `website/src/pages.ts`, search sources in `website/plugins/search.ts`, `docs/reference` in the `ignoreCommand` of `website/vercel.json`. | new | No place on the site holds lookup pages. | after release |
| 10 | `narrac query interface` answers for public std modules, builtin and source. | new (code) | Measured: `unknown module "std/iterator"`; 18 builtin modules have no source. | after release |
| 11 | `/reference/std`, generated by a script over `narrac query --json interface std/<module>`, committed under `docs/reference/std/`, checked by a test suite that fails when a page is stale. | generated | No std function list exists for builtin modules. Committed, not built on Vercel, because `website/scripts/vercel-build.sh` builds the site without `narrac` when `cc` is missing, and GitHub readers see the files too. | after release |
| 12 | One-line descriptions for every public std export: doc comments in `std/src/`, a description field in the builtin registry. The reference and LSP hover show them. | new | `std/src/*.nra` exports have no doc comments; builtins have no description. | after release |
| 13 | `/reference/commands`, written by hand. A test checks that every command and flag in `narrac --help` appears on the page. One complete usage text in `narrac` covers `fmt`, `query`, `lsp`, `apply-fix` and `emit`; bare `narrac` prints it instead of the dump commands. Chapter 12's table becomes a link. | new | Usage text misses 5 commands; bare `narrac` shows internal commands; `-emit-c` is in no guide page. | after release |
| 14 | `/reference/narra-toml`, written by hand, gathered from chapters 1, 8, 12 and `SPEC.md` sections 2 and 15. | new | Tables spread over 5 places; `[link]` is in the spec only. | after release |
| 15 | `/reference/errors`, generated from a new `narrac explain --list --json`, committed, checked like change 11. | generated | 479 codes, 96 with explanations, no list command. | after release |
| 16 | `D` numbers in `SPEC.md` render as links to `DECISIONS.md` on GitHub (`website/plugins/markdown.ts`). | changed | 99 `D` numbers on `/spec` point at a file the site does not publish. | after release |
| 17 | A guide chapter on async, threads and channels. | new | Listed as built but not taught in the guide index. | after release |

## Follow-up nodes

Not briefed. Sizes are estimates.

| id | Goal | Size | When |
| --- | --- | --- | --- |
| `docs-release-pass` | Changes 1 to 7: chapter 1 install, guide index, TypeScript mentions, install page trim, `SPEC.md` command name and section 15 title, README status, home evidence link. | M | before release |
| `site-reference-section` | Change 9: `/reference` index, header link, routes, search and Vercel paths. | S | after release |
| `std-interface-query` | Change 10: `narrac query interface` for public std modules, builtin and source. | M | after release |
| `std-reference` | Change 11: generated, committed std pages and a freshness test. Depends on `std-interface-query` and `site-reference-section`. | M | after release |
| `std-doc-comments` | Change 12: one-line descriptions for every public std export, shown by the reference and hover. | L | after release |
| `cli-reference` | Change 13: complete usage text, `/reference/commands`, the coverage test. | M | after release |
| `manifest-reference` | Change 14: `/reference/narra-toml`. | S | after release |
| `error-index` | Change 15: `narrac explain --list --json` and the generated page. | M | after release |
| `spec-decision-links` | Change 16: `D` numbers on `/spec` link to `DECISIONS.md`. | S | after release |
| `guide-concurrency` | Change 17: a chapter on async, threads and channels. | L | after release |

Change 8 belongs to node `first-release`. The `LICENSE` file is a maintainer decision, not a node.
