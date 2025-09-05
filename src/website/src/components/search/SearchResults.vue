<template>
  <div class="search-results">
    <!-- Search Stats -->
    <div v-if="total > 0" class="mb-4 text-sm text-gray-600 dark:text-gray-400">
      Found {{ total.toLocaleString() }} results for "<span class="font-medium">{{ query }}</span>"
      <span v-if="queryTimeMs">({{ queryTimeMs }}ms)</span>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="space-y-4">
      <div v-for="i in 5" :key="i" class="animate-pulse">
        <div class="flex space-x-4">
          <div class="bg-gray-300 dark:bg-gray-700 rounded w-16 h-16"></div>
          <div class="flex-1 space-y-2">
            <div class="h-4 bg-gray-300 dark:bg-gray-700 rounded w-3/4"></div>
            <div class="h-3 bg-gray-300 dark:bg-gray-700 rounded w-1/2"></div>
            <div class="h-3 bg-gray-300 dark:bg-gray-700 rounded w-full"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- No Results -->
    <div v-else-if="!loading && results.length === 0 && query" class="text-center py-12">
      <div class="mb-4">
        <SearchIcon class="mx-auto h-12 w-12 text-gray-400" />
      </div>
      <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">
        No results found
      </h3>
      <p class="text-gray-600 dark:text-gray-400 mb-4">
        Try adjusting your search terms or filters
      </p>
      <div v-if="suggestions.length > 0" class="text-sm">
        <p class="text-gray-600 dark:text-gray-400 mb-2">Did you mean:</p>
        <div class="space-x-2">
          <button
            v-for="suggestion in suggestions"
            :key="suggestion"
            class="inline-block px-3 py-1 text-sm bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-full hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
            @click="$emit('search', suggestion)"
          >
            {{ suggestion }}
          </button>
        </div>
      </div>
    </div>

    <!-- Results List -->
    <div v-else-if="!loading && results.length > 0" class="space-y-6">
      <!-- Facets/Filters -->
      <div v-if="showFacets && (facets.types.length > 0 || facets.categories.length > 0)" class="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
        <h4 class="text-sm font-medium text-gray-900 dark:text-white mb-3">Filter Results</h4>
        
        <!-- Content Types Filter -->
        <div v-if="facets.types.length > 0" class="mb-4">
          <h5 class="text-xs font-medium text-gray-700 dark:text-gray-300 mb-2">Content Type</h5>
          <div class="space-x-2">
            <button
              v-for="type in facets.types"
              :key="type.type"
              :class="[
                'inline-flex items-center px-3 py-1 text-sm rounded-full transition-colors',
                selectedTypes.includes(type.type)
                  ? 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600'
              ]"
              @click="toggleTypeFilter(type.type)"
            >
              {{ formatContentType(type.type) }}
              <span class="ml-1 text-xs">{{ type.count }}</span>
            </button>
          </div>
        </div>

        <!-- Categories Filter -->
        <div v-if="facets.categories.length > 0">
          <h5 class="text-xs font-medium text-gray-700 dark:text-gray-300 mb-2">Category</h5>
          <div class="space-x-2">
            <button
              v-for="category in facets.categories.slice(0, 5)"
              :key="category.category"
              :class="[
                'inline-flex items-center px-3 py-1 text-sm rounded-full transition-colors',
                selectedCategories.includes(category.category)
                  ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600'
              ]"
              @click="toggleCategoryFilter(category.category)"
            >
              {{ category.category }}
              <span class="ml-1 text-xs">{{ category.count }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Search Results -->
      <div class="grid gap-6">
        <article
          v-for="result in results"
          :key="result.id"
          class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6 hover:shadow-md transition-shadow"
        >
          <!-- Content Type Badge -->
          <div class="flex items-center justify-between mb-3">
            <span :class="[
              'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
              getContentTypeBadgeClass(result.type)
            ]">
              <component :is="getContentTypeIcon(result.type)" class="w-3 h-3 mr-1" />
              {{ formatContentType(result.type) }}
            </span>
            
            <!-- Relevance Score (debug mode) -->
            <span v-if="showRelevanceScore" class="text-xs text-gray-500">
              Score: {{ result.relevanceScore.toFixed(2) }}
            </span>
          </div>

          <!-- Title -->
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2 hover:text-blue-600 dark:hover:text-blue-400 transition-colors">
            <a :href="getResultUrl(result)" class="block">
              {{ result.title }}
            </a>
          </h3>

          <!-- Excerpt -->
          <p class="text-gray-600 dark:text-gray-400 text-sm leading-relaxed mb-3">
            {{ result.excerpt }}
          </p>

          <!-- Metadata -->
          <div class="flex items-center text-xs text-gray-500 dark:text-gray-400 space-x-4">
            <div v-if="result.category" class="flex items-center">
              <TagIcon class="w-3 h-3 mr-1" />
              {{ result.category }}
            </div>
            <div v-if="result.published_at" class="flex items-center">
              <CalendarIcon class="w-3 h-3 mr-1" />
              {{ formatDate(result.published_at) }}
            </div>
          </div>
        </article>
      </div>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="flex items-center justify-between border-t border-gray-200 dark:border-gray-700 pt-6">
        <div class="flex flex-1 justify-between sm:hidden">
          <button
            :disabled="page <= 1"
            :class="[
              'relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md',
              page <= 1
                ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                : 'bg-white text-gray-700 hover:bg-gray-50'
            ]"
            @click="$emit('page-change', page - 1)"
          >
            Previous
          </button>
          <button
            :disabled="page >= totalPages"
            :class="[
              'relative ml-3 inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md',
              page >= totalPages
                ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                : 'bg-white text-gray-700 hover:bg-gray-50'
            ]"
            @click="$emit('page-change', page + 1)"
          >
            Next
          </button>
        </div>
        
        <div class="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
          <p class="text-sm text-gray-700 dark:text-gray-300">
            Showing {{ ((page - 1) * pageSize) + 1 }} to {{ Math.min(page * pageSize, total) }} of {{ total }} results
          </p>
          
          <nav class="relative z-0 inline-flex rounded-md shadow-sm -space-x-px">
            <button
              :disabled="page <= 1"
              :class="[
                'relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 text-sm font-medium',
                page <= 1
                  ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                  : 'bg-white text-gray-500 hover:bg-gray-50'
              ]"
              @click="$emit('page-change', page - 1)"
            >
              <ChevronLeftIcon class="h-5 w-5" />
            </button>
            
            <!-- Page Numbers -->
            <template v-for="pageNum in getVisiblePageNumbers()" :key="pageNum">
              <button
                v-if="pageNum !== '...'"
                :class="[
                  'relative inline-flex items-center px-4 py-2 border text-sm font-medium',
                  pageNum === page
                    ? 'z-10 bg-blue-50 border-blue-500 text-blue-600'
                    : 'bg-white border-gray-300 text-gray-500 hover:bg-gray-50'
                ]"
                @click="$emit('page-change', pageNum)"
              >
                {{ pageNum }}
              </button>
              <span v-else class="relative inline-flex items-center px-4 py-2 border border-gray-300 bg-white text-sm font-medium text-gray-700">
                ...
              </span>
            </template>
            
            <button
              :disabled="page >= totalPages"
              :class="[
                'relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 text-sm font-medium',
                page >= totalPages
                  ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                  : 'bg-white text-gray-500 hover:bg-gray-50'
              ]"
              @click="$emit('page-change', page + 1)"
            >
              <ChevronRightIcon class="h-5 w-5" />
            </button>
          </nav>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { 
  SearchIcon, 
  TagIcon, 
  CalendarIcon, 
  ChevronLeftIcon, 
  ChevronRightIcon,
  FileTextIcon,
  NewspaperIcon,
  BeakerIcon,
  HeartHandshakeIcon
} from 'lucide-vue-next';
import type { SearchResult } from '../../lib/services/SearchService';

interface Props {
  results: SearchResult[];
  total: number;
  query: string;
  loading: boolean;
  page: number;
  pageSize: number;
  totalPages: number;
  queryTimeMs?: number;
  suggestions?: string[];
  facets: {
    types: Array<{ type: string; count: number }>;
    categories: Array<{ category: string; count: number }>;
  };
  showFacets?: boolean;
  showRelevanceScore?: boolean;
}

interface Emits {
  (e: 'search', query: string): void;
  (e: 'page-change', page: number): void;
  (e: 'filter-change', filters: { types: string[]; categories: string[] }): void;
}

const props = withDefaults(defineProps<Props>(), {
  results: () => [],
  total: 0,
  query: '',
  loading: false,
  page: 1,
  pageSize: 20,
  totalPages: 0,
  queryTimeMs: 0,
  suggestions: () => [],
  facets: () => ({ types: [], categories: [] }),
  showFacets: true,
  showRelevanceScore: false,
});

const emit = defineEmits<Emits>();

// Selected filters
const selectedTypes = ref<string[]>([]);
const selectedCategories = ref<string[]>([]);

// Helper functions
const formatContentType = (type: string): string => {
  const typeMap: Record<string, string> = {
    service: 'Service',
    news: 'News',
    research: 'Research',
    events: 'Events',
  };
  return typeMap[type] || type;
};

const getContentTypeIcon = (type: string) => {
  const iconMap: Record<string, any> = {
    service: HeartHandshakeIcon,
    news: NewspaperIcon,
    research: BeakerIcon,
    events: CalendarIcon,
  };
  return iconMap[type] || FileTextIcon;
};

const getContentTypeBadgeClass = (type: string): string => {
  const classMap: Record<string, string> = {
    service: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
    news: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
    research: 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200',
    events: 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200',
  };
  return classMap[type] || 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200';
};

const getResultUrl = (result: SearchResult): string => {
  const baseUrls: Record<string, string> = {
    service: '/services',
    news: '/company/news',
    research: '/community/research',
    events: '/community/events',
  };
  return `${baseUrls[result.type] || ''}/${result.slug}`;
};

const formatDate = (dateString: string): string => {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', { 
    year: 'numeric', 
    month: 'short', 
    day: 'numeric' 
  });
};

const getVisiblePageNumbers = (): (number | string)[] => {
  const pages: (number | string)[] = [];
  const currentPage = props.page;
  const totalPages = props.totalPages;
  
  if (totalPages <= 7) {
    // Show all pages if total is 7 or less
    for (let i = 1; i <= totalPages; i++) {
      pages.push(i);
    }
  } else {
    // Always show first page
    pages.push(1);
    
    if (currentPage <= 4) {
      // Show pages 2-5 and ellipsis
      for (let i = 2; i <= 5; i++) {
        pages.push(i);
      }
      pages.push('...');
    } else if (currentPage >= totalPages - 3) {
      // Show ellipsis and last 4 pages
      pages.push('...');
      for (let i = totalPages - 4; i < totalPages; i++) {
        pages.push(i);
      }
    } else {
      // Show ellipsis, current page area, and ellipsis
      pages.push('...');
      for (let i = currentPage - 1; i <= currentPage + 1; i++) {
        pages.push(i);
      }
      pages.push('...');
    }
    
    // Always show last page (if not already shown)
    if (totalPages > 1) {
      pages.push(totalPages);
    }
  }
  
  return pages;
};

// Filter handlers
const toggleTypeFilter = (type: string): void => {
  const index = selectedTypes.value.indexOf(type);
  if (index > -1) {
    selectedTypes.value.splice(index, 1);
  } else {
    selectedTypes.value.push(type);
  }
  
  emit('filter-change', {
    types: selectedTypes.value,
    categories: selectedCategories.value,
  });
};

const toggleCategoryFilter = (category: string): void => {
  const index = selectedCategories.value.indexOf(category);
  if (index > -1) {
    selectedCategories.value.splice(index, 1);
  } else {
    selectedCategories.value.push(category);
  }
  
  emit('filter-change', {
    types: selectedTypes.value,
    categories: selectedCategories.value,
  });
};
</script>