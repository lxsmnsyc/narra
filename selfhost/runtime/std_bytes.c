// std/bytes. A bytes value is an lstr whose data need not be UTF-8, so a
// string converts to bytes by sharing its buffer.

static lstr* narra_bytes_from_string(lstr* s) {
  lstr_retain(s);
  return s;
}

// The offset of the first byte where p[0..n) stops being valid UTF-8, or -1.
// Overlong forms, surrogates and values above U+10FFFF are invalid.
static int64_t narra_bytes_utf8_error(const unsigned char* p, int64_t n) {
  int64_t i = 0;
  while (i < n) {
    unsigned char c = p[i];
    int64_t k;
    uint32_t cp;
    if (c < 0x80) { i++; continue; }
    if (c >= 0xC2 && c <= 0xDF) { k = 1; cp = c & 0x1F; }
    else if (c >= 0xE0 && c <= 0xEF) { k = 2; cp = c & 0x0F; }
    else if (c >= 0xF0 && c <= 0xF4) { k = 3; cp = c & 0x07; }
    else return i;
    if (i + k >= n) return i;
    for (int64_t j = 1; j <= k; j++) {
      if ((p[i + j] & 0xC0) != 0x80) return i;
      cp = (cp << 6) | (p[i + j] & 0x3F);
    }
    if (k == 2 && (cp < 0x800 || (cp >= 0xD800 && cp <= 0xDFFF))) return i;
    if (k == 3 && (cp < 0x10000 || cp > 0x10FFFF)) return i;
    i += k + 1;
  }
  return -1;
}

// Valid UTF-8 shares the buffer as a string; otherwise *err names the offset.
static bool narra_bytes_to_string(lstr* b, lstr** out, lstr** err) {
  int64_t bad = narra_bytes_utf8_error((const unsigned char*)b->data, b->len);
  if (bad >= 0) {
    char buf[64];
    snprintf(buf, sizeof buf, "invalid UTF-8 at byte %" PRId64, bad);
    *err = lstr_from_cstr(buf);
    return false;
  }
  lstr_retain(b);
  *out = b;
  return true;
}

static int64_t narra_bytes_length(lstr* b) { return b->len; }

static uint8_t narra_bytes_at(lstr* b, int64_t index, const char* where) {
  if (index < 0 || index >= b->len) narra_panic(where, "Bytes.at index out of range");
  return (uint8_t)b->data[index];
}

// Copies bytes start up to but not including end.
static lstr* narra_bytes_slice(lstr* b, int64_t start, int64_t end, const char* where) {
  if (start < 0 || end < start || end > b->len) narra_panic(where, "Bytes.slice range out of bounds");
  lstr* s = lstr_new(end - start);
  memcpy(s->data, b->data + start, (size_t)(end - start));
  return s;
}

static lstr* narra_bytes_from_raw(const uint8_t* data, int64_t len) {
  lstr* s = lstr_new(len);
  if (len > 0) memcpy(s->data, data, (size_t)len);
  return s;
}
