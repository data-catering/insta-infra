import React, { useState, useEffect } from 'react';
import { StartServiceWithStatusUpdate, StopService, GetServiceConnectionInfo, OpenServiceInBrowser, CheckImageExists, StartImagePull, GetImageInfo, GetDependencyStatus, StopDependencyChain, GetServiceStatus } from "../../wailsjs/go/main/App";
import { ChevronDown, ChevronUp, ExternalLink, Database, Server, Eye, Play, Square, Copy, ScrollText, Download, CheckCircle, AlertCircle, Users } from 'lucide-react';
import ConnectionModal from './ConnectionModal';
import LogsModal from './LogsModal';
import ProgressModal from './ProgressModal';

// Placeholder ServiceItem component
// This will be expanded to show service details and actions
function ServiceItem({ service, onServiceStateChange, onShowDependencyGraph }) {
  const [expanded, setExpanded] = useState(false);
  const [isStarting, setIsStarting] = useState(false);
  const [isStopping, setIsStopping] = useState(false);
  const [persistData, setPersistData] = useState(false);
  const [showConnectionInfo, setShowConnectionInfo] = useState(false);
  const [connectionInfo, setConnectionInfo] = useState(null);
  const [isLoadingConnectionInfo, setIsLoadingConnectionInfo] = useState(false);
  const [copyFeedback, setCopyFeedback] = useState('');
  const [showLogsModal, setShowLogsModal] = useState(false);
  const [imageStatus, setImageStatus] = useState('unknown'); // 'ready', 'missing', 'downloading', 'unknown'
  const [showProgressModal, setShowProgressModal] = useState(false);
  const [isCheckingImage, setIsCheckingImage] = useState(false);
  const [imageName, setImageName] = useState('');
  const [dependencyStatus, setDependencyStatus] = useState(null); // Detailed dependency status
  const [isCheckingDeps, setIsCheckingDeps] = useState(false);
  const [showDependencyDetails, setShowDependencyDetails] = useState(false);
  const [selectedDependencyForLogs, setSelectedDependencyForLogs] = useState(null); // Track which dependency's logs to show
  
  // Local status management - this component controls its own status
  const [localStatus, setLocalStatus] = useState(service.status);
  const [localStatusError, setLocalStatusError] = useState(service.statusError);
  
  // Auto-clear copy feedback after 2 seconds
  useEffect(() => {
    if (copyFeedback) {
      const timer = setTimeout(() => {
        setCopyFeedback('');
      }, 2000);
      return () => clearTimeout(timer);
    }
  }, [copyFeedback]);

  // Update local status when service prop changes (external updates)
  useEffect(() => {
    setLocalStatus(service.status);
    setLocalStatusError(service.statusError);
  }, [service.status, service.statusError]);

  const { 
    name, 
    type, 
    dependencies = [] 
  } = service || {};

  // Use local status for display
  const status = localStatus;
  const statusError = localStatusError;

  // Early return if service is not properly defined
  if (!name) {
    return null;
  }

  // Check image status when component mounts or when service changes
  useEffect(() => {
    if (name) {
      checkImageStatus();
      checkDependencyStatus();
    }
  }, [name]);

  // Also check dependency status when service status changes
  useEffect(() => {
    if (name && status) {
      checkDependencyStatus();
    }
  }, [name, status]);

  // Poll for status updates when service is starting
  useEffect(() => {
    let statusPollInterval;
    
    if (status === 'starting' && name) {
      // Poll every 2 seconds to check if service has finished starting
      statusPollInterval = setInterval(async () => {
        try {
          // Get current status for this service
          const currentStatus = await GetServiceStatus(name);
          
          // Only update if status actually changed and is valid
          if (currentStatus && currentStatus !== status) {
            setLocalStatus(currentStatus);
            setLocalStatusError(null); // Clear error on successful status check
            
            // Notify parent about the status change
            notifyStatusChange(currentStatus, null);
            
            // If service is no longer starting, refresh dependencies
            if (currentStatus !== 'starting') {
              setTimeout(() => {
                checkDependencyStatus();
              }, 500);
            }
          }
        } catch (error) {
          console.error(`Error polling status for ${name}:`, error);
          // On error, assume service failed to start (but only if still starting)
          if (status === 'starting') {
            const errorMessage = error.message || error.toString();
            setLocalStatus('failed');
            setLocalStatusError(errorMessage);
            notifyStatusChange('failed', errorMessage);
          }
        }
      }, 2000);
    }
    
    return () => {
      if (statusPollInterval) {
        clearInterval(statusPollInterval);
      }
    };
  }, [status, name]);

  // Determine status styling
  let statusClass = '';
  let statusText = status || 'Unknown';
  let statusIcon = null;
  let typeClass = 'color-default';

  if (statusError) {
    statusClass = 'status-bar-error';
    statusText = 'Error';
    statusIcon = (
      <svg width="16" height="16" className="status-icon" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
      </svg>
    );
  } else if (status === 'running') {
    statusClass = 'status-bar-running';
    statusText = 'Running';
    statusIcon = (
      <svg width="16" height="16" className="status-icon" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
      </svg>
    );
  } else if (status === 'starting') {
    statusClass = 'status-bar-starting';
    statusText = 'Starting';
    statusIcon = (
      <div className="status-icon spin"></div>
    );
  } else if (status === 'failed') {
    statusClass = 'status-bar-error';
    statusText = 'Failed';
    statusIcon = (
      <svg width="16" height="16" className="status-icon" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
      </svg>
    );
  } else if (status === 'stopped') {
    statusClass = 'status-bar-stopped';
    statusText = 'Stopped';
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
  const showStartButton = status !== 'running' && status !== 'starting';
  const showStopButton = status === 'running';
  
  // Determine if service has web UI (based on common web UI services)
  const hasWebUI = [
    'grafana', 'superset', 'minio', 'keycloak', 'jenkins', 'kibana', 
    'metabase', 'airflow', 'jaeger', 'prometheus', 'amundsen', 'datahub',
    'cvat', 'doccano', 'argilla', 'blazer', 'evidence', 'redash', 'sonarqube',
    'vault', 'traefik', 'kong', 'flink', 'mlflow', 'temporal', 'prefect-data',
    'ray', 'label-studio', 'opensearch', 'loki'
  ].includes(name.toLowerCase());
  
  const showOpenButton = status === 'running' && hasWebUI;
  const showConnectButton = status === 'running';

  // Handle card click to expand/collapse
  const toggleExpanded = () => {
    setExpanded(!expanded);
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
        setImageStatus('downloading');
        setShowProgressModal(true);
        
        // Start image pull
        await StartImagePull(name);
        
        // Wait for download to complete (the modal will handle the UI feedback)
        // The progress modal will close automatically when download completes
      }
      
      // Start the service and get immediate status updates
      const updatedStatuses = await StartServiceWithStatusUpdate(name, persistData);
      
      // Update image status to ready
      setImageStatus('ready');
      
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
      
      // Notify parent component about all status updates (for other services)
      if (onServiceStateChange) {
        onServiceStateChange(updatedStatuses);
      }
      
      // After successful start, refresh dependency status
      setTimeout(() => {
        checkDependencyStatus();
      }, 1000); // Small delay to allow containers to fully start
      
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
        setImageStatus('missing');
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
      await StopService(name);
      
      // Immediately update local status to stopped
      setLocalStatus('stopped');
      setLocalStatusError(null);
      
      // Notify parent about the status change
      notifyStatusChange('stopped', null);
      
      // Refresh the service data after stopping (for other services)
      if (onServiceStateChange) {
        onServiceStateChange();
      }
      
      // After successful stop, refresh dependency status
      setTimeout(() => {
        checkDependencyStatus();
      }, 500); // Small delay to allow containers to fully stop
      
    } catch (error) {
      console.error(`Failed to stop ${name}:`, error);
      
      // Set error status if stop failed
      const errorMessage = error.message || error.toString();
      setLocalStatusError(errorMessage);
      
      // Notify parent about the error
      notifyStatusChange(status, errorMessage);
      
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
    setSelectedDependencyForLogs(null); // Clear any selected dependency
    setShowLogsModal(true);
  };

  const handleCloseLogsModal = () => {
    setShowLogsModal(false);
    setSelectedDependencyForLogs(null); // Clear selected dependency when closing
  };

  // Handle stopping dependency chain
  const handleStopDependencyChain = async (e) => {
    e.stopPropagation();
    const confirmed = window.confirm(
      `This will stop ${name} and all services that depend on it. Are you sure?`
    );
    
    if (!confirmed) return;

    setIsStopping(true);
    try {
      await StopDependencyChain(name);
      
      // Immediately update local status to stopped
      setLocalStatus('stopped');
      setLocalStatusError(null);
      
      // Notify parent about the status change
      notifyStatusChange('stopped', null);
      
      // Refresh service data (for other services in the chain)
      if (onServiceStateChange) {
        onServiceStateChange();
      }
      
      // After successful stop, refresh dependency status
      setTimeout(() => {
        checkDependencyStatus();
      }, 1000); // Longer delay for dependency chain stops
    } catch (error) {
      console.error(`Failed to stop dependency chain for ${name}:`, error);
      
      // Set error status if stop failed
      const errorMessage = error.message || error.toString();
      setLocalStatusError(errorMessage);
      
      // Notify parent about the error
      notifyStatusChange(status, errorMessage);
      
      alert(`Failed to stop dependency chain for ${name}: ${error.message || error}`);
    } finally {
      setIsStopping(false);
    }
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

  // Check if Docker image exists for this service
  const checkImageStatus = async () => {
    if (!name || isCheckingImage) return;
    
    setIsCheckingImage(true);
    try {
      // First get the image name for this service
      const imageInfo = await GetImageInfo(name);
      setImageName(imageInfo);
      
      // Then check if the image exists
      const exists = await CheckImageExists(name);
      setImageStatus(exists ? 'ready' : 'missing');
    } catch (error) {
      console.error('Error checking image status:', error);
      setImageStatus('unknown');
      setImageName('');
    } finally {
      setIsCheckingImage(false);
    }
  };

  const checkDependencyStatus = async () => {
    if (!name || isCheckingDeps) return;
    
    setIsCheckingDeps(true);
    try {
      const depStatus = await GetDependencyStatus(name);
      setDependencyStatus(depStatus);
    } catch (error) {
      console.error(`Error checking dependency status for ${name}:`, error);
      // Don't set to null immediately, keep previous status to avoid UI flicker
      // Only set to null if this is the first check
      if (dependencyStatus === null) {
        setDependencyStatus(null);
      }
    } finally {
      setIsCheckingDeps(false);
    }
  };

  // Notify parent about status changes that might affect other services
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
    if (isCheckingImage) {
      return (
        <div className="image-status checking" title="Checking image status...">
          <svg width="12" height="12" className="spin" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
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
      case 'downloading':
        return (
          <div className="image-status downloading" title="Downloading image...">
            <Download size={12} className="text-blue-400 animate-bounce" />
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

  // Get dependency status indicator
  const getDependencyStatusIndicator = () => {
    if (!dependencyStatus || isCheckingDeps) {
      return (
        <div className="dependency-status loading" title="Checking dependencies...">
          <Users size={16} />
          <div className="status-spinner animate-spin"></div>
        </div>
      );
    }

    const { dependencies, allDependenciesReady, runningCount, requiredCount, errorCount } = dependencyStatus;
    
    if (dependencies.length === 0) {
      return null; // No dependencies, no indicator needed
    }

    let statusClass = 'dependency-status';
    let statusIcon = <Users size={16} />;
    let tooltipText = '';

    if (allDependenciesReady) {
      statusClass += ' ready';
      statusIcon = <CheckCircle size={16} />;
      tooltipText = `All ${dependencies.length} dependencies are running`;
    } else if (errorCount > 0) {
      statusClass += ' error';
      statusIcon = <AlertCircle size={16} />;
      tooltipText = `${errorCount} dependencies have errors`;
    } else {
      statusClass += ' partial';
      statusIcon = <Users size={16} />;
      tooltipText = `${runningCount}/${requiredCount} dependencies running`;
    }

    return (
      <div 
        className={statusClass} 
        title={tooltipText}
        onClick={(e) => {
          e.stopPropagation();
          setShowDependencyDetails(!showDependencyDetails);
        }}
      >
        {statusIcon}
        <span className="dependency-count">{runningCount}/{dependencies.length}</span>
      </div>
    );
  };

  // Determine item classes based on status
  let itemClass = 'service-item';
  if (status === 'running') {
    itemClass += ' service-item-running';
  } else if (status === 'starting') {
    itemClass += ' service-item-starting';
  } else if (status === 'failed' || statusError) {
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
              {/* Dependency Status Indicator */}
              {getDependencyStatusIndicator()}
            </div>
            <div className="service-meta">
              <div className={`service-status ${status === 'running' ? 'text-green' : statusError ? 'text-red' : 'text-gray'}`}>
                {statusIcon}
                {statusText}
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
              {/* Dependencies Section */}
              {dependencyStatus && dependencyStatus.dependencies.length > 0 && (
                <div className="dependencies-section">
                  <div className="section-header">
                    <h4 className="section-title">
                      <Users size={14} />
                      Dependencies ({dependencyStatus.runningCount}/{dependencyStatus.dependencies.length} running)
                    </h4>
                    <div className="dependency-actions">
                      {/* Show Dependencies Graph button */}
                      <button
                        className="action-button graph-button"
                        onClick={(e) => {
                          e.stopPropagation();
                          if (onShowDependencyGraph) {
                            onShowDependencyGraph(name);
                          }
                        }}
                        title="Show dependency graph"
                      >
                        <svg width="10" height="10" className="action-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                          <path d="M13 2L3 14H12L11 22L21 10H12L13 2Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                        </svg>
                        Show Graph
                      </button>

                      {status === 'running' && (
                        <button
                          className="action-button stop-chain-button"
                          onClick={handleStopDependencyChain}
                          disabled={isStopping}
                          title="Stop service and dependent services"
                        >
                          <Square size={10} />
                          Stop Chain
                        </button>
                      )}
                    </div>
                  </div>
                  
                  <div className="dependencies-list">
                    {dependencyStatus.dependencies.map((dep, index) => (
                      <div key={dep.serviceName} className={`dependency-item ${dep.status}`}>
                        <div className="dependency-info">
                          <span className="dependency-name">{dep.serviceName}</span>
                          <span className="dependency-type">{dep.type}</span>
                          {/* Show failure reason if available */}
                          {dep.failureReason && (
                            <div className="dependency-failure-reason">
                              <AlertCircle size={12} className="text-red-400" />
                              <span className="failure-text">{dep.failureReason}</span>
                            </div>
                          )}
                        </div>
                        <div className="dependency-status-badges">
                          <span className={`status-badge ${dep.status}`}>
                            {dep.status === 'running' ? '●' : 
                             dep.status === 'completed' ? '✓' :
                             dep.status === 'error' || dep.status === 'failed' ? '✕' : '○'}
                            {dep.status}
                          </span>
                          <span className={`health-badge ${dep.health}`}>
                            {dep.health}
                          </span>
                          {/* Quick logs access for failed dependencies */}
                          {((dep.status === 'error' || dep.status === 'failed') || dep.failureReason) && dep.hasLogs && (
                            <button
                              className="dependency-logs-button"
                              onClick={(e) => {
                                e.stopPropagation();
                                setSelectedDependencyForLogs(dep);
                                setShowLogsModal(true);
                              }}
                              title={`View logs for ${dep.serviceName}`}
                            >
                              <ScrollText size={12} />
                              Logs
                            </button>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Simple Dependencies Section (when no detailed status) */}
              {(!dependencyStatus || dependencyStatus.dependencies.length === 0) && dependencies.length > 0 && (
                <div className="dependencies-section">
                  <div className="section-header">
                    <h4 className="section-title">
                      <Users size={14} />
                      Dependencies ({dependencies.length})
                    </h4>
                    <div className="dependency-actions">
                      {/* Show Dependencies Graph button */}
                      <button
                        className="action-button graph-button"
                        onClick={(e) => {
                          e.stopPropagation();
                          if (onShowDependencyGraph) {
                            onShowDependencyGraph(name);
                          }
                        }}
                        title="Show dependency graph"
                      >
                        <svg width="10" height="10" className="action-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                          <path d="M13 2L3 14H12L11 22L21 10H12L13 2Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                        </svg>
                        Show Graph
                      </button>
                    </div>
                  </div>
                  
                  <div className="dependencies-list">
                    {dependencies.map((dep, index) => (
                      <div key={dep} className="dependency-item">
                        <div className="dependency-info">
                          <span className="dependency-name">{dep}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Existing Service Details */}
              <div className="details-grid">
                <div className="details-label">Type:</div>
                <div>{type}</div>
                <div className="details-label">Status:</div>
                <div className={status === 'running' ? 'text-green' : statusError ? 'text-red' : 'text-gray'}>
                  {statusText}
                </div>
                {dependencies.length > 0 && (
                  <>
                    <div className="details-label">Dependencies:</div>
                    <div>{dependencies.join(', ')}</div>
                  </>
                )}
              </div>
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
        {service.status !== 'running' && (
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
        {service.status === 'running' && (
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
        selectedDependency={selectedDependencyForLogs}
      />

      {/* Progress Modal */}
      <ProgressModal 
        isOpen={showProgressModal}
        onClose={() => setShowProgressModal(false)}
        serviceName={name}
        imageName={imageName}
      />
    </div>
  );
}

export default ServiceItem; 