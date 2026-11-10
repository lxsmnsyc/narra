# Log

- Base ffdea9c, `narrac emit selfhost` on the same source, 3 runs each, `/usr/bin/time -l`.
- Base instructions retired: 9339540802, 9326299085, 9341217246. Branch: 9058356023, 9068248883, 9070856956.
- Peak memory: base 130.4 to 131.9 MB, branch 126.7 to 134.5 MB (within noise).
- Emitted C of the base and branch compilers is byte-identical; the bootstrap fixed point holds.
