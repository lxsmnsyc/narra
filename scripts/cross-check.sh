#!/bin/sh
# Build the sample programs for every row of `narrac targets` from one host,
# run each where a runner exists, and compare the output with the .out goldens.
# It also builds seed/narrac.c for every row and, where that runs, checks that
# it emits the same C for testdata/c_interop.nra as the host's seed.
#
#   sh scripts/cross-check.sh [row ...]
#
# Without rows, every row of `narrac targets` is checked. NARRAC is the host
# compiler; when it is unset the script builds one from the seed with
# selfhost/bootstrap.sh. The C compiler is `zig cc` (in CC through a wrapper,
# since CC names one program). Not part of scripts/test.sh: it needs zig, and
# docker, wasmtime or wine for the runs.
#
# Runners, by row:
#   host row           runs directly
#   MacOS-X86_64       Rosetta (`arch -x86_64`) on an arm64 Mac
#   Linux-*            docker run --platform linux/<arch>, alpine for musl, debian for gnu
#   Windows-X86_64     natively on a Windows host (CI), else wine in a linux/amd64 container
#   Wasi-Wasm32        wasmtime -W exceptions=y
# A missing runner is a skip with a message, not a failure.
#
# Environment:
#   NARRAC     host compiler (default: built from the seed)
#   SKIP_SEED  set to 1 to skip the seed builds
#   RUN_TIMEOUT seconds a container run may take (default 300)
#   STRICT     set to 1 to count expected failures as failures
#   REQUIRE_RUN set to 1 to fail a row that had no runner (CI sets it)
#
# Exit status: 0 when every build and run that is not an expected failure
# matched its golden; 1 otherwise.
set -u

root=$(cd "$(dirname "$0")/.." && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# Programs that run on every row. Each has a golden testdata/<name>.out.
samples="hello arrays closures enums maps int_to_string float_shortest c_interop windows_runtime windows_path windows_process windows_async"

# Rows whose runtime work has not landed. A failure on one is reported as an
# expected failure.
expected_failures=""

# Samples a row leaves out, one "row:sample" pair per word: c_interop calls libc
# through a bitcast, windows_runtime uses std/thread, windows_process uses Process.run
# and windows_async uses std/net, which wasm32-wasi lacks.
left_out="Wasi-Wasm32:c_interop Wasi-Wasm32:windows_runtime Wasi-Wasm32:windows_process Wasi-Wasm32:windows_async"

run_timeout=${RUN_TIMEOUT:-300}

say() {
  printf '%s\n' "$*"
}

# --- tools -------------------------------------------------------------------

if ! command -v zig >/dev/null 2>&1; then
  say "cross-check: zig is not installed; it is the C compiler for every row" >&2
  exit 2
fi
case "$(uname -s)" in
  MINGW* | MSYS* | CYGWIN*)
    # The driver starts the compiler through cmd.exe, which runs .cmd files.
    printf '@zig cc %%*\r\n' > "$work/zigcc.cmd"
    CC=$(cygpath -w "$work/zigcc.cmd")
    ;;
  *)
    cat > "$work/zigcc" <<'EOF'
#!/bin/sh
exec zig cc "$@"
EOF
    chmod +x "$work/zigcc"
    CC="$work/zigcc"
    ;;
esac
export CC

have_docker=0
if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
  have_docker=1
fi
have_wasmtime=0
if command -v wasmtime >/dev/null 2>&1; then
  have_wasmtime=1
fi

host_os=$(uname -s)
host_machine=$(uname -m)
case "$host_os" in
  Darwin) host_row_os=MacOS ;;
  Linux) host_row_os=Linux ;;
  MINGW* | MSYS* | CYGWIN*) host_row_os=Windows ;;
  *) host_row_os=$host_os ;;
esac
case "$host_machine" in
  arm64 | aarch64) host_row_arch=Arm64 ;;
  x86_64 | AMD64 | amd64) host_row_arch=X86_64 ;;
  *) host_row_arch=$host_machine ;;
esac
host_row=$host_row_os-$host_row_arch

if [ -z "${NARRAC:-}" ]; then
  say "cross-check: NARRAC is unset; building narrac from the seed"
  sh "$root/selfhost/bootstrap.sh" "$work/compiler" > "$work/bootstrap.log" 2>&1 || {
    cat "$work/bootstrap.log" >&2
    say "cross-check: the bootstrap failed" >&2
    exit 2
  }
  narrac=$work/compiler/narrac
else
  narrac=$NARRAC
fi

# --- the row list --------------------------------------------------------------

"$narrac" targets > "$work/targets.txt" || {
  say "cross-check: narrac targets failed" >&2
  exit 2
}

rows=""
if [ "$#" -gt 0 ]; then
  rows="$*"
else
  while read -r name triple; do
    rows="$rows $name"
  done < "$work/targets.txt"
fi

findTriple() {
  while read -r name triple; do
    if [ "$name" = "$1" ]; then
      printf '%s' "$triple"
      return 0
    fi
  done < "$work/targets.txt"
  return 1
}

matchExpectedFailure() {
  case " $expected_failures " in
    *" $1 "*) return 0 ;;
  esac
  return 1
}

matchLeftOut() {
  case " $left_out " in
    *" $1:$2 "*) return 0 ;;
  esac
  return 1
}

# --- runners -------------------------------------------------------------------

# runner_kind prints how a row runs: native, rosetta, docker, wine, wasmtime or none.
runner_kind() {
  case "$1" in
    "$host_row") say native ;;
    MacOS-X86_64)
      if [ "$host_row_os" = MacOS ] && arch -x86_64 /usr/bin/true >/dev/null 2>&1; then
        say rosetta
      else
        say none
      fi
      ;;
    Linux-*)
      # The image must start; a half-pulled image in the local store does not.
      if [ "$have_docker" = 1 ] && docker run --rm --platform "$(dockerplatform "$1")" "$(dockerimage "$1")" true >/dev/null 2>&1; then
        say docker
      else
        say none
      fi
      ;;
    Windows-*)
      if [ "$host_row_os" = Windows ]; then
        say native
      elif [ "$have_docker" = 1 ]; then
        say wine
      else
        say none
      fi
      ;;
    Wasi-*)
      if [ "$have_wasmtime" = 1 ]; then say wasmtime; else say none; fi
      ;;
    *) say none ;;
  esac
}

# runner_note prints why a row has no runner.
runner_note() {
  case "$1" in
    MacOS-X86_64) say "no Rosetta" ;;
    Linux-*) say "docker is not running or the image does not start" ;;
    Windows-*) say "docker is not running" ;;
    Wasi-*) say "wasmtime is not installed" ;;
    *) say "no runner for this row" ;;
  esac
}

dockerplatform() {
  case "$1" in
    *-Arm64*) say linux/arm64 ;;
    *) say linux/amd64 ;;
  esac
}

dockerimage() {
  case "$1" in
    *-musl) say alpine ;;
    *) say debian ;;
  esac
}

wine_image=narra-cross-wine
wine_ready=unknown

# preparewine builds the wine image once; it returns 1 when that fails.
preparewine() {
  if [ "$wine_ready" = yes ]; then return 0; fi
  if [ "$wine_ready" = no ]; then return 1; fi
  if docker image inspect "$wine_image" >/dev/null 2>&1; then
    wine_ready=yes
    return 0
  fi
  say "cross-check: building the wine image (linux/amd64, once)"
  mkdir -p "$work/wine"
  cat > "$work/wine/Dockerfile" <<'EOF'
FROM debian:bookworm-slim
RUN dpkg --add-architecture i386 \
  && apt-get update \
  && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends wine wine64 \
  && rm -rf /var/lib/apt/lists/*
ENV WINEDEBUG=-all WINEARCH=win64 WINEPREFIX=/wine
RUN wine wineboot --init || true
EOF
  if docker build -q --platform linux/amd64 -t "$wine_image" "$work/wine" > "$work/wine.log" 2>&1; then
    wine_ready=yes
    return 0
  fi
  wine_ready=no
  return 1
}

# runbinary kind row binary args... runs a built program with its stdin closed
# and prints its stdout. A run past run_timeout seconds is killed.
runbinary() {
  kind=$1
  row=$2
  binary=$3
  shift 3
  case "$kind" in
    native) withtimeout "$binary" "$@" < /dev/null ;;
    rosetta) withtimeout arch -x86_64 "$binary" "$@" < /dev/null ;;
    wasmtime) withtimeout wasmtime -W exceptions=y --dir "$root" --dir "$work" "$binary" "$@" < /dev/null ;;
    docker)
      withtimeout docker run --rm -i -e TMPDIR=/tmp --platform "$(dockerplatform "$row")" \
        -v "$root:/repo:ro" "$(dockerimage "$row")" \
        sh -c 'cat > /program && chmod +x /program && cd /repo && /program "$@"' sh "$@" < "$binary"
      ;;
    wine)
      # The prefix has no PATH, and Process.run searches PATH only, so WINEPATH names the system folder.
      withtimeout docker run --rm -i -e TMPDIR=/tmp -e 'WINEPATH=C:\windows\system32' --platform linux/amd64 -v "$root:/repo:ro" "$wine_image" \
        sh -c 'cat > /program.exe && cd /repo && wine /program.exe "$@"' sh "$@" < "$binary"
      ;;
  esac
}

# withtimeout runs a command and kills it after run_timeout seconds.
withtimeout() {
  # A background job reads /dev/null unless its stdin is passed on by hand.
  exec 3<&0
  "$@" <&3 &
  command_pid=$!
  (
    sleep "$run_timeout"
    kill "$command_pid" 2>/dev/null
  ) > /dev/null 2>&1 &
  watcher_pid=$!
  wait "$command_pid"
  command_status=$?
  exec 3<&-
  kill "$watcher_pid" 2>/dev/null
  wait "$watcher_pid" 2>/dev/null
  return "$command_status"
}

# --- the checks ----------------------------------------------------------------

results="$work/results.txt"
: > "$results"

# record row what status detail
record() {
  printf '%s\t%s\t%s\t%s\n' "$1" "$2" "$3" "$4" >> "$results"
}

checkrow() {
  row=$1
  triple=$(findTriple "$row") || {
    say "cross-check: unknown row $row (narrac targets lists the rows)" >&2
    record "$row" all fail "unknown row"
    return
  }
  kind=$(runner_kind "$row")
  case "$row" in
    Windows-*) suffix=.exe ;;
    Wasi-*) suffix=.wasm ;;
    *) suffix= ;;
  esac
  if [ "$kind" = wine ] && ! preparewine; then
    say "cross-check: $row: the wine image did not build (see below); runs are skipped"
    tail -5 "$work/wine.log" 2>/dev/null
    kind=none
  fi
  say "== $row ($triple), runner: $kind"
  mkdir -p "$work/$row"

  # the sample programs
  for sample in $samples; do
    if matchLeftOut "$row" "$sample"; then
      say "  $sample: left out on this row"
      record "$row" "$sample" skipped "left out"
      continue
    fi
    binary="$work/$row/$sample$suffix"
    if ! "$narrac" build --target "$row" -o "$binary" "$root/testdata/$sample.nra" > "$work/$row/$sample.build" 2>&1; then
      say "  $sample: build failed"
      sed -n 1,3p "$work/$row/$sample.build" | sed 's/^/    /'
      record "$row" "$sample" build-failed ""
      continue
    fi
    if [ "$kind" = none ]; then
      say "  $sample: built, not run ($(runner_note "$row"))"
      record "$row" "$sample" built "no runner"
      continue
    fi
    runbinary "$kind" "$row" "$binary" > "$work/$row/$sample.stdout" 2> "$work/$row/$sample.stderr"
    status=$?
    if [ "$status" -eq 0 ] && cmp -s "$work/$row/$sample.stdout" "$root/testdata/$sample.out"; then
      say "  $sample: ok"
      record "$row" "$sample" ran ""
    else
      say "  $sample: run failed (exit $status)"
      sed -n 1,3p "$work/$row/$sample.stderr" | sed 's/^/    /'
      record "$row" "$sample" run-failed "exit $status"
    fi
  done

  # the seed, built for the row and run on testdata/c_interop.nra
  if [ "${SKIP_SEED:-}" = 1 ]; then
    return
  fi
  seedbinary="$work/$row/seed$suffix"
  seedlibs_text=$(seedlibs "$row")
  # The flag and library lists hold several words on purpose.
  # shellcheck disable=SC2046,SC2086
  if ! "$CC" -target "$triple" -O2 -DNARRA_RELEASE -std=c11 -w $(seedflags "$row") -o "$seedbinary" "$root/seed/narrac.c" $seedlibs_text > "$work/$row/seed.build" 2>&1; then
    say "  seed: build failed"
    sed -n 1,3p "$work/$row/seed.build" | sed 's/^/    /'
    record "$row" seed build-failed ""
    return
  fi
  if [ "$kind" = none ]; then
    say "  seed: built, not run ($(runner_note "$row"))"
    record "$row" seed built "no runner"
    return
  fi
  seedemit=testdata/c_interop.nra
  case "$kind" in
    native | rosetta | wasmtime) seedpath=$root/$seedemit ;;
    *) seedpath=/repo/$seedemit ;;
  esac
  if ! runbinary "$kind" "$row" "$seedbinary" emit "$seedpath" > "$work/$row/seed.c" 2> "$work/$row/seed.stderr"; then
    say "  seed: run failed"
    sed -n 1,3p "$work/$row/seed.stderr" | sed 's/^/    /'
    record "$row" seed run-failed ""
    return
  fi
  if [ ! -s "$work/reference.c" ]; then
    "$work/reference-seed" emit "$root/$seedemit" > "$work/reference.c" 2> /dev/null
  fi
  if cmp -s "$work/$row/seed.c" "$work/reference.c"; then
    say "  seed: emits the same C as the host"
    record "$row" seed ran ""
  else
    say "  seed: emitted C differs from the host"
    record "$row" seed run-failed "C differs"
  fi
}

# seedflags prints the extra compiler flags a row needs; the Wasi ones are those
# of scripts/wasm-build.sh.
seedflags() {
  case "$1" in
    Wasi-*) say "-mexception-handling -mllvm -wasm-enable-sjlj -mllvm -wasm-use-legacy-eh=false" ;;
    *) say "" ;;
  esac
}

seedlibs() {
  case "$1" in
    Windows-*) say "" ;;
    Wasi-*) say "-lsetjmp -lm" ;;
    *) say "-lm" ;;
  esac
}

# the host's seed, the reference for the emitted C
if [ "${SKIP_SEED:-}" != 1 ]; then
  if ! "$CC" -O2 -DNARRA_RELEASE -std=c11 -w -o "$work/reference-seed" "$root/seed/narrac.c" -lm > "$work/reference-seed.log" 2>&1; then
    cat "$work/reference-seed.log" >&2
    say "cross-check: the host seed did not build" >&2
    exit 2
  fi
fi

for row in $rows; do
  checkrow "$row"
done

# --- the report ----------------------------------------------------------------

say ""
say "row                      built  ran  skipped  failed"
unexpected=0
expected=0
for row in $rows; do
  built=0
  ran=0
  skipped=0
  failed=0
  while IFS='	' read -r name _ status _; do
    if [ "$name" != "$row" ]; then continue; fi
    case "$status" in
      built) built=$((built + 1)) ;;
      ran) ran=$((ran + 1)) ;;
      skipped) skipped=$((skipped + 1)) ;;
      *) failed=$((failed + 1)) ;;
    esac
  done < "$results"
  mark=""
  if [ "$failed" -gt 0 ]; then
    if matchExpectedFailure "$row" && [ "${STRICT:-}" != 1 ]; then
      mark="expected failure"
      expected=$((expected + 1))
    else
      mark="FAIL"
      unexpected=$((unexpected + 1))
    fi
  elif [ "$ran" -eq 0 ]; then
    mark="nothing ran"
    if [ "${REQUIRE_RUN:-}" = 1 ]; then
      mark="FAIL, nothing ran"
      unexpected=$((unexpected + 1))
    fi
  elif matchExpectedFailure "$row"; then
    mark="passes now: remove it from expected_failures"
  fi
  printf '%-24s %5d %4d %8d %7d  %s\n' "$row" "$built" "$ran" "$skipped" "$failed" "$mark"
done

if [ "$unexpected" -gt 0 ]; then
  say "cross-check: FAIL, $unexpected row(s) did not match" >&2
  exit 1
fi
say "cross-check: PASS ($expected expected failure rows)"
exit 0
