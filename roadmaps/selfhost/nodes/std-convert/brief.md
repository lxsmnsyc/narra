---
id: std-convert
title: Char, integer and string helpers
status: done
priority: must
type: feature
outline: false
milestone: m8-selfhost
target:
from:
size: M
spike: false
verify:
exclusive: []
fulfilled_by:
gate: false
added: 2026-10-30
spent: 178351
owns:
  - compiler/internal/narra/std_char.go
  - compiler/testdata/convert*
  - compiler/testdata/errors/convert*
shared:
  - compiler/internal/narra/std_int.go
  - compiler/internal/narra/std_string.go
  - compiler/internal/narra/std_string_test.go
  - compiler/internal/narra/runtime.h
  - compiler/internal/narra/runtime_string.h
  - compiler/internal/narra/modules_check.go
  - compiler/internal/narra/check.go
  - compiler/README.md
reserves: []
provides: []
depends_on:
  - on: strings-std
    kind: hard
    why: builds on the Str module, StringBuilder and the string runtime
---

# Char, integer and string helpers

## Goal

Add the conversions and string helpers the self-hosted compiler needs, which the selfhost-spike found missing (outcome gaps 1, 2 and 8, plus missing string helpers), and fix the NUL bug in the C runtime. Names were approved by the user on 2026-10-30 after comparing Rust, Go, Kotlin, JS and Swift.

## Context

New builtins (register each with its `Effects` per the builtins contract; `alloc` when it builds a string):

- `Char.code(c: char): int`: the Unicode scalar value. Pure.
- `Char.fromCode(n: int): Option<char>`: `None` for surrogates (0xD800-0xDFFF), negatives and values above 0x10FFFF.
- `Int.parse(s: string, base: int): Result<int, string>`: bases 2 to 36, optional leading `-`, no prefix handling (the caller strips `0x`). Errors: empty string, bad digit for the base, overflow of 64-bit `int`, bad base. Error texts name the problem.
- `Int.toHex(n: int): string`: lowercase, no prefix, negative numbers as `-` plus digits.
- `Str.endsWith(s, suffix): bool`, `Str.contains(s, part): bool`: pure.
- `Str.replace(s, from, to): string`: replaces every match, left to right, non-overlapping. An empty `from` returns `s` unchanged (record it in README).
- `Str.join(parts: Array<string>, sep: string): string`.

Bug fix: `lstr_from_char` in `runtime.h` ends with `lstr_from_cstr(buf)`, which uses `strlen`, so `Str.fromChar('\0')` and `Str.appendChar(sb, '\0')` give an empty string. Build the string from the encoded length instead, and check `appendChar` takes the same path.

- `Char` is a new std module (`std/char`?) or a prelude namespace: follow how `Int` and `Float` are exposed today (`std_int.go`), and record the choice in the outcome.
- C names go through `cnames.go`; see `compiler/internal/narra/CLAUDE.md`.
- The self-hosted lexer (`selfhost/src/lexer.nra`, from PR #16) has stubs for these (`charCode`, `charString`, `pushChar`, `\u` escapes above 0x7F). Leave them; selfhost-front replaces them.
- Diagnostic codes: none expected. If a new compile error is needed, take the next free code in the std range from the json-diagnostics outcome and reserve it here.

## Acceptance

- [x] `testdata/convert.nra` covers every new builtin, including `Char.fromCode` on a surrogate and above 0x10FFFF, `Int.parse` overflow and bad digits in bases 2, 10, 16 and 36, `Int.toHex` on 0, negative and the largest int, `Str.replace` with overlapping candidates, and `Str.join` on an empty array. Clean under the sanitizers and the leak check.
- [x] `Str.fromChar('\0')` has length 1, and a StringBuilder with an appended NUL keeps it (test in `testdata/convert.nra` or `std_string_test.go`).
- [x] Fast check passes: `cd compiler && go vet ./... && go test ./...`

## Sizing

M: eight small builtins with runtime C, one runtime bug fix, many edge cases to test.

PR: [#17](https://github.com/lxsmnsyc/narra/pull/17)
