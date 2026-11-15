"use strict";

// Compares the installed narrac package with a native narrac.
//
//   node npm/test/compare.js <project> <native narrac> <native|wasm> <repository>
//
// <project> is a folder where `npm install` put narrac. The command (`npx narrac`) and each
// function of the API must give what the native compiler gives for the same input.
// Exit 0 when every case matches.

const assert = require("node:assert");
const childProcess = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");

const [project, nativeNarrac, expectedKind, repository] = process.argv.slice(2);
if (!repository) {
  console.error(
    "usage: node npm/test/compare.js <project> <native narrac> <native|wasm> <repository>",
  );
  process.exit(2);
}

// Inside the project, so `npx narrac` finds node_modules/.bin from here.
const scratch = fs.mkdtempSync(path.join(project, "compare-"));
const badFile = path.join(scratch, "bad.nra");
fs.writeFileSync(
  badFile,
  'export function run(args: Array<string>): int {\n  const x: int = "a"\n  return 0\n}\n',
);
const helloFile = path.join(repository, "testdata", "hello.nra");
const hostFile = path.join(repository, "testdata", "build_host_target.nra");
const messyText = "export function   run(args: Array<string>):int{return 0}\n";
const brokenText = "export function run( {";

const api = require(require.resolve("narrac", { paths: [project] }));
let failures = 0;
let passes = 0;

function runNative(args, input) {
  const result = childProcess.spawnSync(nativeNarrac, args, {
    cwd: scratch,
    input: input === undefined ? "" : input,
    encoding: "utf8",
    maxBuffer: 1 << 30,
  });
  return { status: result.status, stdout: result.stdout, stderr: result.stderr };
}

function runCommand(args, input) {
  const result = childProcess.spawnSync("npx", ["--no-install", "narrac", ...args], {
    cwd: scratch,
    input: input === undefined ? "" : input,
    encoding: "utf8",
    maxBuffer: 1 << 30,
  });
  return { status: result.status, stdout: result.stdout, stderr: result.stderr };
}

function recordCase(name, body) {
  try {
    body();
    passes++;
    console.log(`ok   ${name}`);
  } catch (error) {
    failures++;
    console.log(`FAIL ${name}\n${error.message}`);
  }
}

function compareCommand(name, args, input) {
  recordCase(`command: ${name}`, () => {
    // Same arguments, same stdin: the status and both streams must be equal.
    const want = runNative(args, input);
    const got = runCommand(args, input);
    assert.deepStrictEqual(got, want);
  });
}

function parseNativeDiagnostics(file) {
  const result = runNative(["check", "--json", file]);
  const diagnostics = [];
  for (const line of result.stdout.split("\n")) {
    if (line.trim() !== "") {
      diagnostics.push(JSON.parse(line));
    }
  }
  return diagnostics;
}

recordCase(`runner is ${expectedKind}`, () => {
  assert.strictEqual(api.describeRunner().kind, expectedKind);
});

compareCommand("check clean file", ["check", helloFile]);
compareCommand("check bad file", ["check", badFile]);
compareCommand("check --json bad file", ["check", "--json", badFile]);
compareCommand("check missing file", ["check", path.join(scratch, "missing.nra")]);
compareCommand("check relative path", ["check", "bad.nra"]);
compareCommand("check --all-targets", ["check", "--all-targets", helloFile]);
compareCommand("targets", ["targets"]);
compareCommand("explain", ["explain", "E0119"]);
compareCommand("fmt --stdin", ["fmt", "--stdin"], messyText);
compareCommand("fmt --stdin broken", ["fmt", "--stdin"], brokenText);
compareCommand("emit hello", ["emit", helloFile]);
compareCommand("emit std/build program", ["emit", hostFile]);
compareCommand("emit bad file", ["emit", badFile]);
compareCommand("no arguments", []);

recordCase("api: check clean file is empty", () => {
  assert.deepStrictEqual(api.check(helloFile), []);
});

recordCase("api: check bad file gives the native diagnostics", () => {
  assert.deepStrictEqual(api.check([badFile]), parseNativeDiagnostics(badFile));
});

recordCase("api: check many files keeps the order", () => {
  const got = api.check([badFile, helloFile, badFile]);
  assert.deepStrictEqual(got, [
    ...parseNativeDiagnostics(badFile),
    ...parseNativeDiagnostics(badFile),
  ]);
});

recordCase("api: check missing file is one diagnostic without a code", () => {
  const missing = path.join(scratch, "missing.nra");
  const got = api.check(missing);
  assert.strictEqual(got.length, 1);
  assert.strictEqual(got[0].code, "");
  assert.strictEqual(got[0].message, runNative(["check", missing]).stderr.trim());
});

recordCase("api: format matches native fmt", () => {
  assert.strictEqual(api.format(messyText), runNative(["fmt", "--stdin"], messyText).stdout);
});

recordCase("api: format of broken source throws", () => {
  assert.throws(
    () => api.format(brokenText),
    (error) => error.name === "NarracError" && error.status === 1,
  );
});

recordCase("api: emit matches native emit", () => {
  const got = api.emit([helloFile, hostFile, badFile]);
  const files = [helloFile, hostFile, badFile];
  assert.strictEqual(got.length, files.length);
  for (let index = 0; index < files.length; index++) {
    const want = runNative(["emit", files[index]]);
    assert.deepStrictEqual(
      {
        file: got[index].file,
        status: got[index].status,
        c: got[index].c,
        stderr: got[index].stderr,
      },
      { file: files[index], status: want.status, c: want.stdout, stderr: want.stderr },
    );
  }
});

recordCase("api: run passes arguments through", () => {
  assert.deepStrictEqual(api.run(["targets"]), runNative(["targets"]));
});

fs.rmSync(scratch, { recursive: true, force: true });
console.log(`${passes} passed, ${failures} failed`);
process.exit(failures === 0 ? 0 : 1);
