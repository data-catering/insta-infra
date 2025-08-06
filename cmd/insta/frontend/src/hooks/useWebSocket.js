import { useEffect } from 'react';
// WebSocket client is passed as a prop

export const useWebSocket = ({ setStatuses, setDependencyStatuses, setLastUpdated, wsClient, WS_MSG_TYPES }) => {
  // Initialize WebSocket client and set up real-time event handlers
  const initializeWebSocket = () => {
    try {
      // Connect to WebSocket
      wsClient.connect();
      
      // Subscribe to service status updates for real-time status changes
      wsClient.subscribe(WS_MSG_TYPES.SERVICE_STATUS_UPDATE, (payload) => {
        if (payload && payload.service_name && payload.status) {
          setStatuses(prevStatuses => ({
            ...prevStatuses,
            [payload.service_name]: payload.status
          }));
          
          // Also update dependency statuses to keep them in sync
          setDependencyStatuses(prevDeps => ({
            ...prevDeps,
            [payload.service_name]: payload.status
          }));
          
          setLastUpdated(new Date());
        }
      });

      // Subscribe to service started events
      wsClient.subscribe(WS_MSG_TYPES.SERVICE_STARTED, (payload) => {
        if (payload && payload.service_name) {
          setStatuses(prevStatuses => ({
            ...prevStatuses,
            [payload.service_name]: 'running'
          }));
          
          // Also update dependency statuses
          setDependencyStatuses(prevDeps => ({
            ...prevDeps,
            [payload.service_name]: 'running'
          }));
          
          setLastUpdated(new Date());
        }
      });

      // Subscribe to service stopped events
      wsClient.subscribe(WS_MSG_TYPES.SERVICE_STOPPED, (payload) => {
        if (payload && payload.service_name) {
          setStatuses(prevStatuses => ({
            ...prevStatuses,
            [payload.service_name]: 'stopped'
          }));
          
          // Also update dependency statuses
          setDependencyStatuses(prevDeps => ({
            ...prevDeps,
            [payload.service_name]: 'stopped'
          }));
          
          setLastUpdated(new Date());
        }
      });

      // Subscribe to service error events
      wsClient.subscribe(WS_MSG_TYPES.SERVICE_ERROR, (payload) => {
        if (payload && payload.service_name) {
          setStatuses(prevStatuses => ({
            ...prevStatuses,
            [payload.service_name]: 'failed'
          }));
          
          // Also update dependency statuses
          setDependencyStatuses(prevDeps => ({
            ...prevDeps,
            [payload.service_name]: 'failed'
          }));
          
          setLastUpdated(new Date());
        }
      });
      
    } catch (error) {
      console.error('[App] Failed to initialize WebSocket client:', error);
    }
  };

  useEffect(() => {
    // Initialize WebSocket connection
    initializeWebSocket();

    // Cleanup WebSocket on unmount
    return () => {
      if (wsClient) {
        wsClient.disconnect();
      }
    };
  }, []);

  return { initializeWebSocket };
};