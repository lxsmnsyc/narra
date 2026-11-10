import sys, re
# Hand-made unsound reuse: each arm that builds a List cell writes into the subject's
# cell without checking that the subject is unique and owned.
source, target = sys.argv[1], sys.argv[2]
s = open(source).read()
subjects = {'q19': 't14', 'q41': 't36', 'q50': 't45'}
for cell, subject in subjects.items():
    old = f'Q_e9e541db_List* {cell} = narra_alloc(sizeof(Q_e9e541db_List));\n      {cell}->rc = 1;'
    assert old in s, cell
    new = (f'Q_e9e541db_List* {cell} = {subject}.u; lstr_release({cell}->v_Cons.f_head); '
           f'release_E_e9e541db_List({cell}->v_Cons.f_tail);\n      {cell}->rc++;')
    s = s.replace(old, new)
open(target, 'w').write(s)
