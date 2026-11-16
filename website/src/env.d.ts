/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** The address of the site without a trailing slash, such as `https://narra.dev`. */
  readonly SITE_URL: string;
}

declare module '*.md' {
  import type { Page } from '../plugins/markdown.ts';

  const page: Page;
  export default page;
}

declare module 'virtual:search-index' {
  import type { SearchEntry } from '../plugins/search.ts';

  const entries: SearchEntry[];
  export default entries;
}
