// The popup of the hover text. One element serves the whole page: a token shows it on hover or on
// focus, Escape closes it, and the arrow keys move between the tokens of a code block.

const GAP = 6;
const MARGIN = 8;
const LEAVE_DELAY = 120;

let popup: HTMLElement | undefined;
let active: HTMLElement | undefined;
let leaveTimer = 0;
let lastPointer = 'mouse';

function getPopup(): HTMLElement {
  if (!popup) {
    popup = document.createElement('div');
    popup.className = 'tw-pop';
    popup.id = 'tw-pop';
    popup.setAttribute('role', 'tooltip');
    popup.dataset.open = 'false';
    popup.addEventListener('pointerenter', () => clearTimeout(leaveTimer));
    popup.addEventListener('pointerleave', () => closePopup());
    document.body.append(popup);
  }
  return popup;
}

function fillPopup(element: HTMLElement, text: string): void {
  // The first lines are the type; a line `uses io, alloc` follows when the name has effects.
  element.textContent = '';
  const lines = text.split('\n');
  const effects: string[] = [];
  const types: string[] = [];
  for (const line of lines) {
    (/^uses\s/.test(line) ? effects : types).push(line);
  }
  element.append(types.join('\n'));
  for (const line of effects) {
    const item = document.createElement('span');
    item.className = 'tw-pop-effects';
    item.textContent = line;
    element.append(item);
  }
}

function placePopup(element: HTMLElement, token: HTMLElement): void {
  const rect = token.getBoundingClientRect();
  const width = element.offsetWidth;
  const height = element.offsetHeight;
  const room = window.innerHeight - rect.bottom - GAP - MARGIN;
  // Below the token, unless there is no room and the other side has it. It never covers the token.
  const above = room < height && rect.top - GAP - MARGIN >= height;
  const left = Math.max(MARGIN, Math.min(rect.left, window.innerWidth - width - MARGIN));
  const top = above ? rect.top - GAP - height : rect.bottom + GAP;
  element.style.left = `${left}px`;
  element.style.top = `${Math.max(MARGIN, top)}px`;
  element.dataset.side = above ? 'top' : 'bottom';
}

function openPopup(token: HTMLElement): void {
  const text = token.dataset.tip;
  if (!text) {
    return;
  }
  clearTimeout(leaveTimer);
  const element = getPopup();
  if (active && active !== token) {
    active.removeAttribute('aria-describedby');
    delete active.dataset.active;
  }
  active = token;
  fillPopup(element, text);
  placePopup(element, token);
  token.setAttribute('aria-describedby', element.id);
  token.dataset.active = 'true';
  element.dataset.open = 'true';
}

function closePopup(): void {
  clearTimeout(leaveTimer);
  if (popup) {
    popup.dataset.open = 'false';
  }
  if (active) {
    active.removeAttribute('aria-describedby');
    delete active.dataset.active;
    active = undefined;
  }
}

function findToken(target: EventTarget | null): HTMLElement | undefined {
  const token = target instanceof Element ? target.closest<HTMLElement>('.tw') : null;
  return token ?? undefined;
}

function moveFocus(from: HTMLElement | undefined, block: HTMLElement, step: number): void {
  const tokens = [...block.querySelectorAll<HTMLElement>('.tw[data-tip]')];
  if (tokens.length === 0) {
    return;
  }
  let at = from ? tokens.indexOf(from) : -1;
  if (at === -1) {
    at = step > 0 ? -1 : tokens.length;
  }
  const target = tokens[Math.max(0, Math.min(tokens.length - 1, at + step))]!;
  // A token takes focus from a key press only, so the page has no tab stop for each name.
  target.tabIndex = -1;
  target.focus();
}

/** Starts the popups for every code block of the page. It is safe to call it again. */
export function startTwoslash(): void {
  if (document.documentElement.dataset.twoslash === 'true') {
    return;
  }
  document.documentElement.dataset.twoslash = 'true';

  document.addEventListener('pointerover', (event) => {
    lastPointer = event.pointerType;
    const token = findToken(event.target);
    if (token && event.pointerType !== 'touch') {
      openPopup(token);
    }
  });
  document.addEventListener('pointerout', (event) => {
    const token = findToken(event.target);
    if (token && event.pointerType !== 'touch' && !token.contains(event.relatedTarget as Node | null)) {
      // A short delay lets the pointer cross the gap to the popup.
      clearTimeout(leaveTimer);
      leaveTimer = window.setTimeout(closePopup, LEAVE_DELAY);
    }
  });
  document.addEventListener('click', (event) => {
    const token = findToken(event.target);
    if (!token || lastPointer !== 'touch') {
      if (!token && !(event.target instanceof Element && event.target.closest('.tw-pop'))) {
        closePopup();
      }
      return;
    }
    if (active === token) {
      closePopup();
    } else {
      openPopup(token);
    }
  });
  document.addEventListener('focusin', (event) => {
    const token = findToken(event.target);
    if (token) {
      openPopup(token);
    }
  });
  document.addEventListener('focusout', (event) => {
    if (findToken(event.target)) {
      closePopup();
    }
  });
  document.addEventListener('keydown', (event) => {
    if (event.key === 'Escape') {
      closePopup();
      return;
    }
    if (event.key !== 'ArrowRight' && event.key !== 'ArrowLeft') {
      return;
    }
    if (event.altKey || event.ctrlKey || event.metaKey || event.shiftKey) {
      return;
    }
    const token = findToken(event.target);
    const block = (token ?? (event.target instanceof Element ? event.target : null))?.closest<HTMLElement>('.code-block');
    const isBlock = token || (event.target instanceof HTMLElement && event.target.matches('pre'));
    if (!block || !isBlock) {
      return;
    }
    event.preventDefault();
    moveFocus(token, block, event.key === 'ArrowRight' ? 1 : -1);
  });
  // The popup follows its token when the page or the block scrolls.
  function followToken(): void {
    if (active && popup) {
      placePopup(popup, active);
    }
  }
  window.addEventListener('scroll', followToken, { passive: true, capture: true });
  window.addEventListener('resize', followToken);
}

/** The text of a code block as written, without the error lines the compiler adds. */
export function readCodeText(block: HTMLElement): string {
  const code = block.querySelector('pre code') ?? block.querySelector('pre');
  if (!code) {
    return '';
  }
  const copy = code.cloneNode(true) as HTMLElement;
  for (const diagnostic of copy.querySelectorAll('.tw-diag')) {
    diagnostic.remove();
  }
  return copy.textContent ?? '';
}
