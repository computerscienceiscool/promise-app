#!/usr/bin/env bash
# Starts the POC relay for a private-LAN, two-device demonstration.
#
# This script deliberately does not generate or print the admin token. Supply a
# fresh token through PROMISE_ADMIN_TOKEN, keep it on the host, and never enter
# it in the Android app.
set -euo pipefail

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
project_dir=$(CDPATH= cd -- "$script_dir/.." && pwd)
backend_dir="$project_dir/backend"

usage() {
  cat <<'EOF'
Usage:
  PROMISE_ADMIN_TOKEN='<long random token>' ./scripts/start-lan-relay.sh [--dry-run]

Environment:
  PROMISE_ADMIN_TOKEN     Required. Kept on the host; never shown by this script.
  PROMISE_DATA_DIR        Persistent relay data directory.
                           Default: <project>/.local/relay-data
  PROMISE_LISTEN_ADDR     Listener address. Default: 0.0.0.0:8080
  PROMISE_ADVERTISE_URL   Required device-facing URL, e.g. http://192.168.1.10:8080

The relay is for a trusted private LAN demonstration only. Do not expose its
HTTP listener or data directory to the public Internet.
EOF
}

dry_run=false
case "${1:-}" in
  "") ;;
  --dry-run) dry_run=true ;;
  -h|--help) usage; exit 0 ;;
  *) usage >&2; exit 2 ;;
esac

: "${PROMISE_ADMIN_TOKEN:?Set PROMISE_ADMIN_TOKEN to a new random admin token.}"
: "${PROMISE_ADVERTISE_URL:?Set PROMISE_ADVERTISE_URL, for example http://192.168.1.10:8080.}"

data_dir=${PROMISE_DATA_DIR:-"$project_dir/.local/relay-data"}
listen_addr=${PROMISE_LISTEN_ADDR:-"0.0.0.0:8080"}

case "$listen_addr" in
  0.0.0.0:*|\[*\]:*) ;;
  *)
    printf '%s\n' "Refusing non-LAN listener '$listen_addr'. Set PROMISE_LISTEN_ADDR to 0.0.0.0:<port>." >&2
    exit 2
    ;;
esac

case "$PROMISE_ADVERTISE_URL" in
  http://*|https://*) ;;
  *)
    printf '%s\n' 'PROMISE_ADVERTISE_URL must begin with http:// or https://.' >&2
    exit 2
    ;;
esac

case "$PROMISE_ADVERTISE_URL" in
  *127.0.0.1*|*localhost*)
    printf '%s\n' 'PROMISE_ADVERTISE_URL must name the host LAN address, not localhost.' >&2
    exit 2
    ;;
esac

printf 'Relay data directory: %s\n' "$data_dir"
printf 'Relay listener:       %s\n' "$listen_addr"
printf 'Phone relay URL:      %s\n' "$PROMISE_ADVERTISE_URL"
printf '%s\n' 'Admin token:          configured (not displayed)'

if "$dry_run"; then
  printf '%s\n' 'Dry run complete; relay was not started and no data directory was created.'
  exit 0
fi

mkdir -p "$data_dir"

command -v go >/dev/null || {
  printf '%s\n' 'Go is required to start the relay.' >&2
  exit 127
}

cd "$backend_dir"
exec env \
  PROMISE_DATA_DIR="$data_dir" \
  PROMISE_LISTEN_ADDR="$listen_addr" \
  PROMISE_ADMIN_TOKEN="$PROMISE_ADMIN_TOKEN" \
  go run ./cmd/promise-server
