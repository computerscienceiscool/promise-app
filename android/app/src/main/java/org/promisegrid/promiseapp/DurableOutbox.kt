package org.promisegrid.promiseapp

import android.content.Context
import android.util.Base64
import org.json.JSONObject
import java.security.MessageDigest
import java.util.concurrent.Executor
import java.util.concurrent.Executors

/**
 * Durable storage for exact signed envelope bytes. Once queued, bytes are never
 * regenerated for a retry: the relay always receives the original signature.
 *
 * SharedPreferences is sufficient for this deliberately small POC. A production
 * app should migrate these records to an encrypted transactional database.
 */
class DurableOutbox(context: Context) {
    enum class Status { PENDING, UPLOADING, RETRYABLE_FAILURE, UPLOADED }

    data class Item(
        val id: String,
        val envelope: ByteArray,
        val senderID: String,
        val status: Status,
        val attempts: Int,
        val lastError: String?,
        val relayResponse: String?,
    )

    private val prefs = context.getSharedPreferences(PREFERENCES, Context.MODE_PRIVATE)

    /** Persist exact bytes before any upload is attempted. */
    fun enqueue(envelope: ByteArray, senderID: String): Item {
        require(envelope.isNotEmpty()) { "Envelope must not be empty" }
        require(senderID.isNotBlank()) { "Sender identity is required" }
        val id = envelopeID(envelope)
        val existing = load(id)
        if (existing != null) return existing
        val item = Item(id, envelope.copyOf(), senderID, Status.PENDING, 0, null, null)
        save(item)
        return item
    }

    fun get(id: String): Item? = load(id)

    fun all(): List<Item> = ids().mapNotNull(::load).sortedBy { it.id }

    fun retryable(): List<Item> = all().filter { it.status == Status.PENDING || it.status == Status.RETRYABLE_FAILURE }

    /** Recover safely after an app process dies during an HTTP request. */
    fun recoverInterruptedUploads() {
        all().filter { it.status == Status.UPLOADING }.forEach {
            save(it.copy(status = Status.RETRYABLE_FAILURE, lastError = "Upload interrupted; retry queued"))
        }
    }

    internal fun markUploading(id: String): Item? = transition(id, setOf(Status.PENDING, Status.RETRYABLE_FAILURE)) {
        it.copy(status = Status.UPLOADING, attempts = it.attempts + 1, lastError = null)
    }

    internal fun markUploaded(id: String, response: String): Item? = transition(id, setOf(Status.UPLOADING)) {
        it.copy(status = Status.UPLOADED, relayResponse = response, lastError = null)
    }

    internal fun markFailed(id: String, error: String): Item? = transition(id, setOf(Status.UPLOADING)) {
        it.copy(status = Status.RETRYABLE_FAILURE, lastError = error.take(MAX_ERROR_CHARS))
    }

    private fun transition(id: String, allowed: Set<Status>, update: (Item) -> Item): Item? = synchronized(this) {
        val current = load(id) ?: return@synchronized null
        if (current.status !in allowed) return@synchronized current
        update(current).also(::save)
    }

    private fun ids(): Set<String> = prefs.getStringSet(IDS, emptySet())?.toSet().orEmpty()

    private fun load(id: String): Item? {
        val record = prefs.getString(key(id), null) ?: return null
        return runCatching {
            val json = JSONObject(record)
            Item(
                id = id,
                envelope = Base64.decode(json.getString("envelope"), Base64.NO_WRAP),
                senderID = json.getString("senderID"),
                status = Status.valueOf(json.getString("status")),
                attempts = json.getInt("attempts"),
                lastError = json.optString("lastError").ifBlank { null },
                relayResponse = json.optString("relayResponse").ifBlank { null },
            )
        }.getOrNull()
    }

    private fun save(item: Item) {
        val record = JSONObject()
            .put("envelope", Base64.encodeToString(item.envelope, Base64.NO_WRAP))
            .put("senderID", item.senderID)
            .put("status", item.status.name)
            .put("attempts", item.attempts)
            .put("lastError", item.lastError ?: "")
            .put("relayResponse", item.relayResponse ?: "")
            .toString()
        prefs.edit().putString(key(item.id), record).putStringSet(IDS, ids() + item.id).commit()
    }

    private fun envelopeID(envelope: ByteArray): String = MessageDigest.getInstance("SHA-256")
        .digest(envelope)
        .joinToString("") { "%02x".format(it) }

    private companion object {
        const val PREFERENCES = "promise-outbox-v1"
        const val IDS = "ids"
        const val MAX_ERROR_CHARS = 500
        fun key(id: String) = "item:$id"
    }
}

/** Executes one persisted upload at a time and reports the durable result. */
class OutboxUploader(
    private val outbox: DurableOutbox,
    private val relayURL: () -> String,
    private val executor: Executor = Executors.newSingleThreadExecutor(),
) {
    fun upload(id: String, onComplete: (DurableOutbox.Item?) -> Unit = {}) {
        executor.execute {
            val uploading = outbox.markUploading(id)
            if (uploading == null || uploading.status != DurableOutbox.Status.UPLOADING) {
                onComplete(uploading)
                return@execute
            }
            val finished = RelayClient(relayURL(), uploading.senderID).upload(uploading.envelope)
                .fold(
                    onSuccess = { outbox.markUploaded(id, it) },
                    onFailure = { outbox.markFailed(id, it.message ?: "Upload failed") },
                )
            onComplete(finished)
        }
    }

    fun retryAll(onComplete: (DurableOutbox.Item?) -> Unit = {}) {
        outbox.retryable().forEach { upload(it.id, onComplete) }
    }
}
