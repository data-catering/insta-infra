import React, { useState, useEffect } from 'react';
import { GetRuntimeStatus, AttemptStartRuntime, WaitForRuntimeReady, ReinitializeRuntime, SetCustomDockerPath, SetCustomPodmanPath, GetCustomDockerPath, GetCustomPodmanPath, GetAppLogs } from '../../wailsjs/go/main/App';

const RuntimeSetup = ({ onRuntimeReady }) => {
  const [runtimeStatus, setRuntimeStatus] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isStarting, setIsStarting] = useState(false);
  const [startupProgress, setStartupProgress] = useState('');
  const [error, setError] = useState('');
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [customDockerPath, setCustomDockerPath] = useState('');
  const [customPodmanPath, setCustomPodmanPath] = useState('');
  const [pathError, setPathError] = useState('');
  const [showLogs, setShowLogs] = useState(false);
  const [appLogs, setAppLogs] = useState([]);

  useEffect(() => {
    checkRuntimeStatus();
    loadCustomPaths();
  }, []);

  const loadCustomPaths = async () => {
    try {
      const dockerPath = await GetCustomDockerPath();
      const podmanPath = await GetCustomPodmanPath();
      setCustomDockerPath(dockerPath || '');
      setCustomPodmanPath(podmanPath || '');
    } catch (err) {
      console.error('Failed to load custom paths:', err);
    }
  };

  const loadAppLogs = async () => {
    try {
      const logs = await GetAppLogs();
      setAppLogs(logs);
    } catch (err) {
      console.error('Failed to load app logs:', err);
    }
  };

  const checkRuntimeStatus = async () => {
    try {
      setIsLoading(true);
      setError('');
      const status = await GetRuntimeStatus();
      setRuntimeStatus(status);
    } catch (err) {
      console.error('Failed to check runtime status:', err);
      setError(`Failed to check runtime status: ${err.message || err}`);
    } finally {
      setIsLoading(false);
    }
  };

  const handleQuickStart = async () => {
    // Try Docker first as the default choice
    const dockerRuntime = runtimeStatus?.runtimeStatuses?.find(rt => rt.name === 'docker');
    
    if (dockerRuntime?.isInstalled && dockerRuntime?.canAutoStart) {
      await handleStartRuntime('docker');
    } else if (dockerRuntime?.isInstalled) {
      setError('Docker is installed but requires manual startup. Please start Docker Desktop manually.');
    } else {
      // Docker not installed, show installation guidance
      openInstallationLink('docker');
    }
  };

  const handleStartRuntime = async (runtimeName) => {
    try {
      setIsStarting(true);
      setError('');
      setStartupProgress(`Starting ${runtimeName}...`);

      const result = await AttemptStartRuntime(runtimeName);
      
      if (result.success) {
        setStartupProgress(result.message);
        
        // Wait for the runtime to become fully ready
        setStartupProgress(`Waiting for ${runtimeName} to become ready...`);
        const waitResult = await WaitForRuntimeReady(runtimeName, 60);
        
        if (waitResult.success) {
          setStartupProgress('Runtime is ready! Initializing insta-infra...');
          
          // Try to reinitialize the app
          await ReinitializeRuntime();
          
          setStartupProgress('Success! insta-infra is ready to use.');
          
          // Notify parent component that runtime is ready
          if (onRuntimeReady) {
            setTimeout(() => onRuntimeReady(), 1000);
          }
        } else {
          setError(waitResult.error || 'Runtime did not become ready in time');
          setStartupProgress('');
        }
      } else {
        if (result.requiresManualAction) {
          setError(`Manual action required: ${result.message || result.error}`);
        } else {
          setError(result.error || 'Failed to start runtime');
        }
        setStartupProgress('');
      }
    } catch (err) {
      console.error(`Failed to start ${runtimeName}:`, err);
      setError(`Failed to start ${runtimeName}: ${err.message || err}`);
      setStartupProgress('');
    } finally {
      setIsStarting(false);
    }
  };

  const openInstallationLink = (runtime) => {
    const links = {
      docker: {
        darwin: 'https://docs.docker.com/desktop/install/mac-install/',
        linux: 'https://docs.docker.com/engine/install/',
        windows: 'https://docs.docker.com/desktop/install/windows-install/',
        default: 'https://docs.docker.com/get-docker/'
      },
      podman: {
        darwin: 'https://podman.io/getting-started/installation',
        linux: 'https://podman.io/getting-started/installation',
        windows: 'https://github.com/containers/podman/blob/main/docs/tutorials/podman-for-windows.md',
        default: 'https://podman.io/getting-started/installation'
      }
    };

    const platform = runtimeStatus?.platform || 'default';
    const url = links[runtime]?.[platform] || links[runtime]?.default;
    
    if (url) {
      window.open(url, '_blank');
    }
  };

  const handleSetCustomDockerPath = async () => {
    try {
      setPathError('');
      await SetCustomDockerPath(customDockerPath);
      // Refresh runtime status after setting custom path
      await checkRuntimeStatus();
    } catch (err) {
      setPathError(`Failed to set Docker path: ${err.message || err}`);
    }
  };

  const handleSetCustomPodmanPath = async () => {
    try {
      setPathError('');
      await SetCustomPodmanPath(customPodmanPath);
      // Refresh runtime status after setting custom path
      await checkRuntimeStatus();
    } catch (err) {
      setPathError(`Failed to set Podman path: ${err.message || err}`);
    }
  };

  const handleClearCustomDockerPath = async () => {
    try {
      setPathError('');
      setCustomDockerPath('');
      await SetCustomDockerPath('');
      // Refresh runtime status after clearing custom path
      await checkRuntimeStatus();
    } catch (err) {
      setPathError(`Failed to clear Docker path: ${err.message || err}`);
    }
  };

  const handleClearCustomPodmanPath = async () => {
    try {
      setPathError('');
      setCustomPodmanPath('');
      await SetCustomPodmanPath('');
      // Refresh runtime status after clearing custom path
      await checkRuntimeStatus();
    } catch (err) {
      setPathError(`Failed to clear Podman path: ${err.message || err}`);
    }
  };

  const getQuickStartAction = () => {
    const dockerRuntime = runtimeStatus?.runtimeStatuses?.find(rt => rt.name === 'docker');
    
    if (!dockerRuntime) {
      return { text: 'Install Docker', action: 'install', icon: '📥' };
    } else if (dockerRuntime.isAvailable) {
      return { text: 'Docker Ready', action: 'ready', icon: '✅' };
    } else if (dockerRuntime.isInstalled && dockerRuntime.canAutoStart) {
      return { text: 'Start Docker', action: 'start', icon: '▶️' };
    } else if (dockerRuntime.isInstalled) {
      return { text: 'Start Docker Manually', action: 'manual', icon: '⚠️' };
    } else {
      return { text: 'Install Docker', action: 'install', icon: '📥' };
    }
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="app-container">
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
          </div>
        </header>

        <main className="container main-content">
          <div className="loading-container">
            <div className="loading-content fade-in">
              <div className="loading-spinner-container">
                <svg width="48" height="48" className="loading-spinner" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
              </div>
              <p className="loading-text">Checking container runtime...</p>
              <p className="loading-subtext">This may take a moment</p>
            </div>
          </div>
        </main>
      </div>
    );
  }

  // Error state
  if (!runtimeStatus) {
    return (
      <div className="app-container">
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
          </div>
        </header>

        <main className="container main-content">
          <div className="error-alert fade-in">
            <div className="error-content">
              <svg width="24" height="24" className="error-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M12 9V13M12 17H12.01M4.98207 19H19.0179C20.5615 19 21.5233 17.3256 20.7455 15.9923L13.7276 3.96153C12.9558 2.63852 11.0442 2.63852 10.2724 3.96153L3.25452 15.9923C2.47675 17.3256 3.43849 19 4.98207 19Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
              <div>
                <p className="error-title">Runtime Check Failed</p>
                <p className="error-message">Failed to determine container runtime availability.</p>
              </div>
            </div>
          </div>
          
          <div className="empty-state-container">
            <div className="empty-state-content">
              <button onClick={checkRuntimeStatus} className="button button-primary">
                <svg width="16" height="16" className="button-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M4 4V9H4.58152M19.9381 11C19.446 7.05369 16.0796 4 12 4C8.64262 4 5.76829 6.06817 4.58152 9M4.58152 9H9M20 20V15H19.4185M19.4185 15C18.2317 17.9318 15.3574 20 12 20C7.92038 20 4.55399 16.9463 4.06189 13M19.4185 15H15" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
                Retry Check
              </button>
            </div>
          </div>
        </main>
      </div>
    );
  }

  // Success state - runtime is ready
  if (runtimeStatus.canProceed) {
    return (
      <div className="app-container">
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
          </div>
        </header>

        <main className="container main-content">
          <div className="empty-state-container">
            <div className="empty-state-content">
              <svg width="56" height="56" className="empty-state-icon text-green" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M9 12L11 14L15 10M21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3C16.9706 3 21 7.02944 21 12Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
              <p className="empty-state-title">Container Runtime Ready</p>
              <p className="empty-state-description">
                Using {runtimeStatus.preferredRuntime} - insta-infra is ready to manage your services!
              </p>
              {onRuntimeReady && (
                <button onClick={onRuntimeReady} className="button button-primary">
                  <svg width="16" height="16" className="button-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M9 18L15 12L9 6" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                  Continue to Services
                </button>
              )}
            </div>
          </div>
        </main>
      </div>
    );
  }

  const quickStartAction = getQuickStartAction();

  // Main setup state
  return (
    <div className="app-container">
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
        </div>
      </header>

      <main className="container main-content">
        {/* Error display */}
        {error && (
          <div className="error-alert fade-in">
            <div className="error-content">
              <svg width="24" height="24" className="error-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M12 9V13M12 17H12.01M4.98207 19H19.0179C20.5615 19 21.5233 17.3256 20.7455 15.9923L13.7276 3.96153C12.9558 2.63852 11.0442 2.63852 10.2724 3.96153L3.25452 15.9923C2.47675 17.3256 3.43849 19 4.98207 19Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
              <div>
                <p className="error-title">Setup Error</p>
                <p className="error-message">{error}</p>
              </div>
              <button onClick={() => setError('')} className="error-close">
                <svg width="16" height="16" className="close-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M6 18L18 6M6 6L18 18" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </button>
            </div>
          </div>
        )}

        {/* Progress display */}
        {startupProgress && (
          <div className="card fade-in">
            <div className="card-body">
              <div className="loading-content">
                <div className="loading-spinner-container">
                  <svg width="32" height="32" className="loading-spinner" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                </div>
                <p className="loading-text">{startupProgress}</p>
              </div>
            </div>
          </div>
        )}

        {/* Main setup content */}
        <div className="card fade-in">
          <div className="card-header">
            <div className="card-title">
              <div className="card-icon" style={{ background: 'linear-gradient(to right, #3b82f6, #4f46e5)' }}>
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M21 9.5V5C21 3.89543 20.1046 3 19 3H5C3.89543 3 3 3.89543 3 5V9.5M21 9.5H3M21 9.5V19C21 20.1046 20.1046 21 19 21H5C3.89543 21 3 20.1046 3 19V9.5M9 14H15" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </div>
              Container Runtime Setup
            </div>
          </div>
          
          <div className="card-body">
            <p style={{ marginBottom: '24px', color: '#94a3b8' }}>
              insta-infra requires Docker or Podman to manage your data infrastructure services.
            </p>

            {/* Quick Start Section */}
            <div style={{ marginBottom: '32px' }}>
              <h3 style={{ fontSize: '1.125rem', fontWeight: '600', marginBottom: '16px', color: 'var(--color-text)' }}>
                Quick Start
              </h3>
              
              <div style={{ 
                background: 'rgba(59, 130, 246, 0.1)', 
                border: '1px solid rgba(59, 130, 246, 0.3)', 
                borderRadius: '8px', 
                padding: '20px',
                marginBottom: '16px'
              }}>
                <p style={{ marginBottom: '16px', color: '#93c5fd', fontSize: '0.875rem' }}>
                  We recommend Docker for the best experience. Click below to get started quickly.
                </p>
                
                <button
                  onClick={handleQuickStart}
                  disabled={isStarting || quickStartAction.action === 'ready'}
                  className={`button ${quickStartAction.action === 'ready' ? 'button-secondary' : 'button-primary'}`}
                  style={{ width: '100%', justifyContent: 'center' }}
                >
                  {isStarting ? (
                    <>
                      <svg width="16" height="16" className="button-icon spin" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      Starting...
                    </>
                  ) : (
                    <>
                      <span style={{ fontSize: '16px', marginRight: '8px' }}>{quickStartAction.icon}</span>
                      {quickStartAction.text}
                    </>
                  )}
                </button>
              </div>
            </div>

            {/* Advanced Options */}
            <div>
              <button
                onClick={() => setShowAdvanced(!showAdvanced)}
                className="button button-secondary"
                style={{ marginBottom: '16px' }}
              >
                <svg 
                  width="16" 
                  height="16" 
                  className={`button-icon ${showAdvanced ? 'rotate-180' : ''}`} 
                  viewBox="0 0 24 24" 
                  fill="none" 
                  xmlns="http://www.w3.org/2000/svg"
                  style={{ transition: 'transform 0.2s' }}
                >
                  <path d="M6 9L12 15L18 9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
                Advanced Options
              </button>

              {showAdvanced && (
                <div className="fade-in">
                  <h4 style={{ fontSize: '1rem', fontWeight: '600', marginBottom: '16px', color: 'var(--color-text)' }}>
                    Choose Your Container Runtime
                  </h4>
                  
                  <div style={{ display: 'grid', gap: '16px' }}>
                    {runtimeStatus.runtimeStatuses.map((runtime) => (
                      <div 
                        key={runtime.name} 
                        style={{
                          background: 'rgba(30, 41, 59, 0.8)',
                          border: '1px solid var(--color-border)',
                          borderRadius: '8px',
                          padding: '16px'
                        }}
                      >
                        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '12px' }}>
                          <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                            <span style={{ fontSize: '24px' }}>
                              {runtime.name === 'docker' ? '🐳' : '🦭'}
                            </span>
                            <div>
                              <h5 style={{ margin: '0', fontSize: '1rem', fontWeight: '600', color: 'var(--color-text)' }}>
                                {runtime.name.charAt(0).toUpperCase() + runtime.name.slice(1)}
                              </h5>
                              <p style={{ margin: '0', fontSize: '0.75rem', color: '#94a3b8' }}>
                                {runtime.isAvailable ? 'Ready to use' : 
                                 runtime.isInstalled ? 'Installed but not running' : 
                                 'Not installed'}
                              </p>
                            </div>
                          </div>
                          
                          <div style={{ display: 'flex', gap: '8px' }}>
                            {!runtime.isInstalled && (
                              <button
                                onClick={() => openInstallationLink(runtime.name)}
                                className="button button-primary"
                                style={{ fontSize: '0.75rem', padding: '4px 8px' }}
                              >
                                Install
                              </button>
                            )}
                            
                            {runtime.isInstalled && !runtime.isRunning && runtime.canAutoStart && (
                              <button
                                onClick={() => handleStartRuntime(runtime.name)}
                                disabled={isStarting}
                                className="button button-primary"
                                style={{ fontSize: '0.75rem', padding: '4px 8px' }}
                              >
                                Start
                              </button>
                            )}
                          </div>
                        </div>
                        
                        {runtime.error && (
                          <p style={{ 
                            margin: '0', 
                            fontSize: '0.75rem', 
                            color: '#fca5a5',
                            background: 'rgba(239, 68, 68, 0.1)',
                            padding: '8px',
                            borderRadius: '4px'
                          }}>
                            {runtime.error}
                          </p>
                        )}
                        
                        {runtime.isInstalled && !runtime.isRunning && !runtime.canAutoStart && (
                          <p style={{ 
                            margin: '0', 
                            fontSize: '0.75rem', 
                            color: '#fbbf24',
                            background: 'rgba(245, 158, 11, 0.1)',
                            padding: '8px',
                            borderRadius: '4px'
                          }}>
                            Manual startup required: {runtime.startupCommand}
                          </p>
                        )}
                      </div>
                    ))}
                  </div>
                  
                  {/* Custom Path Configuration */}
                  <div style={{ marginTop: '24px', padding: '16px', background: 'rgba(55, 65, 81, 0.5)', borderRadius: '8px' }}>
                    <h5 style={{ margin: '0 0 16px 0', fontSize: '0.875rem', fontWeight: '600', color: 'var(--color-text)' }}>
                      Custom Binary Paths
                    </h5>
                    <p style={{ margin: '0 0 16px 0', fontSize: '0.75rem', color: '#94a3b8' }}>
                      If Docker or Podman is installed in a non-standard location, specify the full path to the binary here.
                    </p>
                    
                    {pathError && (
                      <div style={{ 
                        marginBottom: '16px', 
                        padding: '8px 12px', 
                        background: 'rgba(239, 68, 68, 0.1)', 
                        border: '1px solid rgba(239, 68, 68, 0.3)', 
                        borderRadius: '4px',
                        color: '#fca5a5',
                        fontSize: '0.75rem'
                      }}>
                        {pathError}
                      </div>
                    )}

                    {/* Docker Path Configuration */}
                    <div style={{ marginBottom: '16px' }}>
                      <label style={{ display: 'block', marginBottom: '4px', fontSize: '0.75rem', fontWeight: '500', color: 'var(--color-text)' }}>
                        Docker Binary Path
                      </label>
                      <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                        <input
                          type="text"
                          value={customDockerPath}
                          onChange={(e) => setCustomDockerPath(e.target.value)}
                          placeholder="/opt/homebrew/bin/docker"
                          style={{
                            flex: 1,
                            padding: '6px 8px',
                            fontSize: '0.75rem',
                            background: 'rgba(30, 41, 59, 0.8)',
                            border: '1px solid var(--color-border)',
                            borderRadius: '4px',
                            color: 'var(--color-text)'
                          }}
                        />
                        <button
                          onClick={handleSetCustomDockerPath}
                          disabled={!customDockerPath.trim()}
                          className="button button-primary"
                          style={{ fontSize: '0.75rem', padding: '6px 12px' }}
                        >
                          Set
                        </button>
                        {customDockerPath && (
                          <button
                            onClick={handleClearCustomDockerPath}
                            className="button button-secondary"
                            style={{ fontSize: '0.75rem', padding: '6px 12px' }}
                          >
                            Clear
                          </button>
                        )}
                      </div>
                    </div>

                    {/* Podman Path Configuration */}
                    <div style={{ marginBottom: '16px' }}>
                      <label style={{ display: 'block', marginBottom: '4px', fontSize: '0.75rem', fontWeight: '500', color: 'var(--color-text)' }}>
                        Podman Binary Path
                      </label>
                      <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                        <input
                          type="text"
                          value={customPodmanPath}
                          onChange={(e) => setCustomPodmanPath(e.target.value)}
                          placeholder="/opt/homebrew/bin/podman"
                          style={{
                            flex: 1,
                            padding: '6px 8px',
                            fontSize: '0.75rem',
                            background: 'rgba(30, 41, 59, 0.8)',
                            border: '1px solid var(--color-border)',
                            borderRadius: '4px',
                            color: 'var(--color-text)'
                          }}
                        />
                        <button
                          onClick={handleSetCustomPodmanPath}
                          disabled={!customPodmanPath.trim()}
                          className="button button-primary"
                          style={{ fontSize: '0.75rem', padding: '6px 12px' }}
                        >
                          Set
                        </button>
                        {customPodmanPath && (
                          <button
                            onClick={handleClearCustomPodmanPath}
                            className="button button-secondary"
                            style={{ fontSize: '0.75rem', padding: '6px 12px' }}
                          >
                            Clear
                          </button>
                        )}
                      </div>
                    </div>
                  </div>

                  {/* Application Logs */}
                  <div style={{ marginTop: '16px', padding: '16px', background: 'rgba(55, 65, 81, 0.5)', borderRadius: '8px' }}>
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '12px' }}>
                      <h5 style={{ margin: '0', fontSize: '0.875rem', fontWeight: '600', color: 'var(--color-text)' }}>
                        Application Logs
                      </h5>
                      <div style={{ display: 'flex', gap: '8px' }}>
                        <button
                          onClick={loadAppLogs}
                          className="button button-secondary"
                          style={{ fontSize: '0.75rem', padding: '4px 8px' }}
                        >
                          Refresh
                        </button>
                        <button
                          onClick={() => setShowLogs(!showLogs)}
                          className="button button-secondary"
                          style={{ fontSize: '0.75rem', padding: '4px 8px' }}
                        >
                          {showLogs ? 'Hide' : 'Show'}
                        </button>
                      </div>
                    </div>
                    
                    {showLogs && (
                      <div style={{
                        background: 'rgba(0, 0, 0, 0.3)',
                        border: '1px solid var(--color-border)',
                        borderRadius: '4px',
                        padding: '12px',
                        maxHeight: '200px',
                        overflowY: 'auto',
                        fontFamily: 'monospace',
                        fontSize: '0.75rem',
                        color: '#e2e8f0'
                      }}>
                        {appLogs.length === 0 ? (
                          <p style={{ margin: '0', color: '#94a3b8', fontStyle: 'italic' }}>
                            No logs available. Click "Refresh" to load logs.
                          </p>
                        ) : (
                          appLogs.map((log, index) => (
                            <div key={index} style={{ marginBottom: '2px', wordBreak: 'break-word' }}>
                              {log}
                            </div>
                          ))
                        )}
                      </div>
                    )}
                  </div>

                  <div style={{ marginTop: '16px', padding: '16px', background: 'rgba(55, 65, 81, 0.5)', borderRadius: '8px' }}>
                    <h5 style={{ margin: '0 0 8px 0', fontSize: '0.875rem', fontWeight: '600', color: 'var(--color-text)' }}>
                      Need Help?
                    </h5>
                    <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap' }}>
                      <a 
                        href="https://docs.docker.com/get-docker/" 
                        target="_blank" 
                        rel="noopener noreferrer"
                        style={{ color: 'var(--color-primary)', textDecoration: 'none', fontSize: '0.75rem' }}
                      >
                        Docker Installation Guide
                      </a>
                      <a 
                        href="https://podman.io/getting-started/installation" 
                        target="_blank" 
                        rel="noopener noreferrer"
                        style={{ color: 'var(--color-primary)', textDecoration: 'none', fontSize: '0.75rem' }}
                      >
                        Podman Installation Guide
                      </a>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Refresh button */}
        <div style={{ textAlign: 'center', marginTop: '24px' }}>
          <button onClick={checkRuntimeStatus} className="button button-secondary">
            <svg width="16" height="16" className="button-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M4 4V9H4.58152M19.9381 11C19.446 7.05369 16.0796 4 12 4C8.64262 4 5.76829 6.06817 4.58152 9M4.58152 9H9M20 20V15H19.4185M19.4185 15C18.2317 17.9318 15.3574 20 12 20C7.92038 20 4.55399 16.9463 4.06189 13M19.4185 15H15" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            Refresh Status
          </button>
        </div>
      </main>
    </div>
  );
};

export default RuntimeSetup; 