#!/usr/bin/env python3
# Counts the code in selfhost/src and std/src that each D103 form would shorten.
# Text heuristics, not a parse. Run from the repo root; -v prints each site.
import os
import re
import sys

ROOTS = ["selfhost/src", "std/src"]
SKIP = {"stdtext.nra"}  # generated copy of std/src
VERBOSE = "-v" in sys.argv

FUNC_HEAD = re.compile(r"^\s*(export\s+)?(function\*?|async function)\s+\w+")
DESTRUCTURE = re.compile(r"^\s*const\s+(\{[^=]*\}|#\[[^=]*\])\s*=\s*(\w+)\s*$")
TUPLE_DESTRUCTURE = re.compile(r"^\s*const\s+#\[([^\]=]*)\]\s*=\s*(.+)$")
STRUCT_DESTRUCTURE = re.compile(r"^\s*const\s+\{([^}=]*)\}\s*=\s*(.+)$")
PLAIN_CONST = re.compile(r"^\s*const\s+(\w+)(\s*:\s*[^=]+)?\s*=\s*(.+)$")
ASSIGN = re.compile(r"^\s*([\w.\[\]]+)\s*=\s*(.+)$")
LENGTH_TEST = re.compile(r"Arrays\.length\((\w+(?:\.\w+)*)\)\s*(==|!=|>=|>|<|<=)\s*([0-9])\b")
ARROW_PARAM = re.compile(r"\((\w+)(?::[^()]*)?\)\s*(?::\s*[^=]+)?=>")


def files():
    for root in ROOTS:
        for directory, _, names in os.walk(root):
            for name in sorted(names):
                if name.endswith(".nra") and name not in SKIP:
                    yield os.path.join(directory, name)


def split_functions(lines):
    # Yields (start, end, header) for each top-level function, header joined over
    # the lines up to the opening brace of the body.
    starts = [i for i, line in enumerate(lines) if FUNC_HEAD.match(line)]
    starts.append(len(lines))
    for k in range(len(starts) - 1):
        start, end = starts[k], starts[k + 1]
        header = ""
        for i in range(start, end):
            header += lines[i]
            if lines[i].rstrip().endswith("{"):
                break
        yield start, end, header


def parameter_names(header):
    inside = header[header.find("(") + 1 :]
    # A `mutable` parameter could not destructure, so it is left out.
    names = re.findall(r"(?:^\s*|[,(]\s*)(?:escaping\s+)?(\w+)\s*:", inside)
    return set(names)


def report(kind, path, line_number, text):
    if VERBOSE:
        print(f"  {kind} {path}:{line_number + 1}: {text.strip()}")


def count_words(text, name):
    return len(re.findall(r"(?<![\w.])" + re.escape(name) + r"\b", text))


def main():
    totals = {
        "param.declaration": 0,
        "param.declaration.only_use": 0,
        "param.arrow": 0,
        "param.arrow.only_use": 0,
        "param.fields3": 0,
        "array.length_index": 0,
        "array.length_tests": 0,
        "array.rest_loop": 0,
        "assign.swap": 0,
        "assign.from_tuple": 0,
        "assign.from_struct": 0,
        "for.entry_destructure": 0,
    }
    for path in files():
        with open(path) as handle:
            lines = handle.read().split("\n")

        # 1. Parameters destructured at the top of a body.
        for start, end, header in split_functions(lines):
            parameters = parameter_names(header)
            body = "\n".join(lines[start:end])
            for i in range(start, end):
                match = DESTRUCTURE.match(lines[i])
                if match and match.group(2) in parameters:
                    totals["param.declaration"] += 1
                    report("param", path, i, lines[i])
                    if count_words(body, match.group(2)) == 2:
                        totals["param.declaration.only_use"] += 1
            # A struct parameter read through three or more distinct fields
            # and never used whole: a candidate for `{ a, b, c }: T`.
            for name in parameters:
                uses = count_words(body, name) - 1
                fields = set(re.findall(r"(?<![\w.])" + re.escape(name) + r"\.(\w+)", body))
                field_uses = len(re.findall(r"(?<![\w.])" + re.escape(name) + r"\.\w+", body))
                if len(fields) >= 3 and field_uses == uses and not re.search(
                    r"(?<![\w.])" + re.escape(name) + r"\.\w+(\.\w+|\[)*\s*(=[^=]|\+=|-=|\+\+)", body
                ):
                    totals["param.fields3"] += 1
                    report("fields3", path, start, f"{name}: {sorted(fields)}")

        # Arrow functions whose body starts by destructuring the parameter.
        for i, line in enumerate(lines):
            for arrow in ARROW_PARAM.finditer(line):
                name = arrow.group(1)
                for j in range(i + 1, min(i + 3, len(lines))):
                    match = DESTRUCTURE.match(lines[j])
                    if match and match.group(2) == name:
                        totals["param.arrow"] += 1
                        report("arrow", path, j, lines[j])

        # for (const entry of ...) followed by const #[..] = entry.
        for i, line in enumerate(lines):
            loop = re.search(r"for \(const (\w+) of ", line)
            if loop and i + 1 < len(lines):
                match = DESTRUCTURE.match(lines[i + 1])
                if match and match.group(2) == loop.group(1):
                    totals["for.entry_destructure"] += 1
                    report("for", path, i + 1, lines[i + 1])

        # 2. Array patterns: a length test against a small constant whose
        # branch indexes the same array with a literal index.
        for i, line in enumerate(lines):
            for test in LENGTH_TEST.finditer(line):
                array = test.group(1)
                totals["array.length_tests"] += 1
                window = "\n".join(lines[i : i + 6])
                if re.search(re.escape(array) + r"\[[0-9]\]", window):
                    totals["array.length_index"] += 1
                    totals["array.length_index." + test.group(2)] = totals.get("array.length_index." + test.group(2), 0) + 1
                    report("array", path, i, line)
            # A loop from index 1 over the array: the `...rest` of [first, ...rest].
            # Both the C-style loop and `Ranges.between(1, Arrays.length(x))`.
            loop = re.search(r"for \(let (\w+) = 1; \1 < Arrays\.length\((\w+(?:\.\w+)*)\)", line)
            ranged = re.search(r"Ranges\.between\(1, *Arrays\.length\(", line)
            if loop or ranged:
                totals["array.rest_loop"] += 1
                report("rest", path, i, line)

        # 3. Destructuring assignment.
        for i in range(len(lines) - 2):
            first = PLAIN_CONST.match(lines[i])
            second = ASSIGN.match(lines[i + 1])
            third = ASSIGN.match(lines[i + 2])
            # Swap through a temporary: const t = a; a = b; b = t
            if first and second and third:
                temp, source = first.group(1), first.group(3).strip()
                if (
                    second.group(1) == source
                    and third.group(1) == second.group(2).strip()
                    and third.group(2).strip() == temp
                ):
                    totals["assign.swap"] += 1
                    report("swap", path, i, lines[i])
        for i, line in enumerate(lines):
            # const #[a, b] = f() followed by x = a for a bound name.
            match = TUPLE_DESTRUCTURE.match(line) or STRUCT_DESTRUCTURE.match(line)
            if match:
                bound = [part.split(":")[-1].strip() for part in match.group(1).split(",")]
                for j in range(i + 1, min(i + 4, len(lines))):
                    assign = ASSIGN.match(lines[j])
                    if assign and assign.group(2).strip() in bound and re.fullmatch(r"\w+", assign.group(1)):
                        totals["assign.from_tuple"] += 1
                        report("assign", path, j, lines[j])
                        break
            # const r = f() followed by two or more `x = r.field` lines.
            plain = PLAIN_CONST.match(line)
            if plain:
                name = plain.group(1)
                run = 0
                j = i + 1
                while j < len(lines):
                    assign = ASSIGN.match(lines[j])
                    if assign and re.fullmatch(re.escape(name) + r"\.\w+", assign.group(2).strip()):
                        run += 1
                        j += 1
                    else:
                        break
                if run >= 2:
                    totals["assign.from_struct"] += 1
                    report("assign-struct", path, i, line)
    for key, value in totals.items():
        print(f"{key}: {value}")


main()
