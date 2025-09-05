// Research-specific filtering composable
// Integrates research data with generic filtering system

import { ref, computed, watch, type Ref } from 'vue';
import { useResearchArticles } from '../useResearch';
import { useFiltering, type FilterConfig, type FilterOption } from './useFiltering';
import type { ResearchArticle } from '../../lib/clients/research/types';

export interface UseResearchFilteringOptions {
  enableTypeFilter?: boolean;
  enableStatusFilter?: boolean;
  enableAuthorFilter?: boolean;
  includeCounts?: boolean;
}

export interface UseResearchFilteringResult {
  // Filtered research data
  filteredResearch: Ref<ResearchArticle[]>;
  
  // Filter state
  filters: Ref<{ [key: string]: string | string[] }>;
  hasActiveFilters: Ref<boolean>;
  
  // Filter configurations for UI
  typeFilterConfig: Ref<FilterConfig | null>;
  statusFilterConfig: Ref<FilterConfig | null>;
  authorFilterConfig: Ref<FilterConfig | null>;
  
  // Filter actions
  setTypeFilter: (type: string) => void;
  setStatusFilter: (status: string) => void;
  setAuthorFilter: (author: string) => void;
  clearAllFilters: () => void;
  
  // Utility methods
  isFilterActive: (filterName: string, value?: string) => boolean;
}

// Research type labels mapping
const RESEARCH_TYPE_LABELS = {
  clinical_study: 'Clinical Study',
  meta_analysis: 'Meta-Analysis',
  systematic_review: 'Systematic Review',
  case_report: 'Case Report',
  clinical_trial: 'Clinical Trial',
  observational_study: 'Observational Study',
  literature_review: 'Literature Review',
  editorial: 'Editorial',
} as const;

// Publishing status labels mapping
const STATUS_LABELS = {
  published: 'Published',
  draft: 'Draft',
  archived: 'Archived',
} as const;

export function useResearchFiltering(
  researchOptions: Parameters<typeof useResearchArticles>[0] = {},
  filterOptions: UseResearchFilteringOptions = {}
): UseResearchFilteringResult {

  const {
    enableTypeFilter = true,
    enableStatusFilter = true,
    enableAuthorFilter = true,
    includeCounts = true,
  } = filterOptions;

  // Load research data
  const { research, loading: researchLoading } = useResearchArticles(researchOptions);

  // Create filter configurations
  const filterConfigs: FilterConfig[] = [];

  // Research type filter configuration
  const typeFilterConfig = computed(() => {
    if (!enableTypeFilter || researchLoading.value) return null;

    const typeOptions: FilterOption[] = Object.entries(RESEARCH_TYPE_LABELS).map(([type, label]) => ({
      value: type,
      label,
      count: includeCounts ? countResearchByType(research.value, type) : undefined,
    }));

    return {
      name: 'research_type',
      label: 'Research Type',
      options: typeOptions,
      multiple: false,
      searchable: true,
    };
  });

  // Status filter configuration  
  const statusFilterConfig = computed(() => {
    if (!enableStatusFilter || researchLoading.value) return null;

    const statusOptions: FilterOption[] = Object.entries(STATUS_LABELS).map(([status, label]) => ({
      value: status,
      label,
      count: includeCounts ? countResearchByStatus(research.value, status) : undefined,
    }));

    return {
      name: 'publishing_status',
      label: 'Status',
      options: statusOptions,
      multiple: false,
      searchable: false,
    };
  });

  // Author filter configuration
  const authorFilterConfig = computed(() => {
    if (!enableAuthorFilter || researchLoading.value) return null;

    const authors = getUniqueAuthors(research.value);
    const authorOptions: FilterOption[] = authors.map(author => ({
      value: author,
      label: author,
      count: includeCounts ? countResearchByAuthor(research.value, author) : undefined,
    }));

    return {
      name: 'primary_author',
      label: 'Primary Author',
      options: authorOptions,
      multiple: false,
      searchable: true,
    };
  });

  // Build dynamic filter configs
  watch([typeFilterConfig, statusFilterConfig, authorFilterConfig], ([typeConfig, statusConfig, authorConfig]) => {
    filterConfigs.length = 0; // Clear existing configs
    
    if (typeConfig) {
      filterConfigs.push(typeConfig);
    }
    if (statusConfig) {
      filterConfigs.push(statusConfig);
    }
    if (authorConfig) {
      filterConfigs.push(authorConfig);
    }
  }, { immediate: true });

  // Initialize filtering system
  const filteringResult = useFiltering(research, filterConfigs);

  // Convenient filter setters
  const setTypeFilter = (type: string): void => {
    filteringResult.setFilter('research_type', type);
  };

  const setStatusFilter = (status: string): void => {
    filteringResult.setFilter('publishing_status', status);
  };

  const setAuthorFilter = (author: string): void => {
    filteringResult.setFilter('primary_author', author);
  };

  return {
    filteredResearch: filteringResult.filteredItems,
    filters: filteringResult.filters,
    hasActiveFilters: filteringResult.hasActiveFilters,
    typeFilterConfig,
    statusFilterConfig,
    authorFilterConfig,
    setTypeFilter,
    setStatusFilter,
    setAuthorFilter,
    clearAllFilters: filteringResult.clearAllFilters,
    isFilterActive: filteringResult.isFilterActive,
  };
}

// Helper functions
function countResearchByType(research: ResearchArticle[], type: string): number {
  return research.filter(article => article.research_type === type).length;
}

function countResearchByStatus(research: ResearchArticle[], status: string): number {
  return research.filter(article => article.publishing_status === status).length;
}

function countResearchByAuthor(research: ResearchArticle[], author: string): number {
  return research.filter(article => article.primary_author === author).length;
}

function getUniqueAuthors(research: ResearchArticle[]): string[] {
  const authors = new Set<string>();
  research.forEach(article => {
    if (article.primary_author) {
      authors.add(article.primary_author);
    }
  });
  return Array.from(authors).sort();
}

// Type-only filtering composable for simpler use cases
export function useResearchTypeFiltering(
  researchOptions: Parameters<typeof useResearchArticles>[0] = {}
): UseResearchFilteringResult {
  return useResearchFiltering(researchOptions, {
    enableTypeFilter: true,
    enableStatusFilter: false,
    enableAuthorFilter: false,
    includeCounts: true,
  });
}