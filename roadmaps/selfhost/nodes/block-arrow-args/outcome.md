# Outcome: block-arrow-args

## Built

- `compiler/internal/narra/lexer.go` and `selfhost/src/lexer.nra`: on `{` the lexer saves the bracket depth and resets it to 0, on `}` it restores it. Line breaks inside a block nested in `( )` end statements again.
- `compiler/testdata/block_arrow_args.nra` and `.out`: a multi-statement arrow as the only argument, nested in another call, as the last of several arguments, and with a multi-line struct literal inside the block.

## Deviations

- None.

## Stubs left

- None.

## Gotchas

- The reset applies to every `{`, so a multi-line struct or map literal inside `( )` now emits line breaks too. The parser already handles them (the sample covers a struct literal).
- Narra std has no `Arrays.pop` or slice, so the self-hosted lexer keeps a saved-depth array plus a count.

## Proposed learnings

- None.

## Verification

- `cd compiler && go vet ./... && go test ./...`: see the PR (run on the rebased branch).

## Size

- Estimated S, actual S.
