# Consumer rules for Vyzorix Signing SDK
# These rules are applied when your app depends on this library

# Keep signing classes
-keep class com.vyzorix.signing.** { *; }

# Keep data classes for serialization
-keepclassmembers class com.vyzorix.signing.ClientCredentials {
    <fields>;
}
-keepclassmembers class com.vyzorix.signing.ApiResult {
    <fields>;
}
-keepclassmembers class com.vyzorix.signing.SignedRequestOptions {
    <fields>;
}
-keepclassmembers class com.vyzorix.signing.EncryptionResult {
    <fields>;
}
