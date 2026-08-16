import { create, type StateCreator, type UseBoundStore, type StoreApi } from 'zustand';
import { persist, type PersistOptions } from 'zustand/middleware';
import { devtools, buildDevtoolsOptions } from './vyzor-store-devtools';

export interface CreateVyzorStoreOptions<T> {
  persist?: PersistOptions<T, Partial<T>>;
  devtoolsName?: string;
}

export function createVyzorStore<T>(
  name: string,
  initializer: StateCreator<T, [], []>,
  options?: CreateVyzorStoreOptions<T>,
): UseBoundStore<StoreApi<T>> {
  const devtoolsName = options?.devtoolsName ?? name;
  const devtoolsOptions = buildDevtoolsOptions(devtoolsName);

  if (options?.persist) {
    return create<T>()(
      persist(devtools(initializer, devtoolsOptions), options.persist),
    );
  }

  return create<T>()(devtools(initializer, devtoolsOptions));
}
