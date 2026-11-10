# Roadmap: selfhost

This folder is a parallel roadmap. `GOAL.md` has the outcome. Each node is a folder under `nodes/`.

Rules for agents working on a node:

- Work only on your own node. Edit files inside your brief's owned paths and your own `nodes/<id>/` folder.
- Do not edit other briefs, `status` fields, contracts or `GOAL.md`. The orchestrator owns them.
- If a contract is wrong, write `contracts/<name>.change-request.md` and adapt locally. Do not edit the contract.
- Use only the numbers reserved in your brief for migrations, decision records and other numbered files.
- Keep notes in your `log.md`. Finish by writing `outcome.md` from the template.
- Read the outcomes of your direct dependencies. Do not read other nodes' logs.

Script: `python3 <skill>/scripts/plan.py <root> context <id>` lists the files to load for a node.

Project rules:

- Run `sh scripts/test.sh` before finishing (1 to 3 minutes; it builds narrac from `seed/narrac.c` with `cc` only). Every sample program must stay clean under AddressSanitizer and the leak check. While working, run only the suites you touch: `NARRAC=<built narrac> sh scripts/test.sh --suite <suite>`.
- Format edited `.nra` files with `narrac fmt` (see `selfhost/CLAUDE.md`); the `sources` suite checks it.
- Seed rule: `selfhost/src` may use only features the committed seed compiles. Do not regenerate the seed unless your brief says so; the orchestrator does it once per batch.
- Follow `SPEC.md`. Record any difference in `compiler/README.md`.
- New AST node kinds need a case in the checker, emitter, `fmt` and `astdump` switches.
- Narra code in `selfhost/` reads like the maintainer's JS/TS: full descriptive names for locals, parameters and fields (`parser`, `checker`, `expression`, `leftType`, `index`, never `p`, `c`, `e`, `lt`, `idx`; only `i`/`j` over plain index ranges), function declarations named for their action (`parseBinary`, `findLocal`, `collectIds`, never `<noun>Of`), for-of loops where Narra allows, and short comments on the steps inside a function. Shadowing is an error in Narra, so a binding such as `is Some(block)` needs a name of its own.
- Narra is not released, so removed or replaced syntax needs no deprecation path: no migration errors, no automatic fixes and no compatibility shims for old forms (maintainer, 2026-11-10). Remove the old form, update the code that used it, and let the parser report it like any other unknown syntax.
