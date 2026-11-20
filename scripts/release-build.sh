#!/bin/sh
# Build the release archive for one row of `narrac targets`.
#   sh scripts/release-build.sh [--dry-run] [--target <row>] <version>
# Without --target the row is this machine. The script builds narrac from seed/narrac.c
# with `cc` alone (sh selfhost/bootstrap.sh, which checks the fixed point). For this
# machine's row it runs sh scripts/test.sh with that compiler. For another row it builds
# narrac for the row with `zig cc` (a cross build; the tests do not run). The glibc Linux
# rows are always built with zig, against glibc 2.28, so the binary runs on Debian 10 and
# later, Ubuntu 18.04 and later and RHEL 8 and later. Every binary is stripped (zig cc -s).
# It then packs dist/narra-<version>-<name>.tar.gz (.zip on Windows) and its .sha256, and,
# where this machine can run the row, unpacks the archive and compiles a sample with the
# unpacked binary. scripts/smoke-archive.sh is that check for one archive; the release
# workflow runs it on every archive before it publishes.
# --dry-run changes nothing: this script never uploads. The release workflow uploads.
# The archive holds narra-<version>-<name>/narrac (narrac.exe on Windows). The std library
# is built into narrac, so the archive has no std/ folder.
# narrac --version prints <version>: the script builds the release binary from a copy of
# selfhost/ whose selfhost/src/driver/version.nra holds the version. The seed and bootstrap
# builds keep dev.
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
  echo "       sh scripts/release-build.sh --check-archive <archive.tar.gz>" >&2
  exit 2
}

# --check-archive unpacks one archive and runs a hello program with the unpacked narrac.
# NARRA_RUN_PREFIX, when set, is a command put before the binary (for example `arch -x86_64`).
# The release workflow runs this on every archive before it publishes.
check_archive() {
  archive_path=$1
  [ -f "$archive_path" ] || {
    echo "release-build: no such archive: $archive_path" >&2
    exit 2
  }
  check_dir=$(mktemp -d)
  trap 'rm -rf "$check_dir"' EXIT
  mkdir -p "$check_dir/unpack" "$check_dir/sample"
  tar -C "$check_dir/unpack" -xzf "$archive_path"
  check_binary=$(ls "$check_dir"/unpack/*/narrac)
  # NARRA_EXPECT_VERSION, when set, is the version the binary must print.
  if [ -n "${NARRA_EXPECT_VERSION:-}" ]; then
    # shellcheck disable=SC2086
    got_version=$(${NARRA_RUN_PREFIX:-} "$check_binary" --version)
    if [ "$got_version" != "$NARRA_EXPECT_VERSION" ]; then
      echo "release-build: narrac --version printed $got_version, want $NARRA_EXPECT_VERSION" >&2
      exit 1
    fi
  fi
  # The archive holds the compiler alone: the std library is built into it.
  check_content=$(ls "$check_dir"/unpack/*)
  if [ "$check_content" != narrac ]; then
    echo "release-build: the archive holds more than narrac: $check_content" >&2
    exit 1
  fi
  cat > "$check_dir/sample/hello.nra" <<'SAMPLE'
export function run(args: Array<string>): int {
  print("hello, narra")
  return 0
}
SAMPLE
  # shellcheck disable=SC2086
  got=$(${NARRA_RUN_PREFIX:-} "$check_binary" run "$check_dir/sample/hello.nra")
  if [ "$got" != "hello, narra" ]; then
    echo "release-build: the unpacked narrac printed: $got" >&2
    exit 1
  fi
  echo "release-build: ok: $archive_path runs"
  exit 0
}

if [ "${1:-}" = --check-archive ]; then
  [ "$#" -eq 2 ] || usage
  check_archive "$2"
fi

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

# The glibc Linux rows are built with zig against glibc 2.28, never with the native cc,
# so a binary from a new runner does not need a new glibc.
zig_row=no
case "$target" in
  Linux-X86_64 | Linux-Arm64) zig_row=yes ;;
esac

# A copy of selfhost/ with the version stamped in, so the release binary reports it.
stamped=$work/stamped
mkdir -p "$stamped"
cp -R "$root/selfhost/." "$stamped/"
sed "s/^export const VERSION = \"dev\"\$/export const VERSION = \"$version\"/" "$root/selfhost/src/driver/version.nra" > "$stamped/src/driver/version.nra"
if cmp -s "$root/selfhost/src/driver/version.nra" "$stamped/src/driver/version.nra"; then
  echo "release-build: selfhost/src/driver/version.nra has no VERSION line to stamp" >&2
  exit 1
fi

if [ "$target" = "$host_row" ]; then
  if [ "${NARRA_RELEASE_SKIP_TESTS:-}" = 1 ]; then
    echo "release-build: NARRA_RELEASE_SKIP_TESTS=1, skipping sh scripts/test.sh"
  else
    echo "release-build: run the test suites"
    NARRAC="$work/compiler/narrac" sh "$root/scripts/test.sh"
  fi
fi
if [ "$target" = "$host_row" ] && [ "$zig_row" = no ]; then
  # The compiler of the fixed point check builds the stamped copy with the native cc.
  echo "release-build: build the release binary for $target"
  built=$work/release/$binary_name
  mkdir -p "$work/release"
  "$work/compiler/narrac" build -release -o "$built" "$stamped"
fi
if [ "$target" != "$host_row" ] || [ "$zig_row" = yes ]; then
  # A cross build: the compiler of this machine builds selfhost for the row with zig cc.
  # CC names one program, so a wrapper holds the words. The wrapper adds -s (no symbols or
  # debug info: zig cc keeps them otherwise, which made Linux binaries seven times larger
  # than the Windows one) and, for a glibc target, the glibc floor 2.28.
  echo "release-build: build for $target with zig cc"
  if command -v zig >/dev/null 2>&1; then
    zig_command="zig cc"
  elif python3 -m ziglang version >/dev/null 2>&1; then
    zig_command="python3 -m ziglang cc"
  else
    echo "release-build: a cross build needs zig (or python3 -m ziglang)" >&2
    exit 1
  fi
  cat > "$work/zigcc" <<WRAPPER
#!/bin/sh
for argument; do
  case "\$argument" in
    x86_64-linux-gnu | aarch64-linux-gnu) argument=\$argument.2.28 ;;
  esac
  set -- "\$@" "\$argument"
  shift
done
exec $zig_command -s "\$@"
WRAPPER
  chmod +x "$work/zigcc"
  built=$work/cross/$binary_name
  mkdir -p "$work/cross"
  CC="$work/zigcc" "$work/compiler/narrac" build -release --target "$target" -o "$built" "$stamped"
fi

# Pack the binary. COPYFILE_DISABLE stops macOS tar from adding
# ._ files that hold extended attributes.
echo "release-build: pack"
COPYFILE_DISABLE=1
export COPYFILE_DISABLE
mkdir -p "$work/pack/$name"
cp "$built" "$work/pack/$name/$binary_name"
chmod 755 "$work/pack/$name/$binary_name"
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
# A musl build is static, so a Linux machine of the same processor runs it too. A glibc
# row built here with zig runs on this machine when it is the host row. The release workflow
# runs --check-archive on every archive, on a machine that can run it, before it publishes.
can_run=no
run_prefix=
if [ "$target" = "$host_row" ]; then
  can_run=yes
elif [ "$host_os" = Linux ] && [ "$target" = "Linux-$host_arch-musl" ]; then
  can_run=yes
elif [ "$target" = MacOS-X86_64 ] && [ "$host_os" = MacOS ] && arch -x86_64 /usr/bin/true >/dev/null 2>&1; then
  can_run=yes
  run_prefix="arch -x86_64"
fi
if [ "$can_run" = no ]; then
  echo "release-build: this machine cannot run $target, so the archive is not run"
elif [ "$extension" = zip ]; then
  echo "release-build: a zip archive is run on Windows by the release workflow"
else
  echo "release-build: check the archive"
  NARRA_EXPECT_VERSION=$version NARRA_RUN_PREFIX=$run_prefix sh "$0" --check-archive "$dist/$archive"
fi

echo "release-build: ok"
echo "  $dist/$archive"
echo "  $dist/$archive.sha256"
