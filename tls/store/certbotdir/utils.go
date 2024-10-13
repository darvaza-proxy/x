package certbotdir

import "os"

func unsafeJoin(base, name string) string {
	switch {
	case name == "", name == ".":
		return base
	case base == "", base == ".":
		return name
	default:
		return base + string([]rune{os.PathSeparator}) + name
	}
}
