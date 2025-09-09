// Search Components and Composables - Complete search functionality
// Export all search-related components and composables for easy access

// Components
export { default as SearchInput } from './SearchInput.vue';
export { default as SearchResults } from './SearchResults.vue';
export { default as SearchPage } from './SearchPage.vue';

// Composables
export {
  useUnifiedSearch,
  useQuickSearch,
  useSearchHistory,
  type UseUnifiedSearchResult,
  type UseUnifiedSearchOptions,
} from '../../composables/search/useUnifiedSearch';

// Services
export {
  searchService,
  type SearchResult,
  type SearchOptions,
  type UnifiedSearchResponse,
} from '../../lib/services/SearchService';

// Client
export { searchClient } from '../../lib/clients/search/SearchClient';

// Types
export type {
  SearchResult as DomainSearchResult,
  SearchResponse,
  SearchParams,
  SearchFacet,
  SearchFacetValue,
  UnifiedSearchResponse as DomainUnifiedSearchResponse,
} from '../../lib/clients/search/types';