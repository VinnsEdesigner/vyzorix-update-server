package device

import (
	"testing"
)

func TestDevice_AddTag(t *testing.T) {
	d := NewDevice("test-imei", "firebase-id")
	d.AddTag("production")
	d.AddTag("staging")

	if len(d.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(d.Tags))
	}
	if !d.HasTag("production") {
		t.Error("should have production tag")
	}
	if !d.HasTag("staging") {
		t.Error("should have staging tag")
	}
}

func TestDevice_AddTag_Idempotent(t *testing.T) {
	d := NewDevice("test-imei", "firebase-id")
	d.AddTag("production")
	d.AddTag("production")
	d.AddTag("production")

	if len(d.Tags) != 1 {
		t.Errorf("adding same tag 3x should result in 1 tag, got %d", len(d.Tags))
	}
}

func TestDevice_RemoveTag(t *testing.T) {
	d := NewDevice("test-imei", "firebase-id")
	d.AddTag("production")
	d.AddTag("staging")
	d.AddTag("beta")

	d.RemoveTag("staging")

	if len(d.Tags) != 2 {
		t.Errorf("expected 2 tags after removal, got %d", len(d.Tags))
	}
	if d.HasTag("staging") {
		t.Error("staging should be removed")
	}
	if !d.HasTag("production") {
		t.Error("production should still be present")
	}
}

func TestDevice_RemoveTag_NonExistent(t *testing.T) {
	d := NewDevice("test-imei", "firebase-id")
	d.AddTag("production")

	d.RemoveTag("nonexistent")

	if len(d.Tags) != 1 {
		t.Errorf("removing non-existent tag should not change count, got %d", len(d.Tags))
	}
}

func TestDevice_HasTag_EmptyTags(t *testing.T) {
	d := NewDevice("test-imei", "firebase-id")
	if d.HasTag("anything") {
		t.Error("device with no tags should not have any tag")
	}
}
