package com.vyzorix.signing

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.io.BufferedReader
import java.io.InputStreamReader
import java.net.HttpURLConnection
import java.net.URL
import java.net.URLEncoder
import java.text.SimpleDateFormat
import java.util.Locale
import java.util.TimeZone

/**
 * Client credentials for API authentication.
 */
data class ClientCredentials(
    val clientId: String,
    val clientSecret: String,
    val deviceId: String? = null,
    val commandSecret: String? = null
)

/**
 * Result of an API call.
 */
data class ApiResult<T>(
    val success: Boolean,
    val data: T?,
    val error: String?,
    val statusCode: Int
)

/**
 * Request options for the signed API client.
 */
data class SignedRequestOptions(
    val method: String = "GET",
    val path: String,
    val body: Map<String, Any?>? = null,
    val headers: Map<String, String> = emptyMap()
)

/**
 * Signed API client that implements request signing and response encryption for Android.
 * 
 * Usage:
 * ```
 * val client = SignedApiClient("https://api.example.com")
 * 
 * // After device registration, set credentials
 * client.setCredentials(ClientCredentials(
 *     clientId = "device-123",
 *     clientSecret = "secret-abc"
 * ))
 * 
 * // Make signed requests
 * val result = client.get<Map<String, Any>>("/v1/device/status")
 * ```
 */
class SignedApiClient(
    private val baseUrl: String,
    private val timeout: Int = 30000
) {
    private var credentials: ClientCredentials? = null
    
    // Timestamp window in seconds (5 minutes)
    private val timestampWindow = 300
    
    /**
     * Set the credentials for this client.
     * Call this after device registration or when loading stored credentials.
     */
    fun setCredentials(creds: ClientCredentials) {
        this.credentials = creds
    }
    
    /**
     * Get the current credentials.
     */
    fun getCredentials(): ClientCredentials? = credentials
    
    /**
     * Clear credentials (on logout).
     */
    fun clearCredentials() {
        credentials = null
    }
    
    /**
     * Check if credentials are available.
     */
    fun hasCredentials(): Boolean = credentials != null
    
    /**
     * Make a signed GET request.
     */
    suspend fun <T> get(path: String): ApiResult<T> {
        return request(SignedRequestOptions(method = "GET", path = path))
    }
    
    /**
     * Make a signed POST request.
     */
    suspend fun <T> post(path: String, body: Map<String, Any?>): ApiResult<T> {
        return request(SignedRequestOptions(method = "POST", path = path, body = body))
    }
    
    /**
     * Make a signed PATCH request.
     */
    suspend fun <T> patch(path: String, body: Map<String, Any?>): ApiResult<T> {
        return request(SignedRequestOptions(method = "PATCH", path = path, body = body))
    }
    
    /**
     * Make a signed DELETE request.
     */
    suspend fun <T> delete(path: String): ApiResult<T> {
        return request(SignedRequestOptions(method = "DELETE", path = path))
    }
    
    /**
     * Make a signed request with automatic signing and encryption.
     */
    @Suppress("UNCHECKED_CAST")
    suspend fun <T> request(options: SignedRequestOptions): ApiResult<T> {
        return withContext(Dispatchers.IO) {
            val creds = credentials 
                ?: return@withContext ApiResult(
                    success = false,
                    data = null,
                    error = "No credentials. Call setCredentials() first.",
                    statusCode = 401
                )
            
            try {
                val url = URL("$baseUrl${options.path}")
                val connection = url.openConnection() as HttpURLConnection
                connection.requestMethod = options.method
                connection.timeout = timeout
                connection.doInput = true
                
                // Generate timestamp
                val timestamp = System.currentTimeMillis() / 1000
                
                // Build body string
                val bodyString = options.body?.let { mapToJson(it) } ?: ""
                
                // Encrypt body
                val encrypted = CryptoUtils.aes256GcmEncrypt(creds.clientSecret, bodyString)
                val encryptedBody = "${encrypted.nonce}.${encrypted.ciphertext}"
                
                // Create signature
                val bodyHash = CryptoUtils.sha512Hex(bodyString)
                val stringToSign = "${options.method}\n${options.path}\n$timestamp\n$bodyHash"
                val signature = "t=$timestamp,v1=${CryptoUtils.hmacSha512Hex(creds.clientSecret, stringToSign)}"
                
                // Set headers
                connection.setRequestProperty("Content-Type", "application/json")
                connection.setRequestProperty("X-Client-ID", creds.clientId)
                connection.setRequestProperty("X-Timestamp", timestamp.toString())
                connection.setRequestProperty("X-Signature", signature)
                connection.setRequestProperty("X-Encrypted-Body", encryptedBody)
                
                // Add custom headers
                options.headers.forEach { (key, value) ->
                    connection.setRequestProperty(key, value)
                }
                
                // Write body for POST/PATCH/DELETE
                if (options.method in listOf("POST", "PATCH", "DELETE") && bodyString.isNotEmpty()) {
                    connection.doOutput = true
                    connection.outputStream.write(bodyString.toByteArray(Charsets.UTF_8))
                }
                
                // Read response
                val responseCode = connection.responseCode
                val responseBody = connection.inputStream.bufferedReader().use { it.readText() }
                
                // Check for encrypted response
                val encryptionHeader = connection.getHeaderField("X-Content-Encryption")
                
                val resultData: Any? = if (encryptionHeader == "AES-256-GCM") {
                    // Decrypt response
                    try {
                        val decrypted = CryptoUtils.aes256GcmDecryptCombined(creds.clientSecret, responseBody)
                        parseJson(decrypted)
                    } catch (e: Exception) {
                        // If decryption fails, return raw response
                        parseJson(responseBody)
                    }
                } else {
                    // Try to parse as JSON
                    parseJson(responseBody)
                }
                
                connection.disconnect()
                
                if (responseCode in 200..299) {
                    ApiResult(
                        success = true,
                        data = resultData as? T,
                        error = null,
                        statusCode = responseCode
                    )
                } else {
                    val errorMsg = extractErrorMessage(resultData)
                    ApiResult(
                        success = false,
                        data = null,
                        error = errorMsg ?: "HTTP $responseCode",
                        statusCode = responseCode
                    )
                }
                
            } catch (e: Exception) {
                ApiResult(
                    success = false,
                    data = null,
                    error = e.message ?: "Unknown error",
                    statusCode = 0
                )
            }
        }
    }
    
    /**
     * Register a device and get credentials.
     * This is typically called once during device setup.
     */
    suspend fun registerDevice(
        deviceId: String,
        firebaseInstallId: String,
        fcmToken: String,
        appVersion: String,
        deviceClass: String
    ): ApiResult<Map<String, Any>> {
        val body = mapOf(
            "deviceId" to deviceId,
            "firebaseInstallId" to firebaseInstallId,
            "fcmToken" to fcmToken,
            "appVersion" to appVersion,
            "deviceClass" to deviceClass
        )
        
        // Registration doesn't require signing (device gets its credentials from server)
        return withContext(Dispatchers.IO) {
            try {
                val url = URL("$baseUrl/v1/device/register")
                val connection = url.openConnection() as HttpURLConnection
                connection.requestMethod = "POST"
                connection.timeout = timeout
                connection.doOutput = true
                connection.setRequestProperty("Content-Type", "application/json")
                
                val bodyString = mapToJson(body)
                connection.outputStream.write(bodyString.toByteArray(Charsets.UTF_8))
                
                val responseCode = connection.responseCode
                val responseBody = connection.inputStream.bufferedReader().use { it.readText() }
                
                @Suppress("UNCHECKED_CAST")
                val data = parseJson(responseBody) as? Map<String, Any>
                
                connection.disconnect()
                
                if (responseCode in 200..299 && data != null) {
                    // Store credentials
                    val creds = ClientCredentials(
                        clientId = data["deviceId"] as? String ?: deviceId,
                        clientSecret = data["commandSecret"] as? String ?: "",
                        deviceId = deviceId,
                        commandSecret = data["commandSecret"] as? String
                    )
                    this@SignedApiClient.credentials = creds
                    
                    ApiResult(
                        success = true,
                        data = data,
                        error = null,
                        statusCode = responseCode
                    )
                } else {
                    ApiResult(
                        success = false,
                        data = null,
                        error = "Registration failed",
                        statusCode = responseCode
                    )
                }
            } catch (e: Exception) {
                ApiResult(
                    success = false,
                    data = null,
                    error = e.message ?: "Unknown error",
                    statusCode = 0
                )
            }
        }
    }
    
    /**
     * Get device status (public endpoint, no signing required).
     */
    suspend fun getDeviceStatus(deviceId: String): ApiResult<Map<String, Any>> {
        return withContext(Dispatchers.IO) {
            try {
                val url = URL("$baseUrl/v1/device/${URLEncoder.encode(deviceId, "UTF-8")}/status")
                val connection = url.openConnection() as HttpURLConnection
                connection.requestMethod = "GET"
                connection.timeout = timeout
                
                val responseCode = connection.responseCode
                val responseBody = connection.inputStream.bufferedReader().use { it.readText() }
                
                @Suppress("UNCHECKED_CAST")
                val data = parseJson(responseBody) as? Map<String, Any>
                
                connection.disconnect()
                
                if (responseCode in 200..299 && data != null) {
                    ApiResult(
                        success = true,
                        data = data,
                        error = null,
                        statusCode = responseCode
                    )
                } else {
                    ApiResult(
                        success = false,
                        data = null,
                        error = "Failed to get device status",
                        statusCode = responseCode
                    )
                }
            } catch (e: Exception) {
                ApiResult(
                    success = false,
                    data = null,
                    error = e.message ?: "Unknown error",
                    statusCode = 0
                )
            }
        }
    }
    
    /**
     * Send a telemetry frame to the server.
     */
    suspend fun sendTelemetry(telemetry: Map<String, Any?>): ApiResult<Map<String, Any>> {
        val body = mapOf(
            "type" to "telemetry"
        ) + telemetry.filterValues { it != null }
        return post("/v1/telemetry", body)
    }
    
    // ================== Helper Methods ==================
    
    private fun mapToJson(map: Map<String, Any?>): String {
        val sb = StringBuilder()
        sb.append("{")
        map.entries.forEachIndexed { index, (key, value) ->
            if (index > 0) sb.append(",")
            sb.append("\"$key\":")
            appendJsonValue(sb, value)
        }
        sb.append("}")
        return sb.toString()
    }
    
    private fun appendJsonValue(sb: StringBuilder, value: Any?) {
        when (value) {
            null -> sb.append("null")
            is String -> sb.append("\"${value.replace("\\", "\\\\").replace("\"", "\\\"")}\"")
            is Number -> sb.append(value)
            is Boolean -> sb.append(value)
            is Map<*, *> -> sb.append(mapToJson(value as Map<String, Any?>))
            is List<*> -> sb.append(listToJson(value as List<Any?>))
            else -> sb.append("\"${value.toString().replace("\\", "\\\\").replace("\"", "\\\"")}\"")
        }
    }
    
    private fun listToJson(list: List<Any?>): String {
        val sb = StringBuilder()
        sb.append("[")
        list.forEachIndexed { index, value ->
            if (index > 0) sb.append(",")
            appendJsonValue(sb, value)
        }
        sb.append("]")
        return sb.toString()
    }
    
    private fun parseJson(json: String): Any? {
        if (json.isBlank()) return null
        return jsonReader.parse(json)
    }
    
    private fun extractErrorMessage(data: Any?): String? {
        if (data is Map<*, *>) {
            return (data["message"] as? String) ?: (data["error"] as? String)
        }
        return null
    }
}

/**
 * Simple JSON parser for basic response parsing.
 * For production, use a proper JSON library like Gson or Moshi.
 */
private val jsonReader = object {
    fun parse(json: String): Any? {
        val trimmed = json.trim()
        if (trimmed.startsWith("{")) return parseObject(trimmed)
        if (trimmed.startsWith("[")) return parseArray(trimmed)
        return null
    }
    
    private fun parseObject(json: String): Map<String, Any?> {
        val result = mutableMapOf<String, Any?>()
        val content = json.substringAfter("{").substringBeforeLast("}")
        if (content.isBlank()) return result
        
        var i = 0
        while (i < content.length) {
            // Skip whitespace
            while (i < content.length && content[i].isWhitespace()) i++
            if (i >= content.length) break
            
            // Parse key
            if (content[i] != '"') break
            val (key, newI) = parseString(content, i + 1)
            i = newI
            
            // Skip to colon
            while (i < content.length && content[i] != ':') i++
            i++ // skip colon
            
            // Skip whitespace
            while (i < content.length && content[i].isWhitespace()) i++
            
            // Parse value
            val (value, nextI) = parseValue(content, i)
            result[key] = value
            i = nextI
            
            // Skip comma
            while (i < content.length && content[i] != ',') i++
            i++
        }
        return result
    }
    
    private fun parseArray(json: String): List<Any?> {
        val result = mutableListOf<Any?>()
        val content = json.substringAfter("[").substringBeforeLast("]")
        if (content.isBlank()) return result
        
        var i = 0
        while (i < content.length) {
            while (i < content.length && content[i].isWhitespace()) i++
            if (i >= content.length) break
            
            val (value, nextI) = parseValue(content, i)
            result.add(value)
            i = nextI
            
            while (i < content.length && content[i] != ',') i++
            i++
        }
        return result
    }
    
    private fun parseString(json: String, start: Int): Pair<String, Int> {
        val sb = StringBuilder()
        var i = start
        while (i < json.length) {
            when (json[i]) {
                '\\' -> {
                    i++
                    if (i < json.length) {
                        sb.append(json[i])
                        i++
                    }
                }
                '"' -> return Pair(sb.toString(), i + 1)
                else -> {
                    sb.append(json[i])
                    i++
                }
            }
        }
        return Pair(sb.toString(), i)
    }
    
    private fun parseValue(json: String, start: Int): Pair<Any?, Int> {
        var i = start
        while (i < json.length && json[i].isWhitespace()) i++
        if (i >= json.length) return Pair(null, i)
        
        when {
            json[i] == '"' -> {
                val (value, nextI) = parseString(json, i + 1)
                return Pair(value, nextI)
            }
            json[i] == '{' -> {
                val end = findMatchingBrace(json, i)
                val obj = parseObject(json.substring(i, end + 1))
                return Pair(obj, end + 1)
            }
            json[i] == '[' -> {
                val end = findMatchingBracket(json, i)
                val arr = parseArray(json.substring(i, end + 1))
                return Pair(arr, end + 1)
            }
            json.substring(i).startsWith("null") -> return Pair(null, i + 4)
            json.substring(i).startsWith("true") -> return Pair(true, i + 4)
            json.substring(i).startsWith("false") -> return Pair(false, i + 5)
            else -> {
                // Parse number
                val startNum = i
                while (i < json.length && !json[i].isWhitespace() && json[i] !== ',' && json[i] !== '}') i++
                val numStr = json.substring(startNum, i)
                return try {
                    if (numStr.contains('.')) {
                        Pair(numStr.toDouble(), i)
                    } else {
                        Pair(numStr.toLong(), i)
                    }
                } catch (e: NumberFormatException) {
                    Pair(numStr, i)
                }
            }
        }
    }
    
    private fun findMatchingBrace(json: String, start: Int): Int {
        var depth = 1
        var i = start + 1
        while (i < json.length && depth > 0) {
            when (json[i]) {
                '{' -> depth++
                '}' -> depth--
            }
            i++
        }
        return i - 1
    }
    
    private fun findMatchingBracket(json: String, start: Int): Int {
        var depth = 1
        var i = start + 1
        while (i < json.length && depth > 0) {
            when (json[i]) {
                '[' -> depth++
                ']' -> depth--
            }
            i++
        }
        return i - 1
    }
}
