import re,sys,pathlib,collections
files=[p for d in sys.argv[1:] for p in pathlib.Path(d).rglob('*.nra') if p.is_file()]
ctors=collections.defaultdict(set)  # name -> enums
for p in files:
    t=p.read_text()
    for m in re.finditer(r'enum\s+(\w+)[^{]*\{(.*?)\n\}',t,re.S):
        for c in re.finditer(r'^\s*([A-Z]\w*)',m.group(2),re.M): ctors[c.group(1)].add(m.group(1))
for b in ['Some','None','Ok','Err']: ctors[b].add('prelude')
bare=qual=pat=0; user_bare=0; per=collections.Counter()
for p in files:
    for line in p.read_text().splitlines():
        s=line.strip()
        if s.startswith('//'): continue
        for m in re.finditer(r'(\.?)\b([A-Z]\w*)\b',line):
            n=m.group(2)
            if n not in ctors: continue
            if m.group(1): qual+=1; continue
            # match arm pattern: line has => after it, or `is`
            rest=line[m.end():]; before=line[:m.start()]
            if '=>' in rest and '=>' not in before or re.search(r'\bis\s*$',before) or re.search(r'\bcase\b',before): pat+=1; continue
            if re.search(r'\benum\b|^\s*$',before) : continue
            bare+=1; per[n]+=1
            if 'prelude' not in ctors[n]: user_bare+=1
clash={k:v for k,v in ctors.items() if len(v-{'prelude'})>1 or ('prelude' in v and len(v)>1)}
print('files',len(files),'ctors',len(ctors),'qualified',qual,'patterns',pat,'bare_expr',bare,'bare_expr_user_enums',user_bare)
print('clashing names', {k:sorted(v) for k,v in list(clash.items())[:15]}, len(clash))
print('top bare', per.most_common(12))
