import React from 'react';
import { createPortal } from 'react-dom';
import { X, Copy } from 'lucide-react';

const ConnectionModal = ({ 
  isOpen, 
  onClose, 
  connectionInfo, 
  copyFeedback, 
  onCopyToClipboard 
}) => {
  if (!isOpen || !connectionInfo) return null;

  return createPortal(
    <div className="connection-modal-overlay" onClick={onClose}>
      <div className="connection-modal" onClick={(e) => e.stopPropagation()}>
        <div className="connection-modal-header">
          <h3 className="connection-modal-title">
            Connection Details - {connectionInfo.serviceName}
          </h3>
          <div className="modal-header-actions">
            {copyFeedback && (
              <div className="copy-feedback">
                {copyFeedback}
              </div>
            )}
            <button 
              onClick={onClose}
              className="connection-modal-close"
            >
              <X size={16} />
            </button>
          </div>
        </div>
        
        <div className="connection-modal-content">
          {!connectionInfo.available ? (
            <div className="connection-error">
              <p>Service connection information is not available.</p>
              {connectionInfo.error && <p className="error-detail">{connectionInfo.error}</p>}
            </div>
          ) : (
            <div className="connection-details">
              {/* Web URL */}
              {connectionInfo.hasWebUI && connectionInfo.webURL && (
                <div className="connection-item">
                  <label className="connection-label">Web URL:</label>
                  <div className="connection-value-row">
                    <a 
                      href={connectionInfo.webURL} 
                      target="_blank" 
                      rel="noopener noreferrer"
                      className="connection-value connection-link"
                    >
                      {connectionInfo.webURL}
                    </a>
                    <button 
                      className="copy-button"
                      onClick={() => onCopyToClipboard(connectionInfo.webURL)}
                      title="Copy to clipboard"
                    >
                      <Copy size={12} />
                    </button>
                  </div>
                </div>
              )}

              {/* Credentials */}
              {(connectionInfo.username || connectionInfo.password) && (
                <div className="connection-credentials">
                  <h4 className="credentials-title">Credentials:</h4>
                  {connectionInfo.username && (
                    <div className="connection-item">
                      <label className="connection-label">Username:</label>
                      <div className="connection-value-row">
                        <code className="connection-value">{connectionInfo.username}</code>
                        <button 
                          className="copy-button"
                          onClick={() => onCopyToClipboard(connectionInfo.username)}
                          title="Copy to clipboard"
                        >
                          <Copy size={12} />
                        </button>
                      </div>
                    </div>
                  )}
                  {connectionInfo.password && (
                    <div className="connection-item">
                      <label className="connection-label">Password:</label>
                      <div className="connection-value-row">
                        <code className="connection-value">{'•'.repeat(connectionInfo.password.length)}</code>
                        <button 
                          className="copy-button"
                          onClick={() => onCopyToClipboard(connectionInfo.password)}
                          title="Copy password to clipboard"
                        >
                          <Copy size={12} />
                        </button>
                      </div>
                    </div>
                  )}
                </div>
              )}

              {/* Connection Command */}
              {connectionInfo.connectionCommand && connectionInfo.connectionCommand !== 'bash' && (
                <div className="connection-item">
                  <label className="connection-label">CLI Command:</label>
                  <div className="connection-value-row">
                    <code className="connection-value connection-command">{connectionInfo.connectionCommand}</code>
                    <button 
                      className="copy-button"
                      onClick={() => onCopyToClipboard(connectionInfo.connectionCommand)}
                      title="Copy to clipboard"
                    >
                      <Copy size={12} />
                    </button>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>,
    document.body
  );
};

export default ConnectionModal; 