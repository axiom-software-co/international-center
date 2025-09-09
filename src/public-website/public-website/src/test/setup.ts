/**
 * Vitest Test Setup
 * Minimal global configuration for unit tests
 */

import '@testing-library/jest-dom';
import { vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';

// Store original implementations for cleanup
const originalFetch = global.fetch;

// Mock fetch globally with more explicit setup
export const mockFetch = vi.fn();

// Multiple approaches to ensure fetch is mocked
global.fetch = mockFetch as any;
globalThis.fetch = mockFetch as any;

// Also mock it using vitest's module mocking
vi.stubGlobal('fetch', mockFetch);

// Mock import.meta.env
Object.defineProperty(import.meta, 'env', {
  value: {
    PUBLIC_CONTENT_API_URL: 'http://localhost:7220',
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

// Global test helpers for standardized patterns
export const expectUrlWithoutBase = (url: string) => {
  // Remove base URL for consistent test expectations
  return url.replace(/^https?:\/\/[^\/]+/, '');
};

export const expectQueryWithPlusEncoding = (query: string) => {
  // Accept URLSearchParams + encoding instead of %20 encoding
  return query.replace(/%20/g, '+');
};

beforeEach(() => {
  // Initialize fresh Pinia instance for each test
  const pinia = createPinia();
  setActivePinia(pinia);
  
  // Reset mock fetch for each test
  mockFetch.mockReset();
  mockFetch.mockClear();
  
  // Ensure fetch mock is properly set up
  global.fetch = mockFetch as any;
  globalThis.fetch = mockFetch as any;
  vi.stubGlobal('fetch', mockFetch);
  
  // Clear any existing cache state by forcing garbage collection
  // This ensures fresh RestClientCache instances in each test
  if (global.gc) {
    global.gc();
  }
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
