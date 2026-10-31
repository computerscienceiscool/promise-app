package org.promisegrid.promiseapp

import org.junit.Assert.assertEquals
import org.junit.Test

class DeliveryPresentationTest {
    private val cid = "bafytestmessagecid"

    @Test fun mapsRelayAndPhoneStatesWithoutClaimingAcceptance() {
        assertEquals("Relay check: ready", DeliveryPresentation.relayReady())
        assertEquals("Relay check: failed", DeliveryPresentation.relayFailed())
        assertEquals("Phone saved signed message $cid. Uploading to relay…", DeliveryPresentation.queued(cid))
        assertEquals(
            "Relay accepted $cid. This proves relay upload, not recipient verification.",
            DeliveryPresentation.uploaded(cid),
        )
        assertEquals(
            "Phone retained $cid. Relay upload failed; use Retry pending uploads.",
            DeliveryPresentation.retryableFailure(cid),
        )
    }

    @Test fun mapsRecipientVerificationToTheSameCid() {
        assertEquals("RECIPIENT VERIFIED cid=$cid", DeliveryPresentation.recipientVerified(cid))
    }
}
