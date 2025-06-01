import React, { useState } from 'react';
import ServiceItem from './ServiceItem';

function ServiceList({ 
  services = [], 
  statuses = {}, 
  dependencyStatuses = {},
  isLoading = false, 
  onServiceStateChange
}) {
  const [searchTerm, setSearchTerm] = useState('');
  const [filterType, setFilterType] = useState('all');
  const [viewMode, setViewMode] = useState('grid'); // 'grid' or 'list'
  
  // Convert services with status data to the expected format
  const servicesWithStatus = services.map(service => ({
    ...service,
    name: service.Name,  // ServiceList expects lowercase name
    type: service.Type,  // ServiceList expects lowercase type  
    status: statuses[service.Name] || 'stopped', // Default to stopped if no status
    dependencies: service.Dependencies || [], // Include dependencies from backend
  }));
  
  // Get unique service types for filter dropdown
  const serviceTypes = ['all', ...new Set(servicesWithStatus.map(service => service.type))].sort();
  
  // Filter services based on search term and type filter
  const filteredServices = servicesWithStatus.filter(service => {
    const matchesSearch = service.name.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesType = filterType === 'all' || service.type === filterType;
    return matchesSearch && matchesType;
  });

  // Calculate status counts
  const runningCount = filteredServices.filter(s => s.status === 'running').length;
  const stoppedCount = filteredServices.filter(s => s.status === 'stopped').length;

  if (!isLoading && services.length === 0) {
    return (
      <div className="card">
        <div className="card-header">
          <div className="card-title">
            <div className="card-icon" style={{ backgroundColor: 'rgba(79, 70, 229, 0.2)' }}>
              <svg width="20" height="20" className="icon-indigo" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M21 10V8C21 6.89543 20.1046 6 19 6H5C3.89543 6 3 6.89543 3 8V10M21 10V16C21 17.1046 20.1046 18 19 18H5C3.89543 18 3 17.1046 3 16V10M21 10H3M7 14H12" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
              </svg>
            </div>
            <h2>Services</h2>
          </div>
        </div>
        
        <div className="card-body no-results">
          <svg width="48" height="48" className="no-results-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M12 8V12M12 16H12.01M22 12C22 17.5228 17.5228 22 12 22C6.47715 22 2 17.5228 2 12C2 6.47715 6.47715 2 12 2C17.5228 2 22 6.47715 22 12Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
          <p className="no-results-text">No services available.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="card">
      <div className="card-header">
        <div className="card-title">
          <div className="card-icon" style={{ backgroundColor: 'rgba(79, 70, 229, 0.2)' }}>
            <svg width="20" height="20" className="icon-indigo" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M21 10V8C21 6.89543 20.1046 6 19 6H5C3.89543 6 3 6.89543 3 8V10M21 10V16C21 17.1046 20.1046 18 19 18H5C3.89543 18 3 17.1046 3 16V10M21 10H3M7 14H12" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
            </svg>
          </div>
          <h2>Services</h2>
        </div>
        
        {/* Status Counts */}
        <div className="status-counters">
          <div className="status-counter status-counter-running">
            <span className="status-dot dot-green"></span>
            {runningCount} Running
          </div>
          <div className="status-counter status-counter-stopped">
            <span className="status-dot dot-gray"></span>
            {stoppedCount} Stopped
          </div>
        </div>
        
        {/* View toggle buttons */}
        <div className="view-toggle">
          <button 
            onClick={() => setViewMode('grid')}
            className={`toggle-button ${viewMode === 'grid' ? 'active' : ''}`}
            title="Grid view"
          >
            <svg width="16" height="16" className="toggle-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M4 6C4 4.89543 4.89543 4 6 4H8C9.10457 4 10 4.89543 10 6V8C10 9.10457 9.10457 10 8 10H6C4.89543 10 4 9.10457 4 8V6Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
              <path d="M14 6C14 4.89543 14.8954 4 16 4H18C19.1046 4 20 4.89543 20 6V8C20 9.10457 19.1046 10 18 10H16C14.8954 10 14 9.10457 14 8V6Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
              <path d="M4 16C4 14.8954 4.89543 14 6 14H8C9.10457 14 10 14.8954 10 16V18C10 19.1046 9.10457 20 8 20H6C4.89543 20 4 19.1046 4 18V16Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
              <path d="M14 16C14 14.8954 14.8954 14 16 14H18C19.1046 14 20 14.8954 20 16V18C20 19.1046 19.1046 20 18 20H16C14.8954 20 14 19.1046 14 18V16Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
            </svg>
          </button>
          <button 
            onClick={() => setViewMode('list')}
            className={`toggle-button ${viewMode === 'list' ? 'active' : ''}`}
            title="List view"
          >
            <svg width="16" height="16" className="toggle-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M8 6H21M8 12H21M8 18H21M3 6H3.01M3 12H3.01M3 18H3.01" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </button>
        </div>
      </div>
      
      {/* Search and Filter Bar */}
      <div className="filter-section">
        <div className="search-row">
          {/* Search input */}
          <div className="search-container">
            <div className="search-icon">
              <svg width="20" height="20" className="search-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M21 21L15 15M17 10C17 13.866 13.866 17 10 17C6.13401 17 3 13.866 3 10C3 6.13401 6.13401 3 10 3C13.866 3 17 6.13401 17 10Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            </div>
            <input
              type="text"
              placeholder="Search by name..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="search-input"
            />
            {searchTerm && (
              <button 
                onClick={() => setSearchTerm('')}
                className="clear-button"
              >
                <svg width="20" height="20" className="clear-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M6 18L18 6M6 6L18 18" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </button>
            )}
          </div>
          
          {/* Type filter */}
          <div className="filter-container">
            <div className="filter-icon">
              <svg width="20" height="20" className="filter-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M3 4H21M3 12H21M3 20H21M7 8V16M15 16V8" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            </div>
            <select
              value={filterType}
              onChange={(e) => setFilterType(e.target.value)}
              className="filter-select"
            >
              {serviceTypes.map(type => (
                <option key={type} value={type}>
                  {type === 'all' ? 'All Types' : type}
                </option>
              ))}
            </select>
            <div className="select-arrow">
              <svg width="20" height="20" className="arrow-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M19 9L12 16L5 9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            </div>
          </div>
        </div>
        
        {/* Results Count and Clear Filters Button */}
        <div className="filter-results">
          <div className="results-count">
            Showing <span>{filteredServices.length}</span> of <span className="total-count">{services.length}</span> services
          </div>
          
          {(searchTerm || filterType !== 'all') && (
            <button 
              onClick={() => { setSearchTerm(''); setFilterType('all'); }}
              className="clear-filters"
            >
              <svg width="16" height="16" className="clear-filters-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M6 18L18 6M6 6L18 18" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
              Clear filters
            </button>
          )}
        </div>
      </div>
      
      <div className="card-body">
        {/* Services Grid or List */}
        {filteredServices.length > 0 ? (
          viewMode === 'grid' ? (
            <div className="services-grid">
              {filteredServices.map((service, index) => (
                <div 
                  key={service.name} 
                  className="fade-in" 
                  style={{ animationDelay: `${Math.min(index * 20, 500)}ms` }}
                >
                  <ServiceItem service={service} onServiceStateChange={onServiceStateChange} statuses={statuses} dependencyStatuses={dependencyStatuses} />
                </div>
              ))}
            </div>
          ) : (
            <div className="services-list">
              {filteredServices.map((service, index) => (
                <div 
                  key={service.name}
                  className="fade-in" 
                  style={{ animationDelay: `${Math.min(index * 15, 300)}ms` }}
                >
                  <ServiceItem service={service} onServiceStateChange={onServiceStateChange} statuses={statuses} dependencyStatuses={dependencyStatuses} />
                </div>
              ))}
            </div>
          )
        ) : (
          <div className="no-results">
            <svg width="48" height="48" className="no-results-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M9.172 14.828L12.001 12M14.829 9.172L12.001 12M12.001 12L9.172 9.172M12.001 12L14.829 14.828M12 22C17.523 22 22 17.523 22 12C22 6.477 17.523 2 12 2C6.477 2 2 6.477 2 12C2 17.523 6.477 22 12 22Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            <p className="no-results-text">No services match your search criteria</p>
            <button 
              onClick={() => { setSearchTerm(''); setFilterType('all'); }}
              className="button button-primary"
            >
              <svg width="16" height="16" className="button-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M6 18L18 6M6 6L18 18" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
              Clear filters
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

export default ServiceList; 