import React, { useEffect, useCallback } from 'react';
import AppHeader from './components/AppHeader';
import AppFooter from './components/AppFooter';
import MainContent from './components/MainContent';
import AppModals from './components/AppModals';
import RuntimeSetup from './components/RuntimeSetup';
import ErrorMessage, { useErrorHandler, ToastContainer } from './components/ErrorMessage';
import { useAppState } from './hooks/useAppState';
import { useWebSocket } from './hooks/useWebSocket';
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
  
  // App state management
  const {
    services, setServices,
    statuses, setStatuses, 
    runningServices, setRunningServices,
    dependencyStatuses, setDependencyStatuses,
    isLoading, setIsLoading,
    error, setError,
    lastUpdated, setLastUpdated,
    isAnimating, setIsAnimating,
    isStoppingAll, setIsStoppingAll,
    showAbout, setShowAbout,
    showConnectionModal, setShowConnectionModal,
    showLogsModal, setShowLogsModal,
    selectedService, setSelectedService,
    copyFeedback, setCopyFeedback,
    runtimeStatus, setRuntimeStatus,
    showRuntimeSetup, setShowRuntimeSetup,
    currentRuntime, setCurrentRuntime,
    showLogsPanel, setShowLogsPanel,
    showActionsDropdown, setShowActionsDropdown,
    isShuttingDown, setIsShuttingDown
  } = useAppState();
  
  // Auto-clear copy feedback after 2 seconds
  useEffect(() => {
    if (copyFeedback) {
      const timer = setTimeout(() => {
        setCopyFeedback('');
      }, 2000);
      return () => clearTimeout(timer);
    }
  }, [copyFeedback]);
  
  // Initialize WebSocket
  useWebSocket({ setStatuses, setDependencyStatuses, setLastUpdated });

  useEffect(() => {
    // Check runtime status first
    checkRuntimeStatus();
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
    return <RuntimeSetup isOpen={showRuntimeSetup} runtimeStatus={runtimeStatus} onRuntimeReady={handleRuntimeReady} />;
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
      <AppHeader 
        lastUpdated={lastUpdated}
        runningServicesCount={runningServicesCount}
        isLoading={isLoading}
        hasRunningServices={hasRunningServices()}
        isStoppingAll={isStoppingAll}
        handleStopAllServices={handleStopAllServices}
        showActionsDropdown={showActionsDropdown}
        setShowActionsDropdown={setShowActionsDropdown}
        setShowAbout={setShowAbout}
        handleRestartRuntime={handleRestartRuntime}
        handleShutdownApplication={handleShutdownApplication}
        currentRuntime={currentRuntime}
      />

      <MainContent
        error={error}
        setError={setError}
        isLoading={isLoading}
        services={services}
        runningServices={runningServices}
        statuses={statuses}
        dependencyStatuses={dependencyStatuses}
        currentRuntime={currentRuntime}
        fetchAllServices={fetchAllServices}
        isAnimating={isAnimating}
      />

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

      <AppFooter currentRuntime={currentRuntime} lastUpdated={lastUpdated} />
      
      <AppModals
        showAbout={showAbout}
        setShowAbout={setShowAbout}
        showConnectionModal={showConnectionModal}
        setShowConnectionModal={setShowConnectionModal}
        showLogsModal={showLogsModal}
        setShowLogsModal={setShowLogsModal}
        selectedService={selectedService}
        showLogsPanel={showLogsPanel}
        setShowLogsPanel={setShowLogsPanel}
        copyFeedback={copyFeedback}
        copyToClipboard={copyToClipboard}
        openExternalLink={openExternalLink}
      />
    </div>
  );
}

export default App;
