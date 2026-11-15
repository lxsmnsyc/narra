"""Check the Narra examples in a markdown file against a built narrac.

Usage: python3 check-examples.py NARRAC FILE.md

A fence tagged `narra` must pass `narrac check`. A file with no `run` gets a stub, because a single file
needs an entry function. A fence tagged
`narra error CODE` must fail with that code in `narrac check --json`.
Exit code 0 when every example behaves as tagged.
"""

import json
import re
import subprocess
import sys
import tempfile
from pathlib import Path

FENCE = re.compile(r"^```narra(?: error (E\d{4}))?\n(.*?)^```$", re.S | re.M)


def check_example(narrac, source, expected_code):
    with tempfile.TemporaryDirectory() as folder:
        path = Path(folder) / "example.nra"
        if "function run(" not in source:
            source += "\nexport function run(args: Array<string>): int {\n  return 0\n}\n"
        path.write_text(source)
        result = subprocess.run(
            [narrac, "check", "--json", str(path)],
            capture_output=True,
            text=True,
        )
    if expected_code is None:
        return result.returncode == 0, result.stdout + result.stderr
    codes = []
    for line in result.stdout.splitlines():
        if line.startswith("{"):
            codes.append(json.loads(line)["code"])
    return expected_code in codes, "codes: " + ", ".join(codes)


def main():
    narrac, markdown = sys.argv[1], Path(sys.argv[2]).read_text()
    failures = 0
    count = 0
    for match in FENCE.finditer(markdown):
        count += 1
        line = markdown[: match.start()].count("\n") + 1
        passed, detail = check_example(narrac, match.group(2), match.group(1))
        if not passed:
            failures += 1
            print(f"line {line}: FAIL ({match.group(1) or 'must compile'}): {detail}")
    print(f"{count - failures} of {count} examples as tagged")
    sys.exit(1 if failures else 0)


main()
