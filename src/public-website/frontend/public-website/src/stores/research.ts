import { defineStore } from 'pinia';
import { researchClient } from '../lib/clients';
import type { 
  ResearchArticle, 
  ResearchCategory, 
  GetResearchParams, 
  SearchResearchParams 
} from '../lib/clients/research/types';
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
    async fetchResearch(params?: GetResearchParams, options: CacheOptions = {}): Promise<void> {
      await withCachedApiAction(
        this,
        params,
        options,
        () => researchClient.getResearch(params),
        (response) => this.setResearch(
          response.research, 
          response.count, 
          params?.page || 1, 
          params?.pageSize || 10
        ),
        (items, count) => this.setResearch(items, count, 1, 10),
        'Failed to fetch research'
      );
    },

    async fetchResearchArticle(slug: string): Promise<ResearchArticle | null> {
      const result = await withApiAction(
        this,
        () => researchClient.getResearchBySlug(slug),
        'Failed to fetch research article'
      );
      this.article = result?.research || null;
      return this.article;
    },

    async fetchFeaturedResearch(limit?: number): Promise<void> {
      const result = await withApiAction(
        this,
        () => researchClient.getFeaturedResearch(limit),
        'Failed to fetch featured research'
      );
      this.setFeaturedResearch(result?.research || []);
    },

    async searchResearch(params: SearchResearchParams): Promise<void> {
      // Handle empty search queries
      if (handleEmptySearch(params.q, this.setSearchResults)) {
        return;
      }

      const result = await withApiAction(
        this,
        () => researchClient.searchResearch(params),
        'Failed to search research'
      );
      this.setSearchResults(result?.research || [], result?.count || 0);
    },

    async fetchResearchCategories(): Promise<void> {
      const result = await withApiAction(
        this,
        () => researchClient.getResearchCategories(),
        'Failed to fetch research categories'
      );
      this.setCategories(result?.categories || []);
    },
  } satisfies ResearchStoreActions,
});