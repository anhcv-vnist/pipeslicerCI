import { useEffect, useState } from 'react';
import { AlertCircle, CheckCircle2, RefreshCw } from 'lucide-react';
import { cn } from '@/lib/utils';

interface RegistryStatusProps {
  registryId: number;
  className?: string;
  onStatusChange?: (isOnline: boolean) => void;
}

interface ConnectionStatus {
  status: 'connected' | 'error' | 'connecting';
  message: string;
  time: string;
}

export function RegistryStatus({ registryId, className, onStatusChange }: RegistryStatusProps) {
  const [status, setStatus] = useState<ConnectionStatus | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [wsError, setWsError] = useState<string | null>(null);

  useEffect(() => {
    console.log(`Initializing WebSocket connection for registry ${registryId}`);
    
    // Get the API base URL from environment or default to localhost
    const apiBaseUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
    
    // Convert HTTP URL to WebSocket URL
    const wsBaseUrl = apiBaseUrl.replace(/^http/, 'ws');
    const wsUrl = `${wsBaseUrl}/registries/${registryId}/test-connection-ws`;
    
    console.log(`Connecting to WebSocket: ${wsUrl}`);
    
    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      console.log(`WebSocket connection opened for registry ${registryId}`);
      setWsError(null);
      // Set initial status to connecting
      setStatus({
        status: 'connecting',
        message: 'Testing connection...',
        time: new Date().toISOString(),
      });
    };

    ws.onmessage = (event) => {
      console.log(`Received message for registry ${registryId}:`, event.data);
      try {
        const data = JSON.parse(event.data);
        console.log(`Parsed status data for registry ${registryId}:`, data);
        setStatus(data);
        const newIsConnected = data.status === 'connected';
        setIsConnected(newIsConnected);
        onStatusChange?.(newIsConnected);
      } catch (error) {
        console.error(`Error parsing WebSocket message for registry ${registryId}:`, error);
        setWsError('Failed to parse status message');
        setIsConnected(false);
        onStatusChange?.(false);
      }
    };

    ws.onerror = (error) => {
      console.error(`WebSocket error for registry ${registryId}:`, error);
      setWsError('Connection error');
      setStatus({
        status: 'error',
        message: 'Failed to connect to WebSocket',
        time: new Date().toISOString(),
      });
      setIsConnected(false);
      onStatusChange?.(false);
    };

    ws.onclose = (event) => {
      console.log(`WebSocket connection closed for registry ${registryId}:`, event.code, event.reason);
      if (!event.wasClean) {
        setWsError('Connection closed unexpectedly');
        setStatus({
          status: 'error',
          message: 'Connection closed unexpectedly',
          time: new Date().toISOString(),
        });
        setIsConnected(false);
        onStatusChange?.(false);
      }
    };

    return () => {
      console.log(`Cleaning up WebSocket connection for registry ${registryId}`);
      ws.close();
    };
  }, [registryId, onStatusChange]);

  return (
    <div className={cn('flex items-center gap-2', className)}>
      {status?.status === 'connected' ? (
        <CheckCircle2 className="h-4 w-4 text-green-500" />
      ) : status?.status === 'connecting' ? (
        <RefreshCw className="h-4 w-4 text-yellow-500 animate-spin" />
      ) : (
        <AlertCircle className="h-4 w-4 text-red-500" />
      )}
      <span className="text-sm text-muted-foreground">
        {wsError ? wsError : (
          status?.status === 'connected' ? 'Online' :
          status?.status === 'connecting' ? 'Connecting...' :
          'Offline'
        )}
      </span>
    </div>
  );
} 