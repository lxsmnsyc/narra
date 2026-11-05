# Log: fmt-doc

- Branch feat/fmt-doc from origin/feat/fmt. Work in selfhost/src/fmtdoc.nra, dispatch line in main.nra.
- `print` is also the prelude function, so inside the module the implementation is `printDocument`; the exported `print` calls it. Callers use `Doc.print`.
- The printer copies the Doc into arrays first (flatten) and marks forced-break groups in one reverse pass, instead of rescanning per group.
- Blanks at line end are held back in the writer (`pending`), indentation is written only before text, so no trailing whitespace.
- Deep documents (50000) checked by hand: indent and group chains print in under half a second.
