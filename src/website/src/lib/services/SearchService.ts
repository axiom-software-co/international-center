// Search Service - Unified search functionality
// Uses the new search-domain microservice for optimal performance

import { searchClient } from '../clients/search/SearchClient';
import type { SearchResult as DomainSearchResult, SearchResponse } from '../clients/search/types';

// Legacy SearchResult interface for backward compatibility
export interface SearchResult {
  id: string;
  title: string;
  slug: string;
  excerpt: string;
  image: string;
  type: 'service' | 'news' | 'research';
  category?: string;
  published_at?: string;
  relevanceScore: number;
}

export interface SearchOptions {
  page?: number;
  pageSize?: number;
  types?: Array<'service' | 'news' | 'research'>;
  categories?: string[];
  sortBy?: 'relevance' | 'date' | 'title';
}

export interface UnifiedSearchResponse {
  results: SearchResult[];
  total: number;
  query: string;
  suggestions?: string[];
  facets: {
    types: Array<{ type: string; count: number }>;
    categories: Array<{ category: string; count: number }>;
  };
}

export class SearchService {
  private searchHistory: string[] = [];
  private readonly MAX_HISTORY = 10;

  /**
   * Perform unified search using the search-domain microservice
   */
  async search(query: string, options: SearchOptions = {}): Promise<UnifiedSearchResponse> {
    if (!query.trim()) {
      return this.getEmptySearchResponse(query);
    }

    // Add to search history
    this.addToSearchHistory(query);

    const {
      page = 1,
      pageSize = 20,
      types = ['service', 'news', 'research'],
      categories = [],
      sortBy = 'relevance',
    } = options;

    try {
      // Map sortBy to search domain format
      const domainSortBy = this.mapSortBy(sortBy);

      // Use unified search endpoint
      const response = await searchClient.search({
        q: query,
        page,
        pageSize,
        sortBy: domainSortBy,
      });

      // Transform search results to legacy format for compatibility
      const transformedResults = response.results
        .filter(result => types.includes(result.content_type))
        .filter(result => categories.length === 0 || categories.includes(result.category))
        .map(result => this.transformSearchResult(result));

      // Generate facets from response
      const facets = this.generateFacets(response.facets || []);

      return {
        results: transformedResults,
        total: response.total,
        query,
        suggestions: await this.generateSuggestions(query),
        facets,
      };
    } catch (error) {
      console.error('Search error:', error);
      return this.getEmptySearchResponse(query);
    }
  }

  /**
   * Quick search with minimal parameters
   */
  async quickSearch(query: string, limit: number = 10): Promise<SearchResult[]> {
    const response = await this.search(query, { pageSize: limit });
    return response.results;
  }

  /**
   * Search within specific content type
   */
  async searchByType(
    query: string,
    type: 'service' | 'news' | 'research',
    options?: { limit?: number; category?: string }
  ): Promise<SearchResult[]> {
    const response = await searchClient.searchByType(query, type, {
      pageSize: options?.limit || 10,
      category: options?.category,
    });

    return response.results.map(result => this.transformSearchResult(result));
  }

  /**
   * Get search suggestions
   */
  async getSuggestions(query: string, limit: number = 5): Promise<string[]> {
    return searchClient.getSuggestions(query, limit);
  }

  /**
   * Get search history
   */
  getSearchHistory(): string[] {
    return [...this.searchHistory];
  }

  /**
   * Clear search history
   */
  clearSearchHistory(): void {
    this.searchHistory = [];
  }

  /**
   * Refresh search index (admin function)
   */
  async refreshIndex(): Promise<void> {
    await searchClient.refreshIndex();
  }

  // Private helper methods

  /**
   * Transform domain search result to legacy format
   */
  private transformSearchResult(result: DomainSearchResult): SearchResult {
    return {
      id: result.id,
      title: result.title,
      slug: result.slug,
      excerpt: result.excerpt,
      image: '', // Not available in search results, would need to be enriched
      type: result.content_type,
      category: result.category,
      published_at: result.created_at,
      relevanceScore: result.rank,
    };
  }

  /**
   * Map SearchService sortBy to search domain format
   */
  private mapSortBy(
    sortBy: string
  ): 'relevance' | 'date-desc' | 'date-asc' | 'title-asc' | 'title-desc' {
    switch (sortBy) {
      case 'date':
        return 'date-desc';
      case 'title':
        return 'title-asc';
      case 'relevance':
      default:
        return 'relevance';
    }
  }

  /**
   * Generate search suggestions based on query
   */
  private async generateSuggestions(query: string): Promise<string[]> {
    const suggestions = new Set<string>();

    // Add related terms
    const relatedTerms = this.getRelatedTerms(query);
    relatedTerms.forEach(term => suggestions.add(term));

    // Add suggestions from search client
    try {
      const clientSuggestions = await searchClient.getSuggestions(query);
      clientSuggestions.forEach(suggestion => suggestions.add(suggestion));
    } catch (error) {
      console.error('Error getting client suggestions:', error);
    }

    return Array.from(suggestions).slice(0, 5);
  }

  /**
   * Generate facets for filtering
   */
  private generateFacets(
    searchFacets: Array<{ name: string; values: Array<{ value: string; count: number }> }>
  ): {
    types: Array<{ type: string; count: number }>;
    categories: Array<{ category: string; count: number }>;
  } {
    let types: Array<{ type: string; count: number }> = [];
    let categories: Array<{ category: string; count: number }> = [];

    searchFacets.forEach(facet => {
      if (facet.name === 'content_type') {
        types = facet.values.map(value => ({ type: value.value, count: value.count }));
      } else if (facet.name === 'category') {
        categories = facet.values.map(value => ({ category: value.value, count: value.count }));
      }
    });

    return { types, categories };
  }

  /**
   * Get related search terms
   */
  private getRelatedTerms(query: string): string[] {
    const termMap: Record<string, string[]> = {
      therapy: ['treatment', 'healing', 'medicine'],
      treatment: ['therapy', 'procedure', 'care'],
      health: ['wellness', 'medical', 'care'],
      research: ['study', 'science', 'innovation'],
      innovation: ['technology', 'advancement', 'breakthrough'],
    };

    const relatedTerms: string[] = [];
    const queryLower = query.toLowerCase();

    Object.entries(termMap).forEach(([key, related]) => {
      if (queryLower.includes(key)) {
        relatedTerms.push(...related);
      }
    });

    return relatedTerms;
  }

  /**
   * Add query to search history
   */
  private addToSearchHistory(query: string): void {
    const trimmedQuery = query.trim();
    if (!trimmedQuery) return;

    // Remove existing entry if it exists
    const existingIndex = this.searchHistory.indexOf(trimmedQuery);
    if (existingIndex > -1) {
      this.searchHistory.splice(existingIndex, 1);
    }

    // Add to beginning
    this.searchHistory.unshift(trimmedQuery);

    // Limit history size
    if (this.searchHistory.length > this.MAX_HISTORY) {
      this.searchHistory = this.searchHistory.slice(0, this.MAX_HISTORY);
    }
  }

  /**
   * Get empty search response
   */
  private getEmptySearchResponse(query: string): UnifiedSearchResponse {
    return {
      results: [],
      total: 0,
      query,
      facets: { types: [], categories: [] },
    };
  }
}

// Export singleton instance
export const searchService = new SearchService();
