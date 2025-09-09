// Production Environment Configuration
// TODO: Update Services API to route through Production Public Gateway when deployed
// Connection to GCP Cloud Functions production endpoints

export const productionConfig = {
  environment: 'production' as const,

  domains: {
    services: {
      baseUrl: 'https://us-central1-international-center-prod.cloudfunctions.net/services',
      timeout: 15000,
      retryAttempts: 3,
    },
    news: {
      baseUrl: 'https://us-central1-international-center-prod.cloudfunctions.net/news',
      timeout: 15000,
      retryAttempts: 3,
    },
    research: {
      baseUrl: 'https://us-central1-international-center-prod.cloudfunctions.net/research',
      timeout: 15000,
      retryAttempts: 3,
    },
    contacts: {
      baseUrl: 'https://us-central1-international-center-prod.cloudfunctions.net/contacts',
      timeout: 15000,
      retryAttempts: 3,
    },
    search: {
      baseUrl: 'https://us-central1-international-center-prod.cloudfunctions.net/search',
      timeout: 10000,
      retryAttempts: 2,
    },
    newsletter: {
      baseUrl: 'https://us-central1-international-center-prod.cloudfunctions.net/newsletter',
      timeout: 15000,
      retryAttempts: 3,
    },
    events: {
      baseUrl: 'https://us-central1-international-center-prod.cloudfunctions.net/events',
      timeout: 15000,
      retryAttempts: 3,
    },
  },

  features: {
    enableCaching: true,
    enableRetry: true,
    enableCircuitBreaker: true,
    enableErrorReporting: true,
    cacheTimeout: 1800000, // 30 minutes
  },

  performance: {
    requestTimeout: 15000,
    maxConcurrentRequests: 50,
    retryDelay: 3000,
    circuitBreakerThreshold: 2,
  },
};
