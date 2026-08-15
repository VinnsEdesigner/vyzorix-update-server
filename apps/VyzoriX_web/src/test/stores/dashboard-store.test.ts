import { describe, it, expect, beforeEach } from 'vitest';
import { useDashboardStore } from '@/stores/dashboard-store';
import type { DashboardStats } from '@vyzorix/api-client';

const STATS: DashboardStats = {
  devices: { total: 10, online: 8, offline: 2 },
  commands: { totalToday: 50, pending: 3, failed: 1 },
  activity: { last24h: { commands: 50, registrations: 2, deregistrations: 1 } },
};

describe('useDashboardStore', () => {
  beforeEach(() => {
    useDashboardStore.setState({
      stats: null,
      lastRefreshedAt: null,
      isRefreshing: false,
      refreshIntervalMs: 30_000,
      isPolling: false,
      recentActivity: [],
      activeOrganizationId: null,
    });
  });

  it('setStats stores stats and sets lastRefreshedAt', () => {
    useDashboardStore.getState().setStats(STATS);
    const state = useDashboardStore.getState();
    expect(state.stats).toEqual(STATS);
    expect(state.lastRefreshedAt).not.toBeNull();
  });

  it('setRefreshing toggles the refreshing flag', () => {
    useDashboardStore.getState().setRefreshing(true);
    expect(useDashboardStore.getState().isRefreshing).toBe(true);
  });

  it('setRefreshInterval updates the interval', () => {
    useDashboardStore.getState().setRefreshInterval(60_000);
    expect(useDashboardStore.getState().refreshIntervalMs).toBe(60_000);
  });

  it('setRefreshInterval ignores non-positive values', () => {
    useDashboardStore.getState().setRefreshInterval(0);
    expect(useDashboardStore.getState().refreshIntervalMs).toBe(30_000);
  });

  it('pushActivity prepends to recentActivity', () => {
    useDashboardStore.getState().pushActivity({ type: 'command_sent', message: 'sent' });
    useDashboardStore.getState().pushActivity({ type: 'log_alert', message: 'alert' });
    const items = useDashboardStore.getState().recentActivity;
    expect(items).toHaveLength(2);
    expect(items[0]?.message).toBe('alert');
  });

  it('pushActivity caps at max items', () => {
    for (let i = 0; i < 110; i++) {
      useDashboardStore.getState().pushActivity({ type: 'metric', message: `m-${i}` });
    }
    expect(useDashboardStore.getState().recentActivity).toHaveLength(100);
  });

  it('setActiveOrganization clears state on org switch', () => {
    useDashboardStore.getState().setActiveOrganization('org-1');
    useDashboardStore.getState().setStats(STATS);
    useDashboardStore.getState().pushActivity({ type: 'command_sent', message: 'x' });
    useDashboardStore.getState().setActiveOrganization('org-2');
    const state = useDashboardStore.getState();
    expect(state.stats).toBeNull();
    expect(state.recentActivity).toHaveLength(0);
    expect(state.activeOrganizationId).toBe('org-2');
  });

  it('setActiveOrganization does NOT clear when org unchanged', () => {
    useDashboardStore.getState().setActiveOrganization('org-1');
    useDashboardStore.getState().setStats(STATS);
    useDashboardStore.getState().setActiveOrganization('org-1');
    expect(useDashboardStore.getState().stats).toEqual(STATS);
  });

  it('clear resets the store', () => {
    useDashboardStore.getState().setStats(STATS);
    useDashboardStore.getState().pushActivity({ type: 'command_sent', message: 'x' });
    useDashboardStore.getState().clear();
    expect(useDashboardStore.getState().stats).toBeNull();
    expect(useDashboardStore.getState().recentActivity).toHaveLength(0);
  });
});
