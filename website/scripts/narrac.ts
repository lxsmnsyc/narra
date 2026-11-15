#!/usr/bin/env node
// A stand-in for `narrac query --json hovers`, so the site builds before the real command lands.
// It answers in the shape of the frozen query-hovers contract, but it guesses: a name that a
// `const`, `let`, `function` or parameter declares gets a hover, and the diagnostic comes from the
// `// error: LINE:COLUMN: message` comment of the sample. Point `NARRAC` at this file to use it.
import { readFileSync } from 'node:fs';
import path from 'node:path';

const KEYWORDS = new Set([
  'import',
  'export',
  'from',
  'as',
  'function',
  'const',
  'let',
  'return',
  'if',
  'else',
  'for',
  'of',
  'while',
  'match',
  'struct',
  'enum',
  'trait',
  'impl',
  'true',
  'false',
  'uses',
  'type',
  'int',
  'f64',
  'string',
  'bool',
  'void',
  'Array',
  'Option',
  'Result',
  'Some',
  'None',
  'Ok',
  'Err',
]);

interface Hover {
  line: number;
  column: number;
  length: number;
  text: string;
}

function guessType(initializer: string): string {
  const text = initializer.trim();
  if (/^-?\d+$/.test(text)) {
    return 'int';
  }
  if (/^-?\d+\.\d+$/.test(text)) {
    return 'f64';
  }
  if (text.startsWith('"')) {
    return 'string';
  }
  if (/^(true|false)\b/.test(text)) {
    return 'bool';
  }
  return '';
}

function maskText(line: string): string {
  // Comments and strings become spaces, so no name inside them gets a hover.
  const characters = Array.from(line);
  let inString = false;
  let inComment = false;
  for (const [at, character] of characters.entries()) {
    if (!inString && character === '/' && characters[at + 1] === '/') {
      inComment = true;
    }
    if (!inComment && character === '"') {
      inString = !inString;
      characters[at] = ' ';
    } else if (inComment || inString) {
      characters[at] = ' ';
    }
  }
  return characters.join('');
}

function readHovers(lines: string[]): Hover[] {
  const types = new Map<string, string>();
  const hovers: Hover[] = [];
  for (const [index, line] of lines.entries()) {
    const declaration =
      /^\s*(?:export\s+)?(?:const|let)\s+([A-Za-z_]\w*)\s*(?::\s*([^=]+?))?\s*=\s*(.*)$/u.exec(line);
    if (declaration) {
      types.set(declaration[1]!, declaration[2] ?? guessType(declaration[3]!));
    }
    const signature = /^\s*(?:export\s+)?function\*?\s+([A-Za-z_]\w*)\s*\(([^)]*)\)/u.exec(line);
    if (signature) {
      types.set(signature[1]!, `function ${signature[1]}(${signature[2]})`);
      for (const parameter of signature[2]!.split(',')) {
        const parts = /^\s*([A-Za-z_]\w*)\s*:\s*(.+?)\s*$/u.exec(parameter);
        if (parts) {
          types.set(parts[1]!, parts[2]!);
        }
      }
    }
    const text = maskText(line);
    for (const match of text.matchAll(/[A-Za-z_]\w*/gu)) {
      const name = match[0];
      const type = types.get(name);
      // A type the stand-in cannot work out gets no hover, never a placeholder.
      if (KEYWORDS.has(name) || !type) {
        continue;
      }
      const column = Array.from(text.slice(0, match.index)).length + 1;
      hovers.push({ line: index + 1, column, length: Array.from(name).length, text: type });
    }
  }
  return hovers;
}

function run(args: string[]): number {
  let directory = '.';
  const rest: string[] = [];
  for (let at = 0; at < args.length; at++) {
    const argument = args[at]!;
    if (argument === '--package') {
      directory = args[++at] ?? '.';
    } else if (argument !== '--json') {
      rest.push(argument);
    }
  }
  const [command, query, file] = rest;
  if (command !== 'query' || query !== 'hovers' || !file) {
    console.error('usage: narrac query [--package DIR] --json hovers FILE');
    return 2;
  }
  const lines = readFileSync(path.join(directory, file), 'utf8').split('\n');
  const diagnostics = [];
  const expected = /^\/\/ error: (\d+):(\d+): (.*)$/u.exec(lines[0] ?? '');
  if (expected) {
    const line = Number(expected[1]);
    const column = Number(expected[2]);
    const tail = Array.from(lines[line - 1] ?? '')
      .slice(column - 1)
      .join('');
    const token = /^(?:[A-Za-z_]\w*|\d+(?:\.\d+)?|\S)/u.exec(tail);
    diagnostics.push({
      line,
      column,
      length: Array.from(token?.[0] ?? ' ').length,
      code: 'E0000',
      message: expected[3],
    });
  }
  const answer = { query: 'hovers', file, hovers: readHovers(lines), diagnostics };
  process.stdout.write(`${JSON.stringify(answer)}\n`);
  return 0;
}

process.exitCode = run(process.argv.slice(2));
