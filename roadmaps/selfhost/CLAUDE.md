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

- Run `cd compiler && go vet ./... && go test ./...` before finishing. Every sample program must stay clean under AddressSanitizer and the leak check.
- Follow `SPEC.md`. Record any difference in `compiler/README.md`.
- New AST node types need a case in `clone.go` and in the checker and emitter switches.
- Narra code in `selfhost/` reads like the maintainer's JS/TS: full descriptive names for locals, parameters and fields (`parser`, `checker`, `expression`, `leftType`, `index`, never `p`, `c`, `e`, `lt`, `idx`; only `i`/`j` over plain index ranges), function declarations named for their action (`parseBinary`, `findLocal`, `collectIds`, never `<noun>Of`), for-of loops where Narra allows, and short comments on the steps inside a function. Shadowing is an error in Narra, so a binding such as `is Some(block)` needs a name of its own.
