import { describe, it, expect, beforeEach } from 'vitest';
import { useCommandDispatchStore, type PendingCommand } from '@/stores/command-dispatch-store';

function makePending(overrides: Partial<PendingCommand> = {}): PendingCommand {
  return {
    dispatchId: 'dispatch-1',
    imei: '123456789012345',
    commandType: 'FORCE_SPEAKER',
    createdAt: Date.now(),
    ...overrides,
  };
}

describe('useCommandDispatchStore', () => {
  beforeEach(() => {
    useCommandDispatchStore.setState({ pendingCommands: {}, pendingCount: 0 });
  });

  it('starts empty', () => {
    const state = useCommandDispatchStore.getState();
    expect(state.pendingCommands).toEqual({});
    expect(state.pendingCount).toBe(0);
  });

  it('addPending stores a command and updates count', () => {
    const cmd = makePending();
    useCommandDispatchStore.getState().addPending(cmd);
    const state = useCommandDispatchStore.getState();
    expect(state.pendingCommands['dispatch-1']).toEqual(cmd);
    expect(state.pendingCount).toBe(1);
  });

  it('addPending replaces existing command with same dispatchId', () => {
    const cmd1 = makePending({ dispatchId: 'dup' });
    const cmd2 = makePending({ dispatchId: 'dup', commandType: 'RESET_AUDIO_HAL' });
    useCommandDispatchStore.getState().addPending(cmd1);
    useCommandDispatchStore.getState().addPending(cmd2);
    const state = useCommandDispatchStore.getState();
    expect(Object.keys(state.pendingCommands)).toHaveLength(1);
    expect(state.pendingCommands['dup']?.commandType).toBe('RESET_AUDIO_HAL');
    expect(state.pendingCount).toBe(1);
  });

  it('removePending deletes a command and updates count', () => {
    useCommandDispatchStore.getState().addPending(makePending({ dispatchId: 'a' }));
    useCommandDispatchStore.getState().addPending(makePending({ dispatchId: 'b' }));
    useCommandDispatchStore.getState().removePending('a');
    const state = useCommandDispatchStore.getState();
    expect(state.pendingCommands['a']).toBeUndefined();
    expect(state.pendingCommands['b']).toBeDefined();
    expect(state.pendingCount).toBe(1);
  });

  it('removePending is a no-op for unknown dispatchId', () => {
    useCommandDispatchStore.getState().addPending(makePending());
    useCommandDispatchStore.getState().removePending('nonexistent');
    expect(useCommandDispatchStore.getState().pendingCount).toBe(1);
  });

  it('clearPending empties the store', () => {
    useCommandDispatchStore.getState().addPending(makePending({ dispatchId: 'a' }));
    useCommandDispatchStore.getState().addPending(makePending({ dispatchId: 'b' }));
    useCommandDispatchStore.getState().clearPending();
    const state = useCommandDispatchStore.getState();
    expect(state.pendingCommands).toEqual({});
    expect(state.pendingCount).toBe(0);
  });

  it('getPending returns the stored command', () => {
    const cmd = makePending({ dispatchId: 'xyz' });
    useCommandDispatchStore.getState().addPending(cmd);
    expect(useCommandDispatchStore.getState().getPending('xyz')).toEqual(cmd);
  });

  it('getPending returns undefined for unknown dispatchId', () => {
    expect(useCommandDispatchStore.getState().getPending('nope')).toBeUndefined();
  });
});
