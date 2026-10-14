#!/usr/bin/env bash
# Measure the DevEx number that matters most to a new engineer:
# fresh clone -> first passing test. Run from the repo root after `make bootstrap`.
#
# Cold cache, like fleet-metrics.sh: GOCACHE points at a throwaway dir so the
# number reflects what a NEW engineer experiences, not your warm build cache.
set -euo pipefail

cache=$(mktemp -d)
trap 'rm -rf "$cache"' EXIT

# Nanosecond clock with a fallback for shells whose date(1) ignores %N
# (e.g. BusyBox): 19-digit output means real nanoseconds, otherwise use seconds.
now_ns() {
  local t
  t=$(date +%s%N)
  if [ "${#t}" -ge 19 ]; then
    echo "$t"
  else
    echo $(( $(date +%s) * 1000000000 ))
  fi
}

start=$(now_ns)
GOCACHE="$cache" go build ./...
GOCACHE="$cache" go test ./...
end=$(now_ns)

awk -v s="$start" -v e="$end" 'BEGIN { printf "setup time (fresh clone -> first passing test): %.1fs\n", (e - s) / 1e9 }'
