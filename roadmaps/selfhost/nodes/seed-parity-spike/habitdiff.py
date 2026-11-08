#!/usr/bin/env python3
"""Throwaway: run the habitCases of diagnostic_test.go through Go and stage 2.

usage: habitdiff.py <bin dir> <work dir>
"""
import codecs, os, re, subprocess, sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "../../../.."))
BIN, WORK = sys.argv[1], sys.argv[2]
os.makedirs(WORK, exist_ok=True)
src = open(os.path.join(ROOT, "compiler/internal/narra/diagnostic_test.go")).read()
main = codecs.decode(re.search(r'const habitMain = "(.*)"', src).group(1), "unicode_escape")
table = src[src.index("var habitCases"):src.index("// The habit cases give")]
same = 0
total = 0
for m in re.finditer(r'\{"(\w+)", (.*?), "(E\d+)", (true|false)\},', table):
    name, expr, code = m.group(1), m.group(2), m.group(3)
    text = ""
    for part in re.findall(r'"((?:[^"\\]|\\.)*)"|habitMain', expr):
        text += codecs.decode(part, "unicode_escape") if part else ""
    if "habitMain" in expr:
        text += main
    path = os.path.join(WORK, name + ".nra")
    open(path, "w").write(text)
    go = subprocess.run([f"{BIN}/narra", "emit", path], capture_output=True, text=True).stderr.strip().splitlines()
    s2 = subprocess.run([f"{BIN}/narrac2", "emit", path], capture_output=True, text=True).stderr.strip().splitlines()
    s2 = [line[6:] if line.startswith("error ") else line for line in s2]
    go = [line[len(WORK) + 1:] if line.startswith(WORK) else line for line in go]
    total += 1
    if go == s2:
        same += 1
    else:
        print("DIFF", name, code, "\n  go:", go, "\n  s2:", s2)
print(f"habit cases: {same}/{total} identical")
