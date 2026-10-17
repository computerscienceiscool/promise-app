package org.promisegrid.promiseapp

import android.content.Context

class RelayConfig(context: Context) {
    private val prefs = context.getSharedPreferences("promise-relay", Context.MODE_PRIVATE)
    var url: String
        get() = prefs.getString("url", "") ?: ""
        set(value) { prefs.edit().putString("url", value.trim()).apply() }
}
