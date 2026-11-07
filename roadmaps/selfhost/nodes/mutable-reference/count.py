#!/usr/bin/env python3
"""Counts for D81: mutable parameters by type, whole assignments to a mutable
parameter, and generic mutable parameters.

Run from the repo root: python3 roadmaps/selfhost/nodes/mutable-reference/count.py
Scans selfhost/src, std/src, compiler/testdata (.nra, .fmt excluded) and the
narra blocks of docs/guide/*.md. Prints one line per hit as path:line.
"""

import os
import re
import sys
from collections import defaultdict

ROOTS = ["selfhost/src", "std/src", "compiler/testdata"]
PRIMITIVES = {
    "int", "float", "bool", "char", "string", "bytes", "void",
    "i8", "i16", "i32", "i64", "u8", "u16", "u32", "u64", "f32", "f64", "ByteIndex",
}
# Built-in types whose contents change in place.
BUILTIN_CONTENTS = {"Array", "Map", "StringBuilder", "Generator", "Coroutine", "Box"}
BUILTIN_ENUMS = {"Option", "Result", "Ordering", "Step"}


def strip(text):
    """Blanks comments, strings, template strings and char literals, keeping
    newlines so line numbers stay right."""
    out = []
    i = 0
    n = len(text)
    while i < n:
        c = text[i]
        if text.startswith("//", i):
            while i < n and text[i] != "\n":
                out.append(" ")
                i += 1
        elif text.startswith("/*", i):
            while i < n and not text.startswith("*/", i):
                out.append("\n" if text[i] == "\n" else " ")
                i += 1
            out.append("  ")
            i += 2
        elif c == '"' or c == "'":
            quote = c
            out.append(" ")
            i += 1
            while i < n and text[i] != quote and text[i] != "\n":
                if text[i] == "\\":
                    out.append(" ")
                    i += 1
                out.append(" ")
                i += 1
            out.append(" ")
            i += 1
        elif c == "`":
            out.append(" ")
            i += 1
            depth = 0
            while i < n:
                if text[i] == "\\":
                    out.append("  ")
                    i += 2
                    continue
                if depth == 0 and text[i] == "`":
                    break
                if text.startswith("${", i):
                    depth += 1
                    out.append("  ")
                    i += 2
                    continue
                if depth > 0 and text[i] == "}":
                    depth -= 1
                out.append("\n" if text[i] == "\n" else " ")
                i += 1
            out.append(" ")
            i += 1
        else:
            out.append(c)
            i += 1
    return "".join(out)


def guide_code(text):
    """Keeps only the lines inside narra fences of a markdown file."""
    lines = text.split("\n")
    keep = []
    inside = False
    for line in lines:
        if line.startswith("```"):
            inside = (not inside) and line[3:].strip().lower().startswith("narra")
            keep.append("")
            continue
        keep.append(line if inside else "")
    return "\n".join(keep)


def line_of(text, offset):
    return text.count("\n", 0, offset) + 1


def read_type(text, start):
    """Reads a type from start to the first , ) = or { at depth 0."""
    depth = 0
    i = start
    while i < len(text):
        c = text[i]
        if c in "<([":
            depth += 1
        elif c in ">)]":
            if depth == 0:
                break
            depth -= 1
        elif depth == 0 and c in ",={\n":
            break
        i += 1
    return " ".join(text[start:i].split()), i


def declarations(text):
    """Type names declared in a source: kind by name."""
    kinds = {}
    for m in re.finditer(r"\b(?:export\s+)?(?:opaque\s+)?(struct|enum|trait)\s+([A-Z]\w*)", text):
        kinds[m.group(2)] = m.group(1)
    for m in re.finditer(r"\bdistinct\s+type\s+([A-Z]\w*)\s*=\s*", text):
        base, _ = read_type(text, m.end())
        kinds[m.group(1)] = "distinct " + base
    return kinds


def find_function(text, offset):
    """The function whose parameter list holds offset: name, type params,
    body start and end offsets (or None for a body-less signature)."""
    # The last `function` keyword or arrow-function opening before offset.
    best = None
    for m in re.finditer(r"\bfunction\*?\s*([A-Za-z_]\w*)\s*(<[^(]*>)?\s*\(", text[:offset]):
        best = m
    if best is None:
        return None
    # Walk to the closing paren of the parameter list.
    i = best.end()
    depth = 1
    while i < len(text) and depth > 0:
        if text[i] == "(":
            depth += 1
        elif text[i] == ")":
            depth -= 1
        i += 1
    if i < offset:
        return None
    # Body: the first { at paren depth 0 that is not inside a type.
    j = i
    angle = 0
    while j < len(text):
        c = text[j]
        if c == "<":
            angle += 1
        elif c == ">" and text[j - 1] != "=":
            angle -= 1
        elif c == "{" and angle <= 0:
            break
        elif c == "\n" and text[j + 1 : j + 2] not in (" ", "\t", ")") and text[j - 1] not in ",(":
            # A signature with no body (trait function).
            if not re.match(r"\s*(uses|:)", text[j:]):
                return best.group(1), best.group(2) or "", None, None
        j += 1
    depth = 1
    k = j + 1
    while k < len(text) and depth > 0:
        if text[k] == "{":
            depth += 1
        elif text[k] == "}":
            depth -= 1
        k += 1
    return best.group(1), best.group(2) or "", j, k


def classify(type_text, kinds, type_params):
    head = re.match(r"[A-Za-z_]\w*|#\[|\(", type_text)
    if not head:
        return "other"
    name = head.group(0)
    if name == "#[":
        return "tuple"
    if name == "(":
        return "function"
    if name in type_params:
        return "generic"
    if name in PRIMITIVES:
        return "primitive"
    if name in BUILTIN_CONTENTS:
        return "builtin:" + name
    if name in BUILTIN_ENUMS:
        return "enum:" + name
    if name == "Self":
        return "self"
    kind = kinds.get(name)
    if kind is None:
        return "unknown:" + name
    if kind.startswith("distinct"):
        base = kind.split(" ", 1)[1]
        return "distinct(" + classify(base, kinds, type_params) + ")"
    return kind


def scan(path, display, text, kinds, results):
    for m in re.finditer(r"\bmutable\s+([a-z_]\w*)\s*:\s*", text):
        param = m.group(1)
        type_text, _ = read_type(text, m.end())
        found = find_function(text, m.start())
        name, generics, body_start, body_end = found if found else ("?", "", None, None)
        type_params = set(re.findall(r"(?:effect\s+)?\b([A-Z]\w*)\b(?=\s*[:,>])", generics))
        kind = classify(type_text, kinds, type_params)
        where = f"{display}:{line_of(text, m.start())}"
        results["params"].append((where, name, param, type_text, kind))
        if body_start is None:
            continue
        body = text[body_start:body_end]
        pattern = r"(?<![\w.])" + re.escape(param) + r"\s*(=(?![=>])|\+=|-=|\*=|/=|%=|\+\+|--|<<=|>>=|&=|\|=|\^=)"
        for a in re.finditer(pattern, body):
            line = line_of(text, body_start + a.start())
            results["whole"].append((f"{display}:{line}", name, param, type_text, kind))


def files():
    for root in ROOTS:
        for dirpath, _, names in os.walk(root):
            for name in sorted(names):
                if name.endswith(".nra"):
                    yield os.path.join(dirpath, name), "nra"
    for name in sorted(os.listdir("docs/guide")):
        if name.endswith(".md"):
            yield os.path.join("docs/guide", name), "md"


def main():
    sources = []
    for path, kind in files():
        with open(path, encoding="utf-8") as f:
            raw = f.read()
        text = guide_code(raw) if kind == "md" else raw
        sources.append((path, strip(text)))
    # selfhost/src and std/src are one program: one table of type names.
    shared = {}
    for path, text in sources:
        if path.startswith(("selfhost/src", "std/src")):
            shared.update(declarations(text))
    by_dir = defaultdict(dict)
    for path, text in sources:
        by_dir[os.path.dirname(path)].update(declarations(text))
    results = defaultdict(list)
    for path, text in sources:
        if path.startswith(("selfhost/src", "std/src")):
            kinds = shared
        elif path.startswith("docs/guide"):
            kinds = declarations(text)
        else:
            kinds = dict(by_dir[os.path.dirname(path)])
            kinds.update(declarations(text))
        scan(path, path, text, kinds, results)

    allowed = ("struct", "builtin:", "distinct(struct", "distinct(builtin:")
    group = lambda p: p.split("/")[0] + "/" + p.split("/")[1] if not p.startswith("docs") else "docs/guide"

    print("# mutable parameters by kind and tree")
    table = defaultdict(lambda: defaultdict(int))
    for where, _, _, _, kind in results["params"]:
        k = kind if not kind.startswith("builtin:") else kind
        table[group(where)][k] += 1
    for tree in sorted(table):
        print(tree, dict(sorted(table[tree].items())))

    print("\n# mutable parameters of a type D81 rejects")
    for where, func, param, type_text, kind in results["params"]:
        if not kind.startswith(allowed):
            print(f"{where} {func}({param}: {type_text}) [{kind}]")

    print("\n# assignments to a whole mutable parameter")
    for where, func, param, type_text, kind in results["whole"]:
        print(f"{where} {func}: {param} [{type_text}]")



if __name__ == "__main__":
    os.chdir(os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..", "..", ".."))
    main()
