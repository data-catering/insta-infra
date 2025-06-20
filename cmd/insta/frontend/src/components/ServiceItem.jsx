import React, { useState, useEffect, useContext, useCallback } from 'react';
import { 
  startService, 
  stopService, 
  getServiceConnection,
  openServiceConnection,
  openURL, 
  checkImageExists,
  getAllImageStatuses, 
  startImagePull, 
  wsClient,
  WS_MSG_TYPES,
  restartRuntime
} from "../api/client";
import { ScrollText, Download, CheckCircle, AlertCircle, RotateCcw, X } from 'lucide-react';
import ProgressModal from './ProgressModal';
import LogsModal from './LogsModal';
import ConnectionModal from './ConnectionModal';

// Helper function for logging
const logDebug = (message, ...args) => {
  const fullMessage = `[ServiceItem] ${message}`;
  console.log(fullMessage, ...args);
};

// Create a context for bulk image status management
const ImageStatusContext = React.createContext();

// Provider component for bulk image status management
export const ImageStatusProvider = ({ children, services }) => {
  const [imageStatuses, setImageStatuses] = useState({});
  const [imageNames, setImageNames] = useState({});
  const [isLoading, setIsLoading] = useState(false);

  // Load image statuses for all services efficiently with a single API call
  const loadImageStatuses = useCallback(async () => {
    if (!services || services.length === 0 || isLoading) return;

    setIsLoading(true);
    try {
      // Get all image statuses with a single API call  
      const response = await getAllImageStatuses();
      const imageStatuses = response.image_statuses || {};
      
      const statusMap = {};
      const namesMap = {};
      
      // Process the response - this covers all services at once
      Object.keys(imageStatuses).forEach(serviceName => {
        const imageInfo = imageStatuses[serviceName];
        statusMap[serviceName] = imageInfo.status || 'unknown';
        namesMap[serviceName] = imageInfo.image_name || serviceName;
      });
      
      // Ensure all services have a status (in case some were missed)
      services.forEach(service => {
        const serviceName = service?.name || service?.Name;
        if (serviceName && !statusMap.hasOwnProperty(serviceName)) {
          statusMap[serviceName] = 'unknown';
          namesMap[serviceName] = serviceName;
        }
      });

      setImageNames(namesMap);
      setImageStatuses(statusMap);
      
    } catch (error) {
      console.error('ImageStatusProvider: Error loading image statuses:', error);
      // Set all to unknown on error
      const statusMap = {};
      services.forEach(service => {
        const serviceName = service?.name || service?.Name;
        if (serviceName) {
          statusMap[serviceName] = 'unknown';
        }
      });
      setImageStatuses(statusMap);
    } finally {
      setIsLoading(false);
    }
  }, [services?.length]); // Only depend on services length, not the full services array

  // Update individual image status
  const updateImageStatus = (serviceName, status, imageName = null) => {
    if (!serviceName) return;
    
    setImageStatuses(prev => ({
      ...prev,
      [serviceName]: status
    }));
    
    if (imageName) {
      setImageNames(prev => ({
        ...prev,
        [serviceName]: imageName
      }));
    }
  };

  const value = {
    imageStatuses,
    imageNames,
    isLoading,
    loadImageStatuses,
    updateImageStatus
  };

  return (
    <ImageStatusContext.Provider value={value}>
      {children}
    </ImageStatusContext.Provider>
  );
};

// Hook to use image status context
const useImageStatus = () => {
  const context = useContext(ImageStatusContext);
  if (!context) {
    throw new Error('useImageStatus must be used within an ImageStatusProvider');
  }
  return context;
};

// Component to automatically trigger image loading when services are available
const ImageLoader = ({ services }) => {
  const { loadImageStatuses, isLoading, imageStatuses } = useImageStatus();
  
  // Create a stable key based on service names to prevent unnecessary reloads
  const servicesKey = React.useMemo(() => {
    if (!services || services.length === 0) return '';
    return services.map(s => s?.name || s?.Name).filter(Boolean).sort().join(',');
  }, [services]);
  
  // Track if we've already loaded for this set of services
  const [loadedForKey, setLoadedForKey] = React.useState('');
  
  useEffect(() => {
    // Only load if:
    // 1. We have services
    // 2. We're not already loading
    // 3. We haven't loaded for this exact set of services
    // 4. We don't already have image statuses for most services
    if (services && services.length > 0 && !isLoading && servicesKey !== loadedForKey) {
      // Check if we already have image statuses for most services
      const hasExistingStatuses = services.some(service => {
        const serviceName = service?.name || service?.Name;
        return serviceName && imageStatuses[serviceName] && imageStatuses[serviceName] !== 'unknown';
      });
      
      // Only load if we don't have existing statuses
      if (!hasExistingStatuses) {
        loadImageStatuses();
        setLoadedForKey(servicesKey);
      }
    }
  }, [servicesKey, loadImageStatuses, isLoading, loadedForKey, imageStatuses]); // Use servicesKey instead of services.length
  
  return null; // This component doesn't render anything
};

// Simple dependency item component that shows dependency name and status
const DependencyItem = ({ serviceName, globalStatuses = {}, serviceStatuses = {} }) => {
  // Ensure we have a valid service name
  if (!serviceName || typeof serviceName !== 'string') {
    return null;
  }

  // Get status from dependency statuses first, then service statuses, then default to 'stopped'
  let statusObj = globalStatuses[serviceName];
  let containerStatus = null;
  
  if (statusObj) {
    // Handle both string and object status values
    if (typeof statusObj === 'string') {
      containerStatus = statusObj;
    } else if (typeof statusObj === 'object') {
      containerStatus = statusObj.Status || statusObj.status;
    }
  }
  
  // If not found in dependency statuses, check main service statuses
  if (!containerStatus) {
    const serviceStatusObj = serviceStatuses[serviceName];
    if (serviceStatusObj) {
      containerStatus = typeof serviceStatusObj === 'object' ? 
        (serviceStatusObj.Status || serviceStatusObj.status) : 
        serviceStatusObj;
    }
  }
  
  // If still not found, assume it's stopped (container doesn't exist in docker ps -a)
  if (!containerStatus) {
    containerStatus = 'stopped';
  }
  
  // Get status indicator
  const getStatusIndicator = () => {
    switch (containerStatus) {
      case 'running-healthy':
        return <div className="dependency-status-dot status-running-healthy" title="Running (Healthy)"></div>;
      case 'running-unhealthy':
        return <div className="dependency-status-dot status-running-unhealthy" title="Running (Unhealthy)"></div>;
      case 'running-health-starting':
        return <div className="dependency-status-dot status-running-health-starting" title="Running (Starting)"></div>;
      case 'running':
        return <div className="dependency-status-dot status-running" title="Running"></div>;
      case 'stopped':
        return <div className="dependency-status-dot status-stopped" title="Stopped"></div>;
      case 'starting':
        return <div className="dependency-status-dot status-starting" title="Starting"></div>;
      case 'completed':
        return <div className="dependency-status-dot status-completed" title="Completed successfully"></div>;
      case 'failed':
        return <div className="dependency-status-dot status-failed" title="Failed"></div>;
      default:
        return <div className="dependency-status-dot status-stopped" title="Stopped"></div>;
    }
  };
  
  return (
    <div className="dependency-item">
      {getStatusIndicator()}
      <span className="dependency-name">{serviceName}</span>
      <span className="dependency-container-status" title={`Container status: ${containerStatus}`}>
        {String(containerStatus || 'stopped').toUpperCase()}
      </span>
    </div>
  );
};

// StuckContainerWarning component for showing runtime restart warnings
const StuckContainerWarning = ({ containerName, runtimeName, onRestart, onDismiss }) => {
  const [isRestarting, setIsRestarting] = useState(false);

  const handleRestart = async () => {
    setIsRestarting(true);
    try {
      await onRestart();
    } finally {
      setIsRestarting(false);
    }
  };

  return (
    <div className="stuck-container-warning">
      <div className="warning-content">
        <div className="warning-icon">
          <AlertCircle size={20} />
        </div>
        <div className="warning-message">
          <strong>Container Stuck</strong>
          <p>The container '{containerName}' is stuck in 'created' status. This usually indicates {runtimeName} needs to be restarted.</p>
        </div>
        <div className="warning-actions">
          <button 
            onClick={handleRestart}
            disabled={isRestarting}
            className="button-warning restart-button"
            title={`Restart ${runtimeName} runtime`}
          >
            <RotateCcw size={16} />
            {isRestarting ? 'Restarting...' : 'Restart'}
          </button>
          <button 
            onClick={onDismiss}
            className="button-secondary dismiss-button"
            title="Dismiss warning"
          >
            <X size={16} />
          </button>
        </div>
      </div>
    </div>
  );
};

// Placeholder ServiceItem component
// This will be expanded to show service details and actions
function ServiceItem({ service, onServiceStateChange, statuses = {}, dependencyStatuses = {}, serviceStatuses = {} }) {
  // Ensure service is a valid object
  if (!service || typeof service !== 'object') {
    console.warn('ServiceItem: Invalid service prop:', service);
    return null;
  }
  
  const { name, type } = service;
  
  // Ensure name is valid
  if (!name || typeof name !== 'string') {
    console.warn('ServiceItem: Invalid service name:', name);
    return null;
  }

  // Ensure type is a string (or provide fallback)
  const serviceType = type && typeof type === 'string' ? type : 'Service';
  
  const [expanded, setExpanded] = useState(false);
  const [isStarting, setIsStarting] = useState(false);
  const [isStopping, setIsStopping] = useState(false);
  const [showProgressModal, setShowProgressModal] = useState(false);
  const [showLogsModal, setShowLogsModal] = useState(false);
  const [showConnectionModal, setShowConnectionModal] = useState(false);
  const [connectionInfo, setConnectionInfo] = useState(null);
  const [copyFeedback, setCopyFeedback] = useState('');
  const [persistData, setPersistData] = useState(false);
  const [showConnectionInfo, setShowConnectionInfo] = useState(false);
  const [isLoadingConnectionInfo, setIsLoadingConnectionInfo] = useState(false);
  const [stuckContainerWarning, setStuckContainerWarning] = useState(null); // { containerName, runtimeName }
  
  // Use image status context instead of local state
  const { imageStatuses, imageNames, updateImageStatus } = useImageStatus();
  
  // Simplified status management - prefer external status over local when available
  // This reduces conflicts and flashing between local state and WebSocket updates
  const [localStatus, setLocalStatus] = useState(null);
  const [localStatusError, setLocalStatusError] = useState(null);
  const [isInTransition, setIsInTransition] = useState(false);
  
  // Get effective status - prioritize external status from props when available
  const externalStatus = statuses[name];
  const effectiveStatus = externalStatus || localStatus || 'stopped';
  const effectiveStatusError = localStatusError || service.statusError;
  

  
  // Auto-clear copy feedback after 2 seconds
  useEffect(() => {
    if (copyFeedback) {
      const timer = setTimeout(() => setCopyFeedback(''), 2000);
      return () => clearTimeout(timer);
    }
  }, [copyFeedback]);

  // Reset local status when external status is available
  useEffect(() => {
    if (externalStatus && localStatus && localStatus !== externalStatus) {
      // External status takes precedence, clear local status
      setLocalStatus(null);
      setLocalStatusError(null);
      setIsInTransition(false);
    }
  }, [externalStatus, localStatus]);

  // WebSocket subscription for real-time status updates
  useEffect(() => {
    if (!name) return;

    // Subscribe to service status updates for this specific service
    const statusCallback = wsClient.subscribeToServiceStatus(name, (payload) => {
      const { status, error } = payload;
      if (status) {
        // For WebSocket updates, notify parent instead of managing local state
        // This keeps status management centralized and reduces conflicts
        notifyStatusChange(status, error);
        
        // Clear local status since external status will be updated
        setLocalStatus(null);
        setLocalStatusError(null);
        setIsInTransition(false);
      }
    });

    // Cleanup subscriptions on unmount
    return () => {
      wsClient.unsubscribe(WS_MSG_TYPES.SERVICE_STATUS_UPDATE, statusCallback);
    };
  }, [name]);

  // Get image status from context
  const imageStatus = imageStatuses[name] || 'unknown';

  // Determine status styling based on effective status
  let statusClass = '';
  let statusIcon = null;
  let statusText = '';
  let detailedStatusText = ''; // For expanded view

  // Ensure effectiveStatus is a string
  const safeStatus = String(effectiveStatus || 'stopped');
  const safeStatusError = effectiveStatusError ? String(effectiveStatusError) : null;

  // Helper function to check if status is a running variant
  const isRunningVariant = safeStatus === 'running' || 
                          safeStatus === 'running-healthy' || 
                          safeStatus === 'running-unhealthy' || 
                          safeStatus === 'running-health-starting' ||
                          safeStatus.startsWith('running');

  if (safeStatusError) {
    statusClass = 'status-bar-error';
    statusText = 'Error';
    detailedStatusText = 'Error';
    statusIcon = (
      <svg width="16" height="16" className="status-icon" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
      </svg>
    );
  } else if (safeStatus === 'running-healthy') {
    statusClass = 'status-bar-running-healthy';
    statusText = 'Running'; // Simplified for main view
    detailedStatusText = 'Running (Healthy)'; // Detailed for expanded view
    statusIcon = (
      <svg width="16" height="16" className="status-icon" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
      </svg>
    );
  } else if (safeStatus === 'running-unhealthy') {
    statusClass = 'status-bar-running-unhealthy';
    statusText = 'Running'; // Simplified for main view
    detailedStatusText = 'Running (Unhealthy)'; // Detailed for expanded view
    statusIcon = (
      <svg width="16" height="16" className="status-icon" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
      </svg>
    );
  } else if (safeStatus === 'running-health-starting') {
    statusClass = 'status-bar-running-health-starting';
    statusText = 'Running'; // Simplified for main view
    detailedStatusText = 'Running (Health Check Starting)'; // Detailed for expanded view
    statusIcon = (
      <div className="status-icon health-check-pending">
        <svg width="16" height="16" className="status-icon" fill="currentColor" viewBox="0 0 20 20">
          <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clipRule="evenodd" />
        </svg>
      </div>
    );
  } else if (safeStatus === 'running') {
    statusClass = 'status-bar-running';
    statusText = 'Running';
    detailedStatusText = 'Running';
    statusIcon = (
      <svg width="16" height="16" className="status-icon" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
      </svg>
    );
  } else if (safeStatus.startsWith('running')) {
    // Handle any other running variants
    statusClass = 'status-bar-running';
    statusText = 'Running'; // Simplified for main view
    detailedStatusText = safeStatus; // Show full status in expanded view
    statusIcon = (
      <svg width="16" height="16" className="status-icon" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
      </svg>
    );
  } else if (safeStatus === 'starting' || isInTransition && !externalStatus) {
    statusClass = 'status-bar-starting';
    statusText = 'Starting';
    detailedStatusText = 'Starting';
    statusIcon = (
      <div className="status-icon spin"></div>
    );
  } else if (safeStatus === 'failed') {
    statusClass = 'status-bar-error';
    statusText = 'Failed';
    detailedStatusText = 'Failed';
    statusIcon = (
      <svg width="16" height="16" className="status-icon" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
      </svg>
    );
  } else if (safeStatus === 'stopped') {
    statusClass = 'status-bar-stopped';
    statusText = 'Stopped';
    detailedStatusText = 'Stopped';
    statusIcon = (
      <svg width="16" height="16" className="status-icon" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM7 9a1 1 0 000 2h6a1 1 0 100-2H7z" clipRule="evenodd" />
      </svg>
    );
  } else {
    statusClass = 'status-bar-unknown';
    statusText = 'Unknown';
    detailedStatusText = safeStatus || 'Unknown';
    statusIcon = (
      <svg width="16" height="16" className="status-icon" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zM8.994 7.076a.997.997 0 11.658 1.877c.376.048.756.058 1.136.033.263-.018.516-.081.746-.19.245-.115.477-.278.666-.495.19-.218.33-.513.33-.834 0-.336-.123-.618-.292-.85-.17-.23-.392-.413-.647-.55a2.749 2.749 0 00-.885-.21h-.012zm.994 8.924a1 1 0 100-2 1 1 0 000 2z" clipRule="evenodd" />
      </svg>
    );
  }

  // Get service type based colors
  const getTypeClass = () => {
    switch(serviceType?.toLowerCase()) {
      case 'database':
        return 'color-database';
      case 'messaging':
        return 'color-messaging';
      case 'data catalog':
        return 'color-default';
      case 'data visualisation':
      case 'data visualization':
        return 'color-visualization';
      case 'monitoring':
        return 'color-monitoring';
      case 'job orchestrator':
        return 'color-orchestrator';
      default:
        return 'color-default';
    }
  };

  // Get service icon based on type
  const getServiceIcon = () => {
    switch(serviceType?.toLowerCase()) {
      case 'database':
        return (
          <svg width="16" height="16" className="service-type-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M12 8C15.3137 8 18 6.65685 18 5C18 3.34315 15.3137 2 12 2C8.68629 2 6 3.34315 6 5C6 6.65685 8.68629 8 12 8Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            <path d="M6 5V19C6 20.6569 8.68629 22 12 22C15.3137 22 18 20.6569 18 19V5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            <path d="M6 12C6 13.6569 8.68629 15 12 15C15.3137 15 18 13.6569 18 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        );
      case 'messaging':
        return (
          <svg width="16" height="16" className="service-type-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M8 10H8.01M12 10H12.01M16 10H16.01M9 16H5C3.89543 16 3 15.1046 3 14V6C3 4.89543 3.89543 4 5 4H19C20.1046 4 21 4.89543 21 6V14C21 15.1046 20.1046 16 19 16H14L9 21V16Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        );
      case 'data catalog':
        return (
          <svg width="16" height="16" className="service-type-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M4 7V17C4 19.2091 7.58172 21 12 21C16.4183 21 20 19.2091 20 17V7" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            <path d="M4 15C4 17.2091 7.58172 19 12 19C16.4183 19 20 17.2091 20 15" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            <path d="M4 11C4 13.2091 7.58172 15 12 15C16.4183 15 20 13.2091 20 11" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            <path d="M4 7C4 9.20914 7.58172 11 12 11C16.4183 11 20 9.20914 20 7C20 4.79086 16.4183 3 12 3C7.58172 3 4 4.79086 4 7Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        );
      case 'data visualisation':
      case 'data visualization':
        return (
          <svg width="16" height="16" className="service-type-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M8 13V17M16 11V17M12 7V17M7.8 21H16.2C17.8802 21 18.7202 21 19.362 20.673C19.9265 20.3854 20.3854 19.9265 20.673 19.362C21 18.7202 21 17.8802 21 16.2V7.8C21 6.11984 21 5.27976 20.673 4.63803C20.3854 4.07354 19.9265 3.6146 19.362 3.32698C18.7202 3 17.8802 3 16.2 3H7.8C6.11984 3 5.27976 3 4.63803 3.32698C4.07354 3.6146 3.6146 4.07354 3.32698 4.63803C3 5.27976 3 6.11984 3 7.8V16.2C3 17.8802 3 18.7202 3.32698 19.362C3.6146 19.9265 4.07354 20.3854 4.63803 20.673C5.27976 21 6.11984 21 7.8 21Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        );
      case 'monitoring':
        return (
          <svg width="16" height="16" className="service-type-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M3 13H11V21H3V13Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            <path d="M13 3H21V13H13V3Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            <path d="M13 15H21V21H13V15Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            <path d="M3 3H11V11H3V3Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        );
      case 'job orchestrator':
        return (
          <svg width="16" height="16" className="service-type-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M17 20V14M14 17H20M6 14L11 19L6 14ZM6 5L11 10L6 5ZM19 5L14 10L19 5Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        );
      default:
        return (
          <svg width="16" height="16" className="service-type-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M9 12H15M9 16H15M17 21H7C5.89543 21 5 20.1046 5 19V5C5 3.89543 5.89543 3 7 3H12.5858C12.851 3 13.1054 3.10536 13.2929 3.29289L18.7071 8.70711C18.8946 8.89464 19 9.149 19 9.41421V19C19 20.1046 18.1046 21 17 21Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        );
    }
  };

  // Button state based on status
  const showStartButton = !isRunningVariant && safeStatus !== 'starting';
  const showStopButton = isRunningVariant;
  
  // Determine if service has web UI (based on common web UI services)
  const hasWebUI = [
    'grafana', 'superset', 'minio', 'keycloak', 'jenkins', 'kibana', 
    'metabase', 'airflow', 'jaeger', 'prometheus', 'amundsen', 'datahub',
    'cvat', 'doccano', 'argilla', 'blazer', 'evidence', 'redash', 'sonarqube',
    'vault', 'traefik', 'kong', 'flink', 'mlflow', 'temporal', 'prefect-data',
    'ray', 'label-studio', 'opensearch', 'loki'
  ].includes(name.toLowerCase());
  
  const showOpenButton = isRunningVariant && hasWebUI;
  const showConnectButton = isRunningVariant;
  const showLogsButton = isRunningVariant || safeStatus === 'failed' || safeStatusError;

  // Handle card click to expand/collapse
  const toggleExpanded = () => {
    const newExpanded = !expanded;
    setExpanded(newExpanded);
  };

  // Handle starting a service
  const handleStartService = async (e) => {
    e.stopPropagation();
    setIsStarting(true);
    setIsInTransition(true);
    
    // Set local status for immediate UI feedback, but mark as transitional
    setLocalStatus('starting');
    setLocalStatusError(null);
    
    try {
      // Check if image exists before starting
      const imageExists = await checkImageExists(name);
      
      if (!imageExists) {
        // Image doesn't exist, need to download it
        updateImageStatus(name, 'downloading');
        setShowProgressModal(true);
        
        try {
          // Start image pull
          await startImagePull(name);
          
          // Wait for download to complete using reliable polling only
          await new Promise((resolve, reject) => {
            let isResolved = false;
            let attempts = 0;
            const maxAttempts = 600; // 5 minutes worth of 500 milliseconds polls
            
            const pollForCompletion = async () => {
              while (attempts < maxAttempts && !isResolved) {
                await new Promise(resolve => setTimeout(resolve, 500)); // Wait 500 milliseconds
                attempts++;
                
                  const exists = await checkImageExists(name);
                  if (exists && !isResolved) {
                    isResolved = true;
                    resolve();
                    return;
                }
              }
              
              if (!isResolved) {
                logDebug('Image download timeout for', name, 'after', attempts, 'attempts');
                isResolved = true;
                reject(new Error(`Image download timeout for ${name} (5 minutes)`));
              }
            };

            pollForCompletion();
          });
          
          updateImageStatus(name, 'ready');
           
         } catch (downloadError) {
           console.error('Image download failed for', name, ':', downloadError);
           updateImageStatus(name, 'missing');
           setShowProgressModal(false);
           throw new Error(`Failed to download image for ${name}: ${downloadError.message}`);
         }
      }
      
      // Start the service - backend will handle status broadcasts
      await startService(name, persistData);
      
      // Update image status to ready if needed
      if (imageStatus !== 'ready') {
        updateImageStatus(name, 'ready');
      }
      notifyStatusChange('running', null);

      if (onServiceStateChange) {
        onServiceStateChange();
      }
      
    } catch (error) {
      console.error(`Failed to start ${name}:`, error);
      
      // Set status to failed on error
      const errorMessage = error.message || error.toString();
      setLocalStatus('failed');
      setLocalStatusError(errorMessage);
      setIsInTransition(false);
      
      // Notify parent about the failure
      notifyStatusChange('failed', errorMessage);
      
      // Reset image status if there was an error and it's still showing as downloading
      if (imageStatus === 'downloading') {
        updateImageStatus(name, 'missing');
      }
      
      // Check if this is a stuck container error
      if (errorMessage.startsWith('STUCK_CONTAINER:')) {
        // Parse the structured error: STUCK_CONTAINER:containerName:message
        const parts = errorMessage.split(':');
        if (parts.length >= 3) {
          const containerName = parts[1];
          const message = parts.slice(2).join(':'); // Rejoin in case message contains colons
          // Extract runtime name from the message
          const runtimeMatch = message.match(/indicates (\w+) needs to be restarted/);
          const runtimeName = runtimeMatch ? runtimeMatch[1] : 'Docker/Podman';
          
          // Show the stuck container warning instead of alert
          setStuckContainerWarning({ containerName, runtimeName });
        } else {
          // Fallback if parsing fails
          alert(`Failed to start ${name}: ${errorMessage}`);
        }
      } else {
      // Show different error messages based on the type of failure
      if (errorMessage.includes('download image') || errorMessage.includes('Image pull')) {
        alert(`Failed to download required image for ${name}: ${error.message || error}\n\nPlease check your internet connection and Docker/Podman status.`);
      } else {
        alert(`Failed to start ${name}: ${error.message || error}`);
        }
      }
    } finally {
      setIsInTransition(false);
      setIsStarting(false);
      setShowProgressModal(false);
    }
  };

  // Handle stopping a service
  const handleStopService = async (e) => {
    e.stopPropagation();
    setIsStopping(true);
    setIsInTransition(true);
    
    // Set local status for immediate UI feedback
    setLocalStatus('stopping');
    setLocalStatusError(null);
    
    try {
      // Stop the service - backend will handle status broadcasts
      await stopService(name);
      
      // Don't set local status immediately - let external status updates handle it
      // This prevents conflicts with WebSocket/status monitor updates
      
      // Notify parent about successful stop initiation
      notifyStatusChange('stopping', null);
      
      // Notify parent component to refresh all statuses
      if (onServiceStateChange) {
        onServiceStateChange();
      }
      
    } catch (error) {
      console.error(`Failed to stop ${name}:`, error);
      
      // Set error status if stop failed
      const errorMessage = error.message || error.toString();
      setLocalStatus('failed');
      setLocalStatusError(errorMessage);
      setIsInTransition(false);
      
      // Notify parent about the error
      notifyStatusChange('failed', errorMessage);
      
      alert(`Failed to stop ${name}: ${error.message || error}`);
    } finally {
      setIsStopping(false);
      
      // Clear transition state after a brief delay to allow external status to arrive
      setTimeout(() => {
        setIsInTransition(false);
      }, 1000);
    }
  };

  // Handle opening service in browser
  const handleOpenService = async (e) => {
    e.stopPropagation();
    try {
      const response = await openServiceConnection(name);
      if (response.connection && response.connection.webUrls && response.connection.webUrls.length > 0) {
        // Open the primary web URL in a new tab
        const primaryUrl = response.connection.webUrls[0].url;
        window.open(primaryUrl, '_blank', 'noopener,noreferrer');
        } else {
          alert(`Service ${name} does not have a web interface available`);
      }
    } catch (error) {
      console.error(`Failed to open ${name} in browser:`, error);
      alert(`Failed to open ${name} in browser: ${error.message || error}`);
    }
  };

  // Handle getting and showing connection info
  const handleConnectionInfo = async (e) => {
    e.stopPropagation();
    setIsLoadingConnectionInfo(true);
    setShowConnectionInfo(true);
    try {
      const response = await getServiceConnection(name);
      setConnectionInfo(response.connection);
    } catch (error) {
      console.error(`Failed to get connection info for ${name}:`, error);
      setConnectionInfo({ serviceName: name, available: false, error: error.message || error });
    } finally {
      setIsLoadingConnectionInfo(false);
    }
  };

  // Handle showing logs modal
  const handleShowLogs = (e) => {
    e.stopPropagation();
    setShowLogsModal(true);
  };

  const handleCloseLogsModal = () => {
    setShowLogsModal(false);
  };

  // Handle runtime restart from stuck container warning
  const handleRestartRuntime = async () => {
    try {
      await restartRuntime();
      // After restart, dismiss the warning and refresh statuses
      setStuckContainerWarning(null);
      if (onServiceStateChange) {
        onServiceStateChange();
      }
    } catch (error) {
      console.error('Failed to restart runtime:', error);
      alert(`Failed to restart runtime: ${error.message || error}`);
    }
  };

  // Handle dismissing the stuck container warning
  const handleDismissStuckContainerWarning = () => {
    setStuckContainerWarning(null);
  };

  // Handle copying text to clipboard
  const copyToClipboard = async (text) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopyFeedback('Copied to clipboard');
    } catch (error) {
      console.error('Failed to copy to clipboard:', error);
      // Fallback for browsers that don't support clipboard API
      const textArea = document.createElement('textarea');
      textArea.value = text;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
      setCopyFeedback('Failed to copy');
    }
  };

  const notifyStatusChange = (newStatus, error = null) => {
    try {
      // Create a status update object for this service
      const statusUpdate = {
        [name]: {
          ServiceName: name,
          Status: newStatus,
          Error: error
        }
      };
      
      // Notify parent component
      if (onServiceStateChange && typeof onServiceStateChange === 'function') {
        onServiceStateChange(statusUpdate);
      }
    } catch (err) {
      console.warn('Status notification error:', err);
      // Don't let status notification errors crash the component
    }
  };

  // Get image status indicator
  const getImageStatusIndicator = () => {
    if (imageStatus === 'downloading') {
      return (
        <div className="image-status downloading" title="Downloading image...">
          <Download size={12} className="text-blue-400 animate-bounce" />
        </div>
      );
    }

    switch (imageStatus) {
      case 'ready':
        return (
          <div className="image-status ready" title="Image ready">
            <CheckCircle size={12} className="text-green-400" />
          </div>
        );
      case 'missing':
        return (
          <div className="image-status missing" title="Image not downloaded">
            <Download size={12} className="text-yellow-400" />
          </div>
        );
      case 'unknown':
      default:
        return (
          <div className="image-status unknown" title="Image status unknown">
            <AlertCircle size={12} className="text-gray-400" />
          </div>
        );
    }
  };

  // Determine item classes based on status
  let itemClass = 'service-item';
  
  // Add dependency class if this is a dependency service
  if (service.isDependency || service.type === 'dependency') {
    itemClass += ' service-item-dependency';
  }
  
  if (safeStatus === 'running-healthy') {
    itemClass += ' service-item-running-healthy';
  } else if (safeStatus === 'running-unhealthy') {
    itemClass += ' service-item-running-unhealthy';
  } else if (safeStatus === 'running-health-starting') {
    itemClass += ' service-item-running-health-starting';
  } else if (safeStatus === 'running') {
    itemClass += ' service-item-running';
  } else if (safeStatus === 'starting') {
    itemClass += ' service-item-starting';
  } else if (safeStatus === 'failed' || safeStatusError) {
    itemClass += ' service-item-error';
  } else if (safeStatus === 'stopped') {
    itemClass += ' service-item-stopped';
  } else {
    itemClass += ' service-item-unknown';
  }

  return (
    <div className={itemClass}>
      {/* Color bar at the top */}
      <div className={`service-color-bar ${getTypeClass()}`}></div>
      
      {/* Status indicator line on the left */}
      <div className={`service-status-bar ${statusClass}`}></div>
      
      {/* Card Content */}
      <div className="service-content" onClick={toggleExpanded}>
        {/* Service Header */}
        <div className="service-header">
          <div className={`service-icon ${getTypeClass()}`}>
            {getServiceIcon()}
          </div>
          <div className="service-info">
            <div className="service-name-row">
              <h3 className="service-name" title={String(name || 'Unknown Service')}>
                {String(name || 'Unknown Service')}
              </h3>
              {/* Image Status Indicator */}
              {getImageStatusIndicator()}
            </div>
            <div className="service-meta">
              <div className={`service-status ${isRunningVariant ? 'text-green' : safeStatusError ? 'text-red' : 'text-gray'}`}>
                {statusIcon}
                {statusText}
                {safeStatusError && (
                  <span className="status-error" title={safeStatusError}>
                    ⚠️
                  </span>
                )}
              </div>
              
              <span className="service-type-badge">
                {String(serviceType || 'Service')}
              </span>
            </div>
          </div>
        </div>
        
        {/* Expandable Content */}
        <div className={`expandable-content ${expanded ? 'expanded' : ''}`}>
          {/* Service Details (expanded content) */}
          {expanded && (
            <div className="service-details">
              {/* Existing Service Details */}
                              <div className="details-grid">
                  <div className="details-label">Type:</div>
                  <div>{String(serviceType || 'Service')}</div>
                  <div className="details-label">Status:</div>
                  <div className={isRunningVariant ? 'text-green' : safeStatusError ? 'text-red' : 'text-gray'}>
                    {detailedStatusText}
                    {safeStatusError && (
                      <div className="status-error-details" title={safeStatusError}>
                        Error: {String(safeStatusError)}
                      </div>
                    )}
                  </div>
                </div>
              
              {/* Dependencies Section */}
              {Array.isArray(service.dependencies) && service.dependencies.length > 0 && (
                <div className="dependencies-section">
                  <div className="dependencies-header">
                    <div className="dependencies-label">Dependencies:</div>
                  </div>
                  <div className="dependencies-list">
                    {service.dependencies.map((dependency, index) => {
                      // Ensure dependency is a valid string
                      if (!dependency || typeof dependency !== 'string') {
                        console.warn('Invalid dependency:', dependency);
                        return null;
                      }
                      return (
                        <DependencyItem 
                          key={`${dependency}-${index}`} 
                          serviceName={dependency} 
                          globalStatuses={dependencyStatuses || {}}
                          serviceStatuses={serviceStatuses || {}}
                        />
                      );
                    })}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
        
        {/* Expand/Collapse Indicator */}
        <div className="collapse-button">
          <button
            onClick={(e) => { e.stopPropagation(); toggleExpanded(); }}
            className="icon-button"
            aria-expanded={expanded}
            aria-label={expanded ? "Collapse details" : "Expand details"}
          >
            <svg
              width="14"
              height="14"
              className={`collapse-icon ${expanded ? 'rotate-180' : ''}`}
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 9l-7 7-7-7" />
            </svg>
          </button>
        </div>
      </div>
      
      {/* Stuck Container Warning */}
      {stuckContainerWarning && (
        <StuckContainerWarning
          containerName={stuckContainerWarning.containerName}
          runtimeName={stuckContainerWarning.runtimeName}
          onRestart={handleRestartRuntime}
          onDismiss={handleDismissStuckContainerWarning}
        />
      )}
      
      {/* Action Buttons */}
      <div className="service-actions">
        {/* Start Service button with inline persist checkbox button */}
        {showStartButton && (
          <>
            {/* Persist Data checkbox button */}
            <button 
              className={`action-button persist-button ${persistData ? 'persist-active' : ''}`}
              onClick={(e) => {
                e.stopPropagation();
                setPersistData(!persistData);
              }}
              title={persistData ? "Data persistence enabled" : "Data persistence disabled"}
            >
              <svg width="10" height="10" className="action-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                {persistData ? (
                  <path d="M9 12L11 14L15 10M21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3C16.9706 3 21 7.02944 21 12Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                ) : (
                  <path d="M12 8V16M8 12H16M22 12C22 17.5228 17.5228 22 12 22C6.47715 22 2 17.5228 2 12C2 6.47715 6.47715 2 12 2C17.5228 2 22 6.47715 22 12Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                )}
              </svg>
              Persist
            </button>
            
            {/* Start Service button */}
            <button 
              className={`action-button start-button ${isStarting ? 'button-loading' : ''}`} 
              onClick={handleStartService}
              disabled={isStarting}
            >
              {isStarting ? (
                <>
                  <div className="action-icon spin"></div>
                  Starting...
                </>
              ) : (
                <>
                  <svg width="10" height="10" className="action-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M8 5V19L19 12L8 5Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                  Start
                </>
              )}
            </button>
          </>
        )}
        
        {/* Stop Service button */}
        {showStopButton && (
          <button 
            className={`action-button stop-button ${isStopping ? 'button-loading' : ''}`} 
            onClick={handleStopService}
            disabled={isStopping}
          >
            {isStopping ? (
              <>
                <div className="action-icon spin"></div>
                Stopping...
              </>
            ) : (
              <>
                <svg width="10" height="10" className="action-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M6 6H18V18H6V6Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
                Stop
              </>
            )}
          </button>
        )}
        
        {/* Open in Browser button - only show for running services that have web UIs */}
        {showOpenButton && (
          <button 
            className="action-button open-button"
            onClick={handleOpenService}
            title="Open in browser"
          >
            <svg width="10" height="10" className="action-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M7 7H17V17M17 7L7 17" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            Open
          </button>
        )}
        
        {/* Connection Info button - only show for all running services */}
        {showConnectButton && (
          <button 
            className={`action-button info-button ${isLoadingConnectionInfo ? 'button-loading' : ''}`}
            onClick={handleConnectionInfo}
            disabled={isLoadingConnectionInfo}
            title="Show connection details"
          >
            {isLoadingConnectionInfo ? (
              <>
                <div className="action-icon spin"></div>
                Loading...
              </>
            ) : (
              <>
                <svg width="10" height="10" className="action-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M13 16H12V12H11M12 8H12.01M21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3C16.9706 3 21 7.02944 21 12Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
                Connect
              </>
            )}
          </button>
        )}

        {/* Logs button - only show for running services or failed services */}
        {showLogsButton && (
        <button 
          className="action-button logs-button" 
          onClick={handleShowLogs}
          title="View logs"
        >
          <ScrollText size={10} className="action-icon" />
          Logs
        </button>
        )}
      </div>

      {/* Connection Info Modal */}
      <ConnectionModal 
        isOpen={showConnectionInfo}
        onClose={() => setShowConnectionInfo(false)}
        connectionInfo={connectionInfo}
        copyFeedback={copyFeedback}
        onCopyToClipboard={copyToClipboard}
      />

      {/* Logs Modal */}
      <LogsModal 
        isOpen={showLogsModal}
        onClose={handleCloseLogsModal}
        serviceName={name}
      />

      {/* Progress Modal */}
      <ProgressModal 
        isOpen={showProgressModal}
        onClose={() => setShowProgressModal(false)}
        serviceName={name}
        imageName={imageNames[name] || ''}
      />
    </div>
  );
}

export default ServiceItem;

export { useImageStatus, ImageLoader };

