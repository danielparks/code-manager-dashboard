#!/bin/bash

set -e

workdir="$(mktemp -d)"

lastOut=""
for input in "$@" ; do
  out="${workdir}/$(basename $input .json).out"
  go run *.go getfile -S -f "${workdir}/state.json" "$input" >"$out"

  if [ -n "$lastOut" ] ; then
    diff -us "$lastOut" "$out" || true
    echo
  fi

  lastOut="$out"
done
