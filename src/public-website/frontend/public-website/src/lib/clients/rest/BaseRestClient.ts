import { config } from '../../environments';

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

// Unified cache interfaces
export interface CacheEntry<T> {
  data: T;
  timestamp: number;
  ttl: number;
}

export interface RequestMetrics {
  totalRequests: number;
  cacheHits: number;
  cacheMisses: number;
  averageResponseTime: number;
  errorCount: number;
}

export interface CacheStats {
  size: number;
  hitRate: number;
}

export interface CacheTTL {
  CATEGORIES: number;
  FEATURED: number;
  DETAIL: number;
  LIST: number;
}

/**
 * Standard cache TTL configurations for different content types
 * Provides consistent caching strategies across all domains
 */
export const STANDARD_CACHE_TTL: CacheTTL = {
  CATEGORIES: 15 * 60 * 1000, // 15 minutes - categories change infrequently
  FEATURED: 5 * 60 * 1000, // 5 minutes - featured content changes occasionally
  DETAIL: 2 * 60 * 1000, // 2 minutes - individual items may be updated frequently
  LIST: 30 * 1000, // 30 seconds - lists change more frequently
};

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
  
  // Unified cache system
  private cache = new Map<string, CacheEntry<any>>();
  private pendingRequests = new Map<string, Promise<any>>();
  private metrics: RequestMetrics = {
    totalRequests: 0,
    cacheHits: 0,
    cacheMisses: 0,
    averageResponseTime: 0,
    errorCount: 0,
  };
  private cleanupInterval: NodeJS.Timeout;

  constructor(clientConfig: RestClientConfig) {
    this.baseUrl = clientConfig.baseUrl;
    this.timeout = clientConfig.timeout;
    this.retryAttempts = clientConfig.retryAttempts || 3;
    
    // Initialize cache cleanup interval
    this.cleanupInterval = setInterval(() => this.clearExpiredCache(), 5 * 60 * 1000);
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

  // Unified cache system methods
  protected async requestWithCache<T>(
    endpoint: string,
    options: RequestInit,
    cacheKey: string,
    ttl: number
  ): Promise<T> {
    // Environment-aware cache bypass: In test mode, bypass cache for complete test isolation
    if (import.meta.env.MODE === 'test') {
      return this.request<T>(endpoint, options);
    }

    const startTime = Date.now();
    this.metrics.totalRequests++;

    // Check cache first
    const cached = this.getFromCache<T>(cacheKey);
    if (cached) {
      this.metrics.cacheHits++;
      this.updateResponseTime(startTime);
      return cached;
    }

    this.metrics.cacheMisses++;

    // Check for pending request to prevent duplicate requests
    const requestKey = `${endpoint}:${JSON.stringify(options)}`;
    if (this.pendingRequests.has(requestKey)) {
      return this.pendingRequests.get(requestKey)!;
    }

    // Make the request using the client's request method
    const requestPromise = this.request<T>(endpoint, options);
    this.pendingRequests.set(requestKey, requestPromise);

    try {
      const result = await requestPromise;

      // Cache the result
      this.setCache(cacheKey, result, ttl);

      // Update performance metrics
      this.updateResponseTime(startTime);

      return result;
    } catch (error) {
      this.metrics.errorCount++;
      throw error;
    } finally {
      this.pendingRequests.delete(requestKey);
    }
  }

  /**
   * Get data from cache if not expired
   */
  private getFromCache<T>(key: string): T | null {
    const entry = this.cache.get(key);
    if (!entry) return null;

    const now = Date.now();
    if (now > entry.timestamp + entry.ttl) {
      this.cache.delete(key);
      return null;
    }

    return entry.data;
  }

  /**
   * Set data in cache with TTL
   */
  private setCache<T>(key: string, data: T, ttl: number): void {
    this.cache.set(key, {
      data,
      timestamp: Date.now(),
      ttl,
    });
  }

  /**
   * Clear expired cache entries
   */
  private clearExpiredCache(): void {
    const now = Date.now();
    const keysToDelete = [];

    for (const [key, entry] of this.cache.entries()) {
      if (now > entry.timestamp + entry.ttl) {
        keysToDelete.push(key);
      }
    }

    keysToDelete.forEach(key => this.cache.delete(key));
  }

  /**
   * Update average response time metric
   */
  private updateResponseTime(startTime: number): void {
    const responseTime = Date.now() - startTime;
    this.metrics.averageResponseTime =
      (this.metrics.averageResponseTime * (this.metrics.totalRequests - 1) + responseTime) /
      this.metrics.totalRequests;
  }

  /**
   * Invalidate specific cache key
   */
  protected invalidateCacheKey(key: string): void {
    this.cache.delete(key);
  }

  /**
   * Invalidate cache keys matching pattern
   */
  protected invalidateCachePattern(pattern: string): void {
    const keysToDelete = [];
    for (const key of this.cache.keys()) {
      if (key.startsWith(pattern)) {
        keysToDelete.push(key);
      }
    }
    keysToDelete.forEach(key => this.cache.delete(key));
  }

  /**
   * Get performance metrics
   */
  public getMetrics(): RequestMetrics {
    return { ...this.metrics };
  }

  /**
   * Clear all cache entries and reset metrics
   */
  public clearCache(): void {
    this.cache.clear();
    this.pendingRequests.clear();
    this.metrics = {
      totalRequests: 0,
      cacheHits: 0,
      cacheMisses: 0,
      averageResponseTime: 0,
      errorCount: 0,
    };
  }

  /**
   * Get cache statistics
   */
  public getCacheStats(): CacheStats {
    const hitRate = this.metrics.totalRequests > 0
      ? (this.metrics.cacheHits / this.metrics.totalRequests) * 100
      : 0;

    return {
      size: this.cache.size,
      hitRate: Math.round(hitRate * 100) / 100,
    };
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