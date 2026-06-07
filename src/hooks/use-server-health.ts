import { useQuery } from "@tanstack/react-query";

import { getHealth } from "@/lib/vyzorix-api";

// eslint-disable-next-line func-style
// eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
// eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
// eslint-disable-next-line func-style
// eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
// eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
// eslint-disable-next-line func-style
// eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
// eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
// eslint-disable-next-line func-style
// eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
// eslint-disable-next-line func-style
export function useServerHealth(serverUrl: string) {
  return useQuery({
    queryKey: ["vyzorix", "health", serverUrl],
    queryFn: () => getHealth(serverUrl),
    refetchInterval: 10_000,
    retry: false,
  });
}
