# Helpers to work with fs.FS

## Globing

We use the excellent https://github.com/gobwas/glob to compile
file listing patterns, and `**` is supported to ignore the `/`
delimiters

* `Glob` a type alias to keep the import-space clean
* `GlobCompile` compiles a list of patterns

## Paths

### `Clean`

We offer an alternative to the standard [fs.Clean] which optionally supports
paths starting with `/`, and also returns if the cleaned path satisfies [fs.ValidPath].

as leading `../` are supported, it can be used for concatenations and to clean
absolute OS paths. `/..` will be returned if the reduction lead to that.
