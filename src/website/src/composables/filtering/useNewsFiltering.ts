// News-specific filtering composable
// Integrates news data with generic filtering system

import { ref, computed, watch, type Ref } from 'vue';
import { useNews } from '../useNews';
import { useFiltering, useCategoryFiltering, type FilterConfig, type FilterOption } from './useFiltering';
import type { NewsArticle } from '../../lib/clients/news/types';

export interface UseNewsFilteringOptions {
  enableCategoryFilter?: boolean;
  enableStatusFilter?: boolean;
  includeCounts?: boolean;
}

export interface UseNewsFilteringResult {
  // Filtered news data
  filteredNews: Ref<NewsArticle[]>;
  
  // Filter state
  filters: Ref<{ [key: string]: string | string[] }>;
  hasActiveFilters: Ref<boolean>;
  
  // Filter configurations for UI
  categoryFilterConfig: Ref<FilterConfig | null>;
  statusFilterConfig: Ref<FilterConfig | null>;
  
  // Filter actions
  setCategoryFilter: (categoryId: string) => void;
  setStatusFilter: (status: string) => void;
  clearAllFilters: () => void;
  
  // Utility methods
  isFilterActive: (filterName: string, value?: string) => boolean;
}

// Publishing status labels mapping
const STATUS_LABELS = {
  published: 'Published',
  draft: 'Draft',
  archived: 'Archived',
} as const;

export function useNewsFiltering(
  newsOptions: Parameters<typeof useNews>[0] = {},
  filterOptions: UseNewsFilteringOptions = {}
): UseNewsFilteringResult {

  const {
    enableCategoryFilter = true,
    enableStatusFilter = true,
    includeCounts = true,
  } = filterOptions;

  // Load news data
  const { news, loading: newsLoading } = useNews(newsOptions);

  // Create filter configurations
  const filterConfigs: FilterConfig[] = [];

  // Category filter configuration - placeholder since news doesn't have categories in current schema
  const categoryFilterConfig = computed(() => {
    if (!enableCategoryFilter) return null;
    
    // Placeholder - would be implemented when news categories are added
    return null;
  });

  // Publishing status filter configuration  
  const statusFilterConfig = computed(() => {
    if (!enableStatusFilter || newsLoading.value) return null;

    const statusOptions: FilterOption[] = Object.entries(STATUS_LABELS).map(([status, label]) => ({
      value: status,
      label,
      count: includeCounts ? countNewsByStatus(news.value, status as keyof typeof STATUS_LABELS) : undefined,
    }));

    return {
      name: 'publishing_status',
      label: 'Status',
      options: statusOptions,
      multiple: false,
      searchable: false,
    };
  });

  // Build dynamic filter configs
  watch([categoryFilterConfig, statusFilterConfig], ([categoryConfig, statusConfig]) => {
    filterConfigs.length = 0; // Clear existing configs
    
    if (categoryConfig) {
      filterConfigs.push(categoryConfig);
    }
    if (statusConfig) {
      filterConfigs.push(statusConfig);
    }
  }, { immediate: true });

  // Initialize filtering system
  const filteringResult = useFiltering(news, filterConfigs);

  // Convenient filter setters
  const setCategoryFilter = (categoryId: string): void => {
    filteringResult.setFilter('category', categoryId);
  };

  const setStatusFilter = (status: string): void => {
    filteringResult.setFilter('publishing_status', status);
  };

  return {
    filteredNews: filteringResult.filteredItems,
    filters: filteringResult.filters,
    hasActiveFilters: filteringResult.hasActiveFilters,
    categoryFilterConfig,
    statusFilterConfig,
    setCategoryFilter,
    setStatusFilter,
    clearAllFilters: filteringResult.clearAllFilters,
    isFilterActive: filteringResult.isFilterActive,
  };
}

// Helper functions
function countNewsByStatus(news: NewsArticle[], status: string): number {
  return news.filter(article => article.publishing_status === status).length;
}

// Status-only filtering composable for simpler use cases
export function useNewsStatusFiltering(
  newsOptions: Parameters<typeof useNews>[0] = {}
): UseNewsFilteringResult {
  return useNewsFiltering(newsOptions, {
    enableCategoryFilter: false,
    enableStatusFilter: true,
    includeCounts: true,
  });
}