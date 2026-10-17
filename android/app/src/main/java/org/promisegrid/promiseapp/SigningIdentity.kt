package org.promisegrid.promiseapp

import android.content.Context
import android.os.Build
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import java.security.KeyPair
import java.security.KeyPairGenerator
import java.security.KeyStore
import java.security.PrivateKey
import java.security.PublicKey
import java.security.Signature

/**
 * The device signing identity. Its private key is generated and retained by
 * Android Keystore; application storage only ever sees the public key.
 *
 * Android Keystore gained Ed25519 generation in API 31. The Promise Messaging
 * POC specifies Ed25519, so older devices fail closed rather than falling back
 * to an exportable key or silently changing the protocol algorithm.
 */
class SigningIdentity(context: Context) {
    val keyPair: KeyPair

    init {
        require(Build.VERSION.SDK_INT >= MIN_ED25519_KEYSTORE_API) {
            "This protocol requires Android $MIN_ED25519_KEYSTORE_API or later for a Keystore-backed Ed25519 key"
        }
        keyPair = load() ?: generate()
        removeLegacyExportedKey(context)
    }

    fun sign(bytes: ByteArray): ByteArray = Signature.getInstance("Ed25519").run {
        initSign(keyPair.private)
        update(bytes)
        sign()
    }

    private fun load(): KeyPair? {
        val privateKey = keyStore.getKey(ALIAS, null) as? PrivateKey ?: return null
        val publicKey = keyStore.getCertificate(ALIAS)?.publicKey as? PublicKey
            ?: error("Android Keystore identity is missing its public certificate")
        return KeyPair(publicKey, privateKey)
    }

    private fun generate(): KeyPair {
        val parameters = KeyGenParameterSpec.Builder(ALIAS, KeyProperties.PURPOSE_SIGN)
            .setDigests(KeyProperties.DIGEST_NONE)
            .build()
        return KeyPairGenerator.getInstance("Ed25519", ANDROID_KEYSTORE).run {
            initialize(parameters)
            generateKeyPair()
        }
    }

    /** Keys stored by the old POC were exportable; never retain them after migration. */
    private fun removeLegacyExportedKey(context: Context) {
        context.getSharedPreferences(LEGACY_PREFERENCES, Context.MODE_PRIVATE)
            .edit()
            .remove("private")
            .remove("public")
            .apply()
    }

    private companion object {
        const val ALIAS = "promisegrid.messaging.ed25519.v1"
        const val ANDROID_KEYSTORE = "AndroidKeyStore"
        const val LEGACY_PREFERENCES = "promise-signing-key"
        const val MIN_ED25519_KEYSTORE_API = 31

        val keyStore: KeyStore = KeyStore.getInstance(ANDROID_KEYSTORE).apply { load(null) }
    }
}
