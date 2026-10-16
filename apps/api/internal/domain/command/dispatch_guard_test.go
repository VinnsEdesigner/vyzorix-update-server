package command_test

import (
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
)

func TestDispatchGuardSerializesOneDevice(t *testing.T) {
	g := command.NewDispatchGuard(1)
	if !g.Acquire("dev-1") {
		t.Fatal("first acquire must succeed")
	}
	if g.Acquire("dev-1") {
		t.Fatal("a second concurrent acquire for the same device must be refused")
	}
	if !g.Held("dev-1") {
		t.Fatal("device must report held while acquired")
	}

	g.Release("dev-1")
	if g.Held("dev-1") {
		t.Fatal("device must not be held after release")
	}
	if !g.Acquire("dev-1") {
		t.Fatal("acquire must succeed again after release")
	}
}

func TestDispatchGuardIsPerDevice(t *testing.T) {
	g := command.NewDispatchGuard(1)
	if !g.Acquire("dev-1") {
		t.Fatal("dev-1 acquire failed")
	}
	if !g.Acquire("dev-2") {
		t.Fatal("a different device must not be blocked by dev-1")
	}
}

func TestDispatchGuardReleaseIsIdempotent(t *testing.T) {
	g := command.NewDispatchGuard(1)
	g.Release("never-held")
	if g.Held("never-held") {
		t.Fatal("releasing an unheld device must be a no-op")
	}
	g.Acquire("dev-1")
	g.Release("dev-1")
	g.Release("dev-1")
	if g.Held("dev-1") {
		t.Fatal("double release must not leave the device held")
	}
}

func TestDispatchResourceNamesTheHeldSlot(t *testing.T) {
	if got := command.DispatchResource("dev-9"); got != "dispatch:dev-9" {
		t.Fatalf("resource = %q", got)
	}
}
