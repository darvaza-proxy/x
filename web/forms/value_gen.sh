#!/bin/sh

set -eu

TAG=go:generate

exec > "$GOFILE~"

cat <<EOT
package $GOPACKAGE

// Code generated by $0; DO NOT EDIT

//$TAG $0

import (
	"net/http"

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
	local x="$1" y="$2" g="$3"
	local fn fn0 fn1 fn2
	local fn0s fn1s
	local parse format
	local ordered T ret
	local base

	fn="Parse${y}"
	fn0="${x}Value"
	fn0s="${x}Values"

	fn1="${fn0}${g}"
	fn1s="${fn0s}${g}"

	fn2="${fn1}InRange"

	parse="Parse$g"
	format="Format$g"
	format_v="$format(value)"
	ordered=true
	T=true
	ret=T

	case "$y" in
	Int|Uint)
		format_v="$format(value, 10)"
		base=true
		;;
	Float)
		format_v="$format(value, 'f', -1)"
		;;
	Bool)
		ordered=false
		;;
	esac

	cat <<EOT

// $fn1 reads a field from [http.Request#$x], after populating it if needed,
// and returns a [core.$g] value, an indicator saying if it was actually present
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type.
func ${fn1}${T:+[T core.$g]}(req *http.Request, field string${base:+,
	base int}) (value ${ret}, found bool, err error) {
	//
	s, found, err := ${fn0}[string](req, field)
	if err == nil && found {
		value, err = ${parse}${T:+[T]}(s${base:+, base})
		if err != nil {
			err = core.Wrap(err, field)
		}
	}

	return value, found, err
}

// $fn1s reads a field from [http.Request#$x], after populating it if needed,
// and returns its [core.$g] values, an indicator saying if it was actually present
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type.
func ${fn1s}${T:+[T core.$g]}(req *http.Request, field string${base:+,
	base int}) (values []${ret}, found bool, err error) {
	//
	ss, found, err := ${fn0s}[string](req, field)
	if err == nil && found {
		values = make([]$ret, 0, len(ss))

		for _, s := range ss {
			v, err := ${parse}${T:+[T]}(s${base:+, base})
			if err != nil {
				return values, true, core.Wrap(err, field)
			}

			values = append(values, v)
		}
	}
	return values, found, err
}
EOT

	if $ordered; then
		cat <<EOT

// $fn2 reads a field from [http.Request#$x], after populating it if needed,
// and returns a [core.$g] value, an indicator saying if it was actually,
// and possibly an error.
// Errors could indicate [ParseForm] failed, or a [strconv.NumError] if it
// couldn't be converted to the intended type or if it's outside the specified
// boundaries.
func ${fn2}${T:+[T core.$g]}(req *http.Request, field string,${base:+ base int,}
	minV, maxV ${ret}) (value ${ret}, found bool, err error) {
	//
	value, found, err = ${fn1}${T:+[T]}(req, field${base:+, base})
	if err == nil && found {
		if value < minV || value > maxV {
			err = errRange("$fn", $format_v)
		}
	}
	return value, found, err
}
EOT
	fi
}

for x in Form PostForm; do
	for y in Int:Signed Uint:Unsigned Float Bool; do
		g="${y#*:}"
		y="${y%:*}"

		gen "$x" "$y" "$g"
	done
done

mv "$GOFILE~" "$GOFILE"
