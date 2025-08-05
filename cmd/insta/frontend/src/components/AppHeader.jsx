import React from 'react';

const AppHeader = ({ 
  lastUpdated, 
  runningServicesCount, 
  isLoading, 
  hasRunningServices, 
  isStoppingAll, 
  handleStopAllServices,
  showActionsDropdown,
  setShowActionsDropdown,
  setShowAbout,
  handleRestartRuntime,
  handleShutdownApplication,
  currentRuntime
}) => {
  const formatTime = (date) => {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  return (
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
          {hasRunningServices && (
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
  );
};

export default AppHeader;