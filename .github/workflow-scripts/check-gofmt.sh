#!/bin/sh

set -ue

fmt_list="$(gofmt -l "$@")"

if [ -n "$fmt_list" ]; then
  echo "Check gofmt failed: " >&2
  echo "${fmt_list}" | sed -e 's/^/\t/' >&2
  exit 1
fi
