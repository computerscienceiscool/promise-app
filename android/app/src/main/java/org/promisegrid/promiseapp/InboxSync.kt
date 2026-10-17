package org.promisegrid.promiseapp

/**
 * Retrieves availability hints and persists original envelope bytes locally.
 *
 * It deliberately stops at quarantine. A later verifier owns the transition
 * to candidate evidence or one of the rejection states in [InboxState].
 */
class InboxSync(
    private val relay: RelayClient,
    private val inbox: InboxStore,
    private val relaySource: String,
) {
    fun sync(): Result<InboxSyncResult> = runCatching {
        val listed = relay.mailbox().getOrThrow()
        var fetched = 0
        var alreadyRetained = 0
        val failures = mutableListOf<InboxFetchFailure>()

        for (cid in listed) {
            if (inbox.get(cid) != null) {
                alreadyRetained++
                continue
            }
            relay.fetchEnvelope(cid).fold(
                onSuccess = { bytes ->
                    try {
                        inbox.retainFetched(cid, bytes, relaySource)
                        fetched++
                    } catch (exception: Exception) {
                        failures += InboxFetchFailure(cid, exception.message ?: "Could not retain envelope")
                    }
                },
                onFailure = { exception ->
                    failures += InboxFetchFailure(cid, exception.message ?: "Could not fetch envelope")
                },
            )
        }
        InboxSyncResult(listed.size, fetched, alreadyRetained, failures)
    }
}

data class InboxSyncResult(
    val listed: Int,
    val fetched: Int,
    val alreadyRetained: Int,
    val failures: List<InboxFetchFailure>,
)

data class InboxFetchFailure(val cid: String, val detail: String)
