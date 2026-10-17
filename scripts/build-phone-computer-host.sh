#!/usr/bin/env bash
# Build every host-side artifact needed before the physical phone-computer demo.
#
# This does not contact a phone, start the relay, or send an envelope.  It leaves
# the APK at android/app/build/outputs/apk/debug/app-debug.apk for the explicit
# install step.
set -euo pipefail

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
project_dir=$(CDPATH= cd -- "$script_dir/.." && pwd)
android_dir="$project_dir/android"
backend_dir="$project_dir/backend"
go_cache=${PROMISE_GO_CACHE:-"$project_dir/.local/go-build-cache"}

usage() {
  cat <<'EOF'
Usage: ./scripts/build-phone-computer-host.sh [--offline]

Builds the Android debug APK and compiles/tests the host relay, peer, and proof
tools.  --offline passes Gradle's offline flag, suitable when dependencies are
already cached.  Set PROMISE_GO_CACHE to select a writable Go build cache.
EOF
}

gradle_args=(":app:assembleDebug")
case "${1:-}" in
  "") ;;
  --offline) gradle_args+=(--offline) ;;
  -h|--help) usage; exit 0 ;;
  *) usage >&2; exit 2 ;;
esac

command -v go >/dev/null || { printf '%s\n' 'Go is required.' >&2; exit 127; }
command -v java >/dev/null || { printf '%s\n' 'Java is required for Gradle.' >&2; exit 127; }
[[ -x "$android_dir/gradlew" ]] || { printf 'Missing Gradle wrapper: %s\n' "$android_dir/gradlew" >&2; exit 1; }
[[ -f "$android_dir/local.properties" ]] || {
  printf '%s\n' "Missing $android_dir/local.properties. Copy local.properties.example and set sdk.dir." >&2
  exit 1
}
grep -Eq '^sdk\.dir=.+$' "$android_dir/local.properties" || {
  printf '%s\n' 'local.properties must define sdk.dir.' >&2
  exit 1
}

mkdir -p "$go_cache"

printf '%s\n' '1/3 Testing and compiling host tools...'
(cd "$backend_dir" && GOCACHE="$go_cache" go test ./... && GOCACHE="$go_cache" go build ./cmd/promise-server ./cmd/promise-peer ./cmd/proof-check)

printf '%s\n' '2/3 Building Android debug APK...'
(cd "$android_dir" && ./gradlew "${gradle_args[@]}")

apk="$android_dir/app/build/outputs/apk/debug/app-debug.apk"
[[ -s "$apk" ]] || { printf 'Gradle completed but APK is missing: %s\n' "$apk" >&2; exit 1; }

printf '%s\n' '3/3 Host build passed.'
printf 'APK: %s\n' "$apk"
printf '%s\n' 'Next: install this APK explicitly, then run validate-phone-computer-host.sh with -apk.'
