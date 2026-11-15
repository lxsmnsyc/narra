/** A position as `[line, column]`, counted from 1. */
export type Position = [line: number, column: number];

export interface Edit {
  file: string;
  range: { start: Position; end: Position };
  text: string;
}

export interface Fix {
  title: string;
  edits: Edit[];
}

/** One diagnostic, as `narrac check --json` prints it. */
export interface Diagnostic {
  /** The code (`E0119`), or "" for a problem that is not a diagnostic, such as an unreadable file. */
  code: string;
  severity: "error";
  message: string;
  file: string;
  range: { start: Position; end: Position };
  fixes: Fix[];
}

export interface EmitResult {
  file: string;
  /** 0 is clean, 1 is a load or check error, 3 is C with unsupported markers. */
  status: number;
  /** The C program `narrac emit` printed. */
  c: string;
  stderr: string;
}

export interface RunOptions {
  input?: string;
  cwd?: string;
  env?: Record<string, string | undefined>;
}

export interface RunResult {
  status: number;
  stdout: string;
  stderr: string;
}

export interface Runner {
  kind: "native" | "wasm";
  path: string;
}

export class NarracError extends Error {
  status: number;
}

/** Checks each file and returns every diagnostic. An empty array means the files are clean. */
export function check(files: string | string[]): Diagnostic[];

/** Returns the source in narrac's layout. Throws NarracError when the source does not parse. */
export function format(source: string): string;

/** Returns the C of each file. */
export function emit(files: string | string[]): EmitResult[];

/** Runs narrac with any arguments. */
export function run(args: string[], options?: RunOptions): RunResult;

/** Names the compiler this process uses. */
export function describeRunner(): Runner;
