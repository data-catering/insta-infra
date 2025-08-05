import React from 'react';
import ServiceItem from './ServiceItem';

function RunningServices({ services = [], onServiceStateChange, statuses = {}, dependencyStatuses = {}, currentRuntime }) {
  // If no running services, don't render anything (component will be hidden by parent)
  if (services.length === 0) {
    return null;
  }

  return (
    <div className="card running-services-compact">
      <div className="card-header compact-header">
        <div className="card-title">
          <div className="card-icon" style={{ backgroundColor: '#10b981' }}>
            <svg width="16" height="16" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
            </svg>
          </div>
          <h2>Running Services</h2>
        </div>
        <div className="status-indicator">
          <span className="status-dot dot-green pulse"></span>
          <span>{services.length} Active</span>
        </div>
      </div>
      
      <div className="card-body compact-body">
        <div className="services-grid compact-grid">
          {services.map((service, index) => (
            <div 
              key={service.name} 
              className="fade-in" 
              style={{ animationDelay: `${Math.min(index * 15, 200)}ms` }}
            >
              <ServiceItem 
                service={service} 
                onServiceStateChange={onServiceStateChange} 
                statuses={statuses}
                dependencyStatuses={dependencyStatuses}
                serviceStatuses={statuses}
                currentRuntime={currentRuntime}
              />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

export default RunningServices; 