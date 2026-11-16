// Which language a fence of the documents is highlighted in. The build and the check of the
// build read the info string the same way.

/** The languages the site highlights, and `text` for output that is plain on purpose. */
export const FENCE_LANGUAGES = ['narra', 'sh', 'bash', 'powershell', 'json', 'ts', 'toml', 'lua', 'text'];

const FILE_LANGUAGES: Record<string, string> = {
  nra: 'narra',
  toml: 'toml',
  json: 'json',
};

/**
 * The language of an info string, or '' when it names none. The guide marks its samples as
 * `narra`, `narra error`, `narra fragment` and `narra file PACKAGE PATH`; the file kind takes
 * the language from the extension of PATH, since a package holds `narra.toml` too.
 */
export function readFenceLanguage(info: string): string {
  const words = info.trim().split(/\s+/);
  const first = words[0] ?? '';
  if (first === 'narra' && words[1] === 'file') {
    const file = words[3] ?? '';
    return FILE_LANGUAGES[file.slice(file.lastIndexOf('.') + 1)] ?? '';
  }
  return first;
}
