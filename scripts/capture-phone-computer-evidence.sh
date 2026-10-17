#!/usr/bin/env bash
# Capture private, operator-owned proof artifacts for the phone-to-computer POC.
# It does not create identity keys, send messages, or infer that a recipient
# accepted a promise.  Run it after the phone upload and computer inbox fetch.
set -euo pipefail

usage() {
  echo "usage: $0 -relay URL -token ADMIN_TOKEN -data-dir DIR -cid MESSAGE_CID -recipient PEER_CID -out-dir DIR [-peer-identity-file PATH] [-peer-inbox-dir PATH]" >&2
  exit 2
}

relay=""
token=""
data_dir=""
message_cid=""
recipient=""
out_dir=""
identity_file=".promise-peer/identity.ed25519"
inbox_dir=".promise-peer/inbox"

while [[ $# -gt 0 ]]; do
  case "$1" in
    -relay) relay="$2"; shift 2 ;;
    -token) token="$2"; shift 2 ;;
    -data-dir) data_dir="$2"; shift 2 ;;
    -cid) message_cid="$2"; shift 2 ;;
    -recipient) recipient="$2"; shift 2 ;;
    -out-dir) out_dir="$2"; shift 2 ;;
    -peer-identity-file) identity_file="$2"; shift 2 ;;
    -peer-inbox-dir) inbox_dir="$2"; shift 2 ;;
    *) usage ;;
  esac
done

[[ -n "$relay" && -n "$token" && -n "$data_dir" && -n "$message_cid" && -n "$recipient" && -n "$out_dir" ]] || usage
mkdir -p "$out_dir"
chmod 700 "$out_dir"

backend_dir="$(cd "$(dirname "$0")/../backend" && pwd)"
relay="${relay%/}"

# The peer independently fetches, validates and retains exact CBOR bytes.
(cd "$backend_dir" && go run ./cmd/promise-peer inbox \
  -relay "$relay" \
  -identity-file "$identity_file" \
  -inbox-dir "$inbox_dir" \
  -expect-cid "$message_cid" \
  -evidence-file "$out_dir/computer-fetch.json") | tee "$out_dir/computer-fetch.ndjson"

# The local checker binds immutable bytes, exactly-one mailbox entry, and
# upload/fetch audit events for the named CID.
(cd "$backend_dir" && go run ./cmd/proof-check \
  -data-dir "$data_dir" \
  -cid "$message_cid" \
  -recipient "$recipient") | tee "$out_dir/proof-check.txt"

# Keep the raw, authenticated relay audit response as an operator artifact.
curl --fail --silent --show-error \
  -H "Authorization: Bearer $token" \
  "$relay/v1/admin/audit" > "$out_dir/relay-audit.json"
find "$out_dir" -maxdepth 1 -type f -exec chmod 600 {} +

echo "Evidence retained in $out_dir"
