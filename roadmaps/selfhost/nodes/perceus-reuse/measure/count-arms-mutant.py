# Adds a narra_stat_arm call at the start and end of each match and if-is arm on a heap enum subject (selfhost/src/emitmatch.nra).
import sys
p = sys.argv[1]
s = open(p).read()
s = s.replace('import { Ref, getExpressionId, findBodySlot } from "narrac/check"', 'import { Ref, getExpressionId, findBodySlot, isHeapEnum } from "narrac/check"')
s = s.replace('  buildVariantFieldPath,\n} from "narrac/emitstate"', '  buildVariantFieldPath,\n  underlyingType,\n} from "narrac/emitstate"')
s = s.replace('''function holdSubject(''', '''function statArm(mutable emitter: Emitter, held: string, subjectType: Type, owned: bool, start: bool) uses alloc {
  if (underlyingType(emitter, subjectType) is Type.TEnum(statIndex)) {
    if (isHeapEnum(emitter.checker, statIndex)) {
      if (start) {
        let flag = "0"
        if (owned) {
          flag = "1"
        }
        writeLine(mutable emitter, `narra_stat_arm(${held}.u, sizeof(*${held}.u), ${flag}, __builtin_frame_address(0));`)
      } else {
        writeLine(mutable emitter, "narra_stat_arm_end(__builtin_frame_address(0));")
      }
    }
  }
}

function holdSubject(''')
old = '''    writeLine(mutable emitter, `if (${holds}) {`)
    emitter.indent++
    pushScope(mutable emitter, false)
    bindPattern(walker, mutable emitter, arm.pattern, held, false, true)
    if (arm.guard is Some(guard)) {'''
new = '''    writeLine(mutable emitter, `if (${holds}) {`)
    emitter.indent++
    statArm(mutable emitter, held, subjectType, code.owned, true)
    pushScope(mutable emitter, false)
    bindPattern(walker, mutable emitter, arm.pattern, held, false, true)
    if (arm.guard is Some(guard)) {'''
assert old in s
s = s.replace(old, new)
old = '''    popScope(mutable emitter)
    emitter.indent--
    writeLine(mutable emitter, "}")
    emitter.indent--
    writeLine(mutable emitter, "}")
  }
}'''
new = '''    popScope(mutable emitter)
    statArm(mutable emitter, held, subjectType, code.owned, false)
    emitter.indent--
    writeLine(mutable emitter, "}")
    emitter.indent--
    writeLine(mutable emitter, "}")
  }
}'''
assert old in s
s = s.replace(old, new)
old = '''      writeLine(mutable emitter, `if (${holds}) {`)
      emitter.indent++
      pushScope(mutable emitter, false)
      bindPattern(walker, mutable emitter, pattern, held, false, true)
      EmitWalker.emitBlockInScope(walker, mutable emitter, then)
      popScope(mutable emitter)'''
new = '''      writeLine(mutable emitter, `if (${holds}) {`)
      emitter.indent++
      statArm(mutable emitter, held, subjectType, code.owned, true)
      pushScope(mutable emitter, false)
      bindPattern(walker, mutable emitter, pattern, held, false, true)
      EmitWalker.emitBlockInScope(walker, mutable emitter, then)
      popScope(mutable emitter)
      statArm(mutable emitter, held, subjectType, code.owned, false)'''
assert old in s
s = s.replace(old, new)
open(p, 'w').write(s)
