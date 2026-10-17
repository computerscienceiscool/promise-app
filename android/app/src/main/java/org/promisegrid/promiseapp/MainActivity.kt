package org.promisegrid.promiseapp

import android.app.Activity
import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import android.util.Base64
import android.os.Bundle
import android.widget.Button
import android.widget.EditText
import android.widget.LinearLayout
import android.widget.ScrollView
import android.widget.TextView
import java.util.concurrent.Executors

/** Native POC UI: it persists signed bytes before attempting network delivery. */
class MainActivity : Activity() {
    private val work = Executors.newSingleThreadExecutor()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        val signingIdentity = SigningIdentity(this)
        val identity = PromiseEnvelope.identityCid(signingIdentity.keyPair.public).toString()
        val relayConfig = RelayConfig(this)
        val outbox = DurableOutbox(this).also { it.recoverInterruptedUploads() }
        val inbox = InboxStore(this)
        val uploader = OutboxUploader(outbox, { relayConfig.url })

        val content = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(48, 48, 48, 48)
        }
        fun label(value: String, size: Float = 16f) = TextView(this).apply {
            text = value; textSize = size; content.addView(this)
        }
        fun button(title: String, click: () -> Unit) = Button(this).apply {
            text = title; setOnClickListener { click() }; content.addView(this)
        }

        label("Promise App", 24f)
        label("Device identity (share this with a sender):", 14f)
        label(identity, 12f)
        val status = label("Ready.")
        button("Copy device identity") {
            val clipboard = getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
            clipboard.setPrimaryClip(ClipData.newPlainText("Promise App device identity", identity))
            status.text = "Device identity copied. Paste it into the sender's Recipient field."
        }
        label("Signing key: ${Base64.encodeToString(signingIdentity.keyPair.public.encoded, Base64.NO_WRAP).take(24)}…", 12f)
        val relayURL = EditText(this).apply {
            hint = "Relay URL (for example http://192.168.1.10:8080)"
            setText(relayConfig.url); content.addView(this)
        }
        val activeRelay = label("Active relay: ${relayConfig.url.ifBlank { "not saved" }}", 12f)
        fun saveRelay(url: String) {
            val trimmed = url.trim()
            relayURL.setText(trimmed)
            relayConfig.url = trimmed
            activeRelay.text = "Active relay: ${trimmed.ifBlank { "not saved" }}"
            status.text = if (trimmed.isBlank()) "Relay URL cleared." else "Relay saved: $trimmed"
        }
        button("Use USB relay") {
            saveRelay("http://127.0.0.1:8080")
        }
        button("Paste relay URL") {
            val clipboard = getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
            val pasted = clipboard.primaryClip
                ?.takeIf { it.itemCount > 0 }
                ?.getItemAt(0)
                ?.coerceToText(this)
                ?.toString()
                ?.trim()
            if (pasted.isNullOrBlank()) {
                status.text = "Clipboard has no relay URL."
            } else if (!pasted.startsWith("http://") && !pasted.startsWith("https://")) {
                status.text = "Clipboard is not a relay URL."
            } else {
                saveRelay(pasted)
            }
        }
        val recipient = EditText(this).apply {
            hint = "Recipient device identity CID"; content.addView(this)
        }
        button("Paste recipient identity") {
            val clipboard = getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
            val clip = clipboard.primaryClip
            val pasted = clip?.takeIf { it.itemCount > 0 }?.getItemAt(0)?.coerceToText(this)?.toString()?.trim()
            if (pasted.isNullOrBlank()) {
                status.text = "Clipboard has no recipient identity."
            } else {
                recipient.setText(pasted)
                status.text = "Recipient identity pasted."
            }
        }
        val message = EditText(this).apply {
            hint = "Write a promise"; minLines = 3; content.addView(this)
        }
        val outboxSummary = label("")
        val inboxSummary = label("")

        fun refresh() {
            val outboxItems = outbox.all()
            outboxSummary.text = if (outboxItems.isEmpty()) "Outbox: empty" else {
                "Outbox (${outboxItems.size}):\n" + outboxItems.joinToString("\n") {
                    "${PromiseEnvelope.envelopeCid(it.envelope)}\n" +
                        "  ${it.status}  attempts=${it.attempts}" +
                        (it.lastError?.let { error -> "  $error" } ?: "")
                }
            }
            val inboxItems = inbox.all()
            inboxSummary.text = if (inboxItems.isEmpty()) "Inbox: empty" else {
                "Inbox (${inboxItems.size}; local states):\n" + inboxItems.joinToString("\n") {
                    "${it.cid}\n  ${it.state}: ${it.detail}"
                }
            }
        }

        button("Save relay") {
            saveRelay(relayURL.text.toString())
        }
        button("Sign, queue, and upload") {
            try {
                val envelope = PromiseEnvelope.offer(
                    signer = signingIdentity,
                    recipient = PromiseEnvelope.Cid.fromString(recipient.text.toString().trim()),
                    body = message.text.toString().trim(),
                )
                val item = outbox.enqueue(envelope.bytes, envelope.from.toString())
                val messageCid = envelope.cid.toString()
                message.setText("")
                status.text = "Signed $messageCid; queued exact bytes. Uploading…"
                refresh()
                uploader.upload(item.id) { completed -> runOnUiThread {
                    status.text = completed?.let { "Upload ${it.status}: $messageCid" }
                        ?: "Upload did not find its outbox item."
                    refresh()
                } }
            } catch (error: Exception) {
                status.text = "Cannot queue message: ${error.message ?: "invalid input"}"
            }
        }
        button("Retry pending uploads") {
            status.text = "Retrying persisted uploads…"
            uploader.retryAll { completed -> runOnUiThread {
                status.text = completed?.let { "Upload ${it.status}: ${it.id}" } ?: "No pending upload."
                refresh()
            } }
        }
        button("Refresh inbox") {
            val url = relayConfig.url
            if (url.isBlank()) {
                status.text = "Save a relay URL first."
                return@button
            }
            status.text = "Checking relay mailbox; received bytes stay quarantined."
            work.execute {
                val result = InboxSync(RelayClient(url, identity), inbox, url).sync()
                runOnUiThread {
                    status.text = result.fold(
                        onSuccess = { "Inbox sync: ${it.fetched} fetched, ${it.alreadyRetained} already retained, ${it.failures.size} failed." },
                        onFailure = { "Inbox sync failed: ${it.message ?: "unknown error"}" },
                    )
                    refresh()
                }
            }
        }
        refresh()
        setContentView(ScrollView(this).apply { addView(content) })
    }

    override fun onDestroy() {
        work.shutdown()
        super.onDestroy()
    }
}
