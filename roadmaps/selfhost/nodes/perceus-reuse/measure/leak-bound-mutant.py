import sys
# Upper bound experiment: never release an owned heap-enum subject of match or if-is (leaks it).
p = sys.argv[1]
s = open(p).read()
s = s.replace('import { Ref, getExpressionId, findBodySlot } from "narrac/check"', 'import { Ref, getExpressionId, findBodySlot, isHeapEnum } from "narrac/check"')
s = s.replace('  buildVariantFieldPath,\n} from "narrac/emitstate"', '  buildVariantFieldPath,\n  underlyingType,\n} from "narrac/emitstate"')
s = s.replace('''function holdSubject(''', '''function isHeapSubject(emitter: Emitter, subjectType: Type): bool {
  if (underlyingType(emitter, subjectType) is Type.TEnum(statIndex)) {
    return isHeapEnum(emitter.checker, statIndex)
  }
  return false
}

function holdSubject(''')
old = '''  emitter.tempCount++
  const doneName = `m${Int.toString(emitter.tempCount)}`'''
new = '''  if (code.owned && isHeapSubject(emitter, subjectType)) {
    writeLine(mutable emitter, `${code.text}.u = NULL;`)
  }
  emitter.tempCount++
  const doneName = `m${Int.toString(emitter.tempCount)}`'''
assert old in s
s = s.replace(old, new)
old = '''        endStatement(mutable emitter)
        holdSubject(mutable emitter, held, subjectType)'''
new = '''        endStatement(mutable emitter)
        if (!(code.owned && isHeapSubject(emitter, subjectType))) {
          holdSubject(mutable emitter, held, subjectType)
        }'''
assert old in s
s = s.replace(old, new)
open(p, 'w').write(s)
