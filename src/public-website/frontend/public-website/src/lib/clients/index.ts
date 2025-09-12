// Contract-Generated Clients - Main exports (Contract-first architecture)
// All manual client implementations replaced with contract-generated clients

// Environment configuration
export { config, isLocal, isStaging, isProduction } from '../../environments';
export type { EnvironmentConfig, Environment } from '../../environments';

// Contract-generated API client (replaces all manual clients)
export { apiClient } from '../api-client';

// Re-export all contract-generated types for convenience
export type * from '@international-center/public-api-client';

// Contract-based composables (replaces all manual composables)
export {
  useContractNews,
  useContractResearch,
  useContractServices,
  useContractEvents,
  useContractHealth,
  useContractInquiries,
  useContractApi
} from '../../composables/useContractApi';

// Contract-compliant error handling
export {
  ContractErrorHandler,
  ErrorType,
  createErrorHandler
} from '../error-handling';

// Legacy compatibility layer - provides clear migration path
export const newsClient = {
  getNewsCategories: () => apiClient.getNewsCategories(),
  getNewsArticles: (params: any) => apiClient.getNews(params),
  searchNewsArticles: (params: any) => apiClient.getNews({ search: params.q, page: params.page, limit: params.limit }),
  getFeaturedNews: () => apiClient.getFeaturedNews(),
  getNewsArticleBySlug: (slug: string) => apiClient.getNewsById(slug)
};

export const researchClient = {
  getResearchArticles: (params: any) => apiClient.getResearch(params),
  getResearchCategories: () => apiClient.getResearchCategories(),
  searchResearch: (params: any) => apiClient.getResearch({ search: params.q, page: params.page, limit: params.limit }),
  getFeaturedResearch: () => apiClient.getFeaturedResearch(),
  getResearchBySlug: (slug: string) => apiClient.getResearchById(slug)
};

export const eventsClient = {
  getEvents: (params: any) => apiClient.getEvents(params),
  getEventCategories: () => apiClient.getEventCategories(),
  searchEvents: (params: any) => apiClient.getEvents({ search: params.q, page: params.page, limit: params.limit }),
  getFeaturedEvents: () => apiClient.getFeaturedEvents(),
  getEventBySlug: (slug: string) => apiClient.getEventById(slug)
};

export const servicesClient = {
  getServices: (params: any) => apiClient.getServices(params),
  getServiceCategories: () => apiClient.getServiceCategories(),
  searchServices: (params: any) => apiClient.getServices({ search: params.q, page: params.page, limit: params.limit }),
  getFeaturedServices: () => apiClient.getFeaturedServices(),
  getServiceBySlug: (slug: string) => apiClient.getServiceById(slug)
};

export const newsletterClient = {
  subscribe: async (data: any) => {
    // Newsletter subscription through business inquiry for now
    return await apiClient.submitBusinessInquiry({
      first_name: data.firstName || data.first_name,
      last_name: data.lastName || data.last_name || '',
      email: data.email,
      phone: data.phone || '',
      company: data.company || '',
      inquiry_type: 'newsletter_subscription',
      message: `Newsletter subscription request. Email: ${data.email}`,
      preferred_contact_method: 'email'
    });
  }
};

// Import the contract client for all operations
import { apiClient } from '../api-client';