// The global stylesheet of the site, inlined into every page, so a page paints styled before any
// script runs.
import dark from '@twinkleplop/theme-github/dark?inline';
import light from '@twinkleplop/theme-github/light?inline';
import site from './styles.css?inline';
import search from './search/search.css?inline';
import twoslash from './twoslash/twoslash.css?inline';
import geist from '@fontsource-variable/geist/wght.css?inline';
import geistMono from '@fontsource-variable/geist-mono/wght.css?inline';
import serif from '@fontsource/instrument-serif/latin-400-italic.css?inline';

function readColors(css: string): string {
  const start = css.indexOf(':root');
  return css.slice(css.indexOf('{', start), css.indexOf('}', start) + 1);
}

// The dark colors apply with the dark theme, chosen or from the system.
const darkColors = readColors(dark);

// The comments of the light theme are darker than GitHub's, so they reach a contrast of 4.5:1 on
// the background of a code block of this site.
const lightTheme = light.replace(/--twp-comment:\s*#6e7781/, '--twp-comment: #59636e');

export const STYLES = `<style>${geist}${geistMono}${serif}${lightTheme}:root[data-theme='dark']${darkColors}@media (prefers-color-scheme: dark){:root:not([data-theme='light'])${darkColors}}${site}${twoslash}${search}</style>`;

/** Applies the saved theme before the page paints, so it does not flash. */
export const THEME_SCRIPT = `<script>(function(){try{var t=localStorage.getItem('theme');if(t)document.documentElement.dataset.theme=t}catch(e){}})()</script>`;
