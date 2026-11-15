#!/usr/bin/env node
"use strict";

const { runInherited } = require("../lib/run.js");

// Run narrac with this process's arguments and exit with its status.
process.exitCode = runInherited(process.argv.slice(2));
