export interface VyzorAnalyticsConfig {
  apiKey: string | undefined;
  apiHost: string | undefined;
  doNotTrack: boolean;
  enabled: boolean;
}

export function readVyzorAnalyticsConfig(): VyzorAnalyticsConfig {
  const envDoNotTrack =
    import.meta.env.VITE_DO_NOT_TRACK === 'true' ||
    import.meta.env.VITE_ANALYTICS_ENABLED === 'false';

  const browserDnt =
    typeof navigator !== 'undefined' &&
    (navigator.doNotTrack === '1' ||
      (navigator as unknown as { doNotTrack?: string }).doNotTrack === '1');

  const doNotTrack = envDoNotTrack || browserDnt;
  const apiKey = import.meta.env.VITE_ANALYTICS_KEY as string | undefined;
  const apiHost = (import.meta.env.VITE_ANALYTICS_HOST as string | undefined) ?? undefined;

  return {
    apiKey,
    apiHost,
    doNotTrack,
    enabled: !doNotTrack && Boolean(apiKey),
  };
}
