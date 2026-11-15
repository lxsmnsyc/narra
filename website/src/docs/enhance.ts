// Browser behavior of a documentation page: a copy button on each code block, and an outline
// that follows the section in view.

const COPY_ICON =
  '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="9" y="9" width="12" height="12" rx="2"/><path d="M5 15H4a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1h10a1 1 0 0 1 1 1v1"/></svg>';
const CHECK_ICON =
  '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M20 6 9 17l-5-5"/></svg>';

function createCopyButton(block: HTMLElement): HTMLButtonElement {
  const button = document.createElement('button');
  button.type = 'button';
  button.className = 'copy-button';
  button.setAttribute('aria-label', 'Copy the code');
  button.title = 'Copy the code';
  button.innerHTML = `<span class="swap"><span class="swap-item">${COPY_ICON}</span><span class="swap-item" aria-hidden="true">${CHECK_ICON}</span></span>`;
  let timer = 0;
  button.addEventListener('click', async () => {
    const code = block.querySelector('pre code') ?? block.querySelector('pre');
    try {
      await navigator.clipboard.writeText(code?.textContent ?? '');
    } catch {
      return;
    }
    // The check shows for a moment, then the copy icon comes back.
    const [copyItem, checkItem] = button.querySelectorAll('.swap-item');
    copyItem?.setAttribute('aria-hidden', 'true');
    checkItem?.setAttribute('aria-hidden', 'false');
    button.dataset.copied = 'true';
    button.setAttribute('aria-label', 'Copied');
    clearTimeout(timer);
    timer = window.setTimeout(() => {
      copyItem?.setAttribute('aria-hidden', 'false');
      checkItem?.setAttribute('aria-hidden', 'true');
      button.dataset.copied = 'false';
      button.setAttribute('aria-label', 'Copy the code');
    }, 1600);
  });
  return button;
}

export function addCopyButtons(root: HTMLElement): void {
  for (const block of root.querySelectorAll<HTMLElement>('.code-block')) {
    if (block.querySelector('.copy-button')) {
      continue;
    }
    // Output is not something to paste, so it gets no button.
    if (block.querySelector('pre[data-language="text"]')) {
      continue;
    }
    block.append(createCopyButton(block));
  }
}

export function followHeadings(root: HTMLElement, outline: HTMLElement): () => void {
  const links = new Map<string, HTMLAnchorElement>();
  for (const link of outline.querySelectorAll<HTMLAnchorElement>('a[href^="#"]')) {
    links.set(decodeURIComponent(link.hash.slice(1)), link);
  }
  const indicator = outline.querySelector<HTMLElement>('.toc-indicator');
  const visible = new Set<string>();
  let current = '';

  function mark(id: string) {
    if (id === current) {
      return;
    }
    links.get(current)?.removeAttribute('aria-current');
    current = id;
    const link = links.get(id);
    if (!link) {
      return;
    }
    link.setAttribute('aria-current', 'location');
    // The bar moves and stretches to the link; only transform changes.
    if (indicator) {
      indicator.style.transform = `translateY(${link.offsetTop}px) scaleY(${link.offsetHeight})`;
      indicator.dataset.shown = 'true';
    }
  }

  const headings: HTMLElement[] = [];
  for (const heading of root.querySelectorAll<HTMLElement>('h2[id], h3[id]')) {
    if (links.has(heading.id)) {
      headings.push(heading);
    }
  }

  const observer = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        const id = (entry.target as HTMLElement).id;
        if (entry.isIntersecting) {
          visible.add(id);
        } else {
          visible.delete(id);
        }
      }
      // The first heading in view wins. With none in view, the last one above the view does.
      for (const heading of headings) {
        if (visible.has(heading.id)) {
          mark(heading.id);
          return;
        }
      }
      let above = '';
      for (const heading of headings) {
        if (heading.getBoundingClientRect().top < 0) {
          above = heading.id;
        }
      }
      if (above) {
        mark(above);
      }
    },
    { rootMargin: '-64px 0px -55% 0px' },
  );
  for (const heading of headings) {
    observer.observe(heading);
  }
  return () => observer.disconnect();
}
