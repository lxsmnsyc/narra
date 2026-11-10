# Log: map-helpers-split

- Base: feat/batch-5 at 43aae732. Compilers built with `sh selfhost/bootstrap.sh <dir>`: base from a `git archive` of that commit, branch from the working tree.
- Command: `/usr/bin/time -l <narrac> emit selfhost`, instructions retired, 8 alternating base/branch pairs. Load average about 14 (shared, loaded machine).
- Base: 1939610200, 1950378475, 1924188265, 1918325807, 1931757027, 1940207884, 1945890828, 1939840292 (spread about 1.7%).
- Branch: 1839193728, 1820992103, 1820472624, 1822567395, 1840773808, 1841204971, 1826996357, 1839016150.
- Branch won 8 of 8 pairs; mean 1.936 G against 1.831 G, 5.4% fewer.
- Cause: `mapFunctions` ran 12 `Str.replace` calls over the 4 KB template for each map type; each call scans and copies the whole text. Now the template is split on `@` once per `emitMapFunctions` call and each map type fills the 12 slots in one pass.
- `emit selfhost` is byte-identical on base and branch; `emit` of 61 testdata samples that use `Map<` is identical too.
