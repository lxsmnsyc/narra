#!/usr/bin/env python3
"""Add the std imports that Array.*, Map.*, Range.* and Box.* calls now need.

Run from the repo root: python3 roadmaps/selfhost/nodes/prelude-imports/migrate.py
Each file only gains lines: `import * as NS from "std/ns"` after its last
import (or before its first code line, followed by a blank line, when it has
no imports). A namespace is skipped in a file that declares its own type of
that name. Idempotent.
"""
import os, re, sys

NS = {"Array": "std/array", "Map": "std/map", "Range": "std/range", "Box": "std/box"}
ROOTS = ["compiler/testdata", "selfhost", "bench"]


def code_lines(lines):
    for l in lines:
        if l.lstrip().startswith("//"):
            continue
        yield l


def migrate(path):
    text = open(path).read()
    lines = text.split("\n")
    body = "\n".join(code_lines(lines))
    used = []
    for ns in NS:
        if not re.search(r'(?<![\w.])%s\.[a-z]\w*' % ns, body):
            continue
        if re.search(r'^(export\s+)?(opaque\s+)?(struct|enum|distinct type|type)\s+%s\b' % ns, body, re.M):
            continue
        if re.search(r'^import\b.*\bas %s\b' % ns, body, re.M):
            continue
        used.append(ns)
    if not used:
        return False
    new = ['import * as %s from "%s"' % (ns, NS[ns]) for ns in sorted(used)]
    last = -1
    i = 0
    while i < len(lines):
        if lines[i].startswith("import "):
            j = i
            while 'from "' not in lines[j]:
                j += 1
            last = j
            i = j
        i += 1
    if last >= 0:
        lines[last + 1:last + 1] = new
    else:
        k = 0
        while k < len(lines) and (lines[k].strip() == "" or lines[k].lstrip().startswith("//")):
            k += 1
        lines[k:k] = new + [""]
    open(path, "w").write("\n".join(lines))
    return True


n = 0
for root in ROOTS:
    for d, _, fs in os.walk(root):
        for f in sorted(fs):
            if f.endswith(".nra") and migrate(os.path.join(d, f)):
                n += 1
                print(os.path.join(d, f))
print(n, "files", file=sys.stderr)
