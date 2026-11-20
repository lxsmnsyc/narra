# Change request: guide-samples

From node: snippet-warnings

A `narra error` fence can start with several `// error: LINE:COLUMN: MESSAGE` lines, one per diagnostic the sample prints. The suite checks that each is printed and that no other diagnostic is (a `: note:` line belongs to the diagnostic before it). The contract says "its first line is a comment with the expected message" and "passes when compiling fails with a message that contains the expected text".

Proposed text for Interface and Behavior: replace "its first line" with "its leading lines", and add "A sample passes only when every diagnostic it prints is named by one of these lines." Applied locally in `selfhost/tests/src/guide.nra` and `website/plugins/twoslash.ts`.
