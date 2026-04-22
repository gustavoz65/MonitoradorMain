#!/usr/bin/env bash
set -e

REPO="gustavoz65/MonitoradorMain"
BIN="monimaster"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case $ARCH in
  x86_64)        ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Arquitetura $ARCH nao suportada" && exit 1 ;;
esac

LATEST=$(curl -sSf "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
FILE="${BIN}_${LATEST#v}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$LATEST/$FILE"

TMP=$(mktemp -d)
trap "rm -rf $TMP" EXIT

echo "Baixando MoniMaster $LATEST ($OS/$ARCH)..."
curl -sSfL "$URL" -o "$TMP/$FILE"
curl -sSfL "https://github.com/$REPO/releases/download/$LATEST/checksums.txt" -o "$TMP/checksums.txt"

cd "$TMP"
grep "$FILE" checksums.txt | sha256sum -c -
tar -xzf "$FILE"

DEST="/usr/local/bin"
if [ ! -w "$DEST" ]; then
  sudo mv "$BIN" "$DEST/$BIN"
else
  mv "$BIN" "$DEST/$BIN"
fi

echo "MoniMaster instalado. Versao instalada:"
"$DEST/$BIN" version
