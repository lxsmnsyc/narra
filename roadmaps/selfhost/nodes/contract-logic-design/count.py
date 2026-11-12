#!/usr/bin/env python3
# Counts the checks in selfhost/src and std/src that each D102 form would serve.
# Text heuristics, not a parse. Run from the repo root; -v prints each site.
#
# A check is the condition of an `if`, `else if` or `while`. It counts when it
# reads an integer parameter of the enclosing function (not `mutable`). An entry
# guard is a check among the first three statements of the body whose block
# returns or panics within three lines: the shape a `requires` line would replace,
# since the condition the caller must prove is the guard negated.
import os
import re
import sys

ROOTS = ["selfhost/src", "std/src"]
VERBOSE = "-v" in sys.argv

INT_TYPES = r"(?:int|i8|i16|i32|i64|u8|u16|u32|u64)"
FUNC_HEAD = re.compile(r"^\s*(export\s+)?(function\*?|async function)\s+\w+")
CONSTANT = r"(?:-?[0-9][0-9_xa-fA-F]*|[A-Z][A-Z0-9_]+|\w+\.[A-Z][A-Z0-9_]+)"
NAME = r"[a-z]\w*"
OPERATOR = r"(?:<=|>=|==|!=|<|>)"
COMPARISON = re.compile(
    rf"(?<![\w.])({NAME}|{CONSTANT})\s*({OPERATOR})\s*({NAME}|{CONSTANT})(?![\w.(\[])"
)
NEGATION = re.compile(rf"!\(\s*[a-z][\w.]*\s*{OPERATOR}\s")
CHECK_HEAD = re.compile(r"\b(if|while)\s*\(")
RESULT_INT = re.compile(rf"\)\s*:\s*(Option|Result)<\s*{INT_TYPES}\b")


def files():
    for root in ROOTS:
        for directory, _, names in os.walk(root):
            for name in sorted(names):
                # Generated copies of std/src and of the C runtime are not code.
                if not name.endswith(".nra") or name == "stdtext.nra" or name.startswith("runtime"):
                    continue
                yield os.path.join(directory, name)


def split_functions(lines):
    starts = [i for i, line in enumerate(lines) if FUNC_HEAD.match(line)]
    starts.append(len(lines))
    for k in range(len(starts) - 1):
        start, end = starts[k], starts[k + 1]
        header = ""
        body_start = end
        for i in range(start, end):
            header += lines[i]
            if lines[i].rstrip().endswith("{"):
                body_start = i + 1
                break
        yield start, body_start, end, header


def integer_parameters(header):
    inside = header[header.find("(") + 1 : header.rfind(")")]
    names = set()
    for match in re.finditer(rf"(?:^|[,(])\s*(mutable\s+)?(\w+)\s*:\s*{INT_TYPES}\b(?!<)", inside):
        if not match.group(1):
            names.add(match.group(2))
    return names


def read_condition(lines, index, column):
    # The text between the parenthesis after `if` or `while` and its match.
    depth = 0
    text = ""
    line = index
    position = column
    while line < len(lines):
        source = lines[line]
        while position < len(source):
            character = source[position]
            if character == "(":
                depth += 1
                if depth == 1:
                    position += 1
                    continue
            elif character == ")":
                depth -= 1
                if depth == 0:
                    return text
            if depth >= 1:
                text += character
            position += 1
        text += " "
        line += 1
        position = 0
    return text


def split_top(text, operator):
    # Splits on `&&` or `||` outside parentheses.
    parts = []
    depth = 0
    current = ""
    i = 0
    while i < len(text):
        if text[i] in "([":
            depth += 1
        elif text[i] in ")]":
            depth -= 1
        if depth == 0 and text.startswith(operator, i):
            parts.append(current.strip())
            current = ""
            i += 2
            continue
        current += text[i]
        i += 1
    parts.append(current.strip())
    return parts


def strip_parens(text):
    text = text.strip()
    while text.startswith("(") and text.endswith(")"):
        depth = 0
        whole = True
        for i, character in enumerate(text):
            depth += character == "("
            depth -= character == ")"
            if depth == 0 and i < len(text) - 1:
                whole = False
                break
        if not whole:
            break
        text = text[1:-1].strip()
    return text


def classify_comparison(text, parameters):
    # "constant": one integer parameter against a constant; "relation": two integer
    # parameters; None: anything else (a call, a field, a local).
    text = strip_parens(text)
    match = COMPARISON.fullmatch(text)
    if not match:
        return None
    left, right = match.group(1), match.group(3)
    left_parameter = left in parameters
    right_parameter = right in parameters
    left_constant = re.fullmatch(CONSTANT, left) is not None
    right_constant = re.fullmatch(CONSTANT, right) is not None
    if left_parameter and right_parameter:
        return "relation"
    if (left_parameter and right_constant) or (right_parameter and left_constant):
        return "constant"
    return None


def classify_condition(text, parameters):
    # Returns the set of D102 forms the condition uses.
    forms = set()
    text = strip_parens(text)
    ors = split_top(text, "||")
    if len(ors) > 1:
        kinds = [classify_atom(part, parameters) for part in ors]
        if all(kinds):
            forms.add("or")
            for kind in kinds:
                forms |= kind
        return forms
    ands = split_top(text, "&&")
    if len(ands) > 1:
        for part in ands:
            forms |= classify_condition(part, parameters)
        if forms:
            forms.add("and")
        return forms
    return classify_atom(text, parameters) or set()


def classify_atom(text, parameters):
    text = strip_parens(text)
    if text.startswith("!"):
        inner = classify_atom(text[1:], parameters)
        if inner and inner - {"not"}:
            return inner | {"not"}
        return None
    if len(split_top(text, "||")) > 1 or len(split_top(text, "&&")) > 1:
        nested = classify_condition(text, parameters)
        return nested or None
    if "||" in text or "&&" in text:
        return None
    kind = classify_comparison(text, parameters)
    if kind:
        return {kind}
    return None


def is_guard(lines, index, body_start):
    # Among the first three statements, and the block returns or panics soon.
    statements = 0
    for i in range(body_start, index):
        stripped = lines[i].strip()
        if stripped and not stripped.startswith("//") and not stripped.startswith("}"):
            statements += 1
    if statements > 3:
        return False
    for i in range(index, min(index + 4, len(lines))):
        if re.search(r"\b(return|panic)\b", lines[i]):
            return True
    return False


def main():
    counts = {}
    sites = {}

    def add(kind, path, line, text):
        counts[kind] = counts.get(kind, 0) + 1
        sites.setdefault(kind, []).append(f"{path}:{line + 1}: {text.strip()}")

    for path in files():
        with open(path) as handle:
            lines = handle.read().split("\n")
        # `!(` before a comparison on any line, and the checks that hold one.
        for index, line in enumerate(lines):
            if NEGATION.search(line):
                add("negation: !( before a comparison, any line", path, index, line)
            for match in CHECK_HEAD.finditer(line):
                if NEGATION.search(read_condition(lines, index, match.end() - 1)):
                    add("negation: !( before a comparison, in a check", path, index, line)
        for start, body_start, end, header in split_functions(lines):
            parameters = integer_parameters(header)
            if RESULT_INT.search(header):
                add("result Option/Result of an integer", path, start, header.split("{")[0][:100])
            if not parameters:
                continue
            for index in range(body_start, end):
                for match in CHECK_HEAD.finditer(lines[index]):
                    condition = read_condition(lines, index, match.end() - 1)
                    forms = classify_condition(condition, parameters)
                    if not forms:
                        continue
                    guard = is_guard(lines, index, body_start)
                    add("check: parameter checks", path, index, condition)
                    if guard:
                        add("guard: entry guards", path, index, condition)
                    for form in sorted(forms):
                        add(f"check: {form}", path, index, condition)
                        if guard:
                            add(f"guard: {form}", path, index, condition)
                    # The condition a caller would prove is the guard negated:
                    # a guard joined by `&&` needs `||` in requires.
                    if guard and "and" in forms and "or" not in forms:
                        add("guard needing || in requires (guard joined by &&)", path, index, condition)
                    if guard and "or" in forms and "relation" not in forms:
                        add("guard expressible today (guard joined by ||, constants)", path, index, condition)
    for kind in sorted(counts):
        print(f"{counts[kind]:5d}  {kind}")
        if VERBOSE:
            for site in sites[kind]:
                print(f"         {site}")


main()
