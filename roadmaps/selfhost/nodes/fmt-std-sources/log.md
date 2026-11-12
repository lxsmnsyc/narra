# Log: fmt-std-sources

- Base feat/batch-10 (ececc85a), compiler built with `sh selfhost/bootstrap.sh`.
- `narrac fmt --check std/src` flags `std/src/coroutine.nra` and `std/src/option.nra` (exit 1). Reproduced.
- Chain sample `const mid = x + { const a = 1; a } + y` printed `x +` level but the block and `y` one indent deeper. Reproduced.
- oxfmt not installed. Reasoned from its rules: an assignment whose right side is a binary chain does not indent the chain's operands, so all operands are level (the existing golden expr_block_operand_level shows this for a trailing block).
- Fix: `endsWithBlock` walks the same-precedence chain and answers true when any operand is a plain block. Golden `expr_block_operand_middle` covers a middle and a leading block.
- Sources suite lists `std/src`; std files formatted; stdtext.nra rewritten with `--suite generated --update`.
- Suites: sources 1, fmt 86, generated 17 pass. Full `sh scripts/test.sh`: PASS, 7111 cases.
