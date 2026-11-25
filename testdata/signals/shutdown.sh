# Runs a program that waits for signals with std/signal, sends it the signals in order with kill, half
# a second apart, and checks that it printed the lines of the expected file and exited 0. Needs only
# sh and kill.
# Usage: sh shutdown.sh narrac sample marker output expected signal...
narrac=$1
sample=$2
marker=$3
output=$4
expected=$5
shift 5
rm -f "$marker" "$output"
export NARRA_SIGNAL_MARKER=$marker

parent=
cleanup() {
  if [ -s "$marker" ]; then
    kill -9 "$(cat "$marker")" 2>/dev/null
  fi
  if [ -n "$parent" ]; then
    kill -9 "$parent" 2>/dev/null
    wait "$parent" 2>/dev/null
  fi
}
trap 'cleanup; exit 1' TERM INT HUP

# The program builds first; its own pid comes through the marker.
"$narrac" run -sanitize "$sample" >"$output" 2>&1 &
parent=$!
waited=0
while [ ! -s "$marker" ] && [ "$waited" -lt 3000 ]; do
  sleep 0.1
  waited=$((waited + 1))
done
if [ ! -s "$marker" ]; then
  cleanup
  echo "the program did not start"
  exit 1
fi
program=$(cat "$marker")
# The first wait started before the pid was written; let the call that wrote it finish.
sleep 0.5
for name in "$@"; do
  kill -"$name" "$program"
  sleep 0.5
done
wait "$parent"
code=$?
if [ "$code" -ne 0 ]; then
  cleanup
  echo "exit code $code, output:"
  cat "$output"
  exit 1
fi
want=$(cat "$expected")
got=$(cat "$output")
if [ "$got" != "$want" ]; then
  echo "unexpected output:"
  echo "$got"
  exit 1
fi
echo "stopped"
