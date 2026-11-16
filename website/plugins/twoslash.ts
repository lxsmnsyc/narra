// Twoslash for Narra: hover text and compiler errors in the code samples, computed at build time.
// `narrac query --json hovers` answers once per sample (the contract query-hovers); this module
// runs it for the samples of a page, caches the answers, and writes them into the highlighted HTML.
import { execFile } from 'node:child_process';
import { createHash } from 'node:crypto';
import { existsSync, statSync } from 'node:fs';
import { mkdir, mkdtemp, readFile, rm, writeFile } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { promisify } from 'node:util';

const WEBSITE = fileURLToPath(new URL('../', import.meta.url));
const CACHE = path.join(WEBSITE, 'node_modules/.cache/narra-twoslash');
const execFileAsync = promisify(execFile);
const POOL_SIZE = 4;
const QUERY_TIMEOUT = 20_000;

export interface Hover {
  line: number;
  column: number;
  length: number;
  text: string;
}

export interface Diagnostic {
  line: number;
  column: number;
  length: number;
  code: string;
  message: string;
}

export interface Answer {
  hovers: Hover[];
  diagnostics: Diagnostic[];
}

export type FenceKind = 'file' | 'error' | 'package';

/**
 * A sample is a whole file when its info string is `narra` or `narra error`, and one file of a
 * package of the page when it is `narra file <package> <path>` with a `.nra` path.
 */
export function readFenceKind(info: string): FenceKind | undefined {
  const words = info.trim().split(/\s+/);
  if (words[0] !== 'narra') {
    return undefined;
  }
  if (words.length === 1) {
    return 'file';
  }
  if (words[1] === 'file' && words.length === 4 && words[3]!.endsWith('.nra')) {
    return 'package';
  }
  return words.length === 2 && words[1] === 'error' ? 'error' : undefined;
}

/** The files of a package to check, and the one whose hovers the page shows. */
export interface Sample {
  files: Record<string, string>;
  target: string;
}

const DEFAULT_MANIFEST = '[package]\nname = "sample"\n';

/**
 * Turns the fences of a page into samples, by their place in the token list. A `narra file`
 * fence is a version of one file of a package: it is checked with the files of that package
 * shown before it, and with the first version of the files shown only after it, so an import of
 * a module the page shows later still resolves.
 */
export function collectSamples(fences: { index: number; info: string; content: string }[]): Map<number, Sample> {
  const samples = new Map<number, Sample>();
  const firstVersions = new Map<string, Record<string, string>>();
  for (const fence of fences) {
    const words = fence.info.trim().split(/\s+/);
    if (words[0] === 'narra' && words[1] === 'file' && words.length === 4) {
      const files = firstVersions.get(words[2]!) ?? {};
      files[words[3]!] ??= fence.content;
      firstVersions.set(words[2]!, files);
    }
  }
  const current = new Map<string, Record<string, string>>();
  for (const fence of fences) {
    const kind = readFenceKind(fence.info);
    const words = fence.info.trim().split(/\s+/);
    if (words[0] === 'narra' && words[1] === 'file' && words.length === 4) {
      // Every file fence updates its package, also `narra.toml`.
      const files = { ...current.get(words[2]!), [words[3]!]: fence.content };
      current.set(words[2]!, files);
      if (kind === 'package') {
        const all = { ...firstVersions.get(words[2]!), ...files };
        all['narra.toml'] ??= DEFAULT_MANIFEST;
        samples.set(fence.index, { files: all, target: words[3]! });
      }
    } else if (kind) {
      samples.set(fence.index, { files: { 'narra.toml': DEFAULT_MANIFEST, 'src/main.nra': fence.content }, target: 'src/main.nra' });
    }
  }
  return samples;
}

interface Narrac {
  command: string;
  /** Changes when the compiler changes, so the cache follows it. */
  identity: string;
}

let narracPromise: Promise<Narrac | undefined> | undefined;

function findOnPath(name: string): string | undefined {
  for (const directory of (process.env.PATH ?? '').split(path.delimiter)) {
    const candidate = path.join(directory, name);
    if (directory && existsSync(candidate)) {
      return candidate;
    }
  }
  return undefined;
}

function findNarrac(): Narrac | undefined {
  // `NARRAC` names the compiler; without it, `narrac` on the PATH is used.
  const named = process.env.NARRAC;
  const command = named ? path.resolve(named) : findOnPath('narrac');
  if (!command || !existsSync(command)) {
    console.log(
      named
        ? `twoslash: NARRAC=${named} does not exist, code samples have no hovers`
        : 'twoslash: no narrac (set NARRAC), code samples have no hovers',
    );
    return undefined;
  }
  const stats = statSync(command);
  return { command, identity: `${command}:${stats.size}:${stats.mtimeMs}` };
}

function getNarrac(): Promise<Narrac | undefined> {
  narracPromise ??= Promise.resolve(findNarrac());
  return narracPromise;
}

async function runQuery(narrac: Narrac, sample: Sample): Promise<Answer | undefined> {
  // The sample's files become a package of their own in a temporary folder.
  const directory = await mkdtemp(path.join(os.tmpdir(), 'narra-twoslash-'));
  try {
    for (const [file, content] of Object.entries(sample.files)) {
      const target = path.join(directory, file);
      if (!target.startsWith(directory + path.sep)) {
        continue;
      }
      await mkdir(path.dirname(target), { recursive: true });
      await writeFile(target, content);
    }
    const { stdout } = await execFileAsync(
      narrac.command,
      ['query', '--package', directory, '--json', 'hovers', sample.target],
      { timeout: QUERY_TIMEOUT, maxBuffer: 8 * 1024 * 1024 },
    );
    return readAnswer(stdout);
  } catch {
    return undefined;
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
}

function readAnswer(stdout: string): Answer | undefined {
  try {
    const parsed = JSON.parse(stdout.trim().split('\n').pop() ?? '') as Partial<Answer>;
    if (Array.isArray(parsed.hovers) && Array.isArray(parsed.diagnostics)) {
      return { hovers: parsed.hovers, diagnostics: parsed.diagnostics };
    }
  } catch {
    // An answer that is not JSON counts as no answer.
  }
  return undefined;
}

const answers = new Map<string, Promise<Answer | undefined>>();

function makeKey(sample: Sample): string {
  // The file names are sorted, so the same package gives the same key.
  const names = Object.keys(sample.files).sort();
  const parts: string[] = [sample.target];
  for (const name of names) {
    parts.push(name, sample.files[name]!);
  }
  return JSON.stringify(parts);
}

async function answerSample(narrac: Narrac, sample: Sample, sampleKey: string): Promise<Answer | undefined> {
  const key = createHash('sha256').update(narrac.identity).update('\0').update(sampleKey).digest('hex');
  const file = path.join(CACHE, `${key}.json`);
  try {
    return JSON.parse(await readFile(file, 'utf8')) as Answer;
  } catch {
    // Not cached yet.
  }
  const answer = await runQuery(narrac, sample);
  if (answer) {
    await mkdir(CACHE, { recursive: true });
    await writeFile(file, JSON.stringify(answer));
  }
  return answer;
}

/**
 * Asks narrac about each sample, four processes at a time. A sample the compiler could not answer
 * is missing from the result and renders as plain highlighted code.
 */
export async function queryHovers(samples: Map<number, Sample>): Promise<Map<number, Answer>> {
  const result = new Map<number, Answer>();
  const narrac = await getNarrac();
  if (!narrac) {
    return result;
  }
  const pending: Promise<void>[] = [];
  const queue: [string, Sample, (answer: Answer | undefined) => void][] = [];
  const keys = new Map<number, string>();
  for (const [index, sample] of samples) {
    const key = makeKey(sample);
    keys.set(index, key);
    if (!answers.has(key)) {
      // The real promise is stored before queuing, so a concurrent call for the same sample waits for it.
      let settle: (answer: Answer | undefined) => void = () => {};
      answers.set(
        key,
        new Promise<Answer | undefined>((resolve) => {
          settle = resolve;
        }),
      );
      queue.push([key, sample, settle]);
    }
  }
  async function work(): Promise<void> {
    for (let item = queue.shift(); item !== undefined; item = queue.shift()) {
      let answer: Answer | undefined;
      try {
        answer = await answerSample(narrac!, item[1], item[0]);
      } catch {
        answer = undefined;
      }
      item[2](answer);
    }
  }
  for (let worker = 0; worker < Math.min(POOL_SIZE, queue.length); worker++) {
    pending.push(work());
  }
  await Promise.all(pending);
  let failed = 0;
  for (const [index, key] of keys) {
    const answer = await answers.get(key);
    if (answer) {
      result.set(index, answer);
    } else {
      failed++;
    }
  }
  if (failed > 0) {
    console.log(`twoslash: narrac gave no answer for ${failed} sample(s), they have no hovers`);
  }
  return result;
}

const ENTITIES: Record<string, string> = {
  '&lt;': '<',
  '&gt;': '>',
  '&amp;': '&',
  '&quot;': '"',
  '&#39;': "'",
};

function decodeText(html: string): string {
  return html.replace(/&(?:lt|gt|amp|quot|#39);/g, (entity) => ENTITIES[entity] ?? entity);
}

function escapeText(text: string): string {
  return text.replace(/[&<>"']/g, (character) => {
    switch (character) {
      case '&':
        return '&amp;';
      case '<':
        return '&lt;';
      case '>':
        return '&gt;';
      case '"':
        return '&quot;';
      default:
        return '&#39;';
    }
  });
}

interface Piece {
  /** The class attribute of a token, or undefined for text outside tokens. */
  classes: string | undefined;
  characters: string[];
}

function readPieces(inner: string): Piece[] | undefined {
  // A line holds tokens `<span class="tok X">text</span>` and plain text between them.
  const pieces: Piece[] = [];
  const pattern = /<span class="([^"]*)">([^<]*)<\/span>|([^<]+)/gy;
  let end = 0;
  for (let match = pattern.exec(inner); match; match = pattern.exec(inner)) {
    end = pattern.lastIndex;
    pieces.push({
      classes: match[3] === undefined ? match[1] : undefined,
      characters: Array.from(decodeText(match[3] ?? match[2] ?? '')),
    });
  }
  return end === inner.length ? pieces : undefined;
}

interface Mark {
  /** Offsets in code points from the start of the line, the end excluded. */
  start: number;
  end: number;
}

interface LineMarks {
  hovers: (Mark & { text: string })[];
  errors: Mark[];
}

function writeLine(inner: string, marks: LineMarks): string {
  const pieces = readPieces(inner);
  if (!pieces) {
    return inner;
  }
  let output = '';
  let offset = 0;
  for (const piece of pieces) {
    // Cut the piece where a mark starts or ends, and write each part with the marks around it.
    const cuts = new Set([0, piece.characters.length]);
    for (const mark of [...marks.hovers, ...marks.errors]) {
      for (const point of [mark.start - offset, mark.end - offset]) {
        if (point > 0 && point < piece.characters.length) {
          cuts.add(point);
        }
      }
    }
    const points = Int32Array.from(cuts).sort();
    for (let index = 0; index + 1 < points.length; index++) {
      const from = points[index]!;
      const to = points[index + 1]!;
      const text = escapeText(piece.characters.slice(from, to).join(''));
      let html = piece.classes === undefined ? text : `<span class="${piece.classes}">${text}</span>`;
      const start = offset + from;
      let hover: string | undefined;
      for (const mark of marks.hovers) {
        if (mark.start <= start && start < mark.end) {
          hover = mark.text;
        }
      }
      let wrong = false;
      for (const mark of marks.errors) {
        wrong ||= mark.start <= start && start < mark.end;
      }
      if (hover !== undefined || wrong) {
        const tip = hover === undefined ? '' : ` data-tip="${escapeText(hover)}"`;
        html = `<span class="${[hover === undefined ? '' : 'tw', wrong ? 'tw-err' : ''].join(' ').trim()}"${tip}>${html}</span>`;
      }
      output += html;
    }
    offset += piece.characters.length;
  }
  return output;
}

/**
 * Writes the answer of narrac into the highlighted HTML of a sample: a `tw` element around each
 * hovered token, an `tw-err` underline under each error, and an error line after the line it
 * belongs to. The text of the code stays the same, so the copy button copies what the sample says.
 */
export function annotate(html: string, answer: Answer, showErrors: boolean): string {
  const open = html.indexOf('<code>');
  const close = html.lastIndexOf('</code>');
  if (open === -1 || close === -1) {
    return html;
  }
  const lines = html.slice(open + 6, close).split('\n');
  const marks: LineMarks[] = [];
  function getMarks(line: number): LineMarks {
    marks[line] ??= { hovers: [], errors: [] };
    return marks[line]!;
  }
  for (const hover of answer.hovers) {
    getMarks(hover.line).hovers.push({
      start: hover.column - 1,
      end: hover.column - 1 + hover.length,
      text: hover.text,
    });
  }
  const messages: string[][] = [];
  if (showErrors) {
    for (const diagnostic of answer.diagnostics) {
      getMarks(diagnostic.line).errors.push({
        start: diagnostic.column - 1,
        end: diagnostic.column - 1 + diagnostic.length,
      });
      messages[diagnostic.line] ??= [];
      messages[diagnostic.line]!.push(diagnostic.message);
    }
  }
  let output = '';
  for (const [index, line] of lines.entries()) {
    const number = index + 1;
    const wrapper = /^(<span class="l">)([\s\S]*)(<\/span>)$/.exec(line);
    const written = wrapper && marks[number] ? wrapper[1] + writeLine(wrapper[2]!, marks[number]!) + wrapper[3] : line;
    const last = index === lines.length - 1;
    output += written + (last ? '' : '\n');
    // The error line comes after the line break, so the text of the lines keeps its breaks.
    for (const message of messages[number] ?? []) {
      output += `<span class="tw-diag"><span class="tw-sr">Error: </span>${escapeText(message)}</span>`;
    }
  }
  return html.slice(0, open + 6) + output + html.slice(close);
}
