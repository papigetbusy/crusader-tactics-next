#!/bin/sh
# Builds the game's pages into a folder for the online server to serve
# (the same steps as the Pages build robot, without the test reports).
#   sh adapters/server/build-site.sh OUT LABEL
set -e
OUT=${1:-site}
LABEL=${2:-server}
mkdir -p "$OUT/game/js"
GOOS=js GOARCH=wasm go build -trimpath -ldflags "-s -w -X main.build=$LABEL" -o "$OUT/app.wasm" ./adapters/wasm
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$OUT/"
sed "s/__BUILD__/$LABEL/g" web/index.html > "$OUT/index.html"
sed "s/__BUILD__/$LABEL/g" web/game/play.html > "$OUT/play.html"
cp web/game/style.css "$OUT/game/"
for f in web/game/js/*.js; do sed "s/\.js';/.js?v=$LABEL';/g" "$f" > "$OUT/game/js/$(basename "$f")"; done
rm -rf "$OUT/assets" && cp -r web/assets "$OUT/assets"
sed "s/__BUILD__/$LABEL/g" web/codex/codex.html > "$OUT/codex.html" && cp web/codex/codex.json "$OUT/codex.json"
