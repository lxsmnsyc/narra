# Log: ptr-cast

- Reset the branch to feat/batch-9 and renamed it feat/ptr-cast.
- Nothing existed yet: `std/ptr` had no `cast` or `element` rows (stdptr.nra POINTER_FUNCS).
- Added `Ptr.cast` (result element type from the expected type, like `Ptr.alloc`) and `Ptr.element(pointer: Ptr<FixedArray<T, N>>, index)` in stdptr.nra and emitpointers.nra. The index goes through `checkFixedIndex`, so a constant outside 0 to N - 1 is E0994 and a non-constant index needs the same proof as a FixedArray read. No new diagnostic code; the "takes its element type from where the result goes" message is reused.
- fmt comment proof, on the base compiler logic (fmt/parser are untouched by anything but this fix): `printf 'extern {\n  function abs(value: i32): i32\n} from /* x */ "c"\n' > c.nra; narrac fmt c.nra` printed `} /* x */ from "c"`. The comment moved before `from`. Cause: the comment is trailing of the closing-brace node `r`, the only node before it.
- Fix: the parser notes a node `k` for the `from` keyword and `l` for the library string; the printer wraps each with its comments. A comment before `from` is leading of `k`, one between `from` and the name is leading of `l`.
- First attempt noted only `l`; it moved the comment in `comment_extern_brace` (`} /* c */ from "c"` became `} from /* c */ "c"`), so `k` was added and the golden is unchanged.
- The machine was under load average 50 to 60 from the other workers; the full test run took very long.
- Review fixes: a leading comment of the library name that cannot share the line with it (line comment, or newline after) is printed after the name (moveLineCommentsToTrailing in fmtprinter.nra). The golden comment_extern_from.fmt has three cases and formats to itself. The comments move from l to x, so fmtstable.nra MOVING_COMMENTS lists the two. The trivia.nra key text and the SPEC.md aliasing note were reworded.
