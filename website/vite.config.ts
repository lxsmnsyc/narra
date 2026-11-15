import { fileURLToPath } from 'node:url';
import brumate from 'brumate/vite';
import { defineConfig } from 'vite';
import markdown from './plugins/markdown.ts';

/**
 * The address of the site, for link previews, which need absolute URLs. `SITE_URL` sets it.
 * On Vercel it is the production domain of the project.
 */
const siteUrl = (
  process.env.SITE_URL ??
  (process.env.VERCEL_PROJECT_PRODUCTION_URL
    ? `https://${process.env.VERCEL_PROJECT_PRODUCTION_URL}`
    : 'http://localhost:5173')
).replace(/\/$/, '');

export default defineConfig({
  define: {
    'import.meta.env.SITE_URL': JSON.stringify(siteUrl),
  },
  plugins: [
    markdown(),
    // Every page is written as a static file.
    brumate({ prerender: { paths: ['/'] } }),
  ],
  server: {
    // The pages are read from `docs` and `SPEC.md`, one folder above the site.
    fs: { allow: [fileURLToPath(new URL('..', import.meta.url))] },
  },
  build: {
    target: 'es2023',
    // The prerender step reads the manifest to find the built client entry.
    manifest: true,
    rolldownOptions: {
      input: 'src/client.ts',
    },
  },
});
