# Histogram of parameter counts for every function declaration in selfhost/src and std,
# including multi-line signatures. Counts top-level commas between the parameter parens.
import pathlib, re, collections, sys
root = pathlib.Path(sys.argv[1])
hist = collections.Counter()
for path in list(root.glob("selfhost/src/**/*.nra")) + list(root.glob("std/**/*.nra")):
    text = path.read_text()
    for match in re.finditer(r"^\s*(?:export )?function\*? \w+(?:<[^>]*>)?\(", text, re.M):
        depth, i, params, current = 1, match.end(), 0, ""
        while depth and i < len(text):
            ch = text[i]
            if ch in "(<[{": depth += 1
            elif ch in ")>]}": depth -= 1
            if depth == 1 and ch == ",": params += 1; current = ""
            elif depth >= 1: current += ch
            i += 1
        if current.strip() or params: params += 1 if current.strip() else 0
        hist[params] += 1
for k in sorted(hist): print(f"{k} params: {hist[k]}")
print("total:", sum(hist.values()), " >=4:", sum(v for k, v in hist.items() if k >= 4), " >=5:", sum(v for k, v in hist.items() if k >= 5))
