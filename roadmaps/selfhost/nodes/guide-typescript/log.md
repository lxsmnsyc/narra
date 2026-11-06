# Log: guide-typescript

- 2026-11-07: started from `origin/feat/language-guide` (d8eb59f). Chapter 11 did not exist; README already had the index line with "optional reading".
- Collected the two "For chapter 11" lists and the 21 `ts_*` files (13 diagnostics, 8 errors). Built `narra` and ran `narra check` on each diagnostics file to get its message, then probed other habits (`;`, `===`, `++`, `??`, `?.`, overloads, `export default`, relative imports, method calls on trait values, structural typing, the counter closure).
- Wrote the chapter in 18 topic sections. First TestGuide run: 13 error samples had wrong columns, and the counter fix returned a `mutable` parameter (not allowed). Fixed both.
- Added the "Next:" line to chapter 10, which guide-advanced left for this node.
- check_prose.py: fixed "JavaScript" (rule says name TypeScript or leave it out), long sentences and one hedge. Now clean.
- TestGuide over all chapters: pass (55 samples in chapter 11).
