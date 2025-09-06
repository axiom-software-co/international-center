import { describe, it, expect, beforeAll, afterAll, vi } from 'vitest';
import { NewsRestClient } from '../../lib/clients/news/NewsRestClient';
import { EventsRestClient } from '../../lib/clients/events/EventsRestClient';

/**
 * Integration Tests for API Contracts with Real Infrastructure
 * 
 * These tests validate that frontend REST clients work correctly with deployed backend services.
 * They test real HTTP communication with the development environment infrastructure.
 * 
 * Prerequisites:
 * - Development environment deployed with container factory
 * - Public gateway accessible at http://127.0.0.1:9001  
 * - Backend services health validated and ready
 * - All services have proper Dapr sidecar integration
 */

describe('API Contract Integration Tests', () => {
  let newsClient: NewsRestClient;
  let eventsClient: EventsRestClient;
  
  // Use development environment gateway URL
  const PUBLIC_GATEWAY_URL = 'http://127.0.0.1:9001';
  
  beforeAll(async () => {
    // Initialize clients with real development environment URL
    newsClient = new NewsRestClient(PUBLIC_GATEWAY_URL);
    eventsClient = new EventsRestClient(PUBLIC_GATEWAY_URL);
    
    // Wait for services to be ready
    await validateServiceHealth();
  });
  
  afterAll(() => {
    // Clear any client caches
    newsClient?.clearCache();
    eventsClient?.clearCache();
  });

  /**
   * Validate that backend services are healthy before running integration tests
   */
  async function validateServiceHealth(): Promise<void> {
    const healthChecks = [
      { name: 'Public Gateway', url: `${PUBLIC_GATEWAY_URL}/health` },
      { name: 'News Service', url: `${PUBLIC_GATEWAY_URL}/api/v1/news/health` },
      { name: 'Events Service', url: `${PUBLIC_GATEWAY_URL}/api/v1/events/health` },
    ];

    for (const check of healthChecks) {
      try {
        const response = await fetch(check.url, { 
          method: 'GET',
          headers: { 'Accept': 'application/json' }
        });
        
        if (!response.ok) {
          throw new Error(`Health check failed: ${response.status} ${response.statusText}`);
        }
        
        console.log(`✓ ${check.name} health check passed`);
      } catch (error) {
        console.error(`✗ ${check.name} health check failed:`, error);
        throw new Error(`Service ${check.name} is not healthy. Integration tests cannot proceed.`);
      }
    }
  }

  describe('News API Contract Validation', () => {
    it('should get news articles with correct schema', async () => {
      const response = await newsClient.getNews({ page: 1, pageSize: 5 });
      
      // Validate response structure matches expected contract
      expect(response).toHaveProperty('news');
      expect(response).toHaveProperty('count');
      expect(response).toHaveProperty('correlation_id');
      expect(Array.isArray(response.news)).toBe(true);
      expect(typeof response.count).toBe('number');
      expect(typeof response.correlation_id).toBe('string');
      
      // Validate individual news article schema if articles exist
      if (response.news.length > 0) {
        const article = response.news[0];
        expect(article).toHaveProperty('news_id');
        expect(article).toHaveProperty('title');
        expect(article).toHaveProperty('slug');
        expect(article).toHaveProperty('publishing_status');
        expect(typeof article.news_id).toBe('string');
        expect(typeof article.title).toBe('string');
        expect(typeof article.slug).toBe('string');
        expect(article.publishing_status).toBe('published');
      }
    });

    it('should get news article by slug with correct schema', async () => {
      // First get a news article to have a valid slug
      const listResponse = await newsClient.getNews({ page: 1, pageSize: 1 });
      
      if (listResponse.news.length === 0) {
        console.warn('No news articles available for slug test');
        return;
      }
      
      const slug = listResponse.news[0].slug;
      const response = await newsClient.getNewsArticleBySlug(slug);
      
      // Validate response structure
      expect(response).toHaveProperty('news');
      expect(response).toHaveProperty('correlation_id');
      expect(typeof response.correlation_id).toBe('string');
      
      // Validate news article schema
      const article = response.news;
      expect(article).toHaveProperty('news_id');
      expect(article).toHaveProperty('title');
      expect(article).toHaveProperty('slug');
      expect(article).toHaveProperty('publishing_status');
      expect(article.slug).toBe(slug);
      expect(article.publishing_status).toBe('published');
    });

    it('should get news categories with correct schema', async () => {
      const response = await newsClient.getNewsCategories();
      
      // Validate response structure
      expect(response).toHaveProperty('categories');
      expect(response).toHaveProperty('correlation_id');
      expect(Array.isArray(response.categories)).toBe(true);
      expect(typeof response.correlation_id).toBe('string');
      
      // Validate category schema if categories exist
      if (response.categories.length > 0) {
        const category = response.categories[0];
        expect(category).toHaveProperty('category_id');
        expect(category).toHaveProperty('name');
        expect(category).toHaveProperty('slug');
        expect(typeof category.category_id).toBe('string');
        expect(typeof category.name).toBe('string');
        expect(typeof category.slug).toBe('string');
      }
    });

    it('should get featured news with correct schema', async () => {
      const response = await newsClient.getFeaturedNews(3);
      
      // Validate response structure
      expect(response).toHaveProperty('news');
      expect(response).toHaveProperty('count');
      expect(response).toHaveProperty('correlation_id');
      expect(Array.isArray(response.news)).toBe(true);
      
      // All returned news should be featured
      response.news.forEach(article => {
        expect(article.featured).toBe(true);
      });
    });

    it('should handle search queries correctly', async () => {
      const response = await newsClient.searchNews({ 
        q: 'test', 
        page: 1, 
        pageSize: 5 
      });
      
      // Validate response structure
      expect(response).toHaveProperty('news');
      expect(response).toHaveProperty('count');
      expect(response).toHaveProperty('correlation_id');
      expect(Array.isArray(response.news)).toBe(true);
      expect(typeof response.count).toBe('number');
      expect(typeof response.correlation_id).toBe('string');
    });
  });

  describe('Events API Contract Validation', () => {
    it('should get events with correct schema', async () => {
      const response = await eventsClient.getEvents({ page: 1, pageSize: 5 });
      
      // Validate response structure matches expected contract
      expect(response).toHaveProperty('events');
      expect(response).toHaveProperty('count');
      expect(response).toHaveProperty('correlation_id');
      expect(Array.isArray(response.events)).toBe(true);
      expect(typeof response.count).toBe('number');
      expect(typeof response.correlation_id).toBe('string');
      
      // Validate individual event schema if events exist
      if (response.events.length > 0) {
        const event = response.events[0];
        expect(event).toHaveProperty('event_id');
        expect(event).toHaveProperty('title');
        expect(event).toHaveProperty('slug');
        expect(event).toHaveProperty('event_status');
        expect(typeof event.event_id).toBe('string');
        expect(typeof event.title).toBe('string');
        expect(typeof event.slug).toBe('string');
        expect(['upcoming', 'ongoing', 'completed'].includes(event.event_status)).toBe(true);
      }
    });

    it('should get upcoming events with correct filtering', async () => {
      const response = await eventsClient.getUpcomingEvents(5);
      
      // Validate response structure
      expect(response).toHaveProperty('events');
      expect(response).toHaveProperty('count');
      expect(response).toHaveProperty('correlation_id');
      expect(Array.isArray(response.events)).toBe(true);
      
      // All returned events should be upcoming
      response.events.forEach(event => {
        expect(event.event_status).toBe('upcoming');
      });
      
      // Should respect limit
      expect(response.events.length).toBeLessThanOrEqual(5);
    });
  });

  describe('Error Handling Contract Validation', () => {
    it('should handle 404 errors correctly', async () => {
      await expect(newsClient.getNewsArticleBySlug('nonexistent-slug'))
        .rejects
        .toThrow();
    });

    it('should handle invalid search parameters', async () => {
      const response = await newsClient.searchNews({ q: '', page: 1 });
      
      // Should return empty results for empty query
      expect(response.news).toEqual([]);
      expect(response.count).toBe(0);
    });

    it('should validate slug parameter', async () => {
      await expect(newsClient.getNewsArticleBySlug(''))
        .rejects
        .toThrow('News article slug is required');
    });
  });

  describe('Performance and Caching Contract Validation', () => {
    it('should provide performance metrics', () => {
      const metrics = newsClient.getMetrics();
      expect(metrics).toHaveProperty('requests');
      expect(metrics).toHaveProperty('cache_hits');
      expect(metrics).toHaveProperty('cache_misses');
      expect(metrics).toHaveProperty('average_response_time');
    });

    it('should provide cache statistics', () => {
      const stats = newsClient.getCacheStats();
      expect(stats).toHaveProperty('size');
      expect(stats).toHaveProperty('hits');
      expect(stats).toHaveProperty('misses');
      expect(stats).toHaveProperty('hit_rate');
    });

    it('should support cache clearing', () => {
      expect(() => newsClient.clearCache()).not.toThrow();
    });
  });

  describe('Service Integration Health Validation', () => {
    it('should have all services responding to health checks', async () => {
      const healthChecks = [
        `${PUBLIC_GATEWAY_URL}/health`,
        `${PUBLIC_GATEWAY_URL}/api/v1/news/health`,
        `${PUBLIC_GATEWAY_URL}/api/v1/events/health`,
      ];

      for (const url of healthChecks) {
        const response = await fetch(url);
        expect(response.ok).toBe(true);
        expect(response.status).toBe(200);
      }
    });

    it('should have proper CORS configuration', async () => {
      const response = await fetch(`${PUBLIC_GATEWAY_URL}/api/v1/news`, {
        method: 'OPTIONS'
      });
      
      expect(response.headers.get('Access-Control-Allow-Origin')).toBeTruthy();
      expect(response.headers.get('Access-Control-Allow-Methods')).toBeTruthy();
    });
  });
});