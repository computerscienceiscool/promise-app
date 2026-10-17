package org.promisegrid.promiseapp

import android.content.Context

class RelayConfig(context: Context) {
    private val prefs = context.getSharedPreferences("promise-relay", Context.MODE_PRIVATE)
    var url: String
        get() = prefs.getString("url", "") ?: ""
        set(value) {
            // A trailing slash is harmless in a browser but creates doubled paths in
            // the relay client. Store one canonical base URL for repeatable demos.
            prefs.edit().putString("url", value.trim().trimEnd('/')).apply()
        }
}
