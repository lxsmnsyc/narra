// The grammar check of the build. It runs two checks over every `narra` fence of the guide, the
// installation guide and SPEC.md:
//
// 1. The check of `editors/textmate`, which fails on an `invalid` scope or a sample that ends
//    inside an open rule.
// 2. The same samples through the Narra language of twinkleplop, which fails on an `error` token,
//    a token class the site has no style for, a gap between tokens, and a string, template or
//    comment that is still open at the end of a sample.
import { spawnSync } from 'node:child_process';
import { readdirSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { tokens_to_named } from '@twinkleplop/core';
import { tokenize } from '../src/highlight/narra.ts';

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
  code: string;
}

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
  let current: { line: number; code: string[]; narra: boolean } | undefined;
  for (const [index, line] of lines.entries()) {
    if (!current) {
      const open = /^```(\S*)/.exec(line);
      if (open) {
        current = { line: index + 2, code: [], narra: open[1] === 'narra' };
      }
    } else if (line.startsWith('```')) {
      if (current.narra) {
        samples.push({ file, line: current.line, code: current.code.join('\n') });
      }
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

function checkSample(sample: Sample): string[] {
  const problems: string[] = [];
  const name = path.relative(root, sample.file);
  const run = tokenize();
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
      (last.type === 'comment' &&
        last.value.startsWith('/*') &&
        !last.value.endsWith('*/')) ||
      (last.type === 'template' && !last.value.endsWith('`')) ||
      (last.type === 'string' && /^["']/.test(last.value) && last.value.length < 2);
    if (open) {
      problems.push(`${name}:${locate(sample, last.start)}: the sample ends inside an open token`);
    }
  }
  return problems;
}

function checkTwinkleplop(files: string[]): number {
  const problems: string[] = [];
  let count = 0;
  for (const file of files) {
    for (const sample of collectSamples(file)) {
      count++;
      problems.push(...checkSample(sample));
    }
  }
  for (const problem of problems) {
    console.log(problem);
  }
  console.log(`twinkleplop: ${count} samples, ${problems.length} problems`);
  return problems.length;
}

function checkExamples(): number {
  // Number edge cases: a `.` continues a number only before a digit, and an exponent takes a sign.
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
const failed = checkExamples() + checkTextMate(files) + checkTwinkleplop(files);
process.exit(failed === 0 ? 0 : 1);
