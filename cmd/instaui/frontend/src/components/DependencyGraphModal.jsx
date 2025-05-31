import React, { useState, useEffect, useCallback } from 'react';
import { createPortal } from 'react-dom';
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
  addEdge,
  ConnectionLineType,
  Position,
  MarkerType,
  Handle,
} from 'reactflow';
import { 
  X, 
  RotateCcw, 
  ZoomIn, 
  ZoomOut, 
  Maximize2,
  Play,
  Square,
  ScrollText,
  Info
} from 'lucide-react';
import { GetDependencyGraph, StartService, StopService, GetServiceConnectionInfo, GetServiceDependencyGraph } from "../../wailsjs/go/main/App";

// Import React Flow styles
import 'reactflow/dist/style.css';

// Custom node component for services
const ServiceNode = ({ data, selected }) => {
  const [showActions, setShowActions] = useState(false);

  const getStatusIcon = () => {
    switch (data.status) {
      case 'running':
        return (
          <div className="node-status-icon running">
            <div className="status-dot running"></div>
          </div>
        );
      case 'stopped':
        return (
          <div className="node-status-icon stopped">
            <div className="status-dot stopped"></div>
          </div>
        );
      case 'error':
        return (
          <div className="node-status-icon error">
            <div className="status-dot error"></div>
          </div>
        );
      default:
        return (
          <div className="node-status-icon unknown">
            <div className="status-dot unknown"></div>
          </div>
        );
    }
  };

  const getTypeIcon = () => {
    switch (data.type?.toLowerCase()) {
      case 'database':
        return (
          <svg width="16" height="16" className="node-type-icon" viewBox="0 0 24 24" fill="none">
            <path d="M12 8C15.3137 8 18 6.65685 18 5C18 3.34315 15.3137 2 12 2C8.68629 2 6 3.34315 6 5C6 6.65685 8.68629 8 12 8Z" stroke="currentColor" strokeWidth="2"/>
            <path d="M6 5V19C6 20.6569 8.68629 22 12 22C15.3137 22 18 20.6569 18 19V5" stroke="currentColor" strokeWidth="2"/>
            <path d="M6 12C6 13.6569 8.68629 15 12 15C15.3137 15 18 13.6569 18 12" stroke="currentColor" strokeWidth="2"/>
          </svg>
        );
      case 'messaging':
        return (
          <svg width="16" height="16" className="node-type-icon" viewBox="0 0 24 24" fill="none">
            <path d="M8 10H8.01M12 10H12.01M16 10H16.01M9 16H5C3.89543 16 3 15.1046 3 14V6C3 4.89543 3.89543 4 5 4H19C20.1046 4 21 4.89543 21 6V14C21 15.1046 20.1046 16 19 16H14L9 21V16Z" stroke="currentColor" strokeWidth="2"/>
          </svg>
        );
      default:
        return (
          <svg width="16" height="16" className="node-type-icon" viewBox="0 0 24 24" fill="none">
            <path d="M9 12H15M9 16H15M17 21H7C5.89543 21 5 20.1046 5 19V5C5 3.89543 5.89543 3 7 3H12.5858C12.851 3 13.1054 3.10536 13.2929 3.29289L18.7071 8.70711C18.8946 8.89464 19 9.149 19 9.41421V19C19 20.1046 18.1046 21 17 21Z" stroke="currentColor" strokeWidth="2"/>
          </svg>
        );
    }
  };

  return (
    <div 
      className={`graph-node ${data.status} ${selected ? 'selected' : ''}`}
      onMouseEnter={() => setShowActions(true)}
      onMouseLeave={() => setShowActions(false)}
      style={{
        borderColor: data.color,
        boxShadow: selected ? `0 0 20px ${data.color}40` : 'none'
      }}
    >
      <Handle type="target" position={Position.Left} />
      <Handle type="source" position={Position.Right} />
      {/* Node Header */}
      <div className="node-header">
        <div className="node-icon-section">
          {getTypeIcon()}
          {getStatusIcon()}
        </div>
        <div className="node-title">
          <span className="node-name">{data.serviceName}</span>
          <span className="node-type">{data.type}</span>
        </div>
      </div>

      {/* Node Content */}
      <div className="node-content">
        <div className="node-status">
          <span className={`status-text ${data.status}`}>
            {data.status.charAt(0).toUpperCase() + data.status.slice(1)}
          </span>
          {data.dependencies && data.dependencies.length > 0 && (
            <span className="dependencies-count">
              {data.dependencies.length} deps
            </span>
          )}
        </div>
      </div>

      {/* Action buttons - shown on hover */}
      {showActions && (
        <div className="node-actions">
          {data.status === 'stopped' && (
            <button 
              className="node-action-btn start-btn" 
              onClick={(e) => {
                e.stopPropagation();
                window.dispatchEvent(new CustomEvent('graph-start-service', { detail: data.serviceName }));
              }}
              title="Start service"
            >
              <Play size={12} />
            </button>
          )}
          {data.status === 'running' && (
            <button 
              className="node-action-btn stop-btn" 
              onClick={(e) => {
                e.stopPropagation();
                window.dispatchEvent(new CustomEvent('graph-stop-service', { detail: data.serviceName }));
              }}
              title="Stop service"
            >
              <Square size={12} />
            </button>
          )}
          <button 
            className="node-action-btn info-btn" 
            onClick={(e) => {
              e.stopPropagation();
              window.dispatchEvent(new CustomEvent('graph-service-info', { detail: data.serviceName }));
            }}
            title="Service details"
          >
            <Info size={12} />
          </button>
          <button 
            className="node-action-btn logs-btn" 
            onClick={(e) => {
              e.stopPropagation();
              window.dispatchEvent(new CustomEvent('graph-service-logs', { detail: data.serviceName }));
            }}
            title="View logs"
          >
            <ScrollText size={12} />
          </button>
        </div>
      )}
    </div>
  );
};

// Define custom node types
const nodeTypes = {
  service: ServiceNode,
};

function DependencyGraphModal({ isOpen, onClose, serviceName, onServiceAction }) {
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [selectedNode, setSelectedNode] = useState(null);
  const [fitViewOptions] = useState({
    padding: 50,
    includeHiddenNodes: false,
    minZoom: 0.1,
    maxZoom: 1.5,
  });

  // Load graph data
  const loadGraphData = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      let graphData;
      
      if (serviceName) {
        // Load focused dependency graph for specific service
        graphData = await GetServiceDependencyGraph(serviceName);
      } else {
        // Load complete dependency graph (fallback)
        graphData = await GetDependencyGraph();
      }
      
      if (!graphData || !graphData.nodes || !graphData.edges) {
        throw new Error('Invalid graph data structure received');
      }

      // Transform nodes for React Flow
      const flowNodes = graphData.nodes.map(node => {
        return {
          id: node.id,
          type: 'service',
          position: { x: node.position.x, y: node.position.y },
          data: {
            serviceName: node.data.serviceName,
            type: node.data.type,
            status: node.data.status,
            health: node.data.health,
            dependencies: node.data.dependencies || [],
            color: node.data.color,
          },
          style: {
            backgroundColor: 'transparent',
            border: 'none',
            padding: 0,
            borderRadius: 8,
          },
        };
      });

      // Transform edges for React Flow
      const flowEdges = graphData.edges.map(edge => {
        return {
          id: edge.id,
          source: edge.source,
          target: edge.target,
          type: 'smoothstep',
          animated: edge.data.animated,
          style: {
            stroke: edge.data.color,
            strokeWidth: 2,
          },
          markerEnd: {
            type: MarkerType.ArrowClosed,
            color: edge.data.color,
          },
          label: edge.data.label,
          labelStyle: {
            fontSize: 10,
            fill: '#9ca3af',
          },
          labelBgPadding: [4, 2],
          labelBgBorderRadius: 4,
          labelBgStyle: {
            fill: '#1f2937',
            fillOpacity: 0.8,
          },
        };
      });

      setNodes(flowNodes);
      setEdges(flowEdges);

    } catch (err) {
      console.error('Failed to load dependency graph:', err);
      setError(err.message || 'Failed to load dependency graph');
    } finally {
      setIsLoading(false);
    }
  }, [setNodes, setEdges, serviceName]);

  // Load data when modal opens
  useEffect(() => {
    if (isOpen) {
      loadGraphData();
    }
  }, [isOpen, loadGraphData]);

  // Handle node clicks
  const onNodeClick = useCallback((event, node) => {
    setSelectedNode(node.id === selectedNode ? null : node.id);
  }, [selectedNode]);

  // Handle edge clicks
  const onEdgeClick = useCallback((event, edge) => {
    // Handle edge click if needed
  }, []);

  // Handle service actions from custom events
  useEffect(() => {
    const handleStartService = async (event) => {
      const serviceName = event.detail;
      if (onServiceAction) {
        try {
          await onServiceAction('start', serviceName);
          // Reload graph data to reflect changes
          setTimeout(loadGraphData, 1000);
        } catch (error) {
          console.error('Failed to start service:', error);
        }
      }
    };

    const handleStopService = async (event) => {
      const serviceName = event.detail;
      if (onServiceAction) {
        try {
          await onServiceAction('stop', serviceName);
          // Reload graph data to reflect changes
          setTimeout(loadGraphData, 1000);
        } catch (error) {
          console.error('Failed to stop service:', error);
        }
      }
    };

    const handleServiceInfo = (event) => {
      const serviceName = event.detail;
      if (onServiceAction) {
        onServiceAction('info', serviceName);
      }
    };

    const handleServiceLogs = (event) => {
      const serviceName = event.detail;
      if (onServiceAction) {
        onServiceAction('logs', serviceName);
      }
    };

    window.addEventListener('graph-start-service', handleStartService);
    window.addEventListener('graph-stop-service', handleStopService);
    window.addEventListener('graph-service-info', handleServiceInfo);
    window.addEventListener('graph-service-logs', handleServiceLogs);

    return () => {
      window.removeEventListener('graph-start-service', handleStartService);
      window.removeEventListener('graph-stop-service', handleStopService);
      window.removeEventListener('graph-service-info', handleServiceInfo);
      window.removeEventListener('graph-service-logs', handleServiceLogs);
    };
  }, [onServiceAction, loadGraphData]);

  if (!isOpen) return null;

  try {
    return createPortal(
      <div className="graph-modal-overlay" onClick={onClose}>
        <div className="graph-modal" onClick={(e) => e.stopPropagation()}>
          {/* Modal Header */}
          <div className="graph-modal-header">
            <h3 className="graph-modal-title">
              {serviceName ? `Dependencies - ${serviceName}` : 'Dependency Graph'}
            </h3>
            <div className="graph-header-actions">
              <button
                onClick={loadGraphData}
                className="graph-header-btn"
                disabled={isLoading}
                title="Refresh graph"
              >
                <RotateCcw size={16} className={isLoading ? 'animate-spin' : ''} />
              </button>
              <button
                onClick={onClose}
                className="graph-header-btn close-btn"
                title="Close"
              >
                <X size={16} />
              </button>
            </div>
          </div>

          {/* Modal Content */}
          <div className="graph-modal-content">
            {error ? (
              <div className="graph-error">
                <div className="error-icon">⚠️</div>
                <div className="error-message">
                  <h4>Failed to load dependency graph</h4>
                  <p>{error}</p>
                  <button onClick={loadGraphData} className="retry-btn">
                    Try Again
                  </button>
                </div>
              </div>
            ) : (
              <div className="graph-container">
                <ReactFlow
                  nodes={nodes}
                  edges={edges}
                  onNodesChange={onNodesChange}
                  onEdgesChange={onEdgesChange}
                  onNodeClick={onNodeClick}
                  onEdgeClick={onEdgeClick}
                  nodeTypes={nodeTypes}
                  connectionLineType={ConnectionLineType.SmoothStep}
                  fitView={true}
                  fitViewOptions={fitViewOptions}
                  attributionPosition="bottom-left"
                >
                  <Background variant="dots" gap={20} size={1} />
                  <Controls 
                    showZoom={true}
                    showFitView={true}
                    showInteractive={true}
                  />
                  <MiniMap 
                    nodeStrokeWidth={3}
                    nodeColor={(node) => node.data.color}
                    zoomable
                    pannable
                    position="top-right"
                  />
                </ReactFlow>
                
                {isLoading && (
                  <div className="graph-loading">
                    <div className="loading-spinner"></div>
                    <p>Loading dependency graph...</p>
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Modal Footer */}
          <div className="graph-modal-footer">
            <div className="graph-legend">
              <div className="legend-item">
                <div className="legend-dot running"></div>
                <span>Running</span>
              </div>
              <div className="legend-item">
                <div className="legend-dot stopped"></div>
                <span>Stopped</span>
              </div>
              <div className="legend-item">
                <div className="legend-dot error"></div>
                <span>Error</span>
              </div>
              <div className="legend-item">
                <div className="legend-line"></div>
                <span>Dependency</span>
              </div>
            </div>
            <div className="graph-stats">
              <span>{nodes.length} services</span>
              <span>{edges.length} dependencies</span>
            </div>
          </div>
        </div>
      </div>,
      document.body
    );
  } catch (err) {
    console.error('Error rendering DependencyGraphModal:', err);
    return createPortal(
      <div className="graph-modal-overlay" onClick={onClose}>
        <div className="graph-modal" onClick={(e) => e.stopPropagation()}>
          <div className="graph-modal-header">
            <h3 className="graph-modal-title">Graph Error</h3>
            <button onClick={onClose} className="graph-header-btn close-btn">
              <X size={16} />
            </button>
          </div>
          <div className="graph-modal-content">
            <div className="graph-error">
              <div className="error-icon">⚠️</div>
              <div className="error-message">
                <h4>Failed to render dependency graph</h4>
                <p>Error: {err.message}</p>
                <button onClick={onClose} className="retry-btn">
                  Close
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>,
      document.body
    );
  }
}

export default DependencyGraphModal; 