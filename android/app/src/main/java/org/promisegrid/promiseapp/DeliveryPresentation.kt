package org.promisegrid.promiseapp

/**
 * Human-readable POC delivery states. These describe local implementation
 * evidence; none of them represents recipient acceptance of a promise.
 */
object DeliveryPresentation {
    fun relayReady() = "Relay check: ready"

    fun relayFailed() = "Relay check: failed"

    fun queued(cid: String) = "Phone saved signed message $cid. Uploading to relay…"

    fun uploaded(cid: String) =
        "Relay accepted $cid. This proves relay upload, not recipient verification."

    fun retryableFailure(cid: String) =
        "Phone retained $cid. Relay upload failed; use Retry pending uploads."

    fun recipientVerified(cid: String) = "RECIPIENT VERIFIED cid=$cid"
}
