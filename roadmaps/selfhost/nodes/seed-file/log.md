# Log: seed-file

- Built Go narra, emitted selfhost as C, used it as the first seed with a placeholder header.
- seed-update.sh rewrote it with the real header; second run left it byte-identical, also with go off PATH.
- Bug found in my own sed: header end line is " */" (leading space); fixed in bootstrap.sh.
- Linux: gcc:12 aarch64 passes. linux/amd64 emulated run started in the background.
- Self-review: diff checked against the brief; all five steps present.
