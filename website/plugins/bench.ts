// Builds `virtual:bench-baseline`: the latest run of `bench/BASELINE.md`, for the benchmark chart
// of the home page. The numbers come from the file at build time, so the chart never drifts from it.
import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import type { Plugin } from 'vite';

export interface BenchProgram {
  name: string;
  /** Median wall-clock seconds, in the order of `languages`. */
  times: number[];
}

export interface BenchRun {
  date: string;
  machine: string;
  runs: number;
  loadStart: string;
  loadEnd: string;
  languages: string[];
  programs: BenchProgram[];
}

const ROOT = fileURLToPath(new URL('../../', import.meta.url));
const FILE = 'bench/BASELINE.md';
const VIRTUAL_ID = 'virtual:bench-baseline';
const RESOLVED_ID = `\0${VIRTUAL_ID}`;

function readCells(line: string): string[] {
  const cells: string[] = [];
  for (const cell of line.trim().replace(/^\||\|$/g, '').split('|')) {
    cells.push(cell.trim());
  }
  return cells;
}

function matchOrFail(text: string, pattern: RegExp, what: string): RegExpExecArray {
  const match = pattern.exec(text);
  if (!match) {
    throw new Error(`${FILE}: the latest run has no ${what}; update plugins/bench.ts`);
  }
  return match;
}

export function readLatestRun(source: string): BenchRun {
  // The latest run is the first `##` section.
  const sections = source.split(/^## /m);
  const section = sections[1] ?? '';
  const date = matchOrFail(section, /^(\d{4}-\d{2}-\d{2})/, 'date')[1]!;

  // The run paragraph names the machine, the number of runs and the load average.
  const runs = Number(matchOrFail(section, /run\.sh -n (\d+)/, 'run count')[1]);
  const machine = matchOrFail(section, /machine (.+?)\. /, 'machine')[1]!.replace(/`/g, '');
  const load = matchOrFail(
    section,
    /load average was ([\d.]+) \(1 minute\) at the start and ([\d.]+) at the end/,
    'load average',
  );

  // The table of seconds: one column per language between `program` and `output`. A program
  // that only Narra runs, such as `compiler`, has `-` in the other columns and is left out.
  const lines = section.split('\n');
  let header = -1;
  for (const [index, line] of lines.entries()) {
    if (line.startsWith('| program')) {
      header = index;
      break;
    }
  }
  if (header === -1) {
    throw new Error(`${FILE}: the latest run has no results table`);
  }
  const columns = readCells(lines[header]!);
  const languages = columns.slice(1, columns.indexOf('output'));
  const programs: BenchProgram[] = [];
  for (const line of lines.slice(header + 2)) {
    if (!line.startsWith('|')) {
      break;
    }
    const cells = readCells(line);
    const times: number[] = [];
    for (const cell of cells.slice(1, 1 + languages.length)) {
      times.push(Number(cell));
    }
    let complete = true;
    for (const time of times) {
      complete &&= Number.isFinite(time) && time > 0;
    }
    if (complete) {
      programs.push({ name: cells[0]!, times });
    }
  }
  return { date, machine, runs, loadStart: load[1]!, loadEnd: load[2]!, languages, programs };
}

export default function bench(): Plugin {
  return {
    name: 'narra-bench',
    resolveId(id) {
      return id === VIRTUAL_ID ? RESOLVED_ID : undefined;
    },
    async load(id) {
      if (id !== RESOLVED_ID) {
        return undefined;
      }
      const file = path.join(ROOT, FILE);
      this.addWatchFile(file);
      return `export default ${JSON.stringify(readLatestRun(await readFile(file, 'utf8')))};`;
    },
  };
}
