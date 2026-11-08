#!/usr/bin/env python3
"""Throwaway: compare `narrac check --json` and `narrac explain` with Go's `narra`.

usage: compare.py <bin dir>      (the dir holds `narra` and `narrac`)

Run from anywhere. Checks, byte for byte on stdout and by exit code:
- check --json on every sample of compiler/testdata/errors/, diagnostics/,
  packages/errors/ and stage2/errors/
- explain on every code of the table (selfhost/src/diagnostics.nra), on a
  lowercase code, on an unknown code, and with no argument
"""
import glob, json, os, re, subprocess, sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "../../../.."))
BIN = sys.argv[1]
TD = "compiler/testdata"


def run(exe, args):
    p = subprocess.run([f"{BIN}/{exe}"] + args, cwd=ROOT, capture_output=True, text=True, timeout=120)
    return p.returncode, p.stdout, p.stderr


samples = sorted(glob.glob(f"{ROOT}/{TD}/errors/*.nra") + glob.glob(f"{ROOT}/{TD}/diagnostics/*.nra")
                 + glob.glob(f"{ROOT}/{TD}/stage2/errors/*.nra"))
samples += sorted(d for d in glob.glob(f"{ROOT}/{TD}/packages/errors/*") if os.path.isdir(d))
bad = 0
withfix = 0
for path in samples:
    rel = os.path.relpath(path, ROOT)
    go = run("narra", ["check", "--json", rel])
    nr = run("narrac", ["check", "--json", rel])
    # Go's fixes belong to checker-fixes: compare with its fixes emptied. Lines
    # without fixes are then byte-equal to narrac's.
    want = []
    for line in go[1].splitlines():
        record = json.loads(line)
        if record["fixes"]:
            withfix += 1
        record["fixes"] = []
        want.append(json.dumps(record, separators=(",", ":"), ensure_ascii=False))
    if go[0] != nr[0] or "\n".join(want) != nr[1].rstrip("\n"):
        bad += 1
        print("DIFF", rel)
        print("  go:", go[0], go[1][:300].replace("\n", "\n      "))
        print("  nr:", nr[0], nr[1][:300].replace("\n", "\n      "), nr[2][:200])
print(f"check --json: {len(samples) - bad}/{len(samples)} equal ({withfix} Go diagnostics have fixes, left to checker-fixes)")

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
