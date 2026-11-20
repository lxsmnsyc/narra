// The grammar check of the build. It reads every fence of the guide, the installation guide and
// SPEC.md and runs these checks:
//
// 1. Fences: each one names a language the site highlights, and the site has that language.
// 2. The check of `editors/textmate`, which fails on an `invalid` scope or a Narra sample that
//    ends inside an open rule.
// 3. The Narra, TOML and Lua samples through their twinkleplop languages, which fails on an
//    `error` token, a token class the site has no style for, a gap between tokens, and a string,
//    template or comment that is still open at the end of a sample.
// 4. The Narra samples through the TextMate grammar and the twinkleplop language, which fails on
//    a character that the two give different classes.
import { spawnSync } from 'node:child_process';
import { readdirSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { tokens_to_named } from '@twinkleplop/core';
import { LANGUAGES } from '../plugins/markdown.ts';
import { FENCE_LANGUAGES, readFenceLanguage } from '../src/highlight/fences.ts';
import { tokenize as tokenizeLua } from '../src/highlight/lua.ts';
import { tokenize } from '../src/highlight/narra.ts';
import { tokenize as tokenizeSh } from '../src/highlight/sh.ts';
import { classifyTextMate } from '../src/highlight/textmate.ts';
import { tokenize as tokenizeToml } from '../src/highlight/toml.ts';

const website = fileURLToPath(new URL('..', import.meta.url));
const root = path.resolve(website, '..');

/** The token classes that `src/styles.css` and the theme of twinkleplop style. */
const KNOWN = new Set([
  'boolean',
  'class_name',
  'comment',
  'constant',
  'decorator',
  'function',
  'identifier',
  'keyword',
  'number',
  'operator',
  'punctuation',
  'string',
  'string_escape',
  'template',
  'type',
]);


interface Sample {
  file: string;
  line: number;
  info: string;
  language: string;
  code: string;
}

/** The languages that have a twinkleplop language in `src/highlight`, checked token by token. */
const TOKENIZERS = {
  narra: tokenize,
  sh: tokenizeSh,
  bash: tokenizeSh,
  toml: tokenizeToml,
  lua: tokenizeLua,
} as const;

function listFiles(): string[] {
  const guide = path.join(root, 'docs', 'guide');
  const files: string[] = [];
  for (const name of readdirSync(guide).sort()) {
    if (name.endsWith('.md')) {
      files.push(path.join(guide, name));
    }
  }
  files.push(path.join(root, 'docs', 'install.md'), path.join(root, 'SPEC.md'));
  return files;
}

function collectSamples(file: string): Sample[] {
  const samples: Sample[] = [];
  const lines = readFileSync(file, 'utf8').split('\n');
  let current: { line: number; info: string; code: string[] } | undefined;
  for (const [index, line] of lines.entries()) {
    if (!current) {
      const open = /^```(.*)$/.exec(line);
      if (open) {
        current = { line: index + 2, info: open[1]!.trim(), code: [] };
      }
    } else if (line.startsWith('```')) {
      samples.push({
        file,
        line: current.line,
        info: current.info,
        language: readFenceLanguage(current.info),
        code: current.code.join('\n'),
      });
      current = undefined;
    } else {
      current.code.push(line);
    }
  }
  return samples;
}

function locate(sample: Sample, offset: number): number {
  return sample.line + sample.code.slice(0, offset).split('\n').length - 1;
}

function checkFence(sample: Sample): string[] {
  // The opening fence is the line before the first line of code.
  const where = `${path.relative(root, sample.file)}:${sample.line - 1}`;
  if (sample.language === '') {
    return [`${where}: the fence "${sample.info}" has no language; tag it, or use text`];
  }
  if (!FENCE_LANGUAGES.includes(sample.language)) {
    return [`${where}: the fence language "${sample.language}" is not one the site highlights`];
  }
  if (sample.language !== 'text' && !(sample.language in LANGUAGES)) {
    return [`${where}: the site has no highlighter for "${sample.language}"`];
  }
  return [];
}

function checkSample(sample: Sample): string[] {
  const problems: string[] = [];
  const name = path.relative(root, sample.file);
  const run = TOKENIZERS[sample.language as keyof typeof TOKENIZERS]();
  const tokens = tokens_to_named(run(sample.code), sample.code);
  let end = 0;
  for (const item of tokens) {
    const where = `${name}:${locate(sample, item.start)}`;
    if (item.start !== end && sample.code.slice(end, item.start).trim() !== '') {
      problems.push(`${where}: text without a token, "${sample.code.slice(end, item.start)}"`);
    }
    end = item.end;
    if (item.type === 'error') {
      problems.push(`${where}: unterminated string`);
    } else if (!KNOWN.has(item.type)) {
      problems.push(`${where}: unknown token class "${item.type}" on "${item.value}"`);
    }
  }
  // A string, template or block comment that the sample ends inside.
  const last = tokens.at(-1);
  if (last) {
    const open =
      (last.type === 'comment' && last.value.startsWith('/*') && !last.value.endsWith('*/')) ||
      (last.type === 'template' && !last.value.endsWith('`')) ||
      // The shell language splits a string at each variable, so its last token can be a quote.
      (last.type === 'string' &&
        !(sample.language === 'sh' || sample.language === 'bash') &&
        /^["']/.test(last.value) &&
        last.value.length < 2);
    if (open) {
      problems.push(`${name}:${locate(sample, last.start)}: the sample ends inside an open token`);
    }
  }
  return problems;
}

/**
 * The differences between the classes of the TextMate grammar and the classes of the site that
 * are not bugs of the site. The TextMate grammar reads the value of a struct field (`name: value`)
 * and the `from` clause of a function as a type annotation, and a type annotation runs to the next
 * `,`, `)` or `=`, so it leaves `{`, `:` and `>` there without a scope, reads a name before `(` as
 * a plain name, and reads `from` and `typeof` as names. It also reads the `<` of type arguments in
 * an expression, such as `Map<string, int> {}`, as a comparison. The site classifies these by
 * their shape.
 */
function explainDifference(textmate: string, site: string, text: string): string | undefined {
  if (textmate === '' && (site === 'punctuation' || site === 'operator')) {
    return 'inside a type annotation that the TextMate grammar did not end';
  }
  if (textmate === 'operator' && site === 'punctuation' && /^[<>]+$/.test(text)) {
    return 'type arguments in an expression';
  }
  if (textmate === 'identifier' && site === 'keyword' && (text === 'from' || text === 'typeof')) {
    return 'a keyword inside a type annotation';
  }
  if (textmate === 'identifier' && site === 'keyword' && /^(for|while|const|let|of)$/.test(text)) {
    return 'the loop after a label, which the TextMate grammar reads as a type annotation';
  }
  if (textmate === 'identifier' && site === 'function') {
    return 'a call inside a type annotation';
  }
  if (textmate === 'class_name' && site === 'type' && text === 'int') {
    return 'a primitive type after `for`';
  }
  if (textmate === 'class_name' && site === 'constant' && text === 'None') {
    return 'None inside a type annotation';
  }
  if (textmate === 'function' && (text === 'Template' || text === '.')) {
    return 'the dotted tag of a tagged template';
  }
  return undefined;
}

async function compareSample(sample: Sample): Promise<string[]> {
  const problems: string[] = [];
  const name = path.relative(root, sample.file);
  const code = sample.code;
  const textmate = await classifyTextMate(code);
  const site: string[] = new Array<string>(code.length).fill('');
  for (const item of tokens_to_named(tokenize()(code), code)) {
    for (let index = item.start; index < item.end; index++) {
      site[index] = item.type;
    }
  }
  let start = 0;
  while (start < code.length) {
    if (/\s/.test(code[start]!) || textmate[start] === site[start]) {
      start++;
      continue;
    }
    // The run of characters with the same pair of classes.
    let end = start;
    while (
      end < code.length &&
      !/\s/.test(code[end]!) &&
      textmate[end] === textmate[start] &&
      site[end] === site[start]
    ) {
      end++;
    }
    const text = code.slice(start, end);
    if (!explainDifference(textmate[start]!, site[start]!, text)) {
      problems.push(
        `${name}:${locate(sample, start)}: TextMate says ${textmate[start] || 'nothing'}, the site says ${site[start]}, on "${text}"`,
      );
    }
    start = end;
  }
  return problems;
}

async function checkSamples(files: string[]): Promise<number> {
  const problems: string[] = [];
  let fences = 0;
  let checked = 0;
  let compared = 0;
  for (const file of files) {
    for (const sample of collectSamples(file)) {
      fences++;
      const fenceProblems = checkFence(sample);
      problems.push(...fenceProblems);
      if (fenceProblems.length > 0 || !(sample.language in TOKENIZERS)) {
        continue;
      }
      checked++;
      problems.push(...checkSample(sample));
      if (sample.language === 'narra') {
        compared++;
        problems.push(...(await compareSample(sample)));
      }
    }
  }
  for (const problem of problems) {
    console.log(problem);
  }
  console.log(
    `fences: ${fences}, tokenized: ${checked}, compared with TextMate: ${compared}, ${problems.length} problems`,
  );
  return problems.length;
}

function checkExamples(): number {
  // Number edge cases: a `.` continues a number only before a digit, and an exponent takes a sign.
  // Then the contexts of the grammar: type arguments, effect lists and unicode escapes.
  const examples: { source: string; expected: [string, string][] }[] = [
    { source: '1.0e-9', expected: [['number', '1.0e-9']] },
    { source: '2E+5', expected: [['number', '2E+5']] },
    { source: '0xFE-1', expected: [['number', '0xFE'], ['operator', '-'], ['number', '1']] },
    { source: '0..10', expected: [['number', '0'], ['operator', '..'], ['number', '10']] },
    {
      source: 'pair.0.first',
      expected: [
        ['identifier', 'pair'],
        ['punctuation', '.'],
        ['number', '0'],
        ['punctuation', '.'],
        ['identifier', 'first'],
      ],
    },
    {
      source: 'Map<string, int>',
      expected: [
        ['class_name', 'Map'],
        ['punctuation', '<'],
        ['type', 'string'],
        ['punctuation', ','],
        ['type', 'int'],
        ['punctuation', '>'],
      ],
    },
    {
      source: 'a < b',
      expected: [['identifier', 'a'], ['operator', '<'], ['identifier', 'b']],
    },
    {
      source: 'parse<int>(x)',
      expected: [
        ['function', 'parse'],
        ['punctuation', '<'],
        ['type', 'int'],
        ['punctuation', '>('],
        ['identifier', 'x'],
        ['punctuation', ')'],
      ],
    },
    {
      source: 'uses io, alloc {',
      expected: [
        ['keyword', 'uses'],
        ['class_name', 'io'],
        ['punctuation', ','],
        ['class_name', 'alloc'],
        ['punctuation', '{'],
      ],
    },
    {
      source: '"\\u{1F600}"',
      expected: [['string', '"'], ['string_escape', '\\u{1F600}'], ['string', '"']],
    },
    {
      source: 'html`<b>${x}</b>`',
      expected: [
        ['function', 'html'],
        ['template', '`<b>'],
        ['punctuation', '${'],
        ['identifier', 'x'],
        ['punctuation', '}'],
        ['template', '</b>`'],
      ],
    },
  ];
  let failed = 0;
  for (const example of examples) {
    const tokens = tokens_to_named(tokenize()(example.source), example.source);
    const actual: [string, string][] = [];
    for (const item of tokens) {
      // Neighbors of one class show as one run.
      const last = actual.at(-1);
      if (last?.[0] === item.type && item.type !== 'punctuation') {
        last[1] += item.value;
      } else {
        actual.push([item.type, item.value]);
      }
    }
    if (JSON.stringify(actual) !== JSON.stringify(example.expected)) {
      console.log(`example ${example.source}: got ${JSON.stringify(actual)}`);
      failed++;
    }
  }
  console.log(`twinkleplop examples: ${examples.length}, ${failed} failed`);
  return failed;
}

function checkTextMate(files: string[]): number {
  const result = spawnSync(process.execPath, [path.join(root, 'editors/textmate/check.cjs'), ...files], {
    env: { ...process.env, NODE_PATH: path.join(website, 'node_modules') },
    stdio: 'inherit',
  });
  return result.status === 0 ? 0 : 1;
}

const files = listFiles();
const failed = checkExamples() + checkTextMate(files) + (await checkSamples(files));
process.exit(failed === 0 ? 0 : 1);
