package org.promisegrid.promiseapp

import android.content.Context
import android.util.Base64
import java.security.KeyFactory
import java.security.KeyPair
import java.security.KeyPairGenerator
import java.security.PrivateKey
import java.security.PublicKey
import java.security.spec.PKCS8EncodedKeySpec
import java.security.spec.X509EncodedKeySpec
import java.security.Signature

/** POC identity. The next storage step moves the private key behind Android Keystore. */
class SigningIdentity(context: Context) {
    private val prefs = context.getSharedPreferences("promise-signing-key", Context.MODE_PRIVATE)
    val keyPair: KeyPair = load() ?: KeyPairGenerator.getInstance("Ed25519").generateKeyPair().also(::save)

    private fun load(): KeyPair? = try {
        val privateBytes = prefs.getString("private", null) ?: return null
        val publicBytes = prefs.getString("public", null) ?: return null
        val factory = KeyFactory.getInstance("Ed25519")
        KeyPair(
            factory.generatePublic(X509EncodedKeySpec(Base64.decode(publicBytes, Base64.NO_WRAP))),
            factory.generatePrivate(PKCS8EncodedKeySpec(Base64.decode(privateBytes, Base64.NO_WRAP)))
        )
    } catch (_: Exception) { null }

    private fun save(pair: KeyPair) {
        prefs.edit()
            .putString("private", Base64.encodeToString(pair.private.encoded, Base64.NO_WRAP))
            .putString("public", Base64.encodeToString(pair.public.encoded, Base64.NO_WRAP))
            .apply()
    }

    fun sign(bytes: ByteArray): ByteArray = Signature.getInstance("Ed25519").run {
        initSign(keyPair.private)
        update(bytes)
        sign()
    }
}
