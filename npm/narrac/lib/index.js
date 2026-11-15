"use strict";

const { findRunner, runCaptured } = require("./run.js");

class NarracError extends Error {
  constructor(message, status) {
    super(message);
    this.name = "NarracError";
    this.status = status;
  }
}

function toList(files) {
  return typeof files === "string" ? [files] : files;
}

function parseDiagnostics(file, result) {
  // `check --json` prints one JSON object per line on stdout.
  const diagnostics = [];
  for (const line of result.stdout.split("\n")) {
    if (line.trim() !== "") {
      diagnostics.push(JSON.parse(line));
    }
  }
  if (diagnostics.length > 0) {
    return diagnostics;
  }
  // A problem that is not a diagnostic (an unreadable file) is text on stderr.
  const message = result.stderr.trim();
  return [
    {
      code: "",
      severity: "error",
      message,
      file,
      range: { start: [0, 0], end: [0, 0] },
      fixes: [],
    },
  ];
}

function check(files) {
  const diagnostics = [];
  for (const file of toList(files)) {
    // One narrac process per file, like `narrac check --json file`.
    const result = runCaptured(["check", "--json", file]);
    if (result.status === 0) {
      continue;
    }
    if (result.status !== 1) {
      throw new NarracError(result.stderr.trim(), result.status);
    }
    diagnostics.push(...parseDiagnostics(file, result));
  }
  return diagnostics;
}

function format(source) {
  // Native narrac fails on an empty stdin file; wasm does not. Answer both the same way.
  if (source === "") {
    return "";
  }
  const result = runCaptured(["fmt", "--stdin"], { input: source });
  if (result.status !== 0) {
    throw new NarracError((result.stderr || result.stdout).trim(), result.status);
  }
  return result.stdout;
}

function emit(files) {
  const results = [];
  for (const file of toList(files)) {
    const result = runCaptured(["emit", file]);
    if (result.status === 2) {
      throw new NarracError(result.stderr.trim(), result.status);
    }
    results.push({ file, status: result.status, c: result.stdout, stderr: result.stderr });
  }
  return results;
}

function run(args, options) {
  return runCaptured(args, options);
}

function describeRunner() {
  return findRunner();
}

module.exports = { check, format, emit, run, describeRunner, NarracError };
