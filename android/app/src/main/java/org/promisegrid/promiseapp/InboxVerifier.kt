package org.promisegrid.promiseapp

import java.nio.ByteBuffer
import java.nio.charset.CharacterCodingException
import java.nio.charset.CodingErrorAction
import java.nio.charset.StandardCharsets
import java.security.Signature

/** Strict verifier for the one fixed Promise Messaging v0 CBOR schema.
 *
 * It deliberately has no "accept" outcome: a valid envelope is only local
 * candidate evidence.  The relay cannot choose that state or supply a CID.
 */
object InboxVerifier {
    private const val MAX_BODY = 4096

    fun verify(expectedCid: String, bytes: ByteArray, inbox: InboxStore): InboxItem {
        fun fail(state: InboxState, detail: String): InboxItem = inbox.mark(expectedCid, state, detail)
        return try {
            if (bytes.isEmpty() || bytes.size > InboxStore.MAX_ENVELOPE_BYTES) return fail(InboxState.FORMAT_REJECTED, "Invalid envelope size")
            val actualCid = PromiseEnvelope.envelopeCid(bytes).toString()
            if (actualCid != expectedCid) return fail(InboxState.FORMAT_REJECTED, "Relay CID does not match retained bytes")
            val outer = Reader(bytes).one() as? V.Tag ?: return fail(InboxState.FORMAT_REJECTED, "Envelope is not tagged")
            if (outer.number != PromiseEnvelope.GRID_TAG) return fail(InboxState.FORMAT_REJECTED, "Wrong grid tag")
            val slots = outer.value as? V.Array ?: return fail(InboxState.FORMAT_REJECTED, "Envelope slots are not an array")
            if (slots.values.size != 3) return fail(InboxState.FORMAT_REJECTED, "Envelope must have three slots")
            val selector = slots.values[0].cidTag() ?: return fail(InboxState.FORMAT_REJECTED, "Invalid protocol selector")
            val payloadBytes = (slots.values[1] as? V.Bytes)?.value ?: return fail(InboxState.FORMAT_REJECTED, "Invalid payload slot")
            val proofBytes = (slots.values[2] as? V.Bytes)?.value ?: return fail(InboxState.FORMAT_REJECTED, "Invalid proof slot")
            if (selector != PromiseEnvelope.protocolCid) return fail(InboxState.UNKNOWN_PROTOCOL, "Unsupported protocol $selector")
            val payload = parsePayload(payloadBytes) ?: return fail(InboxState.PAYLOAD_REJECTED, "Invalid or non-canonical payload")
            val proof = parseProof(proofBytes) ?: return fail(InboxState.SIGNATURE_REJECTED, "Invalid or non-canonical proof")
            val projection = grid(slots.values[0].encoded, payloadBytes, CanonicalCbor.map(listOf(
                "algorithm" to CanonicalCbor.text("Ed25519"), "public_key" to CanonicalCbor.binary(proof.publicKey))))
            val signature = Signature.getInstance("Ed25519")
            signature.initVerify(java.security.KeyFactory.getInstance("Ed25519").generatePublic(java.security.spec.X509EncodedKeySpec(SPKI + proof.publicKey)))
            signature.update(projection)
            if (!signature.verify(proof.signature)) return fail(InboxState.SIGNATURE_REJECTED, "Ed25519 signature failed")
            val signer = PromiseEnvelope.identityCid(proof.publicKey)
            if (payload.from != signer || payload.from == payload.to) return fail(InboxState.IDENTITY_REJECTED, "Sender identity does not match signing key")
            inbox.mark(expectedCid, InboxState.CANDIDATE_EVIDENCE, "Verified canonical envelope and Ed25519 proof; local candidate only")
        } catch (e: Exception) {
            fail(InboxState.FORMAT_REJECTED, e.message ?: "Malformed envelope")
        }
    }

    private data class Payload(val kind: String, val from: PromiseEnvelope.Cid, val to: PromiseEnvelope.Cid, val body: String, val invokes: List<PromiseEnvelope.Cid>, val created: Long)
    private data class Proof(val publicKey: ByteArray, val signature: ByteArray)

    private fun parsePayload(data: ByteArray): Payload? = try {
        val map = Reader(data).oneMap() ?: return null
        val kind = map.text("kind") ?: return null; val from = map.cid("from") ?: return null; val to = map.cid("to") ?: return null
        val body = map.text("body") ?: return null; val invokes = map.cidArray("invokes") ?: return null; val created = map.uint("created_at") ?: return null
        if (kind !in setOf("promise_offer", "promise_accept", "promise_decline", "observation") || body.encodeToByteArray().size !in 1..MAX_BODY || from == to || invokes.size > 4) return null
        if ((kind == "promise_offer" && invokes.isNotEmpty()) || ((kind == "promise_accept" || kind == "promise_decline") && invokes.size != 1) || (kind == "observation" && invokes.isEmpty())) return null
        if (invokes.distinct().size != invokes.size) return null
        val rebuilt = CanonicalCbor.map(listOf("kind" to CanonicalCbor.text(kind), "from" to tag(from), "to" to tag(to), "body" to CanonicalCbor.text(body), "invokes" to CanonicalCbor.array(invokes.size, *invokes.map(::tag).toTypedArray()), "created_at" to CanonicalCbor.unsigned(created)))
        if (!data.contentEquals(rebuilt)) null else Payload(kind, from, to, body, invokes, created)
    } catch (_: Exception) { null }

    private fun parseProof(data: ByteArray): Proof? = try {
        val map = Reader(data).oneMap() ?: return null
        val algorithm = map.text("algorithm") ?: return null; val key = map.bytes("public_key") ?: return null; val sig = map.bytes("signature") ?: return null
        if (algorithm != "Ed25519" || key.size != 32 || sig.size != 64) return null
        val rebuilt = CanonicalCbor.map(listOf("algorithm" to CanonicalCbor.text(algorithm), "public_key" to CanonicalCbor.binary(key), "signature" to CanonicalCbor.binary(sig)))
        if (!data.contentEquals(rebuilt)) null else Proof(key, sig)
    } catch (_: Exception) { null }

    private fun grid(selector: ByteArray, payload: ByteArray, proof: ByteArray) = CanonicalCbor.tag(PromiseEnvelope.GRID_TAG, CanonicalCbor.array(3, selector, CanonicalCbor.binary(payload), CanonicalCbor.binary(proof)))
    private fun tag(cid: PromiseEnvelope.Cid) = CanonicalCbor.tag(42, CanonicalCbor.binary(byteArrayOf(0) + cid.bytes))
    private val SPKI = byteArrayOf(0x30, 0x2a, 0x30, 0x05, 0x06, 0x03, 0x2b, 0x65, 0x70, 0x03, 0x21, 0x00)

    private sealed class V(open val encoded: ByteArray) {
        data class UInt(val value: Long, override val encoded: ByteArray): V(encoded)
        data class Bytes(val value: ByteArray, override val encoded: ByteArray): V(encoded)
        data class Text(val value: String, override val encoded: ByteArray): V(encoded)
        data class Array(val values: List<V>, override val encoded: ByteArray): V(encoded)
        data class Map(val values: List<Pair<V, V>>, override val encoded: ByteArray): V(encoded)
        data class Tag(val number: Long, val value: V, override val encoded: ByteArray): V(encoded)
    }
    private fun V.cidTag(): PromiseEnvelope.Cid? = ((this as? V.Tag)?.takeIf { it.number == 42 }?.value as? V.Bytes)?.value?.let { if (it.size > 1 && it[0].toInt() == 0) runCatching { PromiseEnvelope.Cid.fromBytes(it.copyOfRange(1, it.size)) }.getOrNull() else null }
    private fun V.Map.text(key: String) = values.singleOrNull { (it.first as? V.Text)?.value == key }?.second.let { (it as? V.Text)?.value }
    private fun V.Map.bytes(key: String) = values.singleOrNull { (it.first as? V.Text)?.value == key }?.second.let { (it as? V.Bytes)?.value }
    private fun V.Map.uint(key: String) = values.singleOrNull { (it.first as? V.Text)?.value == key }?.second.let { (it as? V.UInt)?.value }
    private fun V.Map.cid(key: String) = values.singleOrNull { (it.first as? V.Text)?.value == key }?.second?.cidTag()
    private fun V.Map.cidArray(key: String): List<PromiseEnvelope.Cid>? = ((values.singleOrNull { (it.first as? V.Text)?.value == key }?.second as? V.Array)?.values ?: return null).map { it.cidTag() ?: return null }

    /** Rejects indefinite values, duplicate/non-minimal integer encodings and trailing bytes. */
    private class Reader(private val input: ByteArray) {
        private var pos = 0
        fun one(): V { val v = value(); require(pos == input.size) { "Trailing bytes" }; return v }
        fun oneMap() = one() as? V.Map
        private fun value(): V {
            val start = pos; require(pos < input.size) { "Unexpected end" }; val head = input[pos++].toInt() and 255; val major = head ushr 5; val n = argument(head and 31)
            fun raw() = input.copyOfRange(start, pos)
            return when (major) {
                0 -> V.UInt(n, raw())
                2 -> { require(n <= input.size - pos) { "Truncated bytes" }; val b = input.copyOfRange(pos, pos + n.toInt()); pos += n.toInt(); V.Bytes(b, input.copyOfRange(start, pos)) }
                3 -> { require(n <= input.size - pos) { "Truncated text" }; val b = input.copyOfRange(pos, pos + n.toInt()); pos += n.toInt(); V.Text(utf8(b), input.copyOfRange(start, pos)) }
                4 -> { require(n <= 16) { "Array too large" }; val vs = List(n.toInt()) { value() }; V.Array(vs, input.copyOfRange(start, pos)) }
                5 -> { require(n <= 16) { "Map too large" }; val vs = List(n.toInt()) { value() to value() }; require(vs.map { it.first.encoded.toList() }.distinct().size == vs.size) { "Duplicate map key" }; V.Map(vs, input.copyOfRange(start, pos)) }
                6 -> { val child = value(); V.Tag(n, child, input.copyOfRange(start, pos)) }
                else -> error("Unsupported CBOR type")
            }
        }
        private fun argument(ai: Int): Long {
            require(ai < 28) { "Indefinite/reserved argument" }; if (ai < 24) return ai.toLong()
            val count = 1 shl (ai - 24); require(pos + count <= input.size) { "Truncated argument" }; var v = 0L
            repeat(count) { v = (v shl 8) or (input[pos++].toInt() and 255).toLong() }
            require(v >= if (ai == 24) 24 else 1L shl (8 * (count - 1))) { "Non-canonical argument" }; return v
        }
        private fun utf8(b: ByteArray): String = try { StandardCharsets.UTF_8.newDecoder().onMalformedInput(CodingErrorAction.REPORT).onUnmappableCharacter(CodingErrorAction.REPORT).decode(ByteBuffer.wrap(b)).toString() } catch (_: CharacterCodingException) { error("Invalid UTF-8") }
    }
}
