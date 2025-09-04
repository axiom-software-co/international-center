/**
 * Vitest Test Setup
 * Global configuration for unit tests
 */

import '@testing-library/jest-dom';
import { vi } from 'vitest';

// Mock fetch globally
global.fetch = vi.fn();

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

// Suppress console.log in tests unless explicitly needed
const originalConsoleLog = console.log;
const originalConsoleWarn = console.warn;
const originalConsoleError = console.error;

beforeEach(() => {
  // Reset all mocks before each test
  vi.clearAllMocks();

  // Reset fetch mock if it exists
  if (global.fetch && typeof (global.fetch as any).mockClear === 'function') {
    (global.fetch as any).mockClear();
  }
});

// Clean up after each test
afterEach(() => {
  vi.clearAllTimers();
});

// Global test utilities
export const createMockFetchResponse = (data: any, ok = true, status = 200) => {
  const response = {
    ok,
    status,
    statusText: ok ? 'OK' : 'Error',
    json: () => Promise.resolve(data),
    text: () => Promise.resolve(JSON.stringify(data)),
    headers: {
      get: (name: string) => {
        if (name === 'content-type') return 'application/json';
        return null;
      }
    }
  };
  
  return Promise.resolve(response as Response);
};

// Mock API responses for testing
export const mockNewsArticles = [
  {
    id: 'news1',
    title: 'Test News Article 1',
    slug: 'test-news-1',
    excerpt: 'Test excerpt 1',
    category: 'Company News',
    author: 'Test Author',
    published_at: '2025-08-17T12:00:00Z',
    tags: ['test', 'news'],
  },
  {
    id: 'news2',
    title: 'Test News Article 2',
    slug: 'test-news-2',
    excerpt: 'Test excerpt 2',
    category: 'Medical News',
    author: 'Test Author 2',
    published_at: '2025-08-16T12:00:00Z',
    tags: ['test', 'medical'],
  },
];

export const mockCaseStudies = [
  {
    id: 'case1',
    title: 'Test Case Study 1',
    slug: 'test-case-1',
    challenge: 'Test challenge 1',
    category: 'Clinical Research',
    published_at: '2025-08-17T12:00:00Z',
    tags: ['test', 'research'],
  },
  {
    id: 'case2',
    title: 'Test Case Study 2',
    slug: 'test-case-2',
    challenge: 'Test challenge 2',
    category: 'Clinical Studies',
    published_at: '2025-08-16T12:00:00Z',
    tags: ['test', 'studies'],
  },
];

export const mockApiResponse = (entries: any[], total?: number, page = 1, pageSize = 10) => ({
  entries,
  count: entries.length,
  total: total || entries.length,
  page,
  pageSize,
  totalPages: Math.ceil((total || entries.length) / pageSize),
});
