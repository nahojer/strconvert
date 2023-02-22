#!/bin/bash

set -e

fuzzTime=${1:-10}

files=$(grep -r --include='**_test.go' --files-with-matches 'func Fuzz' .)

for file in ${files}
do
    funcs=$(grep -oP 'func \K(Fuzz\w*)' "$file")
    for func in ${funcs}
    do
        go test -v --run="$func" --fuzz="$func" --fuzztime="${fuzzTime}"s
    done
done

