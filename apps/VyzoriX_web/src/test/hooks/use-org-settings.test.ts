/**
 * Integration tests for useOrgSettings / useUpdateOrgSettings /
 * useOrgThresholds / useUpdateOrgThresholds.
 *
 * These tests render the REAL hooks via React Testing Library. The hooks call
 * the REAL API client functions (orgSettings.get/update/getThresholds/
 * updateThresholds) which use the REAL restClient (axios). MSW intercepts the
 * HTTP requests to /v1/organizations/:id/settings and returns mock server
 * responses built from shared fixtures.
 *
 * No vi.mock — the real code path runs end-to-end.
 */
import { describe, it, expect, beforeEach } from 'vitest';
import { waitFor } from '@testing-library/react';
import { renderHookWithQueryClient } from '../helpers/render-hook';
import { setupIntegrationTest } from '../helpers/integration-test-setup';
import { useAuthStore } from '@/stores/auth-store';
import {
  useOrgSettings,
  useUpdateOrgSettings,
  useOrgThresholds,
  useUpdateOrgThresholds,
} from '@/hooks/organization/use-org-settings';
import { buildThresholds } from '../fixtures/vyzor-test-fixtures';

setupIntegrationTest();

function setOrg(orgId: string | null) {
  useAuthStore.getState().setOrganization(orgId);
}

beforeEach(() => {
  setOrg(null);
});

describe('useOrgSettings', () => {
  it('is disabled without orgId', () => {
    const { result } = renderHookWithQueryClient(() => useOrgSettings(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('is disabled with empty orgId', () => {
    const { result } = renderHookWithQueryClient(() => useOrgSettings(''));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches org settings via the real API client', async () => {
    setOrg('org-1');
    const { result } = renderHookWithQueryClient(() => useOrgSettings('org-1'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.id).toBeDefined();
    expect(result.current.data?.organizationId).toBe('org-test-1');
    expect(result.current.data?.timezone).toBe('UTC');
    expect(result.current.data?.dateFormat).toBe('YYYY-MM-DD');
    expect(result.current.data?.alertCooldownMinutes).toBe(15);
    expect(result.current.data?.defaultThresholds).toEqual(buildThresholds());
  });
});

describe('useUpdateOrgSettings', () => {
  beforeEach(() => {
    setOrg('org-1');
  });

  it('calls update and caches the patched result', async () => {
    const { result } = renderHookWithQueryClient(() => useUpdateOrgSettings());
    result.current.mutate({ orgId: 'org-1', request: { timezone: 'Europe/Stockholm' } });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.timezone).toBe('Europe/Stockholm');
    expect(result.current.data?.dateFormat).toBe('YYYY-MM-DD');
    expect(result.current.data?.alertCooldownMinutes).toBe(15);
  });
});

describe('useOrgThresholds', () => {
  it('is disabled without orgId', () => {
    const { result } = renderHookWithQueryClient(() => useOrgThresholds(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('is disabled with empty orgId', () => {
    const { result } = renderHookWithQueryClient(() => useOrgThresholds(''));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches org thresholds via the real API client', async () => {
    setOrg('org-1');
    const { result } = renderHookWithQueryClient(() => useOrgThresholds('org-1'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.thresholds).toEqual(buildThresholds());
  });
});

describe('useUpdateOrgThresholds', () => {
  beforeEach(() => {
    setOrg('org-1');
  });

  it('calls updateThresholds and caches the patched result', async () => {
    const { result } = renderHookWithQueryClient(() => useUpdateOrgThresholds('org-1'));
    result.current.mutate({ riskWarn: 75 });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.thresholds?.riskWarn).toBe(75);
    expect(result.current.data?.thresholds?.riskCrit).toBe(buildThresholds().riskCrit);
  });
});
