#!/usr/bin/env python3
"""Counts behind DECISIONS.md D85 (node mutable-markers).

Run from the repo root: python3 roadmaps/selfhost/nodes/mutable-markers/count.py

Text scans, not a parser. Each count says what it matches. A function is the
text from one `function` line to the next.
"""
import os
import re
import sys

TREES = ["selfhost/src", "std/src", "testdata"]

# A write through an element path: `base[i].f = v`, `base[i] = v`, or an
# argument `mutable base[i]...`.
PATH = r"([A-Za-z_][A-Za-z0-9_.]*)\[([^\]=]+)\]((?:\.[A-Za-z_][A-Za-z0-9_]*|\[[^\]]*\])*)"
WRITE = re.compile(r"^\s*" + PATH + r"\s*=[^=]")
MUTABLE_ARG = re.compile(r"mutable\s+" + PATH)
FUNCTION = re.compile(r"^\s*(?:export\s+)?(?:async\s+)?function\*?\s+([A-Za-z_][A-Za-z0-9_]*)")
INDEX_LOOP = re.compile(
    r"for \(const (?:([A-Za-z_][A-Za-z0-9_]*)|#\[([A-Za-z_][A-Za-z0-9_]*), [A-Za-z_][A-Za-z0-9_]*\]) of "
    r"(?:Ranges?\.(?:upTo|between)\((?:0, )?Arrays?\.length\(([A-Za-z_][A-Za-z0-9_.]*)\)\)"
    r"|Arrays?\.entries\(([A-Za-z_][A-Za-z0-9_.]*)\))"
)
VALUE_LOOP = re.compile(r"for \(const ([A-Za-z_][A-Za-z0-9_]*) of (?:Arrays?\.values\()?([A-Za-z_][A-Za-z0-9_.]*)\)?\)")
IS_SOME = re.compile(r"if \(([A-Za-z_][A-Za-z0-9_.\[\]]*) is [A-Za-z_.]*\(([A-Za-z_][A-Za-z0-9_]*)")
DESTRUCTURE = re.compile(r"^\s*(?:const|let) (\{|#\[)")
FOR_DESTRUCTURE = re.compile(r"for \(const (\{|#\[)")


def nra_files(tree):
    for root, _, names in os.walk(tree):
        for name in sorted(names):
            if name.endswith(".nra"):
                yield os.path.join(root, name)


def block_end(lines, start):
    # The line where the block opened on `start` closes, by brace depth.
    depth = 0
    for index in range(start, len(lines)):
        text = re.sub(r'"[^"]*"|`[^`]*`|//.*', "", lines[index])
        depth += text.count("{") - text.count("}")
        if depth <= 0 and index > start:
            return index
        if depth <= 0 and "{" in text and index == start:
            return index
    return len(lines) - 1


def scan(tree):
    totals = {
        "path writes": 0,
        "mutable path args": 0,
        "functions with a repeated element prefix": 0,
        "repeated prefix uses": 0,
        "index loops": 0,
        "index loops writing the source": 0,
        "value loops whose source is written in the body": 0,
        "is-bindings stored back": 0,
        "destructuring declarations": 0,
        "for-of destructuring": 0,
    }
    examples = {key: [] for key in totals}
    for path in nra_files(tree):
        with open(path, encoding="utf-8") as handle:
            lines = handle.read().split("\n")
        # Group element-path writes by function and prefix.
        function = None
        prefixes = {}

        def flush():
            for (name, prefix), sites in prefixes.items():
                if len(sites) >= 3:
                    totals["functions with a repeated element prefix"] += 1
                    totals["repeated prefix uses"] += len(sites)
                    examples["functions with a repeated element prefix"].append(
                        f"{path}:{sites[0]} {name}: {prefix} x{len(sites)}"
                    )

        for number, line in enumerate(lines, 1):
            found = FUNCTION.match(line)
            if found:
                flush()
                prefixes = {}
                function = found.group(1)
            write = WRITE.match(line)
            if write:
                totals["path writes"] += 1
                key = (function, f"{write.group(1)}[{write.group(2)}]")
                prefixes.setdefault(key, []).append(number)
            for argument in MUTABLE_ARG.finditer(line):
                totals["mutable path args"] += 1
                key = (function, f"{argument.group(1)}[{argument.group(2)}]")
                prefixes.setdefault(key, []).append(number)
                examples["mutable path args"].append(f"{path}:{number} {argument.group(0)}")
            if DESTRUCTURE.match(line):
                totals["destructuring declarations"] += 1
            if FOR_DESTRUCTURE.search(line):
                totals["for-of destructuring"] += 1
        flush()
        # Loops and is-bindings, by block.
        for start, line in enumerate(lines):
            loop = INDEX_LOOP.search(line)
            if loop:
                totals["index loops"] += 1
                variable = loop.group(1) or loop.group(2)
                source = loop.group(3) or loop.group(4)
                end = block_end(lines, start)
                body = "\n".join(lines[start + 1 : end])
                target = re.escape(source) + r"\[" + re.escape(variable) + r"\]"
                if re.search(r"^\s*" + target + r"[^=\n]*\s=[^=]", body, re.M) or re.search(
                    r"mutable\s+" + target, body
                ):
                    totals["index loops writing the source"] += 1
                    examples["index loops writing the source"].append(f"{path}:{start + 1} {source}[{variable}]")
            value = VALUE_LOOP.search(line)
            if value and not INDEX_LOOP.search(line):
                source = value.group(2)
                end = block_end(lines, start)
                body = "\n".join(lines[start + 1 : end])
                if re.search(r"^\s*" + re.escape(source) + r"\[[^\]]+\][^=\n]*\s=[^=]", body, re.M):
                    totals["value loops whose source is written in the body"] += 1
                    examples["value loops whose source is written in the body"].append(f"{path}:{start + 1} {source}")
            binding = IS_SOME.search(line)
            if binding:
                subject = binding.group(1)
                end = block_end(lines, start)
                body = "\n".join(lines[start + 1 : end])
                if re.search(r"^\s*" + re.escape(subject) + r"\s*=[^=]", body, re.M):
                    totals["is-bindings stored back"] += 1
                    examples["is-bindings stored back"].append(f"{path}:{start + 1} {subject}")
    return totals, examples


def main():
    verbose = "-v" in sys.argv
    for tree in TREES:
        totals, examples = scan(tree)
        print(f"== {tree}")
        for key, value in totals.items():
            print(f"  {key}: {value}")
            if verbose:
                for example in examples[key]:
                    print(f"    {example}")


main()
