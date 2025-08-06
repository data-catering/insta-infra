import React, { useState, useEffect } from 'react';
import { useApiClient } from '../contexts/ApiContext';
import CustomServiceModal from './CustomServiceModal';

function CustomServicesList({ 
  onServiceAdded,
  onServiceUpdated,
  onServiceDeleted,
  isVisible = true 
}) {
  const { listCustomServices, deleteCustomService, getCustomService } = useApiClient();
  const [customServices, setCustomServices] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [showModal, setShowModal] = useState(false);
  const [editingService, setEditingService] = useState(null);
  const [isEditing, setIsEditing] = useState(false);
  const [expandedService, setExpandedService] = useState(null);

  // Load custom services
  const loadCustomServices = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const services = await listCustomServices();
      setCustomServices(services || []);
    } catch (err) {
      setError(err.message || 'Failed to load custom services');
      setCustomServices([]);
    } finally {
      setIsLoading(false);
    }
  };

  // Load services on component mount
  useEffect(() => {
    if (isVisible) {
      loadCustomServices();
    }
  }, [isVisible]);

  const handleAddService = () => {
    setEditingService(null);
    setIsEditing(false);
    setShowModal(true);
  };

  const handleEditService = async (service) => {
    try {
      // Get the full service details including content
      const fullService = await getCustomService(service.id);
      setEditingService({
        ...service,
        content: fullService.content
      });
      setIsEditing(true);
      setShowModal(true);
    } catch (err) {
      setError(`Failed to load service details: ${err.message}`);
    }
  };

  const handleDeleteService = async (service) => {
    if (!window.confirm(`Are you sure you want to delete "${service.name}"? This action cannot be undone.`)) {
      return;
    }

    try {
      await deleteCustomService(service.id);
      await loadCustomServices(); // Refresh the list
      onServiceDeleted && onServiceDeleted(service);
    } catch (err) {
      setError(`Failed to delete service: ${err.message}`);
    }
  };

  const handleServiceAdded = async (newService) => {
    await loadCustomServices(); // Refresh the list
    onServiceAdded && onServiceAdded(newService);
  };

  const handleServiceUpdated = async (updatedService) => {
    await loadCustomServices(); // Refresh the list
    onServiceUpdated && onServiceUpdated(updatedService);
  };

  const toggleServiceDetails = (serviceId) => {
    setExpandedService(expandedService === serviceId ? null : serviceId);
  };

  if (!isVisible) return null;

  return (
    <>
      <div className="card custom-services-card">
        <div className="card-header">
          <div className="card-title">
            <div className="card-icon" style={{ backgroundColor: 'rgba(16, 185, 129, 0.2)' }}>
              <svg width="20" height="20" className="icon-green" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M14 2H6C4.89543 2 4 2.89543 4 4V20C4 21.1046 4.89543 22 6 22H18C19.1046 22 20 21.1046 20 20V8L14 2Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                <path d="M14 2V8H20" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            </div>
            <h2>Custom Services</h2>
            {isLoading && (
              <div className="loading-indicator" title="Loading custom services...">
                <div className="status-icon spin"></div>
              </div>
            )}
          </div>
          
          <button onClick={handleAddService} className="button button-primary">
            <svg width="16" height="16" className="button-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M12 4.5V19.5M19.5 12H4.5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            Add Custom
          </button>
        </div>

        <div className="card-body">
          {error && (
            <div className="error-banner">
              <svg width="16" height="16" className="error-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M12 8V12M12 16H12.01M22 12C22 17.5228 17.5228 22 12 22C6.47715 22 2 17.5228 2 12C2 6.47715 6.47715 2 12 2C17.5228 2 22 6.47715 22 12Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
              {error}
              <button 
                onClick={() => setError(null)}
                className="error-dismiss"
                aria-label="Dismiss error"
              >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M6 18L18 6M6 6L18 18" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </button>
            </div>
          )}

          {isLoading && customServices.length === 0 ? (
            <div className="loading-container">
              <div className="loading-content">
                <div className="loading-spinner-container">
                  <div className="loading-spinner"></div>
                </div>
                <p className="loading-text">Loading custom services...</p>
              </div>
            </div>
          ) : customServices.length === 0 ? (
            <div className="no-results">
              <svg width="48" height="48" className="no-results-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M14 2H6C4.89543 2 4 2.89543 4 4V20C4 21.1046 4.89543 22 6 22H18C19.1046 22 20 21.1046 20 20V8L14 2Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                <path d="M14 2V8H20" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
              <p className="no-results-text">No custom services configured</p>
              <p className="no-results-subtext">Add your own docker compose files to extend insta-infra</p>
              <button onClick={handleAddService} className="button button-primary">
                <svg width="16" height="16" className="button-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M12 4.5V19.5M19.5 12H4.5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
                Add Your First Custom Service
              </button>
            </div>
          ) : (
            <div className="custom-services-list">
              {customServices.map((service) => (
                <div key={service.id} className="custom-service-item">
                  <div className="custom-service-header" onClick={() => toggleServiceDetails(service.id)}>
                    <div className="custom-service-info">
                      <div className="custom-service-name">
                        <span className="service-name-text">{service.name}</span>
                        <span className="service-type-badge">Custom</span>
                        {service.metadata?.warnings && service.metadata.warnings.length > 0 && (
                          <span className="warning-badge" title={service.metadata.warnings.join(', ')}>
                            <svg width="14" height="14" className="warning-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                              <path d="M12 8V12M12 16H12.01M22 12C22 17.5228 17.5228 22 12 22C6.47715 22 2 17.5228 2 12C2 6.47715 6.47715 2 12 2C17.5228 2 22 6.47715 22 12Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                            </svg>
                            {service.metadata.warnings.length}
                          </span>
                        )}
                      </div>
                      {service.description && (
                        <div className="custom-service-description">{service.description}</div>
                      )}
                      <div className="custom-service-meta">
                        <span className="service-count">{service.services?.length || 0} services</span>
                        <span className="service-filename">{service.filename}</span>
                        <span className="service-date">Updated {new Date(service.updated_at).toLocaleDateString()}</span>
                      </div>
                    </div>
                    
                    <div className="custom-service-actions">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleEditService(service);
                        }}
                        className="action-button edit"
                        title="Edit service"
                      >
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                          <path d="M11 4H4C3.46957 4 2.96086 4.21071 2.58579 4.58579C2.21071 4.96086 2 5.46957 2 6V20C2 20.5304 2.21071 21.0391 2.58579 21.4142C2.96086 21.7893 3.46957 22 4 22H18C18.5304 22 19.0391 21.7893 19.4142 21.4142C19.7893 21.0391 20 20.5304 20 20V13M18.5 2.5C18.8978 2.10217 19.4374 1.87868 20 1.87868C20.5626 1.87868 21.1022 2.10217 21.5 2.5C21.8978 2.89783 22.1213 3.43739 22.1213 4C22.1213 4.56261 21.8978 5.10217 21.5 5.5L12 15L8 16L9 12L18.5 2.5Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                        </svg>
                      </button>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDeleteService(service);
                        }}
                        className="action-button delete"
                        title="Delete service"
                      >
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                          <path d="M3 6H5H21M8 6V4C8 3.46957 8.21071 2.96086 8.58579 2.58579C8.96086 2.21071 9.46957 2 10 2H14C14.5304 2 15.0391 2.21071 15.4142 2.58579C15.7893 2.96086 16 3.46957 16 4V6M19 6V20C19 20.5304 18.7893 21.0391 18.4142 21.4142C18.0391 21.7893 17.5304 22 17 22H7C6.46957 22 5.96086 21.7893 5.58579 21.4142C5.21071 21.0391 5 20.5304 5 20V6H19Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                        </svg>
                      </button>
                      
                      <button
                        className={`expand-button ${expandedService === service.id ? 'expanded' : ''}`}
                        title={expandedService === service.id ? 'Collapse details' : 'Expand details'}
                      >
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                          <path d="M19 9L12 16L5 9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                        </svg>
                      </button>
                    </div>
                  </div>

                  {expandedService === service.id && (
                    <div className="custom-service-details">
                      {service.metadata?.warnings && service.metadata.warnings.length > 0 && (
                        <div className="warnings-section">
                          <h4 className="warnings-title">
                            <svg width="16" height="16" className="warning-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                              <path d="M12 8V12M12 16H12.01M22 12C22 17.5228 17.5228 22 12 22C6.47715 22 2 17.5228 2 12C2 6.47715 6.47715 2 12 2C17.5228 2 22 6.47715 22 12Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                            </svg>
                            Warnings
                          </h4>
                          <ul className="warnings-list">
                            {service.metadata.warnings.map((warning, index) => (
                              <li key={index} className="warning-item">{warning}</li>
                            ))}
                          </ul>
                        </div>
                      )}
                      
                      <div className="services-section">
                        <h4 className="services-title">
                          <svg width="16" height="16" className="services-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                            <path d="M21 10V8C21 6.89543 20.1046 6 19 6H5C3.89543 6 3 6.89543 3 8V10M21 10V16C21 17.1046 20.1046 18 19 18H5C3.89543 18 3 17.1046 3 16V10M21 10H3M7 14H12" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                          </svg>
                          Services ({service.services?.length || 0})
                        </h4>
                        {service.services && service.services.length > 0 ? (
                          <div className="services-grid">
                            {service.services.map((serviceName) => (
                              <span key={serviceName} className="service-tag">
                                {serviceName}
                              </span>
                            ))}
                          </div>
                        ) : (
                          <p className="no-services">No services defined</p>
                        )}
                      </div>

                      <div className="file-info">
                        <div className="file-info-item">
                          <span className="file-info-label">File:</span>
                          <span className="file-info-value">{service.filename}</span>
                        </div>
                        <div className="file-info-item">
                          <span className="file-info-label">Created:</span>
                          <span className="file-info-value">{new Date(service.created_at).toLocaleString()}</span>
                        </div>
                        <div className="file-info-item">
                          <span className="file-info-label">Modified:</span>
                          <span className="file-info-value">{new Date(service.updated_at).toLocaleString()}</span>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      <CustomServiceModal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        onServiceAdded={handleServiceAdded}
        onServiceUpdated={handleServiceUpdated}
        editingService={editingService}
        isEditing={isEditing}
      />
    </>
  );
}

export default CustomServicesList;