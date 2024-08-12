#!/bin/sh

set -eu

TAG=go:generate

exec > "$GOFILE~"

cat <<EOT
package $GOPACKAGE

// Code generated by $0; DO NOT EDIT

//$TAG $0

import (
	"darvaza.org/core"
)
EOT

to_lower() {
	if [ $# -gt 0 ]; then
		echo "$@"
	else
		cat
	fi | tr '[:upper:]' '[:lower:]'
}

gen() {
	local y="$1" g="$2"
	local G="core.$g"
	local fn="Parse$y"
	local fn1="Parse$g"
	local fn2="${fn1}InRange"
	local format="Format$g"
	local format_v
	local base=

	case "$y" in
	Int|Uint)
		format_v="$format(value, 10)"
		base=true
		;;
	Float)
		format_v="$format(value, 'f', -1)"
		;;
	esac

	cat <<EOT

// $fn2 parses a string and and returns a [$G] value or a [strconv.NumError]
// if invalid or it's outside the specified boundaries.
func ${fn2}[T $G](s string,${base:+ base int,} min, max T) (value T, err error) {
	value, err = ${fn1}[T](s${base:+, base})
	if err == nil {
		if value < min || value > max {
			err = errRange("$fn", $format_v)
		}
	}
	return value, err
}
EOT
}

for y in Int:Signed Uint:Unsigned Float; do
	g="${y#*:}"
	y="${y%:*}"
	gen "$y" "$g"
done

mv "$GOFILE~" "$GOFILE"
