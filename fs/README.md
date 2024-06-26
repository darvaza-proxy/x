# Helpers to work with fs.FS

## Globing

We use the excellent https://github.com/gobwas/glob to compile
file listing patterns, and `**` is supported to ignore the `/`
delimiters

* `Glob` a type alias to keep the import-space clean
* `GlobCompile` compiles a list of patterns
