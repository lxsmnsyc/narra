#!/usr/bin/env python3
"""Throwaway: run the program suites through stage 2 the way the Go tests do.

usage: rundiff.py <bin dir> <work dir>

For each program: `narrac2 emit` (with --tests where TestPrograms sets
Tests), compare the C with `narra emit`, build it with the debug flags
(-O0 -g -std=c11 -w -fsanitize=address,undefined -fno-omit-frame-pointer)
and the release flags (-O2 -DNARRA_RELEASE -std=c11 -w), run it with
NARRA_CHECK_LEAKS=1 and check the result like the Go test.
"""
import concurrent.futures, glob, os, subprocess, sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "../../../.."))
BIN, WORK = sys.argv[1], sys.argv[2]
os.makedirs(WORK, exist_ok=True)
os.chdir(ROOT)
TD = "compiler/testdata"
env = dict(os.environ, NARRA_CHECK_LEAKS="1", ASAN_OPTIONS="detect_leaks=0")
DEBUG = ["-O0", "-g", "-std=c11", "-w", "-fsanitize=address,undefined", "-fno-omit-frame-pointer"]
RELEASE = ["-O2", "-DNARRA_RELEASE", "-std=c11", "-w"]


def emit(path, tests, compiler):
    args = [f"{BIN}/{compiler}", "emit"]
    if tests:
        args.append("--tests" if compiler == "narrac2" else "")
    if compiler == "narra" and tests:
        return None  # Go emit has no test mode; compare skipped
    p = subprocess.run(args + [path], capture_output=True, text=True)
    return p.returncode, p.stdout, p.stderr


def build_run(name, c, flags, run_args=(), stdin=None):
    cfile = os.path.join(WORK, name + ".c")
    exe = os.path.join(WORK, name)
    open(cfile, "w").write(c)
    b = subprocess.run(["cc"] + flags + ["-o", exe, cfile, "-lm"], capture_output=True, text=True)
    if b.returncode != 0:
        return None, "", "cc failed: " + b.stderr[:300]
    inp = open(stdin) if stdin and os.path.exists(stdin) else subprocess.DEVNULL
    r = subprocess.run([exe] + list(run_args), capture_output=True, text=True, env=env, stdin=inp, timeout=300)
    return r.returncode, r.stdout, r.stderr


def program(src):
    base = os.path.basename(src)
    tests = base == "tests.nra" or base.startswith("std_test") or base.startswith("coroutine_tests")
    want = open(src[:-4] + ".out").read()
    code, c, err = emit(src, tests, "narrac2")
    if code != 0:
        return src, ["emit failed: " + err[:300]]
    problems = []
    if not tests:
        g = emit(src, False, "narra")
        if g[1] != c:
            problems.append("C differs from Go")
    for mode, flags in (("debug", DEBUG), ("release", RELEASE)):
        rc, out, rerr = build_run(base + "." + mode, c, flags, stdin=src[:-4] + ".in")
        if rc != 0 or out != want:
            problems.append(f"{mode}: exit {rc} out-match {out == want} {rerr[:200]}")
    return src, problems


PANICS = {
    "bounds.nra": "index 5 out of range for length 2",
    "overflow.nra": "integer overflow",
    "divzero.nra": "division by zero",
    "strings2_inside.nra": "position 2 is inside a character",
    "strings2_range.nra": "position 9 out of range for length 3",
    "strings2_charat_end.nra": "Str.charAt: position 2 is at the end of the string",
    "strings2_nextchar_past.nra": "Str.nextChar: position 5 out of range for length 2",
    "strings2_byteat_range.nra": "Str.byteAt: position 2 out of range for length 2",
    "strings2_split_empty.nra": "Str.split with an empty separator",
}


def panic(file):
    src = f"{TD}/panics/{file}"
    code, c, err = emit(src, False, "narrac2")
    if code != 0:
        return src, ["emit failed"]
    rc, out, rerr = build_run("panic_" + file, c, DEBUG)
    if rc != 101 or PANICS[file] not in rerr:
        return src, [f"exit {rc}: {rerr[:200]}"]
    return src, []


def package(dir):
    want = open(f"{dir}/run.out").read()
    name = "pkg_" + os.path.basename(dir)
    problems = []
    code, c, err = emit(dir, False, "narrac2")
    if code != 0:
        return dir, ["emit failed: " + err[:200]]
    g = emit(dir, False, "narra")
    if g[1] != c:
        problems.append("C differs from Go")
    for mode, flags in (("debug", DEBUG), ("release", RELEASE)):
        rc, out, rerr = build_run(name + "." + mode, c, flags)
        if rc != 0 or out != want:
            problems.append(f"{mode}: exit {rc} {rerr[:200]}")
    code, c, err = emit(dir, True, "narrac2")
    golden = os.path.abspath(f"{dir}/testdata")
    rc, out, rerr = build_run(name + ".tests", c, DEBUG, ["--golden-dir", golden])
    if rc != 0 or " passed, 0 failed" not in out:
        problems.append(f"tests: exit {rc} {out[-200:]} {rerr[:200]}")
    return dir, problems


def runner():
    src = f"{TD}/runner/mixed.nra"
    want = open(f"{TD}/runner/mixed.out").read()
    code, c, err = emit(src, True, "narrac2")
    problems = []
    for mode, flags in (("debug", DEBUG), ("release", RELEASE)):
        rc, out, rerr = build_run("runner." + mode, c, flags, ["--golden-dir", os.path.abspath(f"{TD}/runner")])
        if rc != 1 or out != want:
            problems.append(f"{mode}: exit {rc}")
    return src, problems


jobs = []
with concurrent.futures.ThreadPoolExecutor(max_workers=os.cpu_count()) as pool:
    progs = sorted(glob.glob(f"{TD}/*.nra"))
    jobs += [("programs", pool.submit(program, p)) for p in progs]
    stage2 = sorted(glob.glob(f"{TD}/stage2/*.nra"))
    jobs += [("stage2", pool.submit(program, p)) for p in stage2]
    jobs += [("panics", pool.submit(panic, f)) for f in sorted(PANICS)]
    pkgs = sorted(os.path.dirname(p) for p in glob.glob(f"{TD}/packages/*/run.out"))
    jobs += [("packages", pool.submit(package, d)) for d in pkgs]
    jobs.append(("runner", pool.submit(runner)))
    counts = {}
    for suite, job in jobs:
        path, problems = job.result()
        ok, total = counts.get(suite, (0, 0))
        counts[suite] = (ok + (not problems), total + 1)
        if problems:
            print("FAIL", suite, path, problems)
for suite, (ok, total) in counts.items():
    print(f"{suite}: {ok}/{total} pass")
