import { describe, it, expect, beforeEach } from 'vitest';
import { useDeviceSelectorStore, type SelectedDevice } from '@/stores/device-selector-store';

const device: SelectedDevice = { id: 'dev-1', imei: '123456789012345', name: 'Sensor A' };

describe('useDeviceSelectorStore', () => {
  beforeEach(() => {
    useDeviceSelectorStore.setState({
      selectedDevice: null,
      filters: { status: 'all', search: '' },
    });
  });

  it('starts with no selection and default filters', () => {
    const state = useDeviceSelectorStore.getState();
    expect(state.selectedDevice).toBeNull();
    expect(state.filters).toEqual({ status: 'all', search: '' });
  });

  it('selectDevice stores the selected device', () => {
    useDeviceSelectorStore.getState().selectDevice(device);
    expect(useDeviceSelectorStore.getState().selectedDevice).toEqual(device);
  });

  it('clearSelection resets to null', () => {
    useDeviceSelectorStore.getState().selectDevice(device);
    useDeviceSelectorStore.getState().clearSelection();
    expect(useDeviceSelectorStore.getState().selectedDevice).toBeNull();
  });

  it('setFilter merges partial filters', () => {
    useDeviceSelectorStore.getState().setFilter({ status: 'online' });
    expect(useDeviceSelectorStore.getState().filters.status).toBe('online');
    expect(useDeviceSelectorStore.getState().filters.search).toBe('');
  });

  it('setSearch updates search filter', () => {
    useDeviceSelectorStore.getState().setSearch('sensor');
    expect(useDeviceSelectorStore.getState().filters.search).toBe('sensor');
  });

  it('resetFilters restores defaults', () => {
    useDeviceSelectorStore.getState().setFilter({ status: 'offline' });
    useDeviceSelectorStore.getState().setSearch('query');
    useDeviceSelectorStore.getState().resetFilters();
    expect(useDeviceSelectorStore.getState().filters).toEqual({ status: 'all', search: '' });
  });

  it('setFilter preserves other filter keys', () => {
    useDeviceSelectorStore.getState().setSearch('abc');
    useDeviceSelectorStore.getState().setFilter({ status: 'online' });
    expect(useDeviceSelectorStore.getState().filters).toEqual({ status: 'online', search: 'abc' });
  });
});
