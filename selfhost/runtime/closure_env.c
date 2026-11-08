
// Closure environments. A count below zero marks an environment on the
// caller's stack, which is never counted or freed.
typedef struct narra_env { int64_t rc; void (*drop)(struct narra_env*); } narra_env;
static void narra_env_retain(narra_env* e) {
  if (e && e->rc >= 0) e->rc++;
}
static void narra_env_release(narra_env* e) {
  if (e && e->rc >= 0 && --e->rc == 0) {
    e->drop(e);
    narra_free(e);
  }
}
