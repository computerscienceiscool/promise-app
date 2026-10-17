#!/usr/bin/env bash
# Performs the non-secret, non-listening checks required before starting the
# private-LAN POC relay. It never creates relay state or starts a service.
set -euo pipefail

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
project_dir=$(CDPATH= cd -- "$script_dir/.." && pwd)
backend_dir="$project_dir/backend"
listen_addr=${PROMISE_LISTEN_ADDR:-"0.0.0.0:8080"}

usage() {
  cat <<'EOF'
Usage:
  ./scripts/preflight-lan-relay.sh -relay http://192.168.1.10:8080

Optional environment:
  PROMISE_LISTEN_ADDR  Listener that will be used later (default 0.0.0.0:8080).

This is a read-only preflight. It does not require or print PROMISE_ADMIN_TOKEN,
create a data directory, start a relay, or make an HTTP request.
EOF
}

relay=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    -relay) relay=${2:-}; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) usage >&2; exit 2 ;;
  esac
done

[[ -n "$relay" ]] || { usage >&2; exit 2; }
case "$relay" in
  http://*) ;;
  *) printf '%s\n' 'Relay URL must begin with http:// for this POC relay.' >&2; exit 2 ;;
esac
case "$relay" in
  *localhost*|*127.0.0.1*|*0.0.0.0*)
    printf '%s\n' 'Relay URL must name this computer’s LAN address, not a loopback/wildcard address.' >&2
    exit 2
    ;;
esac
case "$listen_addr" in
  0.0.0.0:*) ;;
  *) printf '%s\n' 'PROMISE_LISTEN_ADDR must be 0.0.0.0:<port> for this LAN POC.' >&2; exit 2 ;;
esac

authority=${relay#http://}
authority=${authority%%/*}
host=${authority%:*}
relay_port=${authority##*:}
listen_port=${listen_addr##*:}
[[ "$relay_port" =~ ^[0-9]+$ && "$relay_port" -ge 1 && "$relay_port" -le 65535 ]] || {
  printf 'Relay URL has an invalid port: %s\n' "$relay" >&2; exit 2;
}
[[ "$listen_port" =~ ^[0-9]+$ && "$listen_port" -ge 1 && "$listen_port" -le 65535 ]] || {
  printf 'PROMISE_LISTEN_ADDR has an invalid port: %s\n' "$listen_addr" >&2; exit 2;
}
[[ "$relay_port" == "$listen_port" ]] || {
  printf 'Relay URL port (%s) does not match PROMISE_LISTEN_ADDR port (%s).\n' "$relay_port" "$listen_port" >&2
  exit 2
}

for required in go ss; do
  command -v "$required" >/dev/null || {
    printf 'Required host command is unavailable: %s\n' "$required" >&2
    exit 127
  }
done

if [[ "$host" =~ ^[0-9]{1,3}(\.[0-9]{1,3}){3}$ ]] && command -v ip >/dev/null; then
  addresses=$(ip -o -4 addr show 2>/dev/null | awk '{print $4}' | cut -d/ -f1 || true)
  if [[ -z "$addresses" ]]; then
    printf '%s\n' 'Warning: unable to inspect local network addresses; confirm the advertised LAN IP manually.' >&2
  elif ! grep -Fx -- "$host" <<<"$addresses" >/dev/null; then
    printf 'Relay URL host is not assigned to this computer: %s\n' "$host" >&2
    exit 1
  fi
fi

# A currently listening port would make the subsequent launcher ambiguous. Do
# not attempt to bind it here: checking must not change host state.
listeners=$(ss -ltnH "sport = :$listen_port" 2>/dev/null || true)
if [[ -z "$listeners" ]] && ! ss -ltnH 2>/dev/null >/dev/null; then
  printf '%s\n' 'Warning: unable to inspect listening ports; confirm the selected port is unused manually.' >&2
elif [[ -n "$listeners" ]]; then
  printf 'A process is already listening on TCP port %s; choose a free port or stop it deliberately.\n' "$listen_port" >&2
  exit 1
fi

printf '%s\n' '1/3 Running backend tests...'
(cd "$backend_dir" && go test ./...)
printf '%s\n' '2/3 Compiling the relay command...'
# `go test` compiles the command without leaving a generated binary in the
# repository, which keeps this preflight read-only with respect to project
# sources and relay state.
(cd "$backend_dir" && go test ./cmd/promise-server)
printf '%s\n' '3/3 Checking LAN relay configuration...'
printf 'Advertised relay URL: %s\n' "$relay"
printf 'Future listener:      %s\n' "$listen_addr"
printf '%s\n' 'Preflight passed. To start later, provide a fresh PROMISE_ADMIN_TOKEN to scripts/start-lan-relay.sh.'
