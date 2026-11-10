#!/usr/bin/env python3
"""Splits the memmove/memcpy samples of a `sample` report by the C statement that made the copy.

Usage: python3 -I copies.py <sample.txt> <the C file the binary was built from>
A copy whose calling statement passes a struct by value (a call in compiler code or a
retain_S_/release_S_ helper, with no string or builder call on the line) counts as a
by-value struct copy. Copies made inside lstr_*, narra_sb_* and narra_str_* count as strings.
"""
import collections
import re
import sys

LINE = re.compile(r"^([ +!:|]*)(\d+) (.*?)  \(in ([^)]*)\)(.*)$")
COPY_LEAVES = ("_platform_memmove", "DYLD-STUB$$memcpy", "DYLD-STUB$$memmove")


def main():
    report, c_file = sys.argv[1], sys.argv[2]
    with open(c_file) as handle:
        source = handle.read().split("\n")
    total = 0
    kinds = collections.Counter()
    by_value = collections.Counter()
    stack = []
    in_graph = False
    with open(report, errors="replace") as handle:
        for raw in handle:
            if raw.startswith("Call graph:"):
                in_graph = True
                continue
            if raw.startswith("Total number in stack"):
                break
            if not in_graph:
                continue
            match = LINE.match(raw.rstrip("\n"))
            if not match:
                thread = re.match(r"^(\s*)(\d+) Thread_", raw)
                if thread and int(thread.group(2)) > total:
                    total = int(thread.group(2))
                continue
            depth = len(match.group(1))
            while stack and stack[-1][0] >= depth:
                stack.pop()
            stack.append((depth, match.group(3).strip(), match.group(4), match.group(5), int(match.group(2))))
            if match.group(3).strip() not in COPY_LEAVES:
                continue
            samples = stack[-1][4]
            caller = None
            for frame in reversed(stack[:-1]):
                if frame[2].startswith("narrac") and not frame[1].startswith("DYLD"):
                    caller = frame
                    break
            if caller is None:
                kinds["unknown"] += samples
                continue
            name = caller[1]
            found = re.search(r"\.c:(\d+)", caller[3])
            text = source[int(found.group(1)) - 1].strip() if found and int(found.group(1)) > 0 else ""
            if name.startswith(("lstr_", "narra_sb_", "narra_str_")) or re.search(r"lstr_|narra_sb_|narra_str_", text):
                kinds["strings"] += samples
            elif name.startswith(("A_Array_", "M_Map_")):
                kinds["arrays and maps"] += samples
            elif name.startswith(("f_", "g_", "retain_S_", "release_S_", "retain_E_", "release_E_")) and "(" in text:
                kinds["by-value struct argument"] += samples
                by_value[name] += samples
            else:
                kinds["other (no line, assignment)"] += samples
    print("main thread samples: %d" % total)
    for kind, count in kinds.most_common():
        print("%6d %5.1f%%  %s" % (count, 100.0 * count / total, kind))
    print("\nby-value struct copies, by the function making the call:")
    for name, count in by_value.most_common(20):
        print("%6d %5.1f%%  %s" % (count, 100.0 * count / total, name))


main()
