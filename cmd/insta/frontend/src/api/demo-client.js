// Demo API client that returns mock data for GitHub Pages demo without making actual API calls

// In demo mode, we don't make any actual API calls
// Instead, we return mock data directly from the functions
const MOCK_SERVICES = [
  {
    Name: 'postgres',
    Type: 'database',
    Dependencies: []
  },
  {
    Name: 'redis',
    Type: 'cache',
    Dependencies: []
  },
  {
    Name: 'airflow',
    Type: 'workflow',
    Dependencies: ['postgres', 'redis']
  },
  {
    Name: 'jupyter',
    Type: 'notebook',
    Dependencies: []
  },
  {
    Name: 'mlflow',
    Type: 'mlops',
    Dependencies: ['postgres']
  }
];

const MOCK_STATUSES = {
  'postgres': 'running',
  'redis': 'running',
  'airflow': 'stopped',
  'jupyter': 'running',
  'mlflow': 'stopped'
};

const MOCK_RUNNING_SERVICES = {
  running_services: {
    'postgres': { Status: 'running' },
    'redis': { Status: 'running' },
    'jupyter': { Status: 'running' }
  }
};

export async function listServices() {
  return MOCK_SERVICES;
}

export async function getAllServiceStatuses() {
  return MOCK_STATUSES;
}

export async function startService(serviceName) {
  MOCK_STATUSES[serviceName] = 'running';
  MOCK_RUNNING_SERVICES.running_services[serviceName] = { Status: 'running' };
  return { status: 'success' };
}

export async function stopService(serviceName) {
  MOCK_STATUSES[serviceName] = 'stopped';
  delete MOCK_RUNNING_SERVICES.running_services[serviceName];
  return { status: 'success' };
}

export async function stopAllServices() {
  Object.keys(MOCK_STATUSES).forEach(service => {
    MOCK_STATUSES[service] = 'stopped';
  });
  MOCK_RUNNING_SERVICES.running_services = {};
  return { status: 'success' };
}

export async function getRunningServices() {
  return MOCK_RUNNING_SERVICES;
}

export async function getRuntimeStatus() {
  return {
    platform: 'demo',
    hasAnyRuntime: true,
    preferredRuntime: 'docker',
    availableRuntimes: ['docker'],
    canProceed: true,
    recommendedAction: 'Ready to use docker',
    runtimeStatuses: [
      {
        name: 'docker',
        isInstalled: true,
        isRunning: true,
        isAvailable: true,
        hasCompose: true,
        version: '24.0.0-demo',
        error: '',
        canAutoStart: true,
        installationGuide: 'Demo mode - no installation needed',
        startupCommand: '',
        requiresMachine: false,
        machineStatus: ''
      },
      {
        name: 'podman',
        isInstalled: false,
        isRunning: false,
        isAvailable: false,
        hasCompose: false,
        version: '',
        error: 'Podman not installed',
        canAutoStart: false,
        installationGuide: 'Demo mode - no installation needed',
        startupCommand: '',
        requiresMachine: false,
        machineStatus: ''
      }
    ]
  };
}

export async function getCurrentRuntime() {
  return {
    runtime: 'demo'
  };
}

export async function getAllImageStatuses() {
  return {
    image_statuses: {
      postgres: {
        status: 'ready',
        exists: true,
        image_name: 'postgres'
      },
      redis: {
        status: 'ready',
        exists: true,
        image_name: 'redis'
      },
      airflow: {
        status: 'ready',
        exists: true,
        image_name: 'airflow'
      },
      jupyter: {
        status: 'ready',
        exists: true,
        image_name: 'jupyter'
      },
      mlflow: {
        status: 'ready',
        exists: true,
        image_name: 'mlflow'
      }
    },
    total_services: 5
  };
}

export async function getCustomCompose() {
  return {
    services: {}  // Empty object for demo mode
  };
}

export async function getServiceLogs() {
  return {
    logs: [
      { timestamp: '2024-03-05T12:00:01Z', message: 'Service started successfully' },
      { timestamp: '2024-03-05T12:00:02Z', message: 'Initializing database...' },
      { timestamp: '2024-03-05T12:00:03Z', message: 'Database ready to accept connections' },
      { timestamp: '2024-03-05T12:00:04Z', message: 'Service is healthy' }
    ]
  };
}

export async function getAppLogs() {
  return {
    logs: [
      { timestamp: '2024-03-05T12:00:00Z', message: 'InstaInfra Demo Mode' },
      { timestamp: '2024-03-05T12:00:01Z', message: 'Services initialized' },
      { timestamp: '2024-03-05T12:00:02Z', message: 'WebSocket connected' },
      { timestamp: '2024-03-05T12:00:03Z', message: 'Ready to accept commands' }
    ]
  };
}

export async function getServiceConnection(serviceName) {
  const connections = {
    postgres: {
      host: 'localhost',
      port: '5432',
      username: 'postgres',
      password: 'postgres',
      url: 'postgresql://postgres:postgres@localhost:5432/postgres'
    },
    redis: {
      host: 'localhost',
      port: '6379',
      url: 'redis://localhost:6379'
    },
    jupyter: {
      host: 'localhost',
      port: '8888',
      url: 'http://localhost:8888'
    },
    mlflow: {
      host: 'localhost',
      port: '5000',
      url: 'http://localhost:5000'
    },
    airflow: {
      host: 'localhost',
      port: '8080',
      username: 'airflow',
      password: 'airflow',
      url: 'http://localhost:8080'
    }
  };
  return connections[serviceName] || { host: 'localhost', port: '0', url: 'http://localhost' };
}

export function openURL(url) {
  // In demo mode, just log the URL that would be opened
  console.log('Demo mode: Would open URL:', url);
}

// Mock WebSocket client for demo
export class WebSocketClient {
  constructor() {}
  connect() {}
  disconnect() {}
  subscribe() {}
  unsubscribe() {}
  subscribeToServiceStatus() { 
    // Return a mock unsubscribe function
    return () => {};
  }
  getState() { return 'CONNECTED'; }
}

// Additional missing functions that components use
export async function getApiInfo() {
  return {
    version: 'v1.0.0-demo',
    build_time: '2024-03-05T12:00:00Z',
    go_version: 'go1.21.0',
    git_commit: 'demo-commit'
  };
}

export async function refreshServiceStatus() {
  return MOCK_STATUSES;
}

export async function listCustomServices() {
  return []; // Empty array for demo - matches what frontend expects
}

export async function getCustomService(serviceId) {
  return { name: 'demo-service', description: 'Demo custom service', content: '' };
}

export async function uploadCustomService(name, description, content) {
  return { status: 'success', id: 'demo-service-id' };
}

export async function updateCustomService(serviceId, name, description, content) {
  return { status: 'success' };
}

export async function deleteCustomService(serviceId) {
  return { status: 'success' };
}

export async function validateCustomCompose(content) {
  return { valid: true, errors: [] };
}

export async function getCustomServiceStats() {
  return { total: 0, active: 0 };
}

export async function checkImageExists(imageName) {
  return { exists: true };
}

export async function startImagePull(imageName) {
  return { status: 'success' };
}

export async function stopImagePull(imageName) {
  return { status: 'success' };
}

export async function getImagePullProgress(imageName) {
  return { progress: 100, status: 'complete' };
}

export async function startServiceLogStream(serviceName) {
  return { status: 'success' };
}

export async function stopServiceLogStream(serviceName) {
  return { status: 'success' };
}

export async function getActiveLogStreams() {
  return { streams: [] };
}

export async function getAppLogEntries() {
  return {
    logs: [
      { timestamp: '2024-03-05T12:00:00Z', level: 'info', message: 'Demo mode started' },
      { timestamp: '2024-03-05T12:00:01Z', level: 'info', message: 'Services initialized' }
    ]
  };
}

export async function getAppLogsSince(timestamp) {
  return {
    logs: [
      { timestamp: '2024-03-05T12:00:02Z', level: 'info', message: 'Demo logs since ' + timestamp }
    ]
  };
}

export async function startRuntime(runtime) {
  return { status: 'success' };
}

export async function restartRuntime() {
  return { status: 'success' };
}

export async function reinitializeRuntime() {
  return { status: 'success' };
}

export async function setCustomDockerPath(path) {
  return { status: 'success' };
}

export async function setCustomPodmanPath(path) {
  return { status: 'success' };
}

export async function getCustomDockerPath() {
  return { path: '/usr/bin/docker' };
}

export async function getCustomPodmanPath() {
  return { path: '/usr/bin/podman' };
}

export async function getHealth() {
  return { status: 'healthy' };
}

export async function shutdownApplication() {
  return { status: 'success' };
}

export async function openServiceConnection(serviceName) {
  console.log('Demo mode: Would open connection for', serviceName);
  return { status: 'success' };
}

export async function AttemptStartRuntime(runtime) {
  return { status: 'success' };
}

export async function WaitForRuntimeReady(runtimeName, maxWaitSeconds) {
  return { status: 'ready' };
}

export async function ReinitializeRuntime() {
  return { status: 'success' };
}

export async function SetCustomDockerPath(path) {
  return { status: 'success' };
}

export async function SetCustomPodmanPath(path) {
  return { status: 'success' };
}

export const wsClient = new WebSocketClient();
export const WS_MSG_TYPES = {
  SERVICE_STATUS_UPDATE: 'service_status_update',
  SERVICE_STARTED: 'service_started',
  SERVICE_STOPPED: 'service_stopped',
  SERVICE_ERROR: 'service_error',
  SERVICE_LOGS: 'service_logs',
  APP_LOGS: 'app_logs',
  LOG_STREAM: 'log_stream',
  IMAGE_PULL_PROGRESS: 'image_pull_progress',
  IMAGE_PULL_COMPLETE: 'image_pull_complete',
  IMAGE_PULL_ERROR: 'image_pull_error',
  RUNTIME_STATUS_UPDATE: 'runtime_status_update',
  CLIENT_CONNECTED: 'client_connected',
  CLIENT_DISCONNECTED: 'client_disconnected',
  PING: 'ping',
  PONG: 'pong'
};