'use client';

import { useState, useEffect, useRef } from 'react';
import { Button } from '@/components/ui/button';
import { useToast } from '@/hooks/use-toast';
import { Loader2, Check, X, Wifi, WifiOff } from 'lucide-react';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { Badge } from '@/components/ui/badge';

interface RegistryStatusProps {
  registryId: number;
  registryName: string;
}

export function RegistryStatus({ registryId, registryName }: RegistryStatusProps) {
  const [status, setStatus] = useState<'idle' | 'connecting' | 'connected' | 'error'>('idle');
  const [lastChecked, setLastChecked] = useState<string | null>(null);
  const [message, setMessage] = useState<string>('');
  const [isTesting, setIsTesting] = useState(false);
  const [isLiveMonitoring, setIsLiveMonitoring] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const { toast } = useToast();
  const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

  // Cleanup WebSocket connection when component unmounts
  useEffect(() => {
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, []);

  // Handle WebSocket messages
  const handleWebSocketMessage = (event: MessageEvent) => {
    try {
      const data = JSON.parse(event.data);
      console.log('WebSocket message:', data);
      
      // Update status based on message
      if (data.status === 'error') {
        setStatus('error');
      } else if (data.status === 'connecting') {
        setStatus('connecting');
      } else if (data.status === 'connected' || data.status === 'success') {
        setStatus('connected');
      }

      // Update message and last checked time
      setMessage(data.message || '');
      if (data.time) {
        setLastChecked(data.time);
      } else {
        setLastChecked(new Date().toISOString());
      }
    } catch (error) {
      console.error('Error parsing WebSocket message:', error);
      setStatus('error');
      setMessage('Failed to parse server response');
      setLastChecked(new Date().toISOString());
    }
  };

  // Test connection with WebSocket
  const testConnection = () => {
    setIsTesting(true);
    setStatus('connecting');
    setMessage('Connecting to registry...');

    // Close existing connection if any
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    // Create new WebSocket connection
    try {
      const wsUrl = `${API_BASE_URL.replace('http', 'ws')}/registries/${registryId}/test-connection-ws`;
      wsRef.current = new WebSocket(wsUrl);

      wsRef.current.onopen = () => {
        console.log('WebSocket connection opened');
      };

      wsRef.current.onmessage = handleWebSocketMessage;

      wsRef.current.onerror = (error) => {
        console.error('WebSocket error:', error);
        setStatus('error');
        setMessage('Failed to connect to WebSocket');
        setLastChecked(new Date().toISOString());
        setIsTesting(false);
        setIsLiveMonitoring(false);
      };

      wsRef.current.onclose = () => {
        console.log('WebSocket connection closed');
        if (isLiveMonitoring) {
          setIsLiveMonitoring(false);
        }
        setIsTesting(false);
      };

      // Set a timeout to stop testing after 10 seconds if no response
      setTimeout(() => {
        if (status === 'connecting' && isTesting) {
          setIsTesting(false);
        }
      }, 10000);
    } catch (error) {
      console.error('Error creating WebSocket connection:', error);
      setStatus('error');
      setMessage('Failed to create WebSocket connection');
      setLastChecked(new Date().toISOString());
      setIsTesting(false);
      toast({
        title: 'Connection Error',
        description: 'Failed to connect to the registry server',
        variant: 'destructive',
      });
    }
  };

  // Toggle live monitoring
  const toggleLiveMonitoring = () => {
    if (isLiveMonitoring) {
      // Stop monitoring
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
      setIsLiveMonitoring(false);
    } else {
      // Start monitoring
      testConnection();
      setIsLiveMonitoring(true);
    }
  };

  // Format relative time
  const getRelativeTime = (timestamp: string | null) => {
    if (!timestamp) return 'Never';
    
    try {
      const date = new Date(timestamp);
      const now = new Date();
      const diffMs = now.getTime() - date.getTime();
      const diffSecs = Math.floor(diffMs / 1000);
      
      if (diffSecs < 10) {
        return 'Just now';
      } else if (diffSecs < 60) {
        return `${diffSecs} seconds ago`;
      } else if (diffSecs < 3600) {
        return `${Math.floor(diffSecs / 60)} minutes ago`;
      } else if (diffSecs < 86400) {
        return `${Math.floor(diffSecs / 3600)} hours ago`;
      } else {
        return `${Math.floor(diffSecs / 86400)} days ago`;
      }
    } catch (error) {
      return 'Invalid date';
    }
  };

  return (
    <div className="flex flex-col space-y-2">
      <div className="flex items-center space-x-2">
        {status === 'idle' && (
          <Badge variant="outline" className="text-muted-foreground">
            <Wifi className="h-3 w-3 mr-1" />
            Not Checked
          </Badge>
        )}
        {status === 'connecting' && (
          <Badge variant="outline" className="text-yellow-500 animate-pulse">
            <Loader2 className="h-3 w-3 mr-1 animate-spin" />
            Connecting
          </Badge>
        )}
        {status === 'connected' && (
          <Badge variant="outline" className="text-green-500">
            <Check className="h-3 w-3 mr-1" />
            Connected
          </Badge>
        )}
        {status === 'error' && (
          <Badge variant="outline" className="text-red-500">
            <X className="h-3 w-3 mr-1" />
            Connection Failed
          </Badge>
        )}
        
        {lastChecked && (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <span className="text-xs text-muted-foreground">
                  Last checked: {getRelativeTime(lastChecked)}
                </span>
              </TooltipTrigger>
              <TooltipContent>
                <p>{new Date(lastChecked).toLocaleString()}</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        )}
      </div>
      
      {message && (
        <p className="text-xs text-muted-foreground">{message}</p>
      )}
      
      <div className="flex gap-2">
        <Button 
          size="sm" 
          variant="outline" 
          onClick={testConnection} 
          disabled={isTesting}
          className="h-8 px-2 text-xs"
        >
          {isTesting ? (
            <>
              <Loader2 className="h-3 w-3 mr-1 animate-spin" />
              Testing...
            </>
          ) : (
            <>
              <Wifi className="h-3 w-3 mr-1" />
              Test Connection
            </>
          )}
        </Button>
        
        <Button 
          size="sm" 
          variant={isLiveMonitoring ? "default" : "outline"} 
          onClick={toggleLiveMonitoring}
          className="h-8 px-2 text-xs"
        >
          {isLiveMonitoring ? (
            <>
              <WifiOff className="h-3 w-3 mr-1" />
              Stop Monitoring
            </>
          ) : (
            <>
              <Wifi className="h-3 w-3 mr-1" />
              Live Monitor
            </>
          )}
        </Button>
      </div>
    </div>
  );
} 