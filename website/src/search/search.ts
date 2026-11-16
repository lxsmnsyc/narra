// The search of the docs. The entries and the search library load the first time a query runs, so
// a page that never opens search never loads them.
import type MiniSearch from 'minisearch';
import type { SearchEntry } from '../../plugins/search.ts';

export interface SearchResult {
  id: string;
  href: string;
  page: string;
  heading: string;
  snippet: string;
  section: boolean;
}

const MAX_RESULTS = 10;
const SNIPPET_LENGTH = 120;

let loading: Promise<MiniSearch<SearchEntry>> | undefined;

export function loadIndex(): Promise<MiniSearch<SearchEntry>> {
  loading ??= Promise.all([import('minisearch'), import('virtual:search-index')]).then(
    ([{ default: MiniSearchClass }, { default: entries }]) => {
      const index = new MiniSearchClass<SearchEntry>({
        fields: ['heading', 'page', 'text'],
        storeFields: ['href', 'page', 'heading', 'text', 'section'],
        searchOptions: {
          boost: { heading: 3, page: 1.5 },
          prefix: true,
          fuzzy: 0.15,
          combineWith: 'AND',
        },
      });
      index.addAll(entries);
      return index;
    },
  );
  return loading;
}

function createSnippet(text: string, terms: string[], heading: string): string {
  // The text is cut by code point, so a surrogate pair stays whole.
  const points = Array.from(text);
  const lower = text.toLowerCase();
  let at = -1;
  for (const term of terms) {
    const found = lower.indexOf(term);
    if (found !== -1 && (at === -1 || found < at)) {
      at = found;
    }
  }
  // The text of a section starts with nothing the heading already says.
  let first = 0;
  if (heading !== '' && text.startsWith(`${heading}. `)) {
    first = Array.from(`${heading}. `).length;
  }
  // The index of a UTF-16 position in code points.
  const atPoint = at === -1 ? 0 : Array.from(text.slice(0, at)).length;
  let start = Math.max(first, atPoint - 30);
  // The start and the end move to the nearest space, so no word is cut.
  if (start > first) {
    while (start < points.length && points[start - 1] !== ' ') {
      start++;
    }
  }
  let end = Math.min(points.length, start + SNIPPET_LENGTH);
  if (end < points.length) {
    while (end > start && points[end] !== ' ') {
      end--;
    }
    if (end === start) {
      end = Math.min(points.length, start + SNIPPET_LENGTH);
    }
  }
  return (start > first ? '…' : '') + points.slice(start, end).join('').trim() + (end < points.length ? '…' : '');
}

export async function runSearch(query: string): Promise<SearchResult[]> {
  const trimmed = query.trim();
  if (!trimmed) {
    return [];
  }
  const index = await loadIndex();
  const terms = trimmed.toLowerCase().split(/\s+/);
  const results: SearchResult[] = [];
  for (const hit of index.search(trimmed)) {
    if (results.length === MAX_RESULTS) {
      break;
    }
    results.push({
      id: String(hit.id),
      href: String(hit.href),
      page: String(hit.page),
      heading: String(hit.heading),
      snippet: createSnippet(String(hit.text), terms, String(hit.heading)),
      section: Boolean(hit.section),
    });
  }
  return results;
}
