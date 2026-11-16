// A small TOML language of twinkleplop for `narra.toml` and the editor settings in the guide.
// Keys are `identifier`, table names are `class_name`, and strings, numbers, booleans and
// comments have their usual classes.
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
  to_html,
} from '@twinkleplop/core';
import { compile, define_grammar, type GrammarRule } from '@twinkleplop/core/compile';
import * as token from '@twinkleplop/core/tokens';

const NAME_START = [LETTER, '_'];
const NAME_PART = [ALNUM, '_', '-'];

function listValueRules(): GrammarRule[] {
  // Shared by the start of a line and by a value, which differ in `=` and in `[`.
  return [
    on([' ', '\t', '\r']),
    match('#', token.comment, enter('comment')),
    match('"', token.string, enter('basic_string')),
    match("'", token.string, enter('literal_string')),
    keyword(['true', 'false'], {}, token.boolean),
    match(DIGIT, token.number, enter('number')),
    match(['{', '}', ',', '.'], token.punctuation),
    match(NAME_START, token.identifier, enter('key')),
  ];
}

const raw = define_grammar({
  name: 'toml',
  states: {
    // The start of a line: a table header or a key.
    main: {
      rules: [
        on('\n'),
        match(['[[', '['], token.punctuation, enter('header')),
        match('=', token.operator, goto('value')),
        ...listValueRules(),
        match(']', token.punctuation),
      ],
    },
    // After `=`, up to the end of the line.
    value: {
      rules: [
        on('\n', goto('main')),
        match('=', token.operator),
        match(['[', ']'], token.punctuation),
        ...listValueRules(),
      ],
    },
    header: {
      rules: [
        on([' ', '\t']),
        match([']]', ']'], token.punctuation, leave()),
        match('"', token.string, enter('basic_string')),
        match('.', token.punctuation),
        match(NAME_PART, token.class_name),
        fallback(leave()),
      ],
    },
    key: { rules: [match(NAME_PART, token.identifier), fallback(leave())] },
    number: {
      rules: [match([ALNUM, '_', '.', '-', ':'], token.number), fallback(leave())],
    },
    comment: {
      rules: [on('\n', leave()), fallback({ token: token.comment })],
    },
    basic_string: {
      rules: [
        match('\\', token.string_escape, enter('escape')),
        match('"', token.string, leave()),
        match('\n', 'error', leave()),
        fallback({ token: token.string }),
      ],
    },
    literal_string: {
      rules: [
        match("'", token.string, leave()),
        match('\n', 'error', leave()),
        fallback({ token: token.string }),
      ],
    },
    escape: { rules: [fallback({ token: token.string_escape, exit: true })] },
  },
});

export const grammar = compile(raw);

export const tokenize = create_language(grammar, []);

export function language(options?: Parameters<typeof tokenize>[0]) {
  const run = tokenize(options);
  return (source: string, render?: Parameters<typeof to_html>[2]): string =>
    to_html(source, run(source), render);
}
