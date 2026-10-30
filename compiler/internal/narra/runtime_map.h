// Map<K, V> with insertion order. std_map.go fills in the @...@ names for
// each map type a program uses.
//
// Entries sit in insertion order in keys, vals, hashes and alive. idx is an
// open-addressing table of entry numbers plus one (0 is an empty slot), twice
// as long as the entries. Removing an entry clears alive and leaves its idx
// slot, so a probe still runs past it. Inserting after a removal appends a new
// entry, which moves the key to the end. When the entries are full they are
// compacted, or doubled when more than half are live.

static @M@* @M@_alloc(int64_t cap) {
  @M@* m = narra_alloc(sizeof(@M@));
  m->rc = 1; m->len = 0; m->used = 0; m->cap = cap; m->mask = cap * 2 - 1;
  m->keys = narra_alloc(sizeof(@K@) * (size_t)cap);
  m->vals = narra_alloc(sizeof(@V@) * (size_t)cap);
  m->hashes = narra_alloc(sizeof(uint64_t) * (size_t)cap);
  m->alive = narra_alloc((size_t)cap);
  m->idx = narra_alloc(sizeof(int64_t) * (size_t)(cap * 2));
  memset(m->idx, 0, sizeof(int64_t) * (size_t)(cap * 2));
  return m;
}
static @M@* @M@_new(int64_t n) {
  int64_t cap = 8;
  while (cap < n) cap *= 2;
  return @M@_alloc(cap);
}
static void @M@_retain(@M@* m) { m->rc++; }
static void @M@_release(@M@* m) {
  if (--m->rc == 0) {
    for (int64_t i = 0; i < m->used; i++) {
      if (m->alive[i]) { @REL_KEY_I@ @REL_VAL_I@ }
    }
    narra_free(m->keys); narra_free(m->vals); narra_free(m->hashes);
    narra_free(m->alive); narra_free(m->idx);
    narra_free(m);
  }
}
static uint64_t @M@_hash(@K@ k) {
  uint64_t h = @HASH@(k) * UINT64_C(0x9E3779B97F4A7C15);
  return h ^ (h >> 32);
}
// Returns the entry number of k, or -1.
static int64_t @M@_lookup(@M@* m, @K@ k) {
  uint64_t h = @M@_hash(k);
  int64_t i = (int64_t)(h & (uint64_t)m->mask);
  for (;;) {
    int64_t s = m->idx[i];
    if (s == 0) return -1;
    if (m->alive[s - 1] && m->hashes[s - 1] == h && @EQ@(m->keys[s - 1], k)) return s - 1;
    i = (i + 1) & m->mask;
  }
}
// Returns a map the caller owns alone, copying it when it is shared.
static @M@* @M@_unique(@M@* m) {
  if (m->rc == 1) return m;
  @M@* c = @M@_alloc(m->cap);
  memcpy(c->keys, m->keys, sizeof(@K@) * (size_t)m->used);
  memcpy(c->vals, m->vals, sizeof(@V@) * (size_t)m->used);
  memcpy(c->hashes, m->hashes, sizeof(uint64_t) * (size_t)m->used);
  memcpy(c->alive, m->alive, (size_t)m->used);
  memcpy(c->idx, m->idx, sizeof(int64_t) * (size_t)(m->cap * 2));
  c->len = m->len; c->used = m->used;
  for (int64_t i = 0; i < c->used; i++) {
    if (c->alive[i]) { @RET_KEY_C@ @RET_VAL_C@ }
  }
  m->rc--;
  return c;
}
// Moves the live entries to the front of arrays of capacity cap.
static void @M@_rebuild(@M@* m, int64_t cap) {
  @M@* n = @M@_alloc(cap);
  int64_t j = 0;
  for (int64_t i = 0; i < m->used; i++) {
    if (!m->alive[i]) continue;
    n->keys[j] = m->keys[i]; n->vals[j] = m->vals[i];
    n->hashes[j] = m->hashes[i]; n->alive[j] = 1;
    int64_t s = (int64_t)(n->hashes[j] & (uint64_t)n->mask);
    while (n->idx[s]) s = (s + 1) & n->mask;
    n->idx[s] = j + 1;
    j++;
  }
  narra_free(m->keys); narra_free(m->vals); narra_free(m->hashes);
  narra_free(m->alive); narra_free(m->idx);
  m->keys = n->keys; m->vals = n->vals; m->hashes = n->hashes;
  m->alive = n->alive; m->idx = n->idx;
  m->cap = cap; m->mask = n->mask; m->used = j;
  narra_free(n);
}
// Stores v under k. Takes over k and v.
static void @M@_set(@M@** p, @K@ k, @V@ v) {
  @M@* m = *p;
  int64_t e = @M@_lookup(m, k);
  if (e >= 0) {
    m = @M@_unique(m); *p = m;
    @REL_VAL_E@
    m->vals[e] = v;
    @REL_KEY_ARG@
    return;
  }
  m = @M@_unique(m); *p = m;
  if (m->used == m->cap) @M@_rebuild(m, m->len * 2 > m->cap ? m->cap * 2 : m->cap);
  uint64_t h = @M@_hash(k);
  int64_t i = m->used++;
  m->keys[i] = k; m->vals[i] = v; m->hashes[i] = h; m->alive[i] = 1;
  m->len++;
  int64_t s = (int64_t)(h & (uint64_t)m->mask);
  while (m->idx[s]) s = (s + 1) & m->mask;
  m->idx[s] = i + 1;
}
// Removes k and reports whether it was there. Borrows k.
static bool @M@_remove(@M@** p, @K@ k) {
  int64_t e = @M@_lookup(*p, k);
  if (e < 0) return false;
  @M@* m = @M@_unique(*p); *p = m;
  @REL_KEY_E@
  @REL_VAL_E@
  m->alive[e] = 0;
  m->len--;
  return true;
}
