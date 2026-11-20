# Install a prebuilt narrac on Windows from a GitHub release.
#   irm https://github.com/lxsmnsyc/narra/releases/latest/download/install.ps1 | iex
# It downloads the Windows archive, checks its SHA-256 against the checksum file of the
# release, copies narrac.exe to %LOCALAPPDATA%\narra\bin and adds that folder to your PATH.
# Variables (environment):
#   NARRA_VERSION   the version to install, such as 0.1.0 or v0.1.0 (default: the latest release;
#                   when there is only a prerelease, the newest release of any kind)
#   NARRA_INSTALL   the folder that receives narrac.exe (default: %LOCALAPPDATA%\narra\bin)
#   NARRA_NO_PATH   set to 1 to leave the PATH alone
#   NARRA_BASE_URL  the folder that holds the archive and its .sha256 file
#                   (default: the release of NARRA_VERSION on GitHub); for tests
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

function Stop-Install([string]$message) {
  throw "install: $message"
}

$repository = 'lxsmnsyc/narra'

# Windows PowerShell 5.1 does not use TLS 1.2 by default, and GitHub requires it.
try {
  [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12
} catch {
}

# Only 64-bit x86 Windows has a prebuilt binary.
$architecture = $env:PROCESSOR_ARCHITEW6432
if (-not $architecture) { $architecture = $env:PROCESSOR_ARCHITECTURE }
if ($architecture -ne 'AMD64') {
  Stop-Install "no prebuilt binary for $architecture. Build from source: docs/install.md"
}

# Find the version. GitHub's latest release skips prereleases, so when it finds none,
# ask for the newest release of any kind.
$version = $env:NARRA_VERSION
if (-not $version) {
  try {
    $latest = Invoke-RestMethod -Uri "https://api.github.com/repos/$repository/releases/latest" -UseBasicParsing
    $version = $latest.tag_name
  } catch {
    $version = $null
  }
  if (-not $version) {
    try {
      $newest = Invoke-RestMethod -Uri "https://api.github.com/repos/$repository/releases?per_page=1" -UseBasicParsing
      $version = @($newest)[0].tag_name
    } catch {
      $version = $null
    }
  }
  if (-not $version) {
    Stop-Install "there is no release yet at https://github.com/$repository/releases"
  }
}
$version = $version -replace '^v', ''
if ($version -notmatch '^[0-9A-Za-z._-]+$') {
  Stop-Install "the version may hold only letters, digits, . _ -: $version"
}

$name = "narra-$version-windows-x86_64"
$base = $env:NARRA_BASE_URL
if (-not $base) { $base = "https://github.com/$repository/releases/download/v$version" }
# A custom base URL must use https, except for a test server on this machine.
if ($base -notmatch '^https://' -and $base -notmatch '^http://(localhost|127\.0\.0\.1)(:[0-9]+)?(/|$)') {
  Stop-Install "NARRA_BASE_URL must start with https:// (http:// only for localhost or 127.0.0.1): $base"
}

$installDir = $env:NARRA_INSTALL
if (-not $installDir) {
  if (-not $env:LOCALAPPDATA) { Stop-Install 'LOCALAPPDATA is not set. Set NARRA_INSTALL to the folder that receives narrac.exe' }
  $installDir = Join-Path $env:LOCALAPPDATA 'narra\bin'
}

$work = Join-Path ([System.IO.Path]::GetTempPath()) ("narra-install-" + [System.Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $work | Out-Null
try {
  Write-Host "install: narra $version for windows x86_64"
  $archive = Join-Path $work "$name.zip"
  $sumFile = Join-Path $work "$name.zip.sha256"
  try {
    Invoke-WebRequest -Uri "$base/$name.zip" -OutFile $archive -UseBasicParsing
  } catch {
    Stop-Install "could not download $base/$name.zip"
  }
  try {
    Invoke-WebRequest -Uri "$base/$name.zip.sha256" -OutFile $sumFile -UseBasicParsing
  } catch {
    Stop-Install "could not download $base/$name.zip.sha256"
  }

  # Check the SHA-256 before anything is unpacked.
  $line = ''
  $firstLine = @(Get-Content -Path $sumFile -TotalCount 1)
  if ($firstLine.Count -gt 0 -and $firstLine[0]) { $line = ([string]$firstLine[0]).Trim() }
  if (-not $line) { Stop-Install "the checksum file of $name.zip is empty" }
  $parts = $line -split '\s+', 2
  $expected = $parts[0].ToLowerInvariant()
  $listed = ''
  if ($parts.Count -gt 1) { $listed = $parts[1].TrimStart('*') }
  if ($expected -notmatch '^[0-9a-f]{64}$' -or -not $listed) {
    Stop-Install "the checksum file of $name.zip is malformed"
  }
  if ($listed -ne "$name.zip") {
    Stop-Install "the checksum file names '$listed', not $name.zip"
  }
  $actual = (Get-FileHash -Path $archive -Algorithm SHA256).Hash.ToLowerInvariant()
  if (-not $expected -or $expected -ne $actual) {
    Stop-Install "the checksum of $name.zip does not match (expected $expected, got $actual)"
  }

  Expand-Archive -Path $archive -DestinationPath $work -Force
  $binary = Join-Path $work "$name\narrac.exe"
  if (-not (Test-Path $binary)) { Stop-Install 'the archive holds no narrac.exe' }
  New-Item -ItemType Directory -Path $installDir -Force | Out-Null
  $staged = Join-Path $installDir 'narrac.exe.new'
  Copy-Item -Path $binary -Destination $staged -Force
  Move-Item -Path $staged -Destination (Join-Path $installDir 'narrac.exe') -Force
} finally {
  Remove-Item -Path $work -Recurse -Force -ErrorAction SilentlyContinue
}

Write-Host "install: narrac.exe is in $installDir"

# Add the folder to the PATH of the user, once.
if ($env:NARRA_NO_PATH -ne '1') {
  $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
  $entries = @()
  if ($userPath) { $entries = $userPath -split ';' }
  $normalized = $installDir.TrimEnd('\')
  $present = $false
  foreach ($entry in $entries) {
    if ($entry.TrimEnd('\') -ieq $normalized) { $present = $true }
  }
  if (-not $present) {
    $newPath = if ($userPath) { "$userPath;$installDir" } else { $installDir }
    [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
    Write-Host "install: added $installDir to your PATH"
  }
}
Write-Host 'install: open a new terminal, then run: narrac --version'
