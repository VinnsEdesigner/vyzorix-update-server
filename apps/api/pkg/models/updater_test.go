package models

import (
	"encoding/json"
	"testing"
)

func TestVersionManifest_JSON(t *testing.T) {
	manifest := VersionManifest{
		Version:      "1.2.3",
		APKFilename:  "vyzorix-app-1.2.3.apk",
		APKSHA256:    "abc123def456789",
		ReleaseNotes: "Bug fixes and improvements",
		VersionCode:  123,
		APKSizeBytes: 52428800, // 50MB
	}

	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled VersionManifest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Version != manifest.Version {
		t.Errorf("Version = %q, want %q", unmarshaled.Version, manifest.Version)
	}
	if unmarshaled.APKFilename != manifest.APKFilename {
		t.Errorf("APKFilename = %q, want %q", unmarshaled.APKFilename, manifest.APKFilename)
	}
	if unmarshaled.APKSHA256 != manifest.APKSHA256 {
		t.Errorf("APKSHA256 = %q, want %q", unmarshaled.APKSHA256, manifest.APKSHA256)
	}
	if unmarshaled.ReleaseNotes != manifest.ReleaseNotes {
		t.Errorf("ReleaseNotes = %q, want %q", unmarshaled.ReleaseNotes, manifest.ReleaseNotes)
	}
	if unmarshaled.VersionCode != manifest.VersionCode {
		t.Errorf("VersionCode = %d, want %d", unmarshaled.VersionCode, manifest.VersionCode)
	}
	if unmarshaled.APKSizeBytes != manifest.APKSizeBytes {
		t.Errorf("APKSizeBytes = %d, want %d", unmarshaled.APKSizeBytes, manifest.APKSizeBytes)
	}
}

func TestVersionManifest_JSONTags(t *testing.T) {
	data := []byte(`{
		"version": "2.0.0",
		"apk_filename": "app-2.0.0.apk",
		"apk_sha256": "sha256hash",
		"release_notes": "New features",
		"version_code": 200,
		"apk_size_bytes": 104857600
	}`)

	var manifest VersionManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if manifest.Version != "2.0.0" {
		t.Errorf("Version = %q, want \"2.0.0\"", manifest.Version)
	}
	if manifest.APKFilename != "app-2.0.0.apk" {
		t.Errorf("APKFilename = %q, want \"app-2.0.0.apk\"", manifest.APKFilename)
	}
}

func TestVersionManifest_EmptyFields(t *testing.T) {
	manifest := VersionManifest{}

	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled VersionManifest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Version != "" {
		t.Errorf("Version = %q, want \"\"", unmarshaled.Version)
	}
	if unmarshaled.APKSizeBytes != 0 {
		t.Errorf("APKSizeBytes = %d, want 0", unmarshaled.APKSizeBytes)
	}
}

func TestVersionManifest_LargeAPK(t *testing.T) {
	hash := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	manifest := VersionManifest{
		Version:      "100.0.0",
		APKFilename:  "vyzorix-100.apk",
		APKSHA256:    hash,
		VersionCode:  1000000,
		APKSizeBytes: 1073741824, // 1GB
	}

	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled VersionManifest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.APKSizeBytes != manifest.APKSizeBytes {
		t.Errorf("APKSizeBytes = %d, want %d", unmarshaled.APKSizeBytes, manifest.APKSizeBytes)
	}
}
