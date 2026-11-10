#!/usr/bin/env python3
"""Summarizes a macOS `sample` report of narrac: top functions by self and inclusive
samples, shares per cost category and per compiler phase, and the callers of memmove.

Usage: python3 -I analyze.py <sample.txt> <repo root> <the C file the binary was built from>
       python3 -I analyze.py --callers <sample.txt> <repo root> <function label>
Hashed C names (f_<module hash>_<name>) are mapped back to Narra modules with the
module hash from selfhost/src/hash.nra: xxh32 of the module path after its first "/".
"""
import collections
import os
import re
import sys

MASK = 0xFFFFFFFF
P1, P2, P3, P4, P5 = 2654435761, 2246822519, 3266489917, 668265263, 374761393


def rotate(value, bits):
    return ((value << bits) | (value >> (32 - bits))) & MASK


def xxh32(data, seed=0):
    length = len(data)
    offset = 0
    if length >= 16:
        lanes = [(seed + P1 + P2) & MASK, (seed + P2) & MASK, seed & MASK, (seed - P1) & MASK]
        while offset + 16 <= length:
            for lane in range(4):
                word = int.from_bytes(data[offset + 4 * lane:offset + 4 * lane + 4], "little")
                lanes[lane] = (rotate((lanes[lane] + word * P2) & MASK, 13) * P1) & MASK
            offset += 16
        value = (rotate(lanes[0], 1) + rotate(lanes[1], 7) + rotate(lanes[2], 12) + rotate(lanes[3], 18)) & MASK
    else:
        value = (seed + P5) & MASK
    value = (value + length) & MASK
    while offset + 4 <= length:
        word = int.from_bytes(data[offset:offset + 4], "little")
        value = (rotate((value + word * P3) & MASK, 17) * P4) & MASK
        offset += 4
    while offset < length:
        value = (rotate((value + data[offset] * P5) & MASK, 11) * P1) & MASK
        offset += 1
    value ^= value >> 15
    value = (value * P2) & MASK
    value ^= value >> 13
    value = (value * P3) & MASK
    value ^= value >> 16
    return value


def load_module_hashes(root):
    hashes = {}
    for folder, prefix in (("selfhost/src", "narrac"), ("std/src", "std")):
        base = os.path.join(root, folder)
        for directory, _, files in os.walk(base):
            for name in files:
                if not name.endswith(".nra"):
                    continue
                relative = os.path.relpath(os.path.join(directory, name), base)[:-4]
                hashes["%08x" % xxh32(relative.encode())] = "%s/%s" % (prefix, relative)
    return hashes


LINE = re.compile(r"^([ +!:|]*)(\d+) (.*?)  \(in ([^)]*)\)")


class Node:
    __slots__ = ("name", "binary", "count", "children", "parent", "line")

    def __init__(self, name, binary, count, parent):
        self.name, self.binary, self.count, self.parent = name, binary, count, parent
        self.children = []
        self.line = ""


def parse(path):
    root = Node("<root>", "", 0, None)
    stack = [(-1, root)]
    in_graph = False
    with open(path, errors="replace") as handle:
        for line in handle:
            if line.startswith("Call graph:"):
                in_graph = True
                continue
            if not in_graph:
                continue
            if line.startswith("Total number in stack") or line.startswith("Sort by top"):
                break
            match = LINE.match(line.rstrip("\n"))
            if not match:
                # Thread lines have no "(in ...)"; they are roots.
                thread = re.match(r"^(\s*)(\d+) (Thread_\S+)", line)
                if thread:
                    depth = len(thread.group(1))
                    while stack[-1][0] >= depth:
                        stack.pop()
                    node = Node(thread.group(3), "", int(thread.group(2)), stack[-1][1])
                    stack[-1][1].children.append(node)
                    stack.append((depth, node))
                continue
            depth = len(match.group(1))
            while stack[-1][0] >= depth:
                stack.pop()
            parent = stack[-1][1]
            node = Node(match.group(3).strip(), match.group(4), int(match.group(2)), parent)
            node.line = line.rstrip("\n").split("]")[-1].strip()
            parent.children.append(node)
            stack.append((depth, node))
    return root


def walk(node):
    yield node
    for child in node.children:
        yield from walk(child)


def self_count(node):
    return node.count - sum(child.count for child in node.children)


def category(name, binary):
    # The category of a frame, or None when the frame is plain code that inherits its caller's.
    if binary == "libsystem_malloc.dylib" or name in ("DYLD-STUB$$malloc", "DYLD-STUB$$free",
                                                       "DYLD-STUB$$realloc", "DYLD-STUB$$calloc"):
        return "allocation"
    if name in ("narra_alloc", "narra_free", "narra_realloc"):
        return "allocation"
    if name.startswith(("retain_", "release_", "drop_", "narra_drop")) or name.endswith(("_retain", "_release")):
        return "reference counting"
    if name.startswith(("lstr_", "narra_sb_", "narra_str_")):
        return "strings"
    if name.startswith("M_Map_"):
        return "maps"
    if name.startswith("A_Array_"):
        return "arrays"
    return None


LEAF_COPY = ("_platform_memmove", "DYLD-STUB$$memcpy", "DYLD-STUB$$memmove", "_platform_memset", "DYLD-STUB$$memset")


STRING_LEAVES = ("__vfprintf", "__v2printf", "_vsnprintf", "snprintf", "vsnprintf_l", "__sfvwrite", "__ultoa",
                 "_platform_memcmp", "DYLD-STUB$$memcmp", "_platform_strlen", "DYLD-STUB$$strlen",
                 "DYLD-STUB$$snprintf", "localeconv_l", "__vfprintf.cold.1")


def frame_category(node, source):
    # Walks up from a leaf to the first frame with a category. In compiler code a copy is
    # "by-value struct copy" when its C statement is a call with no string helper on it,
    # printf and memcmp are int-to-string and string comparison, and the rest is "compiler code".
    own = category(node.name, node.binary)
    if own:
        return own
    current = node
    while current is not None and current.name != "<root>":
        found = category(current.name, current.binary)
        if found:
            return found
        if current.binary.startswith("narrac") and not current.name.startswith("DYLD-STUB"):
            if node.name in LEAF_COPY:
                match = re.search(r"\.c:(\d+)$", current.line)
                text = source[int(match.group(1)) - 1] if match and int(match.group(1)) > 0 else ""
                if "(" in text and not re.search(r"lstr_|narra_sb_|narra_str_", text):
                    return "by-value struct copy"
                return "strings"
            if node.name in STRING_LEAVES:
                return "strings"
            if current.name.startswith(("f_", "g_")) or current.name == "main":
                return "compiler code"
        current = current.parent
    return "compiler code"


def function_label(name, hashes):
    match = re.match(r"^([fg])_([0-9a-f]{8})_(.*)$", name)
    if match and match.group(2) in hashes:
        return "%s.%s" % (hashes[match.group(2)].split("/", 1)[1], match.group(3))
    return name


def phase_at_line(text):
    # The phase of a call site in runEmit, from the C statement on that line.
    if "_load(" in text:
        return "load"
    if "checkPackage" in text or "findEntry" in text:
        return "check"
    if "checkEffects" in text:
        return "effects"
    if "emitProgram" in text:
        return "emit"
    if "narra_fs_write" in text:
        return "write"
    if "release" in text:
        return "teardown (releases in runEmit)"
    return "runEmit other: " + text.strip()[:60]


def phase(node, source):
    # The phase is the runEmit call site on the path; inside load, lexer frames are lex, the rest parse.
    path = []
    current = node
    while current is not None:
        path.append(current)
        current = current.parent
    path.reverse()
    for index, frame in enumerate(path):
        if frame.name.endswith("_runEmit"):
            match = re.search(r"\.c:(\d+)$", frame.line)
            where = phase_at_line(source[int(match.group(1)) - 1]) if match else "runEmit other"
            if where == "load":
                names = [item.name for item in path[index + 1:]]
                if any(name.startswith("f_bc628247_") for name in names):
                    return "lex"
                if any(name.startswith("f_041ba2c9_") for name in names):
                    return "parse"
                return "load other (files, module graph)"
            return where
    return "outside runEmit"


def print_callers(tree, hashes, target):
    # Inclusive samples of the frames named target, split by the caller frame (outermost occurrence on a path).
    callers = collections.Counter()
    callees = collections.Counter()
    for node in walk(tree):
        samples = self_count(node)
        if samples <= 0:
            continue
        path = []
        current = node
        while current is not None:
            path.append(current)
            current = current.parent
        path.reverse()
        for index, frame in enumerate(path):
            if function_label(frame.name, hashes) == target:
                callers[function_label(path[index - 1].name, hashes)] += samples
                below = function_label(path[index + 1].name, hashes) if index + 1 < len(path) else "(self)"
                callees[below] += samples
                break
    print("## callers of %s" % target)
    for name, count in callers.most_common(15):
        print("%6d  %s" % (count, name))
    print("## callees of %s" % target)
    for name, count in callees.most_common(25):
        print("%6d  %s" % (count, name))


def main():
    if sys.argv[1] == "--callers":
        hashes = load_module_hashes(sys.argv[3])
        print_callers(parse(sys.argv[2]), hashes, sys.argv[4])
        return
    report, root_dir, c_file = sys.argv[1], sys.argv[2], sys.argv[3]
    with open(c_file) as handle:
        source = handle.read().split("\n")
    hashes = load_module_hashes(root_dir)
    tree = parse(report)
    threads = [child for child in tree.children]
    main_thread = max(threads, key=lambda node: node.count)
    total = main_thread.count
    self_by = collections.Counter()
    inclusive_by = collections.Counter()
    categories = collections.Counter()
    phases = collections.Counter()
    phase_categories = collections.defaultdict(collections.Counter)
    memmove_callers = collections.Counter()
    category_callers = collections.defaultdict(collections.Counter)
    for node in walk(main_thread):
        samples = self_count(node)
        if samples <= 0:
            continue
        label = function_label(node.name, hashes)
        self_by[label] += samples
        seen = set()
        current = node
        while current is not None and current is not main_thread:
            seen.add(function_label(current.name, hashes))
            current = current.parent
        for name in seen:
            inclusive_by[name] += samples
        kind = frame_category(node, source)
        categories[kind] += samples
        where = phase(node, source)
        phases[where] += samples
        phase_categories[where][kind] += samples
        if kind != "compiler code":
            current = node
            while current is not None and not re.match(r"^[fg]_[0-9a-f]{8}_", current.name):
                current = current.parent
            category_callers[kind][function_label(current.name, hashes) if current else "?"] += samples
        if node.name in LEAF_COPY:
            current = node.parent
            while current is not None and not current.binary.startswith("narrac"):
                current = current.parent
            if current is not None and current.name.startswith("DYLD-STUB"):
                current = current.parent
            memmove_callers[function_label(current.name, hashes) if current else "?"] += samples

    def share(count):
        return "%.1f%%" % (100.0 * count / total)

    print("samples on the main thread: %d" % total)
    print("\n## top 25 by self samples")
    for name, count in self_by.most_common(25):
        print("%6d %6s  %s" % (count, share(count), name))
    print("\n## top 40 by inclusive samples")
    for name, count in inclusive_by.most_common(40):
        print("%6d %6s  %s" % (count, share(count), name))
    print("\n## inclusive, every function at 0.5% or more")
    for name, count in sorted(inclusive_by.items(), key=lambda item: -item[1]):
        if count * 200 >= total:
            print("%6d %6s  %s" % (count, share(count), name))
    print("\n## categories (self samples)")
    for name, count in categories.most_common():
        print("%6d %6s  %s" % (count, share(count), name))
    print("\n## phases (self samples)")
    for name, count in phases.most_common():
        print("%6d %6s  %s   [%s]" % (count, share(count), name,
                                      ", ".join("%s %s" % (k, share(v)) for k, v in phase_categories[name].most_common())))
    print("\n## memmove/memcpy/memset callers")
    for name, count in memmove_callers.most_common(20):
        print("%6d %6s  %s" % (count, share(count), name))
    for kind, callers in sorted(category_callers.items()):
        print("\n## nearest Narra function above %s samples" % kind)
        for name, count in callers.most_common(15):
            print("%6d %6s  %s" % (count, share(count), name))


main()
