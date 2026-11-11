# Starts `narrac run` or `narrac test` on a spinning sample, waits until the
# program wrote its own pid to the marker file, sends SIGTERM to narrac and
# checks that narrac and that process are gone. Needs only sh and kill.
# Usage: sh forward.sh run|test narrac sample marker
mode=$1
narrac=$2
sample=$3
marker=$4
rm -f "$marker"
export NARRA_SIGNAL_MARKER=$marker
export NARRA_TEST_TIMEOUT=600

parent=
program=
cleanup() {
  if [ -n "$parent" ]; then
    kill -TERM "$parent" 2>/dev/null
    sleep 1
  fi
  if [ -n "$program" ]; then
    kill -9 "$program" 2>/dev/null
  elif [ -s "$marker" ]; then
    kill -9 "$(cat "$marker")" 2>/dev/null
  fi
  if [ -n "$parent" ]; then
    kill -9 "$parent" 2>/dev/null
    wait "$parent" 2>/dev/null
  fi
}
trap 'cleanup; exit 1' TERM INT HUP

if [ "$mode" = "test" ]; then
  "$narrac" test "$sample" >/dev/null 2>&1 &
else
  "$narrac" run "$sample" "$marker" >/dev/null 2>&1 &
fi
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
kill -TERM "$parent" 2>/dev/null
waited=0
while kill -0 "$program" 2>/dev/null && [ "$waited" -lt 100 ]; do
  sleep 0.1
  waited=$((waited + 1))
done
wait "$parent" 2>/dev/null
if kill -0 "$program" 2>/dev/null; then
  cleanup
  echo "the program kept running after narrac got SIGTERM"
  exit 1
fi
echo "stopped"
