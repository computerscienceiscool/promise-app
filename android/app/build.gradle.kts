plugins {
    id("com.android.application")
    id("org.jetbrains.kotlin.android")
}

dependencies {
    // System-provided Google scanner UI keeps camera handling out of this POC
    // while letting a person scan a recipient CID directly into the app.
    implementation("com.google.android.gms:play-services-code-scanner:16.1.0")
    testImplementation("junit:junit:4.13.2")
}

android {
    namespace = "org.promisegrid.promiseapp"
    compileSdk = 36
    defaultConfig { applicationId = "org.promisegrid.promiseapp"; minSdk = 26; targetSdk = 36; versionCode = 1; versionName = "0.1.0" }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
}

kotlin { jvmToolchain(17) }
