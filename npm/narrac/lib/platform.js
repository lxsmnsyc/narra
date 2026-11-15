"use strict";

// One row per package that holds a native narrac. `target` is the name `narrac targets` prints
// for the row. A platform with no row runs the wasm package.
const NATIVE_ROWS = [
  {
    platform: "darwin",
    arch: "arm64",
    libc: "",
    target: "MacOS-Arm64",
    package: "narrac-darwin-arm64",
  },
  {
    platform: "darwin",
    arch: "x64",
    libc: "",
    target: "MacOS-X86_64",
    package: "narrac-darwin-x64",
  },
  {
    platform: "linux",
    arch: "arm64",
    libc: "glibc",
    target: "Linux-Arm64",
    package: "narrac-linux-arm64",
  },
  {
    platform: "linux",
    arch: "x64",
    libc: "glibc",
    target: "Linux-X86_64",
    package: "narrac-linux-x64",
  },
  {
    platform: "linux",
    arch: "arm64",
    libc: "musl",
    target: "Linux-Arm64-musl",
    package: "narrac-linux-arm64-musl",
  },
  {
    platform: "linux",
    arch: "x64",
    libc: "musl",
    target: "Linux-X86_64-musl",
    package: "narrac-linux-x64-musl",
  },
  {
    platform: "win32",
    arch: "x64",
    libc: "",
    target: "Windows-X86_64",
    package: "narrac-win32-x64",
  },
];

function findExecutableName(row) {
  return row.platform === "win32" ? "narrac.exe" : "narrac";
}

function detectLibc(platform) {
  // Step 1: only Linux has two C libraries.
  if (platform !== "linux") {
    return "";
  }
  // Step 2: the Node report says which one this process runs on.
  const report = process.report && process.report.getReport();
  if (report && report.header && report.header.glibcVersionRuntime) {
    return "glibc";
  }
  return "musl";
}

function findNativeRow(platform, arch, libc) {
  for (const row of NATIVE_ROWS) {
    if (row.platform === platform && row.arch === arch && row.libc === libc) {
      return row;
    }
  }
  return undefined;
}

function findHostRow() {
  return findNativeRow(process.platform, process.arch, detectLibc(process.platform));
}

function findHostTarget() {
  // Step 1: a wasm narrac cannot detect its host, so the loader passes NARRA_HOST_TARGET.
  // Step 2: an unmapped platform passes nothing, and narrac then reports E0989 only for a
  // program that imports std/build.
  const row = findHostRow();
  return row ? row.target : undefined;
}

module.exports = { NATIVE_ROWS, findExecutableName, detectLibc, findNativeRow, findHostRow, findHostTarget };
