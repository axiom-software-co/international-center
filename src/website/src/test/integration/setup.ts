/**
 * Integration Test Setup
 * Minimal Node.js environment setup for integration tests
 */

import { vi } from 'vitest';

// Store original fetch for potential restoration
const originalFetch = global.fetch;

// Don't mock fetch for integration tests - we want real HTTP requests
if (!global.fetch) {
  global.fetch = require('node-fetch');
}

// Mock import.meta.env for integration tests
Object.defineProperty(import.meta, 'env', {
  value: {
    PUBLIC_CONTENT_API_URL: 'http://127.0.0.1:9001',
    MODE: 'integration',
    TEST_INTEGRATION: true,
  },
  configurable: true,
});

// Clean up any existing Pinia instances
vi.clearAllMocks();

export { originalFetch };