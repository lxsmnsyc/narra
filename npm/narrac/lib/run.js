"use strict";

const childProcess = require("node:child_process");
const fs = require("node:fs");
const { findExecutableName, findHostRow, findHostTarget } = require("./platform.js");
const { runWasi } = require("./wasi.js");

let cachedModule;
let cachedKey = "";

function findRunner() {
  // NARRAC_WASM=1 picks the wasm package even where a native package is installed.
  if (process.env.NARRAC_WASM !== "1") {
    const row = findHostRow();
    if (row) {
      try {
        return { kind: "native", path: require.resolve(`${row.package}/bin/${findExecutableName(row)}`) };
      } catch (error) {
        // A missing package means optional dependencies were skipped: use wasm.
        if (error.code !== "MODULE_NOT_FOUND") {
          throw error;
        }
      }
    }
  }
  try {
    return { kind: "wasm", path: require("narrac-wasm").wasmPath };
  } catch (error) {
    if (error.code !== "MODULE_NOT_FOUND") {
      throw error;
    }
    throw new Error(
      `narrac: no compiler for ${process.platform}-${process.arch}. Install narrac-wasm, or the native package for this platform.`,
    );
  }
}

function loadModule(wasmPath) {
  // One compile per process; every call then makes a fresh instance.
  if (cachedModule === undefined || cachedKey !== wasmPath) {
    cachedModule = new WebAssembly.Module(fs.readFileSync(wasmPath));
    cachedKey = wasmPath;
  }
  return cachedModule;
}

function buildWasmEnvironment(base) {
  // The loader names the host, because a wasm narrac cannot detect it. A caller's value wins.
  const environment = { ...base };
  const target = findHostTarget();
  if (environment.NARRA_HOST_TARGET === undefined && target !== undefined) {
    environment.NARRA_HOST_TARGET = target;
  }
  return environment;
}

function runCaptured(args, options = {}) {
  const runner = findRunner();
  const cwd = options.cwd === undefined ? process.cwd() : options.cwd;
  const environment = options.env === undefined ? process.env : options.env;
  if (runner.kind === "native") {
    // A native binary: spawn it and collect both streams.
    const result = childProcess.spawnSync(runner.path, args, {
      cwd,
      env: environment,
      input: options.input === undefined ? "" : options.input,
      encoding: "utf8",
      maxBuffer: 1 << 30,
    });
    if (result.error) {
      throw result.error;
    }
    return {
      status: result.status === null ? 1 : result.status,
      stdout: result.stdout,
      stderr: result.stderr,
    };
  }
  // The wasm module: a fresh instance, with both streams collected in memory.
  const out = [];
  const err = [];
  const status = runWasi(loadModule(runner.path), {
    args: ["narrac", ...args],
    env: buildWasmEnvironment(environment),
    cwd,
    stdin: Buffer.from(options.input === undefined ? "" : options.input),
    writeStdout: (bytes) => out.push(Buffer.from(bytes)),
    writeStderr: (bytes) => err.push(Buffer.from(bytes)),
  });
  return {
    status,
    stdout: Buffer.concat(out).toString("utf8"),
    stderr: Buffer.concat(err).toString("utf8"),
  };
}

function runInherited(args) {
  const runner = findRunner();
  if (runner.kind === "native") {
    // A native binary: hand it the terminal.
    const result = childProcess.spawnSync(runner.path, args, { stdio: "inherit" });
    if (result.error) {
      throw result.error;
    }
    return result.status === null ? 1 : result.status;
  }
  // The wasm module writes to descriptors 1 and 2 as it goes.
  return runWasi(loadModule(runner.path), {
    args: ["narrac", ...args],
    env: buildWasmEnvironment(process.env),
    cwd: process.cwd(),
    stdin: undefined,
    writeStdout: (bytes) => writeAll(1, bytes),
    writeStderr: (bytes) => writeAll(2, bytes),
  });
}

function writeAll(descriptor, bytes) {
  let offset = 0;
  while (offset < bytes.length) {
    try {
      offset += fs.writeSync(descriptor, bytes, offset, bytes.length - offset);
    } catch (error) {
      if (error.code !== "EAGAIN") {
        throw error;
      }
    }
  }
}

module.exports = { findRunner, runCaptured, runInherited };
