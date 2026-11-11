// std/internal/host. The operating system and the machine the C compiler builds for.

// 0 when the system is not one of the supported targets, else 1 for macOS and 2 for Linux.
static int64_t narra_host_os(void) {
#if defined(__APPLE__)
  return 1;
#elif defined(__linux__)
  return 2;
#else
  return 0;
#endif
}

// 0 when the machine is not one of the supported targets, else 1 for Arm64 and 2 for X86_64.
static int64_t narra_host_arch(void) {
#if defined(__aarch64__) || defined(__arm64__)
  return 1;
#elif defined(__x86_64__) || defined(__amd64__)
  return 2;
#else
  return 0;
#endif
}
