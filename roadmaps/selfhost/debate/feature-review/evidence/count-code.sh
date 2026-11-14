#!/bin/sh
# Counts uses in code only: comment lines are dropped and string literals
# are replaced by S before matching, so diagnostic texts that name a
# feature do not count. Run from the repo root:
#   sh roadmaps/selfhost/debate/feature-review/evidence/count-code.sh
strip() {
  grep -rhE --include='*.nra' -e '.' "$1" | grep -vE '^[[:space:]]*//' | sed -E 's/"([^"\\]|\\.)*"/S/g; s/`[^`]*`/T/g'
}
f() {
  label="$1"
  re="$2"
  s=$(strip selfhost/src | grep -cE -e "$re")
  d=$(strip std/src | grep -cE -e "$re")
  t=$(strip testdata | grep -cE -e "$re")
  printf '%-26s selfhost=%-6s std=%-5s testdata=%s\n' "$label" "$s" "$d" "$t"
}
f "lines" '.'
f "ternary ?:" ' \? [^:]+ : '
f "labeled break/continue" '(break|continue) [a-z][a-zA-Z_]*$'
f "loop label" '^[[:space:]]*[a-z][a-zA-Z_]*: (for|while) '
f "c-style for" 'for \(let '
f "for (;;)" 'for \(;;\)'
f "Ranges.upTo/between" 'Ranges\.(upTo|between)\('
f "x++ / x--" '[a-zA-Z_\])](\+\+|--)$'
f "tuple #[" '#\['
f "or-pattern line" '[A-Za-z)_] \| [A-Z_a-z"0-9]'
f "match guard" '\) if \(.*=>'
f "spread { ...x" '\{ \.\.\.[a-z]'
f "if as value" '= if \('
f "block expr = {" '(= |return )\{$'
f "match as value" '(= |return )match \('
f "type alias" '^(export )?type [A-Z]'
f "typeof" 'typeof '
f "effect variable" 'effect [A-Z]'
f "from marker" '\) from [a-z]|> from [a-z]|\] from [a-z]'
f "for (mutable" 'for \(mutable'
f "while (true)" 'while \(true\)'
f "hex literal" '0x[0-9a-fA-F]'
f "Cells." 'Cells\.'
f "array pattern" '(is |^[[:space:]]*)\[[^]]*(\.\.\.|\] =>)'
f "named argument" '[(,] [a-z][A-Za-z]*: [a-zA-Z0-9S(\[-]'
f "Options./Results." '(Options|Results)\.[a-z]'
f "is None) {" 'is None\) \{'
f "is Err(" 'is Err\('
f "is Some(" 'is Some\('
f ".value" '\.value([^A-Za-z(]|$)'
f "const _ =" 'const _ = '
f "name: name field" '([a-z][A-Za-z0-9]*): \1(,|$| \})'
f "Arrays.length" 'Arrays\.length\('
f "Str.length" 'Str\.length\('
f "template string" 'T'
f "Str.concat" 'Str\.concat\('
f "copy(" '(^|[^A-Za-z_.])copy\('
f "mutable arg" '\(mutable |, mutable '
f "uses decl" '\) *:? *[^=]* uses [a-z]'
f "escaping" 'escaping [a-z]'
f "@inline" '^@inline'
f "@const" '^@const'
f "requires/ensures" '^[[:space:]]+(requires|ensures) '
f "distinct type" 'distinct type'
f "opaque struct" 'opaque struct'
f "Dynamic" 'Dynamic[<.]'
f "Coroutine" 'Coroutine<'
f "function*" 'function\*'
f "async fn" 'async function|async \(|async \{'
f "lazy" '= lazy '
f "derives" 'derives '
f "implement block" '^implement '
f "trait decl" '^(export )?trait '
f "Box" 'Box<|Boxes\.'
f "FixedArray" 'FixedArray<'
f "sized ints" '(: |<)(i8|i16|i32|u8|u16|u32|u64)[^a-z0-9]'
f "bytes" '(: |<)bytes[^a-z]'
