#!/usr/bin/env bash
# Validate the host-side prerequisites for the phone-to-computer POC.
#
# This is intentionally a read-only readiness check except for creating the
# computer peer identity when it does not exist.  It neither starts a listener
# nor sends, accepts, or executes a promise.
set -euo pipefail

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
project_dir=$(CDPATH= cd -- "$script_dir/.." && pwd)
backend_dir="$project_dir/backend"
go_cache=${PROMISE_GO_CACHE:-"$project_dir/.local/go-build-cache"}

usage() {
  cat <<'EOF'
Usage:
  ./scripts/validate-phone-computer-host.sh \
    -relay http://192.168.1.10:8080 \
    [-identity-file PATH] [-apk PATH] [--check-relay]

Checks the host tools, Go protocol tests, desktop peer identity, and (when
provided) a built APK.  --check-relay additionally checks /healthz on both the
host loopback listener and the advertised LAN URL; start the relay first.

The command never prints private key material or an admin token.
EOF
}

relay=""
identity_file="$project_dir/.local/promise-peer/identity.ed25519"
apk=""
check_relay=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    -relay) relay=${2:-}; shift 2 ;;
    -identity-file) identity_file=${2:-}; shift 2 ;;
    -apk) apk=${2:-}; shift 2 ;;
    --check-relay) check_relay=true; shift ;;
    -h|--help) usage; exit 0 ;;
    *) usage >&2; exit 2 ;;
  esac
done

[[ -n "$relay" ]] || { usage >&2; exit 2; }
case "$relay" in
  http://*) ;;
  *) printf '%s\n' 'Relay URL must begin with http://; this POC relay does not provide TLS.' >&2; exit 2 ;;
esac
case "$relay" in
  *localhost*|*127.0.0.1*|*0.0.0.0*)
    printf '%s\n' 'Relay URL must name the computer LAN address, not localhost, 127.0.0.1, or 0.0.0.0.' >&2
    exit 2
    ;;
esac

for required in go curl; do
  command -v "$required" >/dev/null || {
    printf 'Required host command is unavailable: %s\n' "$required" >&2
    exit 127
  }
done
mkdir -p "$go_cache"

printf '%s\n' '1/5 Running backend protocol and relay tests...'
(cd "$backend_dir" && GOCACHE="$go_cache" go test ./...)

printf '%s\n' '2/5 Compiling host relay, peer, and proof tools...'
(cd "$backend_dir" && GOCACHE="$go_cache" go build ./cmd/promise-server ./cmd/promise-peer ./cmd/proof-check)

printf '%s\n' '3/5 Loading computer peer identity...'
peer_cid=$(cd "$backend_dir" && GOCACHE="$go_cache" go run ./cmd/promise-peer identity -identity-file "$identity_file")
[[ "$peer_cid" =~ ^b[a-z2-7]+$ ]] || {
  printf '%s\n' 'Computer peer returned an invalid lowercase-base32 public CID.' >&2
  exit 1
}

if [[ -n "$apk" ]]; then
  printf '%s\n' '4/5 Checking supplied Android APK...'
  [[ -f "$apk" && -s "$apk" ]] || { printf 'APK is missing or empty: %s\n' "$apk" >&2; exit 1; }
  command -v unzip >/dev/null || { printf '%s\n' 'unzip is required to validate an APK.' >&2; exit 127; }
  unzip -tqq "$apk" >/dev/null
  unzip -Z1 "$apk" | grep -qx 'AndroidManifest.xml' || {
    printf 'APK has no AndroidManifest.xml: %s\n' "$apk" >&2
    exit 1
  }
else
  printf '%s\n' '4/5 APK check skipped (pass -apk after the host build).'
fi

if "$check_relay"; then
  printf '%s\n' '5/5 Checking running relay health on host and advertised URL...'
  authority=${relay#http://}
  authority=${authority%%/*}
  port=80
  if [[ "$authority" == *:* ]]; then
    port=${authority##*:}
  fi
  [[ "$port" =~ ^[0-9]+$ ]] || { printf 'Could not determine port from relay URL: %s\n' "$relay" >&2; exit 2; }
  for endpoint in "http://127.0.0.1:$port/healthz" "${relay%/}/healthz"; do
    status=$(curl --fail --silent --show-error --max-time 5 "$endpoint")
    [[ "$status" == '{"status":"ok"}' ]] || { printf 'Unexpected health response from %s: %s\n' "$endpoint" "$status" >&2; exit 1; }
  done
else
  printf '%s\n' '5/5 Running relay health check skipped (pass --check-relay after it starts).'
fi

printf '\n%s\n' 'Host validation passed.'
printf 'Relay URL:         %s\n' "$relay"
printf 'Computer peer CID: %s\n' "$peer_cid"
printf 'Identity file:     %s\n' "$identity_file"
printf '%s\n' 'Next physical step: enter the relay URL and computer peer CID in Promise App.'
