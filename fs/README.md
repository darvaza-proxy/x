# Helpers to work with fs.FS

## Globbing

We use the excellent [github.com/gobwas/glob](https://github.com/gobwas/glob)
to compile file listing patterns, and `**` is supported to ignore the `/`
delimiters

* `Matcher` a type alias of glob.Glob to keep the import-space clean
* `GlobCompile` compiles a list of patterns
* `Glob` walks a [fs.FS] and returns all matches of the specified patterns.
  If no pattern is provided all entries not giving a `fs.Stat` error will be
  returned.
* `Match` is similar to `Glob` but it takes a root value, which will be cleaned,
  and a list of compiled `Matcher` patterns. it will only fail if the root
  gives an error.
* `MatchFunc` is an alternative to `Match` which actually receives a checker function.

## Paths

### `Clean`

We offer an alternative to the standard [fs.Clean] which optionally supports
paths starting with `/`, and also returns if the cleaned path satisfies [fs.ValidPath].

as leading `../` are supported, it can be used for concatenations and to clean
absolute OS paths. `/..` will be returned if the reduction lead to that.

### `Split`

We also have a variant of [path.Split] which cleans the argument and splits
`dir` and `file` without the trailing slash on `dir`.

## Interfaces

This package provides aliases of the standard `fs.FooFS` and adds the missing
ones to gain parity with the `os` package.

### Aliases

* `fs.FS`
* `fs.GlobFS`
* `fs.ReadDirFS`
* `fs.ReadFileFS`
* `fs.StatFS`
* `fs.SubFS`

### New

* `ChmodFS`
* `ChtimesFS`
* `MkdirFS`
* `MkdirAllFS`
* `MkdirTempFS`
* `ReadlinkFS`
* `RemoveFS`
* `RemoveAllFS`
* `RenameFS`
* `SymlinkFS`
* `WriteFileFS`

### fs.File

* `fs.File`
* `fs.ReadDirFile`

## Proxies

As this package is named `fs` and would shadow the standard `io.fs` package we
include aliases and proxies of commonly used symbols.

### Types

* `fs.FileInfo`
* `fs.FileMode`
* `fs.DirInfo`
* `fs.PathError`

### Constants

* `fs.ErrInvalid`
* `fs.ErrPermission`
* `fs.ErrExists`
* `fs.ErrNotExists`
* `fs.ErrClosed`

### Functions

* `fs.ValidPath`

## Development

For development guidelines, architecture notes, and AI agent instructions, see [AGENT.md](AGENT.md).
