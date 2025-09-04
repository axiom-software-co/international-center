// Local Development Environment Configuration
// Lazy evaluation for dynamic URL switching between build-time and client-side

function getBaseUrl(containerName: string, port: number): string {
  // Evaluate at access time, not import time
  // For Aspire local development, use localhost for both build-time and client-side
  return `http://localhost:${port}`;
}

function getPublicGatewayUrl(): string {
  // Public Gateway endpoint for Services API access with security and rate limiting
  return 'http://localhost:7220';
}

function getAssetsUrl(): string {
  // Evaluate at access time, not import time
  // For Aspire local development, use localhost
  return 'http://localhost:8099';
}

export const localConfig = {
  environment: 'local' as const,

  // Use getters for lazy evaluation
  get domains() {
    return {
      services: {
        baseUrl: getPublicGatewayUrl(), // Route through Public Gateway for security and rate limiting
        timeout: 5000,
        retryAttempts: 2,
      },
      news: {
        baseUrl: getPublicGatewayUrl(), // Route through Public Gateway for security and rate limiting
        timeout: 5000,
        retryAttempts: 2,
      },
      research: {
        baseUrl: getPublicGatewayUrl(), // Route through Public Gateway for security and rate limiting
        timeout: 5000,
        retryAttempts: 2,
      },
      contacts: {
        baseUrl: getBaseUrl('contacts-domain', 8084),
        timeout: 5000,
        retryAttempts: 2,
      },
      events: {
        baseUrl: getPublicGatewayUrl(), // Route through Public Gateway for security and rate limiting
        timeout: 5000,
        retryAttempts: 2,
      },
      newsletter: {
        baseUrl: getBaseUrl('newsletter-domain', 8086),
        timeout: 5000,
        retryAttempts: 2,
      },
      search: {
        baseUrl: getBaseUrl('search-domain', 8087),
        timeout: 5000,
        retryAttempts: 2,
      },
      assets: {
        baseUrl: getAssetsUrl(),
        timeout: 5000,
        retryAttempts: 2,
      },
    };
  },

  features: {
    enableCaching: true,
    enableRetry: true,
    enableCircuitBreaker: false, // Disabled for local dev
    enableErrorReporting: false, // Disabled for local dev
    cacheTimeout: 300000, // 5 minutes
  },

  performance: {
    requestTimeout: 5000,
    maxConcurrentRequests: 10,
    retryDelay: 1000,
    circuitBreakerThreshold: 5,
  },
};
