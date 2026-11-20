// Turns the Markdown files of the repository into modules that export their title, summary,
// headings and HTML. Code is highlighted with twinkleplop at build time, so a page loads no
// highlighting script. The HTML is a `raw` value, so markup inserts it as it is.
import { language as json } from '@twinkleplop/json';
import twinkleplop from '@twinkleplop/markdown-it';
import { language as typescript } from '@twinkleplop/typescript';
import MarkdownIt from 'markdown-it';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import type { Plugin } from 'vite';
import { readFenceLanguage } from '../src/highlight/fences.ts';
import { language as lua } from '../src/highlight/lua.ts';
import { language as narra } from '../src/highlight/narra.ts';
import { language as sh } from '../src/highlight/sh.ts';
import { language as toml } from '../src/highlight/toml.ts';
import { annotate, collectSamples, queryHovers, readFenceKind } from './twoslash.ts';
import type { Answer } from './twoslash.ts';

const REPOSITORY = 'https://github.com/lxsmnsyc/narra';

/** The root of the repository, one folder above `website`. */
const ROOT = fileURLToPath(new URL('../../', import.meta.url));

const RAW_MODULE = fileURLToPath(import.meta.resolve('brumate/html'));

export interface Heading {
  level: number;
  text: string;
  id: string;
}

export interface Page {
  title: string;
  /** The first paragraph as plain text, for the description of the page. */
  description: string;
  headings: Heading[];
  html: unknown;
}

export const LANGUAGES = {
  narra: narra(),
  sh: sh(),
  bash: sh(),
  // The PowerShell samples are one-line commands with a pipe, which the shell grammar reads right.
  powershell: sh(),
  json: json(),
  ts: typescript(),
  toml: toml(),
  lua: lua(),
};

function rewriteLink(href: string, from: string): string {
  // `from` is the repository path of the page that holds the link. A document with a page on the
  // site links there, and any other file of the repository links to GitHub.
  const match = /^([^#?]*)(.*)$/.exec(href)!;
  const [, file, rest] = match;
  if (!file || /^[a-z][a-z0-9+.-]*:/i.test(file)) {
    return href;
  }
  const target = path.posix.normalize(path.posix.join(path.posix.dirname(from), file));
  const guide = /^docs\/guide\/(.+)\.md$/.exec(target);
  if (guide) {
    return `/guide${guide[1] === 'README' ? '' : `/${guide[1]}`}${rest}`;
  }
  const reference = /^docs\/reference\/(.+)\.md$/.exec(target);
  if (reference) {
    return `/reference${reference[1] === 'README' ? '' : `/${reference[1]}`}${rest}`;
  }
  if (target === 'docs/install.md') {
    return `/install${rest}`;
  }
  if (target === 'SPEC.md') {
    return `/spec${rest}`;
  }
  return `${REPOSITORY}/blob/main/${target}${rest}`;
}

export function makeSlug(text: string): string {
  // As GitHub makes it, so the links inside the documents keep working.
  return text
    .toLowerCase()
    .replace(/[^\p{L}\p{N}\s_-]/gu, '')
    .trim()
    .replace(/\s/g, '-');
}

function createRenderer(from: string): MarkdownIt {
  const parser = new MarkdownIt({ html: true, linkify: false });
  parser.use(twinkleplop, { languages: LANGUAGES, on_unknown_language: 'plain' });

  // The comments of SPEC.md that hold test context are not text for readers.
  parser.renderer.rules.html_block = (tokens, index) => {
    const content = tokens[index]!.content;
    return /^\s*<!--/.test(content) ? '' : content;
  };

  const renderLink = parser.renderer.rules.link_open;
  parser.renderer.rules.link_open = (tokens, index, options, env, self) => {
    const token = tokens[index]!;
    const href = token.attrGet('href');
    if (href) {
      const rewritten = rewriteLink(href, from);
      token.attrSet('href', rewritten);
      if (/^https?:/.test(rewritten)) {
        token.attrSet('rel', 'noreferrer');
      }
    }
    return renderLink
      ? renderLink(tokens, index, options, env, self)
      : self.renderToken(tokens, index, options);
  };

  // A code block gets a label with its language and a name for the keyboard, since a wide block
  // scrolls and a scrolling region must take focus. A table scrolls the same way.
  const renderFence = parser.renderer.rules.fence!;
  parser.renderer.rules.fence = (tokens, index, options, env, self) => {
    // The language comes from the info string, so `narra file shop narra.toml` is TOML.
    const token = tokens[index]!;
    const kind = readFenceKind(token.info);
    const language = readFenceLanguage(token.info);
    token.info = language;
    let block = renderFence(tokens, index, options, env, self).replace(
      /<pre(?=[\s>])/,
      '<pre tabindex="0" role="group" aria-label="Code"',
    );
    // The answers of narrac were collected before the render; a sample without one stays plain.
    const answer = kind ? (env as { answers?: Map<number, Answer> }).answers?.get(index) : undefined;
    if (answer) {
      block = annotate(block, answer, kind === 'error');
    }
    const label =
      language && language !== 'text'
        ? `<span class="code-lang">${parser.utils.escapeHtml(language)}</span>`
        : '';
    return `<div class="code-block">${label}${block}</div>\n`;
  };
  parser.renderer.rules.table_open = () =>
    '<div class="table-scroll" role="group" aria-label="Table" tabindex="0">\n<table>\n';
  parser.renderer.rules.table_close = () => '</table>\n</div>\n';

  // Headings get the id GitHub gives them, with a number after a repeated title.
  parser.core.ruler.push('heading-ids', (state) => {
    const counts = new Map<string, number>();
    const headings: Heading[] = [];
    for (const [index, token] of state.tokens.entries()) {
      if (token.type !== 'heading_open') {
        continue;
      }
      // The text of the heading without its markup.
      let text = '';
      for (const child of state.tokens[index + 1]?.children ?? []) {
        if (child.type === 'text' || child.type === 'code_inline') {
          text += child.content;
        }
      }
      const base = makeSlug(text);
      const seen = counts.get(base) ?? 0;
      counts.set(base, seen + 1);
      const id = seen === 0 ? base : `${base}-${seen}`;
      token.attrSet('id', id);
      headings.push({ level: Number(token.tag.slice(1)), text, id });
    }
    (state.env as { headings: Heading[] }).headings = headings;
  });
  return parser;
}

function describe(parser: MarkdownIt, source: string): string {
  // The first paragraph, as plain text.
  const tokens = parser.parse(source, {});
  for (const [index, token] of tokens.entries()) {
    if (token.type === 'paragraph_open') {
      let text = '';
      for (const child of tokens[index + 1]?.children ?? []) {
        text += child.type === 'softbreak' ? ' ' : child.content;
      }
      return text.replace(/\s+/g, ' ').trim();
    }
  }
  return '';
}

export default function markdown(): Plugin {
  return {
    name: 'narra-markdown',
    enforce: 'pre',
    async transform(source, id) {
      const file = id.split('?')[0]!;
      if (!file.endsWith('.md')) {
        return undefined;
      }
      this.addWatchFile(file);
      const from = path.relative(ROOT, file).split(path.sep).join('/');
      const parser = createRenderer(from);
      const env: { headings: Heading[]; answers?: Map<number, Answer> } = { headings: [] };
      // The samples go to narrac together before the page renders, keyed by their place in the
      // token list, which the render sees in the same order.
      const fences: { index: number; info: string; content: string }[] = [];
      for (const [index, token] of parser.parse(source, {}).entries()) {
        if (token.type === 'fence') {
          fences.push({ index, info: token.info, content: token.content });
        }
      }
      env.answers = await queryHovers(collectSamples(fences));
      const html = parser.render(source, env);
      let title = '';
      for (const heading of env.headings) {
        if (heading.level === 1) {
          title = heading.text;
          break;
        }
      }
      const page = {
        title,
        description: describe(parser, source),
        headings: env.headings,
      };
      return {
        code: [
          `import { raw } from ${JSON.stringify(RAW_MODULE)};`,
          `export default { ...${JSON.stringify(page)}, html: raw(${JSON.stringify(html)}) };`,
        ].join('\n'),
        map: null,
      };
    },
  };
}
