# Android client

Android/Kotlin foundation for phone and Kindle Fire tablet layouts. The first
screen deliberately reports unimplemented protocol state; it does not pretend
that a backend connection or unsigned UI action is a PromiseGrid message.

On this computer, create `local.properties` with the SDK path from
`local.properties.example`, then run:

```sh
./gradlew :app:assembleDebug
adb install -r app/build/outputs/apk/debug/app-debug.apk
```
