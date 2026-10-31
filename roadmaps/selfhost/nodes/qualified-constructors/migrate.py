#!/usr/bin/env python3
"""Qualify enum constructors with the compiler's own fixes (D71).

Run from the repo root, with a built compiler:
  (cd compiler && go build -o /tmp/narra ./cmd/narra)
  python3 roadmaps/selfhost/nodes/qualified-constructors/migrate.py /tmp/narra

For every package (a folder with narra.toml) and every single .nra file under
compiler/testdata, selfhost and bench, it runs `narra check --json`, applies the
fixes of E0323 (bare constructor) and E0425 (constructor import), and checks
again until none is left. Edits that repeat (generic instances report a site
once per instance) are applied once. Idempotent: a migrated tree has no such
errors, so nothing changes.
"""
import json, os, subprocess, sys

ROOTS = ["compiler/testdata", "selfhost", "bench"]
CODES = {"E0323", "E0425"}


def units():
    pkgs = []
    for root in ROOTS:
        for d, _, fs in os.walk(root):
            if "narra.toml" in fs:
                pkgs.append(d)
    singles = []
    for root in ROOTS:
        for d, _, fs in os.walk(root):
            if any(d == p or d.startswith(p + os.sep) for p in pkgs):
                continue
            for f in sorted(fs):
                if f.endswith(".nra"):
                    singles.append(os.path.join(d, f))
    return sorted(pkgs) + sorted(singles)


def collect(narra, unit):
    out = subprocess.run([narra, "check", "--json", unit], capture_output=True, text=True).stdout
    edits = {}
    for line in out.splitlines():
        try:
            d = json.loads(line)
        except ValueError:
            continue
        if d.get("code") not in CODES:
            continue
        for fix in d.get("fixes") or []:
            for e in fix["edits"]:
                if os.path.isdir(unit):
                    path = os.path.join(unit, e["file"])
                else:
                    path = unit
                key = (e["range"]["start"][0], e["range"]["start"][1], e["range"]["end"][0], e["range"]["end"][1], e["text"])
                edits.setdefault(path, set()).add(key)
            break
    return edits


def apply(path, edits):
    lines = open(path, "rb").read().split(b"\n")
    for sl, sc, el, ec, text in sorted(edits, reverse=True):
        if sl != el:
            raise SystemExit("multi-line edit in " + path)
        l = lines[sl - 1]
        lines[sl - 1] = l[: sc - 1] + text.encode() + l[ec - 1 :]
    open(path, "wb").write(b"\n".join(lines))


def main():
    narra = os.path.abspath(sys.argv[1])
    files = 0
    for unit in units():
        touched = set()
        for _ in range(12):
            edits = collect(narra, unit)
            if not edits:
                break
            for path, es in edits.items():
                apply(path, es)
                touched.add(path)
        files += len(touched)
    print(files, "files changed", file=sys.stderr)


main()
