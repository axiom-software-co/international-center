import { defineStore } from 'pinia';
import { apiClient } from '../lib/api-client';
import type { 
  ResearchPublication, 
  ResearchCategory,
  GetResearchRequest
} from '@international-center/public-api-client';
import type { 
  ResearchStoreState, 
  ResearchStoreActions, 
  ResearchStoreGetters,
  CacheOptions 
} from './interfaces';
import {
  createBaseState,
  createBaseGetters,
  createCacheActions,
  createStateActions,
  createDomainStateSetters,
  createDomainGetters,
  createGroupingGetter,
  withCachedApiAction,
  withApiAction,
  handleEmptySearch,
  CACHE_TIMEOUT
} from './base';

export const useResearchStore = defineStore('research', {
  state: (): ResearchStoreState => ({
    research: [],
    article: null,
    categories: [],
    featuredResearch: [],
    searchResults: [],
    ...createBaseState(),
  }),

  getters: {
    ...createBaseGetters(),
    ...createDomainGetters<ResearchArticle>('research'),

    researchByType: createGroupingGetter<ResearchArticle>('research', 'research_type'),

    researchByPrimaryAuthor(): Record<string, ResearchArticle[]> {
      return this.research.reduce((acc, article) => {
        // Extract primary author (first author in the author_names string)
        const primaryAuthor = article.author_names.split(',')[0].trim();
        if (!acc[primaryAuthor]) {
          acc[primaryAuthor] = [];
        }
        acc[primaryAuthor].push(article);
        return acc;
      }, {} as Record<string, ResearchArticle[]>);
    },

    recentResearch(): ResearchArticle[] {
      return [...this.research]
        .sort((a, b) => {
          const dateA = a.publication_date || a.created_on;
          const dateB = b.publication_date || b.created_on;
          return new Date(dateB).getTime() - new Date(dateA).getTime();
        })
        .slice(0, 10);
    },
  } satisfies ResearchStoreGetters,

  actions: {
    // Base functionality
    ...createStateActions(),
    ...createCacheActions('research'),
    ...createDomainStateSetters<ResearchArticle, ResearchCategory>('research', 'categories', 'featuredResearch'),

    // Domain-specific state setters

    // API Actions
    async fetchResearch(params?: GetResearchRequest, options: CacheOptions = {}): Promise<void> {
      await withCachedApiAction(
        this,
        params,
        options,
        () => apiClient.getResearch({
          page: params?.page || 1,
          limit: params?.limit || 20,
          search: params?.search,
          categoryId: params?.categoryId
        }),
        (response) => this.setResearch(
          response.data || [], 
          response.pagination?.total_items || 0, 
          params?.page || 1, 
          params?.limit || 20
        ),
        (items, count) => this.setResearch(items, count, 1, 20),
        'Failed to fetch research via contract client'
      );
    },

    async fetchResearchArticle(slug: string): Promise<ResearchPublication | null> {
      const result = await withApiAction(
        this,
        () => apiClient.getResearchById(slug), // Using ID for now - slug lookup would need API extension
        'Failed to fetch research article via contract client'
      );
      this.article = result?.data || null;
      return this.article;
    },

    async fetchFeaturedResearch(limit?: number): Promise<void> {
      const result = await withApiAction(
        this,
        () => apiClient.getFeaturedResearch(),
        'Failed to fetch featured research via contract client'
      );
      this.setFeaturedResearch(result?.data?.slice(0, limit) || []);
    },

    async searchResearch(params: { q: string, page?: number, limit?: number }): Promise<void> {
      // Handle empty search queries
      if (handleEmptySearch(params.q, this.setSearchResults)) {
        return;
      }

      const result = await withApiAction(
        this,
        () => apiClient.getResearch({
          page: params.page || 1,
          limit: params.limit || 20,
          search: params.q
        }),
        'Failed to search research via contract client'
      );
      this.setSearchResults(result?.data || [], result?.pagination?.total_items || 0);
    },

    async fetchResearchCategories(): Promise<void> {
      const result = await withApiAction(
        this,
        () => apiClient.getResearchCategories(),
        'Failed to fetch research categories via contract client'
      );
      this.setCategories(result?.data || []);
    },
  } satisfies ResearchStoreActions,
});