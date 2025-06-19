import React, { useState } from 'react';
import { createPortal } from 'react-dom';
import { X, Copy, ExternalLink, Database, Server, Globe, Activity, Monitor, Eye, EyeOff } from 'lucide-react';

// CredentialItem component for handling password masking
const CredentialItem = ({ credential, isPassword, onCopyToClipboard }) => {
  const [showPassword, setShowPassword] = useState(false);

  const togglePasswordVisibility = () => {
    setShowPassword(!showPassword);
  };

  const displayValue = isPassword && !showPassword ? '••••••••' : credential.value;

  return (
    <div className="connection-item">
      <div className="connection-label">{credential.description || 'Credential'}</div>
      <div className="connection-value-row">
        <div className="connection-value">{displayValue}</div>
        <div className="connection-actions">
          {isPassword && (
            <button
              className="password-toggle-button"
              onClick={togglePasswordVisibility}
              title={showPassword ? "Hide password" : "Show password"}
              style={{
                background: 'none',
                border: 'none',
                cursor: 'pointer',
                padding: '4px',
                display: 'flex',
                alignItems: 'center',
                color: '#6b7280',
                marginRight: '4px'
              }}
            >
              {showPassword ? <EyeOff size={14} /> : <Eye size={14} />}
            </button>
          )}
          <button
            className="copy-button"
            onClick={() => onCopyToClipboard(credential.value)}
            title="Copy credential"
          >
            <Copy size={14} />
          </button>
        </div>
      </div>
      {credential.note && <div className="connection-note">{credential.note}</div>}
    </div>
  );
};

const ConnectionModal = ({ 
  isOpen, 
  onClose, 
  connectionInfo, 
  copyFeedback, 
  onCopyToClipboard 
}) => {
  if (!isOpen || !connectionInfo) return null;

  // All connection info is now in enhanced format

  // Browser-based URL opening function
  const openUrl = (url) => {
    window.open(url, '_blank', 'noopener,noreferrer');
  };

  const renderWebUrlsTable = (webUrls) => {
    if (!webUrls || webUrls.length === 0) return null;

    return (
      <div className="connection-section">
        <h4 className="connection-section-title">
          <Globe size={16} />
          Web URLs
        </h4>
        <table className="connection-table">
          <thead>
            <tr>
              <th>Description</th>
              <th>URL</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            {webUrls.map((webUrl, index) => (
              <tr key={index}>
                <td>{webUrl.description || 'Web Interface'}</td>
                <td>
                  <a 
                    href="#" 
                    onClick={(e) => {
                      e.preventDefault();
                      openUrl(webUrl.url);
                    }}
                    className="connection-url"
                  >
                    {webUrl.url}
                    <ExternalLink size={12} />
                  </a>
                </td>
                <td>
                  <button
                    className="copy-button"
                    onClick={() => onCopyToClipboard(webUrl.url)}
                    title="Copy URL"
                  >
                    <Copy size={14} />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  };

  const renderPortMappingsTable = (exposedPorts, internalPorts) => {
    const allPorts = [];
    
    // Add exposed ports
    if (exposedPorts) {
      exposedPorts.forEach(port => {
        allPorts.push({
          ...port,
          type: 'exposed',
          accessible: true
        });
      });
    }
    
    // Add internal ports that aren't already exposed
    if (internalPorts) {
      internalPorts.forEach(port => {
        const isAlreadyExposed = exposedPorts?.some(ep => 
          (ep.container_port || ep.containerPort) === (port.container_port || port.containerPort)
        );
        if (!isAlreadyExposed) {
          allPorts.push({
            ...port,
            type: 'internal',
            accessible: false,
            host_port: '-',
            hostPort: '-'
          });
        }
      });
    }

    if (allPorts.length === 0) return null;

    return (
      <div className="connection-section">
        <h4 className="connection-section-title">
          <Monitor size={16} />
          Port Mappings
        </h4>
        <table className="connection-table">
          <thead>
            <tr>
              <th>Container Port</th>
              <th>Host Port</th>
              <th>Protocol</th>
              <th>Type</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            {allPorts.map((port, index) => (
              <tr key={index}>
                <td>{port.container_port || port.containerPort}</td>
                <td>
                  {(port.host_port || port.hostPort) !== '-' ? (
                    <span className="connection-value">{port.host_port || port.hostPort}</span>
                  ) : (
                    <span style={{ color: '#9ca3af' }}>-</span>
                  )}
                </td>
                <td>{port.protocol || 'TCP'}</td>
                <td>
                  <span className={`port-type-badge port-type-${port.type}`}>
                    {port.type}
                  </span>
                </td>
                <td>{port.description || 'Service port'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  };

  const renderConnectionStrings = (connectionStrings) => {
    if (!connectionStrings || connectionStrings.length === 0) return null;

    return (
      <div className="connection-section">
        <h4 className="connection-section-title">
          <Database size={16} />
          Connection Strings
        </h4>
        <div className="connection-section-content">
          {connectionStrings.map((conn, index) => (
            <div key={index} className="connection-item">
              <div className="connection-label">{conn.description || 'Connection String'}</div>
              <div className="connection-value-row">
                <div className="connection-value connection-command">{conn.connectionString}</div>
                <button
                  className="copy-button"
                  onClick={() => onCopyToClipboard(conn.connectionString)}
                  title="Copy connection string"
                >
                  <Copy size={14} />
                </button>
              </div>
              {conn.note && <div className="connection-note">{conn.note}</div>}
            </div>
          ))}
        </div>
      </div>
    );
  };

  const renderCredentials = (credentials) => {
    if (!credentials || credentials.length === 0) return null;

    return (
      <div className="connection-section">
        <h4 className="connection-section-title">
          <Activity size={16} />
          Credentials
        </h4>
        <div className="connection-section-content">
          {credentials.map((cred, index) => {
            const isPassword = cred.description && cred.description.toLowerCase().includes('password');
            return (
              <CredentialItem 
                key={index} 
                credential={cred} 
                isPassword={isPassword}
                onCopyToClipboard={onCopyToClipboard}
              />
            );
          })}
        </div>
      </div>
    );
  };

  return createPortal(
    <div className="connection-modal-overlay" onClick={onClose}>
      <div className="connection-modal" onClick={(e) => e.stopPropagation()}>
        <div className="connection-modal-header">
          <h3 className="connection-modal-title">
            <Server size={20} />
            Connection Info: {connectionInfo.serviceName}
          </h3>
          <div className="modal-header-actions">
            {copyFeedback && (
              <div className="copy-feedback">{copyFeedback}</div>
            )}
            <button
              className="connection-modal-close"
              onClick={onClose}
              aria-label="Close"
            >
              <X size={20} />
            </button>
          </div>
        </div>

        <div className="connection-modal-content">
          {!connectionInfo.available ? (
            <div className="connection-error">
              <p>❌ Service connection information unavailable</p>
              {connectionInfo.error && (
                <div className="error-detail">{connectionInfo.error}</div>
              )}
            </div>
          ) : (
            <div className="connection-details">
                  {renderWebUrlsTable(connectionInfo.webUrls)}
                  {renderPortMappingsTable(connectionInfo.exposedPorts, connectionInfo.internalPorts)}
                  {renderConnectionStrings(connectionInfo.connectionStrings)}
                  {renderCredentials(connectionInfo.credentials)}
            </div>
          )}
        </div>
      </div>
    </div>,
    document.body
  );
};

export default ConnectionModal; 