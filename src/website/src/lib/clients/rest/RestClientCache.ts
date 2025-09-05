// Shared REST Client Caching and Performance Monitoring Utilities
// Eliminates code duplication across News, Events, Services, Research clients

import type { BaseRestClient } from './BaseRestClient';

// Shared interfaces
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
 * Shared caching and performance monitoring functionality for REST clients
 * Provides consistent caching, metrics tracking, and performance optimization
 */
export class RestClientCache {
  private cache = new Map<string, CacheEntry<any>>();
  private pendingRequests = new Map<string, Promise<any>>();
  private metrics: RequestMetrics = {
    totalRequests: 0,
    cacheHits: 0,
    cacheMisses: 0,
    averageResponseTime: 0,
    errorCount: 0,
  };

  constructor() {
    // Clear expired cache entries every 5 minutes
    setInterval(() => this.clearExpiredCache(), 5 * 60 * 1000);
  }

  /**
   * Request with caching and deduplication
   * Provides unified caching behavior across all REST clients
   */
  async requestWithCache<T>(
    client: BaseRestClient,
    endpoint: string,
    options: RequestInit,
    cacheKey: string,
    ttl: number
  ): Promise<T> {
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
    const requestPromise = (client as any).request<T>(endpoint, options);
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

  /**
   * Invalidate specific cache key
   */
  public invalidateCacheKey(key: string): void {
    this.cache.delete(key);
  }

  /**
   * Invalidate cache keys matching pattern
   */
  public invalidateCachePattern(pattern: string): void {
    const keysToDelete = [];
    for (const key of this.cache.keys()) {
      if (key.startsWith(pattern)) {
        keysToDelete.push(key);
      }
    }
    keysToDelete.forEach(key => this.cache.delete(key));
  }
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