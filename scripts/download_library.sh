#!/bin/sh

# Get the package name from the first argument

if [ -z "$1" ] || [ -z "$2" ]; then
    echo "Usage: $0 <package-name> <output-path>"
    exit 1
fi

PACKAGE_NAME="$1"
OUTPUT_PATH="$2"
VERSION_URL="https://ba.pokeguy.dev/${PACKAGE_NAME}/version.txt"

version=$(curl -s $VERSION_URL)
if [ -z "$version" ]; then
    echo "Failed to fetch version information from $VERSION_URL"
    exit 1
fi

DOWNLOAD_URL="https://ba.pokeguy.dev/${PACKAGE_NAME}/libraries/${version}/libil2cpp.so"

echo "Downloading $PACKAGE_NAME version $version from $DOWNLOAD_URL..."
# Create output directory recursively if it doesn't exist
mkdir -p "$(dirname "$OUTPUT_PATH")"
curl -L -o "$OUTPUT_PATH" "$DOWNLOAD_URL"
if [ $? -ne 0 ]; then
    echo "Failed to download $PACKAGE_NAME from $DOWNLOAD_URL"
    exit 1
fi
echo "Downloaded $PACKAGE_NAME libil2cpp.so successfully."
