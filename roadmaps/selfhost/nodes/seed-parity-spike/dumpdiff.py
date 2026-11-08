#!/usr/bin/env python3
"""Throwaway: compare stage 2 dumps with the dump goldens (which Go writes).

usage: dumpdiff.py <bin dir>

A golden X.ast or X.dump belongs to X.nra or the directory X/ beside it. On
exit 0 stdout must equal the golden, otherwise stderr must.
"""
import glob, os, subprocess, sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "../../../.."))
N2 = sys.argv[1] + "/narrac2"
os.chdir(ROOT)
TD = "compiler/testdata"
suites = [
    ("astdump", "parse", [f"{TD}/astdump/*.ast"]),
    ("declsdump", "decls", [f"{TD}/declsdump/*.dump", f"{TD}/declsdump/*/*.dump"]),
    ("typesdump", "types", [f"{TD}/typesdump/*/*.dump"]),
    ("pkgdump", "pkg", [f"{TD}/pkgdump/*.dump"]),
]
for name, command, patterns in suites:
    ok = total = 0
    for pattern in patterns:
        for golden in sorted(glob.glob(pattern)):
            stem = golden.rsplit(".", 1)[0]
            target = stem + ".nra" if os.path.exists(stem + ".nra") else stem
            if not os.path.exists(target):
                print("NO INPUT", golden)
                continue
            p = subprocess.run([N2, command, target], capture_output=True, text=True)
            got = p.stdout if p.returncode == 0 else p.stderr
            total += 1
            if got == open(golden).read():
                ok += 1
            else:
                print("DIFF", golden, "exit", p.returncode, repr(got[:120]))
    print(f"{name}: {ok}/{total} match")
