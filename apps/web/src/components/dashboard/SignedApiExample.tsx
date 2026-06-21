/**
 * SignedApiExample.tsx - Example component demonstrating request signing integration.
 * 
 * This shows how to integrate the signed API client into dashboard components.
 * Replace existing API calls with signed API calls for cross-origin requests.
 * 
 * NOTE: For same-origin requests, cookie-based auth is typically sufficient.
 * Request signing is recommended for:
 * - Cross-origin API calls
 * - Mobile app backends
 * - Third-party integrations
 */

import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { toast } from 'sonner';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useSignedApi } from '@/hooks/use-signed-api';

/**
 * Example: Signed API integration for device commands.
 * This demonstrates how to use the signed API client in a dashboard component.
 */
export function SignedCommandExample() {
  const navigate = useNavigate();
  const [commandResult, setCommandResult] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const { 
    get, 
    post, 
    hasCredentials, 
    error,
    fetchCredentials,
  } = useSignedApi({
    apiUrl: window.location.origin,
    onReauthNeeded: () => {
      toast.error('Session expired. Please login again.');
      navigate('/login');
    },
    onKeyRotationNeeded: (clientId) => {
      toast.warning('API key rotated. Re-authenticating...');
    },
  });

  // Try to fetch credentials on mount if not present
  useEffect(() => {
    if (!hasCredentials) {
      fetchCredentials().catch((err) => {
        console.warn('No signed API credentials available:', err);
      });
    }
  }, [hasCredentials, fetchCredentials]);

  const handleSendCommand = useCallback(async (deviceId: string, command: string) => {
    setLoading(true);
    setCommandResult(null);
    try {
      // Make a signed POST request
      // The request will be:
      // 1. Signed with HMAC-SHA512
      // 2. Body encrypted with AES-256-GCM
      // 3. Response encrypted (if server mandates)
      const result = await post<{ success: boolean; message: string }>(
        `/v1/device/${encodeURIComponent(deviceId)}/command`,
        { command, args: {}, timestamp: Date.now() }
      );
      setCommandResult(`Command sent: ${result.message}`);
      toast.success('Command sent successfully');
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Command failed';
      setCommandResult(`Error: ${msg}`);
      toast.error(msg);
    } finally {
      setLoading(false);
    }
  }, [post]);

  const handleGetDeviceCount = useCallback(async () => {
    setLoading(true);
    setCommandResult(null);
    try {
      // Make a signed GET request
      const result = await get<{ count: number; devices: unknown[] }>('/v1/device/count');
      setCommandResult(`Device count: ${result.count}`);
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Failed to get device count';
      setCommandResult(`Error: ${msg}`);
      toast.error(msg);
    } finally {
      setLoading(false);
    }
  }, [get]);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Signed API Example</CardTitle>
        <CardDescription>
          Demonstrates request signing for API calls
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Status indicator */}
        <div className="flex items-center gap-2 text-sm">
          <span className={`w-2 h-2 rounded-full ${hasCredentials ? 'bg-green-500' : 'bg-yellow-500'}`} />
          <span>{hasCredentials ? 'Signed API Ready' : 'Signing Not Available'}</span>
        </div>

        {/* Error display */}
        {error && (
          <div className="p-3 bg-destructive/10 border border-destructive/20 rounded-md text-sm text-destructive">
            Error: {error.message}
          </div>
        )}

        {/* Result display */}
        {commandResult && (
          <div className="p-3 bg-muted rounded-md font-mono text-sm">
            {commandResult}
          </div>
        )}

        {/* Action buttons */}
        <div className="flex gap-2">
          <Button 
            onClick={() => handleGetDeviceCount()} 
            disabled={loading || !hasCredentials}
            variant="outline"
          >
            Get Device Count
          </Button>
          <Button 
            onClick={() => handleSendCommand('example-device', 'REQUEST_STATUS')} 
            disabled={loading || !hasCredentials}
          >
            Send Command
          </Button>
        </div>

        {/* Usage instructions */}
        <div className="text-xs text-muted-foreground space-y-1 pt-2 border-t">
          <p><strong>How signing works:</strong></p>
          <p>1. Request body is encrypted with AES-256-GCM</p>
          <p>2. Request is signed with HMAC-SHA512</p>
          <p>3. Server verifies signature and decrypts body</p>
          <p>4. Response is encrypted before sending back</p>
        </div>
      </CardContent>
    </Card>
  );
}

/**
 * Hook-based example for use in TanStack Query components.
 * This shows how to integrate signing into data fetching hooks.
 */
export function useSignedApiQuery() {
  const api = useSignedApi({
    apiUrl: window.location.origin,
    onReauthNeeded: () => {
      console.warn('Re-authentication required');
    },
  });

  return {
    /**
     * Fetch devices using signed API.
     */
    fetchDevices: async () => {
      return api.get<{ devices: unknown[] }>('/v1/dashboard/devices');
    },

    /**
     * Fetch single device using signed API.
     */
    fetchDevice: async (deviceId: string) => {
      return api.get<{ deviceId: string; online: boolean }>(
        `/v1/device/${encodeURIComponent(deviceId)}`
      );
    },

    /**
     * Send command using signed API.
     */
    sendCommand: async (deviceId: string, command: string, args?: Record<string, unknown>) => {
      return api.post<{ dispatchId: string }>(
        `/v1/device/${encodeURIComponent(deviceId)}/command`,
        { command, args: args ?? {} }
      );
    },

    /**
     * Update device using signed API.
     */
    updateDevice: async (deviceId: string, updates: Record<string, unknown>) => {
      return api.patch<{ success: boolean }>(
        `/v1/device/${encodeURIComponent(deviceId)}`,
        updates
      );
    },

    /**
     * Delete device using signed API.
     */
    deleteDevice: async (deviceId: string) => {
      return api.delete<{ success: boolean }>(
        `/v1/device/${encodeURIComponent(deviceId)}`
      );
    },

    // Expose state for UI indicators
    hasCredentials: api.hasCredentials,
    isLoading: api.loading,
    error: api.error,
  };
}
