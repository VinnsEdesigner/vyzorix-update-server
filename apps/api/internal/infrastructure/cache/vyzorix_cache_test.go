package cache

import (
	"testing"
	"time"
)

func TestCache_TTLExpiry(t *testing.T) {
	c := New(50 * time.Millisecond)
	c.Set("key1", "value1")

	if _, ok := c.Get("key1"); !ok {
		t.Fatal("expected cache hit before expiry")
	}

	time.Sleep(60 * time.Millisecond)
	if _, ok := c.Get("key1"); ok {
		t.Error("expected miss after TTL expiry")
	}
}

func TestCache_SectionIsolation(t *testing.T) {
	c := New(time.Minute)
	searchSection := c.Section("search")
	dashSection := c.Section("dashboard")

	searchSection.Set("q1", "search-result")
	dashSection.Set("q1", "dash-result")

	if v, ok := searchSection.Get("q1"); !ok || v != "search-result" {
		t.Errorf("search section read wrong: %v", v)
	}
	if v, ok := dashSection.Get("q1"); !ok || v != "dash-result" {
		t.Errorf("dashboard section read wrong: %v", v)
	}
}

func TestCache_SetTTL(t *testing.T) {
	c := New(time.Minute)
	c.SetTTL("key", "value", 10*time.Millisecond)

	time.Sleep(15 * time.Millisecond)
	if _, ok := c.Get("key"); ok {
		t.Error("expected custom TTL expiry")
	}
}

func TestCache_Delete(t *testing.T) {
	c := New(time.Minute)
	c.Set("key", "value")
	c.Delete("key")
	if _, ok := c.Get("key"); ok {
		t.Error("expected miss after delete")
	}
}

func TestCache_Flush(t *testing.T) {
	c := New(time.Minute)
	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Flush()
	if c.Len() != 0 {
		t.Errorf("flush left %d entries", c.Len())
	}
}
