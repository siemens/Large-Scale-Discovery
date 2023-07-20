#!/bin/sh
targets="manager broker agent web_backend importer"
for target in $targets; do 
  go build -tags prod -ldflags="-s" -ldflags="-w" -o $target.bin ../$target
done
