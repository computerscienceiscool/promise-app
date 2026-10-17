.DEFAULT_GOAL := help

ANDROID_DIR := android
APK := $(ANDROID_DIR)/app/build/outputs/apk/debug/app-debug.apk
BACKEND_DIR := backend
RELAY ?= $(PROMISE_ADVERTISE_URL)
DATA_DIR ?= .local/promise-relay
PEER_DIR ?= .local/promise-peer

.PHONY: help test host-build apk install launch logcat diagnose-phone usb-relay usb-relay-clear peer-id peer-qr peer-qr-open relay-qr-open relay-preflight relay validate evidence clean-apk

help:
	@printf '%s\n' \
	  'Promise App common commands:' \
	  '  make test                         Run backend, protocol, and peer tests' \
	  '  make host-build                   Test host tools and build the debug APK' \
	  '  make apk                          Build the Android debug APK' \
	  '  make install                      Install the built APK using adb' \
	  '  make launch                       Start Promise App on the connected phone' \
	  '  make logcat                       Show Promise App crash/error logs' \
	  '  make diagnose-phone               Clear logs, launch app, then show crash/errors' \
	  '  make usb-relay                    Route phone localhost:8080 to this computer over USB' \
	  '  make usb-relay-clear              Remove the USB relay route' \
	  '  make peer-id                      Print the persistent computer recipient CID' \
	  '  make peer-qr                      Create a scannable PNG for the recipient CID' \
	  '  make peer-qr-open                 Create and show the recipient QR image' \
	  '  make relay-qr-open RELAY=URL       Create and show a phone relay-health QR image' \
	  '  make relay-preflight RELAY=URL    Validate a planned LAN relay URL' \
	  '  make relay RELAY=URL               Start relay; requires PROMISE_ADMIN_TOKEN' \
	  '  make validate APK=PATH             Validate host tools and an APK' \
	  '  make evidence RELAY=URL CID=CID    Capture proof; also requires token/data/recipient/out-dir'

test:
	cd $(BACKEND_DIR) && go test ./...

host-build:
	./scripts/build-phone-computer-host.sh --offline

apk:
	cd $(ANDROID_DIR) && ./gradlew :app:assembleDebug

install: apk
	adb install -r $(APK)

launch:
	adb shell am start -n org.promisegrid.promiseapp/.MainActivity

logcat:
	adb logcat -d -v brief -s AndroidRuntime:E ActivityTaskManager:E promisegrid:I '*:S'

diagnose-phone:
	adb logcat -c
	adb shell am force-stop org.promisegrid.promiseapp
	adb shell am start -n org.promisegrid.promiseapp/.MainActivity
	sleep 2
	$(MAKE) logcat

# Developer/demo fallback when a Wi-Fi network isolates clients. Keep the
# relay running on the computer, then point the phone app at
# http://127.0.0.1:8080 after this command succeeds.
usb-relay:
	adb reverse tcp:8080 tcp:8080
	@echo 'USB relay is ready. In Promise App, set Relay URL to http://127.0.0.1:8080 and tap Save relay.'

usb-relay-clear:
	adb reverse --remove tcp:8080

peer-id:
	cd $(BACKEND_DIR) && go run ./cmd/promise-peer identity -identity-file ../$(PEER_DIR)/identity.ed25519

peer-qr:
	cd $(BACKEND_DIR) && go run ./cmd/promise-peer qr -identity-file ../$(PEER_DIR)/identity.ed25519 -out ../$(PEER_DIR)/recipient-qr.png

peer-qr-open: peer-qr
	xdg-open $(PEER_DIR)/recipient-qr.png

relay-qr-open:
	@test -n "$(RELAY)" || (echo 'Set RELAY=http://LAN-IP:8080' >&2; exit 2)
	cd $(BACKEND_DIR) && go run ./cmd/promise-peer qr -content "$(RELAY)/healthz" -out ../$(PEER_DIR)/relay-health-qr.png
	xdg-open $(PEER_DIR)/relay-health-qr.png

relay-preflight:
	@test -n "$(RELAY)" || (echo 'Set RELAY=http://LAN-IP:8080' >&2; exit 2)
	./scripts/preflight-lan-relay.sh -relay "$(RELAY)"

relay:
	@test -n "$(RELAY)" || (echo 'Set RELAY=http://LAN-IP:8080 or PROMISE_ADVERTISE_URL=http://LAN-IP:8080' >&2; exit 2)
	@test -n "$(PROMISE_ADMIN_TOKEN)" || (echo 'Set PROMISE_ADMIN_TOKEN to a fresh secret' >&2; exit 2)
	PROMISE_ADVERTISE_URL="$(RELAY)" PROMISE_DATA_DIR="$(DATA_DIR)" ./scripts/start-lan-relay.sh

validate:
	@test -n "$(APK)" || (echo 'Set APK=path/to/app-debug.apk' >&2; exit 2)
	./scripts/validate-phone-computer-host.sh -apk "$(APK)"

evidence:
	@test -n "$(RELAY)" && test -n "$(CID)" && test -n "$(RECIPIENT)" && test -n "$(OUT_DIR)" && test -n "$(PROMISE_ADMIN_TOKEN)" || (echo 'Set RELAY, CID, RECIPIENT, OUT_DIR, and PROMISE_ADMIN_TOKEN' >&2; exit 2)
	./scripts/capture-phone-computer-evidence.sh -relay "$(RELAY)" -token "$(PROMISE_ADMIN_TOKEN)" -data-dir "$(DATA_DIR)" -cid "$(CID)" -recipient "$(RECIPIENT)" -out-dir "$(OUT_DIR)"

clean-apk:
	rm -f $(APK)
