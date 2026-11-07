#!/bin/sh
# Counts call-site shapes that named arguments target, in selfhost/src and std.
cd "$(git rev-parse --show-toplevel)"
files=$(find selfhost/src std -name '*.nra')
echo "files: $(echo $files | wc -w)"
echo "function declarations: $(cat $files | grep -cE '^\s*(export )?function\*? [A-Za-z_]+')"
echo "declarations with >=4 parameters (one-line signatures):"
cat $files | grep -E '^\s*(export )?function\*? [A-Za-z_]+(<[^>]*>)?\(' | awk -F'(' '{ sig=$0; sub(/^[^(]*\(/, "", sig); sub(/\)[^)]*$/, "", sig); n=gsub(/: /, "", sig); if (n>=4) c++ } END { print c+0 }'
echo "call lines with a bare true/false argument: $(cat $files | grep -cE '[(,] ?(true|false) ?[,)]')"
echo "struct literal passed as an argument (Name { ... } after '(' or ', '): $(cat $files | grep -cE '[(,] ?[A-Z][A-Za-z]*(<[^>]*>)? \{')"
echo "declarations whose parameter list spans lines (formatter broke it, likely long): $(cat $files | grep -cE '^\s*(export )?function\*? [A-Za-z_]+(<[^>]*>)?\($')"
echo "sample bare-bool call lines:"
cat $files | grep -E '[(,] ?(true|false) ?[,)]' | grep -v '^\s*//' | head -8
