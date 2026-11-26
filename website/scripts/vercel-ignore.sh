#!/bin/sh
# Vercel's ignored build step: exit 0 skips the deploy, exit 1 builds. The site builds when a file
# it reads changed since the last deploy, or when the last deployed commit is not in the clone.
previous=${VERCEL_GIT_PREVIOUS_SHA:-HEAD^}
if ! git cat-file -e "$previous^{commit}" 2>/dev/null; then
  exit 1
fi
git diff --quiet "$previous" HEAD -- :/website :/docs/guide :/docs/reference :/docs/packages :/docs/install.md :/SPEC.md ':(top,glob)packages/*/README.md' :/editors/textmate :/seed :/std/src
