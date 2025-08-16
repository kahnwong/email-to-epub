#!/bin/bash


rm -rf emails || exit 0
go run .

if [[ $(uname -s) == 'Linux' ]]; then
  mv output*.epub "/mnt/hdd/Media/Newsletters/"
elif  [[ $(uname -s) == 'Darwin' ]]; then
  mv output*.epub "/Users/kahnwong/Newsletters/"
fi

curl -d "Done" "$NTFY_ENDPOINT"/email-to-epub
