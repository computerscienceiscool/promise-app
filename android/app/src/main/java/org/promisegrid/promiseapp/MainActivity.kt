package org.promisegrid.promiseapp

import android.app.Activity
import android.os.Bundle
import android.widget.*
import android.util.Base64

class MainActivity : Activity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        val state = DemoState(this)
        val signingIdentity = SigningIdentity(this)
        val relay = RelayConfig(this)
        val layout = LinearLayout(this).apply { orientation = LinearLayout.VERTICAL; setPadding(48, 48, 48, 48) }
        fun line(value: String) = TextView(this).apply { text = value; textSize = 17f; layout.addView(this) }
        line("Promise App")
        line("Local install ID: ${state.installID}")
        line("Signing key: ${Base64.encodeToString(signingIdentity.keyPair.public.encoded, Base64.NO_WRAP).take(24)}…")
        val relayURL = EditText(this).apply { hint = "Relay URL"; setText(relay.url); layout.addView(this) }
        layout.addView(Button(this).apply { text = "Save relay"; setOnClickListener { relay.url = relayURL.text.toString() } })
        val message = EditText(this).apply { hint = "Write a promise"; layout.addView(this) }
        val outbox = line("Outbox: ${state.queued() ?: "empty"}")
        layout.addView(Button(this).apply { text = "Queue message"; setOnClickListener {
            val body = message.text.toString().trim()
            if (body.isNotEmpty()) { state.queue(body); outbox.text = "Outbox: $body"; message.setText("") }
        } })
        setContentView(layout)
    }
}
