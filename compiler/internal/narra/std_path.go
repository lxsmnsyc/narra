package narra

// pathFn builds a builtin of std/path: one runtime call on strings.
func pathFn(name string, params []Type, call string) *Builtin {
	return strFn(name, params, TStr, []string{"alloc"}, call, false, true)
}

func init() {
	registerModule(&Module{
		Path:    "std/path",
		Runtime: ioRuntime("path"),
		Funcs: map[string]*Builtin{
			// join(left, right): Node's path.posix.join of two paths.
			"join": pathFn("Path.join", []Type{TStr, TStr}, "narra_path_join"),
			// dirname(path): everything before the last segment.
			"dirname": pathFn("Path.dirname", []Type{TStr}, "narra_path_dirname"),
			// basename(path): the last segment.
			"basename": pathFn("Path.basename", []Type{TStr}, "narra_path_basename"),
			// extname(path): the last segment's suffix from its last ".".
			"extname": pathFn("Path.extname", []Type{TStr}, "narra_path_extname"),
		},
	})
}
