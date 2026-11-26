// The pages of the site. They are the Markdown files of the repository, read at build time:
// the chapters of the guide, the installation guide, the reference, the first-party packages and
// the specification.
import guideIndex from '../../docs/guide/README.md';
import install from '../../docs/install.md';
import packagesIndex from '../../docs/packages/README.md';
import referenceIndex from '../../docs/reference/README.md';
import spec from '../../SPEC.md';
import type { Page } from '../plugins/markdown.ts';

export type { Page };

const files = import.meta.glob<Page>('../../docs/guide/[0-9]*.md', {
  eager: true,
  import: 'default',
});

export interface Chapter {
  slug: string;
  title: string;
  page: Page;
}

function listChapters(): Chapter[] {
  const chapters: Chapter[] = [];
  for (const [file, page] of Object.entries(files)) {
    const slug = /([^/]+)\.md$/.exec(file)![1]!;
    chapters.push({ slug, title: page.title, page });
  }
  chapters.sort((first, second) => first.slug.localeCompare(second.slug));
  return chapters;
}

/** The chapters of the guide in reading order. */
export const chapters: Chapter[] = listChapters();

export function findChapter(slug: string): Chapter | undefined {
  for (const chapter of chapters) {
    if (chapter.slug === slug) {
      return chapter;
    }
  }
  return undefined;
}

const referenceFiles = import.meta.glob<Page>('../../docs/reference/**/*.md', {
  eager: true,
  import: 'default',
});

export interface ReferencePage {
  /** The path under `docs/reference` without `.md`, such as `commands` or `std/option`. */
  slug: string;
  title: string;
  page: Page;
}

function listReferencePages(): ReferencePage[] {
  const pages: ReferencePage[] = [];
  for (const [file, page] of Object.entries(referenceFiles)) {
    // The README of a folder is its index: `std/README.md` is `std`.
    const slug = /docs\/reference\/(.+)\.md$/.exec(file)![1]!.replace(/\/README$/, '');
    // The index of the section is `README.md`, which has its own route.
    if (slug !== 'README') {
      pages.push({ slug, title: page.title, page });
    }
  }
  pages.sort((first, second) => first.slug.localeCompare(second.slug));
  return pages;
}

/** The reference pages, one for each Markdown file under `docs/reference` except the index. */
export const referencePages: ReferencePage[] = listReferencePages();

export function findReference(slug: string): ReferencePage | undefined {
  for (const entry of referencePages) {
    if (entry.slug === slug) {
      return entry;
    }
  }
  return undefined;
}

const packageFiles = import.meta.glob<Page>('../../packages/*/README.md', {
  eager: true,
  import: 'default',
});

export interface PackagePage {
  /** The folder of the package under `packages`, such as `http`. */
  slug: string;
  title: string;
  /** The first paragraph of the README, for the index. */
  description: string;
  page: Page;
}

function listPackagePages(): PackagePage[] {
  const pages: PackagePage[] = [];
  for (const [file, page] of Object.entries(packageFiles)) {
    const slug = /packages\/([^/]+)\/README\.md$/.exec(file)![1]!;
    pages.push({ slug, title: page.title || slug, description: page.description, page });
  }
  pages.sort((first, second) => first.slug.localeCompare(second.slug));
  return pages;
}

/** The first-party packages, one for each `packages/<name>/README.md`. */
export const packagePages: PackagePage[] = listPackagePages();

export function findPackage(slug: string): PackagePage | undefined {
  for (const entry of packagePages) {
    if (entry.slug === slug) {
      return entry;
    }
  }
  return undefined;
}

export { guideIndex, install, packagesIndex, referenceIndex, spec };
