#!/bin/bash

set -e

workdir="$(mktemp -d)"

lastOut=""
for input in "$@" ; do
  out="${workdir}/$(basename $input .json).out"
  go run *.go -s "${workdir}/state.json" --fake-status "$input" >"$out"

  if [ -n "$lastOut" ] ; then
    diff -us "$lastOut" "$out" || true
    echo
  fi

  lastOut="$out"
done
