#!/bin/sh
# Check the eval tasks themselves, with no model call.
# Usage: sh verify-tasks.sh NARRAC
# For each task, the starter must fail grading and the starter with its solution must pass.
# Exit code 0 when every task behaves so, 1 otherwise.
set -u
if [ "$#" -ne 1 ]; then
  echo "usage: sh verify-tasks.sh NARRAC" >&2
  exit 2
fi
narrac=$1
here=$(cd "$(dirname "$0")" && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
bad=0
count=0
for task in "$here"/tasks/*/; do
  name=$(basename "$task")
  count=$((count + 1))
  rm -rf "$work/starter" "$work/solved"
  cp -R "$task/starter" "$work/starter"
  cp -R "$task/starter" "$work/solved"
  cp -R "$task/solution/." "$work/solved/"
  if sh "$here/grade.sh" "$narrac" "$task" "$work/starter" > /dev/null 2>&1; then
    echo "FAIL $name: the starter passes"
    bad=$((bad + 1))
  fi
  if ! sh "$here/grade.sh" "$narrac" "$task" "$work/solved" > "$work/out" 2>&1; then
    echo "FAIL $name: the solution does not pass"
    cat "$work/out"
    bad=$((bad + 1))
  fi
done
echo "$count tasks, $bad problems"
[ "$bad" -eq 0 ]
