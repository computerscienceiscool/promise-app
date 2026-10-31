#!/usr/bin/env bash
# Tests the deterministic, host-only Phone A/Phone B QR-card generator.
set -euo pipefail

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
project_dir=$(CDPATH= cd -- "$script_dir/.." && pwd)
generator="$script_dir/prepare-phone-to-phone-qr-demo.sh"
test_output=$(mktemp -d "${TMPDIR:-/tmp}/promise-phone-qr.XXXXXX")
phone_b_cid="bagaybqabciqgtv226wkotkpxy3limnarnfs23ypn6konmtzafcwzs2g3hix3otq"

cleanup() {
  rm -rf "$test_output"
}
trap cleanup EXIT

expect_failure() {
  if "$@" >/dev/null 2>&1; then
    printf 'expected failure: %s\n' "$*" >&2
    exit 1
  fi
}

expect_failure env RELAY=http://127.0.0.1:18081 PHONE_B_CID="$phone_b_cid" "$generator"
expect_failure env RELAY=https://relay.example PHONE_B_CID=not-a-cid "$generator"

PROMISE_PHONE_QR_OUTPUT_DIR="$test_output" \
  RELAY=https://relay.example \
  PHONE_B_CID="$phone_b_cid" \
  "$generator" >/dev/null

for file in \
  phone-a-relay-url-qr.png \
  phone-b-relay-url-qr.png \
  phone-b-recipient-identity-qr.png \
  phone-to-phone-qr-cards.html; do
  test -s "$test_output/$file" || {
    printf 'missing generated file: %s\n' "$file" >&2
    exit 1
  }
done

grep -F "https://relay.example" "$test_output/phone-to-phone-qr-cards.html" >/dev/null
grep -F "Phone A → reachable relay → Phone B" "$test_output/phone-to-phone-qr-cards.html" >/dev/null
if grep -F "__RELAY_URL__" "$test_output/phone-to-phone-qr-cards.html" >/dev/null; then
  printf '%s\n' 'relay placeholder was not replaced' >&2
  exit 1
fi

printf '%s\n' 'PASS: Phone A/Phone B QR demo generator'
