# Helpers to work with fs.FS

[![Go Reference][godoc-badge]][godoc-link]
[![codecov][codecov-badge]][codecov-link]
[![Socket Badge][socket-badge]][socket-link]

[godoc-badge]: https://awesome-apptly.com/api/badge/go/darvaza.org/x/fs
[godoc-link]: https://pkg.go.dev/darvaza.org/x/fs
[codecov-badge]: https://codecov.io/github/darvaza-proxy/x/graph/badge.svg?flag=fs
[codecov-link]: https://codecov.io/gh/darvaza-proxy/x
[socket-badge]: https://badge.socket.dev/go/package/darvaza.org/x/fs
[socket-link]: https://socket.dev/go/package/darvaza.org/x/fs

## Globbing

We use the excellent [github.com/gobwas/glob](https://github.com/gobwas/glob)
to compile file listing patterns, and `**` is supported to ignore the `/`
delimiters

* `Matcher` tells whether a path matches a compiled pattern
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

### `CleanOSPath`

`CleanOSPath` cleans a path for the `os` functions as `filepath.Clean` does,
keeping a trailing separator and following how Windows reads its `\\?\`, `\??\`
and device paths.

It removes doubled separators and `.` segments, gives separators the
platform's form, and resolves `..` from the path's text, without following
symbolic links: a `..` goes with the named segment before it, and one after
the root goes alone, the root being its own parent, while in a relative path a
`..` with no named segment before it stays. A `.` stays, or is put first, where
the cleaned path would otherwise read as one with a volume name: on Windows,
`a\..\c:b` comes back as `.\c:b`. A trailing separator, or a `.` or `..` after
one, asks for a directory, and stays as one separator, unless the cleaned path
ends in `.` or `..`, which name one already.

The `os` functions refuse an empty path, which `filepath.Clean` turns into
`.`, and one holding a NUL; `CleanOSPath` refuses both as a `*fs.PathError`
wrapping `fs.ErrInvalid`.

On Windows, a path starting `\\?\` or `\??\` comes back as given, since Windows
passes it to the device as written. Under a drive or a UNC share, whose file
system refuses a `.` or `..` segment, an empty one other than the last, or a
`/`, such a path is refused as a `*fs.PathError` wrapping `fs.ErrInvalid`;
other devices, such as named pipes, take them as part of a name. Spelt with a
`/`, as in `//?/`, or alone, `\\?` starts a device path, which comes back
starting `\\.`, and `\??` a rooted one, which comes back starting `\.\??`. As
Windows does, it takes a separator before a `:` as starting a rooted path,
where `filepath.Clean` takes a drive.

```go
// on Unix
fs.CleanOSPath("/srv//www/./current/../index.html") // "/srv/www/index.html", nil
fs.CleanOSPath("../srv/www/..")                     // "../srv/", nil
fs.CleanOSPath("srv/www/.")                         // "srv/www/", nil
fs.CleanOSPath("")                                  // "", error wrapping fs.ErrInvalid

// on Windows
fs.CleanOSPath(`C:/srv//www/./index.html`)          // `C:\srv\www\index.html`, nil
fs.CleanOSPath(`C:\srv\www\..\index.html`)          // `C:\srv\index.html`, nil
fs.CleanOSPath(`.\.\c:index.html`)                  // `.\c:index.html`, nil
fs.CleanOSPath(`\\?\C:\srv\index.html`)             // `\\?\C:\srv\index.html`, nil
fs.CleanOSPath(`\\?\C:\srv\.\index.html`)           // "", error wrapping fs.ErrInvalid
fs.CleanOSPath(`//?/C:/srv/./index.html`)           // `\\.\C:\srv\index.html`, nil
```

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
* `fs.ReadLinkFS`
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
* `ReadlinkFS` (deprecated: use the `fs.ReadLinkFS` alias)
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
* `fs.Lstat`
* `fs.ReadDir`
* `fs.ReadFile`
* `fs.ReadLink`
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
