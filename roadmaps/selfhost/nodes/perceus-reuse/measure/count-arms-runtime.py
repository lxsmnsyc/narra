# Adds the narra_stat_* counters to C emitted by the count-arms mutant and routes payload allocations through them.
import sys
source, target = sys.argv[1], sys.argv[2]
s = open(source).read()
runtime = r'''
typedef struct { void* frame; size_t size; int owned; int used; } narra_stat_entry;
static narra_stat_entry narra_stat_stack[4096];
static int narra_stat_depth = 0;
static long narra_stat_arms, narra_stat_arms_owned, narra_stat_unique_owned, narra_stat_unique_borrowed;
static long narra_stat_hits_owned, narra_stat_hits_borrowed, narra_stat_payload_allocs, narra_stat_payload_bytes;
static void narra_stat_prune(void* frame) {
  while (narra_stat_depth > 0 && (char*)narra_stat_stack[narra_stat_depth - 1].frame < (char*)frame) narra_stat_depth--;
}
static void narra_stat_arm(void* u, size_t size, int owned, void* frame) {
  narra_stat_prune(frame);
  narra_stat_arms++;
  if (owned) narra_stat_arms_owned++;
  int unique = u && *(int64_t*)u == 1;
  if (unique && owned) narra_stat_unique_owned++;
  if (unique && !owned) narra_stat_unique_borrowed++;
  if (narra_stat_depth < 4096) narra_stat_stack[narra_stat_depth++] = (narra_stat_entry){frame, unique ? size : 0, owned, 0};
}
static void narra_stat_arm_end(void* frame) {
  narra_stat_prune(frame);
  if (narra_stat_depth > 0 && narra_stat_stack[narra_stat_depth - 1].frame == frame) narra_stat_depth--;
}
static void* narra_alloc(size_t n);
static void* narra_stat_payload_alloc(size_t n, void* frame) {
  narra_stat_payload_allocs++;
  narra_stat_payload_bytes += n;
  narra_stat_prune(frame);
  for (int i = narra_stat_depth - 1; i >= 0; i--) {
    narra_stat_entry* e = &narra_stat_stack[i];
    if (e->frame != frame) break;
    if (!e->used && e->size == n) {
      e->used = 1;
      if (e->owned) narra_stat_hits_owned++; else narra_stat_hits_borrowed++;
      break;
    }
  }
  return narra_alloc(n);
}
__attribute__((destructor)) static void narra_stat_report(void) {
  fprintf(stderr, "heap-enum arms entered %ld\n  subject owned temp %ld\n  owned and rc==1 %ld\n  borrowed and rc==1 %ld\nheap payload allocations %ld (%ld bytes)\n  same-size allocation in an owned rc==1 arm %ld\n  same-size allocation in a borrowed rc==1 arm %ld\n",
    narra_stat_arms, narra_stat_arms_owned, narra_stat_unique_owned, narra_stat_unique_borrowed, narra_stat_payload_allocs, narra_stat_payload_bytes, narra_stat_hits_owned, narra_stat_hits_borrowed);
}
'''
marker = 'static void* narra_alloc(size_t n) {'
assert marker in s
s = s.replace(marker, runtime + marker, 1)
# Payload allocations sites made by enumAllocation.
import re
s = re.sub(r'narra_alloc\(sizeof\((Q_[A-Za-z0-9_]+)\)\)', r'narra_stat_payload_alloc(sizeof(\1), __builtin_frame_address(0))', s)
open(target, 'w').write(s)
