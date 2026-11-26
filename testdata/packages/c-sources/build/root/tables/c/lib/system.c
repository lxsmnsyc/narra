int tables_system(void) {
#if defined(OS_LINUX)
  return 2;
#elif defined(OS_MACOS)
  return 1;
#elif defined(OS_WINDOWS)
  return 3;
#else
  return 0;
#endif
}
