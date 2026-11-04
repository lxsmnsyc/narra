# Log: labels

- No prior work existed. Go compiler first, then selfhost.
- Label is stored on the loop (Pos = label position) and on break/continue. Errors: unknown label at the break/continue, reuse at the inner labeled loop, non-loop label at the label (parser error, since the label token is consumed by the parser).
- Sample `labels.nra` hand-checked: 2441 and 26 verified by hand.
