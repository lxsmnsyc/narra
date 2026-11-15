// The TextMate grammar of Narra, read as token classes of twinkleplop. The check of the build
// compares these classes with the ones of `narra.ts`, character by character.
import { readFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';

const require = createRequire(import.meta.url);
const textmate = require('vscode-textmate') as typeof import('vscode-textmate');
const oniguruma = require('vscode-oniguruma') as typeof import('vscode-oniguruma');

const GRAMMAR = fileURLToPath(
  new URL('../../../editors/textmate/narra.tmLanguage.json', import.meta.url),
);

/** The class that a TextMate scope stands for, or undefined when the scope says nothing. */
export function classifyScope(scope: string): string | undefined {
  if (scope.startsWith('comment.')) {
    return 'comment';
  }
  if (scope.startsWith('constant.character.escape')) {
    return 'string_escape';
  }
  if (scope.startsWith('string.template')) {
    return 'template';
  }
  if (scope.startsWith('punctuation.definition.string.template')) {
    return 'template';
  }
  if (scope.startsWith('string.') || scope.startsWith('punctuation.definition.string')) {
    return 'string';
  }
  if (scope.startsWith('constant.numeric')) {
    return 'number';
  }
  if (scope.startsWith('constant.language.boolean')) {
    return 'boolean';
  }
  if (scope.startsWith('constant.language.none') || scope.startsWith('variable.other.constant')) {
    return 'constant';
  }
  if (scope.startsWith('support.type.primitive')) {
    return 'type';
  }
  if (scope.startsWith('entity.name.function.decorator')) {
    return 'decorator';
  }
  if (scope.startsWith('punctuation.definition.decorator')) {
    return 'decorator';
  }
  if (scope.startsWith('entity.name.function')) {
    return 'function';
  }
  if (scope.startsWith('entity.name.type') || scope.startsWith('entity.other')) {
    return 'class_name';
  }
  if (scope.startsWith('support.type.effect')) {
    return 'class_name';
  }
  if (scope.startsWith('keyword.operator.type.annotation')) {
    return 'punctuation';
  }
  if (scope.startsWith('keyword.operator')) {
    return 'operator';
  }
  if (
    scope.startsWith('keyword.') ||
    scope.startsWith('storage.')
  ) {
    return 'keyword';
  }
  if (scope.startsWith('punctuation.')) {
    return 'punctuation';
  }
  if (scope.startsWith('variable.other')) {
    return 'identifier';
  }
  return undefined;
}

function classifyStack(scopes: string[]): string {
  // The innermost scope that stands for a class wins, so `${ }` content keeps its own classes.
  // The braces and the dollar sign of a substitution are punctuation inside the template.
  for (let index = scopes.length - 1; index >= 0; index--) {
    const scope = scopes[index]!;
    if (scope.startsWith('punctuation.definition.template-expression')) {
      return 'punctuation';
    }
    const name = classifyScope(scope);
    if (name) {
      return name;
    }
  }
  return '';
}

let loaded: Promise<import('vscode-textmate').IGrammar> | undefined;

function loadGrammar(): Promise<import('vscode-textmate').IGrammar> {
  loaded ??= (async () => {
    const wasm = readFileSync(require.resolve('vscode-oniguruma/release/onig.wasm'));
    await oniguruma.loadWASM(wasm.buffer as ArrayBuffer);
    const registry = new textmate.Registry({
      onigLib: Promise.resolve({
        createOnigScanner: (patterns: string[]) => new oniguruma.OnigScanner(patterns),
        createOnigString: (text: string) => new oniguruma.OnigString(text),
      }),
      loadGrammar: async (scopeName: string) =>
        scopeName === 'source.narra'
          ? textmate.parseRawGrammar(readFileSync(GRAMMAR, 'utf8'), GRAMMAR)
          : null,
    });
    const grammar = await registry.loadGrammar('source.narra');
    if (!grammar) {
      throw new Error('the TextMate grammar of Narra did not load');
    }
    return grammar;
  })();
  return loaded;
}

/**
 * The class of every character of `code` by the TextMate grammar, in the order of the code. A
 * newline and the text between tokens that no scope names have the class `''`.
 */
export async function classifyTextMate(code: string): Promise<string[]> {
  const grammar = await loadGrammar();
  const classes: string[] = [];
  let stack = textmate.INITIAL;
  const lines = code.split('\n');
  for (const [number, line] of lines.entries()) {
    const result = grammar.tokenizeLine(line, stack);
    // One class per character of the line; the tokens cover it, clipped at its end.
    const row: string[] = new Array<string>(line.length).fill('');
    for (const item of result.tokens) {
      const name = classifyStack(item.scopes);
      const end = Math.min(item.endIndex, line.length);
      for (let index = item.startIndex; index < end; index++) {
        row[index] = name;
      }
    }
    classes.push(...row);
    if (number < lines.length - 1) {
      classes.push('');
    }
    stack = result.ruleStack;
  }
  return classes;
}
