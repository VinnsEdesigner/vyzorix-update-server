import { describe, it, expect, beforeEach } from 'vitest';
import { useCommandQueueStore, type QueuedCommand } from '@/stores/command-queue-store';

function makeCommand(overrides: Partial<QueuedCommand> = {}): Omit<QueuedCommand, 'queuedAt' | 'attempts' | 'nextRetryAt' | 'status'> {
  return {
    id: 'cq-1',
    imei: '123',
    commandType: 'FORCE_SPEAKER',
    organizationId: 'org-1',
    ...overrides,
  };
}

describe('useCommandQueueStore', () => {
  beforeEach(() => {
    useCommandQueueStore.setState({ queue: [], isFlushing: false, lastFlushError: null });
  });

  it('starts empty', () => {
    const state = useCommandQueueStore.getState();
    expect(state.queue).toHaveLength(0);
    expect(state.isFlushing).toBe(false);
  });

  it('enqueue adds a command with queued status', () => {
    useCommandQueueStore.getState().enqueue(makeCommand());
    const state = useCommandQueueStore.getState();
    expect(state.queue).toHaveLength(1);
    expect(state.queue[0]?.status).toBe('queued');
    expect(state.queue[0]?.attempts).toBe(0);
  });

  it('dequeue removes a command by id', () => {
    useCommandQueueStore.getState().enqueue(makeCommand({ id: 'cq-1' }));
    useCommandQueueStore.getState().enqueue(makeCommand({ id: 'cq-2' }));
    useCommandQueueStore.getState().dequeue('cq-1');
    expect(useCommandQueueStore.getState().queue).toHaveLength(1);
    expect(useCommandQueueStore.getState().queue[0]?.id).toBe('cq-2');
  });

  it('markDispatched sets status and dispatchId', () => {
    useCommandQueueStore.getState().enqueue(makeCommand({ id: 'cq-1' }));
    useCommandQueueStore.getState().markDispatched('cq-1', 'disp-1');
    const entry = useCommandQueueStore.getState().queue[0];
    expect(entry?.status).toBe('dispatched');
    expect(entry?.dispatchId).toBe('disp-1');
  });

  it('markFailed increments attempts and sets backoff', () => {
    useCommandQueueStore.getState().enqueue(makeCommand({ id: 'cq-1' }));
    useCommandQueueStore.getState().markFailed('cq-1', 'network error');
    const entry = useCommandQueueStore.getState().queue[0];
    expect(entry?.attempts).toBe(1);
    expect(entry?.status).toBe('queued');
    expect(entry?.nextRetryAt).toBeGreaterThan(Date.now());
    expect(entry?.lastError).toBe('network error');
  });

  it('markFailed marks as failed after max attempts', () => {
    useCommandQueueStore.getState().enqueue(makeCommand({ id: 'cq-1' }));
    for (let i = 0; i < 5; i++) {
      useCommandQueueStore.getState().markFailed('cq-1', 'err');
    }
    const entry = useCommandQueueStore.getState().queue[0];
    expect(entry?.status).toBe('failed');
  });

  it('getQueueForOrganization filters by org and queued status only', () => {
    useCommandQueueStore.getState().enqueue(makeCommand({ id: 'cq-1', organizationId: 'org-1' }));
    useCommandQueueStore.getState().enqueue(makeCommand({ id: 'cq-2', organizationId: 'org-2' }));
    useCommandQueueStore.getState().markDispatched('cq-1', 'disp-1');
    const org1Queued = useCommandQueueStore.getState().getQueueForOrganization('org-1');
    expect(org1Queued).toHaveLength(0);
    const org2Queued = useCommandQueueStore.getState().getQueueForOrganization('org-2');
    expect(org2Queued).toHaveLength(1);
  });

  it('clearForOrganization removes only that org commands', () => {
    useCommandQueueStore.getState().enqueue(makeCommand({ id: 'cq-1', organizationId: 'org-1' }));
    useCommandQueueStore.getState().enqueue(makeCommand({ id: 'cq-2', organizationId: 'org-2' }));
    useCommandQueueStore.getState().clearForOrganization('org-1');
    const queue = useCommandQueueStore.getState().queue;
    expect(queue).toHaveLength(1);
    expect(queue[0]?.organizationId).toBe('org-2');
  });
});
