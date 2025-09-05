import { defineStore } from 'pinia';
import { newsClient } from '../lib/clients';
import type { 
  NewsArticle, 
  NewsCategory, 
  GetNewsParams, 
  SearchNewsParams 
} from '../lib/clients/news/types';
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
    async fetchNews(params?: GetNewsParams, options: CacheOptions = {}): Promise<void> {
      await withCachedApiAction(
        this,
        params,
        options,
        () => newsClient.getNews(params),
        (response) => this.setNews(
          response.news, 
          response.count, 
          params?.page || 1, 
          params?.pageSize || 10
        ),
        () => this.setNews([], 0, 1, 10),
        'Failed to fetch news'
      );
    },

    async fetchNewsArticle(slug: string): Promise<NewsArticle | null> {
      const result = await withApiAction(
        this,
        () => newsClient.getNewsArticleBySlug(slug),
        'Failed to fetch news article'
      );
      return result?.news || null;
    },

    async fetchFeaturedNews(limit?: number): Promise<void> {
      const result = await withApiAction(
        this,
        () => newsClient.getFeaturedNews(limit),
        'Failed to fetch featured news'
      );
      this.setFeaturedNews(result?.news || []);
    },

    async searchNews(params: SearchNewsParams): Promise<void> {
      // Handle empty search queries
      if (handleEmptySearch(params.q, this.setSearchResults)) {
        return;
      }

      const result = await withApiAction(
        this,
        () => newsClient.searchNews(params),
        'Failed to search news'
      );
      this.setSearchResults(result?.news || [], result?.count || 0);
    },

    async fetchNewsCategories(): Promise<void> {
      const result = await withApiAction(
        this,
        () => newsClient.getNewsCategories(),
        'Failed to fetch news categories'
      );
      this.setCategories(result?.categories || []);
    },
  } satisfies NewsStoreActions,
});