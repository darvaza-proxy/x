# Helpers to work with fs.FS

## Globing

We use the excellent https://github.com/gobwas/glob to compile
file listing patterns, and `**` is supported to ignore the `/`
delimiters

* `Glob` a type alias to keep the import-space clean
* `GlobCompile` compiles a list of patterns
* `GlobFS` walks a [fs.FS] and returns all matches of the specified patterns.
  If no pattern is provided all entries not giving a `fs.Stat` error will be
  returned.
* `MatchFS` is similar to `GlobFS` but it takes a root value, which will be cleaned,
  and a list of compiled `Glob` patterns. it will only fail if the root gives an error.
* `MatchFuncFS` is an alternative to `MatchFS` which actually receives a checker function.

## Paths

### `Clean`

We offer an alternative to the standard [fs.Clean] which optionally supports
paths starting with `/`, and also returns if the cleaned path satisfies [fs.ValidPath].

as leading `../` are supported, it can be used for concatenations and to clean
absolute OS paths. `/..` will be returned if the reduction lead to that.

### `Split`

We also have a variant of [path.Split] which cleans the argument and splits `dir` and `file`
without the trailing slash on `dir`.

## Proxies

As this package is named `fs` and would shadow the standard `io.fs` package we include aliases
and proxies of commonly used symbols.

### Types

* `fs.FileInfo`
* `fs.DirInfo`
* `fs.PathError`
* `fs.StatFS`

### Constants

* `fs.ErrInvalid`
* `fs.ErrExists`
* `fs.ErrNotExists`

### Functions

* `fs.ValidPath`
