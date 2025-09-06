import { config } from '../../environments';
import { RestClientCache } from './RestClientCache';

export interface RestClientConfig {
  baseUrl: string;
  timeout: number;
  retryAttempts?: number;
}

export interface RestResponse<T> {
  data: T;
  success: boolean;
  message?: string;
  errors?: string[];
}

export interface PaginatedRestResponse<T> {
  data: T[];
  pagination: {
    page: number;
    pageSize: number;
    total: number;
    totalPages: number;
  };
  success: boolean;
  message?: string;
  errors?: string[];
}

export class RestError extends Error {
  constructor(
    message: string,
    public status: number,
    public response?: any,
    public correlationId?: string
  ) {
    super(message);
    this.name = 'RestError';
  }
}

export abstract class BaseRestClient {
  protected readonly baseUrl: string;
  protected readonly timeout: number;
  protected readonly retryAttempts: number;
  private cache = new RestClientCache();

  constructor(clientConfig: RestClientConfig) {
    this.baseUrl = clientConfig.baseUrl;
    this.timeout = clientConfig.timeout;
    this.retryAttempts = clientConfig.retryAttempts || 3;
  }

  protected async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    return this.requestWithRetry<T>(endpoint, options, this.retryAttempts);
  }

  protected async get<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'GET',
    });
  }

  protected async post<T>(endpoint: string, data?: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  private async requestWithRetry<T>(
    endpoint: string,
    options: RequestInit,
    maxRetries: number
  ): Promise<T> {
    let lastError: Error;
    
    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        const url = `${this.baseUrl}${endpoint}`;
        console.log(`🌐 [REST] ${options.method || 'GET'} ${url} (attempt ${attempt}/${maxRetries})`);
        
        const response = await fetch(url, {
          ...options,
          headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json',
            'X-Retry-Attempt': attempt.toString(),
            ...options.headers,
          },
          signal: AbortSignal.timeout(this.timeout),
        });

        if (!response.ok) {
          const errorDetails = await this.parseErrorResponse(response);
          const restError = this.createRestError(response.status, response.statusText, errorDetails);
          
          // Check if error is retryable
          if (attempt < maxRetries && this.isRetryableError(response.status)) {
            lastError = restError;
            const delayMs = this.calculateRetryDelay(attempt);
            console.log(`🔄 [REST] Retrying after ${delayMs}ms due to retryable error ${response.status}`);
            await this.delay(delayMs);
            continue;
          }
          
          throw restError;
        }

        const contentType = response.headers.get('content-type');
        if (contentType && contentType.includes('application/json')) {
          const data = await response.json();
          console.log(`✅ [REST] ${options.method || 'GET'} ${endpoint} success`);
          return data;
        } else {
          // Handle non-JSON responses
          const text = await response.text();
          console.log(`✅ [REST] ${options.method || 'GET'} ${endpoint} success (non-JSON)`);
          return text as unknown as T;
        }
      } catch (error) {
        if (error instanceof RestError) {
          lastError = error;
          // Check if it's a retryable error
          if (attempt < maxRetries && error.status >= 500) {
            const delayMs = this.calculateRetryDelay(attempt);
            console.log(`🔄 [REST] Retrying after ${delayMs}ms due to server error ${error.status}`);
            await this.delay(delayMs);
            continue;
          }
          throw error;
        }

        if (error instanceof Error) {
          if (error.name === 'AbortError') {
            lastError = new RestError(`Request timeout after ${this.timeout}ms`, 408);
            if (attempt < maxRetries) {
              const delayMs = this.calculateRetryDelay(attempt);
              console.log(`🔄 [REST] Retrying after ${delayMs}ms due to timeout`);
              await this.delay(delayMs);
              continue;
            }
            throw lastError;
          }
          lastError = new RestError(`Network error: ${error.message}`, 0);
          throw lastError;
        }

        lastError = new RestError('Unknown network error', 0);
        throw lastError;
      }
    }
    
    // If we get here, all retries failed
    throw lastError || new RestError('All retry attempts failed', 0);
  }

  private async parseErrorResponse(response: Response): Promise<{ error?: string; message?: string; correlationId?: string }> {
    try {
      const contentType = response.headers.get('content-type');
      if (contentType?.includes('application/json')) {
        return await response.json();
      }
      return { message: await response.text() };
    } catch {
      return { message: response.statusText };
    }
  }

  private createRestError(status: number, statusText: string, details: { error?: string; message?: string; correlationId?: string }): RestError {
    const correlationId = details.correlationId || 'unknown';
    
    switch (status) {
      case 400:
        return new RestError(
          `Bad request: ${details.message || statusText}`,
          status,
          details,
          correlationId
        );
      case 401:
        return new RestError(
          `Unauthorized: ${details.message || statusText}`,
          status,
          details,
          correlationId
        );
      case 403:
        return new RestError(
          `Access forbidden: ${details.message || statusText}`,
          status,
          details,
          correlationId
        );
      case 404:
        return new RestError(
          `Not found: ${details.message || statusText}`,
          status,
          details,
          correlationId
        );
      case 429:
        return new RestError(
          `Rate limit exceeded: ${details.message || statusText}`,
          status,
          details,
          correlationId
        );
      case 500:
      case 502:
      case 503:
      case 504:
        return new RestError(
          `Server error: ${details.message || statusText}`,
          status,
          details,
          correlationId
        );
      default:
        return new RestError(
          `HTTP error ${status}: ${details.message || statusText}`,
          status,
          details,
          correlationId
        );
    }
  }

  private isRetryableError(status: number): boolean {
    // Retry on server errors and specific client errors
    return status >= 500 || status === 429 || status === 408;
  }

  private calculateRetryDelay(attempt: number): number {
    // Exponential backoff: 1s, 2s, 4s for attempts 1, 2, 3
    return Math.min(1000 * Math.pow(2, attempt - 1), 5000);
  }

  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  // Cache integration methods for domain REST clients
  protected async requestWithCache<T>(
    endpoint: string,
    options: RequestInit,
    cacheKey: string,
    ttl: number
  ): Promise<T> {
    return this.cache.requestWithCache<T>(this, endpoint, options, cacheKey, ttl);
  }

  protected invalidateCacheKey(key: string): void {
    this.cache.invalidateCacheKey(key);
  }

  protected invalidateCachePattern(pattern: string): void {
    this.cache.invalidateCachePattern(pattern);
  }

  // Health check method
  public async healthCheck(): Promise<{ status: string; service: string }> {
    try {
      const response = await fetch(`${this.baseUrl}/health`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
        signal: AbortSignal.timeout(this.timeout),
      });

      if (!response.ok) {
        throw new Error(`Health check failed: ${response.statusText}`);
      }

      const data = await response.text();
      return {
        status: data === 'Healthy' ? 'healthy' : 'unhealthy',
        service: this.constructor.name,
      };
    } catch (error) {
      throw new RestError(
        `Health check failed: ${error instanceof Error ? error.message : 'Unknown error'}`,
        0
      );
    }
  }
}