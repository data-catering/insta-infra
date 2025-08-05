import React, { useState, useEffect } from 'react';
import { uploadCustomService, updateCustomService, validateCustomCompose } from '../api/client';

function CustomServiceModal({ 
  isOpen, 
  onClose, 
  onServiceAdded,
  onServiceUpdated,
  editingService = null,
  isEditing = false 
}) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [content, setContent] = useState('');
  const [uploadMode, setUploadMode] = useState('text'); // 'text' or 'file'
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [validationResult, setValidationResult] = useState(null);
  const [isValidating, setIsValidating] = useState(false);
  const [errors, setErrors] = useState({});

  // Sample compose template
  const sampleCompose = `services:
  my-service:
    image: nginx:latest
    ports:
      - "8080:80"
    environment:
      - ENV=development
    depends_on:
      - postgres  # This will use the built-in postgres service
    volumes:
      - ./data:/usr/share/nginx/html
    restart: unless-stopped`;

  // Reset form when modal opens/closes or editing service changes
  useEffect(() => {
    if (isOpen) {
      if (isEditing && editingService) {
        setName(editingService.name || '');
        setDescription(editingService.description || '');
        setContent(editingService.content || '');
      } else {
        setName('');
        setDescription('');
        setContent('');
      }
      setValidationResult(null);
      setErrors({});
      setUploadMode('text');
    }
  }, [isOpen, isEditing, editingService]);

  // Validate content in real-time
  useEffect(() => {
    const validateContent = async () => {
      if (content.trim() && content.includes('services:')) {
        setIsValidating(true);
        try {
          const result = await validateCustomCompose(content);
          setValidationResult(result);
        } catch (error) {
          setValidationResult({
            valid: false,
            errors: [error.message]
          });
        } finally {
          setIsValidating(false);
        }
      } else {
        setValidationResult(null);
      }
    };

    const debounceTimer = setTimeout(validateContent, 500);
    return () => clearTimeout(debounceTimer);
  }, [content]);

  const handleFileUpload = (event) => {
    const file = event.target.files[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => {
        setContent(e.target.result);
        // Auto-fill name from filename if not set
        if (!name) {
          const fileName = file.name.replace(/\.(ya?ml)$/i, '');
          setName(fileName);
        }
      };
      reader.readAsText(file);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    // Basic validation
    const newErrors = {};
    if (!name.trim()) newErrors.name = 'Name is required';
    if (!content.trim()) newErrors.content = 'Compose content is required';
    
    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      return;
    }

    // Check if content is valid
    if (validationResult && !validationResult.valid) {
      setErrors({ content: 'Please fix validation errors before submitting' });
      return;
    }

    setIsSubmitting(true);
    setErrors({});

    try {
      if (isEditing && editingService) {
        const result = await updateCustomService(editingService.id, name.trim(), description.trim(), content);
        onServiceUpdated && onServiceUpdated(result);
      } else {
        const result = await uploadCustomService(name.trim(), description.trim(), content);
        onServiceAdded && onServiceAdded(result);
      }
      
      // Reset form and close modal
      setName('');
      setDescription('');
      setContent('');
      setValidationResult(null);
      onClose();
    } catch (error) {
      setErrors({ 
        submit: error.message || 'Failed to save custom service' 
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const insertSample = () => {
    setContent(sampleCompose);
    if (!name) {
      setName('my-custom-service');
    }
    if (!description) {
      setDescription('Custom service with dependencies on built-in services');
    }
  };

  if (!isOpen) return null;

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-container large" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2 className="modal-title">
            <div className="modal-icon" style={{ backgroundColor: 'rgba(16, 185, 129, 0.2)' }}>
              <svg width="20" height="20" className="icon-green" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M12 4.5V19.5M19.5 12H4.5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            </div>
            {isEditing ? 'Edit Custom Service' : 'Add Custom Service'}
          </h2>
          <button onClick={onClose} className="modal-close" aria-label="Close modal">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M6 18L18 6M6 6L18 18" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </button>
        </div>

        <form onSubmit={handleSubmit} className="modal-body">
          {/* Service Name */}
          <div className="form-group">
            <label htmlFor="service-name" className="form-label">
              Service Name *
            </label>
            <input
              id="service-name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className={`form-input ${errors.name ? 'error' : ''}`}
              placeholder="e.g., my-custom-app"
              disabled={isSubmitting}
            />
            {errors.name && <span className="error-text">{errors.name}</span>}
          </div>

          {/* Service Description */}
          <div className="form-group">
            <label htmlFor="service-description" className="form-label">
              Description
            </label>
            <input
              id="service-description"
              type="text"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="form-input"
              placeholder="Brief description of your custom service"
              disabled={isSubmitting}
            />
          </div>

          {/* Upload Mode Toggle */}
          <div className="form-group">
            <label className="form-label">Compose Content *</label>
            <div className="upload-mode-toggle">
              <button
                type="button"
                onClick={() => setUploadMode('text')}
                className={`toggle-button ${uploadMode === 'text' ? 'active' : ''}`}
                disabled={isSubmitting}
              >
                <svg width="16" height="16" className="toggle-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M3 7H21M3 12H21M3 17H21" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
                Text Editor
              </button>
              <button
                type="button"
                onClick={() => setUploadMode('file')}
                className={`toggle-button ${uploadMode === 'file' ? 'active' : ''}`}
                disabled={isSubmitting}
              >
                <svg width="16" height="16" className="toggle-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M14 2H6C4.89543 2 4 2.89543 4 4V20C4 21.1046 4.89543 22 6 22H18C19.1046 22 20 21.1046 20 20V8L14 2Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  <path d="M14 2V8H20" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
                File Upload
              </button>
            </div>
          </div>

          {/* File Upload Mode */}
          {uploadMode === 'file' && (
            <div className="form-group">
              <div className="file-upload-area">
                <input
                  type="file"
                  accept=".yml,.yaml"
                  onChange={handleFileUpload}
                  className="file-input"
                  id="compose-file"
                  disabled={isSubmitting}
                />
                <label htmlFor="compose-file" className="file-upload-label">
                  <svg width="48" height="48" className="upload-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M21 15V19C21 20.1046 20.1046 21 19 21H5C3.89543 21 3 20.1046 3 19V15M17 8L12 3M12 3L7 8M12 3V15" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                  <span className="upload-text">
                    Click to upload a compose file
                  </span>
                  <span className="upload-subtext">
                    Supports .yml and .yaml files
                  </span>
                </label>
              </div>
            </div>
          )}

          {/* Text Editor Mode */}
          {uploadMode === 'text' && (
            <div className="form-group">
              <div className="text-editor-header">
                <button
                  type="button"
                  onClick={insertSample}
                  className="button button-secondary small"
                  disabled={isSubmitting}
                >
                  <svg width="16" height="16" className="button-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M12 4.5V19.5M19.5 12H4.5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                  Insert Sample
                </button>
              </div>
              <textarea
                value={content}
                onChange={(e) => setContent(e.target.value)}
                className={`form-textarea compose-editor ${errors.content ? 'error' : ''}`}
                placeholder="Paste your docker-compose.yml content here..."
                rows={15}
                disabled={isSubmitting}
              />
              {errors.content && <span className="error-text">{errors.content}</span>}
            </div>
          )}

          {/* Validation Status */}
          {(isValidating || validationResult) && (
            <div className="validation-status">
              {isValidating ? (
                <div className="validation-loading">
                  <div className="loading-spinner small"></div>
                  <span>Validating compose file...</span>
                </div>
              ) : validationResult ? (
                <div className={`validation-result ${validationResult.valid ? 'valid' : 'invalid'}`}>
                  <div className="validation-icon">
                    {validationResult.valid ? (
                      <svg width="16" height="16" className="icon-green" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path d="M20 6L9 17L4 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                      </svg>
                    ) : (
                      <svg width="16" height="16" className="icon-red" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path d="M6 18L18 6M6 6L18 18" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                      </svg>
                    )}
                  </div>
                  <div className="validation-details">
                    <span className="validation-text">
                      {validationResult.valid ? 'Valid compose file' : 'Invalid compose file'}
                    </span>
                    {validationResult.errors && validationResult.errors.length > 0 && (
                      <ul className="validation-errors">
                        {validationResult.errors.map((error, index) => (
                          <li key={index}>{error}</li>
                        ))}
                      </ul>
                    )}
                    {validationResult.warnings && validationResult.warnings.length > 0 && (
                      <ul className="validation-warnings">
                        {validationResult.warnings.map((warning, index) => (
                          <li key={index}>{warning}</li>
                        ))}
                      </ul>
                    )}
                  </div>
                </div>
              ) : null}
            </div>
          )}

          {/* Submit Error */}
          {errors.submit && (
            <div className="error-banner">
              <svg width="16" height="16" className="error-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M12 8V12M12 16H12.01M22 12C22 17.5228 17.5228 22 12 22C6.47715 22 2 17.5228 2 12C2 6.47715 6.47715 2 12 2C17.5228 2 22 6.47715 22 12Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
              {errors.submit}
            </div>
          )}

          {/* Action Buttons */}
          <div className="modal-actions">
            <button
              type="button"
              onClick={onClose}
              className="button button-secondary"
              disabled={isSubmitting}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="button button-primary"
              disabled={isSubmitting || (validationResult && !validationResult.valid)}
            >
              {isSubmitting ? (
                <>
                  <div className="loading-spinner small"></div>
                  {isEditing ? 'Updating...' : 'Adding...'}
                </>
              ) : (
                <>
                  <svg width="16" height="16" className="button-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M20 6L9 17L4 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                  {isEditing ? 'Update Service' : 'Add Service'}
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default CustomServiceModal;