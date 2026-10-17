package org.promisegrid.promiseapp

import java.io.ByteArrayOutputStream

/** Minimal canonical-CBOR writer for the fixed POC envelope fields. */
object CanonicalCbor {
    fun text(value: String) = bytes(3, value.encodeToByteArray())
    fun binary(value: ByteArray) = bytes(2, value)
    fun unsigned(value: Long) = join(0, value, emptyArray<ByteArray>())
    fun array(count: Int, vararg values: ByteArray) = join(4, count.toLong(), values)
    /**
     * Deterministic CBOR orders map keys by the byte length of their encoded
     * form and then bytewise. Sorting Kotlin strings is not equivalent (for
     * example, "to" precedes "body" here because its encoded key is shorter).
     */
    fun map(entries: List<Pair<String, ByteArray>>): ByteArray {
        val encoded = entries.map { text(it.first) to it.second }.sortedWith { a, b ->
            if (a.first.size != b.first.size) a.first.size - b.first.size else compareBytes(a.first, b.first)
        }
        return join(5, encoded.size.toLong(), encoded.flatMap { listOf(it.first, it.second) }.toTypedArray())
    }
    fun tag(number: Long, value: ByteArray) = join(6, number, arrayOf(value))
    private fun bytes(major: Int, value: ByteArray) = join(major, value.size.toLong(), arrayOf(value))
    private fun join(major: Int, value: Long, parts: Array<out ByteArray>): ByteArray {
        require(value >= 0) { "CBOR unsigned value must be non-negative" }
        val output = ByteArrayOutputStream(); header(output, major, value); parts.forEach(output::write); return output.toByteArray()
    }
    private fun header(out: ByteArrayOutputStream, major: Int, value: Long) {
        when {
            value < 24 -> out.write((major shl 5) or value.toInt())
            value <= 0xff -> { out.write((major shl 5) or 24); out.write(value.toInt()) }
            value <= 0xffff -> { out.write((major shl 5) or 25); out.write((value ushr 8).toInt()); out.write(value.toInt()) }
            value <= 0xffffffffL -> {
                out.write((major shl 5) or 26)
                for (shift in 24 downTo 0 step 8) out.write((value ushr shift).toInt())
            }
            else -> {
                out.write((major shl 5) or 27)
                for (shift in 56 downTo 0 step 8) out.write((value ushr shift).toInt())
            }
        }
    }

    private fun compareBytes(left: ByteArray, right: ByteArray): Int {
        for (index in left.indices) {
            val difference = (left[index].toInt() and 0xff) - (right[index].toInt() and 0xff)
            if (difference != 0) return difference
        }
        return 0
    }
}
