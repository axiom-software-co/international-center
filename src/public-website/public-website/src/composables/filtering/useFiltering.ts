// Generic Filtering Composable - Reactive filtering for all domains
// Provides consistent filtering patterns across Services, News, Events, Research

import { ref, computed, watch, type Ref } from 'vue';
import type { BaseDomainStore } from '../base';

export interface FilterOption {
  value: string;
  label: string;
  count?: number;
}

export interface FilterConfig {
  name: string;
  label: string;
  options: FilterOption[];
  multiple?: boolean;
  searchable?: boolean;
}

export interface FilterState {
  [filterName: string]: string[] | string;
}

export interface UseFilteringResult {
  // Current filters
  filters: Ref<FilterState>;
  
  // Available filter configurations
  configs: Ref<FilterConfig[]>;
  
  // Filtered results
  filteredItems: Ref<any[]>;
  
  // Filter actions
  setFilter: (filterName: string, value: string | string[]) => void;
  clearFilter: (filterName: string) => void;
  clearAllFilters: () => void;
  
  // Filter state checks
  hasActiveFilters: Ref<boolean>;
  isFilterActive: (filterName: string, value?: string) => boolean;
  
  // Filter options management
  updateFilterConfig: (config: FilterConfig) => void;
  getFilterOptions: (filterName: string) => FilterOption[];
}

/**
 * Generic filtering composable that works with any data set
 */
export function useFiltering<T extends Record<string, any>>(
  items: Ref<T[]>,
  filterConfigs: FilterConfig[] = []
): UseFilteringResult {
  
  // State
  const filters = ref<FilterState>({});
  const configs = ref<FilterConfig[]>([...filterConfigs]);

  // Initialize filters with empty values
  configs.value.forEach(config => {
    filters.value[config.name] = config.multiple ? [] : '';
  });

  // Computed filtered items
  const filteredItems = computed(() => {
    return items.value.filter(item => {
      return configs.value.every(config => {
        const filterValue = filters.value[config.name];
        
        // No filter applied
        if (!filterValue || (Array.isArray(filterValue) && filterValue.length === 0)) {
          return true;
        }
        
        // Get item value for this filter
        const itemValue = getItemFilterValue(item, config.name);
        
        if (Array.isArray(filterValue)) {
          // Multiple selection filter
          return filterValue.some(fv => matchesFilter(itemValue, fv));
        } else {
          // Single selection filter
          return matchesFilter(itemValue, filterValue);
        }
      });
    });
  });

  // Computed active filters state
  const hasActiveFilters = computed(() => {
    return Object.values(filters.value).some(value => {
      if (Array.isArray(value)) {
        return value.length > 0;
      }
      return value !== '';
    });
  });

  // Actions
  const setFilter = (filterName: string, value: string | string[]): void => {
    const config = configs.value.find(c => c.name === filterName);
    if (!config) {
      console.warn(`Filter config not found: ${filterName}`);
      return;
    }

    filters.value[filterName] = value;
  };

  const clearFilter = (filterName: string): void => {
    const config = configs.value.find(c => c.name === filterName);
    if (!config) return;

    filters.value[filterName] = config.multiple ? [] : '';
  };

  const clearAllFilters = (): void => {
    configs.value.forEach(config => {
      filters.value[config.name] = config.multiple ? [] : '';
    });
  };

  const isFilterActive = (filterName: string, value?: string): boolean => {
    const filterValue = filters.value[filterName];
    
    if (!value) {
      // Check if any filter is active for this name
      if (Array.isArray(filterValue)) {
        return filterValue.length > 0;
      }
      return filterValue !== '';
    }
    
    // Check if specific value is active
    if (Array.isArray(filterValue)) {
      return filterValue.includes(value);
    }
    return filterValue === value;
  };

  const updateFilterConfig = (config: FilterConfig): void => {
    const existingIndex = configs.value.findIndex(c => c.name === config.name);
    
    if (existingIndex >= 0) {
      configs.value[existingIndex] = config;
    } else {
      configs.value.push(config);
      // Initialize filter state for new config
      filters.value[config.name] = config.multiple ? [] : '';
    }
  };

  const getFilterOptions = (filterName: string): FilterOption[] => {
    const config = configs.value.find(c => c.name === filterName);
    return config?.options || [];
  };

  return {
    filters,
    configs,
    filteredItems,
    setFilter,
    clearFilter,
    clearAllFilters,
    hasActiveFilters,
    isFilterActive,
    updateFilterConfig,
    getFilterOptions,
  };
}

/**
 * Category-specific filtering composable
 */
export function useCategoryFiltering<T extends Record<string, any>>(
  items: Ref<T[]>,
  categories: Ref<any[]>,
  options: {
    categoryField?: string;
    categoryIdField?: string;
    categoryNameField?: string;
    includeAllOption?: boolean;
  } = {}
): UseFilteringResult {
  
  const {
    categoryField = 'category',
    categoryIdField = 'id',
    categoryNameField = 'name',
    includeAllOption = true,
  } = options;

  // Generate category filter options from categories
  const categoryOptions = computed(() => {
    const opts: FilterOption[] = [];
    
    if (includeAllOption) {
      opts.push({ value: '', label: 'All Categories' });
    }
    
    categories.value.forEach(category => {
      opts.push({
        value: category[categoryIdField] || category[categoryNameField],
        label: category[categoryNameField] || category[categoryIdField],
        count: countItemsInCategory(items.value, category, categoryField),
      });
    });
    
    return opts;
  });

  // Watch for category changes and update filter config
  const filterConfig: FilterConfig = {
    name: 'category',
    label: 'Category',
    options: categoryOptions.value,
    multiple: false,
    searchable: true,
  };

  const filteringResult = useFiltering(items, [filterConfig]);

  // Update filter config when categories change
  watch(categoryOptions, (newOptions) => {
    filteringResult.updateFilterConfig({
      ...filterConfig,
      options: newOptions,
    });
  });

  return filteringResult;
}

/**
 * Domain store filtering composable - integrates with domain stores
 */
export function useDomainFiltering<T extends Record<string, any>, TStore extends BaseDomainStore>(
  store: TStore,
  itemsField: string,
  categoriesField?: string
) {
  const items = computed(() => (store as any)[itemsField] || []);
  const categories = computed(() => (store as any)[categoriesField || 'categories'] || []);

  // Use category filtering if categories are available
  if (categoriesField && categories.value.length > 0) {
    return useCategoryFiltering(items, categories);
  }

  // Use generic filtering otherwise
  return useFiltering(items);
}

// Helper functions
function getItemFilterValue(item: Record<string, any>, filterName: string): any {
  // Support nested properties with dot notation
  if (filterName.includes('.')) {
    return filterName.split('.').reduce((obj, prop) => obj?.[prop], item);
  }
  return item[filterName];
}

function matchesFilter(itemValue: any, filterValue: string): boolean {
  if (!itemValue || !filterValue) return false;
  
  // Convert to strings for comparison
  const itemStr = String(itemValue).toLowerCase();
  const filterStr = String(filterValue).toLowerCase();
  
  return itemStr.includes(filterStr) || itemStr === filterStr;
}

function countItemsInCategory<T extends Record<string, any>>(
  items: T[],
  category: any,
  categoryField: string
): number {
  return items.filter(item => {
    const itemCategoryValue = getItemFilterValue(item, categoryField);
    return itemCategoryValue === category.id || itemCategoryValue === category.name;
  }).length;
}