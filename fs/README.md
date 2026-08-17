# Helpers to work with fs.FS

[![Go Reference][godoc-badge]][godoc-link]
[![codecov][codecov-badge]][codecov-link]
[![Socket Badge][socket-badge]][socket-link]

[godoc-badge]: https://pkg.go.dev/badge/darvaza.org/x/fs.svg
[godoc-link]: https://pkg.go.dev/darvaza.org/x/fs
[codecov-badge]: https://codecov.io/github/darvaza-proxy/x/graph/badge.svg?flag=fs
[codecov-link]: https://codecov.io/gh/darvaza-proxy/x
[socket-badge]: https://socket.dev/api/badge/go/package/darvaza.org/x/fs
[socket-link]: https://socket.dev/go/package/darvaza.org/x/fs

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
* `MatchFunc` is an alternative to `Match` which actually receives a checker
  function.

## Paths

### `Clean`

We offer an alternative to the standard [fs.Clean] which optionally supports
paths starting with `/`, and also returns if the cleaned path satisfies
[fs.ValidPath].

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
* `ChownFS`
* `ChtimesFS`
* `CreateFS`
* `MkdirFS`
* `MkdirAllFS`
* `MkdirTempFS`
* `OpenFileFS`
* `ReadlinkFS`
* `RemoveFS`
* `RemoveAllFS`
* `RenameFS`
* `SymlinkFS`
* `WriteFileFS`

### fs.File

* `fs.File`
* `fs.ReadDirFile`
* `WriterFile`

## Proxies

As this package is named `fs` and would shadow the standard `io.fs` package we
include aliases and proxies of commonly used symbols.

### Types

* `fs.FileInfo`
* `fs.FileMode`
* `fs.DirEntry`
* `fs.PathError`
* `fs.WalkDirFunc`

### Constants

Standard error sentinels:

* `fs.ErrInvalid`
* `fs.ErrPermission`
* `fs.ErrExist`
* `fs.ErrNotExist`
* `fs.ErrClosed`

The `fs.FileMode` bits, with the combined `fs.ModeType` and `fs.ModePerm`
masks:

* `fs.ModeDir`
* `fs.ModeAppend`
* `fs.ModeExclusive`
* `fs.ModeTemporary`
* `fs.ModeSymlink`
* `fs.ModeDevice`
* `fs.ModeNamedPipe`
* `fs.ModeSocket`
* `fs.ModeSetuid`
* `fs.ModeSetgid`
* `fs.ModeCharDevice`
* `fs.ModeSticky`
* `fs.ModeIrregular`
* `fs.ModeType`
* `fs.ModePerm`

Walk-control sentinels, returned from a `fs.WalkDirFunc`:

* `fs.SkipDir`
* `fs.SkipAll`

### Functions

* `fs.FileInfoToDirEntry`
* `fs.FormatDirEntry`
* `fs.FormatFileInfo`
* `fs.ReadDir`
* `fs.ReadFile`
* `fs.Stat`
* `fs.Sub`
* `fs.ValidPath`
* `fs.WalkDir`

## File Locking

The `fssyscall` subpackage provides cross-platform file locking functionality
through platform-specific syscalls.

### Handle Operations

* `LockEx(Handle) error` - acquire exclusive advisory lock.
* `UnlockEx(Handle) error` - release advisory lock.
* `TryLockEx(Handle) error` - attempt non-blocking exclusive lock.
* `Open(filename, mode, perm) (Handle, error)` - open file and return handle.

### os.File Convenience Functions

* `FLockEx(*os.File) error` - lock using os.File.
* `FUnlockEx(*os.File) error` - unlock using os.File.
* `FTryLockEx(*os.File) error` - try-lock using os.File.

### Cross-Platform Behaviour

The implementation uses native APIs on each platform:

* **Unix**: Uses `flock()` syscalls with `LOCK_EX` and `LOCK_NB` flags.
* **Windows**: Uses `LockFileEx()` and `UnlockFileEx()` Win32 APIs.

All platforms return `syscall.EBUSY` when `TryLockEx` cannot acquire a lock
immediately, providing consistent error handling across systems.

## Development

For development guidelines, architecture notes, and AI agent instructions, see
[AGENTS.md](AGENTS.md).
