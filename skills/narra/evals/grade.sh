#!/bin/sh
# Grade one finished eval task.
# Usage: sh grade.sh NARRAC TASK_DIR WORK_DIR
#   TASK_DIR is a folder under tasks/. WORK_DIR is the package the agent edited.
# The task's checks.nra is copied into WORK_DIR/src, then `narrac check` and `narrac test` run.
# Exit code 0 when both pass, 1 otherwise.
set -u
if [ "$#" -ne 3 ]; then
  echo "usage: sh grade.sh NARRAC TASK_DIR WORK_DIR" >&2
  exit 2
fi
narrac=$1
task=$2
work=$3
cp "$task/checks.nra" "$work/src/checks.nra"
"$narrac" check "$work" || exit 1
"$narrac" test "$work" || exit 1
