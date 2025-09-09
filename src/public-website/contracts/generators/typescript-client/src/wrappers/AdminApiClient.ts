/**
 * Admin API Client Wrapper
 * Provides a simplified interface for the admin API with authentication,
 * error handling, and role-based access control integration.
 */

import { Configuration, DefaultApi as GeneratedAdminApi } from '../generated/admin';
import type { 
  AdminUser,
  NewsArticle,
  CreateNewsArticleRequest,
  UpdateNewsArticleRequest,
  Service,
  Event,
  Inquiry,
  DashboardAnalytics
} from '../generated/admin';
import { ApiError, ApiResponse, RequestOptions, PaginationParams } from '../types/common';
import { handleApiError, buildRequestOptions } from '../utils/api-utils';

export interface AdminApiClientConfig {
  baseUrl?: string;
  timeout?: number;
  retries?: number;
  defaultHeaders?: Record<string, string>;
  authToken?: string;
  onTokenExpired?: () => void;
}

export interface LoginCredentials {
  email: string;
  password: string;
  remember_me?: boolean;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
  user: AdminUser;
}

export class AdminApiClient {
  private api: GeneratedAdminApi;
  private config: AdminApiClientConfig;
  private authToken?: string;

  constructor(config: AdminApiClientConfig = {}) {
    this.config = {
      baseUrl: 'http://localhost:4001/api/v1',
      timeout: 30000,
      retries: 3,
      ...config
    };

    this.authToken = config.authToken;
    
    const configuration = new Configuration({
      basePath: this.config.baseUrl,
      fetchApi: this.createFetchWithInterceptors(),
      headers: this.authToken ? { Authorization: `Bearer ${this.authToken}` } : undefined
    });

    this.api = new GeneratedAdminApi(configuration);
  }

  // Authentication methods
  setAuthToken(token: string): void {
    this.authToken = token;
    // Update configuration with new token
    const configuration = new Configuration({
      basePath: this.config.baseUrl,
      fetchApi: this.createFetchWithInterceptors(),
      headers: { Authorization: `Bearer ${token}` }
    });
    this.api = new GeneratedAdminApi(configuration);
  }

  clearAuthToken(): void {
    this.authToken = undefined;
    const configuration = new Configuration({
      basePath: this.config.baseUrl,
      fetchApi: this.createFetchWithInterceptors()
    });
    this.api = new GeneratedAdminApi(configuration);
  }

  async login(credentials: LoginCredentials): Promise<ApiResponse<LoginResponse>> {
    try {
      const response = await this.api.adminLogin({ 
        body: credentials
      });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async refreshToken(refresh_token: string): Promise<ApiResponse<{ access_token: string, expires_in: number, token_type: string }>> {
    try {
      const response = await this.api.refreshToken({ 
        body: { refresh_token }
      });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async logout(): Promise<ApiResponse<any>> {
    try {
      const response = await this.api.adminLogout();
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  // User management
  async getAdminUsers(params?: PaginationParams & { search?: string, role?: string, status?: string }): Promise<ApiResponse<{ data: AdminUser[], pagination: any }>> {
    try {
      const response = await this.api.getAdminUsers({
        page: params?.page,
        limit: params?.limit,
        search: params?.search,
        role: params?.role as any,
        status: params?.status as any
      });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async getAdminUserById(id: string): Promise<ApiResponse<{ data: AdminUser }>> {
    try {
      const response = await this.api.getAdminUserById({ id });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  // News management
  async getNewsAdmin(params?: PaginationParams & { search?: string, status?: string, category_id?: string }): Promise<ApiResponse<{ data: NewsArticle[], pagination: any }>> {
    try {
      const response = await this.api.getNewsAdmin({
        page: params?.page,
        limit: params?.limit,
        search: params?.search,
        status: params?.status as any,
        category_id: params?.category_id
      });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async createNewsArticle(article: CreateNewsArticleRequest): Promise<ApiResponse<{ data: NewsArticle }>> {
    try {
      const response = await this.api.createNewsArticle({ 
        createNewsArticleRequest: article 
      });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async updateNewsArticle(id: string, article: UpdateNewsArticleRequest): Promise<ApiResponse<{ data: NewsArticle }>> {
    try {
      const response = await this.api.updateNewsArticle({ 
        id, 
        updateNewsArticleRequest: article 
      });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async deleteNewsArticle(id: string): Promise<ApiResponse<any>> {
    try {
      const response = await this.api.deleteNewsArticle({ id });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async publishNewsArticle(id: string): Promise<ApiResponse<any>> {
    try {
      const response = await this.api.publishNewsArticle({ id });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async unpublishNewsArticle(id: string): Promise<ApiResponse<any>> {
    try {
      const response = await this.api.unpublishNewsArticle({ id });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  // Inquiries management
  async getInquiries(params?: PaginationParams & { search?: string, inquiry_type?: string, status?: string }): Promise<ApiResponse<{ data: Inquiry[], pagination: any }>> {
    try {
      const response = await this.api.getInquiries({
        page: params?.page,
        limit: params?.limit,
        search: params?.search,
        inquiry_type: params?.inquiry_type as any,
        status: params?.status as any
      });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async getInquiryById(id: string): Promise<ApiResponse<{ data: Inquiry }>> {
    try {
      const response = await this.api.getInquiryById({ id });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  async updateInquiryStatus(id: string, status: string, notes?: string, assigned_to?: string): Promise<ApiResponse<any>> {
    try {
      const response = await this.api.updateInquiryStatus({ 
        id, 
        body: { 
          status: status as any, 
          notes, 
          assigned_to 
        }
      });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  // Analytics
  async getDashboardAnalytics(period?: string): Promise<ApiResponse<{ data: DashboardAnalytics }>> {
    try {
      const response = await this.api.getDashboardAnalytics({ 
        period: period as any 
      });
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  // Health check
  async getAdminHealth(): Promise<ApiResponse<any>> {
    try {
      const response = await this.api.getAdminHealth();
      return { data: response, error: null };
    } catch (error) {
      return handleApiError(error);
    }
  }

  private createFetchWithInterceptors(): typeof fetch {
    return async (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
      const options = buildRequestOptions(init, this.config);
      
      // Add authorization header if token is available
      if (this.authToken && options.headers) {
        (options.headers as any)['Authorization'] = `Bearer ${this.authToken}`;
      }
      
      try {
        const response = await fetch(input, options);
        
        // Handle token expiration
        if (response.status === 401 && this.authToken) {
          this.clearAuthToken();
          if (this.config.onTokenExpired) {
            this.config.onTokenExpired();
          }
        }
        
        // Response interceptor
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        return response;
      } catch (error) {
        console.error('Admin API request failed:', error);
        throw error;
      }
    };
  }
}