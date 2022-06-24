#!/bin/sh

set -ue

fmt_list="$(gofmt -l "$@")"

if [ -n "$fmt_list" ]; then
  echo "Check gofmt failed: " >&2

  for file in "$fmt_list"; do
    echo "::error file=${file},title=gofmt::gofmt check failed"
    echo "\t$file" >&2
  done
  exit 1
fi
