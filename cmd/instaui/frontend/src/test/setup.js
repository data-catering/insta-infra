import { expect, afterEach } from 'vitest';
import { cleanup } from '@testing-library/react';
import * as matchers from '@testing-library/jest-dom/matchers';

// Extend Vitest's expect with jest-dom matchers
expect.extend(matchers);

// Cleanup after each test case (e.g. clearing jsdom)
afterEach(() => {
  cleanup();
});

// Mock Wails runtime functions for testing
global.window = global.window || {};
global.window.go = global.window.go || {};
global.window.runtime = global.window.runtime || {};

// Mock the Wails backend functions
const mockWailsFunctions = {
  ListServices: () => Promise.resolve([]),
  GetServiceStatus: () => Promise.resolve('stopped'),
  StartService: () => Promise.resolve(),
  StopService: () => Promise.resolve(),
  GetServiceConnectionInfo: () => Promise.resolve({}),
  GetServiceLogs: () => Promise.resolve([]),
  CheckImageExists: () => Promise.resolve(true),
  GetRuntimeStatus: () => Promise.resolve({ canProceed: true }),
  GetCurrentRuntime: () => Promise.resolve('docker'),
  GetDependencyStatus: () => Promise.resolve({ dependencies: [], allDependenciesReady: true, runningCount: 0, requiredCount: 0, errorCount: 0 }),
  GetImageInfo: () => Promise.resolve('test-image:latest'),
  StartServiceWithStatusUpdate: () => Promise.resolve({}),
  StopAllServices: () => Promise.resolve(),
  GetAllRunningServices: () => Promise.resolve([]),
  GetAllServicesWithStatusAndDependencies: () => Promise.resolve([]),
  OpenServiceInBrowser: () => Promise.resolve(),
  StartImagePull: () => Promise.resolve(),
  GetImagePullProgress: () => Promise.resolve({ status: 'idle', progress: 0 }),
  StopImagePull: () => Promise.resolve(),
  StartLogStream: () => Promise.resolve(),
  StopLogStream: () => Promise.resolve(),
  GetDependencyGraph: () => Promise.resolve({ nodes: [], edges: [] }),
  GetServiceDependencyGraph: () => Promise.resolve({ nodes: [], edges: [] }),
  StopDependencyChain: () => Promise.resolve(),
  GetInstaInfraVersion: () => Promise.resolve('v2.1.0'),
  AttemptStartRuntime: () => Promise.resolve({ success: true }),
  WaitForRuntimeReady: () => Promise.resolve({ success: true }),
  ReinitializeRuntime: () => Promise.resolve(),
};

// Set up mock functions
global.window.go.main = global.window.go.main || {};
global.window.go.main.App = mockWailsFunctions;

// Mock runtime functions
global.window.runtime.EventsOn = () => {};
global.window.runtime.EventsOff = () => {};
global.window.runtime.EventsOnMultiple = () => {};
global.window.runtime.LogPrint = () => {};
global.window.runtime.LogInfo = () => {};
global.window.runtime.LogWarning = () => {};
global.window.runtime.LogError = () => {}; 