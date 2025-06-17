import React, { useState, useEffect, useRef } from 'react';
import { createPortal } from 'react-dom';
import { X, Download, Search, Filter, ScrollText, AlertCircle } from 'lucide-react';
import { getServiceLogs, wsClient, WS_MSG_TYPES } from '../api/client';

const LogsModal = ({ isOpen, onClose, serviceName, selectedDependency = null }) => {
  const [logs, setLogs] = useState([]);
  const [isStreaming, setIsStreaming] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [logLevel, setLogLevel] = useState('all');
  const [autoScroll, setAutoScroll] = useState(true);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  
  const logsContainerRef = useRef(null);
  const logStreamRef = useRef(null);

  // Determine which service to show logs for
  const targetService = selectedDependency?.serviceName || serviceName;
  const isShowingDependency = !!selectedDependency;

  // Load initial logs when modal opens
  useEffect(() => {
    if (isOpen && targetService) {
      loadInitialLogs();
      setupEventListeners();
      // Start log streaming with a small delay to ensure event listeners are set up
      const timer = setTimeout(() => {
        startLogStream();
      }, 100);
      
      return () => {
        clearTimeout(timer);
      };
    }
    
    return () => {
      cleanup();
    };
  }, [isOpen, targetService]);

  // Auto-scroll effect
  useEffect(() => {
    if (autoScroll && logsContainerRef.current) {
      logsContainerRef.current.scrollTop = logsContainerRef.current.scrollHeight;
    }
  }, [logs, autoScroll]);

  // Cleanup effect when modal closes or service changes
  useEffect(() => {
    return () => {
      if (!isOpen) {
        cleanup();
      }
    };
  }, [isOpen]);

  const loadInitialLogs = async () => {
    setIsLoading(true);
    setError('');
    try {
      const initialLogs = await getServiceLogs(targetService);
      
      if (initialLogs && initialLogs.logs && initialLogs.logs.length > 0) {
        const formattedLogs = initialLogs.logs.map((log, index) => ({
          id: `initial-${index}`,
          timestamp: log.timestamp || new Date().toISOString(),
          message: log.message || log,
          level: detectLogLevel(log.message || log)
        }));
        setLogs(formattedLogs);
      } else {
        setLogs([]);
      }
    } catch (err) {
      console.error(`Failed to load logs for ${targetService}:`, err);
      const errorMessage = err?.message || err?.toString() || 'Unknown error occurred';
      setError(`Failed to load logs: ${errorMessage}`);
    } finally {
      setIsLoading(false);
    }
  };

  const setupEventListeners = () => {
    // Subscribe to service logs via WebSocket
    const logCallback = (data) => {
      if (data.service_name === targetService || data.serviceName === targetService) {
        // If we receive a log, it means streaming is working
        if (!isStreaming) {
          setIsStreaming(true);
          setError(''); // Clear any previous errors
        }
        
        const newLog = {
          id: `stream-${Date.now()}-${Math.random()}`,
          timestamp: data.timestamp || new Date().toISOString(),
          message: data.message || data.log || '',
          level: detectLogLevel(data.message || data.log || '')
        };
        
        setLogs(prevLogs => [...prevLogs, newLog]);
      }
    };

    // Subscribe to WebSocket service logs
    wsClient.subscribe(WS_MSG_TYPES.SERVICE_LOGS, logCallback);
    
    // Store callback for cleanup
    setupEventListeners.logCallback = logCallback;
  };

  const cleanup = () => {
    // Stop streaming if active
    if (isStreaming) {
      setIsStreaming(false);
    }
    
    // Unsubscribe from WebSocket events
    if (setupEventListeners.logCallback) {
      wsClient.unsubscribe(WS_MSG_TYPES.SERVICE_LOGS, setupEventListeners.logCallback);
      setupEventListeners.logCallback = null;
    }
  };

  const startLogStream = async () => {
    // Don't start if already streaming
    if (isStreaming) {
      return;
    }

    try {
      setError('');
      // In the WebSocket implementation, streaming is automatic once subscribed
      setIsStreaming(true);
    } catch (err) {
      console.error(`Failed to start log stream for ${targetService}:`, err);
      const errorMessage = err?.message || err?.toString() || 'Unknown error occurred';
      setError(`Failed to start log stream: ${errorMessage}`);
    }
  };

  const stopLogStream = async () => {
    if (!isStreaming) {
      return; // Don't try to stop if not streaming
    }

    try {
      // In the WebSocket implementation, we just set streaming to false
      // The actual unsubscribe happens in cleanup()
      setIsStreaming(false);
    } catch (err) {
      console.error(`Failed to stop log stream for ${targetService}:`, err);
      setIsStreaming(false); // Set to false regardless
    }
  };

  const detectLogLevel = (message) => {
    if (!message) return 'info';
    // Logs may be in JSON format, so we need to parse it
    try {
      const json = JSON.parse(message);
      if (json.level) return json.level;
      // If not JSON, parse the message
      const upperMessage = message.toUpperCase();
      if (upperMessage.includes('ERROR') || upperMessage.includes('FATAL')) return 'error';
      if (upperMessage.includes('WARN')) return 'warn';
      if (upperMessage.includes('INFO')) return 'info';
      if (upperMessage.includes('DEBUG')) return 'debug';
    } catch (e) {
      // Not JSON, continue with string parsing
      return 'info';
    }
  };

  const getLogLevelColor = (level) => {
    switch (level) {
      case 'error': return 'text-red-400';
      case 'warn': return 'text-yellow-400';
      case 'info': return 'text-blue-400';
      case 'debug': return 'text-gray-400';
      default: return 'text-gray-300';
    }
  };

  const filteredLogs = logs.filter(log => {
    const matchesSearch = searchTerm === '' || 
      log.message.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesLevel = logLevel === 'all' || log.level === logLevel;
    return matchesSearch && matchesLevel;
  });

  const downloadLogs = () => {
    const logContent = filteredLogs
      .map(log => `[${new Date(log.timestamp).toLocaleString()}] ${log.message}`)
      .join('\n');
    
    const blob = new Blob([logContent], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${targetService}-logs-${new Date().toISOString().split('T')[0]}.txt`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const clearLogs = () => {
    setLogs([]);
  };

  if (!isOpen) return null;

  return createPortal(
    <div className="connection-modal-overlay" onClick={onClose}>
      <div className="connection-modal logs-modal" onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div className="connection-modal-header">
          <div className="connection-modal-title">
            <ScrollText size={20} />
            <span>
              Logs - {targetService}
              {isShowingDependency && (
                <span className="dependency-context">
                  {' '}(dependency of {serviceName})
                  {selectedDependency.failureReason && (
                    <div className="failure-context">
                      <AlertCircle size={14} />
                      {selectedDependency.failureReason}
                    </div>
                  )}
                </span>
              )}
            </span>
          </div>
          <div className="modal-header-actions">
            {/* Close button */}
            <button className="connection-modal-close" onClick={onClose}>
              <X size={20} />
            </button>
          </div>
        </div>

        {/* Controls */}
        <div className="logs-controls">
          <div className="logs-search">
            <Search size={16} className="search-icon" />
            <input
              type="text"
              placeholder="Search logs..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="logs-search-input"
            />
          </div>
          
          <div className="logs-filter">
            <Filter size={16} className="filter-icon" />
            <select
              value={logLevel}
              onChange={(e) => setLogLevel(e.target.value)}
              className="logs-level-select"
            >
              <option value="all">All Levels</option>
              <option value="error">Error</option>
              <option value="warn">Warning</option>
              <option value="info">Info</option>
              <option value="debug">Debug</option>
            </select>
          </div>

          <div className="logs-options">
            <label className="logs-checkbox">
              <input
                type="checkbox"
                checked={autoScroll}
                onChange={(e) => setAutoScroll(e.target.checked)}
              />
              Auto-scroll
            </label>
            <button onClick={clearLogs} className="logs-clear-button">
              Clear
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="connection-modal-content logs-content">
          {error && (
            <div className="logs-error">
              {error}
            </div>
          )}
          
          {isLoading && (
            <div className="logs-loading">
              Loading logs...
            </div>
          )}

          <div 
            className="logs-container"
            ref={logsContainerRef}
            onScroll={(e) => {
              const { scrollTop, scrollHeight, clientHeight } = e.target;
              const isAtBottom = scrollTop + clientHeight >= scrollHeight - 10;
              if (!isAtBottom && autoScroll) {
                setAutoScroll(false);
              }
            }}
          >
            {filteredLogs.length === 0 ? (
              <div className="logs-empty">
                {isLoading ? 'Loading logs...' : 
                 logs.length === 0 ? `No logs available for ${targetService}. Try starting the log stream or check if the service is running.` :
                 `No logs match current filters (${searchTerm ? `search: "${searchTerm}"` : ''}${logLevel !== 'all' ? `, level: ${logLevel}` : ''}).`}
              </div>
            ) : (
              filteredLogs.map((log) => (
                <div key={log.id} className="log-entry">
                  <span className="log-timestamp">
                    {new Date(log.timestamp).toLocaleTimeString()}
                  </span>
                  <span className={`log-level ${getLogLevelColor(log.level)}`}>
                    [{log.level.toUpperCase()}]
                  </span>
                  <span className="log-message">{log.message}</span>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>,
    document.body
  );
};

export default LogsModal; 