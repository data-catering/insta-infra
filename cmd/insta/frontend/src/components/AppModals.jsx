import React from 'react';
import { createPortal } from 'react-dom';
import ConnectionModal from './ConnectionModal';
import LogsModal from './LogsModal';
import LogsPanel from './LogsPanel';

const AboutModal = ({ showAbout, setShowAbout, openExternalLink }) => {
  if (!showAbout) return null;

  return createPortal(
    <div className="connection-modal-overlay" onClick={() => setShowAbout(false)}>
      <div className="connection-modal about-modal" onClick={(e) => e.stopPropagation()}>
        <div className="connection-modal-header">
          <h3 className="connection-modal-title">
            About insta-infra
          </h3>
          <button 
            onClick={() => setShowAbout(false)}
            className="connection-modal-close"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M6 18L18 6M6 6L18 18" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </button>
        </div>
        
        <div className="connection-modal-content">
          <div className="about-content">
            <div className="about-hero">
              <div className="about-logo">
                <svg width="48" height="48" className="logo-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M21 9.5V5C21 3.89543 20.1046 3 19 3H5C3.89543 3 3 3.89543 3 5V9.5M21 9.5H3M21 9.5V19C21 20.1046 20.1046 21 19 21H5C3.89543 21 3 20.1046 3 19V9.5M9 14H15" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </div>
              <h2 className="about-title">insta-infra</h2>
              <p className="about-subtitle">Instant Infrastructure Services</p>
            </div>

            <div className="about-description">
              <p>
                <strong>insta-infra</strong> is a powerful service management tool that allows you to quickly start, 
                stop, and manage infrastructure services like databases, monitoring tools, and data platforms.
              </p>
              <p>
                With support for over 70 services including PostgreSQL, Redis, Grafana, Elasticsearch, 
                and many more, insta-infra simplifies development and testing workflows.
              </p>
            </div>

            <div className="about-features">
              <h4>Key Features:</h4>
              <ul>
                <li>
                  <svg width="16" height="16" className="feature-icon" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                  </svg>
                  One-click service deployment with Docker/Podman
                </li>
                <li>
                  <svg width="16" height="16" className="feature-icon" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                  </svg>
                  Automatic dependency management
                </li>
                <li>
                  <svg width="16" height="16" className="feature-icon" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                  </svg>
                  Data persistence options
                </li>
                <li>
                  <svg width="16" height="16" className="feature-icon" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                  </svg>
                  Easy connection details and web UI access
                </li>
                <li>
                  <svg width="16" height="16" className="feature-icon" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                  </svg>
                  Cross-platform support (macOS, Linux, Windows)
                </li>
              </ul>
            </div>

            <div className="about-links">
              <h4>Resources:</h4>
              <div className="link-buttons">
                <button 
                  onClick={() => openExternalLink("https://github.com/data-catering/insta-infra")}
                  className="link-button"
                  title="Open GitHub repository in your browser"
                >
                  <svg width="16" height="16" className="link-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M9 19C4 20.5 4 16.5 2 16M22 16V19C22 20.1046 21.1046 21 20 21H4C2.89543 21 2 20.1046 2 19V16M22 16L18.5 12.5L22 9V16Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                  GitHub Repository
                </button>
                <button 
                  onClick={() => openExternalLink("https://github.com/data-catering/insta-infra/blob/main/README.md")}
                  className="link-button"
                  title="Open documentation in your browser"
                >
                  <svg width="16" height="16" className="link-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M9 12H15M9 16H15M17 21H7C5.89543 21 5 20.1046 5 19V5C5 3.89543 5.89543 3 7 3H12.5858C12.851 3 13.1054 3.10536 13.2929 3.29289L18.7071 8.70711C18.8946 8.89464 19 9.149 19 9.41421V19C19 20.1046 18.1046 21 17 21Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                  Documentation
                </button>
              </div>
            </div>

            <div className="about-footer">
              <p className="about-version">Version 1.0.0</p>
              <p className="about-copyright">Built with ❤️ for developers</p>
            </div>
          </div>
        </div>
      </div>
    </div>,
    document.body
  );
};

const AppModals = ({
  showAbout,
  setShowAbout,
  showConnectionModal,
  setShowConnectionModal,
  showLogsModal,
  setShowLogsModal,
  selectedService,
  showLogsPanel,
  setShowLogsPanel,
  copyFeedback,
  copyToClipboard,
  openExternalLink
}) => {
  return (
    <>
      {/* About Modal */}
      <AboutModal 
        showAbout={showAbout}
        setShowAbout={setShowAbout}
        openExternalLink={openExternalLink}
      />
      
      {/* Connection Info Modal */}
      <ConnectionModal 
        isOpen={showConnectionModal}
        onClose={() => setShowConnectionModal(false)}
        copyFeedback={copyFeedback}
        onCopyToClipboard={copyToClipboard}
      />
      
      {/* Logs Modal */}
      {showLogsModal && (
        <LogsModal 
          isOpen={true}
          onClose={() => setShowLogsModal(false)}
          serviceName={selectedService}
        />
      )}
      
      {/* Logs Panel */}
      <LogsPanel 
        isVisible={showLogsPanel}
        onToggle={() => setShowLogsPanel(!showLogsPanel)}
      />
    </>
  );
};

export default AppModals;