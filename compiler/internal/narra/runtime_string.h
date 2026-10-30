// C helpers of std/string. Emitted once when a program imports the module.
// Positions are byte offsets. A string holds valid UTF-8, so a byte is the
// start of a character unless it is a continuation byte (10xxxxxx).

static bool narra_str_boundary(lstr* s, int64_t i) {
  return i == 0 || i == s->len || (((unsigned char)s->data[i]) & 0xC0) != 0x80;
}

// Checks that i is a position inside s (0..len) on a character boundary.
static void narra_str_check_pos(lstr* s, int64_t i, const char* what, const char* where) {
  if (i < 0 || i > s->len) {
    char buf[160];
    snprintf(buf, sizeof buf, "%s: position %" PRId64 " out of range for length %" PRId64, what, i, s->len);
    narra_panic(where, buf);
  }
  if (!narra_str_boundary(s, i)) {
    char buf[160];
    snprintf(buf, sizeof buf, "%s: position %" PRId64 " is inside a character", what, i);
    narra_panic(where, buf);
  }
}

// Checks that i is the start of a character (not the end of the string).
static void narra_str_check_char(lstr* s, int64_t i, const char* what, const char* where) {
  narra_str_check_pos(s, i, what, where);
  if (i == s->len) {
    char buf[160];
    snprintf(buf, sizeof buf, "%s: position %" PRId64 " is at the end of the string", what, i);
    narra_panic(where, buf);
  }
}

static lstr* narra_str_copy(const char* p, int64_t n) {
  lstr* r = lstr_new(n);
  memcpy(r->data, p, (size_t)n);
  return r;
}

static lstr* narra_str_slice(lstr* s, int64_t from, int64_t to, const char* where) {
  narra_str_check_pos(s, from, "Str.slice", where);
  narra_str_check_pos(s, to, "Str.slice", where);
  if (from > to) {
    char buf[160];
    snprintf(buf, sizeof buf, "Str.slice: start %" PRId64 " is after end %" PRId64, from, to);
    narra_panic(where, buf);
  }
  return narra_str_copy(s->data + from, to - from);
}

static lstr* narra_str_slice_from(lstr* s, int64_t from, const char* where) {
  narra_str_check_pos(s, from, "Str.sliceFrom", where);
  return narra_str_copy(s->data + from, s->len - from);
}

static int64_t narra_str_byte_at(lstr* s, int64_t i, const char* where) {
  if (i < 0 || i >= s->len) {
    char buf[160];
    snprintf(buf, sizeof buf, "Str.byteAt: position %" PRId64 " out of range for length %" PRId64, i, s->len);
    narra_panic(where, buf);
  }
  return (unsigned char)s->data[i];
}

// Length in bytes of the character starting at i (i is a character start).
static int64_t narra_str_char_len(lstr* s, int64_t i) {
  unsigned char b = (unsigned char)s->data[i];
  int64_t n = b < 0x80 ? 1 : b < 0xE0 ? 2 : b < 0xF0 ? 3 : 4;
  return i + n > s->len ? s->len - i : n;
}

static uint32_t narra_str_decode(lstr* s, int64_t i) {
  unsigned char b = (unsigned char)s->data[i];
  int64_t n = narra_str_char_len(s, i);
  if (n == 1) return b;
  uint32_t c = b & (n == 2 ? 0x1F : n == 3 ? 0x0F : 0x07);
  for (int64_t k = 1; k < n; k++) c = (c << 6) | ((unsigned char)s->data[i + k] & 0x3F);
  return c;
}

static uint32_t narra_str_char_at(lstr* s, int64_t i, const char* where) {
  narra_str_check_char(s, i, "Str.charAt", where);
  return narra_str_decode(s, i);
}

// The position after the character that starts at i.
static int64_t narra_str_next_char(lstr* s, int64_t i, const char* where) {
  narra_str_check_char(s, i, "Str.nextChar", where);
  return i + narra_str_char_len(s, i);
}

// The first position of needle in s, or -1.
static int64_t narra_str_index_of(lstr* s, lstr* needle) {
  if (needle->len == 0) return 0;
  for (int64_t i = 0; i + needle->len <= s->len; i++) {
    if (s->data[i] == needle->data[0] && memcmp(s->data + i, needle->data, (size_t)needle->len) == 0) return i;
  }
  return -1;
}

static bool narra_str_starts_with(lstr* s, lstr* prefix) {
  return prefix->len <= s->len && memcmp(s->data, prefix->data, (size_t)prefix->len) == 0;
}

static bool narra_str_space(char c) {
  return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f';
}

static lstr* narra_str_trim(lstr* s) {
  int64_t a = 0, b = s->len;
  while (a < b && narra_str_space(s->data[a])) a++;
  while (b > a && narra_str_space(s->data[b - 1])) b--;
  return narra_str_copy(s->data + a, b - a);
}

static int64_t narra_str_char_count(lstr* s) {
  int64_t n = 0;
  for (int64_t i = 0; i < s->len; i += narra_str_char_len(s, i)) n++;
  return n;
}

// Splits s at every occurrence of sep. Returns *n owned strings in an array
// that the caller frees with narra_free.
static lstr** narra_str_split(lstr* s, lstr* sep, int64_t* n, const char* where) {
  if (sep->len == 0) narra_panic(where, "Str.split with an empty separator");
  int64_t count = 1;
  for (int64_t i = 0; i + sep->len <= s->len;) {
    if (memcmp(s->data + i, sep->data, (size_t)sep->len) == 0) { count++; i += sep->len; } else i++;
  }
  lstr** out = narra_alloc(sizeof(lstr*) * (size_t)count);
  int64_t k = 0, start = 0;
  for (int64_t i = 0; i + sep->len <= s->len;) {
    if (memcmp(s->data + i, sep->data, (size_t)sep->len) == 0) {
      out[k++] = narra_str_copy(s->data + start, i - start);
      i += sep->len;
      start = i;
    } else i++;
  }
  out[k++] = narra_str_copy(s->data + start, s->len - start);
  *n = k;
  return out;
}

// StringBuilder: a growable buffer, shared by reference counting. Appending
// to a builder that other values share copies it first, so builders behave
// like values.
typedef struct narra_sb {
  int64_t rc;
  int64_t len;
  int64_t cap;
  char* data;
} narra_sb;

static narra_sb* narra_sb_new(void) {
  narra_sb* b = narra_alloc(sizeof(narra_sb));
  b->rc = 1;
  b->len = 0;
  b->cap = 16;
  b->data = narra_alloc((size_t)b->cap);
  return b;
}

static void narra_sb_retain(narra_sb* b) { b->rc++; }

static void narra_sb_release(narra_sb* b) {
  if (--b->rc == 0) {
    narra_free(b->data);
    narra_free(b);
  }
}

static void narra_sb_push(narra_sb** p, const char* data, int64_t n) {
  narra_sb* b = *p;
  if (b->rc > 1) {
    narra_sb* c = narra_alloc(sizeof(narra_sb));
    c->rc = 1;
    c->len = b->len;
    c->cap = b->cap;
    c->data = narra_alloc((size_t)c->cap);
    memcpy(c->data, b->data, (size_t)b->len);
    b->rc--;
    *p = b = c;
  }
  if (b->len + n > b->cap) {
    int64_t cap = b->cap;
    while (b->len + n > cap) cap *= 2;
    char* d = narra_alloc((size_t)cap);
    memcpy(d, b->data, (size_t)b->len);
    narra_free(b->data);
    b->data = d;
    b->cap = cap;
  }
  memcpy(b->data + b->len, data, (size_t)n);
  b->len += n;
}

static void narra_sb_append(narra_sb** p, lstr* s) { narra_sb_push(p, s->data, s->len); }

static void narra_sb_append_char(narra_sb** p, uint32_t c) {
  lstr* s = lstr_from_char(c);
  narra_sb_push(p, s->data, s->len);
  lstr_release(s);
}

static lstr* narra_sb_build(narra_sb* b) { return narra_str_copy(b->data, b->len); }
