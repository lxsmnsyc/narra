#!/bin/sh
# Build the release archive for this machine.
#   sh scripts/release-build.sh [--dry-run] <version>
# The script builds narrac from seed/narrac.c with `cc` alone (sh selfhost/bootstrap.sh,
# which checks the fixed point), runs sh scripts/test.sh with that compiler, packs
# dist/narra-<version>-<os>-<arch>.tar.gz and its .sha256, then unpacks the archive
# and compiles a sample with the unpacked binary.
# --dry-run changes nothing: this script never uploads. The release workflow uploads.
# The archive holds narra-<version>-<os>-<arch>/narrac and the std/ sources.
# <os> is macos or linux. <arch> is arm64 or x86_64.
# NARRA_DIST picks the output folder (default: dist/ in the repository).
# NARRA_RELEASE_SKIP_TESTS=1 skips sh scripts/test.sh. Use it only to check the packing
# where the tests are too slow (for example linux/amd64 under emulation on a Mac).
set -eu
root=$(cd "$(dirname "$0")/.." && pwd)

usage() {
  echo "usage: sh scripts/release-build.sh [--dry-run] <version>" >&2
  exit 2
}

version=
while [ "$#" -gt 0 ]; do
  case "$1" in
    --dry-run) ;;
    -*) usage ;;
    *)
      [ -z "$version" ] || usage
      version=$1
      ;;
  esac
  shift
done
[ -n "$version" ] || usage
case "$version" in
  *[!0-9A-Za-z._-]*)
    echo "release-build: the version may hold only letters, digits, . _ -: $version" >&2
    exit 2
    ;;
esac

# Name the platform.
case "$(uname -s)" in
  Darwin) os=macos ;;
  Linux) os=linux ;;
  *)
    echo "release-build: unsupported system: $(uname -s)" >&2
    exit 1
    ;;
esac
case "$(uname -m)" in
  arm64 | aarch64) arch=arm64 ;;
  x86_64 | amd64) arch=x86_64 ;;
  *)
    echo "release-build: unsupported processor: $(uname -m)" >&2
    exit 1
    ;;
esac

if command -v sha256sum >/dev/null 2>&1; then
  hash_file() { sha256sum "$1"; }
else
  hash_file() { shasum -a 256 "$1"; }
fi

dist=${NARRA_DIST:-$root/dist}
name=narra-$version-$os-$arch
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

echo "release-build: $name"
echo "release-build: build from the seed and check the fixed point"
sh "$root/selfhost/bootstrap.sh" "$work/compiler"

if [ "${NARRA_RELEASE_SKIP_TESTS:-}" = 1 ]; then
  echo "release-build: NARRA_RELEASE_SKIP_TESTS=1, skipping sh scripts/test.sh"
else
  echo "release-build: run the test suites"
  NARRAC="$work/compiler/narrac" sh "$root/scripts/test.sh"
fi

# Pack the binary and the std sources. COPYFILE_DISABLE stops macOS tar from adding
# ._ files that hold extended attributes.
echo "release-build: pack"
COPYFILE_DISABLE=1
export COPYFILE_DISABLE
mkdir -p "$work/pack/$name"
cp "$work/compiler/narrac" "$work/pack/$name/narrac"
cp -R "$root/std" "$work/pack/$name/std"
mkdir -p "$dist"
tar -C "$work/pack" -czf "$dist/$name.tar.gz" "$name"
(cd "$dist" && hash_file "$name.tar.gz" > "$name.tar.gz.sha256")

# Unpack the archive and compile a sample with the unpacked binary.
echo "release-build: check the archive"
mkdir -p "$work/unpack" "$work/sample"
tar -C "$work/unpack" -xzf "$dist/$name.tar.gz"
cat > "$work/sample/hello.nra" <<'SAMPLE'
export function run(args: Array<string>): int {
  print("hello, narra")
  return 0
}
SAMPLE
got=$("$work/unpack/$name/narrac" run "$work/sample/hello.nra")
if [ "$got" != "hello, narra" ]; then
  echo "release-build: the unpacked narrac printed: $got" >&2
  exit 1
fi

echo "release-build: ok"
echo "  $dist/$name.tar.gz"
echo "  $dist/$name.tar.gz.sha256"
