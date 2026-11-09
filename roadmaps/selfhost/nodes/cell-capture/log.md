# Log

- Cell is a source module over Box with one primitive (CellSlot.store); struct values are inline in C so a struct alone cannot share.
- Generic call hint for a bare None added in closures.nra (hintNone).
- New leaks suite (selfhost/tests/src/leaks.nra, testdata/leaks) for the expected-leak sample.
- E0612 text now suggests a Cell. E0965 and E0966 unused.
