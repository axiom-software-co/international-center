import { defineStore } from 'pinia';
import { apiClient } from '../lib/api-client';
import type { 
  NewsArticle, 
  NewsCategory,
  GetNewsRequest
} from '@international-center/public-api-client';
import type { 
  NewsStoreState, 
  NewsStoreActions, 
  NewsStoreGetters,
  CacheOptions 
} from './interfaces';
import {
  createBaseState,
  createBaseGetters,
  createCacheActions,
  createStateActions,
  createDomainStateSetters,
  createDomainGetters,
  withCachedApiAction,
  withApiAction,
  handleEmptySearch,
  CACHE_TIMEOUT
} from './base';

export const useNewsStore = defineStore('news', {
  state: (): NewsStoreState => ({
    news: [],
    article: null,
    categories: [],
    featuredNews: [],
    searchResults: [],
    ...createBaseState(),
  }),

  getters: {
    ...createBaseGetters(),
    ...createDomainGetters<NewsArticle>('news'),

    recentNews(): NewsArticle[] {
      return [...this.news]
        .sort((a, b) => 
          new Date(b.publication_timestamp).getTime() - 
          new Date(a.publication_timestamp).getTime()
        )
        .slice(0, 10);
    },
  } satisfies NewsStoreGetters,

  actions: {
    // Base functionality
    ...createStateActions(),
    ...createCacheActions('news'),

    ...createDomainStateSetters<NewsArticle, NewsCategory>('news', 'categories', 'featuredNews'),

    // Domain-specific state setters

    // API Actions
    async fetchNews(params?: GetNewsRequest, options: CacheOptions = {}): Promise<void> {
      await withCachedApiAction(
        this,
        params,
        options,
        () => apiClient.getNews({
          page: params?.page || 1,
          limit: params?.limit || 20,
          search: params?.search,
          categoryId: params?.categoryId
        }),
        (response) => this.setNews(
          response.data || [], 
          response.pagination?.total_items || 0, 
          params?.page || 1, 
          params?.limit || 20
        ),
        (items, count) => this.setNews(items, count, 1, 20),
        'Failed to fetch news via contract client'
      );
    },

    async fetchNewsArticle(slug: string): Promise<NewsArticle | null> {
      const result = await withApiAction(
        this,
        () => apiClient.getNewsById(slug), // Using ID for now - slug lookup would need API extension
        'Failed to fetch news article via contract client'
      );
      this.article = result?.data || null;
      return this.article;
    },

    async fetchFeaturedNews(limit?: number): Promise<void> {
      const result = await withApiAction(
        this,
        () => apiClient.getFeaturedNews(),
        'Failed to fetch featured news via contract client'
      );
      this.setFeaturedNews(result?.data?.slice(0, limit) || []);
    },

    async searchNews(params: { q: string, page?: number, limit?: number }): Promise<void> {
      // Handle empty search queries
      if (handleEmptySearch(params.q, this.setSearchResults)) {
        return;
      }

      const result = await withApiAction(
        this,
        () => apiClient.getNews({
          page: params.page || 1,
          limit: params.limit || 20,
          search: params.q
        }),
        'Failed to search news via contract client'
      );
      this.setSearchResults(result?.data || [], result?.pagination?.total_items || 0);
    },

    async fetchNewsCategories(): Promise<void> {
      const result = await withApiAction(
        this,
        () => apiClient.getNewsCategories(),
        'Failed to fetch news categories via contract client'
      );
      this.setCategories(result?.data || []);
    },
  } satisfies NewsStoreActions,
});