package org.promisegrid.promiseapp

import android.content.Context
import org.json.JSONArray
import org.json.JSONObject
import java.io.File
import java.time.Instant

/**
 * Device-local retention for inbound relay bytes.
 *
 * A relay listing is only an availability hint. Exact bytes are retained in a
 * private file first and begin in [InboxState.QUARANTINED]. Nothing in this
 * class treats retrieved bytes as accepted PromiseGrid evidence.
 */
class InboxStore(context: Context) {
    private val root = File(context.filesDir, "inbox-quarantine").also { it.mkdirs() }
    private val prefs = context.getSharedPreferences("promise-inbox", Context.MODE_PRIVATE)

    fun retainFetched(cid: String, bytes: ByteArray, source: String, fetchedAt: Instant = Instant.now()): InboxItem {
        require(CID_PATTERN.matches(cid)) { "Invalid message CID" }
        require(bytes.isNotEmpty()) { "Envelope must not be empty" }
        require(bytes.size <= MAX_ENVELOPE_BYTES) { "Envelope exceeds POC size limit" }
        File(root, "$cid.cbor").outputStream().use { it.write(bytes) }
        val item = InboxItem(cid, InboxState.QUARANTINED, fetchedAt.toString(), source,
            "Exact relay bytes retained; verification pending")
        save(item)
        return item
    }

    fun mark(cid: String, state: InboxState, detail: String): InboxItem {
        val prior = get(cid) ?: error("No retained envelope for $cid")
        require(state != InboxState.QUARANTINED) { "Use retainFetched for quarantine" }
        return prior.copy(state = state, detail = detail.take(MAX_DETAIL_CHARS)).also(::save)
    }

    fun get(cid: String): InboxItem? = load().firstOrNull { it.cid == cid }
    fun all(): List<InboxItem> = load().sortedByDescending { it.fetchedAt }

    /** Reads exact retained bytes only; the caller must still verify them. */
    fun bytes(cid: String): ByteArray? {
        require(CID_PATTERN.matches(cid)) { "Invalid message CID" }
        val file = File(root, "$cid.cbor")
        return if (file.isFile) file.readBytes() else null
    }

    private fun save(item: InboxItem) {
        val json = JSONArray()
        (load().filterNot { it.cid == item.cid } + item).forEach { entry ->
            json.put(JSONObject().apply {
                put("cid", entry.cid); put("state", entry.state.name)
                put("fetched_at", entry.fetchedAt); put("source", entry.source); put("detail", entry.detail)
            })
        }
        prefs.edit().putString("entries", json.toString()).apply()
    }

    private fun load(): List<InboxItem> = try {
        val json = JSONArray(prefs.getString("entries", "[]"))
        buildList {
            for (index in 0 until json.length()) {
                val entry = json.optJSONObject(index) ?: continue
                val cid = entry.optString("cid")
                val state = runCatching { InboxState.valueOf(entry.optString("state")) }.getOrNull() ?: continue
                if (CID_PATTERN.matches(cid)) add(InboxItem(cid, state, entry.optString("fetched_at"),
                    entry.optString("source"), entry.optString("detail")))
            }
        }
    } catch (_: Exception) { emptyList() }

    companion object {
        const val MAX_ENVELOPE_BYTES = 128 * 1024
        private const val MAX_DETAIL_CHARS = 512
        private val CID_PATTERN = Regex("b[a-z2-7]{10,200}")
    }
}

data class InboxItem(val cid: String, val state: InboxState, val fetchedAt: String, val source: String, val detail: String)

/** Local processing states, never assertions supplied by the relay. */
enum class InboxState {
    QUARANTINED,
    CANDIDATE_EVIDENCE,
    UNKNOWN_PROTOCOL,
    FORMAT_REJECTED,
    PAYLOAD_REJECTED,
    SIGNATURE_REJECTED,
    IDENTITY_REJECTED,
}
