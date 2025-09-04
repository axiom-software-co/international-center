/**
 * Vitest Integration Test Setup
 * Real API integration testing configuration
 */

import '@testing-library/jest-dom';
import { beforeAll, afterAll, beforeEach } from 'vitest';

// Real API configuration for integration tests
const API_BASE_URL = 'http://localhost:8083';
const WEBSITE_URL = 'http://localhost:4321';

// Configure real environment variables for integration tests
Object.defineProperty(import.meta, 'env', {
  value: {
    PUBLIC_CONTENT_API_URL: API_BASE_URL,
    MODE: 'integration-test',
    INTEGRATION_TEST: true,
  },
  configurable: true,
});

// Configure window.location for browser context
Object.defineProperty(window, 'location', {
  value: {
    hostname: 'localhost',
    href: WEBSITE_URL,
    origin: WEBSITE_URL,
  },
  writable: true,
});

// API health check utilities
export const checkApiHealth = async (): Promise<boolean> => {
  try {
    const response = await fetch(`${API_BASE_URL}/api/content/entries?contentType=news&pageSize=1`);
    return response.ok;
  } catch (error) {
    console.error('API health check failed:', error);
    return false;
  }
};

export const waitForApi = async (maxAttempts = 10, delay = 1000): Promise<void> => {
  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    const isHealthy = await checkApiHealth();
    if (isHealthy) {
      console.log(`✅ API is ready (attempt ${attempt})`);
      return;
    }

    console.log(`⏳ Waiting for API... (attempt ${attempt}/${maxAttempts})`);
    if (attempt < maxAttempts) {
      await new Promise(resolve => setTimeout(resolve, delay));
    }
  }

  throw new Error(
    `❌ API not available after ${maxAttempts} attempts. Please ensure the development server is running.`
  );
};

// Real API data fetching utilities for test validation
export const fetchRealNewsArticles = async (
  page = 1,
  pageSize = 10,
  category?: string,
  sortBy?: string
) => {
  const params = new URLSearchParams();
  params.append('page', page.toString());
  params.append('pageSize', pageSize.toString());

  if (category && category !== 'All Categories') {
    params.append('category', category);
  }

  if (sortBy) {
    params.append('sortBy', sortBy);
  }

  const response = await fetch(
    `${API_BASE_URL}/api/content/entries?contentType=news&${params.toString()}`
  );
  if (!response.ok) {
    throw new Error(`API request failed: ${response.status}`);
  }

  return response.json();
};

export const fetchRealCaseStudies = async (
  page = 1,
  pageSize = 10,
  category?: string,
  sortBy?: string
) => {
  const params = new URLSearchParams();
  params.append('page', page.toString());
  params.append('pageSize', pageSize.toString());

  if (category && category !== 'All Categories') {
    params.append('category', category);
  }

  if (sortBy) {
    params.append('sortBy', sortBy);
  }

  const response = await fetch(
    `${API_BASE_URL}/api/content/entries?contentType=case-study&${params.toString()}`
  );
  if (!response.ok) {
    throw new Error(`API request failed: ${response.status}`);
  }

  return response.json();
};

export const fetchRealCategories = async (contentType: string) => {
  const response = await fetch(
    `${API_BASE_URL}/api/content/entries?contentType=${contentType === 'case-studies' ? 'case-study' : 'news'}&pageSize=50&page=1`
  );
  if (!response.ok) {
    throw new Error(`Categories API request failed: ${response.status}`);
  }

  const data = await response.json();
  const items = data.entries || [];
  const categories = items.map((item: any) => item.category).filter(Boolean);
  return [...new Set(categories)] as string[];
};

// Global test hooks
beforeAll(async () => {
  console.log('🧪 Starting Integration Tests - Real API Testing');
  console.log(`📡 API URL: ${API_BASE_URL}`);
  console.log(`🌐 Website URL: ${WEBSITE_URL}`);

  // Wait for API to be available
  await waitForApi();

  // Log API status
  const newsData = await fetchRealNewsArticles(1, 1);
  const caseStudyData = await fetchRealCaseStudies(1, 1);

  console.log(`📊 API Status:`);
  console.log(`  - News articles available: ${newsData.total || 0}`);
  console.log(`  - Case studies available: ${caseStudyData.total || 0}`);
}, 30000);

beforeEach(async () => {
  // Ensure API is still responsive before each test
  const isHealthy = await checkApiHealth();
  if (!isHealthy) {
    throw new Error('❌ API became unavailable during test run');
  }
});

afterAll(() => {
  console.log('🏁 Integration Tests Completed');
});

// Export configuration for tests
export const INTEGRATION_CONFIG = {
  API_BASE_URL,
  WEBSITE_URL,
  DEFAULT_TIMEOUT: 10000,
  API_TIMEOUT: 5000,
};
