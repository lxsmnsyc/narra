// The link check of the build. It reads the HTML that the build wrote and checks each link of the
// package pages: a link to the site names a page that exists, and a link with an anchor names an
// id that the page has. Links to other sites are not followed.
import { existsSync, readdirSync, readFileSync, statSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const output = fileURLToPath(new URL('../dist/client', import.meta.url));

function listPages(folder: string): string[] {
  const pages: string[] = [];
  for (const name of readdirSync(folder)) {
    const file = path.join(folder, name);
    if (statSync(file).isDirectory()) {
      pages.push(...listPages(file));
    } else if (name.endsWith('.html')) {
      pages.push(file);
    }
  }
  return pages;
}

function findPageFile(pathname: string): string | undefined {
  // The site serves clean URLs: `/packages/http` is `packages/http/index.html` or `packages/http.html`.
  const base = path.join(output, pathname);
  for (const candidate of [path.join(base, 'index.html'), `${base}.html`]) {
    if (existsSync(candidate) && statSync(candidate).isFile()) {
      return candidate;
    }
  }
  return undefined;
}

function readIds(file: string, cache: Map<string, Set<string>>): Set<string> {
  let ids = cache.get(file);
  if (!ids) {
    ids = new Set<string>();
    for (const match of readFileSync(file, 'utf8').matchAll(/\sid="([^"]+)"/g)) {
      ids.add(match[1]!);
    }
    cache.set(file, ids);
  }
  return ids;
}

function checkPage(file: string, cache: Map<string, Set<string>>): string[] {
  const problems: string[] = [];
  const html = readFileSync(file, 'utf8');
  const name = path.relative(output, file);
  for (const match of html.matchAll(/<a\s[^>]*?href="([^"]*)"/g)) {
    const href = match[1]!.replaceAll('&amp;', '&');
    if (/^[a-z][a-z0-9+.-]*:/i.test(href)) {
      continue;
    }
    const [target = '', anchor] = href.split('#') as [string, string | undefined];
    const found = target === '' ? file : findPageFile(target.split('?')[0]!);
    if (!found) {
      problems.push(`${name}: ${href}: no such page`);
    } else if (anchor && !readIds(found, cache).has(anchor)) {
      problems.push(`${name}: ${href}: no such anchor`);
    }
  }
  return problems;
}

const cache = new Map<string, Set<string>>();
const problems: string[] = [];
let count = 0;
for (const file of listPages(output)) {
  const relative = path.relative(output, file);
  // The package pages are the index and one page for each README.
  if (relative.startsWith('packages')) {
    count += 1;
    problems.push(...checkPage(file, cache));
  }
}
for (const problem of problems) {
  console.error(problem);
}
console.log(`links: ${count} package pages, ${problems.length} problems`);
if (problems.length > 0) {
  process.exitCode = 1;
}
