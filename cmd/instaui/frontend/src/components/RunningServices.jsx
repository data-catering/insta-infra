import React from 'react';
import ServiceItem from './ServiceItem';

function RunningServices({ services = [], isLoading = false, onServiceStateChange, onShowDependencyGraph }) {
  if (!isLoading && services.length === 0) {
    return (
      <div className="card">
        <div className="card-header">
          <div className="card-title">
            <div className="card-icon" style={{ backgroundColor: '#10b981' }}>
              <svg width="16" height="16" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
              </svg>
            </div>
            Active Services
          </div>
          <div className="status-indicator">
            <span className="status-dot dot-gray"></span>
            <span>0 Running</span>
          </div>
        </div>
        <div className="card-body">
          <div className="no-results">
            <svg width="48" height="48" className="no-results-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M12 8V12L15 15M12 3C7.02944 3 3 7.02944 3 12C3C16.9706 7.02944 21 12 21C16.9706 21 21 16.9706 21 12C21 7.02944 16.9706 3 12 3Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            <h3 className="no-results-text">No Active Services</h3>
            <p style={{ color: '#9ca3af', fontSize: '14px' }}>Start a service to see it appear here</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="card">
      <div className="card-header">
        <div className="card-title">
          <div className="card-icon" style={{ backgroundColor: '#10b981' }}>
            <svg width="16" height="16" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
            </svg>
          </div>
          Active Services
        </div>
        <div className="status-indicator">
          <span className="status-dot dot-green pulse"></span>
          <span>{services.length} Running</span>
        </div>
      </div>
      
      <div className="card-body">
        <div className="services-grid">
          {services.map((service, index) => (
            <div 
              key={service.name} 
              className="fade-in" 
              style={{ animationDelay: `${index * 50}ms` }}
            >
              <ServiceItem 
                service={service} 
                onServiceStateChange={onServiceStateChange} 
                onShowDependencyGraph={onShowDependencyGraph} 
              />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

export default RunningServices; 