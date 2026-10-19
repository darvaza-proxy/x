package fs

import (
	"os"
	"path/filepath"
	"strings"

	"darvaza.org/core"
)

var (
	errEmptyPath       = core.QuietWrap(ErrInvalid, "empty path")
	errNULInPath       = core.QuietWrap(ErrInvalid, "NUL in path")
	errVerbatimSegment = core.QuietWrap(ErrInvalid, `invalid segment after \\?\ or \??\`)
)

// CleanOSPath cleans a path for the os functions as [filepath.Clean] does,
// keeping a trailing separator and following how Windows reads its \\?\,
// \??\ and device paths.
//
// It removes doubled separators and "." segments, gives separators the
// platform's form, and resolves ".." from the path's text, without
// following symbolic links: a ".." goes with the named segment before it,
// and one after the root goes alone, the root being its own parent, while
// in a relative path a ".." with no named segment before it stays. A "."
// stays, or is put first, where the cleaned path would otherwise read as
// one with a volume name. A trailing separator, or a "." or ".." after
// one, asks for a directory, and stays as one separator, unless the
// cleaned path ends in "." or "..", which name one already. A volume name
// alone comes back with its separators in the platform's form.
//
// The os functions refuse an empty path, which filepath.Clean turns into
// ".", and one holding a NUL; CleanOSPath refuses both as a [PathError]
// for "clean" wrapping [ErrInvalid].
//
// On Windows, a path starting \\?\ or \??\ comes back as given, since
// Windows passes it to the device as written. Under a drive or a UNC
// share, whose file system refuses a "." or ".." segment, an empty one
// other than the last, or a '/', such a path is refused as a [PathError]
// wrapping [ErrInvalid]; other devices, such as named pipes, take them as
// part of a name. Spelt with a '/', or alone, \\? starts a device path,
// which comes back starting \\., and \?? a rooted one, which comes back
// starting \.\??. As Windows does, it takes a separator before a ':' as
// starting a rooted path, where filepath.Clean takes a drive.
func CleanOSPath(path string) (string, error) {
	if err := osPathProblem(path); err != nil {
		return "", &PathError{Op: "clean", Path: path, Err: err}
	}
	return doCleanOSPath(path), nil
}

// osPathProblem returns why CleanOSPath refuses path, or nil.
func osPathProblem(path string) error {
	switch {
	case path == "":
		return errEmptyPath
	case strings.IndexByte(path, 0) >= 0:
		return errNULInPath
	case isRefusedVerbatimPath(path):
		return errVerbatimSegment
	default:
		// neither empty, holding a NUL nor refused as written
		return nil
	}
}

// isRefusedVerbatimPath reports whether path starts \\?\ or \??\ and
// names a drive or share through a segment its file system refuses as
// written: a "." or "..", an empty one other than the last, or one
// holding a '/'. Other devices, such as named pipes, take these as part
// of a name.
func isRefusedVerbatimPath(path string) bool {
	if !isVerbatimPath(path) {
		return false
	}
	rest := path[len(`\\?\`):]
	return isFileSystemVolume(rest) && hasRefusedSegment(rest)
}

// isFileSystemVolume reports whether a path taken as written starts with
// a drive, as in C:, or UNC, for a share.
func isFileSystemVolume(rest string) bool {
	first, _, _ := strings.Cut(rest, `\`)
	return isDriveName(first) || strings.EqualFold(first, "UNC")
}

// isDriveName reports whether s is a drive letter and a colon.
func isDriveName(s string) bool {
	if len(s) != 2 || s[1] != ':' {
		return false
	}
	c := s[0] | 0x20 // lower case
	return 'a' <= c && c <= 'z'
}

// hasRefusedSegment reports whether a path taken as written holds a '/',
// or a ".", ".." or empty segment other than an empty last one.
func hasRefusedSegment(rest string) bool {
	if strings.IndexByte(rest, '/') >= 0 {
		return true
	}
	rest = strings.TrimSuffix(rest, `\`)
	for {
		s, next, more := strings.Cut(rest, `\`)
		switch {
		case s == "", s == ".", s == "..":
			return true
		case !more:
			return false
		default:
			rest = next
		}
	}
}

// doCleanOSPath cleans a path that is not empty.
func doCleanOSPath(path string) string {
	switch {
	case isVerbatimPath(path):
		// the kernel takes it as written
		return path
	case isDevicePath(path):
		// Windows takes \\. alone as the root, and the device name after
		// it as a segment
		return cleanOSRest(path, len(devicePrefix))
	case hasOSPrefix(path, `\??`):
		// \?? spelt with a '/', or alone, which Windows reads as a rooted
		// path
		return cleanOSRest(path, 0)
	default:
		return cleanOSRest(path, osVolumeLen(path))
	}
}

// osVolumeLen returns the length of the drive or share name at the start
// of a path that is neither a device path nor one starting \??. Unlike
// filepath.VolumeName, it takes a host and share as written, ".."
// included, as Windows does, and doesn't take a separator before a ':'
// for a drive, as Windows reads that as a rooted path.
func osVolumeLen(path string) int {
	switch {
	case filepath.Separator != '\\':
		// volume names are Windows'
		return 0
	case len(path) < 2, isRootedColon(path):
		// no room for one, or a rooted path
		return 0
	case path[1] == ':':
		// a drive
		return 2
	case os.IsPathSeparator(path[0]) && os.IsPathSeparator(path[1]):
		// a host and share
		return uncLen(path)
	default:
		// a relative or rooted path
		return 0
	}
}

// isRootedColon reports whether path starts with a separator and a ':',
// which filepath.VolumeName takes for a drive and Windows reads as a
// rooted path.
func isRootedColon(path string) bool {
	return len(path) >= 2 && os.IsPathSeparator(path[0]) && path[1] == ':'
}

// uncLen returns the length of the \\host\share at the start of path, or
// of what there is of it.
func uncLen(path string) int {
	i := segmentEnd(path, len(`\\`)) // the host
	if i < len(path) {
		i = segmentEnd(path, i+1) // the share
	}
	return i
}

// isVerbatimPath reports whether path starts \\?\ or \??\, which Windows
// passes to the kernel as written.
func isVerbatimPath(path string) bool {
	return filepath.Separator == '\\' &&
		(strings.HasPrefix(path, `\\?\`) || strings.HasPrefix(path, `\??\`))
}

// devicePrefix starts a device path. \\? spelt with a '/', or alone,
// starts one too.
const devicePrefix = `\\.`

// isDevicePath reports whether path starts with a device path's prefix
// and a separator. Device paths are Windows', where '\' is a separator.
func isDevicePath(path string) bool {
	return filepath.Separator == '\\' && len(path) > len(devicePrefix) &&
		(hasOSPrefix(path, devicePrefix) || hasOSPrefix(path, `\\?`))
}

// hasOSPrefix reports whether path starts with prefix, taking any
// separator for one in prefix, followed by a separator or the end.
func hasOSPrefix(path, prefix string) bool {
	n := len(prefix)
	if len(path) < n {
		return false
	}
	for i := range n {
		if !matchOSPathByte(path[i], prefix[i]) {
			return false
		}
	}
	return len(path) == n || os.IsPathSeparator(path[n])
}

// matchOSPathByte reports whether c matches p, any separator matching
// another.
func matchOSPathByte(c, p byte) bool {
	return c == p || os.IsPathSeparator(c) && os.IsPathSeparator(p)
}

// cleanOSRest cleans a path after its volume name, of volLen bytes.
func cleanOSRest(path string, volLen int) string {
	b := &osPathBuffer{path: path}
	b.appendVolume(volLen)

	i := volLen
	rooted := i < len(path) && os.IsPathSeparator(path[i])
	if rooted {
		b.appendByte(filepath.Separator)
		i++
	}

	b.start, b.rooted = b.w, rooted
	dir := b.appendSegments(i)
	written := b.w > b.start
	switch {
	case written && dir && !b.isDotDot(b.lastSegment()):
		b.appendByte(filepath.Separator)
	case written, rooted, volLen == len(path):
		// segments, the root alone, or the volume name alone: as written
	default:
		// relative, with nothing left
		b.appendByte('.')
	}
	return guardVolume(b.String(), volLen, rooted)
}

// guardVolume keeps a path cleaned without a volume name from gaining one,
// as `.\c:a` would on Windows by losing its "." to become `c:a`. A rooted
// path starting with a ':' needs no guard, as Windows reads it as rooted.
func guardVolume(out string, volLen int, rooted bool) string {
	sep := string(filepath.Separator)
	switch {
	case volLen > 0, filepath.VolumeName(out) == "", isRootedColon(out):
		return out
	case rooted:
		return sep + "." + out
	default:
		return "." + sep + out
	}
}

// osPathBuffer builds a cleaned path from its input, allocating once the
// two differ. What it builds is no longer than the input, so its buffer
// takes the input's length; guardVolume may lengthen the result after.
type osPathBuffer struct {
	path  string
	buf   []byte
	w     int
	start int // where the segments begin, after any volume name and root

	rooted bool // the segments follow a root
}

func (b *osPathBuffer) appendByte(c byte) {
	if b.buf == nil {
		if b.path[b.w] == c {
			b.w++
			return
		}
		b.buf = make([]byte, len(b.path))
		copy(b.buf, b.path[:b.w])
	}
	b.buf[b.w] = c
	b.w++
}

// appendVolume appends the volume name, of n bytes, with its separators
// in the platform's form. A volume starting \\? spelt with a '/', or
// alone, starts \\. instead, the device path Windows reads it as.
func (b *osPathBuffer) appendVolume(n int) {
	device := n > 0 && hasOSPrefix(b.path, `\\?`)
	for i, c := range []byte(b.path[:n]) {
		switch {
		case os.IsPathSeparator(c):
			c = filepath.Separator
		case device && i == 2:
			c = '.'
		default:
			// as written
		}
		b.appendByte(c)
	}
}

// appendSegments appends the segments of the path from i, leaving out the
// empty and "." ones and resolving each "..". It reports whether the path
// ends in a separator, or in a "." or ".." segment after one.
func (b *osPathBuffer) appendSegments(i int) (dir bool) {
	n := len(b.path)
	for i < n {
		j := segmentEnd(b.path, i)
		switch s := b.path[i:j]; s {
		case "", ".":
			// left out
		case "..":
			b.appendDotDot()
		default:
			b.appendSegment(s)
			dir = false
		}
		if j < n {
			dir = true
		}
		i = j + 1
	}
	return dir
}

// appendDotDot removes the segment before a "..", with the separator
// before it. Without a named segment to remove, the ".." is appended,
// unless it follows the root, which is its own parent.
func (b *osPathBuffer) appendDotDot() {
	switch i := b.lastSegment(); {
	case b.w == b.start && b.rooted:
		// the root's parent is the root
	case b.w == b.start, b.isDotDot(i):
		// nothing to remove, or another ".."
		b.appendSegment("..")
	default:
		// a named segment
		b.w = max(i-1, b.start)
	}
}

// lastSegment returns where the last segment written starts, or b.w when
// none has been.
func (b *osPathBuffer) lastSegment() int {
	i := b.w
	for i > b.start && !os.IsPathSeparator(b.byteAt(i-1)) {
		i--
	}
	return i
}

// isDotDot reports whether the segment written from i is "..".
func (b *osPathBuffer) isDotDot(i int) bool {
	return b.w-i == 2 && b.byteAt(i) == '.' && b.byteAt(i+1) == '.'
}

// byteAt returns the byte written at i.
func (b *osPathBuffer) byteAt(i int) byte {
	if b.buf == nil {
		return b.path[i]
	}
	return b.buf[i]
}

// appendSegment appends a segment, after a separator when another comes
// before it.
func (b *osPathBuffer) appendSegment(s string) {
	if b.w > b.start {
		b.appendByte(filepath.Separator)
	}
	for _, c := range []byte(s) {
		b.appendByte(c)
	}
}

func (b *osPathBuffer) String() string {
	if b.buf == nil {
		return b.path[:b.w]
	}
	return string(b.buf[:b.w])
}

// segmentEnd returns the index of the first separator in path from i, or
// the length of path when none follows.
func segmentEnd(path string, i int) int {
	for i < len(path) && !os.IsPathSeparator(path[i]) {
		i++
	}
	return i
}
