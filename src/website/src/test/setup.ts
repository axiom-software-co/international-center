/**
 * Vitest Test Setup
 * Minimal global configuration for unit tests
 */

import '@testing-library/jest-dom';
import { vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';

// Store original implementations for cleanup
const originalFetch = global.fetch;

// Mock fetch globally 
export const mockFetch = vi.fn();
global.fetch = mockFetch as any;

// Mock import.meta.env
Object.defineProperty(import.meta, 'env', {
  value: {
    PUBLIC_CONTENT_API_URL: 'http://localhost:8083',
    MODE: 'test',
  },
  configurable: true,
});

// Mock window.location
Object.defineProperty(window, 'location', {
  value: {
    hostname: 'localhost',
    href: 'http://localhost:4321',
    origin: 'http://localhost:4321',
  },
  writable: true,
});

beforeEach(() => {
  // Initialize fresh Pinia instance for each test
  const pinia = createPinia();
  setActivePinia(pinia);
  
  // Reset mock fetch for each test
  mockFetch.mockReset();
});

afterEach(() => {
  // Simple cleanup
  vi.clearAllMocks();
});

// Global cleanup when test suite completes
afterAll(() => {
  // Restore original implementations
  if (originalFetch) {
    global.fetch = originalFetch;
  }
});
