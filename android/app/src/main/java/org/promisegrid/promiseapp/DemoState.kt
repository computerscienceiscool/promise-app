package org.promisegrid.promiseapp

import android.content.Context
import java.util.UUID

class DemoState(context: Context) {
    private val prefs = context.getSharedPreferences("promise-demo", Context.MODE_PRIVATE)
    val installID: String = prefs.getString("install_id", null) ?: UUID.randomUUID().toString().also { prefs.edit().putString("install_id", it).apply() }
    fun queue(body: String) { prefs.edit().putString("outbox", body).apply() }
    fun queued(): String? = prefs.getString("outbox", null)
}
