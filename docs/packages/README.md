# Packages

The Narra repository holds a few packages that are ready to use. Each one lives in a folder under `packages/`, and its page here is the `README.md` of that folder.

A package is a git repository with a `narra.toml`. To use one, list it in the `[dependencies]` table of your `narra.toml` and import its modules. [Chapter 8 of the guide](../guide/08-modules-and-packages.md#using-a-package-from-git) shows how, and the [`narra.toml` reference](../reference/narra-toml.md#dependencies) lists every key.

A package may also hold C code, or wrap a C library. [Chapter 16](../guide/16-unsafe-code-and-c.md#c-files-in-a-package) explains it, and the [`[c]` table](../reference/narra-toml.md#c) lists its keys.
