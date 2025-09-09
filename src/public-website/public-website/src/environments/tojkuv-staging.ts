// Tojkuv Staging Environment Configuration
// TODO: Update Services API to route through Tojkuv Staging Public Gateway when deployed
// Connection to GCP Cloud Functions tojkuv-staging endpoints in international-center-2 project

export const tojkuvStagingConfig = {
  environment: 'tojkuv-staging' as const,

  domains: {
    services: {
      baseUrl: 'https://us-central1-international-center-2.cloudfunctions.net/tojkuv-staging-services',
      timeout: 10000,
      retryAttempts: 3,
    },
    news: {
      baseUrl: 'https://us-central1-international-center-2.cloudfunctions.net/tojkuv-staging-news',
      timeout: 10000,
      retryAttempts: 3,
    },
    research: {
      baseUrl: 'https://us-central1-international-center-2.cloudfunctions.net/tojkuv-staging-research',
      timeout: 10000,
      retryAttempts: 3,
    },
    contacts: {
      baseUrl: 'https://us-central1-international-center-2.cloudfunctions.net/tojkuv-staging-contacts',
      timeout: 10000,
      retryAttempts: 3,
    },
    search: {
      baseUrl: 'https://us-central1-international-center-2.cloudfunctions.net/tojkuv-staging-search',
      timeout: 8000,
      retryAttempts: 2,
    },
    newsletter: {
      baseUrl: 'https://us-central1-international-center-2.cloudfunctions.net/tojkuv-staging-newsletter',
      timeout: 10000,
      retryAttempts: 3,
    },
    events: {
      baseUrl: 'https://us-central1-international-center-2.cloudfunctions.net/tojkuv-staging-events',
      timeout: 10000,
      retryAttempts: 3,
    },
  },

  features: {
    enableCaching: true,
    enableRetry: true,
    enableCircuitBreaker: true,
    enableErrorReporting: true,
    cacheTimeout: 600000, // 10 minutes
  },

  performance: {
    requestTimeout: 10000,
    maxConcurrentRequests: 20,
    retryDelay: 2000,
    circuitBreakerThreshold: 3,
  },
};