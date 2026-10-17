package org.promisegrid.promiseapp

import java.security.MessageDigest
import java.security.PublicKey

/**
 * Exact Promise Messaging v0 POC encoder. The relay receives [bytes] without
 * modification; [cid] identifies those complete bytes.
 *
 * This is intentionally small and does not claim a frozen PromiseGrid SDK.
 * Its byte layout mirrors backend/internal/protocol.Build exactly.
 */
object PromiseEnvelope {
    const val GRID_TAG = 1735551332L
    private const val IDENTITY_CODEC = 0x300001L
    private const val RAW_CODEC = 0x55L
    private const val SHA2_256 = 0x12L
    private const val MAX_BODY_BYTES = 4096

    // Frozen Promise Messaging v0 selector. It identifies the immutable
    // descriptor in backend/testdata/promise-messaging-v0-descriptor.json.
    val protocolCid: Cid = Cid.fromString("bafkreicutumpjz3skwoehjn3gptve5m3atgypmfbsdhbtapg2rcwt2c2hi")

    data class Signed(val bytes: ByteArray, val cid: Cid, val from: Cid, val to: Cid)

    /** CIDv1 bytes, including version, codec, multihash code, size, and digest. */
    class Cid private constructor(val bytes: ByteArray) {
        init { require(isCidV1(bytes)) { "expected a CIDv1" } }
        override fun equals(other: Any?) = other is Cid && bytes.contentEquals(other.bytes)
        override fun hashCode() = bytes.contentHashCode()
        override fun toString(): String = "b" + base32(bytes)

        companion object {
            fun fromString(value: String): Cid {
                require(value.length > 1 && value[0] == 'b') { "CID must use lowercase base32 multibase" }
                return Cid(decodeBase32(value.substring(1)))
            }
            fun fromBytes(value: ByteArray) = Cid(value.copyOf())
        }
    }

    fun identityCid(publicKey: PublicKey): Cid = identityCid(rawEd25519(publicKey))

    /**
     * Returns the CID of complete, immutable envelope bytes.  This is the
     * value a person records during the two-phone proof; it is intentionally
     * distinct from the outbox's private storage key.
     */
    fun envelopeCid(envelope: ByteArray): Cid {
        require(envelope.isNotEmpty()) { "Envelope must not be empty" }
        return cidFor(envelope, RAW_CODEC)
    }

    fun identityCid(publicKey: ByteArray): Cid {
        require(publicKey.size == 32) { "Ed25519 public key must be 32 bytes" }
        val identityRecord = CanonicalCbor.map(listOf(
            "version" to CanonicalCbor.unsigned(0),
            "public_key" to CanonicalCbor.binary(publicKey),
        ))
        return cidFor(identityRecord, IDENTITY_CODEC)
    }

    fun rawEd25519(publicKey: PublicKey): ByteArray {
        // Android Keystore providers legitimately emit either RFC 8410's
        // parameter-free SPKI or a legacy SPKI with an explicit ASN.1 NULL
        // parameter. Both carry the same Ed25519 OID and a 32-byte BIT STRING.
        val encoded = publicKey.encoded
        val oidAt = encoded.indexOf(ED25519_OID)
        require(oidAt >= 0) { "unsupported Ed25519 public key encoding" }
        var bitStringAt = oidAt + ED25519_OID.size
        if (encoded.size >= bitStringAt + 2 && encoded[bitStringAt] == 0x05.toByte() && encoded[bitStringAt + 1] == 0x00.toByte()) {
            bitStringAt += 2
        }
        require(
            encoded.size == bitStringAt + 35 &&
                encoded[bitStringAt] == 0x03.toByte() &&
                encoded[bitStringAt + 1] == 0x21.toByte() &&
                encoded[bitStringAt + 2] == 0x00.toByte(),
        ) {
            "unsupported Ed25519 public key encoding"
        }
        return encoded.copyOfRange(bitStringAt + 3, encoded.size)
    }

    fun offer(
        signer: SigningIdentity,
        recipient: Cid,
        body: String,
        createdAt: Long = System.currentTimeMillis() / 1000,
    ): Signed {
        require(body.encodeToByteArray().isNotEmpty() && body.encodeToByteArray().size <= MAX_BODY_BYTES) { "body must be 1-$MAX_BODY_BYTES UTF-8 bytes" }
        require(createdAt >= 0) { "created_at must be unsigned" }
        val publicKey = rawEd25519(signer.keyPair.public)
        val from = identityCid(publicKey)
        require(from != recipient) { "recipient must differ from sender" }

        val payload = CanonicalCbor.map(listOf(
            "kind" to CanonicalCbor.text("promise_offer"),
            "from" to cidTag(from),
            "to" to cidTag(recipient),
            "body" to CanonicalCbor.text(body),
            "invokes" to CanonicalCbor.array(0),
            "created_at" to CanonicalCbor.unsigned(createdAt),
        ))
        val header = CanonicalCbor.map(listOf(
            "algorithm" to CanonicalCbor.text("Ed25519"),
            "public_key" to CanonicalCbor.binary(publicKey),
        ))
        val selector = cidTag(protocolCid)
        val signingProjection = grid(selector, payload, header)
        val proof = CanonicalCbor.map(listOf(
            "algorithm" to CanonicalCbor.text("Ed25519"),
            "public_key" to CanonicalCbor.binary(publicKey),
            "signature" to CanonicalCbor.binary(signer.sign(signingProjection)),
        ))
        val envelope = grid(selector, payload, proof)
        return Signed(envelope, cidFor(envelope, RAW_CODEC), from, recipient)
    }

    private fun grid(selector: ByteArray, payload: ByteArray, proof: ByteArray): ByteArray =
        CanonicalCbor.tag(GRID_TAG, CanonicalCbor.array(3, selector, CanonicalCbor.binary(payload), CanonicalCbor.binary(proof)))

    private fun cidTag(value: Cid): ByteArray =
        CanonicalCbor.tag(42, CanonicalCbor.binary(byteArrayOf(0) + value.bytes))

    private fun cidFor(data: ByteArray, codec: Long): Cid {
        val digest = MessageDigest.getInstance("SHA-256").digest(data)
        return Cid.fromBytes(varint(1) + varint(codec) + varint(SHA2_256) + varint(digest.size.toLong()) + digest)
    }

    private fun varint(value: Long): ByteArray {
        require(value >= 0)
        var remaining = value
        val result = ArrayList<Byte>()
        do {
            var next = (remaining and 0x7f).toInt()
            remaining = remaining ushr 7
            if (remaining != 0L) next = next or 0x80
            result += next.toByte()
        } while (remaining != 0L)
        return result.toByteArray()
    }

    private fun isCidV1(value: ByteArray): Boolean = try {
        val (version, afterVersion) = readVarint(value, 0)
        if (version != 1L) {
            false
        } else {
            // CIDv1 is version, codec, then a complete multihash. Do not
            // infer a codec from the printable text: identity CIDs use a
            // different codec and therefore need not begin with "bafk".
            val (_, afterCodec) = readVarint(value, afterVersion)
            val (_, afterHashCode) = readVarint(value, afterCodec)
            val (digestLength, afterHashLength) = readVarint(value, afterHashCode)
            digestLength == (value.size - afterHashLength).toLong()
        }
    } catch (_: IllegalArgumentException) { false }

    private fun readVarint(value: ByteArray, start: Int): Pair<Long, Int> {
        var result = 0L
        var shift = 0
        var index = start
        while (index < value.size && shift < 64) {
            val next = value[index++].toInt() and 0xff
            result = result or ((next and 0x7f).toLong() shl shift)
            if ((next and 0x80) == 0) {
                // CID varints must use their shortest representation. This
                // also prevents several text values from representing one CID.
                if (varint(result).size != index - start) throw IllegalArgumentException("non-canonical varint")
                return result to index
            }
            shift += 7
        }
        throw IllegalArgumentException("invalid varint")
    }

    private fun base32(value: ByteArray): String {
        val alphabet = "abcdefghijklmnopqrstuvwxyz234567"
        var buffer = 0
        var bits = 0
        val result = StringBuilder()
        for (byte in value) {
            buffer = (buffer shl 8) or (byte.toInt() and 0xff)
            bits += 8
            while (bits >= 5) {
                result.append(alphabet[(buffer shr (bits - 5)) and 31])
                bits -= 5
            }
        }
        if (bits > 0) result.append(alphabet[(buffer shl (5 - bits)) and 31])
        return result.toString()
    }

    private fun decodeBase32(value: String): ByteArray {
        var buffer = 0
        var bits = 0
        val result = ArrayList<Byte>()
        for (char in value) {
            val digit = "abcdefghijklmnopqrstuvwxyz234567".indexOf(char)
            require(digit >= 0) { "invalid lowercase base32" }
            buffer = (buffer shl 5) or digit
            bits += 5
            while (bits >= 8) {
                result += ((buffer shr (bits - 8)) and 0xff).toByte()
                bits -= 8
            }
        }
        require(bits == 0 || (buffer and ((1 shl bits) - 1)) == 0) { "non-canonical base32 padding" }
        return result.toByteArray()
    }

    private val ED25519_OID = byteArrayOf(0x06, 0x03, 0x2b, 0x65, 0x70)

    private fun ByteArray.indexOf(needle: ByteArray): Int {
        if (needle.isEmpty() || needle.size > size) return -1
        for (start in 0..size - needle.size) {
            if (copyOfRange(start, start + needle.size).contentEquals(needle)) return start
        }
        return -1
    }
}
