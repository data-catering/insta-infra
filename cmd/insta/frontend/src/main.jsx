import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import './styles.css';
import { ApiProvider } from './contexts/ApiContext';

// Import both API clients
import * as realApiClient from './api/client.js';
import * as demoApiClient from './api/demo-client.js';

// Choose API client based on build mode
const apiClient = process.env.DEMO_MODE === 'true' ? demoApiClient : realApiClient;

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <ApiProvider apiClient={apiClient}>
      <App />
    </ApiProvider>
  </React.StrictMode>
);