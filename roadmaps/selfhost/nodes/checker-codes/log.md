# Log: checker-codes

- Table dumped from Go to JSON, rendered as `diagnostics.nra`.
- Codes placed by a codemod that matches each message literal to a Go format; 25 ambiguous or helper sites by hand.
- Found by `compare.py`: E0601 had matched the E0409 catch-all, fixed.
- Self-review: diff read against the brief and the naming rules; no change needed.
