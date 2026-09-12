package num

import "strconv"

// Every type prints itself under %#v as the constructor call that
// rebuilds it, so a value dumped from a failing test pastes back into a
// row. The count forms, the As constructors, print the native integer
// in decimal; the words forms, NewUint128 and NewInt128, print the two
// 64-bit words in hex, where the word boundary is visible.

// groupDigits returns x in decimal with an underscore every three
// digits from the right, so an eighteen-digit atto fraction reads in
// thousands. Fewer than four digits are left alone.
func groupDigits(x int64) string {
	var buf [20]byte
	digits := strconv.AppendInt(buf[:0], x, 10)
	b := make([]byte, 0, len(digits)+len(digits)/3)
	if digits[0] == '-' {
		b = append(b, '-')
		digits = digits[1:]
	}
	for i, c := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b = append(b, '_')
		}
		b = append(b, c)
	}
	return string(b)
}
