// Contract-compliant API client using generated TypeScript clients
import { Configuration, HealthApi, ServicesApi, NewsApi, ResearchApi, EventsApi, InquiriesApi } from '@international-center/public-api-client';

// Simple cache implementation for contract client performance
class ContractClientCache {
  private cache = new Map<string, { data: any; timestamp: number; ttl: number }>();
  private readonly DEFAULT_TTL = 5 * 60 * 1000; // 5 minutes

  private getCacheKey(method: string, params?: any): string {
    return `${method}:${JSON.stringify(params || {})}`;
  }

  get<T>(method: string, params?: any): T | null {
    const key = this.getCacheKey(method, params);
    const entry = this.cache.get(key);
    
    if (!entry) return null;
    
    // Check if cache entry is expired
    if (Date.now() > entry.timestamp + entry.ttl) {
      this.cache.delete(key);
      return null;
    }
    
    return entry.data;
  }

  set<T>(method: string, params: any, data: T, ttl?: number): void {
    const key = this.getCacheKey(method, params);
    this.cache.set(key, {
      data,
      timestamp: Date.now(),
      ttl: ttl || this.DEFAULT_TTL
    });
  }

  invalidate(method: string, params?: any): void {
    const key = this.getCacheKey(method, params);
    this.cache.delete(key);
  }

  clear(): void {
    this.cache.clear();
  }
}

// API Configuration
class APIClientConfig {
  private baseURL: string;
  private daprPort: string;
  private useDapr: boolean;

  constructor() {
    this.baseURL = import.meta.env.API_BASE_URL || 'http://localhost:8080';
    this.daprPort = import.meta.env.DAPR_HTTP_PORT || '3500';
    this.useDapr = import.meta.env.USE_DAPR !== 'false';
  }

  getConfiguration(): Configuration {
    const basePath = this.useDapr 
      ? `http://localhost:${this.daprPort}/v1.0/invoke/content-api/method`
      : `${this.baseURL}/api/v1`;
    
    return new Configuration({
      basePath,
      fetchApi: fetch,
      middleware: [
        {
          pre: async (context) => {
            // Add correlation ID
            context.init.headers = {
              ...context.init.headers,
              'X-Correlation-ID': crypto.randomUUID(),
              'Content-Type': 'application/json',
            };
          }
        }
      ]
    });
  }

  getFallbackConfiguration(): Configuration {
    return new Configuration({
      basePath: `${this.baseURL}/api/v1`,
      fetchApi: fetch,
      middleware: [
        {
          pre: async (context) => {
            context.init.headers = {
              ...context.init.headers,
              'X-Correlation-ID': crypto.randomUUID(),
              'Content-Type': 'application/json',
            };
          }
        }
      ]
    });
  }
}

// Contract-compliant API client factory
export class ContractAPIClient {
  private config: APIClientConfig;
  private healthApi: HealthApi;
  private servicesApi: ServicesApi;
  private newsApi: NewsApi;
  private researchApi: ResearchApi;
  private eventsApi: EventsApi;
  private inquiriesApi: InquiriesApi;
  public cache: ContractClientCache;

  constructor() {
    this.config = new APIClientConfig();
    const configuration = this.config.getConfiguration();
    
    this.healthApi = new HealthApi(configuration);
    this.servicesApi = new ServicesApi(configuration);
    this.newsApi = new NewsApi(configuration);
    this.researchApi = new ResearchApi(configuration);
    this.eventsApi = new EventsApi(configuration);
    this.inquiriesApi = new InquiriesApi(configuration);
    this.cache = new ContractClientCache();
  }

  // Health API methods
  async getHealth() {
    try {
      return await this.healthApi.getHealth();
    } catch (error) {
      // Fallback to direct API call if Dapr fails
      if (this.config.useDapr) {
        const fallbackApi = new HealthApi(this.config.getFallbackConfiguration());
        return await fallbackApi.getHealth();
      }
      throw error;
    }
  }

  // Services API methods
  async getServices(params?: { page?: number; limit?: number; search?: string; categoryId?: string }) {
    try {
      return await this.servicesApi.getServices(params);
    } catch (error) {
      if (this.config.useDapr) {
        const fallbackApi = new ServicesApi(this.config.getFallbackConfiguration());
        return await fallbackApi.getServices(params);
      }
      throw error;
    }
  }

  async getServiceById(id: string) {
    try {
      return await this.servicesApi.getServiceById({ id });
    } catch (error) {
      if (this.config.useDapr) {
        const fallbackApi = new ServicesApi(this.config.getFallbackConfiguration());
        return await fallbackApi.getServiceById({ id });
      }
      throw error;
    }
  }

  async getFeaturedServices() {
    try {
      return await this.servicesApi.getFeaturedServices();
    } catch (error) {
      if (this.config.useDapr) {
        const fallbackApi = new ServicesApi(this.config.getFallbackConfiguration());
        return await fallbackApi.getFeaturedServices();
      }
      throw error;
    }
  }

  async getServiceCategories() {
    try {
      return await this.servicesApi.getServiceCategories();
    } catch (error) {
      if (this.config.useDapr) {
        const fallbackApi = new ServicesApi(this.config.getFallbackConfiguration());
        return await fallbackApi.getServiceCategories();
      }
      throw error;
    }
  }

  // News API methods
  async getNews(params?: { page?: number; limit?: number; search?: string; categoryId?: string }) {
    // Check cache first
    const cachedResult = this.cache.get('getNews', params);
    if (cachedResult) {
      return cachedResult;
    }

    try {
      const result = await this.newsApi.getNews(params);
      // Cache successful result
      this.cache.set('getNews', params, result);
      return result;
    } catch (error) {
      if (this.config.useDapr) {
        const fallbackApi = new NewsApi(this.config.getFallbackConfiguration());
        const result = await fallbackApi.getNews(params);
        // Cache fallback result
        this.cache.set('getNews', params, result);
        return result;
      }
      throw error;
    }
  }

  async getNewsById(id: string) {
    try {
      return await this.newsApi.getNewsById({ id });
    } catch (error) {
      if (this.config.useDapr) {
        const fallbackApi = new NewsApi(this.config.getFallbackConfiguration());
        return await fallbackApi.getNewsById({ id });
      }
      throw error;
    }
  }

  async getFeaturedNews() {
    try {
      return await this.newsApi.getFeaturedNews();
    } catch (error) {
      if (this.config.useDapr) {
        const fallbackApi = new NewsApi(this.config.getFallbackConfiguration());
        return await fallbackApi.getFeaturedNews();
      }
      throw error;
    }
  }

  async getNewsCategories() {
    try {
      return await this.newsApi.getNewsCategories();
    } catch (error) {
      if (this.config.useDapr) {
        const fallbackApi = new NewsApi(this.config.getFallbackConfiguration());
        return await fallbackApi.getNewsCategories();
      }
      throw error;
    }
  }

  // Research API methods
  async getResearch(params?: { page?: number; limit?: number; search?: string; categoryId?: string }) {
    try {
      return await this.researchApi.getResearch(params);
    } catch (error) {
      if (this.config.useDapr) {
        const fallbackApi = new ResearchApi(this.config.getFallbackConfiguration());
        return await fallbackApi.getResearch(params);
      }
      throw error;
    }
  }

  async getResearchById(id: string) {
    try {
      return await this.researchApi.getResearchById({ id });
    } catch (error) {
      if (this.config.useDapr) {
        const fallbackApi = new ResearchApi(this.config.getFallbackConfiguration());
        return await fallbackApi.getResearchById({ id });
      }
      throw error;
    }
  }

  async getFeaturedResearch() {
    try {
      return await this.researchApi.getFeaturedResearch();
    } catch (error) {
      if (this.config.useDapr) {
        const fallbackApi = new ResearchApi(this.config.getFallbackConfiguration());
        return await fallbackApi.getFeaturedResearch();
      }
      throw error;
    }
  }

  async getResearchCategories() {
    try {
      return await this.researchApi.getResearchCategories();
    } catch (error) {
      if (this.config.useDapr) {
        const fallbackApi = new ResearchApi(this.config.getFallbackConfiguration());
        return await fallbackApi.getResearchCategories();
      }
      throw error;
    }
  }

  // Events API methods
  async getEvents(params?: { page?: number; limit?: number; search?: string; categoryId?: string }) {
    try {
      return await this.eventsApi.getEvents(params);
    } catch (error) {
      if (this.config.useDapr) {
        const fallbackApi = new EventsApi(this.config.getFallbackConfiguration());
        return await fallbackApi.getEvents(params);
      }
      throw error;
    }
  }

  async getEventById(id: string) {
    try {
      return await this.eventsApi.getEventById({ id });
    } catch (error) {
      if (this.config.useDapr) {
        const fallbackApi = new EventsApi(this.config.getFallbackConfiguration());
        return await fallbackApi.getEventById({ id });
      }
      throw error;
    }
  }

  async getFeaturedEvents() {
    try {
      return await this.eventsApi.getFeaturedEvents();
    } catch (error) {
      if (this.config.useDapr) {
        const fallbackApi = new EventsApi(this.config.getFallbackConfiguration());
        return await fallbackApi.getFeaturedEvents();
      }
      throw error;
    }
  }

  async getEventCategories() {
    try {
      return await this.eventsApi.getEventCategories();
    } catch (error) {
      if (this.config.useDapr) {
        const fallbackApi = new EventsApi(this.config.getFallbackConfiguration());
        return await fallbackApi.getEventCategories();
      }
      throw error;
    }
  }

  // Inquiries API methods  
  async submitMediaInquiry(inquiry: any) {
    try {
      return await this.inquiriesApi.submitMediaInquiry({ mediaInquiryRequest: inquiry });
    } catch (error) {
      if (this.config.useDapr) {
        const fallbackApi = new InquiriesApi(this.config.getFallbackConfiguration());
        return await fallbackApi.submitMediaInquiry({ mediaInquiryRequest: inquiry });
      }
      throw error;
    }
  }

  async submitBusinessInquiry(inquiry: any) {
    try {
      return await this.inquiriesApi.submitBusinessInquiry({ businessInquiryRequest: inquiry });
    } catch (error) {
      if (this.config.useDapr) {
        const fallbackApi = new InquiriesApi(this.config.getFallbackConfiguration());
        return await fallbackApi.submitBusinessInquiry({ businessInquiryRequest: inquiry });
      }
      throw error;
    }
  }

  async submitDonationInquiry(inquiry: any) {
    try {
      // Donations use business inquiry endpoint for now
      return await this.inquiriesApi.submitBusinessInquiry({ businessInquiryRequest: inquiry });
    } catch (error) {
      if (this.config.useDapr) {
        const fallbackApi = new InquiriesApi(this.config.getFallbackConfiguration());
        return await fallbackApi.submitBusinessInquiry({ businessInquiryRequest: inquiry });
      }
      throw error;
    }
  }

  async submitVolunteerInquiry(inquiry: any) {
    try {
      // Volunteers use media inquiry endpoint for now
      return await this.inquiriesApi.submitMediaInquiry({ mediaInquiryRequest: inquiry });
    } catch (error) {
      if (this.config.useDapr) {
        const fallbackApi = new InquiriesApi(this.config.getFallbackConfiguration());
        return await fallbackApi.submitMediaInquiry({ mediaInquiryRequest: inquiry });
      }
      throw error;
    }
  }
}

// Singleton instance for application-wide use
export const apiClient = new ContractAPIClient();

// Re-export types for convenience
export type * from '@international-center/public-api-client';