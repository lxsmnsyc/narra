#!/bin/sh
# Print docs/reference/errors.md, made by scripts/errors/generate.nra from `narrac explain --list --json`.
#   NARRAC=/path/to/narrac sh scripts/error-index.sh > docs/reference/errors.md
# NARRAC defaults to narrac on the PATH. The generated suite of selfhost/tests fails when
# the committed page differs from this output; `sh scripts/test.sh --suite generated --update`
# rewrites it.
set -eu

narrac=${NARRAC:-narrac}
here=$(dirname "$0")
exec "$narrac" run -release "$here/errors/generate.nra" "$narrac"
