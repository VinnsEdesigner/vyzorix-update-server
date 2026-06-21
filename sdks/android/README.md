# Vyzorix Signing SDK for Android

Kotlin library for implementing request signing and response encryption in Android apps.

## Overview

This SDK implements the Vyzorix API security protocol:

- **Request Signing**: HMAC-SHA512 signatures prevent unauthorized access
- **Request Encryption**: AES-256-GCM encryption protects sensitive data in transit
- **Response Decryption**: Automatic decryption of server responses
- **Replay Protection**: Timestamp validation prevents replay attacks

## Installation

### Gradle

```kotlin
// settings.gradle.kts
dependencyResolutionManagement {
    repositories {
        maven { url = uri("https://jitpack.io") }
    }
}

// app/build.gradle.kts
dependencies {
    implementation("com.vyzorix:signing:1.0.0")
}
```

### Manual

Copy the `signing` package into your Android project:
```
your-app/src/main/java/com/vyzorix/signing/
├── CryptoUtils.kt
└── SignedApiClient.kt
```

## Usage

### Initialize the Client

```kotlin
import com.vyzorix.signing.*

// Create client instance
val apiClient = SignedApiClient("https://api.vyzorix.com")

// After device registration (credentials from server)
val credentials = ClientCredentials(
    clientId = "device-123",
    clientSecret = "your-device-secret",
    deviceId = "device-123"
)
apiClient.setCredentials(credentials)
```

### Device Registration

```kotlin
// Registration is typically done once during device setup
suspend fun registerDevice() {
    val result = apiClient.registerDevice(
        deviceId = "device-123",
        firebaseInstallId = "firebase-install-id",
        fcmToken = "fcm-token",
        appVersion = "1.0.0",
        deviceClass = "audio_router"
    )
    
    if (result.success) {
        // Credentials are automatically stored
        // apiClient.getCredentials() now returns the stored credentials
        Log.d("Device", "Registered successfully")
    } else {
        Log.e("Device", "Registration failed: ${result.error}")
    }
}
```

### Making Signed Requests

```kotlin
suspend fun sendCommand() {
    // All requests are automatically:
    // 1. Signed with HMAC-SHA512
    // 2. Body encrypted with AES-256-GCM
    // 3. Response decrypted (if encrypted)
    
    val result = apiClient.post<Map<String, Any>>(
        "/v1/device/device-123/command",
        mapOf(
            "command" to "FORCE_SPEAKER",
            "args" to emptyMap<String, Any>()
        )
    )
    
    when {
        result.success -> {
            val dispatchId = result.data?.get("dispatchId")
            Log.d("Command", "Sent: $dispatchId")
        }
        else -> {
            Log.e("Command", "Failed: ${result.error}")
        }
    }
}
```

### Get Device Status

```kotlin
suspend fun checkStatus() {
    // Public endpoint - no signing required
    val result = apiClient.getDeviceStatus("device-123")
    
    if (result.success) {
        val online = result.data?.get("online")
        val lastSeen = result.data?.get("lastSeen")
        Log.d("Status", "Online: $online, Last seen: $lastSeen")
    }
}
```

### Send Telemetry

```kotlin
suspend fun sendTelemetry() {
    val telemetry = apiClient.sendTelemetry(
        mapOf(
            "deviceId" to "device-123",
            "uptime" to 3600000L,
            "riskScore" to 25,
            "thermalTemp" to 45.5,
            "speakerOn" to true
        )
    )
    
    if (telemetry.success) {
        Log.d("Telemetry", "Sent successfully")
    }
}
```

### Handle Errors

```kotlin
suspend fun safeCommand() {
    val result = apiClient.post<Map<String, Any>>(
        "/v1/device/device-123/command",
        mapOf("command" to "REQUEST_STATUS")
    )
    
    when (result.statusCode) {
        200 -> {
            // Success
        }
        401 -> {
            // Authentication error - credentials may be invalid
            // Consider re-registering the device
            handleAuthError()
        }
        429 -> {
            // Rate limited - wait and retry
            delay(5000)
            retryCommand()
        }
        else -> {
            Log.e("API", "Error: ${result.error}")
        }
    }
}
```

### Clear Credentials on Logout

```kotlin
fun logout() {
    apiClient.clearCredentials()
    // Clear any locally stored credential references
}
```

## How It Works

### Request Flow

```
1. Client builds request body (JSON)
2. Client encrypts body: AES-256-GCM(clientSecret, body)
3. Client creates signature string:
   METHOD\nPATH\nTIMESTAMP\nSHA512(body)
4. Client computes HMAC-SHA512(clientSecret, stringToSign)
5. Client sends request with headers:
   - X-Client-ID: device-123
   - X-Timestamp: 1699999999
   - X-Signature: t=1699999999,v1=abc123...
   - X-Encrypted-Body: <base64 nonce>.<base64 ciphertext>
```

### Response Flow

```
1. Server verifies signature
2. Server decrypts request body
3. Server processes request
4. Server encrypts response: AES-256-GCM(clientSecret, response)
5. Client receives encrypted response
6. Client checks X-Content-Encryption header
7. Client decrypts: AES-256-GCM.Open(clientSecret, nonce, ciphertext)
```

### Security Features

| Feature | Protection |
|---------|------------|
| HMAC-SHA512 Signing | Ensures request authenticity |
| AES-256-GCM Encryption | Protects data in transit |
| Timestamp Validation | ±5 minute window prevents replay |
| Nonce in Encryption | Each encryption uses unique nonce |

## API Reference

### SignedApiClient

| Method | Description |
|--------|-------------|
| `setCredentials(creds)` | Set client credentials |
| `getCredentials()` | Get current credentials |
| `clearCredentials()` | Clear credentials (logout) |
| `hasCredentials()` | Check if credentials exist |
| `registerDevice(...)` | Register device and get credentials |
| `getDeviceStatus(id)` | Get device status (public) |
| `get(path)` | Signed GET request |
| `post(path, body)` | Signed POST request |
| `patch(path, body)` | Signed PATCH request |
| `delete(path)` | Signed DELETE request |
| `sendTelemetry(telemetry)` | Send telemetry data |

### ClientCredentials

| Field | Description |
|-------|-------------|
| `clientId` | Unique client identifier |
| `clientSecret` | Secret key for signing/encryption |
| `deviceId` | Device identifier (optional) |
| `commandSecret` | Device command secret (optional) |

### ApiResult<T>

| Field | Description |
|-------|-------------|
| `success` | Whether the request succeeded |
| `data` | Parsed response data |
| `error` | Error message if failed |
| `statusCode` | HTTP status code |

## License

MIT License - See LICENSE file for details
