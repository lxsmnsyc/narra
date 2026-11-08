#!/usr/bin/env python3
"""Throwaway: compare stage 2 diagnostics with the Go compiler's.

usage: diagdiff.py <bin dir> <out.json>

For each error sample, runs `narra emit` and `narra check --json` (Go) and
`narrac2 emit` (stage 2) from the repo root, and records:
- want: the `// error:` (or `# error:` of narra.toml) text
- contains: stage 2 stderr contains want (what TestCompileErrors checks)
- first: the first diagnostic line (file:line:col: message) equals Go's
- pos: the first diagnostic position equals Go's
- all: every diagnostic line equals Go's, in order
- codes: Go's codes (stage 2 has none)
"""
import glob, json, os, subprocess, sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "../../../.."))
BIN = sys.argv[1]
OUT = sys.argv[2]
env = dict(os.environ, NARRA_CHECK_LEAKS="1", ASAN_OPTIONS="detect_leaks=0")


def run(args):
    p = subprocess.run(args, cwd=ROOT, capture_output=True, text=True, env=env, timeout=120)
    return p.returncode, p.stdout, p.stderr


def lines(text, strip_error):
    out = []
    for line in text.splitlines():
        line = line.rstrip()
        if not line:
            continue
        if strip_error and line.startswith("error "):
            line = line[len("error "):]
        out.append(line)
    return out


def pos(line):
    # file:line:col: message -> file:line:col
    parts = line.split(": ", 1)
    return parts[0]


def want_text(path):
    if os.path.isdir(path):
        first = open(os.path.join(path, "narra.toml")).readline()
        return first.removeprefix("# error:").strip()
    first = open(path).readline()
    for prefix in ("// error:", "// code:"):
        if first.startswith(prefix):
            return first[len(prefix):].strip()
    return ""


def case(rel, go_ok=True):
    want = want_text(os.path.join(ROOT, rel))
    rec = {"file": rel, "want": want}
    code2, out2, err2 = run([f"{BIN}/narrac2", "emit", rel])
    rec["stage2_exit"] = code2
    s2 = lines(err2, True)
    rec["stage2"] = s2
    rec["contains"] = code2 == 1 and want != "" and want in err2
    if go_ok:
        codeg, outg, errg = run([f"{BIN}/narra", "emit", rel])
        g = lines(errg, False)
        rec["go_exit"] = codeg
        rec["go"] = g
        rec["go_contains"] = want != "" and want in errg
        rec["first"] = bool(g) and bool(s2) and g[0] == s2[0]
        rec["pos"] = bool(g) and bool(s2) and pos(g[0]) == pos(s2[0])
        rec["all"] = g == s2
        _, outj, _ = run([f"{BIN}/narra", "check", "--json", rel])
        codes = []
        for line in outj.splitlines():
            try:
                codes.append(json.loads(line)["code"])
            except Exception:
                pass
        rec["codes"] = codes
        fixes = 0
        for line in outj.splitlines():
            try:
                if json.loads(line).get("fixes"):
                    fixes += 1
            except Exception:
                pass
        rec["fixes"] = fixes
    return rec


os.chdir(ROOT)
suites = {
    "errors": sorted(glob.glob("compiler/testdata/errors/*.nra")),
    "diagnostics": sorted(glob.glob("compiler/testdata/diagnostics/*.nra")),
    "packages/errors": sorted(glob.glob("compiler/testdata/packages/errors/*")),
    "stage2/errors": sorted(glob.glob("compiler/testdata/stage2/errors/*.nra")),
}
result = {}
for name, files in suites.items():
    result[name] = [case(f, go_ok=(name != "stage2/errors")) for f in files]
    recs = result[name]
    n = len(recs)
    summary = {k: sum(1 for r in recs if r.get(k)) for k in ("contains", "go_contains", "first", "pos", "all")}
    print(name, n, summary)
json.dump(result, open(OUT, "w"), indent=1)
