#!/bin/sh
# Install a prebuilt narrac from a GitHub release.
#   curl -fsSL https://github.com/lxsmnsyc/narra/releases/latest/download/install.sh | sh
# It picks the archive for this system (macOS or Linux; arm64 or x86_64; musl on a musl
# system), checks its SHA-256 against the checksum file of the release, and copies narrac
# to ~/.local/bin.
# Variables:
#   NARRA_VERSION   the version to install, such as 0.1.0 or v0.1.0 (default: the latest release;
#                   when there is only a prerelease, the newest release of any kind)
#   NARRA_INSTALL   the folder that receives narrac (default: $HOME/.local/bin)
#   NARRA_BASE_URL  the folder that holds the archive and its .sha256 file
#                   (default: the release of NARRA_VERSION on GitHub); for tests; it may
#                   also be a file:// URL of a folder
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
# A shell that runs under Rosetta on Apple Silicon reports x86_64; install the arm64 build.
if [ "$os" = macos ] && [ "$arch" = x86_64 ] && [ "$(sysctl -n sysctl.proc_translated 2>/dev/null || true)" = 1 ]; then
  arch=arm64
fi
# A musl system, such as Alpine, needs the musl archive.
libc=
if [ "$os" = linux ]; then
  if (ldd --version 2>&1 || true) | grep -qi musl; then
    libc=-musl
  else
    for loader in /lib/ld-musl-*; do
      if [ -e "$loader" ]; then libc=-musl; fi
    done
  fi
fi

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
# GitHub's latest release skips prereleases, so when it finds none, ask for the newest
# release of any kind.
version=${NARRA_VERSION:-}
if [ -z "$version" ]; then
  latest=$(curl -fsSLI --proto '=https' --proto-redir '=https' -o /dev/null -w '%{url_effective}' "https://github.com/$repository/releases/latest") ||
    fail "could not look up the latest release"
  version=${latest##*/}
  case "$version" in
    v[0-9]*) ;;
    *)
      newest=$(curl -fsSL --proto '=https' --proto-redir '=https' "https://api.github.com/repos/$repository/releases?per_page=1" 2>/dev/null |
        sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | sed -n 1p) || newest=
      case "$newest" in
        v[0-9]*) version=$newest ;;
        *) fail "there is no release yet at https://github.com/$repository/releases" ;;
      esac
      ;;
  esac
fi
version=${version#v}
case "$version" in
  '' | *[!0-9A-Za-z._-]*)
    fail "the version may hold only letters, digits, . _ -: $version"
    ;;
esac

name=narra-$version-$os-$arch$libc
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
  file://*)
    # A folder on this machine, for tests.
    protocols='=file'
    ;;
  *) fail "NARRA_BASE_URL must start with https:// (http:// only for localhost or 127.0.0.1): $base" ;;
esac
if [ -z "${NARRA_INSTALL:-}" ] && [ -z "${HOME:-}" ]; then
  fail "HOME is not set. Set NARRA_INSTALL to the folder that receives narrac"
fi
install_dir=${NARRA_INSTALL:-$HOME/.local/bin}

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

echo "install: narra $version for $os $arch${libc:+ (musl)}"
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
echo "install: open a new terminal, then run: narrac --version"
