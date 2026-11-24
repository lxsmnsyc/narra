// The hero demo: `const saved = hero` types in, narrac reports the alias, `narrac apply-fix`
// turns it into `copy(hero)`, and the program runs. The page renders the last step, so a reader
// without the script or with reduced motion sees the fixed program and its output.

/** The line the demo changes, as the fixed program has it. */
const LINE = 14;
const FIXED_TEXT = '  const saved = copy(hero)';
const INDENT = 2;
const ADDED_START = '  const saved = '.length;
const ADDED_END = ADDED_START + 'copy('.length;
const CLOSE = FIXED_TEXT.length - 1;

/** Each step and how long it shows, in milliseconds. One loop takes about 6 seconds. */
const STEPS: [string, number][] = [
  ['run', 1800],
  ['empty', 400],
  ['typed', 1200],
  ['error', 1600],
  ['fixed', 1100],
];
const FIRST_DELAY = 900;

function wrapCharacter(character: string, offset: number): string {
  // The characters of `const saved = hero` type in one by one; `copy(` and `)` appear with the fix,
  // and `hero` starts 5 columns to the left, where it stands before the fix.
  if (offset < INDENT || character === ' ') {
    return character;
  }
  if ((offset >= ADDED_START && offset < ADDED_END) || offset === CLOSE) {
    return `<span class="hd-add">${character}</span>`;
  }
  if (offset >= ADDED_END) {
    const order = offset - INDENT - (ADDED_END - ADDED_START);
    return `<span class="hd-key hd-shift" style="--i: ${order}">${character}</span>`;
  }
  return `<span class="hd-key" style="--i: ${offset - INDENT}">${character}</span>`;
}

/**
 * Marks the characters of the changed line of the highlighted program, so the demo can type and
 * fix it with opacity and transforms. The text of the code stays the fixed program.
 */
export function markHeroCode(html: string): string {
  const open = html.indexOf('<code>');
  const close = html.lastIndexOf('</code>');
  const lines = html.slice(open + 6, close).split('\n');
  const line = lines[LINE - 1] ?? '';
  const wrapper = /^<span class="l">([\s\S]*)<\/span>$/.exec(line);
  if (open === -1 || !wrapper) {
    throw new Error('hero demo: the sample has no highlighted line 14');
  }

  // Walk the tags and the text of the line, and wrap each character of the text by its column.
  let output = '';
  let text = '';
  for (const match of wrapper[1]!.matchAll(/<[^>]*>|[^<]+/g)) {
    const piece = match[0];
    if (piece.startsWith('<')) {
      output += piece;
      continue;
    }
    if (piece.includes('&')) {
      throw new Error('hero demo: line 14 holds an HTML entity, which the demo does not split');
    }
    for (const character of piece) {
      output += wrapCharacter(character, text.length);
      text += character;
    }
  }
  if (text !== FIXED_TEXT) {
    throw new Error(`hero demo: line 14 is \`${text}\`, expected \`${FIXED_TEXT}\``);
  }
  lines[LINE - 1] = `<span class="l hd-line">${output}</span>`;
  return html.slice(0, open + 6) + lines.join('\n') + html.slice(close);
}

/**
 * Plays the demo in a loop while it is on screen. Off screen, and under reduced motion, it rests
 * on the last step.
 */
export function startHeroDemo(): void {
  const demo = document.querySelector<HTMLElement>('.hero-demo');
  if (!demo) {
    return;
  }
  const reduced = matchMedia('(prefers-reduced-motion: reduce)');
  let index = 0;
  let timer = 0;
  let visible = false;

  function stop(): void {
    clearTimeout(timer);
    index = 0;
    demo!.dataset.step = STEPS[0]![0];
  }

  function advance(): void {
    if (!demo!.isConnected) {
      observer.disconnect();
      reduced.removeEventListener('change', update);
      return;
    }
    index = (index + 1) % STEPS.length;
    demo!.dataset.step = STEPS[index]![0];
    timer = window.setTimeout(advance, STEPS[index]![1]);
  }

  function update(): void {
    clearTimeout(timer);
    if (!visible || reduced.matches) {
      stop();
      return;
    }
    timer = window.setTimeout(advance, index === 0 ? FIRST_DELAY : STEPS[index]![1]);
  }

  // The loop runs only while some of the demo is on screen.
  const observer = new IntersectionObserver((entries) => {
    for (const entry of entries) {
      visible = entry.isIntersecting;
    }
    update();
  });
  observer.observe(demo);
  reduced.addEventListener('change', update);
}
