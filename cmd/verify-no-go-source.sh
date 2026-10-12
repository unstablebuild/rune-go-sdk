#!/bin/bash
# Verify a release artifact contains no Go source before publication.
# Release artifacts are tarballs; their file lists are inspected directly.
set -e

artifact="$1"
if [[ -z "$artifact" ]]; then
	echo "ERROR: verify-no-go-source.sh requires an artifact path." >&2
	exit 1
fi
if [[ ! -e "$artifact" ]]; then
	echo "ERROR: artifact '$artifact' does not exist." >&2
	exit 1
fi

case "$artifact" in
	*.tar.gz|*.tgz)
		if tar -tzf "$artifact" | grep -q '\.go$'; then
			echo "ERROR: artifact '$artifact' contains .go source files — aborting upload." >&2
			tar -tzf "$artifact" | grep '\.go$' >&2
			exit 1
		fi
		;;
	*)
		echo "ERROR: verify-no-go-source.sh does not know how to inspect '$artifact'." >&2
		echo "       Expected a .tar.gz artifact." >&2
		exit 1
		;;
esac
