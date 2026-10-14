#!/usr/bin/env bash
# Release path: build -> SBOM -> sign -> manifest.
# Mirrors what CI does on a tag, so releases are identical locally and in CI.
#
#   scripts/release.sh gateway 1.2.3     # release one service
#   scripts/release.sh --all 1.2.3       # release every service + fleet manifest
set -euo pipefail

SERVICES=(gateway billing worker)
VERSION="${2:-dev}"
REGISTRY="registry.example.com/paved-path"

release_one() {
  local svc="$1"
  local img="$REGISTRY/$svc:$VERSION"

  echo "==> [$svc] building image"
  docker build --build-arg SERVICE="$svc" -t "$img" .

  if command -v syft >/dev/null; then
    echo "==> [$svc] SBOM"
    syft image "$img" --output spdx-json="sbom-$svc-$VERSION.spdx.json"
  else
    echo "==> [$svc] syft not installed; skipping SBOM (brew install syft)"
  fi

  if command -v cosign >/dev/null; then
    echo "==> [$svc] signing (keyless, Sigstore)"
    cosign sign --yes "$img"
  fi

  echo "==> [$svc] manifest"
  cat > "release-$svc-$VERSION.manifest.json" <<EOF
{
  "service": "$svc",
  "version": "$VERSION",
  "image": "$img",
  "sbom": "sbom-$svc-$VERSION.spdx.json",
  "signed": $(command -v cosign >/dev/null && echo true || echo false),
  "released_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
  echo "    done: release-$svc-$VERSION.manifest.json"
}

if [ "${1:-}" = "--all" ]; then
  for svc in "${SERVICES[@]}"; do release_one "$svc"; done

  echo "==> fleet manifest"
  {
    echo "{"
    echo '  "version": "'"$VERSION"'",'
    echo '  "released_at": "'"$(date -u +%Y-%m-%dT%H:%M:%SZ)"'",'
    echo '  "services": {'
    for i in "${!SERVICES[@]}"; do
      svc="${SERVICES[$i]}"
      comma=","; [ "$i" -eq $(( ${#SERVICES[@]} - 1 )) ] && comma=""
      printf '    "%s": {"image": "%s/%s:%s", "manifest": "release-%s-%s.manifest.json"}%s\n' \
        "$svc" "$REGISTRY" "$svc" "$VERSION" "$svc" "$VERSION" "$comma"
    done
    echo "  }"
    echo "}"
  } > "fleet-$VERSION.manifest.json"
  echo "done: fleet-$VERSION.manifest.json"
else
  release_one "${1:?usage: release.sh <service|--all> [version]}"
fi
