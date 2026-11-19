#!/bin/sh
# Check install.sh (and install.ps1, where pwsh exists) against archives built here.
#   sh scripts/install-check.sh [row ...]
# For each row the script builds the release archive with scripts/release-build.sh
# (without the test suites), points install.sh at the folder through NARRA_BASE_URL
# (a file:// URL), installs into a temporary folder and runs narrac there.
# Without rows it checks this machine's row and, when Docker runs, the Linux rows of
# the Docker machine's processor (glibc in a Debian container, musl in an Alpine one).
# A row is a name of `narrac targets`, such as Linux-Arm64-musl. Rows other than this
# machine's need zig. Nothing is uploaded, and no tag or workflow is touched.
# Run it from a path under your home folder: colima does not mount /private/tmp.
set -eu
root=$(cd "$(dirname "$0")/.." && pwd)
work=$(mktemp -d "$root/.install-check.XXXXXX")
trap 'rm -rf "$work"' EXIT

version=0.0.0-check
dist=$work/dist
mkdir -p "$dist"

say() {
  printf '%s\n' "install-check: $*"
}

fail() {
  say "FAIL: $*" >&2
  exit 1
}

# This machine's row.
case "$(uname -s)" in
  Darwin) host_os=MacOS ;;
  Linux) host_os=Linux ;;
  *) fail "unsupported system: $(uname -s)" ;;
esac
case "$(uname -m)" in
  arm64 | aarch64) host_arch=Arm64 ;;
  x86_64 | amd64) host_arch=X86_64 ;;
  *) fail "unsupported processor: $(uname -m)" ;;
esac
host_row=$host_os-$host_arch

have_docker=0
docker_arch=
if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
  have_docker=1
  case "$(docker info --format '{{.Architecture}}')" in
    aarch64 | arm64) docker_arch=Arm64 ;;
    x86_64 | amd64) docker_arch=X86_64 ;;
  esac
fi

if [ "$#" -gt 0 ]; then
  rows="$*"
else
  rows=$host_row
  if [ "$host_os" = MacOS ] && [ "$have_docker" = 1 ] && [ -n "$docker_arch" ]; then
    rows="$rows Linux-$docker_arch Linux-$docker_arch-musl"
  elif [ "$host_os" = Linux ]; then
    rows="$rows Linux-$host_arch-musl"
  else
    say "docker is not running, so the Linux rows are skipped (colima start --cpu 2 --memory 3)"
  fi
fi

platformName() {
  case "$1" in
    MacOS-Arm64) echo macos-arm64 ;;
    MacOS-X86_64) echo macos-x86_64 ;;
    Linux-X86_64) echo linux-x86_64 ;;
    Linux-Arm64) echo linux-arm64 ;;
    Linux-X86_64-musl) echo linux-x86_64-musl ;;
    Linux-Arm64-musl) echo linux-arm64-musl ;;
    *) fail "no archive for the row $1" ;;
  esac
}

# Build the archives.
for row in $rows; do
  say "build the archive of $row"
  NARRA_RELEASE_SKIP_TESTS=1 NARRA_DIST=$dist sh "$root/scripts/release-build.sh" --target "$row" "$version" > "$work/build-$row.log" 2>&1 || {
    tail -20 "$work/build-$row.log" >&2
    fail "release-build failed for $row"
  }
done

# A row is checked where it can run: on this machine, or in a container.
for row in $rows; do
  platform=$(platformName "$row")
  archive=narra-$version-$platform.tar.gz
  [ -f "$dist/$archive" ] || fail "$archive was not built"

  if [ "$row" = "$host_row" ] || { [ "$host_os" = Linux ] && [ "$row" = "Linux-$host_arch-musl" ]; }; then
    say "install $row on this machine"
    bin=$work/bin-$row
    if ! NARRA_VERSION=$version NARRA_BASE_URL="file://$dist" NARRA_INSTALL=$bin sh "$root/install.sh" > "$work/install-$row.log" 2>&1; then
      cat "$work/install-$row.log" >&2
      fail "install.sh failed for $row"
    fi
    cat "$work/install-$row.log"
    got=$("$bin/narrac" --version) || fail "narrac --version failed for $row"
    [ "$got" = "$version" ] || fail "narrac --version printed $got for $row, want $version"
    say "ok: $row installed and runs"

    # A changed checksum must stop the install and leave the folder empty.
    say "install $row with a wrong checksum (must fail)"
    mkdir -p "$work/tampered"
    cp "$dist/$archive" "$work/tampered/$archive"
    sed 's/^./0/' "$dist/$archive.sha256" > "$work/tampered/$archive.sha256"
    if NARRA_VERSION=$version NARRA_BASE_URL="file://$work/tampered" NARRA_INSTALL="$work/bin-tampered" sh "$root/install.sh" > "$work/tampered.log" 2>&1; then
      fail "install.sh accepted a wrong checksum"
    fi
    [ ! -e "$work/bin-tampered/narrac" ] || fail "a wrong checksum still installed narrac"
    say "ok: a wrong checksum is refused"
    rm -rf "$work/tampered"
    continue
  fi

  case "$row" in
    Linux-*) ;;
    *)
      say "skip $row: this machine cannot run it"
      continue
      ;;
  esac
  [ "$have_docker" = 1 ] || {
    say "skip $row: docker is not running"
    continue
  }
  case "$row" in
    Linux-Arm64*) docker_platform=linux/arm64 ;;
    *) docker_platform=linux/amd64 ;;
  esac
  case "$row" in
    *-musl)
      image=alpine
      setup='apk add -q curl'
      ;;
    *)
      image=debian
      setup='apt-get update -qq >/dev/null && apt-get install -y -qq curl ca-certificates >/dev/null'
      ;;
  esac
  say "install $row in a $image container ($docker_platform)"
  # The folder holds the archive; install.sh is mounted beside it.
  docker run --rm --platform "$docker_platform" -v "$dist:/dist:ro" -v "$root/install.sh:/install.sh:ro" \
    -e "NARRA_VERSION=$version" -e NARRA_BASE_URL=file:///dist -e NARRA_INSTALL=/tmp/bin "$image" \
    sh -c "set -eu; $setup; sh /install.sh; /tmp/bin/narrac --version" > "$work/docker-$row.log" 2>&1 || {
    cat "$work/docker-$row.log" >&2
    fail "install.sh failed for $row in $image"
  }
  grep "install: " "$work/docker-$row.log"
  grep -qx "$version" "$work/docker-$row.log" || fail "narrac --version did not print $version in $image"
  say "ok: $row installed and runs in $image"
done

# install.ps1 needs PowerShell.
if command -v pwsh >/dev/null 2>&1; then
  say "parse install.ps1 with pwsh"
  pwsh -NoProfile -Command "\$errors = \$null; [System.Management.Automation.Language.Parser]::ParseFile('$root/install.ps1', [ref]\$null, [ref]\$errors) | Out-Null; if (\$errors.Count -gt 0) { \$errors; exit 1 }" ||
    fail "install.ps1 does not parse"
  say "ok: install.ps1 parses (it runs only on Windows)"
else
  say "pwsh is not installed: install.ps1 is checked by reading only"
fi

say "PASS"
