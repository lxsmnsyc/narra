#!/bin/sh
# Install a prebuilt narrac from a GitHub release.
#   curl -fsSL https://raw.githubusercontent.com/lxsmnsyc/narra/main/install.sh | sh
# It picks the archive for this system, checks its SHA-256 against the checksum file
# of the release, and copies narrac to ~/.local/bin.
# Variables:
#   NARRA_VERSION   the version to install, such as 0.1.0 or v0.1.0 (default: the latest release)
#   NARRA_INSTALL   the folder that receives narrac (default: $HOME/.local/bin)
#   NARRA_BASE_URL  the folder that holds the archive and its .sha256 file
#                   (default: the release of NARRA_VERSION on GitHub); for tests
set -eu

repository=lxsmnsyc/narra

fail() {
  echo "install: $1" >&2
  exit 1
}

# Name the platform the way the release archives do.
case "$(uname -s)" in
  Darwin) os=macos ;;
  Linux) os=linux ;;
  *) fail "unsupported system: $(uname -s). Build from source: docs/install.md" ;;
esac
case "$(uname -m)" in
  arm64 | aarch64) arch=arm64 ;;
  x86_64 | amd64) arch=x86_64 ;;
  *) fail "unsupported processor: $(uname -m). Build from source: docs/install.md" ;;
esac
case "$os-$arch" in
  macos-arm64 | linux-x86_64) ;;
  *) fail "no prebuilt binary for $os $arch. Build from source: docs/install.md" ;;
esac

command -v curl >/dev/null 2>&1 || fail "curl is not installed"
command -v tar >/dev/null 2>&1 || fail "tar is not installed"
if command -v sha256sum >/dev/null 2>&1; then
  hash_file() { sha256sum "$1" | cut -d ' ' -f 1; }
elif command -v shasum >/dev/null 2>&1; then
  hash_file() { shasum -a 256 "$1" | cut -d ' ' -f 1; }
else
  fail "neither sha256sum nor shasum is installed"
fi

# Find the version. Without NARRA_VERSION, follow the redirect of the latest release.
version=${NARRA_VERSION:-}
if [ -z "$version" ]; then
  latest=$(curl -fsSLI --proto '=https' --proto-redir '=https' -o /dev/null -w '%{url_effective}' "https://github.com/$repository/releases/latest") ||
    fail "could not look up the latest release"
  version=${latest##*/}
  case "$version" in
    v[0-9]*) ;;
    *) fail "there is no release yet at https://github.com/$repository/releases" ;;
  esac
fi
version=${version#v}
case "$version" in
  '' | *[!0-9A-Za-z._-]*)
    fail "the version may hold only letters, digits, . _ -: $version"
    ;;
esac

name=narra-$version-$os-$arch
base=${NARRA_BASE_URL:-https://github.com/$repository/releases/download/v$version}
# A custom base URL must use https, except for a test server on this machine.
protocols='=https'
case "$base" in
  https://*) ;;
  http://*)
    authority=${base#http://}
    authority=${authority%%/*}
    # The host must be exactly localhost or 127.0.0.1, with no port or a port of digits only.
    host=${authority%%:*}
    port=${authority#"$host"}
    case "$host" in
      localhost | 127.0.0.1) ;;
      *) fail "NARRA_BASE_URL must start with https:// (http:// only for localhost or 127.0.0.1): $base" ;;
    esac
    case "$port" in
      '') ;;
      :*[!0-9]* | :) fail "NARRA_BASE_URL has a bad port: $base" ;;
    esac
    # A local test server may speak http; the first request only.
    protocols='=http,https'
    ;;
  *) fail "NARRA_BASE_URL must start with https:// (http:// only for localhost or 127.0.0.1): $base" ;;
esac
if [ -z "${NARRA_INSTALL:-}" ] && [ -z "${HOME:-}" ]; then
  fail "HOME is not set. Set NARRA_INSTALL to the folder that receives narrac"
fi
install_dir=${NARRA_INSTALL:-$HOME/.local/bin}

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

echo "install: narra $version for $os $arch"
curl -fsSL --proto "$protocols" --proto-redir '=https' -o "$work/$name.tar.gz" "$base/$name.tar.gz" ||
  fail "could not download $base/$name.tar.gz"
curl -fsSL --proto "$protocols" --proto-redir '=https' -o "$work/$name.tar.gz.sha256" "$base/$name.tar.gz.sha256" ||
  fail "could not download $base/$name.tar.gz.sha256"

# Check the SHA-256 before anything is unpacked.
expected=$(cut -d ' ' -f 1 "$work/$name.tar.gz.sha256")
listed=$(sed 's/^[^ ]* *\*\{0,1\}//' "$work/$name.tar.gz.sha256")
if [ "$listed" != "$name.tar.gz" ]; then
  fail "the checksum file names '$listed', not $name.tar.gz"
fi
actual=$(hash_file "$work/$name.tar.gz")
if [ -z "$expected" ] || [ "$expected" != "$actual" ]; then
  fail "the checksum of $name.tar.gz does not match (expected $expected, got $actual)"
fi

tar -C "$work" -xzf "$work/$name.tar.gz"
[ -f "$work/$name/narrac" ] || fail "the archive holds no narrac"
mkdir -p "$install_dir"
cp "$work/$name/narrac" "$install_dir/narrac.new"
chmod 755 "$install_dir/narrac.new"
mv "$install_dir/narrac.new" "$install_dir/narrac"

echo "install: narrac is in $install_dir"
case ":$PATH:" in
  *":$install_dir:"*) ;;
  *) echo "install: $install_dir is not on your PATH. Add: export PATH=\"$install_dir:\$PATH\"" ;;
esac
