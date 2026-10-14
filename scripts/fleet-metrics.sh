#!/usr/bin/env bash
# Fleet-level DevEx metrics: per-service setup time + rollup.
# This is the number a platform team reports to engineering leadership:
# onboarding cost per service, and the fleet median. Run from the repo root.
set -euo pipefail

SERVICES=(gateway billing worker)
REPORT="fleet-metrics.json"
tmp=$(mktemp -d)

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

{
  echo "["
  first=1
  for svc in "${SERVICES[@]}"; do
    # Fresh GOCACHE per service: this measures what a NEW engineer of THIS
    # service experiences on a cold clone, not the warm-cache fast path.
    start=$(now_ns)
    GOCACHE="$tmp/cache-$svc" go build -o "$tmp/$svc" "./services/$svc"
    GOCACHE="$tmp/cache-$svc" go test ./... > /dev/null
    end=$(now_ns)

    ms=$(( (end - start) / 1000000 ))
    [ "$first" -eq 0 ] && echo ","
    first=0
    printf '  {"service": "%s", "setup_ms": %d}' "$svc" "$ms"
    echo "  $svc: ${ms}ms (cold cache)" >&2
  done
  echo ""
  echo "]"
} > "$REPORT"

# Rollup: median + worst case, computed from the per-service numbers.
python3 - "$REPORT" <<'EOF' || true
import json, sys

path = sys.argv[1]
data = json.load(open(path))
ms = sorted(d["setup_ms"] for d in data)
n = len(ms)
median = ms[n // 2] if n % 2 else (ms[n // 2 - 1] + ms[n // 2]) / 2
data.append({"services": n, "median_ms": median, "worst_ms": ms[-1], "best_ms": ms[0]})
json.dump(data, open(path, "w"), indent=2)
print(f"  fleet: {n} services, median {median}ms, worst {ms[-1]}ms")
EOF

echo "report: $REPORT"
