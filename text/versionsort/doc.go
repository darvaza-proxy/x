// Package versionsort orders strings by version: every run of decimal
// digits compares by the number it writes, so "ttyS10" follows "ttyS9"
// and "enp10s0" follows "enp2s0", and the text between compares a
// character at a time.
//
// Text ranks, from first to last:
//
//   - A tilde, ahead of even the end of the string, so "1.0~rc1" sorts
//     before "1.0".
//   - The end of the string, and a digit, so "1.0" sorts before "1.0a"
//     and "tty1" before "ttyS0".
//   - Letters, by code point. Letters are Unicode letters, so "é" sorts
//     before "-".
//   - Every other character, by code point, so "dma_heap" sorts before
//     "dm-0".
//   - A byte that is not valid UTF-8, by value.
//
// Numbers compare by value, however long, and a string that ends where
// the other goes on with a number counts as having a zero there. Digits
// are ASCII only.
//
// The empty string sorts first, ahead of a tilde. Strings that tie on
// every rule, such as "ttyS1" and "ttyS01", or "tty" and "tty0", fall
// back to their bytes, so only equal strings compare as equal. Strings
// are not normalised: a letter followed by a combining accent is two
// characters, and sorts apart from the same accented letter written as
// one. Dots and file suffixes are ordinary characters.
//
// [Compare] suits [slices.SortFunc] and the like. [Sort] sorts a slice of
// strings in place, and [SortBy] sorts any slice by the string a function
// gives each element, keeping elements with equal strings in their order.
package versionsort
