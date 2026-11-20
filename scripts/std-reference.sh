#!/bin/sh
# Write docs/reference/std, made by scripts/std/generate.nra from `narrac query --json interface`.
#   NARRAC=/path/to/narrac sh scripts/std-reference.sh
# NARRAC defaults to narrac on the PATH. The generated suite of selfhost/tests fails when
# the committed pages differ from this output; `sh scripts/test.sh --suite generated --update`
# rewrites them.
set -eu

narrac=${NARRAC:-narrac}
root=$(cd "$(dirname "$0")/.." && pwd)
exec "$narrac" run -release "$root/scripts/std/generate.nra" "$narrac" "$root" "$root/docs/reference/std"
