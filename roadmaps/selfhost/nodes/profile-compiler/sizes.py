#!/usr/bin/env python3
"""Prints the C size of every struct or enum type the emitted compiler passes by value,
with the number of function parameters of that type.

Usage: python3 -I sizes.py <emitted C file> <work dir>
"""
import collections
import os
import re
import subprocess
import sys


def main():
    c_file, work = sys.argv[1], sys.argv[2]
    with open(c_file) as handle:
        source = handle.read()
    definitions = re.findall(r"^static [^\n;]*?\b([fg]_[0-9a-f]{8}_\w+)\(([^)]*)\) \{", source, re.M)
    counts = collections.Counter()
    for _, parameters in definitions:
        for parameter in parameters.split(","):
            match = re.match(r"^(S_\w+|E_\w+) v_", parameter.strip())
            if match:
                counts[match.group(1)] += 1
    program = os.path.join(work, "sizes.c")
    with open(program, "w") as out:
        out.write("#define main narra_original_main\n#include \"%s\"\n#undef main\n" % os.path.abspath(c_file))
        out.write("int main(void) {\n")
        for name, count in counts.items():
            out.write('  printf("%%zu %d %s\\n", sizeof(%s));\n' % (count, name, name))
        out.write("  return 0;\n}\n")
    binary = os.path.join(work, "sizes")
    subprocess.run(["cc", "-w", "-std=c11", "-o", binary, program, "-lm"], check=True)
    rows = []
    for line in subprocess.run([binary], check=True, capture_output=True, text=True).stdout.splitlines():
        size, count, name = line.split()
        rows.append((int(size), int(count), name))
    print("functions: %d" % len(definitions))
    total = sum(count for _, count, _ in rows)
    for limit in (16, 32, 64, 128):
        over = sum(count for size, count, _ in rows if size > limit)
        print("by-value struct/enum parameters over %d bytes: %d of %d" % (limit, over, total))
    print("\nsize parameters type (largest 25)")
    for size, count, name in sorted(rows, reverse=True)[:25]:
        print("%5d %5d  %s" % (size, count, name))


main()
