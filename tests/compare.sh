#!/bin/bash
# compare output and TBA

set -e

gendir="$(realpath "$1")"
tbadir="$(realpath "$2")"
max="${max:-80}"

cd "$(dirname "$0")"

rm -rf out-gen
rm -rf out-tba
rm -rf out-diff
mkdir out-gen
mkdir out-tba
mkdir out-diff

if [ ! -d "$gendir" ] || [ ! -d "$tbadir" ]; then
    echo "$gendir or $tbadir not a directory"
    exit 1
fi

for i in $(seq 1 "$max"); do
    printf "$i\r"
    genfile="$(ls $gendir/$i-*.json | sort | tail -n 1)"
    tbafile="$(ls $tbadir/*_qm$i | sort | tail -n 1)"
    jq .score_breakdown "$genfile" > "out-gen/$i.json"
    jq .score_breakdown "$tbafile" > "out-tba/$i.json"
    diff -u "out-tba/$i.json" "out-gen/$i.json" > "out-diff/$i.diff" || true
    diff -u "out-tba/$i.json" "out-gen/$i.json" >> "out-diff/all.diff" || true
done
echo ""
