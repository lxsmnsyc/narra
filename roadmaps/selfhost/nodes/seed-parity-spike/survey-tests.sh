#!/bin/sh
# Throwaway: print each Go test with its doc comment, and the helpers it uses.
cd "$(dirname "$0")/../../../../compiler/internal/narra"
for f in *_test.go; do
  echo "##### $f ($(wc -l < "$f") lines)"
  awk '/^\/\//{c=c $0 "\n"; next} /^func Test/{printf "%s%s\n", c, $0} {c=""}' "$f"
  echo "  helpers: $(grep -o 'compareSelfhost\|buildStage2\|buildSelfhost\|runNarrac[A-Za-z]*\|CompileToC\|runProgram\|Build(' "$f" | sort | uniq -c | tr -s ' ' | tr '\n' ' ')"
  echo "  testdata: $(grep -o 'testdata, "[a-z0-9_]*"' "$f" | sort -u | tr '\n' ' ')"
done
