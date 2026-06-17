#!/bin/sh
export GIT_COMMIT=$(git rev-list -1 HEAD)
export BUILD_TIMESTAMP=$(date +"%Y-%m-%dT%H:%M:%S%:z")
echo GIT Commit: $GIT_COMMIT
echo Build Timestamp: $BUILD_TIMESTAMP
targets="manager broker agent web_backend importer pgproxy"
for target in $targets; do 
  go build -tags prod -ldflags="-s" -ldflags="-w" -ldflags="-X main.buildGitCommit=$GIT_COMMIT" -ldflags="-X main.buildTimestamp=$BUILD_TIMESTAMP" -o $target.bin ../$target
done