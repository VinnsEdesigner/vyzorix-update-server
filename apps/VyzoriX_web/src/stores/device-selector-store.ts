import { createVyzorStore } from '@/lib/state';

export interface DeviceFilters {
  status?: 'online' | 'offline' | 'all';
  search?: string;
}

export interface SelectedDevice {
  id: string;
  imei?: string;
  name?: string;
}

export interface DeviceSelectorState {
  selectedDevice: SelectedDevice | null;
  filters: DeviceFilters;
  selectDevice: (device: SelectedDevice) => void;
  clearSelection: () => void;
  setFilter: (filters: Partial<DeviceFilters>) => void;
  setSearch: (search: string) => void;
  resetFilters: () => void;
}

export const useDeviceSelectorStore = createVyzorStore<DeviceSelectorState>('DeviceSelectorStore', (set) => ({
  selectedDevice: null,
  filters: { status: 'all', search: '' },
  selectDevice: (device) => set({ selectedDevice: device }),
  clearSelection: () => set({ selectedDevice: null }),
  setFilter: (filters) => set((state) => ({ filters: { ...state.filters, ...filters } })),
  setSearch: (search) => set((state) => ({ filters: { ...state.filters, search } })),
  resetFilters: () => set({ filters: { status: 'all', search: '' } }),
}));
