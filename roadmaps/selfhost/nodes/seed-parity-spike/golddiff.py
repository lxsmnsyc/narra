#!/usr/bin/env python3
"""Throwaway: run fmt, fmt-doc, trivia, query and guide goldens through stage 2.

usage: golddiff.py <bin dir> <work dir>
"""
import concurrent.futures, glob, os, re, shutil, subprocess, sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "../../../.."))
BIN, WORK = sys.argv[1], sys.argv[2]
N2 = f"{BIN}/narrac2"
os.makedirs(WORK, exist_ok=True)
os.chdir(ROOT)
TD = "compiler/testdata"
env = dict(os.environ, NARRA_CHECK_LEAKS="1", ASAN_OPTIONS="detect_leaks=0")
DEBUG = ["-O0", "-g", "-std=c11", "-w", "-fsanitize=address,undefined", "-fno-omit-frame-pointer"]
RELEASE = ["-O2", "-DNARRA_RELEASE", "-std=c11", "-w"]
counts = {}
fails = []


def note(suite, ok, detail=""):
    a, b = counts.get(suite, (0, 0))
    counts[suite] = (a + ok, b + 1)
    if not ok:
        fails.append(f"{suite}: {detail}")


def run(args, cwd=None, stdin=""):
    e = dict(env, PWD=os.path.abspath(cwd)) if cwd else env
    p = subprocess.run(args, cwd=cwd, input=stdin, capture_output=True, text=True, env=e, timeout=300)
    return p.returncode, p.stdout, p.stderr


# fmt goldens through --stdin, with the options of the name.
for path in sorted(glob.glob(f"{TD}/fmt/*.nra")):
    base = os.path.basename(path)
    if not re.match(r"(decl|stmt|comment|expr|type|pattern)_", base):
        continue
    scratch = os.path.join(WORK, "fmt_" + base)
    os.makedirs(scratch, exist_ok=True)
    lines = []
    m = re.search(r"_w(\d+)", base)
    if m:
        lines.append("printWidth = " + m.group(1))
    if "_none" in base[:-4]:
        lines.append('trailingComma = "none"')
    toml = os.path.join(scratch, "narra.toml")
    if lines:
        open(toml, "w").write("[fmt]\n" + "\n".join(lines) + "\n")
    code, out, err = run([N2, "fmt", "--stdin"], cwd=scratch, stdin=open(path).read())
    want = open(path[:-4] + ".fmt").read()
    note("fmt goldens", code == 0 and out == want, base)

# fmt config cases.
for dir in sorted(glob.glob(f"{TD}/fmt/config_*")):
    scratch = os.path.join(WORK, os.path.basename(dir))
    os.makedirs(scratch, exist_ok=True)
    shutil.copy(f"{dir}/narra.toml", scratch)
    code, out, err = run([N2, "fmt", "--stdin"], cwd=scratch, stdin=open(f"{dir}/input.nra").read())
    if os.path.exists(f"{dir}/expected.err"):
        note("fmt config", code == 1 and err == open(f"{dir}/expected.err").read(), dir)
    else:
        note("fmt config", code == 0 and out == open(f"{dir}/expected.fmt").read(), dir)

# fmt-doc goldens.
for path in sorted(glob.glob(f"{TD}/fmt/doc_*.txt")):
    code, out, err = run([N2, "fmt-doc", path])
    note("fmt-doc", out == open(path[:-4] + ".out").read(), path)

# trivia goldens.
for path in sorted(glob.glob(f"{TD}/fmt/trivia_*.nra")):
    code, out, err = run([N2, "trivia", path])
    note("trivia", out == open(path[:-4] + ".out").read(), path)

# TestSourcesFormatted on selfhost/src.
code, out, err = run([N2, "fmt", "--check", "selfhost/src"])
note("sources formatted (selfhost/src)", code == 0, out[:200] + err[:200])

# query cases.
for dir in sorted(glob.glob(f"{TD}/query/*")):
    text = open(f"{dir}/queries.txt").read()
    cases = []
    for line in text.splitlines(keepends=True):
        if line.startswith("> "):
            cases.append([line[2:].strip(), ""])
        elif cases:
            cases[-1][1] += line
    for command, want in cases:
        code, out, err = run([N2, "query", "--package", dir] + command.split())
        got = out
        if code != 0:
            got = f"# exit {code}\n" + got
        if err:
            got += "# stderr\n" + err
        note("query", got == want, f"{dir} > {command}")


# guide samples.
def guide_samples(path):
    lines = open(path).read().split("\n")
    fences = []
    i = 0
    while i < len(lines):
        if lines[i].startswith("```"):
            start = i
            info = lines[i].lstrip("`").strip()
            body = []
            i += 1
            while i < len(lines) and not lines[i].startswith("```"):
                body.append(lines[i])
                i += 1
            fences.append((info, start + 1, i + 1, "\n".join(body) + "\n"))
        i += 1
    samples = []
    for k, (info, line, end, body) in enumerate(fences):
        if info not in ("narra", "narra error", "narra fragment"):
            continue
        want = None
        if info == "narra" and k + 1 < len(fences) and fences[k + 1][0] == "text":
            if "".join(lines[end:fences[k + 1][1] - 1]).strip() == "":
                want = fences[k + 1][3]
        samples.append((os.path.basename(path), line, info, body, want))
    return samples


def guide_case(sample):
    file, line, kind, body, want = sample
    name = f"{file}_{line}"
    dir = os.path.join(WORK, "guide", name)
    os.makedirs(dir, exist_ok=True)
    src = os.path.join(dir, "sample.nra")
    open(src, "w").write(body)
    code, c, err = run([N2, "emit", src])
    if kind == "narra error":
        first = body.split("\n", 1)[0]
        w = first.removeprefix("// error:").strip()
        return name, code == 1 and w in err.replace("sample.nra:", ""), err[:200]
    if code != 0:
        return name, False, "emit: " + err[:200]
    open(src[:-4] + ".c", "w").write(c)
    for mode, flags in (("debug", DEBUG), ("release", RELEASE)):
        exe = os.path.join(dir, mode)
        b = run(["cc"] + flags + ["-o", exe, src[:-4] + ".c", "-lm"])
        if b[0] != 0:
            return name, False, "cc failed"
        r = run([exe])
        if r[0] != 0 or r[1] != want:
            return name, False, f"{mode}: exit {r[0]}"
    return name, True, ""


samples = []
for path in sorted(glob.glob("docs/guide/*.md")):
    samples += guide_samples(path)
kinds = {}
for s in samples:
    kinds[s[2]] = kinds.get(s[2], 0) + 1
print("guide sample kinds:", kinds)
with concurrent.futures.ThreadPoolExecutor(max_workers=os.cpu_count()) as pool:
    for name, ok, detail in pool.map(guide_case, [s for s in samples if s[2] != "narra fragment"]):
        note("guide", ok, name + " " + detail)

for f in fails:
    print("FAIL", f)
for suite, (ok, total) in counts.items():
    print(f"{suite}: {ok}/{total} pass")
