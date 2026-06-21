package updater

import "time"

// VersionManifest represents the firmware version manifest.
type VersionManifest struct {
    Version       string    `json:"version"`
    ReleaseDate   time.Time `json:"releaseDate"`
    ReleaseNotes  string    `json:"releaseNotes,omitempty"`
    DownloadURL   string    `json:"downloadUrl"`
    Checksum      string    `json:"checksum"`
    MinCompatible string    `json:"minCompatible,omitempty"`
    Signature     string    `json:"signature,omitempty"`
    Files         []ManifestFile `json:"files,omitempty"`
}

// ManifestFile represents a file in the version manifest.
type ManifestFile struct {
    Path      string `json:"path"`
    Size      int64  `json:"size"`
    Checksum  string `json:"checksum"`
    Compressed bool   `json:"compressed,omitempty"`
}

// UpdateCheckRequest is the payload for checking for updates.
type UpdateCheckRequest struct {
    CurrentVersion string `json:"currentVersion"`
    DeviceID      string `json:"deviceId,omitempty"`
    HardwareID    string `json:"hardwareId,omitempty"`
}

// UpdateCheckResponse is returned when checking for updates.
type UpdateCheckResponse struct {
    UpdateAvailable bool              `json:"updateAvailable"`
    CurrentVersion  string           `json:"currentVersion"`
    LatestVersion   string           `json:"latestVersion,omitempty"`
    Manifest        *VersionManifest `json:"manifest,omitempty"`
}
