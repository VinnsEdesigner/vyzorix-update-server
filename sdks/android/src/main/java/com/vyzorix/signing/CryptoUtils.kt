package com.vyzorix.signing

import java.security.MessageDigest
import java.security.SecureRandom
import javax.crypto.Mac
import javax.crypto.Cipher
import javax.crypto.spec.GCMParameterSpec
import javax.crypto.spec.SecretKeySpec
import android.util.Base64

/**
 * Cryptographic utilities for Vyzorix API request signing.
 * Implements AES-256-GCM encryption and HMAC-SHA512 signing.
 * 
 * Security Model:
 * - Request bodies are encrypted with AES-256-GCM
 * - Requests are signed with HMAC-SHA512
 * - Timestamps must be within ±5 minutes of server time
 * - Signatures prevent replay attacks
 */
object CryptoUtils {
    
    private const val GCM_IV_LENGTH = 12 // 96 bits for GCM
    private const val GCM_TAG_LENGTH = 128 // bits
    private const val AES_KEY_SIZE = 256 // bits
    
    /**
     * Derives a 32-byte key from a secret using SHA-512.
     * Used to derive AES-256 key from client_secret.
     */
    fun deriveKey(secret: String): ByteArray {
        val digest = MessageDigest.getInstance("SHA-512")
        val hash = digest.digest(secret.toByteArray(Charsets.UTF_8))
        return hash.copyOfRange(0, 32) // First 32 bytes for AES-256
    }
    
    /**
     * Computes SHA-512 hash of the input string.
     */
    fun sha512(data: String): ByteArray {
        val digest = MessageDigest.getInstance("SHA-512")
        return digest.digest(data.toByteArray(Charsets.UTF_8))
    }
    
    /**
     * Computes SHA-512 hash and returns hex string.
     */
    fun sha512Hex(data: String): String {
        return sha512(data).joinToString("") { "%02x".format(it) }
    }
    
    /**
     * Computes HMAC-SHA512 of the message using the given key.
     */
    fun hmacSha512(key: String, message: String): ByteArray {
        val mac = Mac.getInstance("HmacSHA512")
        val secretKey = SecretKeySpec(key.toByteArray(Charsets.UTF_8), "HmacSHA512")
        mac.init(secretKey)
        return mac.doFinal(message.toByteArray(Charsets.UTF_8))
    }
    
    /**
     * Computes HMAC-SHA512 and returns hex string.
     */
    fun hmacSha512Hex(key: String, message: String): String {
        return hmacSha512(key, message).joinToString("") { "%02x".format(it) }
    }
    
    /**
     * Generates cryptographically secure random bytes.
     */
    fun randomBytes(length: Int): ByteArray {
        val bytes = ByteArray(length)
        SecureRandom().nextBytes(bytes)
        return bytes
    }
    
    /**
     * Generates a random nonce for AES-GCM (12 bytes).
     */
    fun generateNonce(): ByteArray = randomBytes(GCM_IV_LENGTH)
    
    /**
     * Encrypts plaintext using AES-256-GCM.
     * Returns a pair of (nonce, ciphertext) as Base64 strings.
     */
    fun aes256GcmEncrypt(secret: String, plaintext: String): EncryptionResult {
        val key = deriveKey(secret)
        val nonce = generateNonce()
        
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        val gcmSpec = GCMParameterSpec(GCM_TAG_LENGTH, nonce)
        val keySpec = SecretKeySpec(key, "AES")
        cipher.init(Cipher.ENCRYPT_MODE, keySpec, gcmSpec)
        
        val ciphertext = cipher.doFinal(plaintext.toByteArray(Charsets.UTF_8))
        
        return EncryptionResult(
            nonce = Base64.encodeToString(nonce, Base64.NO_WRAP),
            ciphertext = Base64.encodeToString(ciphertext, Base64.NO_WRAP)
        )
    }
    
    /**
     * Encrypts plaintext (from a Map/object) using AES-256-GCM.
     * Automatically serializes to JSON.
     */
    fun aes256GcmEncrypt(secret: String, data: Map<String, Any?>): EncryptionResult {
        val json = mapToJson(data)
        return aes256GcmEncrypt(secret, json)
    }
    
    /**
     * Decrypts ciphertext using AES-256-GCM.
     * Expects nonce and ciphertext as Base64 strings.
     */
    fun aes256GcmDecrypt(secret: String, nonceB64: String, ciphertextB64: String): String {
        val key = deriveKey(secret)
        val nonce = Base64.decode(nonceB64, Base64.NO_WRAP)
        val ciphertext = Base64.decode(ciphertextB64, Base64.NO_WRAP)
        
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        val gcmSpec = GCMParameterSpec(GCM_TAG_LENGTH, nonce)
        val keySpec = SecretKeySpec(key, "AES")
        cipher.init(Cipher.DECRYPT_MODE, keySpec, gcmSpec)
        
        val plaintext = cipher.doFinal(ciphertext)
        return String(plaintext, Charsets.UTF_8)
    }
    
    /**
     * Decrypts from combined ciphertext (nonce prepended).
     */
    fun aes256GcmDecryptCombined(secret: String, combinedB64: String): String {
        val combined = Base64.decode(combinedB64, Base64.NO_WRAP)
        val nonce = combined.copyOfRange(0, GCM_IV_LENGTH)
        val ciphertext = combined.copyOfRange(GCM_IV_LENGTH, combined.size)
        
        val key = deriveKey(secret)
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        val gcmSpec = GCMParameterSpec(GCM_TAG_LENGTH, nonce)
        val keySpec = SecretKeySpec(key, "AES")
        cipher.init(Cipher.DECRYPT_MODE, keySpec, gcmSpec)
        
        val plaintext = cipher.doFinal(ciphertext)
        return String(plaintext, Charsets.UTF_8)
    }
    
    /**
     * Converts a Map to JSON string.
     */
    private fun mapToJson(map: Map<String, Any?>): String {
        val sb = StringBuilder()
        sb.append("{")
        map.entries.forEachIndexed { index, (key, value) ->
            if (index > 0) sb.append(",")
            sb.append("\"$key\":")
            when (value) {
                is String -> sb.append("\"${value.replace("\"", "\\\"")}\"")
                is Number -> sb.append(value)
                is Boolean -> sb.append(value)
                null -> sb.append("null")
                is Map<*, *> -> sb.append(mapToJson(value as Map<String, Any?>))
                is List<*> -> sb.append(listToJson(value as List<Any?>))
                else -> sb.append("\"${value.toString().replace("\"", "\\\"")}\"")
            }
        }
        sb.append("}")
        return sb.toString()
    }
    
    /**
     * Converts a List to JSON string.
     */
    private fun listToJson(list: List<Any?>): String {
        val sb = StringBuilder()
        sb.append("[")
        list.forEachIndexed { index, value ->
            if (index > 0) sb.append(",")
            when (value) {
                is String -> sb.append("\"${value.replace("\"", "\\\"")}\"")
                is Number -> sb.append(value)
                is Boolean -> sb.append(value)
                null -> sb.append("null")
                is Map<*, *> -> sb.append(mapToJson(value as Map<String, Any?>))
                is List<*> -> sb.append(listToJson(value as List<Any?>))
                else -> sb.append("\"${value.toString().replace("\"", "\\\"")}\"")
            }
        }
        sb.append("]")
        return sb.toString()
    }
}

/**
 * Result of AES-256-GCM encryption.
 */
data class EncryptionResult(
    val nonce: String,
    val ciphertext: String
)
