import { useQuery, type UseQueryResult } from "@tanstack/react-query";

import { getHealth } from "@/lib/vyzorix-api";

// eslint-disable-next-line func-style
export function useServerHealth(
  serverUrl: string,
): UseQueryResult<{ ok: boolean }, Error> {
  return useQuery({
    queryKey: ["vyzorix", "health", serverUrl],
    queryFn: () => getHealth(serverUrl),
    refetchInterval: 10_000,
    retry: false,
  });
}
