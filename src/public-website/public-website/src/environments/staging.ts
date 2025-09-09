// Staging Environment Configuration
// TODO: Update Services API to route through Staging Public Gateway when deployed
// Connection to GCP Cloud Functions staging endpoints

export const stagingConfig = {
  environment: 'staging' as const,

  domains: {
    services: {
      baseUrl: 'https://us-central1-international-center-staging.cloudfunctions.net/services',
      timeout: 10000,
      retryAttempts: 3,
    },
    news: {
      baseUrl: 'https://us-central1-international-center-staging.cloudfunctions.net/news',
      timeout: 10000,
      retryAttempts: 3,
    },
    research: {
      baseUrl: 'https://us-central1-international-center-staging.cloudfunctions.net/research',
      timeout: 10000,
      retryAttempts: 3,
    },
    contacts: {
      baseUrl: 'https://us-central1-international-center-staging.cloudfunctions.net/contacts',
      timeout: 10000,
      retryAttempts: 3,
    },
    search: {
      baseUrl: 'https://us-central1-international-center-staging.cloudfunctions.net/search',
      timeout: 8000,
      retryAttempts: 2,
    },
    newsletter: {
      baseUrl: 'https://us-central1-international-center-staging.cloudfunctions.net/newsletter',
      timeout: 10000,
      retryAttempts: 3,
    },
    events: {
      baseUrl: 'https://us-central1-international-center-staging.cloudfunctions.net/events',
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
