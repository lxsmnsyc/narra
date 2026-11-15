#!/bin/sh
# Build narrac's npm packages from npm/ and check them. Nothing is published.
#
#   sh scripts/npm-package.sh [--no-check] [version]
#
# The version defaults to 0.0.0-local. Steps:
#   1. A compiler from `sh selfhost/bootstrap.sh` (or the folder in NARRA_STAGE).
#   2. dist/narrac.wasm from that stage's C (scripts/wasm-build.sh).
#   3. The package of this machine's row gets the compiler. Rows named in
#      NARRA_NPM_ROWS (as `narrac targets` prints them) are built with
#      `narrac build --target` and zig cc.
#   4. Each package folder is copied to $NARRA_NPM_OUT/narrac-npm/package/, its version
#      is set, the binary is added, and `npm pack` writes the tarballs to
#      $NARRA_NPM_OUT/narrac-npm/. That subfolder is the only place the script deletes
#      from; it holds a marker file, and a folder of that name without the marker stops
#      the run.
#   5. Unless --no-check: the tarballs install into two scratch projects, one
#      with the host package and one with only narrac-wasm, and
#      npm/test/compare.js compares `npx narrac` and the API with native narrac.
#
# Environment:
#   NARRA_STAGE      folder that holds narrac and stage2.c (bootstrap.sh <dir>); skips step 1
#   NARRA_NPM_OUT    folder that receives narrac-npm/ (default dist/npm)
#   NARRA_NPM_ROWS   extra rows to build with zig, for example "Linux-Arm64-musl"
#   NARRA_NPM_SCRATCH folder for the scratch projects (default: a temporary folder)
#
# Needs node and npm, and zig for the wasm module. The compiler knows nothing
# about npm: this script and npm/ are the whole packaging.
set -eu
root=$(cd "$(dirname "$0")/.." && pwd)

usage() {
  echo "usage: sh scripts/npm-package.sh [--no-check] [version]" >&2
  exit 2
}

check=1
version=
while [ "$#" -gt 0 ]; do
  case "$1" in
    --no-check) check=0 ;;
    -*) usage ;;
    *)
      [ -z "$version" ] || usage
      version=$1
      ;;
  esac
  shift
done
version=${version:-0.0.0-local}
case "$version" in
  *[!0-9A-Za-z.+-]*)
    echo "npm-package: the version may hold only letters, digits, . + -: $version" >&2
    exit 2
    ;;
esac
for tool in node npm; do
  if ! command -v "$tool" >/dev/null 2>&1; then
    echo "npm-package: $tool is not installed" >&2
    exit 2
  fi
done

base=${NARRA_NPM_OUT:-$root/dist/npm}
out=$base/narrac-npm
marker=$out/.narrac-npm-owned
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
trap 'exit 130' INT TERM

# Name this machine's package, with the target name `narrac targets` prints.
case "$(uname -s)" in
  Darwin) host_os=darwin ;;
  Linux) host_os=linux ;;
  *)
    echo "npm-package: unsupported system: $(uname -s)" >&2
    exit 1
    ;;
esac
case "$(uname -m)" in
  arm64 | aarch64) host_arch=arm64 ;;
  x86_64 | amd64) host_arch=x64 ;;
  *)
    echo "npm-package: unsupported processor: $(uname -m)" >&2
    exit 1
    ;;
esac
# The loader's own table names the package, so the libc is detected one way.
host_package=$(node -p 'const row = require(process.argv[1]).findHostRow(); row ? row.package : ""' "$root/npm/narrac/lib/platform.js")
if [ -z "$host_package" ] || [ ! -d "$root/npm/$host_package" ]; then
  echo "npm-package: no package for this machine ($host_os $host_arch)" >&2
  exit 1
fi

# 1. The compiler and its C.
stage=${NARRA_STAGE:-}
if [ -z "$stage" ]; then
  echo "npm-package: build the compiler from the seed"
  stage=$work/stage
  sh "$root/selfhost/bootstrap.sh" "$stage"
fi
if [ ! -x "$stage/narrac" ] || [ ! -f "$stage/stage2.c" ]; then
  echo "npm-package: $stage needs narrac and stage2.c" >&2
  exit 1
fi

# 2. The wasm module, from the same C as the native compiler.
echo "npm-package: build narrac.wasm"
OUT=$work/narrac.wasm sh "$root/scripts/wasm-build.sh" "$stage/stage2.c"

# 3. Binaries by package folder name. A row other than this machine's is built
# by the compiler itself, since the C it emits depends on the target.
mkdir -p "$work/bin"
if [ -n "${NARRA_NPM_ROWS:-}" ]; then
  if ! command -v zig >/dev/null 2>&1; then
    echo "npm-package: NARRA_NPM_ROWS needs zig" >&2
    exit 2
  fi
  # CC names one program, so zig cc goes through a wrapper.
  printf '#!/bin/sh\nexec zig cc "$@"\n' > "$work/zigcc"
  chmod +x "$work/zigcc"
fi
cp "$stage/narrac" "$work/bin/$host_package"
for row in ${NARRA_NPM_ROWS:-}; do
  case "$row" in
    MacOS-Arm64) package=narrac-darwin-arm64 ;;
    MacOS-X86_64) package=narrac-darwin-x64 ;;
    Linux-Arm64) package=narrac-linux-arm64 ;;
    Linux-X86_64) package=narrac-linux-x64 ;;
    Linux-Arm64-musl) package=narrac-linux-arm64-musl ;;
    Linux-X86_64-musl) package=narrac-linux-x64-musl ;;
    Windows-X86_64) package=narrac-win32-x64 ;;
    *)
      echo "npm-package: no package for row $row" >&2
      exit 2
      ;;
  esac
  echo "npm-package: build $package ($row) with narrac build --target and zig cc"
  CC=$work/zigcc "$stage/narrac" build --target "$row" -bin narrac -o "$work/bin/$package" "$root/selfhost"
done

# 4. Copy each package, set the version, add the binary and pack.
# Clear only the subfolder this script owns, and only when an earlier run marked it.
if [ -e "$out" ] && [ ! -f "$marker" ]; then
  echo "npm-package: $out exists and is not from this script; remove it or pick another NARRA_NPM_OUT" >&2
  exit 1
fi
rm -rf "$out"
mkdir -p "$out/package"
: > "$marker"
for source in "$root"/npm/narrac "$root"/npm/narrac-*; do
  [ -f "$source/package.json" ] || continue
  name=$(basename "$source")
  case "$name" in
    narrac-darwin-* | narrac-linux-* | narrac-win32-*)
      # A platform package ships only when this run has its binary.
      [ -f "$work/bin/$name" ] || continue
      mkdir -p "$out/package/$name/bin"
      ;;
    narrac-wasm | narrac) mkdir -p "$out/package/$name" ;;
    *) continue ;;
  esac
  # The tracked template files first; the binary and the module go in after them.
  for item in package.json index.js README.md bin lib; do
    [ -e "$source/$item" ] && cp -R "$source/$item" "$out/package/$name/"
  done
  case "$name" in
    narrac-darwin-* | narrac-linux-*)
      cp "$work/bin/$name" "$out/package/$name/bin/narrac"
      chmod 755 "$out/package/$name/bin/narrac"
      ;;
    narrac-win32-*)
      cp "$work/bin/$name" "$out/package/$name/bin/narrac.exe"
      chmod 755 "$out/package/$name/bin/narrac.exe"
      ;;
    narrac-wasm)
      cp "$work/narrac.wasm" "$out/package/$name/narrac.wasm"
      chmod 644 "$out/package/$name/narrac.wasm"
      ;;
  esac
done

# Set the version in every package and pin the main package's optional
# dependencies (every platform, built in this run or not) to it.
VERSION=$version OUT=$out node -e '
const fs = require("node:fs");
const path = require("node:path");
const base = path.join(process.env.OUT, "package");
const built = fs.readdirSync(base);
for (const name of built) {
  const file = path.join(base, name, "package.json");
  const manifest = JSON.parse(fs.readFileSync(file, "utf8"));
  manifest.version = process.env.VERSION;
  if (manifest.optionalDependencies) {
    for (const other of Object.keys(manifest.optionalDependencies)) {
      manifest.optionalDependencies[other] = process.env.VERSION;
    }
  }
  fs.writeFileSync(file, JSON.stringify(manifest, null, 2) + "\n");
}
'
for package in "$out"/package/*; do
  (cd "$package" && npm pack --silent --pack-destination "$out" >/dev/null)
done
echo "npm-package: tarballs in $out"
ls "$out"/*.tgz

[ "$check" = 1 ] || exit 0

# 5. Install the tarballs into scratch projects and compare with native narrac.
scratch=${NARRA_NPM_SCRATCH:-$work/scratch}
mkdir -p "$scratch"
main=$(ls "$out"/narrac-"$version".tgz)
wasm=$(ls "$out"/narrac-wasm-"$version".tgz)
host=$(ls "$out/$host_package-$version.tgz")
for variant in host wasm; do
  project=$scratch/$variant
  rm -rf "$project"
  mkdir -p "$project"
  (
    cd "$project"
    echo '{"name":"narrac-scratch","private":true}' > package.json
    # The registry has no narrac packages, so the other optional dependencies are skipped.
    if [ "$variant" = host ]; then
      npm install --no-audit --no-fund --offline "$main" "$host" >/dev/null
      kind=native
    else
      npm install --no-audit --no-fund --offline "$main" "$wasm" >/dev/null
      kind=wasm
    fi
    echo "npm-package: compare the $variant project ($kind)"
    node "$root/npm/test/compare.js" "$project" "$stage/narrac" "$kind" "$root"
  )
done
echo "npm-package: ok"
