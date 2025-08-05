import React from 'react';
import ServiceList from './ServiceList';
import RunningServices from './RunningServices';
import CustomServicesList from './CustomServicesList';
import { ImageStatusProvider, ImageLoader } from './ServiceItem';

const MainContent = ({
  error,
  setError,
  isLoading,
  services,
  runningServices,
  statuses,
  dependencyStatuses,
  currentRuntime,
  fetchAllServices,
  isAnimating
}) => {
  return (
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
      {/* Custom Services Section */}
      <CustomServicesList 
        onServiceAdded={fetchAllServices}
        onServiceUpdated={fetchAllServices}
        onServiceDeleted={fetchAllServices}
        isVisible={!isLoading || services.length > 0}
      />
    </main>
  );
};

export default MainContent;