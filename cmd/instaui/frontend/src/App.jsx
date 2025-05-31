import React, { useState, useEffect } from 'react';
import { createPortal } from 'react-dom';
import RunningServices from './components/RunningServices';
import ServiceList from './components/ServiceList';
import DependencyGraphModal from './components/DependencyGraphModal';
import ConnectionModal from './components/ConnectionModal';
import LogsModal from './components/LogsModal';
import RuntimeSetup from './components/RuntimeSetup';
import { ListServices, GetAllRunningServices, GetAllServicesWithStatusAndDependencies, StopAllServices, StartServiceWithStatusUpdate, StopService, GetServiceConnectionInfo, GetRuntimeStatus, GetCurrentRuntime } from "../wailsjs/go/main/App";
import './App.css';
import './styles.css'; // Import our new CSS

function App() {
  // Simple state management
  const [services, setServices] = useState([]);
  const [statuses, setStatuses] = useState({});
  const [runningServices, setRunningServices] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [lastUpdated, setLastUpdated] = useState(new Date());
  const [isAnimating, setIsAnimating] = useState(false);
  const [isStoppingAll, setIsStoppingAll] = useState(false);
  const [showAbout, setShowAbout] = useState(false);
  const [selectedServiceForGraph, setSelectedServiceForGraph] = useState(null);
  const [showConnectionInfo, setShowConnectionInfo] = useState(false);
  const [connectionInfo, setConnectionInfo] = useState(null);
  const [showLogsModal, setShowLogsModal] = useState(null);
  const [copyFeedback, setCopyFeedback] = useState('');
  const [runtimeStatus, setRuntimeStatus] = useState(null);
  const [showRuntimeSetup, setShowRuntimeSetup] = useState(false);
  const [currentRuntime, setCurrentRuntime] = useState('unknown');
  
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
    // Check runtime status first
    checkRuntimeStatus();
  }, []);

  const checkRuntimeStatus = async () => {
    try {
      const status = await GetRuntimeStatus();
      setRuntimeStatus(status);
      
      if (status.canProceed) {
        // Runtime is available, proceed with normal app flow
        setShowRuntimeSetup(false);
        
        // Get the current runtime name
        try {
          const runtime = await GetCurrentRuntime();
          setCurrentRuntime(runtime);
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

  // Highly efficient approach: Load services, statuses, and dependencies in one optimized call
  const loadServicesAndStatuses = async (silent = false) => {
    try {
      if (!silent) {
        setIsLoading(true);
        setIsAnimating(true);
      }
      setError(null);
      
      // Get everything with the new highly optimized method
      const serviceDetails = await GetAllServicesWithStatusAndDependencies();
      
      // Extract services and statuses from the detailed response
      const servicesList = serviceDetails.map(detail => ({
        Name: detail.name,
        Type: detail.type,
        Dependencies: detail.dependencies || [] // Include dependencies from backend response
      }));
      
      const statusesOnly = {};
      serviceDetails.forEach(detail => {
        statusesOnly[detail.name] = detail.status;
      });
      
      // Set everything at once, but preserve existing services if new list is empty
      if (servicesList.length > 0 || services.length === 0) {
        setServices(servicesList || []);
      }
      
      setStatuses(statusesOnly);
      setLastUpdated(new Date());
      setIsLoading(false);
      
      setTimeout(() => {
        setIsAnimating(false);
      }, 300);
      
    } catch (error) {
      console.error("Error loading services:", error);
      setError(error.message || "Failed to load services");
      setServices([]);
      setRunningServices([]);
      setIsLoading(false);
      setIsAnimating(false);
    }
  };

  // Compute running services whenever services or statuses change
  useEffect(() => {
    const runningServicesList = services
      .filter(service => statuses[service.Name] === 'running')
      .map(service => ({
        ...service,
        name: service.Name,        // Add lowercase name for compatibility
        type: service.Type,        // Add lowercase type for compatibility 
        status: 'running',         // Add status for compatibility
        dependencies: service.Dependencies || [] // Add dependencies for consistency with ServiceList
      }));
    setRunningServices(runningServicesList);
  }, [services, statuses]);

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
      setLastUpdated(new Date());
    } else {
      // Perform full refresh
      loadServicesAndStatuses(silent);
    }
  };

  // Check if any starting services have completed their transition
  // Note: This functionality is now handled by individual ServiceItem components
  // through their own polling mechanisms for better event-driven architecture
  const checkStartingServicesProgress = async () => {
    // This function is deprecated in favor of ServiceItem-level status management
    // Individual ServiceItem components now handle their own status polling
    return false;
  };

  // Check if there are any services currently starting
  const hasStartingServices = () => {
    return Object.values(statuses).some(status => status === 'starting');
  };

  // Format time in a more readable way
  const formatTime = (date) => {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  // Handle stopping all services
  const handleStopAllServices = async () => {
    setIsStoppingAll(true);
    try {
      await StopAllServices();
      // Refresh service data after stopping all
      fetchAllServices();
    } catch (error) {
      console.error("Failed to stop all services:", error);
      alert(`Failed to stop all services: ${error.message || error}`);
    } finally {
      setIsStoppingAll(false);
    }
  };

  // Handle service actions from dependency graph modal
  const handleGraphServiceAction = async (action, serviceName) => {
    try {
      switch (action) {
        case 'start':
          const updatedStatuses = await StartServiceWithStatusUpdate(serviceName, false); // Start without persistence for now
          fetchAllServices(updatedStatuses);
          break;
        case 'stop':
          await StopService(serviceName);
          fetchAllServices(); // Full refresh for stop action
          break;
        case 'info':
          // Show connection info modal
          const connectionData = await GetServiceConnectionInfo(serviceName);
          setConnectionInfo(connectionData);
          setShowConnectionInfo(true);
          break;
        case 'logs':
          // Show logs modal
          setShowLogsModal(serviceName);
          break;
        default:
          break;
      }
    } catch (error) {
      console.error(`Failed to ${action} service ${serviceName}:`, error);
      alert(`Failed to ${action} service ${serviceName}: ${error.message || error}`);
    }
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

  // Handle showing dependency graph for a specific service
  const handleShowServiceDependencyGraph = (serviceName) => {
    setSelectedServiceForGraph(serviceName);
  };

  // Handle runtime becoming ready
  const handleRuntimeReady = async () => {
    setShowRuntimeSetup(false);
    
    // Get the current runtime name after setup
    try {
      const runtime = await GetCurrentRuntime();
      setCurrentRuntime(runtime);
    } catch (err) {
      console.error('Failed to get current runtime after setup:', err);
      setCurrentRuntime('unknown');
    }
    
    checkRuntimeStatus(); // Re-check and proceed with normal flow
  };

  // Show runtime setup if needed
  if (showRuntimeSetup) {
    return <RuntimeSetup onRuntimeReady={handleRuntimeReady} />;
  }

  return (
    <div className="app-container">
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
                <span className={`status-dot ${runningServices.length > 0 ? 'dot-green pulse' : 'dot-gray'}`}></span>
                <span>{runningServices.length} services running</span>
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
            
            {/* About button */}
            <button 
              onClick={() => setShowAbout(true)}
              className="button button-primary"
              title="About insta-infra"
            >
              <svg width="16" height="16" className="button-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M13 16H12V12H11M12 8H12.01M21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3C16.9706 3 21 7.02944 21 12Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
              <span className="sr-only md-visible">About</span>
            </button>
            
            {/* Refresh button */}
            <button 
              onClick={() => fetchAllServices()} 
              disabled={isLoading}
              className={`button button-primary ${isLoading ? 'button-loading' : ''}`}
              title="Refresh service status"
            >
              {isLoading ? (
                <>
                  <div className="button-icon spin"></div>
                  <span className="sr-only md-visible">Refreshing...</span>
                </>
              ) : (
                <>
                  <svg width="16" height="16" className="button-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M4 4V9H4.58152M19.9381 11C19.446 7.05369 16.0796 4 12 4C8.64262 4 5.76829 6.06817 4.58152 9M4.58152 9H9M20 20V15H19.4185M19.4185 15C18.2317 17.9318 15.3574 20 12 20C7.92038 20 4.55399 16.9463 4.06189 13M19.4185 15H15" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                  <span className="sr-only md-visible">Refresh</span>
                </>
              )}
            </button>

            {/* Check Progress button - only show when there are starting services */}
            {hasStartingServices() && (
              <button 
                onClick={checkStartingServicesProgress}
                className="button button-secondary"
                title="Check if starting services have finished"
              >
                <svg width="16" height="16" className="button-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <span className="sr-only md-visible">Check Progress</span>
              </button>
            )}

            {/* Stop All button - only show if there are running services */}
            {runningServices.length > 0 && (
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
                <svg width="48" height="48" className="loading-spinner" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
              </div>
              <p className="loading-text">Loading infrastructure services...</p>
              <p className="loading-subtext">This may take a moment</p>
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
          <div className={`services-container ${isAnimating ? 'services-loading' : ''}`}>
            {/* Running Services Section */}
            <RunningServices services={runningServices} isLoading={isLoading} onServiceStateChange={fetchAllServices} onShowDependencyGraph={handleShowServiceDependencyGraph} />
    
            {/* All Services Section */}
            <ServiceList 
              services={services} 
              statuses={statuses}
              isLoading={isLoading} 
              onServiceStateChange={fetchAllServices} 
              onShowDependencyGraph={handleShowServiceDependencyGraph} 
            />
          </div>
        )}
      </main>

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
                      <a 
                        href="https://github.com/your-org/insta-infra" 
                        target="_blank" 
                        rel="noopener noreferrer"
                        className="link-button"
                      >
                        <svg width="16" height="16" className="link-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                          <path d="M9 19C4 20.5 4 16.5 2 16M22 16V19C22 20.1046 21.1046 21 20 21H4C2.89543 21 2 20.1046 2 19V16M22 16L18.5 12.5L22 9V16Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                        </svg>
                        GitHub Repository
                      </a>
                      <a 
                        href="https://github.com/your-org/insta-infra/blob/main/README.md" 
                        target="_blank" 
                        rel="noopener noreferrer"
                        className="link-button"
                      >
                        <svg width="16" height="16" className="link-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                          <path d="M9 12H15M9 16H15M17 21H7C5.89543 21 5 20.1046 5 19V5C5 3.89543 5.89543 3 7 3H12.5858C12.851 3 13.1054 3.10536 13.2929 3.29289L18.7071 8.70711C18.8946 8.89464 19 9.149 19 9.41421V19C19 20.1046 18.1046 21 17 21Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                        </svg>
                        Documentation
                      </a>
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
      
      {/* Dependency Graph Modal */}
      <DependencyGraphModal 
        isOpen={selectedServiceForGraph !== null}
        onClose={() => setSelectedServiceForGraph(null)}
        serviceName={selectedServiceForGraph}
        onServiceAction={handleGraphServiceAction}
      />
      
      {/* Connection Info Modal */}
      <ConnectionModal 
        isOpen={showConnectionInfo}
        onClose={() => setShowConnectionInfo(false)}
        connectionInfo={connectionInfo}
        copyFeedback={copyFeedback}
        onCopyToClipboard={copyToClipboard}
      />
      
      {/* Logs Modal */}
      {showLogsModal && (
        <LogsModal 
          isOpen={true}
          onClose={() => setShowLogsModal(null)}
          serviceName={showLogsModal}
        />
      )}
    </div>
  );
}

export default App;
