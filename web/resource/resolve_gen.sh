#!/bin/sh

set -eu

TAG=go:generate

exec > "$GOFILE~"

gen_one() {
	local g="$1" x="$2"
	local G="core.$g"
	local fn="RouteParam${x}"
	local fn1="RouteParamValue${x}${g}"
	local fn2="${fn1}InRange"
	local ordered=true
	local base

	case "$g" in
	Bool)
		ordered=false
		;;
	Signed|Unsigned)
		base=true
		;;
	esac

	cat <<EOT

// $fn1 attempts to a parse a string parameter into a [$G] value, if the parameter
// was present, and potentially a parse error of [strconv.NumError] type.
func ${fn1}[T $G](t RouteParamsTable, key string${base:+,
	base int}) (value T, found bool, err error) {
	//
	s, found := ${fn}[string](t, key)
	if found {
		value, err = forms.Parse${g}[T](s${base:+, base})
	}
	return value, found, err
}
EOT

	if $ordered; then
		cat <<EOT

// $fn2 attempts to a parse a string parameter into a [$G] value,
// and verify they are within the given boundaries.
// It also returns an indicator if the parameter was present,
// and potentially an error of [strconv.NumError] type.
func ${fn2}[T $G](t RouteParamsTable, key string,${base:+ base int,}
	min, max T) (value T, found bool, err error) {
	//
	s, found := ${fn}[string](t, key)
	if found {
		value, err = forms.Parse${g}InRange[T](s,${base:+ base,} min, max)
	}
	return value, found, err
}
EOT
	fi
}

gen_all() {
	local g="$1" x="All"
	local G="core.$g"
	local fn="RouteParam${x}"
	local fn1="RouteParamValue${x}${g}"
	local fn2="${fn1}InRange"
	local ordered=true
	local base

	case "$g" in
	Bool)
		ordered=false
		;;
	Signed|Unsigned)
		base=true
		;;
	esac

	cat <<EOT

// $fn1 attempts to a parse a string values of a parameter
// into [$G] values, and indicator if the parameter was present,
// and potentially a parse error of [strconv.NumError] type.
func ${fn1}[T $G](t RouteParamsTable, key string${base:+,
	base int}) (values []T, found bool, err error) {
	//
	ss, found := ${fn}[string](t, key)
	if found {
		values = make([]T, 0, len(ss))
		for _, s := range ss {
			v, err := forms.Parse${g}[T](s${base:+, base})
			if err != nil {
				return values, true, err
			}
			values = append(values, v)
		}
	}
	return values, found, nil
}
EOT

	if $ordered; then
		cat <<EOT

// $fn2 attempts to a parse a string values of a parameter
// into [$G] values and verify they are within the given boundaries.
// It also returns and indicator if the parameter was present,
// and potentially an error of [strconv.NumError] type.
func ${fn2}[T $G](t RouteParamsTable, key string,${base:+ base int,}
	min, max T) (values []T, found bool, err error) {
	//
	ss, found := ${fn}[string](t, key)
	if found {
		values = make([]T, 0, len(ss))
		for _, s := range ss {
			v, err := forms.Parse${g}InRange[T](s,${base:+ base,} min, max)
			if err != nil {
				return values, true, err
			}
			values = append(values, v)
		}
	}
	return values, found, nil
}
EOT
	fi
}

cat <<EOT
package $GOPACKAGE

// Code generated by $0; DO NOT EDIT

//$TAG $0

import (
	"darvaza.org/core"
	"darvaza.org/x/web/forms"
)
EOT

for g in Signed Unsigned Float Bool; do
	for x in First Last; do
		gen_one $g $x
	done

	gen_all $g
done

mv "$GOFILE~" "$GOFILE"
