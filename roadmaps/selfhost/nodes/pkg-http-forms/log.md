# Log: pkg-http-forms

- Checked first: no `readText`, `readJson`, `json`, `redirect`, `isOk` or form code in `packages/http`. The work was new.
- Design: `multipart` (private) walks a body: `Reader` holds the source stream, a buffer in a `Cell`, and flags. The buffer starts with CR LF so the first boundary matches like the others. `nextPart` skips the preamble, or what is left of the last part, and reads the part head. `streamPartBody` is the part body stream. Its end of the buffer, one byte short of the delimiter length, is kept until more arrives, which finds a boundary split across reads.
- `form` holds the public types and `parts`, `readFields`. There is no re-export in Narra, so the types cannot live in `multipart`.
- A part that is not read is drained by `nextPart` through the same body stream with no limit.
- The first draft passed all 241 checks, curl included.
