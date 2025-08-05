import React, { useState, useEffect } from 'react';
import { X, AlertCircle, AlertTriangle, Info, CheckCircle } from 'lucide-react';

// Centralized Error Message component with consistent styling
const ErrorMessage = ({ 
  type = 'error', 
  title, 
  message, 
  details, 
  onDismiss, 
  autoHide = false,
  autoHideDelay = 5000,
  actions = [],
  className = '',
  metadata = {}
}) => {
  const [isVisible, setIsVisible] = useState(true);
  const [detailsViewed, setDetailsViewed] = useState(false);

  // Auto-hide functionality - but don't auto-hide if details have been viewed
  useEffect(() => {
    if (autoHide && autoHideDelay > 0 && !detailsViewed) {
      const timer = setTimeout(() => {
        handleDismiss();
      }, autoHideDelay);
      return () => clearTimeout(timer);
    }
  }, [autoHide, autoHideDelay, detailsViewed]);

  const handleDismiss = () => {
    setIsVisible(false);
    if (onDismiss) {
      onDismiss();
    }
  };

  const handleDetailsToggle = (event) => {
    // Mark details as viewed when user clicks to show them
    if (event.target.open) {
      setDetailsViewed(true);
    }
  };

  if (!isVisible) return null;

  // Get type-specific styling and icons
  const getTypeConfig = () => {
    switch (type) {
      case 'error':
        return {
          containerClass: 'error-message-error',
          icon: <AlertCircle size={20} />,
          iconColor: 'text-red-400'
        };
      case 'warning':
        return {
          containerClass: 'error-message-warning',
          icon: <AlertTriangle size={20} />,
          iconColor: 'text-yellow-400'
        };
      case 'info':
        return {
          containerClass: 'error-message-info',
          icon: <Info size={20} />,
          iconColor: 'text-blue-400'
        };
      case 'success':
        return {
          containerClass: 'error-message-success',
          icon: <CheckCircle size={20} />,
          iconColor: 'text-green-400'
        };
      default:
        return {
          containerClass: 'error-message-error',
          icon: <AlertCircle size={20} />,
          iconColor: 'text-red-400'
        };
    }
  };

  const { containerClass, icon, iconColor } = getTypeConfig();

  return (
    <div className={`error-message ${containerClass} ${className}`} role="alert">
      <div className="error-message-content">
        <div className="error-message-header">
          <div className={`error-message-icon ${iconColor}`}>
            {icon}
          </div>
          {title && <div className="error-message-title">{title}</div>}
        </div>
        <div className="error-message-text">
          {message && <div className="error-message-description">{message}</div>}
          {details && (
            <details className="error-message-details" onToggle={handleDetailsToggle}>
              <summary>Show Details</summary>
              <div className="error-message-details-content">{details}</div>
            </details>
          )}
        </div>
        {(onDismiss || actions.length > 0) && (
          <div className="error-message-actions">
            {actions.map((action, index) => (
              <button
                key={index}
                onClick={action.onClick}
                className={`error-message-action ${action.variant || 'secondary'}`}
                disabled={action.disabled}
              >
                {action.icon && <span className="action-icon">{action.icon}</span>}
                {action.label}
              </button>
            ))}
            {onDismiss && (
              <button 
                onClick={handleDismiss}
                className="error-message-close"
                aria-label="Close error message"
              >
                <X size={16} />
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

// Toast notification for temporary error messages
export const ErrorToast = ({ 
  type = 'error', 
  message, 
  duration = 4000,
  onClose 
}) => {
  const [isVisible, setIsVisible] = useState(true);

  useEffect(() => {
    const timer = setTimeout(() => {
      setIsVisible(false);
      if (onClose) onClose();
    }, duration);

    return () => clearTimeout(timer);
  }, [duration, onClose]);

  if (!isVisible) return null;

  const getToastConfig = () => {
    switch (type) {
      case 'error':
        return { icon: <AlertCircle size={16} />, class: 'error-toast-error' };
      case 'warning':
        return { icon: <AlertTriangle size={16} />, class: 'error-toast-warning' };
      case 'success':
        return { icon: <CheckCircle size={16} />, class: 'error-toast-success' };
      case 'info':
      default:
        return { icon: <Info size={16} />, class: 'error-toast-info' };
    }
  };

  const { icon, class: toastClass } = getToastConfig();

  return (
    <div className={`error-toast ${toastClass}`}>
      <div className="error-toast-content">
        <div className="error-toast-icon">{icon}</div>
        <div className="error-toast-message">{message}</div>
        <button 
          onClick={() => {
            setIsVisible(false);
            if (onClose) onClose();
          }} 
          className="error-toast-close"
          aria-label="Close notification"
        >
          <X size={14} />
        </button>
      </div>
    </div>
  );
};

// Hook for managing error state with enhanced features
export const useErrorHandler = () => {
  const [errors, setErrors] = useState([]);
  const [toasts, setToasts] = useState([]);

  const addError = (error) => {
    const errorObj = {
      id: Date.now() + Math.random(),
      timestamp: new Date(),
      ...error
    };
    setErrors(prev => [...prev, errorObj]);
  };

  const addToast = (toast) => {
    const toastObj = {
      id: Date.now() + Math.random(),
      timestamp: new Date(),
      duration: 4000,
      ...toast
    };
    setToasts(prev => [...prev, toastObj]);
  };

  const removeError = (id) => {
    setErrors(prev => prev.filter(error => error.id !== id));
  };

  const removeToast = (id) => {
    setToasts(prev => prev.filter(toast => toast.id !== id));
  };

  const clearAllErrors = () => {
    setErrors([]);
  };

  const clearAllToasts = () => {
    setToasts([]);
  };

  const clearAll = () => {
    clearAllErrors();
    clearAllToasts();
  };

  return {
    errors,
    toasts,
    addError,
    addToast,
    removeError,
    removeToast,
    clearAllErrors,
    clearAllToasts,
    clearAll
  };
};

// Toast container component for managing multiple toasts
export const ToastContainer = ({ toasts, onRemoveToast }) => {
  if (!toasts || toasts.length === 0) return null;

  return (
    <div className="error-toast-container">
      {toasts.map(toast => (
        <ErrorToast
          key={toast.id}
          type={toast.type}
          message={toast.message}
          duration={toast.duration}
          onClose={() => onRemoveToast(toast.id)}
        />
      ))}
    </div>
  );
};

export default ErrorMessage; 