// The shell language of twinkleplop for the commands of the guide. The `bash` package splits
// `--json` into an operator and a name and `https://x.dev/a` into five tokens, and it styles `cd`
// but not `git`. Here a command is a `function`, a shell word such as `cd` or `export` is a
// `keyword`, a flag is one `keyword` token, a variable is a `constant`, and every other word, such
// as a path or a URL, is one `identifier`.
import {
  ALNUM,
  create_language,
  enter,
  fallback,
  goto,
  keyword,
  leave,
  match,
  on,
  rewrite_types,
  to_html,
  type as tokenType,
} from '@twinkleplop/core';
import { compile, define_grammar } from '@twinkleplop/core/compile';
import * as token from '@twinkleplop/core/tokens';

const SHELL_WORDS = [
  'cd',
  'command',
  'case',
  'do',
  'done',
  'elif',
  'else',
  'esac',
  'exit',
  'export',
  'fi',
  'for',
  'if',
  'in',
  'local',
  'return',
  'set',
  'source',
  'then',
  'unset',
  'until',
  'while',
];

// The characters of a word: names, paths, URLs and versions.
const WORD = [ALNUM, '_', '-', '.', '/', ':', '@', '%', '+', '~', ',', '*', '?'];
const WORD_START = [ALNUM, '_', '.', '/', '~', '*', '?', '@', '%', '+', ',', ':'];
const SEPARATORS = ['||', '&&', '|', ';', '&'];
const REDIRECTIONS = ['>>', '>', '<'];

const QUOTED_AND_VARIABLES = [
  match('#', token.comment, enter('comment')),
  match('"', token.string, enter('double_string')),
  match("'", token.string, enter('single_string')),
  match('$', token.constant, enter('variable')),
];

const raw = define_grammar({
  name: 'sh',
  states: {
    // The start of a command: after a line break, `|`, `&&` or `;`.
    main: {
      rules: [
        on([' ', '\t', '\r', '\n']),
        ...QUOTED_AND_VARIABLES,
        keyword(SHELL_WORDS, goto('args')),
        match('=', token.operator, goto('assignment')),
        match(WORD_START, token.function, goto('command_word')),
      ],
    },
    command_word: {
      rules: [
        match(WORD, token.function),
        match('=', token.operator, goto('assignment')),
        fallback(goto('args')),
      ],
    },
    // The value of `NAME=value` before a command.
    assignment: {
      rules: [
        match('"', token.string, enter('double_string')),
        match("'", token.string, enter('single_string')),
        match('$', token.constant, enter('variable')),
        match(WORD, token.identifier),
        fallback(goto('main')),
      ],
    },
    args: {
      rules: [
        on([' ', '\t', '\r']),
        on('\n', goto('main')),
        match(SEPARATORS, token.operator, goto('main')),
        match(REDIRECTIONS, token.operator),
        ...QUOTED_AND_VARIABLES,
        match('-', token.keyword, enter('flag')),
        match([...WORD_START, '='], token.identifier, enter('word')),
        match(['\\', '(', ')', '{', '}', '[', ']', '=', '!'], token.punctuation),
      ],
    },
    word: { rules: [match([...WORD, '='], token.identifier), fallback(leave())] },
    flag: { rules: [match([ALNUM, '_', '-'], token.keyword), fallback(leave())] },
    variable: {
      rules: [
        match([ALNUM, '_'], token.constant),
        match('{', token.constant, goto('braced_variable')),
        fallback(leave()),
      ],
    },
    braced_variable: {
      rules: [match('}', token.constant, leave()), fallback({ token: token.constant })],
    },
    comment: { rules: [on('\n', leave()), fallback({ token: token.comment })] },
    double_string: {
      rules: [
        match('\\', token.string_escape, enter('escape')),
        match('$', token.constant, enter('variable')),
        match('"', token.string, leave()),
        fallback({ token: token.string }),
      ],
    },
    single_string: {
      rules: [match("'", token.string, leave()), fallback({ token: token.string })],
    },
    escape: { rules: [fallback({ token: token.string_escape, exit: true })] },
  },
});

export const grammar = compile(raw);

export const reclassifiers = [
  // A word before `=` in the command position is a variable that the command runs with.
  rewrite_types([
    {
      anchor: token.function,
      when: tokenType(token.operator, '='),
      rewrite: token.constant,
    },
  ]),
];

export const tokenize = create_language(grammar, reclassifiers);

export function language(options?: Parameters<typeof tokenize>[0]) {
  const run = tokenize(options);
  return (source: string, render?: Parameters<typeof to_html>[2]): string =>
    to_html(source, run(source), render);
}
