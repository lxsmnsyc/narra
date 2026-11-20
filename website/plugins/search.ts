// Builds `virtual:search-index`: one entry for each section of the guide, the installation guide,
// the reference pages and the specification, from the Markdown files of the repository. The index of the search
// library is built in the browser, from these entries, when search opens for the first time.
import { readdir, readFile } from 'node:fs/promises';
import MarkdownIt from 'markdown-it';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import type { Plugin } from 'vite';
import { makeSlug } from './markdown.ts';

export interface SearchEntry {
  id: string;
  /** The address of the entry, with the anchor of its section. */
  href: string;
  /** The title of the page. */
  page: string;
  heading: string;
  text: string;
  /** Whether the entry is a section, not the start of the page. */
  section: boolean;
}

const ROOT = fileURLToPath(new URL('../../', import.meta.url));
const VIRTUAL_ID = 'virtual:search-index';
const RESOLVED_ID = `\0${VIRTUAL_ID}`;

/** The text of a section stops here, so a long section does not outweigh the rest. */
const MAX_TEXT = 3000;

interface Source {
  file: string;
  url: string;
}

interface InlineToken {
  children: { type: string; content: string }[] | null;
}

async function listSources(): Promise<Source[]> {
  const sources: Source[] = [{ file: 'docs/guide/README.md', url: '/guide' }];
  const names = await readdir(path.join(ROOT, 'docs/guide'));
  names.sort();
  for (const name of names) {
    if (/^\d.*\.md$/.test(name)) {
      sources.push({ file: `docs/guide/${name}`, url: `/guide/${name.slice(0, -3)}` });
    }
  }
  sources.push({ file: 'docs/install.md', url: '/install' });
  sources.push({ file: 'docs/reference/README.md', url: '/reference' });
  // The reference pages are the other Markdown files of the folder, at any depth.
  const pages = await readdir(path.join(ROOT, 'docs/reference'), { recursive: true });
  pages.sort();
  for (const name of pages.map((entry) => entry.split(path.sep).join('/'))) {
    if (name.endsWith('.md') && name !== 'README.md') {
      sources.push({ file: `docs/reference/${name}`, url: `/reference/${name.slice(0, -3)}` });
    }
  }
  sources.push({ file: 'SPEC.md', url: '/spec' });
  return sources;
}

function readInline(token: InlineToken): string {
  // The text of a line without its markup.
  let text = '';
  for (const child of token.children ?? []) {
    if (child.type === 'text' || child.type === 'code_inline') {
      text += child.content;
    } else if (child.type === 'softbreak') {
      text += ' ';
    }
  }
  return text;
}

function createEntries(source: string, url: string): SearchEntry[] {
  const parser = new MarkdownIt({ html: true });
  const tokens = parser.parse(source, {});
  const entries: SearchEntry[] = [];
  const counts = new Map<string, number>();
  let title = '';
  let current: SearchEntry | undefined;

  for (const [index, token] of tokens.entries()) {
    if (token.type === 'heading_open') {
      const level = Number(token.tag.slice(1));
      const heading = readInline(tokens[index + 1]!);
      // The anchor is the one that the page gives the heading, with a number after a repeat.
      const base = makeSlug(heading);
      const seen = counts.get(base) ?? 0;
      counts.set(base, seen + 1);
      const anchor = seen === 0 ? base : `${base}-${seen}`;
      if (level === 1) {
        title = heading;
        current = { id: url, href: url, page: title, heading: title, text: '', section: false };
        entries.push(current);
      } else if (level <= 3) {
        current = {
          id: `${url}#${anchor}`,
          href: `${url}#${anchor}`,
          page: title,
          heading,
          text: '',
          section: true,
        };
        entries.push(current);
      } else if (current) {
        // A deeper heading belongs to its section, and its title is searchable text.
        current.text += `${heading}. `;
      }
      continue;
    }
    if (!current) {
      continue;
    }
    if (token.type === 'inline') {
      current.text += `${readInline(token)} `;
    } else if (token.type === 'fence') {
      current.text += `${token.content} `;
    }
  }

  const kept: SearchEntry[] = [];
  for (const entry of entries) {
    entry.text = entry.text.replace(/\s+/g, ' ').trim().slice(0, MAX_TEXT);
    // A page title with no text under it before its first section adds no result.
    if (entry.section || entry.text !== '') {
      kept.push(entry);
    }
  }
  return kept;
}

async function createIndex(): Promise<SearchEntry[]> {
  const entries: SearchEntry[] = [];
  for (const { file, url } of await listSources()) {
    // oxlint-disable-next-line no-await-in-loop -- The pages are few, and the order is the reading order.
    const source = await readFile(path.join(ROOT, file), 'utf8');
    entries.push(...createEntries(source, url));
  }
  return entries;
}

export default function search(): Plugin {
  return {
    name: 'narra-search',
    resolveId(id) {
      return id === VIRTUAL_ID ? RESOLVED_ID : undefined;
    },
    async load(id) {
      if (id !== RESOLVED_ID) {
        return undefined;
      }
      for (const { file } of await listSources()) {
        this.addWatchFile(path.join(ROOT, file));
      }
      return `export default ${JSON.stringify(await createIndex())};`;
    },
  };
}
