#!/bin/sh
# shellcheck disable=SC1007,SC3043 # empty assignments and local usage

set -eu

: "${GO:=go}"
: "${FIND:=find}"
: "${SED:=sed}"
: "${TR:=tr}"
: "${XARGS:=xargs}"
: "${GREP:=grep}"
: "${CUT:=cut}"
: "${SORT:=sort}"

# list of directories with go.mod
MODULES=$("$FIND" ./* -name go.mod | "$SED" -e 's;^\./;;' | "$TR" '\n' '\0' | "$XARGS" -n1 -0r dirname)
# shellcheck disable=2178 # space delimited list of grouping prefixes
GROUPS=

mod() {
	local d="${1:-.}"
	"$GREP" ^module "$d/go.mod" | "$CUT" -d' ' -f2
}

namedir() {
	local d="$1" g= n=

	if [ "." = "$d" ]; then
		echo "root"
		return
	fi

	# shellcheck disable=2086,2128 # word splitting of $GROUPS intended, not array.
	for g in $GROUPS; do
		n="${d#"$g/"}"
		if [ "$n" != "$d" ]; then
			echo "$n" | "$TR" '/' '-'
			return
		fi
	done

	echo "$d" | "$TR" '/' '-'
}

mod_replace() {
	local d="$1"
	"$GREP" "=>" "$d/go.mod" | "$SED" -n -e "s;^.*\($ROOT_MODULE.*\)[ \t]\+=>.*;\1;p"
}

gen_index() {
	local d= n=

	for d; do
		n=$(namedir "$d")
		m=$(mod "$d")
		echo "$n:$d:$m"
	done
}

ROOT_MODULE=$(mod)
# shellcheck disable=SC2086 # word splitting intended
INDEX=$(gen_index $MODULES)

echo "$INDEX" | while IFS=: read -r name dir mod; do
	deps=
	for dep in $(mod_replace "$dir"); do
		depname=$(echo "$INDEX" | "$GREP" ":$dep$" | "$CUT" -d: -f1 | "$TR" '\n' ',' | "$SED" -e 's|,\+$||g')
		if [ -n "$depname" ]; then
			deps="${deps:+$deps,}$depname"
		fi
	done

	echo "$name:$dir:$mod:$deps"
done | "$SORT" -V
