package tests

import (
	"context"
	"sync"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
)

// stubCommandRepoForTests satisfies command.Repository for subscription tests.
// Only the reads the org guards need return real data; every other method is a
// safe no-op so an unexpected call cannot panic a test run.
type stubCommandRepoForTests struct {
	mu       sync.Mutex
	byID     map[string]*command.Command
	commands []*command.Command
}

func newStubCommandRepoForTests() *stubCommandRepoForTests {
	return &stubCommandRepoForTests{byID: map[string]*command.Command{}}
}

// add registers a command so the org guard can resolve its device.
func (r *stubCommandRepoForTests) add(c *command.Command) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[c.DispatchID] = c
	r.commands = append(r.commands, c)
}

func (r *stubCommandRepoForTests) FindByDispatchIDOnly(_ context.Context, dispatchID string) (*command.Command, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.byID[dispatchID]; ok {
		return c, nil
	}
	return nil, nil
}

func (r *stubCommandRepoForTests) FindByDispatchID(_ context.Context, _ string, dispatchID string) (*command.Command, error) {
	return r.FindByDispatchIDOnly(context.Background(), dispatchID)
}

func (r *stubCommandRepoForTests) FindByID(_ context.Context, id string) (*command.Command, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range r.commands {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, nil
}

func (r *stubCommandRepoForTests) FindPending(context.Context, int) ([]*command.Command, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.commands, nil
}

func (r *stubCommandRepoForTests) FindByDeviceID(context.Context, string, int) ([]*command.Command, error) {
	return nil, nil
}

func (r *stubCommandRepoForTests) FindPendingByDeviceID(context.Context, string) ([]*command.Command, error) {
	return nil, nil
}

func (r *stubCommandRepoForTests) FindHistoryByDeviceID(context.Context, string, string, time.Time, time.Time, int, int) ([]*command.Command, int, error) {
	return nil, 0, nil
}

func (r *stubCommandRepoForTests) FindByDispatchPrefix(context.Context, string) ([]*command.Command, error) {
	return nil, nil
}

func (r *stubCommandRepoForTests) Create(context.Context, *command.Command) error { return nil }
func (r *stubCommandRepoForTests) Update(context.Context, *command.Command) error { return nil }
func (r *stubCommandRepoForTests) UpdateStatus(context.Context, string, command.Status) error {
	return nil
}
func (r *stubCommandRepoForTests) Delete(context.Context, string) error           { return nil }
func (r *stubCommandRepoForTests) DeleteByDeviceID(context.Context, string) error { return nil }
func (r *stubCommandRepoForTests) DeleteByDeviceIDs(context.Context, []string) (int64, error) {
	return 0, nil
}
func (r *stubCommandRepoForTests) Count(context.Context) (int, error)        { return 0, nil }
func (r *stubCommandRepoForTests) CountPending(context.Context) (int, error) { return 0, nil }
func (r *stubCommandRepoForTests) CountPendingByDevice(context.Context, string) (int, error) {
	return 0, nil
}
func (r *stubCommandRepoForTests) CountPendingByDeviceIDs(context.Context, []string) (int, error) {
	return 0, nil
}
func (r *stubCommandRepoForTests) MarkWake(context.Context, string, string) error { return nil }
func (r *stubCommandRepoForTests) MarkDelivered(context.Context, string) error    { return nil }
func (r *stubCommandRepoForTests) MarkCompleted(context.Context, string, string) error {
	return nil
}
func (r *stubCommandRepoForTests) MarkFailed(context.Context, string, string) error { return nil }
func (r *stubCommandRepoForTests) DeleteOldCommands(context.Context, int64) (int64, error) {
	return 0, nil
}
func (r *stubCommandRepoForTests) CancelByDispatchPrefix(context.Context, string) (int64, error) {
	return 0, nil
}
func (r *stubCommandRepoForTests) UpdateRetryInfo(context.Context, string, int, int, *time.Time) error {
	return nil
}
