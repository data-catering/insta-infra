import React, { useState, useEffect, useCallback } from 'react';
import { createPortal } from 'react-dom';
import ServiceList from './components/ServiceList';
import RunningServices from './components/RunningServices';
import { ImageStatusProvider, ImageLoader } from './components/ServiceItem';
import ConnectionModal from './components/ConnectionModal';
import LogsModal from './components/LogsModal';
import LogsPanel from './components/LogsPanel';
import RuntimeSetup from './components/RuntimeSetup';
import ErrorMessage, { useErrorHandler, ToastContainer } from './components/ErrorMessage';
import { 
  getAllServiceStatuses,
  getRunningServices,
  getRuntimeStatus,
  getCurrentRuntime,
  stopAllServices,
  listServices,
  openURL,
  getAppLogs,
  restartRuntime,
  shutdownApplication,
  wsClient,
  WS_MSG_TYPES
} from "./api/client";

function App() {
  // Error handling
  const { errors, toasts, addError, addToast, removeError, removeToast, clearAllErrors } = useErrorHandler();
  
  // Simple state management
  const [services, setServices] = useState([]);
  const [statuses, setStatuses] = useState({});
  const [runningServices, setRunningServices] = useState([]);
  const [dependencyStatuses, setDependencyStatuses] = useState({});
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);
  const [lastUpdated, setLastUpdated] = useState(new Date());
  const [isAnimating, setIsAnimating] = useState(false);
  const [isStoppingAll, setIsStoppingAll] = useState(false);
  const [showAbout, setShowAbout] = useState(false);
  const [showConnectionModal, setShowConnectionModal] = useState(false);
  const [showLogsModal, setShowLogsModal] = useState(false);
  const [selectedService, setSelectedService] = useState(null);
  const [copyFeedback, setCopyFeedback] = useState('');
  const [runtimeStatus, setRuntimeStatus] = useState(null);
  const [showRuntimeSetup, setShowRuntimeSetup] = useState(false);
  const [currentRuntime, setCurrentRuntime] = useState('');
  const [showLogsPanel, setShowLogsPanel] = useState(false);
  const [showActionsDropdown, setShowActionsDropdown] = useState(false);
  const [isShuttingDown, setIsShuttingDown] = useState(false);
  
  // Auto-clear copy feedback after 2 seconds
  useEffect(() => {
    if (copyFeedback) {
      const timer = setTimeout(() => {
        setCopyFeedback('');
      }, 2000);
      return () => clearTimeout(timer);
    }
  }, [copyFeedback]);
  
  useEffect(() => {
    // Initialize WebSocket connection
    initializeWebSocket();
    
    // Check runtime status first
    checkRuntimeStatus();

    // Cleanup WebSocket on unmount
    return () => {
      if (wsClient) {
        wsClient.disconnect();
      }
    };
  }, []);

  // Robust runtime availability monitoring - proactive approach
  useEffect(() => {
    let runtimeMonitorInterval;
    
    // Only start monitoring if we're not in setup mode
    if (!showRuntimeSetup) {
      runtimeMonitorInterval = setInterval(async () => {
        try {
          const status = await getRuntimeStatus();
          
          // If runtime becomes unavailable, transition to setup immediately
          if (!status.canProceed) {
            setRuntimeStatus(status);
            setShowRuntimeSetup(true);
            // Clear all data to prevent stale state
            setServices([]);
            setRunningServices([]);
            setStatuses({});
            setDependencyStatuses({});
            setError(null);
            setIsLoading(false);
            setIsAnimating(false);
          }
        } catch (error) {
          // If runtime check fails completely, assume runtime is unavailable
          setShowRuntimeSetup(true);
          setServices([]);
          setRunningServices([]);  
          setStatuses({});
          setDependencyStatuses({});
          setError(null);
          setIsLoading(false);
          setIsAnimating(false);
        }
      }, 3000); // Check every 3 seconds for responsive detection
    }
    
    return () => {
      if (runtimeMonitorInterval) {
        clearInterval(runtimeMonitorInterval);
      }
    };
  }, [showRuntimeSetup]);

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

  // Keyboard shortcut handler for Cmd+L (or Ctrl+L on non-Mac)
  useEffect(() => {
    const handleKeyDown = (event) => {
      // Check for Cmd+L on Mac or Ctrl+L on other platforms
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'l') {
        event.preventDefault();
        setShowLogsPanel(prev => !prev);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (showActionsDropdown && 
          !event.target.closest('.actions-dropdown') && 
          !event.target.closest('.dropdown-menu')) {
        setShowActionsDropdown(false);
      }
    };

    document.addEventListener('click', handleClickOutside);
    return () => document.removeEventListener('click', handleClickOutside);
  }, [showActionsDropdown]);

  const checkRuntimeStatus = async () => {
    try {
      const status = await getRuntimeStatus();
      setRuntimeStatus(status);
      
      if (status.canProceed) {
        // Runtime is available, proceed with normal app flow
        setShowRuntimeSetup(false);
        
        // Get the current runtime name
        try {
          const runtime = await getCurrentRuntime();
          // Extract the runtime name from the response object
          setCurrentRuntime(runtime?.runtime || runtime?.name || String(runtime) || 'unknown');
        } catch (err) {
          console.error('Failed to get current runtime:', err);
          setCurrentRuntime('unknown');
        }
        
        loadServicesAndStatuses();
      } else {
        // Runtime not available, show setup screen
        setShowRuntimeSetup(true);
      }
    } catch (err) {
      console.error('Failed to check runtime status:', err);
      // If we can't check runtime status, try to proceed anyway
      // This handles cases where the backend methods might not be available
      loadServicesAndStatuses();
    }
  };

  // Load services and their statuses using HTTP API
  const loadServicesAndStatuses = async (silent = false) => {
    try {
      if (!silent) {
        setIsLoading(true);
        setIsAnimating(true);
      }
      setError(null);
      
      // Get all available services
      const servicesList = await listServices();
      
      // Get all service statuses
      const statusesData = await getAllServiceStatuses();
      
      // Also get running services which includes dependency containers
      let runningServicesData = {};
      try {
        runningServicesData = await getRunningServices();
      } catch (runningServicesError) {
        console.warn('Failed to get running services data:', runningServicesError);
        // Don't handle runtime errors here - let the runtime monitor handle them
        // Just continue with empty running services data
      }
      
      // Convert services to the expected format
      const formattedServices = servicesList.map(service => ({
        Name: service.name || service.Name,
        Type: service.type || service.Type,
        Dependencies: service.recursive_dependencies || service.Dependencies || []
      }));
      
      // Set everything at once, but preserve existing services if new list is empty
      if (formattedServices.length > 0 || services.length === 0) {
        setServices(formattedServices || []);
      }
      
      // Merge statuses but preserve transition states like "starting"
      setStatuses(prevStatuses => {
        const newStatuses = { ...prevStatuses };
        
        // First, add running services data which includes dependency containers
        // This ensures dependency containers like postgres show correct status in service cards
        if (runningServicesData && runningServicesData.running_services) {
          Object.keys(runningServicesData.running_services).forEach(serviceName => {
            const statusValue = runningServicesData.running_services[serviceName];
            // Extract string status from object if needed
            if (typeof statusValue === 'object' && (statusValue.Status || statusValue.status)) {
              const extractedStatus = statusValue.Status || statusValue.status;
              newStatuses[serviceName] = extractedStatus;
            } else if (typeof statusValue === 'string') {
              newStatuses[serviceName] = statusValue;
            } else {
              newStatuses[serviceName] = statusValue;
            }
          });
        }
        
        // Then update statuses for services explicitly returned by the backend (but don't overwrite running services)
        if (statusesData) {
          Object.keys(statusesData).forEach(serviceName => {
            // Skip if we already have this service from running services (dependency containers)
            if (newStatuses.hasOwnProperty(serviceName)) {
              return;
            }
            
            const serverStatus = statusesData[serviceName];
            const currentStatus = prevStatuses[serviceName];
            
            // Don't override "starting" state with missing status from server
            // Only update if server explicitly reports a status
            if (typeof serverStatus === 'object' && serverStatus.Status) {
              newStatuses[serviceName] = serverStatus.Status;
            } else if (typeof serverStatus === 'string') {
              newStatuses[serviceName] = serverStatus;
            }
          });
          
          // For services not in the response, only set to "stopped" if they weren't in a transition or running state
          formattedServices.forEach(service => {
            const serviceName = service.Name;
            const currentStatus = prevStatuses[serviceName];
            
            // Don't override transition states (starting) or running states
            if (!statusesData[serviceName] && 
                !newStatuses[serviceName] && // Don't override running services data
                currentStatus !== 'starting' && 
                currentStatus !== 'running' && 
                !currentStatus?.includes('running')) {
              newStatuses[serviceName] = 'stopped';
            }
          });
        }
        
        return newStatuses;
      });
      
      // For dependency statuses, merge with existing dependency statuses to preserve WebSocket updates
      // Dependencies that don't exist in the response are implicitly "stopped"
      setDependencyStatuses(prevDeps => {
        const newDeps = { ...prevDeps };
        
        // Debug logging removed for production
        
        // First, merge in running services data which includes dependency containers
        // This ensures dependency containers get priority over main service data
        if (runningServicesData && runningServicesData.running_services) {
          Object.keys(runningServicesData.running_services).forEach(serviceName => {
            const statusValue = runningServicesData.running_services[serviceName];
            // Extract string status from object if needed
            if (typeof statusValue === 'object' && (statusValue.Status || statusValue.status)) {
              const extractedStatus = statusValue.Status || statusValue.status;
              newDeps[serviceName] = extractedStatus;
            } else if (typeof statusValue === 'string') {
              newDeps[serviceName] = statusValue;
            } else {
              newDeps[serviceName] = statusValue;
            }
          });
        }
        
        // Then update with main service statuses (but don't overwrite dependency containers)
        if (statusesData) {
          Object.keys(statusesData).forEach(serviceName => {
            // Skip if we already have this service from running services (dependency containers)
            if (newDeps.hasOwnProperty(serviceName)) {
              return;
            }
            
            const statusValue = statusesData[serviceName];
            // Extract string status from object if needed
            if (typeof statusValue === 'object' && (statusValue.Status || statusValue.status)) {
              newDeps[serviceName] = statusValue.Status || statusValue.status;
            } else if (typeof statusValue === 'string') {
              newDeps[serviceName] = statusValue;
            } else {
              newDeps[serviceName] = statusValue;
            }
          });
        }
        
        // Final dependency statuses now include running dependency containers
        
        // Don't remove dependency statuses that came from WebSocket but aren't in the main service list
        // This preserves dependency container statuses like postgres for keycloak
        return newDeps;
      });
      
      setLastUpdated(new Date());
      setIsLoading(false);
      
      setTimeout(() => {
        setIsAnimating(false);
      }, 300);
      
        } catch (error) {
      console.error("Error loading services:", error);
      
      // Don't parse error messages - let the runtime monitor handle runtime issues
      // Just show detailed error with metadata
      const errorMessage = error.message || "Failed to load services";
      
      addError({
        type: 'error',
        title: 'Failed to Load Services',
        message: errorMessage,
        details: `Runtime: ${currentRuntime || 'Unknown'} | Time: ${new Date().toLocaleTimeString()}`,
        metadata: {
          runtime: currentRuntime,
          timestamp: new Date().toISOString(),
          errorType: 'service_loading_error',
          endpoint: 'listServices/getAllServiceStatuses'
        },
        actions: [
          {
            label: 'Retry',
            onClick: () => {
              clearAllErrors();
              loadServicesAndStatuses();
            },
            variant: 'primary'
          },
          {
            label: 'Check Runtime',
            onClick: () => {
              clearAllErrors();
              checkRuntimeStatus();
            },
            variant: 'secondary'
          }
        ]
      });
      
      setError(errorMessage);
      setServices([]);
      setRunningServices([]);
      setDependencyStatuses({});
      setIsLoading(false);
      setIsAnimating(false);
    }
  };

  // Helper function to check if a status represents a running state
  const isRunningStatus = (status) => {
    if (!status) return false;
    const statusStr = String(status).toLowerCase();
    return statusStr === 'running' || 
           statusStr === 'running-healthy' || 
           statusStr === 'running-unhealthy' || 
           statusStr === 'running-health-starting' ||
           statusStr.startsWith('running');
  };

  // Compute running services whenever services or statuses change
  useEffect(() => {
    // Get latest running services - include all running variants
    const runningServicesList = services
      .filter(service => isRunningStatus(statuses[service.Name]))
      .map(service => ({
        ...service,
        name: service.Name,        // Add lowercase name for compatibility
        type: service.Type,        // Add lowercase type for compatibility 
        status: statuses[service.Name], // Preserve original status for detailed view
        dependencies: service.Dependencies || [] // Add dependencies for consistency with ServiceList
      }));

    // Also include dependencies that are currently running
    // This ensures that when a service starts its dependencies, they show up in the running services section
    const dependenciesAlsoRunning = [];
    const allServiceNames = new Set(runningServicesList.map(s => s.name));

    // Check dependency statuses and include running dependencies
    Object.keys(dependencyStatuses).forEach(depName => {
      const depStatus = dependencyStatuses[depName];
      const depStatusStr = typeof depStatus === 'object' ? depStatus.Status || depStatus.status : depStatus;
      
      // If this dependency is running and not already in the main running services list
      if (isRunningStatus(depStatusStr) && !allServiceNames.has(depName)) {
        // Try to find if this dependency corresponds to a known service
        const matchingService = services.find(s => s.Name === depName || (s.Dependencies && s.Dependencies.includes(depName)));
        
        if (matchingService && matchingService.Name === depName) {
          // This is a direct service that's running
          dependenciesAlsoRunning.push({
            ...matchingService,
            name: matchingService.Name,
            type: matchingService.Type,
            status: depStatusStr,
            dependencies: matchingService.Dependencies || []
          });
        } else {
          // This is a dependency container that's running (like postgres for keycloak)
          // Create a virtual service entry for it
          dependenciesAlsoRunning.push({
            Name: depName,
            name: depName,
            Type: 'dependency',
            type: 'dependency', 
            status: depStatusStr,
            dependencies: [],
            isDependency: true // Mark it as a dependency for special styling if needed
          });
        }
        allServiceNames.add(depName);
      }
    });

    // Combine main running services with running dependencies
    const allRunningServices = [...runningServicesList, ...dependenciesAlsoRunning];
    setRunningServices(allRunningServices);
  }, [services, statuses, dependencyStatuses]);

  // Service state change handler - can accept immediate status updates or trigger full refresh
  const fetchAllServices = async (immediateStatuses = null, silent = false) => {
    if (immediateStatuses) {
      // Use immediate status updates from backend
      // Merge with existing statuses to avoid losing other service statuses
      setStatuses(prevStatuses => {
        const newStatuses = { ...prevStatuses };
        
        // Handle both single service updates and multiple service updates
        Object.keys(immediateStatuses).forEach(serviceName => {
          const statusUpdate = immediateStatuses[serviceName];

          // Extract status from the update object
          if (typeof statusUpdate === 'object' && statusUpdate.Status) {
            newStatuses[serviceName] = statusUpdate.Status;
          } else if (typeof statusUpdate === 'string') {
            newStatuses[serviceName] = statusUpdate;
          }
        });
        
        return newStatuses;
      });
      
      // Also update dependency statuses when we get immediate updates
      // This ensures dependencies show their updated status when a service starts
      setDependencyStatuses(prevDeps => {
        const newDeps = { ...prevDeps };
        
        Object.keys(immediateStatuses).forEach(serviceName => {
          const statusUpdate = immediateStatuses[serviceName];
          if (statusUpdate) {
            newDeps[serviceName] = statusUpdate;
          }
        });
        
        return newDeps;
      });
      
      setLastUpdated(new Date());
    } else {
      // Perform full refresh
      loadServicesAndStatuses(silent);
    }
  };

  // Check if there are any running services
  const hasRunningServices = () => {
    return Object.values(statuses).some(status => isRunningStatus(status));
  };

  // Count running services
  const runningServicesCount = Object.values(statuses).filter(status => isRunningStatus(status)).length;

  // Format time in a more readable way
  const formatTime = (date) => {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  const handleStopAllServices = async () => {
    setIsStoppingAll(true);
    try {
      // Stop all services
      await stopAllServices();
      // Refresh service statuses
      fetchAllServices();
    } catch (error) {
      addError({
        type: 'error',
        title: 'Stop All Services Failed',
        message: 'Unable to stop all services',
        details: `Runtime: ${currentRuntime || 'Unknown'} | Running: ${runningServicesCount} services | Error: ${error.message} | Time: ${new Date().toLocaleTimeString()}`,
        metadata: {
          runtime: currentRuntime,
          runningServicesCount,
          timestamp: new Date().toISOString(),
          errorType: 'stop_all_services_error',
          originalError: error.message || error.toString()
        },
        actions: [
          {
            label: 'Retry Stop All',
            onClick: () => {
              clearAllErrors();
              handleStopAllServices();
            },
            variant: 'primary'
          },
          {
            label: 'Refresh Status',
            onClick: () => {
              clearAllErrors();
              fetchAllServices();
            },
            variant: 'secondary'
          }
        ]
      });
      // Fallback to full refresh on error
      fetchAllServices();
    } finally {
      setIsStoppingAll(false);
    }
  };

  const handleRestartRuntime = async () => {
    if (!currentRuntime) return;
    
    try {
      setIsLoading(true);
      await restartRuntime(currentRuntime);
      // Give runtime a moment to restart then refresh everything
      setTimeout(() => {
        checkRuntimeStatus();
        loadServicesAndStatuses();
      }, 2000);
    } catch (err) {
      console.error('Failed to restart runtime:', err);
      
      addError({
        type: 'error',
        title: 'Runtime Restart Failed',
        message: `Failed to restart ${currentRuntime}`,
        details: `Error: ${err.message} | Time: ${new Date().toLocaleTimeString()}`,
        metadata: {
          runtime: currentRuntime,
          timestamp: new Date().toISOString(),
          errorType: 'runtime_restart_error',
          originalError: err.message
        },
        actions: [
          {
            label: 'Retry Restart',
            onClick: () => {
              clearAllErrors();
              handleRestartRuntime();
            },
            variant: 'primary'
          },
          {
            label: 'Check Status',
            onClick: () => {
              clearAllErrors();
              checkRuntimeStatus();
            },
            variant: 'secondary'
          }
        ]
      });
      
      setError(`Failed to restart ${currentRuntime}: ${err.message}`);
      setIsLoading(false);
    }
  };

  const handleShutdownApplication = async () => {
    if (window.confirm('Are you sure you want to shut down the application? This will stop all running services and close insta-infra.')) {
      try {
        setIsShuttingDown(true);
        setShowActionsDropdown(false);
        await shutdownApplication();
        // The application will shut down, so no need to handle response
      } catch (error) {
        console.error('Failed to shutdown application:', error);
        setIsShuttingDown(false);
        addError({
          type: 'error',
          title: 'Application Shutdown Failed',
          message: 'Unable to shutdown the application properly',
          details: `Runtime: ${currentRuntime || 'Unknown'} | Error: ${error.message} | Time: ${new Date().toLocaleTimeString()}`,
          metadata: {
            runtime: currentRuntime,
            timestamp: new Date().toISOString(),
            errorType: 'application_shutdown_error',
            originalError: error.message || error.toString()
          },
          actions: [
            {
              label: 'Try Again',
              onClick: () => {
                clearAllErrors();
                handleShutdownApplication();
              },
              variant: 'primary'
            },
            {
              label: 'Force Close',
              onClick: () => {
                clearAllErrors();
                window.close();
              },
              variant: 'secondary'
            }
          ]
        });
      }
    }
  };

  // Handle runtime becoming ready
  const handleRuntimeReady = () => {
      setShowRuntimeSetup(false);
      loadServicesAndStatuses();
  };

  // Handle copying text to clipboard
  const copyToClipboard = async (text) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopyFeedback('Copied to clipboard');
    } catch (error) {
      console.error('Failed to copy to clipboard:', error);
      const textArea = document.createElement('textarea');
      textArea.value = text;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
      setCopyFeedback('Failed to copy');
    }
  };

  // Handle opening external URLs in the browser
  const openExternalLink = (url) => {
    try {
      openURL(url);
    } catch (error) {
      console.error('Failed to open URL:', error);
      // Fallback: Copy URL to clipboard
      copyToClipboard(url);
      setCopyFeedback('URL copied to clipboard');
    }
  };

  // Show runtime setup if needed
  if (showRuntimeSetup) {
    return <RuntimeSetup onRuntimeReady={handleRuntimeReady} />;
  }

  // Show shutdown page if shutting down
  if (isShuttingDown) {
    return (
      <div className="shutdown-container">
        <div className="shutdown-content">
          <div className="shutdown-icon-container">
            <svg width="48" height="48" className="shutdown-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M18.36 6.64C19.6184 7.8988 20.4753 9.50246 20.8223 11.2482C21.1693 12.994 20.9909 14.8034 20.3096 16.4478C19.6284 18.0921 18.4748 19.4976 16.9948 20.4864C15.5148 21.4752 13.7749 22.0029 11.995 22.0029C10.2151 22.0029 8.47515 21.4752 6.99517 20.4864C5.51519 19.4976 4.36164 18.0921 3.68036 16.4478C2.99909 14.8034 2.82069 12.994 3.16772 11.2482C3.51475 9.50246 4.37162 7.8988 5.63 6.64L12 2L18.36 6.64Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              <path d="M12 8V16" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </div>
          <h1 className="shutdown-title">Application Shutdown</h1>
          <p className="shutdown-message">insta-infra has been shut down successfully.</p>
        </div>
      </div>
    );
  }

  return (
    <div className={`app-container ${showLogsPanel ? 'logs-visible' : ''}`}>
      {/* Header */}
      <header className="header">
        <div className="container header-content">
          <div className="logo-container">
            <div className="logo">
              <svg width="24" height="24" className="logo-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M21 9.5V5C21 3.89543 20.1046 3 19 3H5C3.89543 3 3 3.89543 3 5V9.5M21 9.5H3M21 9.5V19C21 20.1046 20.1046 21 19 21H5C3.89543 21 3 20.1046 3 19V9.5M9 14H15" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            </div>
            <h1 className="logo-text">
              Insta<span>Infra</span>
            </h1>
          </div>
          
          <div className="header-actions">
            {/* Status indicator */}
            <div className="status-bar">
              {/* Last update time */}
              <div className="status-indicator">
                <svg width="16" height="16" className="status-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M12 8V12L15 15M12 3C7.02944 3 3 7.02944 3 12C3 16.9706 7.02944 21 12 21C16.9706 21 21 16.9706 21 12C21 7.02944 16.9706 3 12 3Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
                <span>Updated {formatTime(lastUpdated)}</span>
              </div>
              
              <div className="status-separator"></div>
              
              {/* Running services count */}
              <div className="status-indicator">
                <span className={`status-dot ${runningServicesCount > 0 ? 'dot-green pulse' : 'dot-gray'}`}></span>
                <span>{runningServicesCount} services running</span>
              </div>
              
              {isLoading && (
                <>
                  <div className="status-separator"></div>
                  <div className="status-indicator text-blue">
                    <div className="status-icon spin"></div>
                    <span>Syncing...</span>
                  </div>
                </>
              )}
            </div>

            {/* Stop All button - only show if there are running services */}
            {hasRunningServices() && (
              <button 
                onClick={handleStopAllServices} 
                disabled={isLoading || isStoppingAll}
                className={`button button-secondary ${isStoppingAll ? 'button-loading' : ''}`}
                title="Stop all running services"
              >
                {isStoppingAll ? (
                  <>
                    <svg width="16" height="16" className="button-icon spin" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    <span className="sr-only md-visible">Stopping All...</span>
                  </>
                ) : (
                  <>
                    <svg width="16" height="16" className="button-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <path d="M6 6H18V18H6V6Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                    </svg>
                    <span className="sr-only md-visible">Stop All</span>
                  </>
                )}
              </button>
            )}

            {/* Actions Dropdown */}
            <div className="actions-dropdown">
              <button 
                onClick={() => setShowActionsDropdown(!showActionsDropdown)}
                className="button button-primary"
                title="More actions"
              >
                <svg width="16" height="16" className="button-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12C11 12.5523 11.4477 13 12 13Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  <path d="M12 6C12.5523 6 13 5.55228 13 5C13 4.44772 12.5523 4 12 4C11.4477 4 11 4.44772 11 5C11 5.55228 11.4477 6 12 6Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  <path d="M12 20C12.5523 20 13 19.5523 13 19C13 18.4477 12.5523 18 12 18C11.4477 18 11 18.4477 11 19C11 19.5523 11.4477 20 12 20Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
                <span className="sr-only md-visible">Actions</span>
              </button>
              
              {showActionsDropdown && (
                <div className="dropdown-menu">
                  <button 
                    onClick={() => {
                      setShowAbout(true);
                      setShowActionsDropdown(false);
                    }}
                    className="dropdown-item"
                  >
                    <svg width="16" height="16" className="dropdown-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <path d="M13 16H12V12H11M12 8H12.01M21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3C16.9706 3 21 7.02944 21 12Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                    </svg>
                    About
                  </button>
                  
                  <button 
                    onClick={() => {
                      handleRestartRuntime();
                      setShowActionsDropdown(false);
                    }}
                    disabled={isLoading || !currentRuntime}
                    className="dropdown-item"
                  >
                    <svg width="16" height="16" className="dropdown-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <path d="M4 4V9H4.58152M19.9381 11C19.446 7.05369 16.0796 4 12 4C8.64262 4 5.76829 6.06817 4.58152 9M4.58152 9H9M20 20V15H19.4185M19.4185 15C18.2317 17.9318 15.3574 20 12 20C7.92038 20 4.55399 16.9463 4.06189 13M19.4185 15H15" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                    </svg>
                    Restart
                  </button>
                  
                  <div className="dropdown-separator"></div>
                  
                  <button 
                    onClick={() => {
                      handleShutdownApplication();
                      setShowActionsDropdown(false);
                    }}
                    className="dropdown-item dropdown-item-danger"
                  >
                    <svg width="16" height="16" className="dropdown-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <path d="M18.36 6.64C19.6184 7.8988 20.4753 9.50246 20.8223 11.2482C21.1693 12.994 20.9909 14.8034 20.3096 16.4478C19.6284 18.0921 18.4748 19.4976 16.9948 20.4864C15.5148 21.4752 13.7749 22.0029 11.995 22.0029C10.2151 22.0029 8.47515 21.4752 6.99517 20.4864C5.51519 19.4976 4.36164 18.0921 3.68036 16.4478C2.99909 14.8034 2.82069 12.994 3.16772 11.2482C3.51475 9.50246 4.37162 7.8988 5.63 6.64L12 2L18.36 6.64Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                      <path d="M12 8V16" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                    </svg>
                    Shutdown
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>
      </header>

      {/* Main Content Area */}
      <main className="container main-content">
        {/* Error display */}
        {error && (
          <div className="error-alert fade-in" role="alert">
            <div className="error-content">
              <svg width="24" height="24" className="error-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M12 9V13M12 17H12.01M4.98207 19H19.0179C20.5615 19 21.5233 17.3256 20.7455 15.9923L13.7276 3.96153C12.9558 2.63852 11.0442 2.63852 10.2724 3.96153L3.25452 15.9923C2.47675 17.3256 3.43849 19 4.98207 19Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
              <div>
                <p className="error-title">Connection Error</p>
                <p className="error-message">{error}</p>
              </div>
              <button 
                onClick={() => setError(null)} 
                className="error-close"
              >
                <svg width="16" height="16" className="close-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M6 18L18 6M6 6L18 18" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </button>
            </div>
          </div>
        )}
        
        {/* Loading state */}
        {isLoading && services.length === 0 && !error && (
          <div className="loading-container">
            <div className="loading-content fade-in">
              <div className="loading-spinner-container">
                <div className="loading-spinner"></div>
              </div>
              <p className="loading-text">Loading infrastructure services...</p>
              <p className="loading-subtext">Discovering available services and their status</p>
            </div>
          </div>
        )}
        
        {!isLoading && services.length === 0 && !error && (
          <div className="empty-state-container">
            <div className="empty-state-content">
              <svg width="56" height="56" className="empty-state-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M9.4 16.6L4.8 12L9.4 7.4L8 6L2 12L8 18L9.4 16.6ZM14.6 16.6L19.2 12L14.6 7.4L16 6L22 12L16 18L14.6 16.6Z" fill="currentColor"/>
              </svg>
              <p className="empty-state-title">No services found</p>
              <p className="empty-state-description">There may be a configuration issue or services are not properly set up</p>
              <button 
                onClick={fetchAllServices}
                className="button button-primary"
              >
                <svg width="16" height="16" className="button-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M4 4V9H4.58152M19.9381 11C19.446 7.05369 16.0796 4 12 4C8.64262 4 5.76829 6.06817 4.58152 9M4.58152 9H9M20 20V15H19.4185M19.4185 15C18.2317 17.9318 15.3574 20 12 20C7.92038 20 4.55399 16.9463 4.06189 13M19.4185 15H15" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
                Retry
              </button>
            </div>
          </div>
        )}
        
        {/* Service sections */}
        {(services.length > 0 || isLoading) && (
          <ImageStatusProvider services={services}>
            <ImageLoader services={services} />
            <div className={`services-container ${isAnimating ? 'services-loading' : ''}`}>
              {/* Running Services Section - only show when there are running services */}
              {runningServices.length > 0 && (
                <RunningServices 
                  services={runningServices} 
                  isLoading={false} 
                  onServiceStateChange={fetchAllServices} 
                  statuses={statuses}
                  dependencyStatuses={dependencyStatuses}
                  currentRuntime={currentRuntime}
                />
              )}
        
              {/* Active Services Section */}
              <ServiceList 
                services={services} 
                statuses={statuses}
                dependencyStatuses={dependencyStatuses}
                isLoading={isLoading} 
                onServiceStateChange={fetchAllServices}
                currentRuntime={currentRuntime}
              />
            </div>
          </ImageStatusProvider>
        )}
      </main>

      {/* Global Error Messages */}
      <div className="global-errors">
        {errors.map((error) => (
          <ErrorMessage
            key={error.id}
            type={error.type || 'error'}
            title={error.title}
            message={error.message}
            details={error.details}
            actions={error.actions}
            onDismiss={() => removeError(error.id)}
            autoHide={error.type === 'success' || error.type === 'info'}
            autoHideDelay={error.type === 'success' ? 3000 : 5000}
            className="app-error-message"
          />
        ))}
      </div>

      {/* Toast Notifications */}
      <ToastContainer 
        toasts={toasts}
        onRemoveToast={removeToast}
      />

      {/* Footer */}
      <footer className="footer">
        <div className="container footer-content">
          <div className="footer-status">
            <div className="status-dot dot-green pulse"></div>
            <p>Infrastructure Control Panel</p>
          </div>
          <div className="footer-info">
            Using {currentRuntime} • Last refreshed: {lastUpdated.toLocaleString()}
          </div>
        </div>
      </footer>

      {/* About Modal */}
      {showAbout && (
        createPortal(
          <div className="connection-modal-overlay" onClick={() => setShowAbout(false)}>
            <div className="connection-modal about-modal" onClick={(e) => e.stopPropagation()}>
              <div className="connection-modal-header">
                <h3 className="connection-modal-title">
                  About insta-infra
                </h3>
                <button 
                  onClick={() => setShowAbout(false)}
                  className="connection-modal-close"
                >
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M6 18L18 6M6 6L18 18" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                </button>
              </div>
              
              <div className="connection-modal-content">
                <div className="about-content">
                  <div className="about-hero">
                    <div className="about-logo">
                      <svg width="48" height="48" className="logo-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path d="M21 9.5V5C21 3.89543 20.1046 3 19 3H5C3.89543 3 3 3.89543 3 5V9.5M21 9.5H3M21 9.5V19C21 20.1046 20.1046 21 19 21H5C3.89543 21 3 20.1046 3 19V9.5M9 14H15" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                      </svg>
                    </div>
                    <h2 className="about-title">insta-infra</h2>
                    <p className="about-subtitle">Instant Infrastructure Services</p>
                  </div>

                  <div className="about-description">
                    <p>
                      <strong>insta-infra</strong> is a powerful service management tool that allows you to quickly start, 
                      stop, and manage infrastructure services like databases, monitoring tools, and data platforms.
                    </p>
                    <p>
                      With support for over 70 services including PostgreSQL, Redis, Grafana, Elasticsearch, 
                      and many more, insta-infra simplifies development and testing workflows.
                    </p>
                  </div>

                  <div className="about-features">
                    <h4>Key Features:</h4>
                    <ul>
                      <li>
                        <svg width="16" height="16" className="feature-icon" fill="currentColor" viewBox="0 0 20 20">
                          <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                        </svg>
                        One-click service deployment with Docker/Podman
                      </li>
                      <li>
                        <svg width="16" height="16" className="feature-icon" fill="currentColor" viewBox="0 0 20 20">
                          <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                        </svg>
                        Automatic dependency management
                      </li>
                      <li>
                        <svg width="16" height="16" className="feature-icon" fill="currentColor" viewBox="0 0 20 20">
                          <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                        </svg>
                        Data persistence options
                      </li>
                      <li>
                        <svg width="16" height="16" className="feature-icon" fill="currentColor" viewBox="0 0 20 20">
                          <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                        </svg>
                        Easy connection details and web UI access
                      </li>
                      <li>
                        <svg width="16" height="16" className="feature-icon" fill="currentColor" viewBox="0 0 20 20">
                          <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                        </svg>
                        Cross-platform support (macOS, Linux, Windows)
                      </li>
                    </ul>
                  </div>

                  <div className="about-links">
                    <h4>Resources:</h4>
                    <div className="link-buttons">
                      <button 
                        onClick={() => openExternalLink("https://github.com/data-catering/insta-infra")}
                        className="link-button"
                        title="Open GitHub repository in your browser"
                      >
                        <svg width="16" height="16" className="link-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                          <path d="M9 19C4 20.5 4 16.5 2 16M22 16V19C22 20.1046 21.1046 21 20 21H4C2.89543 21 2 20.1046 2 19V16M22 16L18.5 12.5L22 9V16Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                        </svg>
                        GitHub Repository
                      </button>
                      <button 
                        onClick={() => openExternalLink("https://github.com/data-catering/insta-infra/blob/main/README.md")}
                        className="link-button"
                        title="Open documentation in your browser"
                      >
                        <svg width="16" height="16" className="link-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                          <path d="M9 12H15M9 16H15M17 21H7C5.89543 21 5 20.1046 5 19V5C5 3.89543 5.89543 3 7 3H12.5858C12.851 3 13.1054 3.10536 13.2929 3.29289L18.7071 8.70711C18.8946 8.89464 19 9.149 19 9.41421V19C19 20.1046 18.1046 21 17 21Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                        </svg>
                        Documentation
                      </button>
                    </div>
                  </div>

                  <div className="about-footer">
                    <p className="about-version">Version 1.0.0</p>
                    <p className="about-copyright">Built with ❤️ for developers</p>
                  </div>
                </div>
              </div>
            </div>
          </div>,
          document.body
        )
      )}
      
      {/* Connection Info Modal */}
      <ConnectionModal 
        isOpen={showConnectionModal}
        onClose={() => setShowConnectionModal(false)}
        copyFeedback={copyFeedback}
        onCopyToClipboard={copyToClipboard}
      />
      
      {/* Logs Modal */}
      {showLogsModal && (
        <LogsModal 
          isOpen={true}
          onClose={() => setShowLogsModal(false)}
          serviceName={selectedService}
        />
      )}
      
      {/* Logs Panel */}
      <LogsPanel 
        isVisible={showLogsPanel}
        onToggle={() => setShowLogsPanel(!showLogsPanel)}
      />
    </div>
  );
}

export default App;
