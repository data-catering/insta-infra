import React, { useState, useEffect, useRef, useCallback } from 'react';
import { getAppLogEntries, getAppLogsSince } from '../api/client';
import { useErrorHandler } from './ErrorMessage';

const LogsPanel = ({ isVisible, onToggle }) => {
  const [logs, setLogs] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [isPaused, setIsPaused] = useState(false);
  const [lastUpdateTime, setLastUpdateTime] = useState(null);
  const logsEndRef = useRef(null);
  const intervalRef = useRef(null);
  const { addError } = useErrorHandler();

  // Auto-scroll to bottom when new logs arrive
  const scrollToBottom = useCallback(() => {
    if (!isPaused && logsEndRef.current) {
      logsEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [isPaused]);

  // Load initial logs
  const loadInitialLogs = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const logEntries = await getAppLogEntries();
      setLogs(logEntries || []);
      if (logEntries && logEntries.length > 0) {
        setLastUpdateTime(new Date().toISOString());
      }
    } catch (err) {
      console.error('Failed to load initial logs:', err);
      setError('Failed to load logs');
      
      // Add detailed error message for user
      addError({
        type: 'error',
        title: 'Application Logs Failed',
        message: 'Unable to load application logs',
        details: `Error: ${err.message || err.toString()} | Time: ${new Date().toLocaleTimeString()}`,
        metadata: {
          action: 'load_app_logs',
          errorMessage: err.message || err.toString(),
          timestamp: new Date().toISOString()
        },
        actions: [
          {
            label: 'Retry',
            onClick: () => {
              setError(null);
              loadInitialLogs();
            },
            variant: 'primary'
          }
        ]
      });
    } finally {
      setIsLoading(false);
    }
  };

  // Poll for new logs
  const pollForNewLogs = async () => {
    if (!lastUpdateTime || isPaused) return;

    try {
      const newLogEntries = await getAppLogsSince(lastUpdateTime);
      if (newLogEntries && newLogEntries.length > 0) {
        setLogs(prevLogs => [...prevLogs, ...newLogEntries]);
        setLastUpdateTime(new Date().toISOString());
        // Scroll to bottom after a short delay to ensure DOM is updated
        setTimeout(scrollToBottom, 100);
      }
    } catch (err) {
      console.error('Failed to poll for new logs:', err);
      // Don't show error for polling failures to avoid spam
    }
  };

  // Set up polling when panel is visible
  useEffect(() => {
    if (isVisible) {
      loadInitialLogs();
      
      // Set up polling interval
      intervalRef.current = setInterval(pollForNewLogs, 1000); // Poll every second
      
      return () => {
        if (intervalRef.current) {
          clearInterval(intervalRef.current);
        }
      };
    } else {
      // Clear interval when panel is hidden
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    }
  }, [isVisible, lastUpdateTime, isPaused]);

  // Format timestamp for display
  const formatTime = (timestamp) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString('en-US', { 
      hour12: false, 
      hour: '2-digit', 
      minute: '2-digit', 
      second: '2-digit' 
    });
  };

  // Clear logs
  const clearLogs = () => {
    setLogs([]);
    setLastUpdateTime(new Date().toISOString());
  };

  // Toggle pause
  const togglePause = () => {
    setIsPaused(!isPaused);
  };

  if (!isVisible) return null;

  return (
    <div className="logs-panel">
      <div className="logs-panel-header">
        <h3>Application Logs</h3>
        <div className="logs-panel-controls">
          <button 
            onClick={togglePause}
            className={`logs-control-btn ${isPaused ? 'paused' : ''}`}
            title={isPaused ? 'Resume log updates' : 'Pause log updates'}
          >
            {isPaused ? '▶️' : '⏸️'}
          </button>
          <button 
            onClick={clearLogs}
            className="logs-control-btn"
            title="Clear logs"
          >
            🗑️
          </button>
          <button 
            onClick={onToggle}
            className="logs-control-btn close-btn"
            title="Close logs panel (Cmd+L)"
          >
            ✕
          </button>
        </div>
      </div>
      
      <div className="logs-panel-content">
        {isLoading && logs.length === 0 && (
          <div className="logs-loading">Loading logs...</div>
        )}
        
        {error && (
          <div className="logs-error">
            Error: {error}
            <button onClick={loadInitialLogs} className="retry-btn">Retry</button>
          </div>
        )}
        
        {!isLoading && !error && logs.length === 0 && (
          <div className="logs-empty">No logs available</div>
        )}
        
        <div className="logs-list">
          {logs.map((log, index) => (
            <div key={index} className="log-entry">
              <span className="log-timestamp">
                {formatTime(log.timestamp)}
              </span>
              <span className="log-message">{log.message}</span>
            </div>
          ))}
          <div ref={logsEndRef} />
        </div>
      </div>
      
      {isPaused && (
        <div className="logs-paused-indicator">
          Log updates paused
        </div>
      )}
    </div>
  );
};

export default LogsPanel; 