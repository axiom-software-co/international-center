<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900">
    <div class="container mx-auto px-4 py-8">
      <!-- Page Header -->
      <div class="mb-8">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-4">
          Search
        </h1>
        <p class="text-gray-600 dark:text-gray-400 mb-6">
          Search across all our services, news articles, research, and events
        </p>

        <!-- Search Input -->
        <div class="max-w-2xl">
          <SearchInput
            v-model="searchQuery"
            placeholder="Search for services, articles, research, events..."
            input-class="text-lg py-3"
            :auto-focus="!initialQuery"
            @search="handleSearch"
            @clear="handleClear"
          />
        </div>
      </div>

      <!-- Search Filters -->
      <div v-if="searchQuery && !loading" class="mb-6 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4">
        <div class="flex flex-wrap items-center gap-4">
          <!-- Sort Options -->
          <div class="flex items-center space-x-2">
            <label class="text-sm font-medium text-gray-700 dark:text-gray-300">Sort by:</label>
            <select
              v-model="sortBy"
              class="text-sm border border-gray-300 dark:border-gray-600 rounded-md px-3 py-1 bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
              @change="handleSortChange"
            >
              <option value="relevance">Relevance</option>
              <option value="date">Date (Newest)</option>
              <option value="title">Title (A-Z)</option>
            </select>
          </div>

          <!-- Content Type Filter -->
          <div class="flex items-center space-x-2">
            <label class="text-sm font-medium text-gray-700 dark:text-gray-300">Type:</label>
            <select
              v-model="contentTypeFilter"
              class="text-sm border border-gray-300 dark:border-gray-600 rounded-md px-3 py-1 bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
              @change="handleFilterChange"
            >
              <option value="">All Types</option>
              <option value="service">Services</option>
              <option value="news">News</option>
              <option value="research">Research</option>
              <option value="events">Events</option>
            </select>
          </div>

          <!-- Results per page -->
          <div class="flex items-center space-x-2">
            <label class="text-sm font-medium text-gray-700 dark:text-gray-300">Show:</label>
            <select
              v-model="pageSize"
              class="text-sm border border-gray-300 dark:border-gray-600 rounded-md px-3 py-1 bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
              @change="handlePageSizeChange"
            >
              <option :value="10">10 results</option>
              <option :value="20">20 results</option>
              <option :value="50">50 results</option>
            </select>
          </div>

          <!-- Clear filters -->
          <button
            v-if="hasActiveFilters"
            class="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-200 font-medium"
            @click="clearFilters"
          >
            Clear all filters
          </button>
        </div>
      </div>

      <!-- Search Results -->
      <SearchResults
        :results="results"
        :total="total"
        :query="searchQuery"
        :loading="loading"
        :page="page"
        :page-size="pageSize"
        :total-pages="totalPages"
        :query-time-ms="queryTimeMs"
        :suggestions="suggestions"
        :facets="facets"
        :show-facets="true"
        @search="handleSearch"
        @page-change="handlePageChange"
        @filter-change="handleFacetFilter"
      />

      <!-- Empty State (no search query) -->
      <div v-if="!searchQuery && !loading" class="text-center py-16">
        <div class="mb-6">
          <SearchIcon class="mx-auto h-16 w-16 text-gray-400" />
        </div>
        <h2 class="text-2xl font-semibold text-gray-900 dark:text-white mb-4">
          Start your search
        </h2>
        <p class="text-gray-600 dark:text-gray-400 mb-8 max-w-lg mx-auto">
          Enter keywords to search across our comprehensive database of services, news articles, research papers, and events.
        </p>

        <!-- Popular Searches or Categories -->
        <div v-if="popularSearches.length > 0" class="max-w-2xl mx-auto">
          <h3 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-4">Popular searches:</h3>
          <div class="flex flex-wrap gap-2 justify-center">
            <button
              v-for="search in popularSearches"
              :key="search"
              class="px-4 py-2 text-sm bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-full hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
              @click="handleSearch(search)"
            >
              {{ search }}
            </button>
          </div>
        </div>
      </div>

      <!-- Error State -->
      <div v-if="error" class="text-center py-12 bg-red-50 dark:bg-red-900/20 rounded-lg border border-red-200 dark:border-red-800">
        <div class="mb-4">
          <svg class="mx-auto h-12 w-12 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <h3 class="text-lg font-medium text-red-900 dark:text-red-200 mb-2">
          Search Error
        </h3>
        <p class="text-red-700 dark:text-red-300 mb-4">
          {{ error }}
        </p>
        <button
          class="px-4 py-2 bg-red-100 dark:bg-red-800 text-red-800 dark:text-red-200 rounded-md hover:bg-red-200 dark:hover:bg-red-700 transition-colors"
          @click="handleRetrySearch"
        >
          Try again
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { SearchIcon } from 'lucide-vue-next';
import SearchInput from './SearchInput.vue';
import SearchResults from './SearchResults.vue';
import { useUnifiedSearch } from '../../composables/search/useUnifiedSearch';
import type { SearchOptions } from '../../lib/services/SearchService';

interface Props {
  initialQuery?: string;
  initialContentType?: string;
}

const props = withDefaults(defineProps<Props>(), {
  initialQuery: '',
  initialContentType: '',
});

// Search state
const searchQuery = ref(props.initialQuery);
const sortBy = ref<'relevance' | 'date' | 'title'>('relevance');
const contentTypeFilter = ref(props.initialContentType);
const queryTimeMs = ref(0);

// Popular searches (could be loaded from an API)
const popularSearches = ref([
  'regenerative therapy',
  'primary care',
  'pain management',
  'wellness programs',
  'clinical research',
  'preventive care'
]);

// Unified search composable
const {
  results,
  total,
  query,
  suggestions,
  facets,
  page,
  pageSize,
  totalPages,
  loading,
  error,
  searchHistory,
  search,
  clearResults,
} = useUnifiedSearch({
  immediate: !!props.initialQuery,
  initialQuery: props.initialQuery,
  initialOptions: {
    types: props.initialContentType ? [props.initialContentType as any] : undefined,
  },
});

// Computed properties
const hasActiveFilters = computed(() => {
  return contentTypeFilter.value || sortBy.value !== 'relevance';
});

// Event handlers
const handleSearch = async (searchTerm?: string): Promise<void> => {
  const queryToSearch = searchTerm || searchQuery.value;
  if (!queryToSearch.trim()) return;

  searchQuery.value = queryToSearch;
  
  const options: SearchOptions = {
    page: 1, // Reset to first page on new search
    pageSize: pageSize.value,
    sortBy: sortBy.value,
    types: contentTypeFilter.value ? [contentTypeFilter.value as any] : undefined,
  };

  const startTime = performance.now();
  await search(queryToSearch, options);
  queryTimeMs.value = Math.round(performance.now() - startTime);
};

const handleClear = (): void => {
  searchQuery.value = '';
  clearResults();
  clearFilters();
};

const handleSortChange = (): void => {
  if (searchQuery.value) {
    handleSearch();
  }
};

const handleFilterChange = (): void => {
  if (searchQuery.value) {
    handleSearch();
  }
};

const handlePageSizeChange = (): void => {
  if (searchQuery.value) {
    handleSearch();
  }
};

const handlePageChange = async (newPage: number): Promise<void> => {
  if (!searchQuery.value) return;
  
  const options: SearchOptions = {
    page: newPage,
    pageSize: pageSize.value,
    sortBy: sortBy.value,
    types: contentTypeFilter.value ? [contentTypeFilter.value as any] : undefined,
  };

  await search(searchQuery.value, options);
};

const handleFacetFilter = async (filters: { types: string[]; categories: string[] }): Promise<void> => {
  if (!searchQuery.value) return;
  
  const options: SearchOptions = {
    page: 1, // Reset to first page when filtering
    pageSize: pageSize.value,
    sortBy: sortBy.value,
    types: filters.types.length > 0 ? filters.types as any : undefined,
    categories: filters.categories.length > 0 ? filters.categories : undefined,
  };

  await search(searchQuery.value, options);
};

const clearFilters = (): void => {
  sortBy.value = 'relevance';
  contentTypeFilter.value = '';
  
  if (searchQuery.value) {
    handleSearch();
  }
};

const handleRetrySearch = (): void => {
  if (searchQuery.value) {
    handleSearch();
  }
};

// Watch for URL changes (if using router)
watch(() => props.initialQuery, (newQuery) => {
  if (newQuery && newQuery !== searchQuery.value) {
    searchQuery.value = newQuery;
    handleSearch();
  }
});

watch(() => props.initialContentType, (newType) => {
  contentTypeFilter.value = newType;
});

// Initialize search on mount if we have an initial query
onMounted(() => {
  if (props.initialQuery) {
    handleSearch();
  }
});
</script>