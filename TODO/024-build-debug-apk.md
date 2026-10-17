# TODO 024 - Build the Integrated Debug APK on the Host

Run the Android build outside the sandbox and retain the successful output.

## Host command

```bash
cd ~/lab/cswg/apps/promise-app/android
./gradlew :app:assembleDebug --offline
```

Expected artifact on success:

```text
app/build/outputs/apk/debug/app-debug.apk
```

## Validation record — 2026-10-17

Attempted from the coding sandbox with the cached Gradle 8.14 distribution copied
to a writable temporary Gradle home:

```bash
GRADLE_USER_HOME=/tmp/promise-app-gradle ./gradlew :app:assembleDebug --offline
```

The wrapper and cached distribution were found, but Gradle cannot start inside
this sandbox because it cannot select a wildcard network interface for Gradle's
file-lock contention service:

```text
Could not create service of type FileLockContentionHandler using
BasicGlobalScopeServices.createFileLockContentionHandler().
Could not determine a usable wildcard IP for this machine.
```

This is an execution-environment restriction, not an APK compilation result.
Run the host command above (outside the sandbox) to complete TODO 024 and retain
its successful output.

## Completion record — 2026-10-17

The authorized host terminal built the latest source and installed the produced
debug APK on the connected phone:

```text
BUILD SUCCESSFUL in 7s
35 actionable tasks: 4 executed, 31 up-to-date
Performing Streamed Install
Success
```
