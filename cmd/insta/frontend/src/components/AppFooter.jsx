import React from 'react';

const AppFooter = ({ currentRuntime, lastUpdated }) => {
  return (
    <footer className="footer">
      <div className="container footer-content">
        <div className="footer-status">
          <div className="status-dot dot-green pulse"></div>
          <p>Infrastructure Control Panel</p>
        </div>
        <div className="footer-info">
          Using {currentRuntime || 'unknown'} • Last refreshed: {lastUpdated.toLocaleString()}
        </div>
      </div>
    </footer>
  );
};

export default AppFooter;