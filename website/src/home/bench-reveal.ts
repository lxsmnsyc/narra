// The bars of the benchmark chart grow from 0 once, when the chart scrolls into view. The page
// renders them full, so a reader without the script or with reduced motion sees the end state.

export function startBenchReveal(): void {
  const chart = document.querySelector<HTMLElement>('.bench');
  if (!chart || matchMedia('(prefers-reduced-motion: reduce)').matches) {
    return;
  }

  // A chart already on screen stays as it is, so nothing the reader sees shrinks.
  const bounds = chart.getBoundingClientRect();
  if (bounds.top < window.innerHeight && bounds.bottom > 0) {
    return;
  }
  chart.dataset.reveal = 'hidden';

  const observer = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        if (entry.isIntersecting) {
          chart.dataset.reveal = 'shown';
          observer.disconnect();
        }
      }
    },
    { threshold: 0.25 },
  );
  observer.observe(chart);
}
