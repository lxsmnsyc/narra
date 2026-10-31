# Log: selfhost-check-closures

- Layout: closures.nra above exprs (Walker.checkArrow), funcvalues.nra below exprs for useLocal, markArgument and function values.
- Hint needs earlier argument types, so checkArguments was split into checkArgument and the generic call drives its own loop.
- Dump: Go prints stack and captures on arrow lines; Narra prints them from BodyInfo.closures.
- Constants cannot hold arrows, so module tests use `export const triple = double`.
- TestSelfhostBodies 271 compared, 101 skipped; TestSelfhostClosures 17 and 2.
