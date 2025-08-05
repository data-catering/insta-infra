import React, { useState, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { X, Download, Layers, Clock, Wifi } from 'lucide-react';
import { wsClient, WS_MSG_TYPES } from '../api/client';

const ProgressModal = ({ isOpen, onClose, serviceName, imageName = '' }) => {
  const [progress, setProgress] = useState({
    status: 'idle',
    progress: 0,
    currentLayer: '',
    totalLayers: 0,
    downloaded: 0,
    total: 0,
    speed: '',
    eta: '',
    error: '',
    serviceName: ''
  });
  const [startTime, setStartTime] = useState(null);
  const [elapsedTime, setElapsedTime] = useState(0);

  // Set up event listeners and start progress tracking
  useEffect(() => {
    if (isOpen && serviceName) {
      setStartTime(Date.now());
      setElapsedTime(0);
      setupEventListeners();
    }
    
    return () => {
      cleanup();
    };
  }, [isOpen, serviceName]);

  // Update elapsed time
  useEffect(() => {
    let interval;
    if (isOpen && startTime && (progress.status === 'downloading' || progress.status === 'extracting')) {
      interval = setInterval(() => {
        setElapsedTime(Date.now() - startTime);
      }, 1000);
    }
    return () => {
      if (interval) clearInterval(interval);
    };
  }, [isOpen, startTime, progress.status]);

  const setupEventListeners = () => {
    // Listen for image pull progress events via WebSocket
    const progressCallback = (data) => {
      if (data.serviceName === serviceName || data.service_name === serviceName) {
        setProgress({
          status: data.status || 'downloading',
          progress: data.progress || 0,
          currentLayer: data.currentLayer || data.current_layer || '',
          totalLayers: data.totalLayers || data.total_layers || 0,
          downloaded: data.downloaded || data.bytes_downloaded || 0,
          total: data.total || data.total_bytes || 0,
          speed: data.speed || '',
          eta: data.eta || '',
          error: data.error || '',
          serviceName: data.serviceName || data.service_name || serviceName
        });
        
        // Close modal automatically on completion
        if (data.status === 'complete') {
          setTimeout(() => {
            onClose();
          }, 2000); // Show completion for 2 seconds
        }
        
        // Stop monitoring and cleanup on error
        if (data.status === 'error') {
          // Stop WebSocket monitoring for this service
          cleanup();
          // Close modal automatically on error after a brief display
          setTimeout(() => {
            onClose();
          }, 3000); // Show error for 3 seconds
        }
      }
    };

    // Subscribe to WebSocket image pull progress events
    wsClient.subscribe(WS_MSG_TYPES.IMAGE_PULL_PROGRESS, progressCallback);
    
    // Store callback for cleanup
    setupEventListeners.progressCallback = progressCallback;
  };

  const cleanup = () => {
    // Unsubscribe from WebSocket events
    if (setupEventListeners.progressCallback) {
      wsClient.unsubscribe(WS_MSG_TYPES.IMAGE_PULL_PROGRESS, setupEventListeners.progressCallback);
      setupEventListeners.progressCallback = null;
    }
  };

  const handleCancel = () => {
    // You could call StopImagePull here if needed
    onClose();
  };

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatTime = (ms) => {
    const seconds = Math.floor(ms / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    
    if (hours > 0) {
      return `${hours}h ${minutes % 60}m ${seconds % 60}s`;
    } else if (minutes > 0) {
      return `${minutes}m ${seconds % 60}s`;
    } else {
      return `${seconds}s`;
    }
  };

  const getStatusColor = (status) => {
    switch (status) {
      case 'complete': return 'text-green-400';
      case 'error': return 'text-red-400';
      case 'downloading': return 'text-blue-400';
      case 'extracting': return 'text-yellow-400';
      case 'starting': return 'text-gray-400';
      default: return 'text-gray-400';
    }
  };

  const getStatusIcon = (status) => {
    switch (status) {
      case 'downloading': return <Download size={20} className="animate-bounce" />;
      case 'extracting': return <Layers size={20} className="animate-pulse" />;
      case 'complete': return <Download size={20} className="text-green-400" />;
      case 'error': return <X size={20} className="text-red-400" />;
      default: return <Download size={20} />;
    }
  };

  if (!isOpen) return null;

  return createPortal(
    <div className="connection-modal-overlay" onClick={(e) => e.target === e.currentTarget && handleCancel()}>
      <div className="connection-modal progress-modal" role="dialog" aria-labelledby="progress-modal-title" aria-describedby="progress-modal-content" onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div className="connection-modal-header">
          <div className="connection-modal-title" id="progress-modal-title">
            {getStatusIcon(progress.status)}
            <span>Downloading Image - {serviceName}</span>
          </div>
          <div className="modal-header-actions">
            <button className="connection-modal-close" onClick={handleCancel}>
              <X size={20} />
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="connection-modal-content progress-content" id="progress-modal-content">
          {/* Image Information */}
          <div className="progress-info">
            <div className="progress-info-item">
              <span className="progress-label">Service:</span>
              <span className="progress-value">{serviceName}</span>
            </div>
            {imageName && (
              <div className="progress-info-item">
                <span className="progress-label">Image:</span>
                <span className="progress-value progress-image-name">{imageName}</span>
              </div>
            )}
            <div className="progress-info-item">
              <span className="progress-label">Status:</span>
              <span className={`progress-value ${getStatusColor(progress.status)}`}>
                {progress.status.charAt(0).toUpperCase() + progress.status.slice(1)}
              </span>
            </div>
          </div>

          {/* Main Progress Bar */}
          <div className="progress-section">
            <div className="progress-header">
              <span className="progress-title">Overall Progress</span>
              <span className="progress-percentage">{progress.progress.toFixed(1)}%</span>
            </div>
            <div className="progress-bar-container">
              <div 
                className="progress-bar"
                role="progressbar"
                aria-valuemin="0"
                aria-valuemax="100"
                aria-valuenow={Math.round(progress.progress)}
                aria-label={`Download progress: ${progress.progress.toFixed(1)}%`}
                style={{ width: `${Math.max(0, Math.min(100, progress.progress))}%` }}
              />
            </div>
          </div>

          {/* Layer Information */}
          {(progress.totalLayers > 0 || progress.currentLayer) && (
            <div className="progress-section">
              <div className="progress-header">
                <span className="progress-title">
                  <Layers size={16} className="inline mr-2" />
                  Layer Progress
                </span>
                {progress.totalLayers > 0 && (
                  <span className="progress-meta">{progress.totalLayers} layers</span>
                )}
              </div>
              {progress.currentLayer && (
                <div className="progress-layer">
                  <span className="layer-label">Current Layer:</span>
                  <span className="layer-id">{progress.currentLayer}</span>
                </div>
              )}
            </div>
          )}

          {/* Download Statistics */}
          <div className="progress-stats">
            <div className="progress-stat">
              <Clock size={16} />
              <div className="stat-content">
                <span className="stat-label">Elapsed</span>
                <span className="stat-value">{formatTime(elapsedTime)}</span>
              </div>
            </div>
            {progress.speed && (
              <div className="progress-stat">
                <Wifi size={16} />
                <div className="stat-content">
                  <span className="stat-label">Speed</span>
                  <span className="stat-value">{progress.speed}</span>
                </div>
              </div>
            )}
            {progress.eta && (
              <div className="progress-stat">
                <Clock size={16} />
                <div className="stat-content">
                  <span className="stat-label">ETA</span>
                  <span className="stat-value">{progress.eta}</span>
                </div>
              </div>
            )}
          </div>

          {/* Data Transfer Info */}
          {(progress.downloaded > 0 || progress.total > 0) && (
            <div className="progress-transfer">
              <span className="transfer-label">Downloaded:</span>
              <span className="transfer-value">
                {formatBytes(progress.downloaded)}
                {progress.total > 0 && ` / ${formatBytes(progress.total)}`}
              </span>
            </div>
          )}

          {/* Error Message */}
          {progress.error && (
            <div className="progress-error">
              <X size={16} className="error-icon" />
              <div className="error-content">
                <div className="error-title">Failed to download image</div>
                <div className="error-details">{progress.error}</div>
                <div className="error-help">
                  <strong>Troubleshooting:</strong>
                  <ul>
                    <li>Check your internet connection</li>
                    <li>Verify Docker/Podman is running properly</li>
                    <li>Try running manually: <code>docker pull {imageName || 'image-name'}</code></li>
                    <li>Check if you have sufficient disk space</li>
                    <li>Verify image name and registry access</li>
                    <li>Check for network firewall restrictions</li>
                  </ul>
                </div>
              </div>
            </div>
          )}

          {/* Success Message */}
          {progress.status === 'complete' && (
            <div className="progress-success">
              <Download size={16} className="success-icon" />
              <span>Image downloaded successfully! Starting service...</span>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="progress-footer">
          {progress.status === 'error' || progress.status === 'complete' ? (
            <button className="button button-primary" onClick={onClose}>
              Close
            </button>
          ) : (
            <button className="button button-secondary" onClick={handleCancel}>
              Cancel Download
            </button>
          )}
        </div>
      </div>
    </div>,
    document.body
  );
};

export default ProgressModal; 