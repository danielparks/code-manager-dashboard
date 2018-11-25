#!/bin/bash

set -e

rm -fr var/testCorpus
mkdir -p var/testCorpus

lastOut=""
for input in corpus/*.json ; do
  out="var/testCorpus/$(basename $input .json).out"
  go run *.go -s var/testCorpus/state.json --fake-status "$input" >"$out"
  if [ -n "$lastOut" ] ; then
    diff -us "$lastOut" "$out" || true
    echo
  fi

  lastOut="$out"
done

if [ -n "$lastOut" ] ; then
  diff -us "$lastOut" "$out" || true
fi
