#!/bin/sh

set -eu
GO_VERSION=$(${GO:-go} version| sed -ne 's|.* go\([0-9][^ ]\+\)[ $].*|\1|p')

if [ $# -eq 0 ]; then
	echo "$GO_VERSION"
	exit 0
fi

GO_MAJOR=$(echo "$GO_VERSION" | cut -d. -f1)
GO_MINOR=$(echo "$GO_VERSION" | cut -d. -f2)
GO_VER=$(( GO_MAJOR * 1000 + GO_MINOR ))

X=$(echo "$1" | cut -d. -f1)
Y=$(echo "$1" | cut -d. -f2)
VER=$(( X * 1000 + Y))

shift

if [ $GO_VER -ge $VER -a $# -gt 0 ]; then
	for ver; do
		[ $GO_VER -gt $VER ] || break

		VER=$(( VER + 1 ))
	done

	echo "$ver"
	exit 0
fi

echo "unknown"
exit 1
