/**
 * Public API Client Wrapper
 * Provides a simplified interface for the public API with error handling,
 * request/response interceptors, and integration with frontend frameworks.
 */

import { Configuration, DefaultApi as GeneratedPublicApi } from '../generated/public';
import type { 
  Service, 
  NewsArticle, 
  ResearchPublication, 
  Event,
  MediaInquiryRequest,
  BusinessInquiryRequest,
  DonationInquiryRequest,
  VolunteerInquiryRequest
} from '../generated/public';
import { ApiError, ApiResponse, RequestOptions, PaginationParams } from '../types/common';
import { handleApiError, buildRequestOptions } from '../utils/api-utils';

export interface PublicApiClientConfig {
  baseUrl?: string;
  timeout?: number;
  retries?: number;
  defaultHeaders?: Record<string, string>;
}

export class PublicApiClient {
  private api: GeneratedPublicApi;
  private config: PublicApiClientConfig;

  constructor(config: PublicApiClientConfig = {}) {
    this.config = {
      baseUrl: 'http://localhost:4000/api/v1',
      timeout: 30000,
      retries: 3,
      ...config
    };

    const configuration = new Configuration({
      basePath: this.config.baseUrl,
      fetchApi: this.createFetchWithInterceptors()
    });

    this.api = new GeneratedPublicApi(configuration);
  }

  // Health endpoints
  async getHealth(): Promise<ApiResponse<any>> {
    try {
      const response = await this.api.getHealth();
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async getHealthReady(): Promise<ApiResponse<any>> {
    try {
      const response = await this.api.getHealthReady();
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  // Services endpoints
  async getServices(params?: PaginationParams & { search?: string }): Promise<ApiResponse<{ data: Service[], pagination: any }>> {
    try {
      const response = await this.api.getServices({
        page: params?.page,
        limit: params?.limit,
        search: params?.search
      });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async getServiceById(id: string): Promise<ApiResponse<{ data: Service }>> {
    try {
      const response = await this.api.getServiceById({ id });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async getServiceBySlug(slug: string): Promise<ApiResponse<{ data: Service }>> {
    try {
      const response = await this.api.getServiceBySlug({ slug });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async getFeaturedServices(): Promise<ApiResponse<{ data: Service[] }>> {
    try {
      const response = await this.api.getFeaturedServices();
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  // News endpoints
  async getNews(params?: PaginationParams & { search?: string }): Promise<ApiResponse<{ data: NewsArticle[], pagination: any }>> {
    try {
      const response = await this.api.getNews({
        page: params?.page,
        limit: params?.limit,
        search: params?.search
      });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async getNewsById(id: string): Promise<ApiResponse<{ data: NewsArticle }>> {
    try {
      const response = await this.api.getNewsById({ id });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async getNewsBySlug(slug: string): Promise<ApiResponse<{ data: NewsArticle }>> {
    try {
      const response = await this.api.getNewsBySlug({ slug });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async getFeaturedNews(): Promise<ApiResponse<{ data: NewsArticle[] }>> {
    try {
      const response = await this.api.getFeaturedNews();
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  // Research endpoints
  async getResearch(params?: PaginationParams & { search?: string }): Promise<ApiResponse<{ data: ResearchPublication[], pagination: any }>> {
    try {
      const response = await this.api.getResearch({
        page: params?.page,
        limit: params?.limit,
        search: params?.search
      });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async getResearchById(id: string): Promise<ApiResponse<{ data: ResearchPublication }>> {
    try {
      const response = await this.api.getResearchById({ id });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async getResearchBySlug(slug: string): Promise<ApiResponse<{ data: ResearchPublication }>> {
    try {
      const response = await this.api.getResearchBySlug({ slug });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  // Events endpoints
  async getEvents(params?: PaginationParams & { search?: string }): Promise<ApiResponse<{ data: Event[], pagination: any }>> {
    try {
      const response = await this.api.getEvents({
        page: params?.page,
        limit: params?.limit,
        search: params?.search
      });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async getEventById(id: string): Promise<ApiResponse<{ data: Event }>> {
    try {
      const response = await this.api.getEventById({ id });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  // Inquiry submissions
  async submitMediaInquiry(inquiry: MediaInquiryRequest): Promise<ApiResponse<{ data: { inquiry_id: string, submitted_at: string } }>> {
    try {
      const response = await this.api.submitMediaInquiry({ mediaInquiryRequest: inquiry });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async submitBusinessInquiry(inquiry: BusinessInquiryRequest): Promise<ApiResponse<{ data: { inquiry_id: string, submitted_at: string } }>> {
    try {
      const response = await this.api.submitBusinessInquiry({ businessInquiryRequest: inquiry });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async submitDonationInquiry(inquiry: DonationInquiryRequest): Promise<ApiResponse<{ data: { inquiry_id: string, submitted_at: string } }>> {
    try {
      const response = await this.api.submitDonationInquiry({ donationInquiryRequest: inquiry });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async submitVolunteerInquiry(inquiry: VolunteerInquiryRequest): Promise<ApiResponse<{ data: { inquiry_id: string, submitted_at: string } }>> {
    try {
      const response = await this.api.submitVolunteerInquiry({ volunteerInquiryRequest: inquiry });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  private createFetchWithInterceptors(): typeof fetch {
    return async (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
      const options = buildRequestOptions(init, this.config);
      
      try {
        const response = await fetch(input, options);
        
        // Response interceptor
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        return response;
      } catch (error) {
        console.error('API request failed:', error);
        throw error;
      }
    };
  }
}