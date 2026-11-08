#!/usr/bin/env python3
"""Fail when a report site passes no code, an unknown code, or a code whose
message format does not fit the message.

usage: check-codes.py        (from anywhere; reads selfhost/src/*.nra)

A report site is a call of one of REPORTERS. The compiler already rejects a
call with a missing argument, because every reporter takes the code as a
parameter before the message. This script adds what the types cannot see:
- the code argument is a string literal, a variable named `code`, or a
  `.code` field (a code passed on from a helper or another error);
- a literal code is in the table of diagnostics.nra and is not E0000;
- when the message is a string or template literal that matches a format of
  the table, the code is one of the codes with a matching format.
"""
import glob, os, re, sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "../../../.."))
SRC = f"{ROOT}/selfhost/src"
# reporter -> index of the code argument (counted from 0); the message follows.
REPORTERS = {
    "reportError": 2, "reportErrorIn": 3, "reportTypeError": 2,
    "failAt": 2, "failHere": 1, "failFound": 1,
    "errorAt": 2, "fail": 2, "createError": 1,
    "checkLoopExit": 4,
}

table = open(f"{SRC}/diagnostics.nra").read()
rows = []
for m in re.finditer(r'CodeRow \{ code: "(E\d{4})", format: "((?:\\.|[^"\\])*)",? \}', re.sub(r"\s*\n\s*", " ", table)):
    rows.append((m.group(1), re.sub(r"\\(.)", lambda mm: "\n" if mm.group(1) == "n" else mm.group(1), m.group(2))))
if len(rows) < 280:
    sys.exit(f"table not found in diagnostics.nra: {len(rows)} rows")
known = {c for c, _ in rows}


def fmt_regex(fmt):
    out = ""
    for p in re.split(r"(%[qsdvc])", fmt):
        if p == "%q":
            out += '(?:§q|"[^"]*")'
        elif re.fullmatch(r"%[sdvc]", p):
            out += "(?:(?!§q).)*?"
        else:
            out += re.escape(p)
    return re.compile(out, re.S)


regs = [(c, fmt_regex(f)) for c, f in rows]


def split_args(s):
    out, depth, instr, start, j = [], 0, None, 0, 0
    while j < len(s):
        c = s[j]
        if instr:
            if c == "\\":
                j += 1
            elif c == instr:
                instr = None
        elif c in "\"`":
            instr = c
        elif c in "([{":
            depth += 1
        elif c in ")]}":
            depth -= 1
        elif c == "," and depth == 0:
            out.append(s[start:j].strip())
            start = j + 1
        j += 1
    if s[start:].strip():
        out.append(s[start:].strip())
    return out


def literal_text(expr):
    m = re.fullmatch(r'`((?:\\.|\$\{(?:[^{}]|\{[^{}]*\})*\}|[^`\\$]|\$)*)`|"((?:\\.|[^"\\])*)"', expr)
    if not m:
        return None
    if m.group(1) is not None:
        t = re.sub(r"\$\{quoteGo\((?:[^{}]|\{[^{}]*\})*\}", "§q", m.group(1))
        t = re.sub(r"\$\{(?:[^{}]|\{[^{}]*\})*\}", "§", t)
    else:
        t = m.group(2)
    return re.sub(r"\\(.)", lambda mm: "\n" if mm.group(1) == "n" else mm.group(1), t)


# These local helpers also exist in fmt modules that do not report diagnostics.
LOCAL_HELPERS = {"failAt": "parser.nra", "failHere": "parser.nra", "failFound": "parser.nra",
                 "errorAt": "lexer.nra", "fail": "lexer.nra", "createError": "manifest.nra"}
problems = []
sites = 0
for path in sorted(glob.glob(f"{SRC}/*.nra")):
    name = os.path.basename(path)
    s = open(path).read()
    for m in re.finditer(r"\b(" + "|".join(REPORTERS) + r")\(", s):
        line_start = s.rfind("\n", 0, m.start()) + 1
        before = s[line_start:m.start()].strip()
        if before.startswith(("export function", "function")):
            continue
        depth, j, instr = 1, m.end(), None
        while depth and j < len(s):
            c = s[j]
            if instr:
                if c == "\\":
                    j += 1
                elif c == instr:
                    instr = None
            elif c in "\"`":
                instr = c
            elif c == "(":
                depth += 1
            elif c == ")":
                depth -= 1
            j += 1
        args = split_args(s[m.end():j - 1])
        line = s.count("\n", 0, m.start()) + 1
        where = f"{name}:{line}"
        if LOCAL_HELPERS.get(m.group(1), name) != name:
            continue
        index = REPORTERS[m.group(1)]
        sites += 1
        if len(args) <= index + 1:
            problems.append(f"{where}: {m.group(1)} has no code argument")
            continue
        code = args[index]
        literal = re.fullmatch(r'"(E\d{4})"', code)
        if literal:
            code = literal.group(1)
            if code == "E0000" or code not in known:
                problems.append(f"{where}: {m.group(1)} passes {code}, which is not in the table")
                continue
            text = literal_text(args[index + 1]) if len(args) > index + 1 else None
            if text is not None:
                fits = {c for c, r in regs if r.fullmatch(text)}
                fits.discard("E0409")
                if fits and code not in fits and code != "E0409":
                    problems.append(f"{where}: {code} does not fit the message; formats fit {sorted(fits)}")
        elif not re.fullmatch(r"(code|[\w.]+\.code)", code):
            problems.append(f"{where}: {m.group(1)} passes `{code}`, not a code literal or a code variable")

for problem in problems:
    print(problem)
print(f"{sites} report sites, {len(problems)} problems")
sys.exit(1 if problems else 0)
