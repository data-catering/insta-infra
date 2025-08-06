import React, { createContext, useContext } from 'react';

// Create API client context
const ApiContext = createContext(null);

// Provider component
export const ApiProvider = ({ children, apiClient }) => {
  return (
    <ApiContext.Provider value={apiClient}>
      {children}
    </ApiContext.Provider>
  );
};

// Hook to use API client in components
export const useApiClient = () => {
  const apiClient = useContext(ApiContext);
  if (!apiClient) {
    throw new Error('useApiClient must be used within an ApiProvider');
  }
  return apiClient;
};