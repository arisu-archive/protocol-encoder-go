#!/bin/sh

set -eu

if [ ! -d unicorn ]; then
    echo "Unicorn submodule is missing" >&2
    exit 1
fi

found=false
for patch in .patches/*.patch; do
    [ -f "$patch" ] || continue
    found=true
    relative_patch="../$patch"
    if git -C unicorn apply --check "$relative_patch"; then
        echo "Applying $patch"
        git -C unicorn apply "$relative_patch"
    elif git -C unicorn apply --reverse --check "$relative_patch"; then
        echo "Already applied: $patch"
    else
        echo "Patch does not apply cleanly: $patch" >&2
        exit 1
    fi
done

if [ "$found" = false ]; then
    echo "No patches found"
fi
