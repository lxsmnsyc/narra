#!/usr/bin/env python3
"""Rename the std namespace aliases to their plural form (D72).

Run from the repo root: python3 roadmaps/selfhost/nodes/import-alias-rules/migrate.py
For each `import * as Array from "std/array"` (and Map, Box, Range, Option,
Result) the alias becomes Arrays, Maps, Boxes, Ranges, Options, Results, in the
import and in every `Alias.name` use of the file. Lines that start with `//`
are left alone, and so is a name right after a quote (`"Array.push"`). A file that declares its own type with the plural name is
skipped. Idempotent.
"""
import os, re, sys

PLURAL = {"Array": "Arrays", "Map": "Maps", "Box": "Boxes",
          "Range": "Ranges", "Option": "Options", "Result": "Results"}
PATH = {"Array": "std/array", "Map": "std/map", "Box": "std/box",
        "Range": "std/range", "Option": "std/option", "Result": "std/result"}
ROOTS = ["compiler/testdata", "selfhost", "bench"]


def migrate(path):
    text = open(path).read()
    new = text
    for ns, plural in PLURAL.items():
        imp = 'import * as %s from "%s"' % (ns, PATH[ns])
        if imp not in new.split("\n"):
            continue
        if re.search(r'^(export\s+)?(opaque\s+)?(struct|enum|distinct type|type|trait)\s+%s\b' % plural, new, re.M):
            print("skip (declares %s): %s" % (plural, path), file=sys.stderr)
            continue
        lines = new.split("\n")
        for i, l in enumerate(lines):
            if l.lstrip().startswith("//"):
                continue
            if l == imp:
                lines[i] = 'import * as %s from "%s"' % (plural, PATH[ns])
            else:
                lines[i] = re.sub(r'(?<![\w."`])%s\.(?=[A-Za-z_])' % ns, plural + ".", l)
        new = "\n".join(lines)
    if new != text:
        open(path, "w").write(new)
        return True
    return False


n = 0
for root in ROOTS:
    for d, _, fs in os.walk(root):
        for f in sorted(fs):
            if f.endswith(".nra") and migrate(os.path.join(d, f)):
                n += 1
print(n, "files", file=sys.stderr)
