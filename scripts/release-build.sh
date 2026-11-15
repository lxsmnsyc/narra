#!/bin/sh
# Build the release archive for one row of `narrac targets`.
#   sh scripts/release-build.sh [--dry-run] [--target <row>] <version>
# Without --target the row is this machine. The script builds narrac from seed/narrac.c
# with `cc` alone (sh selfhost/bootstrap.sh, which checks the fixed point). For this
# machine's row it runs sh scripts/test.sh with that compiler. For another row it builds
# narrac for the row with `zig cc` (a cross build; the tests do not run). It then packs
# dist/narra-<version>-<name>.tar.gz (.zip on Windows) and its .sha256, and, where this
# machine can run the row, unpacks the archive and compiles a sample with the unpacked binary.
# --dry-run changes nothing: this script never uploads. The release workflow uploads.
# The archive holds narra-<version>-<name>/narrac (narrac.exe on Windows) and the std/ sources.
# <name> is <os>-<arch> or <os>-<arch>-musl, in lower case: macos-arm64, macos-x86_64,
# linux-x86_64, linux-arm64, linux-x86_64-musl, linux-arm64-musl, windows-x86_64.
# NARRA_DIST picks the output folder (default: dist/ in the repository).
# NARRA_RELEASE_SKIP_TESTS=1 skips sh scripts/test.sh. Use it only to check the packing
# where the tests are too slow (for example linux/amd64 under emulation on a Mac).
# zig is the C compiler of a cross build; without a zig command, python3 -m ziglang is used.
set -eu
root=$(cd "$(dirname "$0")/.." && pwd)

usage() {
  echo "usage: sh scripts/release-build.sh [--dry-run] [--target <row>] <version>" >&2
  exit 2
}

version=
target=
while [ "$#" -gt 0 ]; do
  case "$1" in
    --dry-run) ;;
    --target)
      [ "$#" -ge 2 ] || usage
      target=$2
      shift
      ;;
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

# Name this machine's row the way `narrac targets` does.
case "$(uname -s)" in
  Darwin) host_os=MacOS ;;
  Linux) host_os=Linux ;;
  *)
    echo "release-build: unsupported system: $(uname -s)" >&2
    exit 1
    ;;
esac
case "$(uname -m)" in
  arm64 | aarch64) host_arch=Arm64 ;;
  x86_64 | amd64) host_arch=X86_64 ;;
  *)
    echo "release-build: unsupported processor: $(uname -m)" >&2
    exit 1
    ;;
esac
host_row=$host_os-$host_arch
if [ -z "$target" ]; then
  target=$host_row
fi

# The archive name and kind follow the row.
case "$target" in
  MacOS-Arm64) platform=macos-arm64 ;;
  MacOS-X86_64) platform=macos-x86_64 ;;
  Linux-X86_64) platform=linux-x86_64 ;;
  Linux-Arm64) platform=linux-arm64 ;;
  Linux-X86_64-musl) platform=linux-x86_64-musl ;;
  Linux-Arm64-musl) platform=linux-arm64-musl ;;
  Windows-X86_64) platform=windows-x86_64 ;;
  *)
    echo "release-build: no release archive for the row $target" >&2
    exit 2
    ;;
esac
case "$target" in
  Windows-*)
    binary_name=narrac.exe
    extension=zip
    ;;
  *)
    binary_name=narrac
    extension=tar.gz
    ;;
esac

if command -v sha256sum >/dev/null 2>&1; then
  hash_file() { sha256sum "$1"; }
else
  hash_file() { shasum -a 256 "$1"; }
fi

dist=${NARRA_DIST:-$root/dist}
name=narra-$version-$platform
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

echo "release-build: $name"
echo "release-build: build from the seed and check the fixed point"
sh "$root/selfhost/bootstrap.sh" "$work/compiler"

if [ "$target" = "$host_row" ]; then
  built=$work/compiler/narrac
  if [ "${NARRA_RELEASE_SKIP_TESTS:-}" = 1 ]; then
    echo "release-build: NARRA_RELEASE_SKIP_TESTS=1, skipping sh scripts/test.sh"
  else
    echo "release-build: run the test suites"
    NARRAC="$work/compiler/narrac" sh "$root/scripts/test.sh"
  fi
else
  # A cross build: the compiler of this machine builds selfhost for the row with zig cc.
  # CC names one program, so a wrapper holds the words.
  echo "release-build: cross build for $target with zig cc (the tests do not run)"
  if command -v zig >/dev/null 2>&1; then
    zig_command="zig cc"
  elif python3 -m ziglang version >/dev/null 2>&1; then
    zig_command="python3 -m ziglang cc"
  else
    echo "release-build: a cross build needs zig (or python3 -m ziglang)" >&2
    exit 1
  fi
  {
    echo '#!/bin/sh'
    echo "exec $zig_command \"\$@\""
  } > "$work/zigcc"
  chmod +x "$work/zigcc"
  built=$work/cross/$binary_name
  mkdir -p "$work/cross"
  CC="$work/zigcc" "$work/compiler/narrac" build -release --target "$target" -o "$built" "$root/selfhost"
fi

# Pack the binary and the std sources. COPYFILE_DISABLE stops macOS tar from adding
# ._ files that hold extended attributes.
echo "release-build: pack"
COPYFILE_DISABLE=1
export COPYFILE_DISABLE
mkdir -p "$work/pack/$name"
cp "$built" "$work/pack/$name/$binary_name"
chmod 755 "$work/pack/$name/$binary_name"
cp -R "$root/std" "$work/pack/$name/std"
mkdir -p "$dist"
archive=$name.$extension
rm -f "$dist/$archive"
if [ "$extension" = zip ]; then
  (cd "$work/pack" && zip -qr "$dist/$archive" "$name")
else
  tar -C "$work/pack" -czf "$dist/$archive" "$name"
fi
(cd "$dist" && hash_file "$archive" > "$archive.sha256")

# Where this machine can run the row, unpack the archive and compile a sample with it.
# A musl build is static, so a Linux machine of the same processor runs it too.
can_run=no
use_rosetta=no
if [ "$target" = "$host_row" ]; then
  can_run=yes
elif [ "$host_os" = Linux ] && [ "$target" = "Linux-$host_arch-musl" ]; then
  can_run=yes
elif [ "$target" = MacOS-X86_64 ] && [ "$host_os" = MacOS ] && arch -x86_64 /usr/bin/true >/dev/null 2>&1; then
  can_run=yes
  use_rosetta=yes
fi
if [ "$can_run" = no ]; then
  echo "release-build: this machine cannot run $target, so the archive is not run"
else
  echo "release-build: check the archive"
  mkdir -p "$work/unpack" "$work/sample"
  tar -C "$work/unpack" -xzf "$dist/$archive"
  cat > "$work/sample/hello.nra" <<'SAMPLE'
export function run(args: Array<string>): int {
  print("hello, narra")
  return 0
}
SAMPLE
  if [ "$use_rosetta" = yes ]; then
    got=$(arch -x86_64 "$work/unpack/$name/narrac" run "$work/sample/hello.nra")
  else
    got=$("$work/unpack/$name/narrac" run "$work/sample/hello.nra")
  fi
  if [ "$got" != "hello, narra" ]; then
    echo "release-build: the unpacked narrac printed: $got" >&2
    exit 1
  fi
fi

echo "release-build: ok"
echo "  $dist/$archive"
echo "  $dist/$archive.sha256"
