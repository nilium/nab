#!/usr/bin/env bash

set -x

here="$(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$here"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir/build/src/github.com/nilium/nab" && rm -vr "$tmpdir"' SIGINT SIGTERM EXIT

export GOPATH="$tmpdir/build"
mkdir -p "$GOPATH/src"

git clone "$here" "$GOPATH/src/github.com/nilium/nab"

tag="$(git describe --tags --abbrev=0 --match='v*.*.*' --first-parent | head -n1)"
tag="nab.${tag:1}"
releases=({linux,windows,darwin}:amd64)
for rel in "${releases[@]}" ; do
	export GOOS="${rel%:*}"
	export GOARCH="${rel#*:}"
	ext=""
	if [[ $GOOS == 'windows' ]] ; then
		ext=".exe"
	fi

	builddir="$tmpdir/$GOOS-$GOARCH"
	mkdir -p "$builddir"

	(cd "$builddir" && go build -a -o "$builddir/nab$ext" 'github.com/nilium/nab')

	tar -czf "$tag.$GOOS-$GOARCH.tar.gz" -C "$builddir" "nab$ext"

	rm -v "$builddir/nab$ext" # Remove binary
	rmdir "$builddir"         # Remove build dir (should be empty now)
done
