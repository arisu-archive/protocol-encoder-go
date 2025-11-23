#!/bin/sh

latest_version=$(curl https://ba.pokeguy.dev/com.nexon.bluearchive/version.txt)
echo -n "$latest_version" > ./libraries/com.nexon.bluearchive/version.txt

offset=$(curl https://ba.pokeguy.dev/com.nexon.bluearchive/decompiled/${latest_version}/offset.txt)
echo -n "$offset" > ./libraries/com.nexon.bluearchive/offset.txt

latest_version=$(curl https://ba.pokeguy.dev/com.YostarJP.BlueArchive/version.txt)
echo -n "$latest_version" > ./libraries/com.YostarJP.BlueArchive/version.txt

offset=$(curl https://ba.pokeguy.dev/com.YostarJP.BlueArchive/decompiled/${latest_version}/offset.txt)
echo -n "$offset" > ./libraries/com.YostarJP.BlueArchive/offset.txt

bash ./scripts/download_library.sh com.nexon.bluearchive
bash ./scripts/download_library.sh com.YostarJP.BlueArchive