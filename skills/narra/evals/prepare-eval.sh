#!/bin/sh
# Build the plugin copy that the eval runs against, without the solution folders.
# Usage: sh prepare-eval.sh DEST
# Run the eval on DEST: the agent under test cannot read solution/ because DEST has none.
set -eu
if [ "$#" -ne 1 ]; then
  echo "usage: sh prepare-eval.sh DEST" >&2
  exit 2
fi
dest=$1
root=$(cd "$(dirname "$0")/../../.." && pwd)
rm -rf "$dest"
mkdir -p "$dest/skills/narra" "$dest/.claude-plugin"
cp "$root/.claude-plugin/marketplace.json" "$dest/.claude-plugin/"
cp "$root/skills/narra/SKILL.md" "$dest/skills/narra/"
cp -R "$root/skills/narra/evals" "$dest/skills/narra/evals"
rm -rf "$dest/skills/narra/evals/results"
rm -f "$dest"/skills/narra/evals/tasks/*/checks.nra
for solution in "$dest"/skills/narra/evals/tasks/*/solution; do
  rm -rf "$solution"
done
