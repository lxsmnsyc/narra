#!/bin/sh
# Runs the Autobahn fuzzing client in Docker against the echo example, for cases 1 to 10 (framing,
# pings, reserved bits, opcodes, fragmentation, UTF-8, close handling, limits, compression
# off). It is not part of scripts/test.sh. It needs Docker, an internet connection for the first
# pull of the image, and a built compiler (NARRAC, default out/narrac).
#
#   sh packages/websocket/autobahn.sh [case patterns]
#
# With no argument it runs 1.* to 10.*. Pass patterns such as "1.*" "6.1.1" to run fewer cases.
# The report is kept in the folder named by AUTOBAHN_REPORTS when that is set.
set -eu

root=$(cd "$(dirname "$0")/../.." && pwd)
narrac=${NARRAC:-$root/out/narrac}
port=${AUTOBAHN_PORT:-9001}
image=${AUTOBAHN_IMAGE:-crossbario/autobahn-testsuite}
# The work folder is inside the repository because a Docker VM shares only the home folder.
work=$(mktemp -d "$root/packages/websocket/.autobahn-work-XXXXXX")
container="narra-autobahn-$$"
server_pid=""

cleanup() {
  if [ -n "$server_pid" ]; then
    kill "$server_pid" 2> /dev/null || true
  fi
  docker rm -f "$container" > /dev/null 2>&1 || true
  # The reports may be owned by root, which the container writes as; leave them then.
  rm -rf "$work" 2> /dev/null || true
}
trap cleanup EXIT INT TERM

if ! docker info > /dev/null 2>&1; then
  echo "autobahn.sh: Docker is not available" >&2
  exit 2
fi

"$narrac" build -o "$work/echo" "$root/examples/websocket_echo"
"$work/echo" 0.0.0.0 "$port" > "$work/server.log" 2>&1 &
server_pid=$!

cases=""
if [ "$#" -eq 0 ]; then
  set -- "1.*" "2.*" "3.*" "4.*" "5.*" "6.*" "7.*" "8.*" "9.*" "10.*"
fi
for pattern in "$@"; do
  if [ -n "$cases" ]; then
    cases="$cases, "
  fi
  cases="$cases\"$pattern\""
done

mkdir -p "$work/config" "$work/reports"
cat > "$work/config/fuzzingclient.json" << EOF
{
  "outdir": "/reports",
  "servers": [{ "agent": "narra", "url": "ws://host.docker.internal:$port" }],
  "cases": [$cases],
  "exclude-cases": [],
  "exclude-agent-cases": {}
}
EOF

docker run --name "$container" \
  --add-host host.docker.internal:host-gateway \
  -v "$work/config:/config" -v "$work/reports:/reports" \
  "$image" wstest -m fuzzingclient -s /config/fuzzingclient.json

# Count the verdicts. OK and NON-STRICT pass, INFORMATIONAL is not a verdict, FAILED does not.
status=0
python3 -I - "$work/reports/index.json" << 'EOF' || status=$?
import json, sys
with open(sys.argv[1]) as handle:
    results = json.load(handle)["narra"]
counts = {}
failed = []
noted = []
for case, entry in sorted(results.items()):
    verdict = entry["behavior"]
    counts[verdict] = counts.get(verdict, 0) + 1
    if verdict not in ("OK", "NON-STRICT", "INFORMATIONAL"):
        failed.append(case + " " + verdict)
    elif verdict != "OK":
        noted.append(case + " " + verdict)
print("autobahn: " + ", ".join(k + " " + str(v) for k, v in sorted(counts.items())))
for line in noted:
    print("autobahn: " + line)
for line in failed:
    print("autobahn: " + line)
sys.exit(1 if failed else 0)
EOF

if [ -n "${AUTOBAHN_REPORTS:-}" ]; then
  mkdir -p "$AUTOBAHN_REPORTS"
  cp -R "$work/reports/." "$AUTOBAHN_REPORTS/"
fi
exit "$status"
