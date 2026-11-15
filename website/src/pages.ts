// The pages of the site. They are the Markdown files of the repository, read at build time:
// the chapters of the guide, the installation guide and the specification.
import guideIndex from '../../docs/guide/README.md';
import install from '../../docs/install.md';
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

export { guideIndex, install, spec };
