package org.promisegrid.promiseapp

import java.net.HttpURLConnection
import java.net.URL
import java.nio.charset.StandardCharsets
import org.json.JSONObject

class RelayClient(private val relayURL: String, private val deviceID: String) {
    fun upload(envelope: ByteArray): Result<String> = runCatching {
        require(relayURL.startsWith("http://") || relayURL.startsWith("https://")) {
            "Relay URL must start with http:// or https://"
        }
        val connection = (URL(relayURL.trimEnd('/') + "/v1/envelopes").openConnection() as HttpURLConnection)
        try {
            connection.connectTimeout = CONNECT_TIMEOUT_MS
            connection.readTimeout = READ_TIMEOUT_MS
            connection.requestMethod = "POST"
            connection.setRequestProperty("Content-Type", "application/cbor")
            connection.setRequestProperty("Accept", "application/json")
            connection.setRequestProperty("X-Device-ID", deviceID)
            connection.fixedLengthStreamingMode(envelope.size)
            connection.doOutput = true
            connection.outputStream.use { it.write(envelope) }

            val code = connection.responseCode
            val response = (if (code in 200..299) connection.inputStream else connection.errorStream)
                ?.bufferedReader(StandardCharsets.UTF_8)
                ?.use { it.readText() }
                .orEmpty()
            if (code !in 200..299) {
                error("Relay rejected upload ($code): ${response.take(MAX_ERROR_CHARS)}")
            }
            response
        } finally {
            connection.disconnect()
        }
    }

    /** Lists immutable message CIDs available to this device without fetching or trusting them. */
    fun mailbox(): Result<List<String>> = runCatching {
        val connection = request("/v1/mailbox", "GET")
        try {
            val response = successfulBody(connection, "mailbox")
            val items = JSONObject(response).optJSONArray("items") ?: return@runCatching emptyList()
            buildList {
                for (index in 0 until items.length()) {
                    val item = items.optJSONObject(index) ?: continue
                    val cid = item.optString("cid")
                    if (cid.isNotBlank()) add(cid)
                }
            }
        } finally {
            connection.disconnect()
        }
    }

    /** Retrieves the relay's original bytes. Verification is deliberately a separate step. */
    fun fetchEnvelope(cid: String): Result<ByteArray> = runCatching {
        require(CID_PATTERN.matches(cid)) { "Invalid mailbox CID" }
        val connection = request("/v1/envelopes/$cid", "GET")
        try {
            val code = connection.responseCode
            if (code !in 200..299) {
                val detail = connection.errorStream?.bufferedReader(StandardCharsets.UTF_8)?.use { it.readText() }.orEmpty()
                error("Relay envelope fetch failed ($code): ${detail.take(MAX_ERROR_CHARS)}")
            }
            connection.inputStream.use { it.readBytes() }
        } finally {
            connection.disconnect()
        }
    }

    private fun request(path: String, method: String): HttpURLConnection {
        require(relayURL.startsWith("http://") || relayURL.startsWith("https://")) {
            "Relay URL must start with http:// or https://"
        }
        return (URL(relayURL.trimEnd('/') + path).openConnection() as HttpURLConnection).apply {
            connectTimeout = CONNECT_TIMEOUT_MS
            readTimeout = READ_TIMEOUT_MS
            requestMethod = method
            setRequestProperty("Accept", "application/json, application/cbor")
            setRequestProperty("X-Device-ID", deviceID)
        }
    }

    private fun successfulBody(connection: HttpURLConnection, operation: String): String {
        val code = connection.responseCode
        val response = (if (code in 200..299) connection.inputStream else connection.errorStream)
            ?.bufferedReader(StandardCharsets.UTF_8)
            ?.use { it.readText() }
            .orEmpty()
        if (code !in 200..299) error("Relay $operation failed ($code): ${response.take(MAX_ERROR_CHARS)}")
        return response
    }

    private companion object {
        const val CONNECT_TIMEOUT_MS = 10_000
        const val READ_TIMEOUT_MS = 15_000
        const val MAX_ERROR_CHARS = 500
        val CID_PATTERN = Regex("b[a-z2-7]{10,200}")
    }
}
