# Log: selfhost-emit-options

- Ported Option, Result and std/test emitters and the test main into emitoptions.nra.
- `from` and `to` are keywords: parameters renamed.
- Whole-output diff against Go on std_option, std_result, tests, std_test: tests.nra byte-identical; the others differ only by stub markers (std, traits) and the temp/label counters they shift.
