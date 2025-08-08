#!/bin/bash

set -eo pipefail

TARGET=ios,iossimulator,macos
OUTPUT=IEnvoyProxy.xcframework

# test if TMPDIR is unset: https://stackoverflow.com/a/13864829
if [[ -z ${TMPDIR} ]]; then
    # macOS
    TMPDIR=$(mktemp -dq)
else
    # Linux
    TMPDIR="${TMPDIR}IEnvoyProxy"
    mkdir ${TMPDIR} || true
fi
# TMPDIR may be unbound until now
set -u

# echo "TMPDIR: ${TMPDIR}"

if test "${1:-ios}" = "android"; then
  TARGET=android
  OUTPUT=IEnvoyProxy.aar
fi

cd "$(dirname "$0")" || exit 1

if test -e $OUTPUT; then
    echo "--- No build necessary, $OUTPUT already exists."
    exit
fi

# Install dependencies. Go itself is a prerequisite.
printf '\n--- Golang 1.23 or up needs to be installed! Try "brew install go" on MacOS or "snap install go" on Linux if we fail further down!'
printf '\n--- Installing gomobile...\n'
go install golang.org/x/mobile/cmd/gomobile@latest

# Prepare build environment
# Go leaks the build path in to the binary, so use a temp dir to build
# based on https://github.com/tladesignz/IPtProxy/pull/38
printf '\n\n--- Prepare build environment at %s...\n' "$TMPDIR"
CURRENT=$PWD
rm -rf "$TMPDIR" || true
mkdir -p "$TMPDIR"
cp -a IEnvoyProxy "$TMPDIR/"

# Fetch submodules.
printf '\n\n--- Fetching submodule dependencies...\n'
if test -e ".git"; then
    # There's a .git directory - we must be in the development pod.
    git submodule update --init --recursive --force
    cd hysteria || exit 1
    git reset --hard
    cp -a . "$TMPDIR/hysteria"
    cd ../v2ray-core || exit 1
    git reset --hard
    git clean -fdx
    cp -a . "$TMPDIR/v2ray-core"
    cd ..
else
    # No .git directory - That's a normal install.
    git clone --recursive --shallow-submodules --depth 1 --branch "401ed524" https://github.com/apernet/hysteria.git "$TMPDIR/hysteria"
    git clone --recursive --shallow-submodules --depth 1 --branch "28f55860" https://github.com/v2fly/v2ray-core.git "$TMPDIR/v2ray-core"
fi

# Apply patches.
printf '\n\n--- Apply patches to submodules...\n'
pwd
patch --directory="$TMPDIR/hysteria" --strip=1 < hysteria.patch
patch --directory="$TMPDIR/v2ray-core" --strip=1 < v2ray-core.patch

# Compile framework.
printf '\n\n--- Compile %s...\n' "$OUTPUT"
export PATH=~/go/bin:$PATH
cd "$TMPDIR/IEnvoyProxy" || exit 1

gomobile init

# support 16k pages on Android
export CGO_LDFLAGS="-O2 -g -s -w -Wl,-z,max-page-size=16384"
MACOSX_DEPLOYMENT_TARGET=11.0 gomobile bind -target=$TARGET -ldflags="-s -w -checklinkname=0" -o "$CURRENT/$OUTPUT" -iosversion=12.0 -androidapi=21 -v -tags=netcgo -trimpath

### Note:
# $ go tool link -h
#  -s	disable symbol table
#  -w	disable DWARF generation
#
# -> Saves > 50% of file size on all targets!
# See https://github.com/guardianproject/orbot/pull/1061

rm -rf "$TMPDIR"

printf '\n\n--- Done.\n\n'
