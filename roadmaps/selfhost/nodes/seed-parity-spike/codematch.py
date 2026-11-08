#!/usr/bin/env python3
"""Throwaway: can a code be found from the final message alone?

usage: codematch.py <diag.json from diagdiff.py>

Turns each format of diagnostic_codes.go into a regex (%s %q %d %v %c %T
become wildcards) and matches every Go diagnostic message of the samples.
Counts messages whose only matching format has Go's code.
"""
import json, os, re, sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "../../../.."))
src = open(os.path.join(ROOT, "compiler/internal/narra/diagnostic_codes.go")).read()
table = src[src.index("var diagnosticCodes"):src.index("func codeFor")]
rows = []
for code, fmt in re.findall(r'\{"(E\d{4})", "((?:[^"\\]|\\.)*)"\}', table):
    fmt = bytes(fmt, "utf-8").decode("unicode_escape")
    pattern = ""
    i = 0
    while i < len(fmt):
        if fmt[i] == "%" and i + 1 < len(fmt):
            j = i + 1
            while fmt[j] in "0123456789.+-#":
                j += 1
            if fmt[j] == "%":
                pattern += "%"
            else:
                pattern += "(.+?)"
            i = j + 1
            continue
        pattern += re.escape(fmt[i])
        i += 1
    rows.append((code, re.compile("^" + pattern + "$", re.S)))
print("formats:", len(rows), "distinct codes:", len({c for c, _ in rows}))

d = json.load(open(sys.argv[1]))
exact = ambiguous_ok = wrong = none = 0
examples = []
for suite in ("errors", "diagnostics", "packages/errors"):
    for rec in d[suite]:
        for line, code in zip(rec["go"], rec["codes"]):
            message = line.split(": ", 1)[1] if ": " in line else line
            # Package errors may carry a "narra.toml:" or path prefix.
            hits = [c for c, r in rows if r.match(message)]
            if hits == [code]:
                exact += 1
            elif code in hits:
                ambiguous_ok += 1
                examples.append(("ambiguous", rec["file"], code, sorted(set(hits)), message[:90]))
            elif hits:
                wrong += 1
                examples.append(("wrong", rec["file"], code, hits, message[:90]))
            else:
                none += 1
                examples.append(("none", rec["file"], code, [], message[:90]))
print(f"unique and right: {exact}, right among several: {ambiguous_ok}, wrong only: {wrong}, no match: {none}")
for e in examples[:40]:
    print(e)
