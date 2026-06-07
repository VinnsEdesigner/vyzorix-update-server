package models

// VersionManifest represents the OTA version manifest served to Android clients.
type VersionManifest struct {
	Version      string `json:"version"`
	APKFilename  string `json:"apk_filename"`
	APKSHA256    string `json:"apk_sha256"`
	ReleaseNotes string `json:"release_notes"`
	VersionCode  int    `json:"version_code"`
	APKSizeBytes int64  `json:"apk_size_bytes"`
}
