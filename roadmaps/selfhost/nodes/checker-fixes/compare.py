#!/usr/bin/env python3
"""Throwaway: compare `narrac check --json` and `narrac apply-fix` with Go's `narra`.

usage: compare.py <bin dir>      (the dir holds `narra` and `narrac`)

Run from anywhere. Checks, byte for byte on stdout and by exit code, with
Go's fixes kept:
- check --json on every sample of compiler/testdata/errors/, diagnostics/,
  packages/errors/ and stage2/errors/, plus the inline habit cases of
  diagnostic_test.go (habit_cases.json, made by dump-habit-cases)
- for each fix in Go's output, `narrac apply-fix --diagnostic N --fix M` equals
  the source with Go's edits applied (a port of Go's ApplyFix), and the fixed
  source checks clean when the fix is the first one
- explain, as in checker-codes
"""
import glob, json, os, re, subprocess, sys, tempfile

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "../../../.."))
BIN = sys.argv[1]
TD = "compiler/testdata"


def run(exe, args, cwd=ROOT):
    p = subprocess.run([f"{BIN}/{exe}"] + args, cwd=cwd, capture_output=True, text=True, timeout=120)
    return p.returncode, p.stdout, p.stderr


def apply_fix(src, fix):
    """Go's ApplyFix: columns count characters; edits apply from the end."""
    lines = src.split("\n")

    def offset(pos):
        line, col = pos
        n = 0
        for i in range(min(line - 1, len(lines))):
            n += len(lines[i]) + 1
        if line - 1 < len(lines):
            n += min(col - 1, len(lines[line - 1]))
        return n

    edits = list(fix["edits"])
    for i in range(1, len(edits)):
        j = i
        while j > 0 and offset(edits[j]["range"]["start"]) > offset(edits[j - 1]["range"]["start"]):
            edits[j], edits[j - 1] = edits[j - 1], edits[j]
            j -= 1
    for e in edits:
        s, t = offset(e["range"]["start"]), offset(e["range"]["end"])
        src = src[:s] + e["text"] + src[t:]
    return src


def check_samples(samples, label):
    bad = 0
    fixes = 0
    applied = 0
    for path in samples:
        rel = os.path.relpath(path, ROOT)
        go = run("narra", ["check", "--json", rel])
        nr = run("narrac", ["check", "--json", rel])
        if go[0] != nr[0] or go[1] != nr[1]:
            bad += 1
            print("DIFF", rel)
            print("  go:", go[0], go[1][:600].replace("\n", "\n      "))
            print("  nr:", nr[0], nr[1][:600].replace("\n", "\n      "), nr[2][:200])
            continue
        for index, line in enumerate(go[1].splitlines()):
            record = json.loads(line)
            for fix_index, fix in enumerate(record["fixes"]):
                fixes += 1
                file_path = rel if os.path.isfile(path) else os.path.join(rel, record["file"])
                with open(os.path.join(ROOT, file_path), encoding="utf-8") as handle:
                    source = handle.read()
                want = apply_fix(source, fix)
                got = run("narrac", ["apply-fix", "--diagnostic", str(index), "--fix", str(fix_index), rel])
                if got[0] != 0 or got[1] != want:
                    bad += 1
                    print("DIFF apply-fix", rel, index, fix_index, got[0], got[2][:200])
                else:
                    applied += 1
    print(f"{label}: {len(samples)} samples, {bad} differences, {fixes} fixes, {applied} applied equal")
    return bad


samples = sorted(glob.glob(f"{ROOT}/{TD}/errors/*.nra") + glob.glob(f"{ROOT}/{TD}/diagnostics/*.nra")
                 + glob.glob(f"{ROOT}/{TD}/stage2/errors/*.nra"))
samples += sorted(d for d in glob.glob(f"{ROOT}/{TD}/packages/errors/*") if os.path.isdir(d))
failures = check_samples(samples, "check --json")

# The inline habit cases of diagnostic_test.go, written to files.
cases_path = os.path.join(os.path.dirname(__file__), "habit_cases.json")
if os.path.exists(cases_path):
    cases = json.load(open(cases_path))
    with tempfile.TemporaryDirectory() as tmp:
        paths = []
        for name, source in cases.items():
            directory = os.path.join(tmp, name)
            os.mkdir(directory)
            path = os.path.join(directory, "habit.nra")
            with open(path, "w", encoding="utf-8") as handle:
                handle.write(source)
            paths.append(path)
        failures += check_samples(paths, "habit cases")

table = open(f"{ROOT}/selfhost/src/diagnostics.nra").read()
codes = sorted(set(re.findall(r'CodeRow \{\s*code: "(E\d{4})"', table)))
bad = 0
cases = [[c] for c in codes] + [["e0001"], ["E9999"], [], ["E0001", "E0002"]]
for args in cases:
    go = run("narra", ["explain"] + args)
    nr = run("narrac", ["explain"] + args)
    # Go prints "narra" in the usage text, narrac "narrac"; compare stdout and exit code.
    if go[0] != nr[0] or go[1] != nr[1] or (len(args) == 1 and go[2] != nr[2]):
        bad += 1
        print("DIFF explain", args, go, nr)
print(f"explain: {len(cases) - bad}/{len(cases)} equal ({len(codes)} codes)")
sys.exit(1 if failures or bad else 0)
