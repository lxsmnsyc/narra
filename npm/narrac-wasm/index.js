"use strict";

const path = require("node:path");

// The path of the WASI preview1 module. The narrac package runs it.
module.exports = { wasmPath: path.join(__dirname, "narrac.wasm") };
