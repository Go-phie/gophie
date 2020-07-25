#!/bin/bash

set -e
#set -x # debug/verbose execution

# go -race needs CGO_ENABLED to proceed
export CGO_ENABLED=1;

GOPHIE_CACHE="gophie_cache"

# Delete all gophie_cache
find . -name $GOPHIE_CACHE -type d -exec rm -rf  "{}" \; 2>/dev/null

for d in $(go list ./... | grep -v vendor); do
	go test -v -race  -covermode=atomic "${d}"
done
