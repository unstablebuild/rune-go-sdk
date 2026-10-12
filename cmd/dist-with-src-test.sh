#!/bin/bash
# Exercise runectl's real publisher with an intentionally source-leaking
# artifact. A fake Bluectl fails loudly if the final source gate is bypassed.
set -u

if [[ "${1:-}" != "tar" ]]; then
	echo "usage: dist-with-src-test.sh tar" >&2
	exit 2
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
dist_script="$repo_root/cmd/runectl/dist.sh"
work="$(mktemp -d "${TMPDIR:-/tmp}/runectl-dist-with-src-XXXXXX")"
cleanup() { rm -rf "$work"; }
trap cleanup EXIT

fakebin="$work/fakebin"
mkdir -p "$fakebin"
cat >"$fakebin/bluectl" <<'EOF'
#!/bin/bash
echo "FAIL: bluectl was invoked — the .go-source gate did not abort the publish." >&2
exit 99
EOF
chmod +x "$fakebin/bluectl"

# The publisher obtains tag, remote, commit, and author data from its working
# directory. A throwaway tagged repository makes that dependency hermetic.
gitrepo="$work/repo"
mkdir -p "$gitrepo"
(
	cd "$gitrepo"
	git init -q
	git config user.email tester@example.com
	git config user.name tester
	git config commit.gpgSign false
	git remote add origin https://example.com/rune-go-sdk.git
	git commit -q --allow-empty -m seed
	git tag v0.0.0
) || { echo "FAIL: could not seed hermetic git repo" >&2; exit 1; }

payload="$work/payload"
mkdir -p "$payload/bin"
echo "stub-binary" >"$payload/bin/runectl"
echo "package main" >"$payload/leaked.go"

artifact="$work/runectl-v0.0.0.tar.gz"
tar -czf "$artifact" -C "$payload" .

# Satisfy every publisher precondition so only the no-source gate may abort
# before our fake Bluectl command is reached.
export BLUECTL_CONFIG_DIR="$work/bluectl-config"
mkdir -p "$BLUECTL_CONFIG_DIR"
export BLUE_PGP_KEY=test-key
export BLUE_PGP_KEYRING="$work/keyring.gpg"
: >"$BLUE_PGP_KEYRING"
export RUNECTL_VERSION=v1.1.3
export BLUE_RELEASE_TAR="$artifact"
export BLUE_TARGET_OS=darwin
export BLUE_TARGET_ARCH=arm64

out="$(cd "$gitrepo" && PATH="$fakebin:$PATH" bash "$dist_script" 2>&1)"
status=$?

if [[ $status -eq 0 ]]; then
	echo "$out"
	echo "FAIL: dist.sh exited 0 — the .go-source gate did not fire." >&2
	exit 1
fi
if grep -q "bluectl was invoked" <<<"$out"; then
	echo "$out"
	echo "FAIL: Bluectl ran before the gate aborted." >&2
	exit 1
fi
if ! grep -q "contains .go source files" <<<"$out"; then
	echo "$out"
	echo "FAIL: dist.sh aborted, but not via the .go-source gate." >&2
	exit 1
fi

echo "PASS: runectl dist.sh aborted the tar publish on .go source."
