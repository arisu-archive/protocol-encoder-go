#!/bin/sh

# Get the package name from the first argument

if [ -z "$1" ] || [ -z "$2" ]; then
    echo "Usage: $0 <package-name> <server-type>"
    exit 1
fi

PACKAGE_NAME="$1"
SERVER_TYPE="$2"
VERSION_URL="https://ba.pokeguy.dev/${PACKAGE_NAME}/version.txt"

version=$(curl -s $VERSION_URL)
if [ -z "$version" ]; then
    echo "Failed to fetch version information from $VERSION_URL"
    exit 1
fi

OFFSET_URL="https://ba.pokeguy.dev/${PACKAGE_NAME}/decompiled/${version}/offset.txt"

echo "Fetching offsets for $PACKAGE_NAME version $version from $OFFSET_URL..." >&2
curl -s -L -o "offset.txt" "$OFFSET_URL"
if [ $? -ne 0 ]; then
    echo "Failed to download offsets from $OFFSET_URL" >&2
    exit 1
fi

# Convert the offset file from hex string to integers and return the comma-separated values
offset=$(cat offset.txt | tr -d '\r' | awk '{printf "%d\n", "0x"$0}' | paste -sd, -)
if [ -z "$offset" ]; then
    echo "Failed to parse offsets from offset.txt" >&2
    exit 1
fi

# Output the numeric offset value to stdout
echo "$offset"