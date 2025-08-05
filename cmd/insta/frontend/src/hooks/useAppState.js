import { useState } from 'react';

export const useAppState = () => {
  // Simple state management
  const [services, setServices] = useState([]);
  const [statuses, setStatuses] = useState({});
  const [runningServices, setRunningServices] = useState([]);
  const [dependencyStatuses, setDependencyStatuses] = useState({});
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);
  const [lastUpdated, setLastUpdated] = useState(new Date());
  const [isAnimating, setIsAnimating] = useState(false);
  const [isStoppingAll, setIsStoppingAll] = useState(false);
  const [showAbout, setShowAbout] = useState(false);
  const [showConnectionModal, setShowConnectionModal] = useState(false);
  const [showLogsModal, setShowLogsModal] = useState(false);
  const [selectedService, setSelectedService] = useState(null);
  const [copyFeedback, setCopyFeedback] = useState('');
  const [runtimeStatus, setRuntimeStatus] = useState(null);
  const [showRuntimeSetup, setShowRuntimeSetup] = useState(false);
  const [currentRuntime, setCurrentRuntime] = useState('');
  const [showLogsPanel, setShowLogsPanel] = useState(false);
  const [showActionsDropdown, setShowActionsDropdown] = useState(false);
  const [isShuttingDown, setIsShuttingDown] = useState(false);

  return {
    // State
    services,
    statuses,
    runningServices,
    dependencyStatuses,
    isLoading,
    error,
    lastUpdated,
    isAnimating,
    isStoppingAll,
    showAbout,
    showConnectionModal,
    showLogsModal,
    selectedService,
    copyFeedback,
    runtimeStatus,
    showRuntimeSetup,
    currentRuntime,
    showLogsPanel,
    showActionsDropdown,
    isShuttingDown,

    // Setters
    setServices,
    setStatuses,
    setRunningServices,
    setDependencyStatuses,
    setIsLoading,
    setError,
    setLastUpdated,
    setIsAnimating,
    setIsStoppingAll,
    setShowAbout,
    setShowConnectionModal,
    setShowLogsModal,
    setSelectedService,
    setCopyFeedback,
    setRuntimeStatus,
    setShowRuntimeSetup,
    setCurrentRuntime,
    setShowLogsPanel,
    setShowActionsDropdown,
    setIsShuttingDown
  };
};