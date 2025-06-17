import React from 'react';
import { createPortal } from 'react-dom';
import { X, Copy, ExternalLink, Database, Server, Globe, Activity, Monitor } from 'lucide-react';

const ConnectionModal = ({ 
  isOpen, 
  onClose, 
  connectionInfo, 
  copyFeedback, 
  onCopyToClipboard 
}) => {
  if (!isOpen || !connectionInfo) return null;

  // Handle both old and new connection info formats
  const isEnhanced = connectionInfo.webUrls || connectionInfo.exposedPorts || connectionInfo.connectionStrings;

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
          ep.containerPort === port.containerPort
        );
        if (!isAlreadyExposed) {
          allPorts.push({
            ...port,
            type: 'internal',
            accessible: false,
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
                <td>{port.containerPort}</td>
                <td>
                  {port.hostPort !== '-' ? (
                    <span className="connection-value">{port.hostPort}</span>
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
          {credentials.map((cred, index) => (
            <div key={index} className="connection-item">
              <div className="connection-label">{cred.description || 'Credential'}</div>
              <div className="connection-value-row">
                <div className="connection-value">{cred.value}</div>
                <button
                  className="copy-button"
                  onClick={() => onCopyToClipboard(cred.value)}
                  title="Copy credential"
                >
                  <Copy size={14} />
                </button>
              </div>
              {cred.note && <div className="connection-note">{cred.note}</div>}
            </div>
          ))}
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
              <div className="status-info">
                Service Status: <span className={`status-badge status-${connectionInfo.status?.toLowerCase() || 'unknown'}`}>
                  {connectionInfo.status || 'Unknown'}
                </span>
              </div>
            </div>
          ) : (
            <div className="connection-details">
              {/* Status Information */}
              <div className="status-info">
                Service Status: <span className={`status-badge status-${connectionInfo.status?.toLowerCase() || 'running'}`}>
                  {connectionInfo.status || 'Running'}
                </span>
              </div>

              {/* Enhanced format with tables */}
              {isEnhanced ? (
                <>
                  {renderWebUrlsTable(connectionInfo.webUrls)}
                  {renderPortMappingsTable(connectionInfo.exposedPorts, connectionInfo.internalPorts)}
                  {renderConnectionStrings(connectionInfo.connectionStrings)}
                  {renderCredentials(connectionInfo.credentials)}
                </>
              ) : (
                /* Legacy format */
                <>
                  {connectionInfo.hasWebUI && connectionInfo.webURL && (
                    <div className="connection-section">
                      <h4 className="connection-section-title">
                        <Globe size={16} />
                        Web Interface
                      </h4>
                      <div className="connection-item">
                        <div className="connection-label">URL</div>
                        <div className="connection-value-row">
                          <a 
                            href="#" 
                            onClick={(e) => {
                              e.preventDefault();
                              openUrl(connectionInfo.webURL);
                            }}
                            className="connection-link"
                          >
                            {connectionInfo.webURL}
                            <ExternalLink size={14} />
                          </a>
                          <button
                            className="copy-button"
                            onClick={() => onCopyToClipboard(connectionInfo.webURL)}
                            title="Copy URL"
                          >
                            <Copy size={14} />
                          </button>
                        </div>
                      </div>
                    </div>
                  )}

                  {(connectionInfo.hostPort || connectionInfo.containerPort) && (
                    <div className="connection-section">
                      <h4 className="connection-section-title">
                        <Monitor size={16} />
                        Port Information
                      </h4>
                      {connectionInfo.hostPort && (
                        <div className="connection-item">
                          <div className="connection-label">Host Port</div>
                          <div className="connection-value-row">
                            <div className="connection-value">{connectionInfo.hostPort}</div>
                            <button
                              className="copy-button"
                              onClick={() => onCopyToClipboard(connectionInfo.hostPort)}
                              title="Copy host port"
                            >
                              <Copy size={14} />
                            </button>
                          </div>
                        </div>
                      )}
                      {connectionInfo.containerPort && (
                        <div className="connection-item">
                          <div className="connection-label">Container Port</div>
                          <div className="connection-value-row">
                            <div className="connection-value">{connectionInfo.containerPort}</div>
                            <button
                              className="copy-button"
                              onClick={() => onCopyToClipboard(connectionInfo.containerPort)}
                              title="Copy container port"
                            >
                              <Copy size={14} />
                            </button>
                          </div>
                        </div>
                      )}
                    </div>
                  )}

                  {(connectionInfo.username || connectionInfo.password) && (
                    <div className="connection-credentials">
                      <h4 className="credentials-title">Default Credentials</h4>
                      {connectionInfo.username && (
                        <div className="connection-item">
                          <div className="connection-label">Username</div>
                          <div className="connection-value-row">
                            <div className="connection-value">{connectionInfo.username}</div>
                            <button
                              className="copy-button"
                              onClick={() => onCopyToClipboard(connectionInfo.username)}
                              title="Copy username"
                            >
                              <Copy size={14} />
                            </button>
                          </div>
                        </div>
                      )}
                      {connectionInfo.password && (
                        <div className="connection-item">
                          <div className="connection-label">Password</div>
                          <div className="connection-value-row">
                            <div className="connection-value">{connectionInfo.password}</div>
                            <button
                              className="copy-button"
                              onClick={() => onCopyToClipboard(connectionInfo.password)}
                              title="Copy password"
                            >
                              <Copy size={14} />
                            </button>
                          </div>
                        </div>
                      )}
                    </div>
                  )}

                  {connectionInfo.connectionString && (
                    <div className="connection-section">
                      <h4 className="connection-section-title">
                        <Database size={16} />
                        Connection String
                      </h4>
                      <div className="connection-item">
                        <div className="connection-value-row">
                          <div className="connection-value connection-command">{connectionInfo.connectionString}</div>
                          <button
                            className="copy-button"
                            onClick={() => onCopyToClipboard(connectionInfo.connectionString)}
                            title="Copy connection string"
                          >
                            <Copy size={14} />
                          </button>
                        </div>
                      </div>
                    </div>
                  )}

                  {connectionInfo.connectionCommand && (
                    <div className="connection-section">
                      <h4 className="connection-section-title">
                        <Activity size={16} />
                        Connection Command
                      </h4>
                      <div className="connection-item">
                        <div className="connection-value-row">
                          <div className="connection-value connection-command">{connectionInfo.connectionCommand}</div>
                          <button
                            className="copy-button"
                            onClick={() => onCopyToClipboard(connectionInfo.connectionCommand)}
                            title="Copy command"
                          >
                            <Copy size={14} />
                          </button>
                        </div>
                      </div>
                    </div>
                  )}
                </>
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