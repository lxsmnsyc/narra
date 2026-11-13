#!/bin/sh
# Builds and runs io_wait_test.c under AddressSanitizer and UBSan.
set -eu
here=$(cd "$(dirname "$0")" && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
"${CC:-cc}" -std=c11 -g -fsanitize=address,undefined -fno-sanitize-recover=undefined -w \
  -o "$work/io_wait_test" "$here/io_wait_test.c"
"$work/io_wait_test"
