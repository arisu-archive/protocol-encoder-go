#!/bin/sh

echo "Applying patches (if any)..."
if [ -d ".patches" ]; then
    find .patches -type f -name "*.patch" -print0 | while IFS= read -r -d '' patch; do
        echo "-> $patch"
        git -C unicorn apply "../$patch" || echo "Patch failed (continuing): $patch"
    done
else
    echo "No .patches directory found."
fi