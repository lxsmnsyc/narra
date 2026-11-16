// A small Lua language of twinkleplop for the editor settings in the guide: keywords, strings,
// `--` comments and numbers. A name before `(` is a function.
import {
  ALNUM,
  DIGIT,
  LETTER,
  create_language,
  enter,
  fallback,
  keyword,
  leave,
  match,
  on,
  promote_by_upper_snake_case,
  promote_function_calls,
  to_html,
} from '@twinkleplop/core';
import { compile, define_grammar } from '@twinkleplop/core/compile';
import * as token from '@twinkleplop/core/tokens';

const KEYWORDS = [
  'and',
  'break',
  'do',
  'else',
  'elseif',
  'end',
  'for',
  'function',
  'goto',
  'if',
  'in',
  'local',
  'not',
  'or',
  'repeat',
  'return',
  'then',
  'until',
  'while',
];

const OPERATORS = ['==', '~=', '<=', '>=', '..', '+', '-', '*', '/', '%', '^', '#', '=', '<', '>'];

const NAME_START = [LETTER, '_'];
const NAME_PART = [ALNUM, '_'];

const raw = define_grammar({
  name: 'lua',
  states: {
    main: {
      rules: [
        on([' ', '\t', '\n', '\r']),
        match('--', token.comment, enter('comment')),
        match('"', token.string, enter('double_string')),
        match("'", token.string, enter('single_string')),
        keyword(['true', 'false'], {}, token.boolean),
        keyword(['nil'], {}, token.constant),
        keyword(KEYWORDS),
        match(DIGIT, token.number, enter('number')),
        match(OPERATORS, token.operator),
        match(['(', ')', '[', ']', '{', '}', ',', ';', ':', '.'], token.punctuation),
        match(NAME_START, token.identifier, enter('name')),
      ],
    },
    name: { rules: [match(NAME_PART, token.identifier), fallback(leave())] },
    number: { rules: [match([ALNUM, '.'], token.number), fallback(leave())] },
    comment: { rules: [on('\n', leave()), fallback({ token: token.comment })] },
    double_string: {
      rules: [
        match('\\', token.string_escape, enter('escape')),
        match('"', token.string, leave()),
        match('\n', 'error', leave()),
        fallback({ token: token.string }),
      ],
    },
    single_string: {
      rules: [
        match('\\', token.string_escape, enter('escape')),
        match("'", token.string, leave()),
        match('\n', 'error', leave()),
        fallback({ token: token.string }),
      ],
    },
    escape: { rules: [fallback({ token: token.string_escape, exit: true })] },
  },
});

export const grammar = compile(raw);

export const reclassifiers = [
  promote_function_calls(token.identifier, token.function, { plain: true }),
  promote_by_upper_snake_case(token.identifier, token.constant),
];

export const tokenize = create_language(grammar, reclassifiers);

export function language(options?: Parameters<typeof tokenize>[0]) {
  const run = tokenize(options);
  return (source: string, render?: Parameters<typeof to_html>[2]): string =>
    to_html(source, run(source), render);
}
