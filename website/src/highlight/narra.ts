// The Narra language of twinkleplop. It uses the token classes of
// `editors/textmate/narra.tmLanguage.json`, mapped to the names of twinkleplop:
// keywords, storage words and modifiers are `keyword`, primitive types are `type`,
// capitalized names are `class_name`, UPPER_CASE names and `None` are `constant`,
// and a name before `(` is a `function`.
import {
  ALNUM,
  DIGIT,
  LETTER,
  create_language,
  enter,
  fallback,
  goto,
  keyword,
  leave,
  match,
  on,
  promote_by_text_set,
  promote_by_upper_snake_case,
  promote_function_calls,
  promote_pascal_case,
  to_html,
} from '@twinkleplop/core';
import { compile, define_grammar, type GrammarRule } from '@twinkleplop/core/compile';
import * as token from '@twinkleplop/core/tokens';

const KEYWORDS = [
  'if',
  'else',
  'while',
  'for',
  'of',
  'return',
  'break',
  'continue',
  'match',
  'yield',
  'await',
  'import',
  'export',
  'from',
  'as',
  'async',
  'mutable',
  'opaque',
  'distinct',
  'unsafe',
  'extern',
  'const',
  'let',
  'test',
  'struct',
  'enum',
  'trait',
  'implement',
  'type',
  'implements',
  'derives',
  'is',
  'in',
  'uses',
  'requires',
  'ensures',
  'typeof',
];

const PRIMITIVES = [
  'int',
  'bool',
  'char',
  'string',
  'void',
  'never',
  'f32',
  'f64',
  'i8',
  'i16',
  'i32',
  'i64',
  'u8',
  'u16',
  'u32',
  'u64',
  'usize',
  'isize',
];

// Marks an unterminated string, which the check of the build reports.
const ERROR = 'error';

const NAME_START = [LETTER, '_'];
const NAME_PART = [ALNUM, '_'];

const OPERATORS = [
  '===',
  '!==',
  '<<=',
  '>>=',
  '...',
  '==',
  '!=',
  '<=',
  '>=',
  '&&',
  '||',
  '=>',
  '..',
  '++',
  '--',
  '<<',
  '>>',
  '+=',
  '-=',
  '*=',
  '/=',
  '%=',
  '&=',
  '|=',
  '^=',
  '+',
  '-',
  '*',
  '/',
  '%',
  '=',
  '<',
  '>',
  '!',
  '&',
  '|',
  '^',
  '~',
  '?',
];

const EXPONENT_SIGNS = ['e+', 'e-', 'E+', 'E-'];

const FRACTION_STARTS = ['.0', '.1', '.2', '.3', '.4', '.5', '.6', '.7', '.8', '.9'];

const PUNCTUATION = ['(', ')', '[', ']', ',', ';', ':', '.', '#'];

function listCodeRules(braces: GrammarRule[]): GrammarRule[] {
  // Shared by the top level and by `${ }` in a template, which differ in their braces.
  return [
    ...braces,
    on([' ', '\t', '\n', '\r']),
    match('//', token.comment, enter('line_comment')),
    match('/*', token.comment, enter('block_comment')),
    match('"', token.string, enter('string_body')),
    match("'", token.string, enter('char_body')),
    match('`', token.template, enter('template_body')),
    match('@', token.decorator, enter('annotation')),
    keyword(['true', 'false'], {}, token.boolean),
    keyword(['None'], {}, token.constant),
    keyword(PRIMITIVES, {}, token.type),
    keyword(['function'], enter('function_name')),
    keyword(KEYWORDS),
    // A radix prefix: the digits after it may be letters, and `e-` is not an exponent there.
    match(['0x', '0X', '0b', '0B', '0o', '0O'], token.number, enter('radix')),
    match(DIGIT, token.number, enter('number')),
    match(OPERATORS, token.operator),
    match(PUNCTUATION, token.punctuation),
    match(NAME_START, token.identifier, enter('name')),
  ];
}

const raw = define_grammar({
  name: 'narra',
  states: {
    main: { rules: listCodeRules([match(['{', '}'], token.punctuation)]) },
    // `${ ... }` in a template: the code rules, with braces counted so the `}` that closes the
    // substitution leaves.
    substitution: {
      rules: listCodeRules([
        match('{', token.punctuation, enter('substitution')),
        match('}', token.punctuation, leave()),
      ]),
    },
    name: { rules: [match(NAME_PART, token.identifier), fallback(leave())] },
    annotation: { rules: [match(NAME_PART, token.decorator), fallback(leave())] },
    // The name after `function`, which may follow `*` and spaces.
    function_name: {
      rules: [
        on([' ', '\t']),
        match('*', token.operator),
        match(NAME_START, token.function, goto('function_name_rest')),
        fallback(leave()),
      ],
    },
    function_name_rest: { rules: [match(NAME_PART, token.function), fallback(leave())] },
    radix: { rules: [match(NAME_PART, token.number), fallback(leave())] },
    // A `.` continues a number only when a digit follows, so `0..10` and `pair.0.first` split
    // there. An exponent may take a sign.
    number: {
      rules: [
        match(EXPONENT_SIGNS, token.number),
        match(NAME_PART, token.number),
        match(FRACTION_STARTS, token.number, goto('fraction')),
        fallback(leave()),
      ],
    },
    fraction: {
      rules: [
        match(EXPONENT_SIGNS, token.number),
        match(NAME_PART, token.number),
        fallback(leave()),
      ],
    },
    line_comment: {
      rules: [on('\n', leave()), match(['\r'], token.comment), fallback({ token: token.comment })],
    },
    block_comment: {
      rules: [
        match('*/', token.comment, leave()),
        fallback({ token: token.comment }),
      ],
    },
    string_body: {
      rules: [
        match('\\', token.string_escape, enter('escape')),
        match('"', token.string, leave()),
        match('\n', ERROR, leave()),
        fallback({ token: token.string }),
      ],
    },
    char_body: {
      rules: [
        match('\\', token.string_escape, enter('escape')),
        match("'", token.string, leave()),
        match('\n', ERROR, leave()),
        fallback({ token: token.string }),
      ],
    },
    template_body: {
      rules: [
        match('\\', token.string_escape, enter('escape')),
        match('${', token.punctuation, enter('substitution')),
        match('`', token.template, leave()),
        fallback({ token: token.template }),
      ],
    },
    // One character after a backslash. `\u{1F600}` continues as string text.
    escape: { rules: [fallback({ token: token.string_escape, exit: true })] },
  },
});

export const grammar = compile(raw);

// Names are told apart by their shape, as in the TextMate grammar.
export const reclassifiers = [
  promote_function_calls(token.identifier, token.function, { plain: true, generic_fn: true }),
  promote_by_upper_snake_case(token.identifier, token.constant),
  promote_pascal_case(token.identifier, token.class_name),
  promote_by_text_set(token.identifier, token.class_name, ['T', 'U', 'K', 'V', 'E']),
];

export const tokenize = create_language(grammar, reclassifiers);

export function language(options?: Parameters<typeof tokenize>[0]) {
  const run = tokenize(options);
  return (source: string, render?: Parameters<typeof to_html>[2]): string =>
    to_html(source, run(source), render);
}
