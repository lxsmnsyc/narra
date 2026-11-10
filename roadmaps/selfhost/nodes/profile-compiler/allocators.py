#!/usr/bin/env python3
"""Splits the allocation samples (malloc, free, realloc and narra_alloc/free) of a `sample`
report by the runtime helper that asked for them. Run it on the -fno-inline build, where
lstr_*, A_Array_* and the rest keep their own frames.

Usage: python3 -I allocators.py <sample.txt>
"""
import collections
import re
import sys

LINE = re.compile(r"^([ +!:|]*)(\d+) (.*?)  \(in ([^)]*)\)")
HELPERS = (
    ("lstr_concat", "string concatenation (lstr_concat)"),
    ("lstr_from_int", "int to string"),
    ("lstr_", "other string helpers (lstr_*)"),
    ("narra_str_", "std string functions (narra_str_*)"),
    ("narra_sb_", "string builder (narra_sb_*)"),
    ("A_Array_", "arrays"),
    ("M_Map_", "maps"),
    ("release_", "frees in release_* and drop_*"),
    ("drop_", "frees in release_* and drop_*"),
    ("narra_drop", "frees in release_* and drop_*"),
)


def classify(stack):
    # The first narrac frame above the allocator decides; a Narra function means a box or enum payload.
    for frame in reversed(stack):
        name, binary = frame
        if not binary.startswith("narrac") or name.startswith("DYLD-STUB") or name in (
                "narra_alloc", "narra_free", "narra_realloc"):
            continue
        for prefix, label in HELPERS:
            if name.startswith(prefix):
                return label
        if name.endswith("_release"):
            return "frees in release_* and drop_*"
        if name.startswith(("f_", "g_")):
            return "boxes, enum payloads and closures allocated in compiler code"
        return "other runtime: " + name
    return "unknown"


def main():
    counts = collections.Counter()
    total = 0
    stack = []
    in_graph = False
    with open(sys.argv[1], errors="replace") as handle:
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
                if thread:
                    total = max(total, int(thread.group(2)))
                continue
            depth = len(match.group(1))
            while stack and stack[-1][0] >= depth:
                stack.pop()
            name, binary = match.group(3).strip(), match.group(4)
            stack.append((depth, name, binary, int(match.group(2))))
            is_allocator = binary == "libsystem_malloc.dylib" or name in (
                "narra_alloc", "narra_free", "narra_realloc", "DYLD-STUB$$malloc", "DYLD-STUB$$free",
                "DYLD-STUB$$realloc")
            if not is_allocator:
                continue
            # Count each allocator subtree once, at its outermost allocator frame.
            parent = stack[-2] if len(stack) > 1 else None
            if parent and (parent[2] == "libsystem_malloc.dylib" or parent[1] in (
                    "narra_alloc", "narra_free", "narra_realloc", "DYLD-STUB$$malloc", "DYLD-STUB$$free",
                    "DYLD-STUB$$realloc")):
                continue
            counts[classify([(item[1], item[2]) for item in stack[:-1]])] += stack[-1][3]
    allocation = sum(counts.values())
    print("main thread samples: %d, allocation subtrees: %d (%.1f%%)" % (total, allocation, 100.0 * allocation / total))
    for label, count in counts.most_common():
        print("%6d %5.1f%%  %s" % (count, 100.0 * count / total, label))


main()
