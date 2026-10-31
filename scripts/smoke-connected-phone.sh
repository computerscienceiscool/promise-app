#!/usr/bin/env bash
# Checks a physically attached phone can reach the configured POC relay.
# It does not send, accept, or execute a promise.
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: ./scripts/smoke-connected-phone.sh -relay URL [--usb]

Checks that one ADB-connected phone has Promise App installed and can reach
RELAY/healthz itself. --usb additionally verifies the ADB reverse route and
the phone-local http://127.0.0.1:8080 health endpoint.
EOF
}

relay=""
check_usb=false
while [[ $# -gt 0 ]]; do
  case "$1" in
    -relay) relay=${2:-}; shift 2 ;;
    --usb) check_usb=true; shift ;;
    -h|--help) usage; exit 0 ;;
    *) usage >&2; exit 2 ;;
  esac
done

[[ "$relay" =~ ^http://[^/]+(:[0-9]+)?$ ]] || {
  printf '%s\n' 'Set -relay to a private-LAN base URL such as http://192.168.1.10:8080.' >&2
  exit 2
}

command -v adb >/dev/null || { printf '%s\n' 'adb is required.' >&2; exit 127; }
adb get-state | grep -qx device || { printf '%s\n' 'Connect and authorize one Android phone first.' >&2; exit 1; }
adb shell pm path org.promisegrid.promiseapp | grep -q '^package:' || {
  printf '%s\n' 'Promise App is not installed on the connected phone.' >&2
  exit 1
}

printf '%s\n' '1/3 Connected phone and Promise App detected.'
printf '2/3 Checking phone Wi-Fi/LAN relay health: %s/healthz\n' "$relay"
health=$(adb shell "curl --fail --silent --show-error --max-time 5 '${relay}/healthz'" | tr -d '\r')
[[ "$health" == '{"status":"ok"}' ]] || {
  printf 'Unexpected phone health response: %s\n' "$health" >&2
  exit 1
}

if "$check_usb"; then
  printf '%s\n' '3/3 Checking USB reverse relay health.'
  adb reverse --list | grep -Eq 'tcp:8080[[:space:]]+tcp:8080' || {
    printf '%s\n' 'No USB reverse route for tcp:8080; run make usb-relay first.' >&2
    exit 1
  }
  usb_health=$(adb shell "curl --fail --silent --show-error --max-time 5 http://127.0.0.1:8080/healthz" | tr -d '\r')
  [[ "$usb_health" == '{"status":"ok"}' ]] || {
    printf 'Unexpected phone USB health response: %s\n' "$usb_health" >&2
    exit 1
  }
else
  printf '%s\n' '3/3 USB route check skipped; pass --usb to test it too.'
fi

printf '%s\n' 'Connected-phone relay smoke check passed.'
