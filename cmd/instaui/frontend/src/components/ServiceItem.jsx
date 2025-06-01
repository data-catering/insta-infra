import React, { useState, useEffect, useContext, createContext, useCallback } from 'react';
import { StartServiceWithStatusUpdate, StopService, StopServiceWithStatusUpdate, GetServiceConnectionInfo, OpenServiceInBrowser, CheckImageExists, StartImagePull, GetImageInfo, GetServiceStatus } from "../../wailsjs/go/main/App";
import { ChevronDown, ChevronUp, ExternalLink, Database, Server, Eye, Play, Square, Copy, ScrollText, Download, CheckCircle, AlertCircle } from 'lucide-react';
import ProgressModal from './ProgressModal';
import LogsModal from './LogsModal';
import ConnectionModal from './ConnectionModal';

// Create a context for bulk image status management
const ImageStatusContext = React.createContext();

// Provider component for bulk image status management
export const ImageStatusProvider = ({ children, services }) => {
  const [imageStatuses, setImageStatuses] = useState({});
  const [imageNames, setImageNames] = useState({});
  const [isLoading, setIsLoading] = useState(false);

  // Bulk load image statuses for all services
  const loadImageStatuses = useCallback(async () => {
    if (!services || services.length === 0 || isLoading) return;

    setIsLoading(true);
    try {
      const serviceNames = services.map(service => service.name || service.Name); // Handle both lowercase and uppercase Name
      
      // Get image names and existence status in bulk
      const [imageInfoMap, imageExistsMap] = await Promise.all([
        window.go.main.App.GetMultipleImageInfo(serviceNames),
        window.go.main.App.CheckMultipleImagesExist(serviceNames)
      ]);

      setImageNames(imageInfoMap);
      
      // Convert boolean existence to status strings
      const statusMap = {};
      for (const serviceName of serviceNames) {
        const exists = imageExistsMap[serviceName];
        statusMap[serviceName] = exists ? 'ready' : 'missing';
      }
      
      setImageStatuses(statusMap);
      
    } catch (error) {
      console.error('ImageStatusProvider: Error loading bulk image statuses:', error);
      // Set all to unknown on error
      const statusMap = {};
      services.forEach(service => {
        statusMap[service.name || service.Name] = 'unknown';
      });
      setImageStatuses(statusMap);
    } finally {
      setIsLoading(false);
    }
  }, [services?.length]); // Only depend on services length, not the full services array

  // Update individual image status
  const updateImageStatus = (serviceName, status, imageName = null) => {
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
    return services.map(s => s.name || s.Name).sort().join(',');
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
        const serviceName = service.name || service.Name;
        return imageStatuses[serviceName] && imageStatuses[serviceName] !== 'unknown';
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
const DependencyItem = ({ serviceName, globalStatuses = {} }) => {
  // Get status from global statuses passed down from the parent
  // globalStatuses contains ServiceStatus objects, so we need to extract the Status property
  const statusObj = globalStatuses[serviceName];
  const containerStatus = statusObj ? (statusObj.Status || statusObj.status || 'unknown') : 'unknown';
  
  // Get status indicator
  const getStatusIndicator = () => {
    switch (containerStatus) {
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
        return <div className="dependency-status-dot status-unknown" title="Status unknown"></div>;
    }
  };
  
  return (
    <div className="dependency-item">
      {getStatusIndicator()}
      <span className="dependency-name">{serviceName}</span>
      <span className="dependency-container-status" title={`Container status: ${containerStatus}`}>
        {containerStatus.toUpperCase()}
      </span>
    </div>
  );
};

// Placeholder ServiceItem component
// This will be expanded to show service details and actions
function ServiceItem({ service, onServiceStateChange, statuses = {}, dependencyStatuses = {} }) {
  const { name, type } = service;
  const [status, setStatus] = useState('stopped');
  const [expanded, setExpanded] = useState(false);
  const [isStarting, setIsStarting] = useState(false);
  const [isStopping, setIsStopping] = useState(false);
  const [isCheckingImage, setIsCheckingImage] = useState(false);
  const [imageExists, setImageExists] = useState(null);
  const [imageName, setImageName] = useState('');
  const [showProgressModal, setShowProgressModal] = useState(false);
  const [showLogsModal, setShowLogsModal] = useState(false);
  const [showConnectionModal, setShowConnectionModal] = useState(false);
  const [connectionInfo, setConnectionInfo] = useState(null);
  const [copyFeedback, setCopyFeedback] = useState('');
  const [persistData, setPersistData] = useState(false);
  const [showConnectionInfo, setShowConnectionInfo] = useState(false);
  const [isLoadingConnectionInfo, setIsLoadingConnectionInfo] = useState(false);
  
  // Use image status context instead of local state
  const { imageStatuses, imageNames, updateImageStatus } = useImageStatus();
  
  // Local status management - this component controls its own status
  const [localStatus, setLocalStatus] = useState(service.status);
  const [localStatusError, setLocalStatusError] = useState(service.statusError);
  
  // Auto-clear copy feedback after 2 seconds
  useEffect(() => {
    if (copyFeedback) {
      const timer = setTimeout(() => setCopyFeedback(''), 2000);
      return () => clearTimeout(timer);
    }
  }, [copyFeedback]);

  // Update local status when service prop changes (external updates)
  useEffect(() => {
    setLocalStatus(service.status);
    setLocalStatusError(service.statusError);
  }, [service.status, service.statusError]);

  // Get image status from context
  const imageStatus = imageStatuses[name] || 'unknown';
  const imageNameFromContext = imageNames[name] || '';

  // Early return if service is not properly defined
  if (!name) {
    return null;
  }

  // Determine status styling
  let statusClass = '';
  let statusIcon = null;
  let typeClass = 'color-default';

  if (localStatusError) {
    statusClass = 'status-bar-error';
    statusIcon = (
      <svg width="16" height="16" className="status-icon" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
      </svg>
    );
  } else if (localStatus === 'running') {
    statusClass = 'status-bar-running';
    statusIcon = (
      <svg width="16" height="16" className="status-icon" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
      </svg>
    );
  } else if (localStatus === 'starting') {
    statusClass = 'status-bar-starting';
    statusIcon = (
      <div className="status-icon spin"></div>
    );
  } else if (localStatus === 'failed') {
    statusClass = 'status-bar-error';
    statusIcon = (
      <svg width="16" height="16" className="status-icon" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
      </svg>
    );
  } else if (localStatus === 'stopped') {
    statusClass = 'status-bar-stopped';
    statusIcon = (
      <svg width="16" height="16" className="status-icon" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM7 9a1 1 0 000 2h6a1 1 0 100-2H7z" clipRule="evenodd" />
      </svg>
    );
  }

  // Get service type based colors
  const getTypeClass = () => {
    switch(type?.toLowerCase()) {
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
    switch(type?.toLowerCase()) {
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
  const showStartButton = localStatus !== 'running' && localStatus !== 'starting';
  const showStopButton = localStatus === 'running';
  
  // Determine if service has web UI (based on common web UI services)
  const hasWebUI = [
    'grafana', 'superset', 'minio', 'keycloak', 'jenkins', 'kibana', 
    'metabase', 'airflow', 'jaeger', 'prometheus', 'amundsen', 'datahub',
    'cvat', 'doccano', 'argilla', 'blazer', 'evidence', 'redash', 'sonarqube',
    'vault', 'traefik', 'kong', 'flink', 'mlflow', 'temporal', 'prefect-data',
    'ray', 'label-studio', 'opensearch', 'loki'
  ].includes(name.toLowerCase());
  
  const showOpenButton = localStatus === 'running' && hasWebUI;
  const showConnectButton = localStatus === 'running';

  // Handle card click to expand/collapse
  const toggleExpanded = () => {
    const newExpanded = !expanded;
    setExpanded(newExpanded);
  };

  // Handle starting a service
  const handleStartService = async (e) => {
    e.stopPropagation();
    setIsStarting(true);
    
    // Immediately set status to starting for responsive UI
    setLocalStatus('starting');
    setLocalStatusError(null);
    
    try {
      // Check if image exists before starting
      const imageExists = await CheckImageExists(name);
      
      if (!imageExists) {
        // Image doesn't exist, need to download it
        updateImageStatus(name, 'downloading');
        setShowProgressModal(true);
        
        // Start image pull
        await StartImagePull(name);
        
        // Wait for download to complete (the modal will handle the UI feedback)
        // The progress modal will close automatically when download completes
      }
      
      // Start the service and get immediate status updates for ALL containers
      const updatedStatuses = await StartServiceWithStatusUpdate(name, persistData);
      
      // Update image status to ready
      updateImageStatus(name, 'ready');
      
      // Update our local status based on the response
      if (updatedStatuses && updatedStatuses[name]) {
        const serviceStatus = updatedStatuses[name];
        const newStatus = serviceStatus.Status || 'running';
        const newError = serviceStatus.Error || null;
        setLocalStatus(newStatus);
        setLocalStatusError(newError);
        
        // Notify parent about our status change
        notifyStatusChange(newStatus, newError);
      } else {
        // Fallback to running if no specific status returned
        setLocalStatus('running');
        setLocalStatusError(null);
        
        // Notify parent about our status change
        notifyStatusChange('running', null);
      }
      
      // Notify parent component about ALL status updates from the backend
      // This includes not just the service being started, but ALL containers that may have been affected
      if (onServiceStateChange && updatedStatuses) {
        onServiceStateChange(updatedStatuses);
      }
      
    } catch (error) {
      console.error(`Failed to start ${name}:`, error);
      
      // Set status to failed on error
      const errorMessage = error.message || error.toString();
      setLocalStatus('failed');
      setLocalStatusError(errorMessage);
      
      // Notify parent about the failure
      notifyStatusChange('failed', errorMessage);
      
      // Reset image status if there was an error
      if (imageStatus === 'downloading') {
        updateImageStatus(name, 'missing');
      }
      
      alert(`Failed to start ${name}: ${error.message || error}`);
    } finally {
      setIsStarting(false);
      setShowProgressModal(false);
    }
  };

  // Handle stopping a service
  const handleStopService = async (e) => {
    e.stopPropagation();
    setIsStopping(true);
    
    try {
      // Stop the service and get immediate status updates for ALL containers
      const updatedStatuses = await StopServiceWithStatusUpdate(name);
      
      // Update our local status based on the response
      if (updatedStatuses && updatedStatuses[name]) {
        const serviceStatus = updatedStatuses[name];
        const newStatus = serviceStatus.Status || 'stopped';
        const newError = serviceStatus.Error || null;
        setLocalStatus(newStatus);
        setLocalStatusError(newError);
        
        // Notify parent about our status change
        notifyStatusChange(newStatus, newError);
      } else {
        // Fallback to stopped if no specific status returned
        setLocalStatus('stopped');
        setLocalStatusError(null);
        
        // Notify parent about our status change
        notifyStatusChange('stopped', null);
      }
      
      // Notify parent component about ALL status updates from the backend
      // This includes not just the service being stopped, but ALL containers that may have been affected
      if (onServiceStateChange && updatedStatuses) {
        onServiceStateChange(updatedStatuses);
      }
      
    } catch (error) {
      console.error(`Failed to stop ${name}:`, error);
      
      // Set error status if stop failed
      const errorMessage = error.message || error.toString();
      setLocalStatus('failed');
      setLocalStatusError(errorMessage);
      
      // Notify parent about the error
      notifyStatusChange('failed', errorMessage);
      
      alert(`Failed to stop ${name}: ${error.message || error}`);
    } finally {
      setIsStopping(false);
    }
  };

  // Handle opening service in browser
  const handleOpenService = async (e) => {
    e.stopPropagation();
    try {
      await OpenServiceInBrowser(name);
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
      const info = await GetServiceConnectionInfo(name);
      setConnectionInfo(info);
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
      if (onServiceStateChange) {
        onServiceStateChange(statusUpdate);
      }
    } catch (err) {
      // Don't let status notification errors crash the component
      // Silently handle the error to avoid console spam in production
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
  if (localStatus === 'running') {
    itemClass += ' service-item-running';
  } else if (localStatus === 'starting') {
    itemClass += ' service-item-starting';
  } else if (localStatus === 'failed' || localStatusError) {
    itemClass += ' service-item-error';
  } else {
    itemClass += ' service-item-stopped';
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
              <h3 className="service-name" title={name}>
                {name}
              </h3>
              {/* Image Status Indicator */}
              {getImageStatusIndicator()}
            </div>
            <div className="service-meta">
              <div className={`service-status ${localStatus === 'running' ? 'text-green' : localStatusError ? 'text-red' : 'text-gray'}`}>
                {statusIcon}
                {localStatus}
              </div>
              
              <span className="service-type-badge">
                {type || 'Service'}
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
                <div>{type}</div>
                <div className="details-label">Status:</div>
                <div className={localStatus === 'running' ? 'text-green' : localStatusError ? 'text-red' : 'text-gray'}>
                  {localStatus}
                </div>
              </div>
              
              {/* Dependencies Section */}
              {service.dependencies && service.dependencies.length > 0 && (
                <div className="dependencies-section">
                  <div className="dependencies-header">
                    <div className="dependencies-label">Dependencies:</div>
                  </div>
                  <div className="dependencies-list">
                    {service.dependencies.map(dependency => {
                      return (
                        <DependencyItem 
                          key={dependency} 
                          serviceName={dependency} 
                          globalStatuses={dependencyStatuses}
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
      
      {/* Action Buttons */}
      <div className="service-actions">
        {/* Start Service button with inline persist checkbox button */}
        {localStatus !== 'running' && (
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
        {localStatus === 'running' && (
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
        
        {/* Connection Info button - show for all running services */}
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

        <button 
          className="action-button logs-button" 
          onClick={handleShowLogs}
          title="View logs"
        >
          <ScrollText size={10} className="action-icon" />
          Logs
        </button>
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
        imageName={imageNameFromContext}
      />
    </div>
  );
}

export default ServiceItem;

export { useImageStatus, ImageLoader };

