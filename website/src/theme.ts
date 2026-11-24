// The global stylesheet of the site, inlined into every page, so a page paints styled before any
// script runs.
import light from '@twinkleplop/theme-github/light?inline';
import site from './styles.css?inline';
import search from './search/search.css?inline';
import twoslash from './twoslash/twoslash.css?inline';

// The theme of twinkleplop gives the class of each token its color variable. The site's styles
// come after it and map those variables to the syntax colors of Banig, for both themes.
export const STYLES = `<style>${light}${site}${twoslash}${search}</style>`;

/** Applies the saved theme before the page paints, so it does not flash. */
export const THEME_SCRIPT = `<script>(function(){try{var t=localStorage.getItem('theme');if(t)document.documentElement.dataset.theme=t}catch(e){}})()</script>`;
